package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ------------------------------------------------------------
// Session history — persisted to ~/.modelgen_history.jsonl
// Each line is a HistoryEntry JSON object.
// ------------------------------------------------------------

type HistoryEntry struct {
	Timestamp  string `json:"ts"`
	SessionID  string `json:"session_id"`
	Prompt     string `json:"prompt"`
	ModelName  string `json:"model_name,omitempty"`
	FileSaved  string `json:"file_saved,omitempty"`
	NumLines   int    `json:"scad_lines,omitempty"`
}

func historyPath() string {
	return filepath.Join(os.Getenv("HOME"), ".modelgen_history.jsonl")
}

func appendHistory(entry HistoryEntry) {
	f, err := os.OpenFile(historyPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	data, _ := json.Marshal(entry)
	f.Write(data)
	f.WriteString("\n")
}

func loadHistory(limit int) []HistoryEntry {
	data, err := os.ReadFile(historyPath())
	if err != nil {
		return nil
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	var entries []HistoryEntry
	for _, line := range lines {
		if line == "" {
			continue
		}
		var e HistoryEntry
		if json.Unmarshal([]byte(line), &e) == nil {
			entries = append(entries, e)
		}
	}
	// Return last N entries
	if limit > 0 && len(entries) > limit {
		entries = entries[len(entries)-limit:]
	}
	return entries
}

const defaultModel = "qwen3:14b"
const ollamaURL = "http://localhost:11434/api/chat"

// Location of the templates/samples relative to the repo root.
// Resolved from the binary's own path.
var repoRoot string

// French cleat dimensions — stored as a system context so user doesn't repeat it
const frenchCleatContext = `
French cleat dimensions on file:
- Wall cleat material: 19mm (3/4" plywood)
- Cleat angle: 45 degrees
- Hook depth: 22mm (19mm + 3mm clearance)
- Standard slot: 40mm tall strips on wall
When designing a French cleat mount, use these dimensions unless user specifies otherwise.
`

// CNC capabilities context
const cncContext = `
CNC machine capabilities on file:
- Material: plywood (default 18mm / 3/4"), MDF, acrylic
- Bit diameter: 3.175mm (1/8") default
- Dogbone reliefs required for inside 90° corners
- Flatpack designs use finger joints (12mm default finger width)
- Designs should be flat panels exported as SVG or DXF for CNC
`

// 3D print context
const printContext = `
3D printer profile:
- Overhang limit: 45 degrees (beyond this, supports are needed)
- Layer height: 0.2mm default
- Nozzle: 0.4mm
- When designing for 3D print, avoid overhangs beyond 45°, or add chamfers/fillets.
- Flag any overhangs > 45° in comments.
`

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type ChatResponse struct {
	Message Message `json:"message"`
}

var conversation []Message

func systemPrompt() string {
	return `You are a 3D modeling assistant. You generate OpenSCAD parametric code based on user descriptions.
Rules:
1. Always output valid OpenSCAD code wrapped in ` + "```" + `openscad ... ` + "```" + ` blocks.
2. Use parametric variables at the top (never hardcode dimensions in geometry).
3. Respect the CNC, French cleat, and 3D print profiles below — the user doesn't need to re-specify them.
4. For 3D print mode: avoid overhangs >45°, add fillets, comment any unavoidable overhangs.
5. For CNC/flatpack: include dogbone reliefs, use finger joints, design flat panels.
6. When user says "change X to Y", output only the modified .scad with the change applied.
7. Include a comment block at the top: model name, date, key parameters.

` + frenchCleatContext + cncContext + printContext
}

func chat(userMsg string) (string, error) {
	if len(conversation) == 0 {
		conversation = append(conversation, Message{Role: "system", Content: systemPrompt()})
	}
	conversation = append(conversation, Message{Role: "user", Content: userMsg})

	reqBody, _ := json.Marshal(ChatRequest{
		Model:    defaultModel,
		Messages: conversation,
		Stream:   false,
	})

	client := &http.Client{Timeout: 180 * time.Second}
	resp, err := client.Post(ollamaURL, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("parse error: %w\nraw: %s", err, string(body))
	}

	conversation = append(conversation, chatResp.Message)
	return chatResp.Message.Content, nil
}

// extractSCAD pulls the first ```openscad ... ``` block from the response
func extractSCAD(response string) string {
	start := strings.Index(response, "```openscad")
	if start == -1 {
		start = strings.Index(response, "```scad")
	}
	if start == -1 {
		// Maybe raw SCAD without fence
		if strings.Contains(response, "module ") || strings.Contains(response, "// ===") {
			return response
		}
		return ""
	}
	// Find end of opening fence line
	lineEnd := strings.Index(response[start:], "\n")
	if lineEnd == -1 {
		return ""
	}
	codeStart := start + lineEnd + 1
	end := strings.Index(response[codeStart:], "```")
	if end == -1 {
		return response[codeStart:]
	}
	return response[codeStart : codeStart+end]
}

func saveAndRender(scadCode, name, outDir string) error {
	scadPath := filepath.Join(outDir, name+".scad")
	stlPath := filepath.Join(outDir, name+".stl")

	if err := os.WriteFile(scadPath, []byte(scadCode), 0644); err != nil {
		return err
	}
	fmt.Printf("💾 Saved: %s\n", scadPath)

	// Try to render to STL if openscad is installed
	if _, err := exec.LookPath("openscad"); err == nil {
		fmt.Printf("🔧 Rendering STL...\n")
		cmd := exec.Command("openscad", "-o", stlPath, scadPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("⚠️  Render failed (OpenSCAD error): %v\n", err)
		} else {
			fmt.Printf("✅ STL: %s\n", stlPath)
		}
	} else {
		fmt.Printf("ℹ️  OpenSCAD not found — install to auto-render STL\n")
	}
	return nil
}

// autosave writes to ~/.modelgen_autosave/<name>_<ts>.scad
// Returns the saved path, or empty string on failure.
func autosave(scadCode, name string) string {
	dir := filepath.Join(os.Getenv("HOME"), ".modelgen_autosave")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return ""
	}
	ts := time.Now().Format("20060102_150405")
	safe := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			return r
		}
		return '_'
	}, name)
	fname := safe + "_" + ts + ".scad"
	path := filepath.Join(dir, fname)
	if err := os.WriteFile(path, []byte(scadCode), 0644); err != nil {
		return ""
	}
	return path
}

// ------------------------------------------------------------
// extractParams reads top-level variable declarations from a .scad file
// and returns them as a map[name]value. Only handles simple assignments:
//   width = 80;   // comment
// ------------------------------------------------------------
var paramRe = regexp.MustCompile(`(?m)^\s*([a-zA-Z_]\w*)\s*=\s*([^;]+);\s*(?://\s*(.*))?$`)

// isDerivedValue returns true if the value contains operators / function calls,
// meaning it's a computed/derived variable rather than a direct user parameter.
func isDerivedValue(val string) bool {
	// Contains arithmetic operators (but not just a negative sign before a number)
	trimmed := strings.TrimSpace(val)
	// Allow: numbers, strings, true/false — block expressions
	if strings.ContainsAny(trimmed, "+*/") {
		return true
	}
	// Contains function calls like max(), floor(), etc.
	if regexp.MustCompile(`\w+\s*\(`).MatchString(trimmed) {
		return true
	}
	// Contains variable references (identifier not a plain number/bool/string)
	if regexp.MustCompile(`[a-zA-Z_]\w*`).MatchString(trimmed) {
		// It's OK if it's just true/false
		if trimmed == "true" || trimmed == "false" {
			return false
		}
		// It's OK if it starts with a quote (string literal)
		if strings.HasPrefix(trimmed, "\"") {
			return false
		}
		// Otherwise, if it contains an identifier that isn't a number — it's derived
		if !regexp.MustCompile(`^-?[0-9]*\.?[0-9]+$`).MatchString(trimmed) {
			return true
		}
	}
	return false
}

func extractParams(scadSrc string) []paramEntry {
	matches := paramRe.FindAllStringSubmatch(scadSrc, -1)
	var out []paramEntry
	seen := map[string]bool{}
	for _, m := range matches {
		name := strings.TrimSpace(m[1])
		val := strings.TrimSpace(m[2])
		comment := strings.TrimSpace(m[3])
		// Skip non-primitive / long values
		if strings.Contains(val, "[") || len(val) > 60 {
			continue
		}
		// Skip derived/computed variables — only show user-tunable params
		if isDerivedValue(val) {
			continue
		}
		if seen[name] {
			continue
		}
		seen[name] = true
		out = append(out, paramEntry{Name: name, Default: val, Comment: comment})
	}
	return out
}

type paramEntry struct {
	Name    string
	Default string
	Comment string
}

// applyOverrides substitutes parameter values in .scad source.
// overrides is a slice of "name=value" strings.
func applyOverrides(scadSrc string, overrides []string) string {
	for _, kv := range overrides {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			fmt.Fprintf(os.Stderr, "⚠️  Skipping malformed override %q — use name=value\n", kv)
			continue
		}
		name := strings.TrimSpace(parts[0])
		newVal := strings.TrimSpace(parts[1])
		// Replace first occurrence of `name = <old_val>;`
		re := regexp.MustCompile(`(?m)(^\s*` + regexp.QuoteMeta(name) + `\s*=\s*)([^;]+)(;)`)
		if !re.MatchString(scadSrc) {
			fmt.Fprintf(os.Stderr, "⚠️  Parameter %q not found in file\n", name)
			continue
		}
		scadSrc = re.ReplaceAllStringFunc(scadSrc, func(s string) string {
			// Preserve leading whitespace and trailing semicolon
			sub := re.FindStringSubmatch(s)
			return sub[1] + newVal + sub[3]
		})
		fmt.Printf("  ✦ %s = %s\n", name, newVal)
	}
	return scadSrc
}

// ------------------------------------------------------------
// Command: samples — list available sample + template .scad files
// ------------------------------------------------------------
func cmdSamples() {
	dirs := []struct {
		label string
		path  string
	}{
		{"templates", filepath.Join(repoRoot, "openscad", "templates")},
		{"samples", filepath.Join(repoRoot, "openscad", "samples")},
		{"cnc-box", filepath.Join(repoRoot, "openscad", "cnc-box")},
		{"flatpack", filepath.Join(repoRoot, "openscad", "flatpack")},
		{"french-cleat", filepath.Join(repoRoot, "openscad", "french-cleat")},
	}

	fmt.Println()
	fmt.Println("📦 Available models")
	fmt.Println()
	for _, d := range dirs {
		entries, err := os.ReadDir(d.path)
		if err != nil {
			continue
		}
		var names []string
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".scad") {
				names = append(names, e.Name())
			}
		}
		if len(names) == 0 {
			continue
		}
		sort.Strings(names)
		fmt.Printf("  ── %s/ ──\n", d.label)
		for _, n := range names {
			p := filepath.Join(d.path, n)
			src, _ := os.ReadFile(p)
			params := extractParams(string(src))
			paramSummary := ""
			if len(params) > 0 {
				var parts []string
				for i, p := range params {
					if i >= 4 {
						parts = append(parts, "…")
						break
					}
					parts = append(parts, p.Name+"="+p.Default)
				}
				paramSummary = "  [" + strings.Join(parts, ", ") + "]"
			}
			fmt.Printf("  %-38s%s\n", strings.TrimSuffix(n, ".scad"), paramSummary)
		}
		fmt.Println()
	}
	fmt.Println("Use 'modelgen from <name> [key=val ...]' to instantiate with overrides.")
	fmt.Println("Use 'modelgen render <file.scad>' to render a specific file to STL.")
}

// ------------------------------------------------------------
// Command: params — show all parameters in a .scad file
// ------------------------------------------------------------
func cmdParams(path string) {
	src, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", path, err)
		os.Exit(1)
	}
	params := extractParams(string(src))
	if len(params) == 0 {
		fmt.Println("No simple parameters found.")
		return
	}
	fmt.Printf("\n📐 Parameters in %s\n\n", filepath.Base(path))
	maxName := 0
	for _, p := range params {
		if len(p.Name) > maxName {
			maxName = len(p.Name)
		}
	}
	for _, p := range params {
		comment := ""
		if p.Comment != "" {
			comment = "  // " + p.Comment
		}
		fmt.Printf("  %-*s = %s%s\n", maxName, p.Name, p.Default, comment)
	}
	fmt.Printf("\nOverride with: modelgen from %s key=newval ...\n\n", filepath.Base(path))
}

// ------------------------------------------------------------
// Command: render — render a single .scad file to STL
// ------------------------------------------------------------
func cmdRender(scadPath, outDir string) {
	if outDir == "" {
		outDir = filepath.Dir(scadPath)
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot create output dir: %v\n", err)
		os.Exit(1)
	}
	base := strings.TrimSuffix(filepath.Base(scadPath), ".scad")
	stlPath := filepath.Join(outDir, base+".stl")

	if _, err := exec.LookPath("openscad"); err != nil {
		fmt.Fprintln(os.Stderr, "❌ openscad not found in PATH")
		os.Exit(1)
	}

	fmt.Printf("🔧 Rendering %s → %s ...\n", filepath.Base(scadPath), stlPath)
	cmd := exec.Command("openscad", "-o", stlPath, scadPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Render failed: %v\n", err)
		os.Exit(1)
	}
	fi, _ := os.Stat(stlPath)
	if fi != nil {
		fmt.Printf("✅ STL written: %s (%s)\n", stlPath, humanSize(fi.Size()))
	}
}

// ------------------------------------------------------------
// Command: preview — render a .scad to PNG for visual inspection
// Requires xvfb-run (headless) or a display.
// ------------------------------------------------------------
func cmdPreview(scadPath, outDir string, imgW, imgH int) {
	if outDir == "" {
		outDir = filepath.Dir(scadPath)
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot create output dir: %v\n", err)
		os.Exit(1)
	}
	base := strings.TrimSuffix(filepath.Base(scadPath), ".scad")
	pngPath := filepath.Join(outDir, base+".png")

	if _, err := exec.LookPath("openscad"); err != nil {
		fmt.Fprintln(os.Stderr, "❌ openscad not found in PATH")
		os.Exit(1)
	}

	imgArg := fmt.Sprintf("%d,%d", imgW, imgH)

	// Build the openscad args for PNG render
	openscadArgs := []string{
		"--camera", "0,0,0,55,0,25,350",
		"--imgsize", imgArg,
		"--render",
		"--projection=ortho",
		"--colorscheme=Cornfield",
		"-o", pngPath,
		scadPath,
	}

	fmt.Printf("📸 Rendering preview of %s → %s ...\n", filepath.Base(scadPath), pngPath)

	var cmd *exec.Cmd
	// Use xvfb-run if no display is available (headless server)
	if os.Getenv("DISPLAY") == "" {
		if _, err := exec.LookPath("xvfb-run"); err == nil {
			args := append([]string{"-a", "openscad"}, openscadArgs...)
			cmd = exec.Command("xvfb-run", args...)
		} else {
			fmt.Fprintln(os.Stderr, "⚠️  No DISPLAY and xvfb-run not found — PNG preview may fail")
			cmd = exec.Command("openscad", openscadArgs...)
		}
	} else {
		cmd = exec.Command("openscad", openscadArgs...)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ Preview render failed: %v\n", err)
		if len(out) > 0 {
			fmt.Printf("   %s\n", strings.TrimSpace(string(out)))
		}
		os.Exit(1)
	}
	fi, _ := os.Stat(pngPath)
	if fi == nil || fi.Size() < 100 {
		fmt.Println("❌ PNG was empty — render likely failed silently. Check xvfb-run is installed.")
		os.Exit(1)
	}
	fmt.Printf("✅ Preview: %s (%s)\n", pngPath, humanSize(fi.Size()))
	fmt.Printf("   Open with:  xdg-open %s\n", pngPath)
}

// ------------------------------------------------------------
// Command: export — export a .scad to a specific format (dxf, svg, 3mf, amf)
// For CNC workflows: export to SVG/DXF for toolpath generation.
// ------------------------------------------------------------
// detectCNCPanels scans a .scad file for CNC-style panel module names.
// Returns the list of panel module names in layout order.
func detectCNCPanels(scadPath string) []string {
	src, err := os.ReadFile(scadPath)
	if err != nil {
		return nil
	}
	// Known CNC panel module patterns
	candidates := []string{"panel_bottom", "panel_front", "panel_back", "panel_side", "panel_left", "panel_right", "panel_lid", "panel_top"}
	var found []string
	for _, mod := range candidates {
		if bytes.Contains(src, []byte("module "+mod+"(")) {
			found = append(found, mod)
		}
	}
	return found
}

// buildCNCPanelWrapper generates a projection() wrapper that uses the source file's
// modules and arranges them flat for SVG/DXF export.
func buildCNCPanelWrapper(abspath string, panels []string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("use <%s>;\n\n", abspath))
	sb.WriteString("projection(cut=false) {\n")
	// Layout panels in a row with 400mm spacing between them
	for i, panel := range panels {
		offset := i * 400
		sb.WriteString(fmt.Sprintf("    translate([%d, 0, 0]) %s();\n", offset, panel))
	}
	sb.WriteString("}\n")
	return sb.String()
}

func cmdExport(scadPath, format, outDir string) {
	supportedFormats := map[string]bool{
		"dxf": true, "svg": true, "3mf": true, "amf": true, "off": true, "stl": true,
	}
	format = strings.ToLower(format)
	if !supportedFormats[format] {
		fmt.Fprintf(os.Stderr, "❌ Unsupported format %q — choose from: dxf, svg, 3mf, amf, stl\n", format)
		os.Exit(1)
	}

	if outDir == "" {
		outDir = filepath.Dir(scadPath)
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot create output dir: %v\n", err)
		os.Exit(1)
	}

	if _, err := exec.LookPath("openscad"); err != nil {
		fmt.Fprintln(os.Stderr, "❌ openscad not found in PATH")
		os.Exit(1)
	}

	base := strings.TrimSuffix(filepath.Base(scadPath), ".scad")
	outPath := filepath.Join(outDir, base+"."+format)

	fmt.Printf("📤 Exporting %s → %s ...\n", filepath.Base(scadPath), outPath)

	// SVG and DXF require a 2D object. CNC flatpack/box designs export as 3D models
	// but can be flattened to 2D via projection() for CNC toolpath software.
	// Strategy: use `use <file>` (imports modules only, no top-level geometry) then
	// call projection() around the known CNC panel modules.
	needs2D := format == "svg" || format == "dxf"
	actualScad := scadPath

	if needs2D {
		abspath, _ := filepath.Abs(scadPath)

		// Detect CNC panel modules in the source file
		panels := detectCNCPanels(scadPath)

		var wrapperBody string
		if len(panels) > 0 {
			// Known CNC panel layout: use modules + arrange panels in projection
			fmt.Printf("   ℹ️  CNC panel mode — detected panels: %s\n", strings.Join(panels, ", "))
			wrapperBody = buildCNCPanelWrapper(abspath, panels)
		} else {
			// Generic fallback: wrap entire file content in projection using include
			// (works only for files without module-only definitions causing conflicts)
			wrapperBody = fmt.Sprintf("projection(cut=false) {\n  include <%s>;\n}\n", abspath)
		}

		tmpf, err := os.CreateTemp("", "modelgen-export-*.scad")
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Cannot create temp file: %v\n", err)
			os.Exit(1)
		}
		defer os.Remove(tmpf.Name())
		if _, err := tmpf.WriteString(wrapperBody); err != nil {
			tmpf.Close()
			fmt.Fprintf(os.Stderr, "❌ Cannot write temp file: %v\n", err)
			os.Exit(1)
		}
		tmpf.Close()
		actualScad = tmpf.Name()
		fmt.Printf("   ℹ️  2D format — wrapping with projection() for CNC export\n")
	}

	cmd := exec.Command("openscad", "-o", outPath, actualScad)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ Export failed: %v\n%s\n", err, strings.TrimSpace(string(out)))
		if needs2D {
			fmt.Printf("   Tip: for 3D CNC box models, render to STL first, then use your slicer/CAM software to export SVG/DXF.\n")
		}
		os.Exit(1)
	}
	fi, _ := os.Stat(outPath)
	if fi != nil {
		fmt.Printf("✅ Exported: %s (%s)\n", outPath, humanSize(fi.Size()))
	} else {
		fmt.Println("⚠️  Export completed but output file not found.")
	}
	if len(out) > 0 {
		// Print any warnings
		for _, line := range strings.Split(string(out), "\n") {
			if strings.Contains(line, "WARNING") || strings.Contains(line, "ERROR") {
				fmt.Printf("   ⚠️  %s\n", line)
			}
		}
	}
}

// ------------------------------------------------------------
// Command: render-all — render every .scad in a directory
// ------------------------------------------------------------
func cmdRenderAll(dir, outDir string) {
	if outDir == "" {
		outDir = filepath.Join(dir, "stl-output")
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot create output dir: %v\n", err)
		os.Exit(1)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read dir %s: %v\n", dir, err)
		os.Exit(1)
	}

	if _, err := exec.LookPath("openscad"); err != nil {
		fmt.Fprintln(os.Stderr, "❌ openscad not found in PATH")
		os.Exit(1)
	}

	passed, failed := 0, 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".scad") {
			continue
		}
		scadPath := filepath.Join(dir, e.Name())
		stlPath := filepath.Join(outDir, strings.TrimSuffix(e.Name(), ".scad")+".stl")
		fmt.Printf("  🔧 %-35s → ", e.Name())
		cmd := exec.Command("openscad", "-o", stlPath, scadPath)
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("❌  %v\n", err)
			if len(out) > 0 {
				fmt.Printf("      %s\n", strings.TrimSpace(string(out)))
			}
			failed++
		} else {
			fi, _ := os.Stat(stlPath)
			size := ""
			if fi != nil {
				size = humanSize(fi.Size())
			}
			fmt.Printf("✅  %s\n", size)
			passed++
		}
	}
	fmt.Printf("\n%d passed, %d failed  →  %s\n\n", passed, failed, outDir)
}

// ------------------------------------------------------------
// Command: from — instantiate a template/sample with param overrides
// ------------------------------------------------------------
func cmdFrom(nameOrPath string, overrides []string, outDir string) {
	// Resolve path — accept bare name (looks in templates + samples) or a full path
	scadPath := resolveScad(nameOrPath)
	if scadPath == "" {
		fmt.Fprintf(os.Stderr, "❌ Cannot find %q — run 'modelgen samples' to list available models\n", nameOrPath)
		os.Exit(1)
	}

	src, err := os.ReadFile(scadPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", scadPath, err)
		os.Exit(1)
	}

	base := strings.TrimSuffix(filepath.Base(scadPath), ".scad")
	if len(overrides) > 0 {
		fmt.Printf("\n🔩 Applying parameter overrides to %s:\n", base)
	}
	modified := applyOverrides(string(src), overrides)

	if outDir == "" {
		outDir = "./models"
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot create output dir: %v\n", err)
		os.Exit(1)
	}

	// Generate output name
	outName := base
	if len(overrides) > 0 {
		outName = base + "_custom_" + time.Now().Format("150405")
	}

	if err := saveAndRender(modified, outName, outDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// resolveScad finds a .scad file by name or path
func resolveScad(nameOrPath string) string {
	// Direct path
	if _, err := os.Stat(nameOrPath); err == nil {
		return nameOrPath
	}
	// Bare name — search openscad subdirs
	name := nameOrPath
	if !strings.HasSuffix(name, ".scad") {
		name += ".scad"
	}
	searchDirs := []string{
		filepath.Join(repoRoot, "openscad", "templates"),
		filepath.Join(repoRoot, "openscad", "samples"),
		filepath.Join(repoRoot, "openscad", "flatpack"),
		filepath.Join(repoRoot, "openscad", "french-cleat"),
		filepath.Join(repoRoot, "openscad", "cnc-box"),
	}
	for _, d := range searchDirs {
		p := filepath.Join(d, name)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// ------------------------------------------------------------
// Command: install — copy binary to ~/.local/bin/modelgen
// ------------------------------------------------------------
func cmdInstall() {
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot determine binary path: %v\n", err)
		os.Exit(1)
	}
	dest := filepath.Join(os.Getenv("HOME"), ".local", "bin", "modelgen")
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot create ~/.local/bin: %v\n", err)
		os.Exit(1)
	}
	// Read source
	data, err := os.ReadFile(exe)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read binary: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(dest, data, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot write %s: %v\n", dest, err)
		os.Exit(1)
	}
	fmt.Printf("✅ Installed: %s\n", dest)
	fmt.Printf("   Make sure ~/.local/bin is in your PATH.\n")
	fmt.Printf("   Add to ~/.bashrc:  export PATH=\"$HOME/.local/bin:$PATH\"\n")
}

// ------------------------------------------------------------
// Command: history — show past generations
// ------------------------------------------------------------
func cmdHistory(n int) {
	entries := loadHistory(n)
	if len(entries) == 0 {
		fmt.Println("No generation history found. Generate some models first!")
		fmt.Printf("History is stored at: %s\n", historyPath())
		return
	}
	fmt.Printf("\n📜 modelgen history (%d most recent)\n\n", len(entries))
	fmt.Printf("  %-20s  %-12s  %-30s  %s\n", "When", "Session", "File saved", "Prompt")
	fmt.Printf("  %-20s  %-12s  %-30s  %s\n",
		strings.Repeat("─", 20), strings.Repeat("─", 12),
		strings.Repeat("─", 30), strings.Repeat("─", 40))
	for _, e := range entries {
		ts := e.Timestamp
		if t, err := time.Parse(time.RFC3339, ts); err == nil {
			ts = t.Format("Jan 02 15:04")
		}
		sessionShort := e.SessionID
		if len(sessionShort) > 12 {
			sessionShort = sessionShort[:12]
		}
		saved := e.FileSaved
		if saved == "" {
			saved = "(not saved)"
		} else if len(saved) > 30 {
			saved = "…" + saved[len(saved)-29:]
		}
		prompt := e.Prompt
		if len(prompt) > 60 {
			prompt = prompt[:57] + "…"
		}
		fmt.Printf("  %-20s  %-12s  %-30s  %s\n", ts, sessionShort, saved, prompt)
	}
	fmt.Printf("\nHistory file: %s\n\n", historyPath())
}

// ------------------------------------------------------------
// Helpers
// ------------------------------------------------------------
func humanSize(n int64) string {
	switch {
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.0f KB", float64(n)/(1<<10))
	default:
		return strconv.FormatInt(n, 10) + " B"
	}
}

func printHelp() {
	fmt.Println(`modelgen — 3D Model Generator (Friday CLI)

SUBCOMMANDS
  modelgen                         Interactive chat → SCAD generation
  modelgen samples                 List available templates and sample models
  modelgen params <file.scad>      Show all adjustable parameters in a file
  modelgen render <file.scad>      Render a .scad file to STL
  modelgen render-all <dir>        Render all .scad files in a directory
  modelgen preview <file.scad>     Render a .scad to PNG for visual inspection
  modelgen export <file.scad> <fmt> Export to dxf, svg, 3mf, amf (for CNC/slicers)
  modelgen from <name> [k=v ...]   Instantiate a template with param overrides
  modelgen history [N]             Show last N generations (default: 20)
  modelgen install                 Copy binary to ~/.local/bin/modelgen
  modelgen help                    Show this help

FLAGS (for interactive + one-shot modes)
  -prompt <text>    One-shot: generate a model from description and exit
  -name <name>      Output filename (without extension)
  -out <dir>        Output directory (default: ./models)

PREVIEW OPTIONS
  --size <WxH>      Image size for preview PNG (default: 800x600)

HISTORY & AUTO-SAVE
  Every generation in interactive and one-shot mode is auto-saved to:
    ~/.modelgen_autosave/<session>_<ts>.scad
  Use 'modelgen history' to review past prompts and saved files.
  History log: ~/.modelgen_history.jsonl

EXAMPLES
  # List everything available
  modelgen samples

  # Show parameters for the box template
  modelgen params box_parametric

  # Make a custom box with overrides and render to STL
  modelgen from box_parametric width=120 depth=90 height=50 fillet=5

  # Preview a model as PNG
  modelgen preview openscad/samples/phone_stand.scad
  modelgen preview openscad/cnc-box/cnc_routed_box.scad --out ./previews

  # Export for CNC toolpath (SVG or DXF)
  modelgen export openscad/cnc-box/cnc_routed_box.scad svg
  modelgen export openscad/cnc-box/cnc_routed_box.scad dxf --out ./cnc-output

  # Export for slicer (3MF includes colour info)
  modelgen export openscad/samples/phone_stand.scad 3mf

  # Render a specific file
  modelgen render openscad/samples/phone_stand.scad

  # Render all samples to stl-output/
  modelgen render-all openscad/samples

  # One-shot: generate from description
  modelgen -prompt "Makita drill French cleat mount" -name drill_mount

  # Review generation history
  modelgen history
  modelgen history 10

  # Interactive chat mode
  modelgen

PRE-LOADED CONTEXT (you never need to specify these)
  French cleat: 19mm ply, 45° angle, 22mm hook depth
  CNC: 3.175mm bit, dogbone reliefs, 18mm plywood, finger joints
  3D print: 0.4mm nozzle, 0.2mm layers, 45° overhang limit`)
}


// ------------------------------------------------------------
// initRepoRoot — find the repo root by walking up from the binary,
// or from MODELGEN_ROOT env var, or the known default install path.
// ------------------------------------------------------------
func initRepoRoot() {
	// 1. Explicit env override
	if env := os.Getenv("MODELGEN_ROOT"); env != "" {
		repoRoot = env
		return
	}

	// 2. Known default location
	defaultPath := filepath.Join(os.Getenv("HOME"), "Documents", "code", "3d-model-generation")
	if _, err := os.Stat(filepath.Join(defaultPath, "openscad")); err == nil {
		repoRoot = defaultPath
		return
	}

	// 3. Walk up from the binary (works when run directly from the repo)
	exe, err := os.Executable()
	if err != nil {
		repoRoot = "."
		return
	}
	dir := filepath.Dir(exe)
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(dir, "openscad")); err == nil {
			repoRoot = dir
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// 4. Fallback: cwd
	repoRoot = "."
}

// ------------------------------------------------------------
// main
// ------------------------------------------------------------
func main() {
	initRepoRoot()

	args := os.Args[1:]

	// Handle subcommands first
	if len(args) > 0 {
		switch args[0] {
		case "samples", "list":
			cmdSamples()
			return

		case "params":
			if len(args) < 2 {
				fmt.Fprintln(os.Stderr, "Usage: modelgen params <file.scad>")
				os.Exit(1)
			}
			p := args[1]
			if !strings.HasSuffix(p, ".scad") {
				p = resolveScad(p)
				if p == "" {
					fmt.Fprintf(os.Stderr, "❌ File not found: %s\n", args[1])
					os.Exit(1)
				}
			}
			cmdParams(p)
			return

		case "render":
			if len(args) < 2 {
				fmt.Fprintln(os.Stderr, "Usage: modelgen render <file.scad> [--out <dir>]")
				os.Exit(1)
			}
			outDir := ""
			for i, a := range args[2:] {
				if a == "--out" && i+1 < len(args[2:]) {
					outDir = args[2:][i+1]
				}
			}
			cmdRender(args[1], outDir)
			return

		case "render-all":
			if len(args) < 2 {
				fmt.Fprintln(os.Stderr, "Usage: modelgen render-all <dir> [--out <dir>]")
				os.Exit(1)
			}
			outDir := ""
			for i, a := range args[2:] {
				if a == "--out" && i+1 < len(args[2:]) {
					outDir = args[2:][i+1]
				}
			}
			cmdRenderAll(args[1], outDir)
			return

		case "preview":
			if len(args) < 2 {
				fmt.Fprintln(os.Stderr, "Usage: modelgen preview <file.scad> [--out <dir>] [--size WxH]")
				os.Exit(1)
			}
			outDir := ""
			imgW, imgH := 800, 600
			for i := 2; i < len(args); i++ {
				switch {
				case args[i] == "--out" && i+1 < len(args):
					outDir = args[i+1]
					i++
				case strings.HasPrefix(args[i], "--out="):
					outDir = strings.TrimPrefix(args[i], "--out=")
				case args[i] == "--size" && i+1 < len(args):
					parts := strings.SplitN(args[i+1], "x", 2)
					if len(parts) == 2 {
						fmt.Sscanf(parts[0], "%d", &imgW)
						fmt.Sscanf(parts[1], "%d", &imgH)
					}
					i++
				case strings.HasPrefix(args[i], "--size="):
					parts := strings.SplitN(strings.TrimPrefix(args[i], "--size="), "x", 2)
					if len(parts) == 2 {
						fmt.Sscanf(parts[0], "%d", &imgW)
						fmt.Sscanf(parts[1], "%d", &imgH)
					}
				}
			}
			scadPath := args[1]
			if !strings.HasSuffix(scadPath, ".scad") {
				scadPath = resolveScad(scadPath)
				if scadPath == "" {
					fmt.Fprintf(os.Stderr, "❌ File not found: %s\n", args[1])
					os.Exit(1)
				}
			}
			cmdPreview(scadPath, outDir, imgW, imgH)
			return

		case "export":
			if len(args) < 3 {
				fmt.Fprintln(os.Stderr, "Usage: modelgen export <file.scad> <format> [--out <dir>]")
				fmt.Fprintln(os.Stderr, "Formats: dxf, svg, 3mf, amf, stl")
				os.Exit(1)
			}
			outDir := ""
			for i := 3; i < len(args); i++ {
				if args[i] == "--out" && i+1 < len(args) {
					outDir = args[i+1]
					i++
				} else if strings.HasPrefix(args[i], "--out=") {
					outDir = strings.TrimPrefix(args[i], "--out=")
				}
			}
			scadPath := args[1]
			if !strings.HasSuffix(scadPath, ".scad") {
				scadPath = resolveScad(scadPath)
				if scadPath == "" {
					fmt.Fprintf(os.Stderr, "❌ File not found: %s\n", args[1])
					os.Exit(1)
				}
			}
			cmdExport(scadPath, args[2], outDir)
			return

		case "from":
			if len(args) < 2 {
				fmt.Fprintln(os.Stderr, "Usage: modelgen from <name|file> [key=val ...]")
				os.Exit(1)
			}
			outDir := "./models"
			remaining := args[2:]
			var overrides []string
			for _, a := range remaining {
				if strings.HasPrefix(a, "--out=") {
					outDir = strings.TrimPrefix(a, "--out=")
				} else if a == "--out" {
					// handled below via index — simple approach: collect non-flag args
				} else if strings.Contains(a, "=") {
					overrides = append(overrides, a)
				}
			}
			cmdFrom(args[1], overrides, outDir)
			return

		case "install":
			cmdInstall()
			return

		case "history":
			n := 20
			if len(args) >= 2 {
				if v, err := strconv.Atoi(args[1]); err == nil && v > 0 {
					n = v
				}
			}
			cmdHistory(n)
			return

		case "help", "--help", "-h":
			printHelp()
			return
		}
	}

	// Legacy flag-based mode (interactive + -prompt one-shot)
	outDir := "./models"
	name := ""
	oneShot := ""

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-out", "--out":
			if i+1 < len(args) {
				outDir = args[i+1]
				i++
			}
		case "-name", "--name":
			if i+1 < len(args) {
				name = args[i+1]
				i++
			}
		case "-prompt", "--prompt":
			if i+1 < len(args) {
				oneShot = args[i+1]
				i++
			}
		}
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot create output dir: %v\n", err)
		os.Exit(1)
	}

	// Session ID: short timestamp for grouping history entries
	sessionID := time.Now().Format("20060102_150405")

	// One-shot mode
	if oneShot != "" {
		fmt.Printf("🤖 Generating model...\n")
		resp, err := chat(oneShot)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		scad := extractSCAD(resp)
		if scad == "" {
			fmt.Println("No SCAD code found in response. Full response:")
			fmt.Println(resp)
			os.Exit(1)
		}
		modelName := name
		if modelName == "" {
			modelName = "model_" + sessionID
		}
		savedPath := filepath.Join(outDir, modelName+".scad")
		if err := saveAndRender(scad, modelName, outDir); err != nil {
			fmt.Fprintf(os.Stderr, "Save error: %v\n", err)
			os.Exit(1)
		}
		appendHistory(HistoryEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			SessionID: sessionID,
			Prompt:    oneShot,
			ModelName: modelName,
			FileSaved: savedPath,
			NumLines:  strings.Count(scad, "\n"),
		})
		return
	}

	// Interactive mode
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Println("║   3D Model Generator — Friday CLI      ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Printf("Session: %s\n", sessionID)
	fmt.Println("Describe a model, say 'save <name>' to save, 'exit' to quit.")
	fmt.Println("French cleat dims, CNC params, and print profiles are pre-loaded.")
	fmt.Println("Tip: run 'modelgen samples' to see available templates.")
	fmt.Println()

	currentSCAD := ""
	currentName := name
	lastPrompt := ""

	for {
		fmt.Print("you> ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if input == "exit" || input == "quit" {
			break
		}
		if strings.HasPrefix(input, "save") {
			parts := strings.Fields(input)
			saveName := currentName
			if len(parts) > 1 {
				saveName = parts[1]
			}
			if saveName == "" {
				saveName = "model_" + sessionID
			}
			if currentSCAD == "" {
				fmt.Println("Nothing to save yet — generate a model first.")
				continue
			}
			if err := saveAndRender(currentSCAD, saveName, outDir); err != nil {
				fmt.Fprintf(os.Stderr, "Save error: %v\n", err)
			} else {
				savedPath := filepath.Join(outDir, saveName+".scad")
				appendHistory(HistoryEntry{
					Timestamp: time.Now().Format(time.RFC3339),
					SessionID: sessionID,
					Prompt:    lastPrompt,
					ModelName: saveName,
					FileSaved: savedPath,
					NumLines:  strings.Count(currentSCAD, "\n"),
				})
			}
			currentName = saveName
			continue
		}

		fmt.Printf("🤖 Thinking...\n")
		resp, err := chat(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}

		scad := extractSCAD(resp)
		if scad != "" {
			currentSCAD = scad
			lastPrompt = input

			// Auto-save to ~/.modelgen_autosave/
			autosaved := autosave(scad, sessionID)

			fmt.Printf("\n📐 Generated SCAD (%d lines):\n", strings.Count(scad, "\n"))
			// Show first 15 lines as preview
			lines := strings.Split(scad, "\n")
			preview := lines
			if len(lines) > 15 {
				preview = lines[:15]
				fmt.Println(strings.Join(preview, "\n"))
				fmt.Printf("... (%d more lines)\n", len(lines)-15)
			} else {
				fmt.Println(strings.Join(preview, "\n"))
			}
			if autosaved != "" {
				fmt.Printf("📂 Auto-saved: %s\n", autosaved)
			}
			fmt.Println("Say 'save <name>' to write to ./models/ and render STL.")
		} else {
			// No SCAD block — show the response as-is (explanation or question)
			fmt.Println()
			fmt.Println(resp)
		}
		fmt.Println()
	}
}
