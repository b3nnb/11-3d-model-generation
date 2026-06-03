package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const defaultModel = "qwen3:14b"
const ollamaURL = "http://localhost:11434/api/chat"

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

	client := &http.Client{Timeout: 120 * time.Second}
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

func main() {
	outDir := flag.String("out", "./models", "Output directory for .scad and .stl files")
	name := flag.String("name", "", "Model name (used for filenames, prompted if empty)")
	oneShot := flag.String("prompt", "", "Non-interactive: describe the model and exit")
	flag.Parse()

	if err := os.MkdirAll(*outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot create output dir: %v\n", err)
		os.Exit(1)
	}

	// One-shot mode
	if *oneShot != "" {
		fmt.Printf("🤖 Generating model...\n")
		resp, err := chat(*oneShot)
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
		modelName := *name
		if modelName == "" {
			modelName = "model_" + time.Now().Format("20060102_150405")
		}
		if err := saveAndRender(scad, modelName, *outDir); err != nil {
			fmt.Fprintf(os.Stderr, "Save error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Interactive mode
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Println("║   3D Model Generator — Friday CLI      ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Println("Describe a model, say 'save <name>' to save, 'exit' to quit.")
	fmt.Println("French cleat dims, CNC params, and print profiles are pre-loaded.")
	fmt.Println()

	currentSCAD := ""
	currentName := *name

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
				saveName = "model_" + time.Now().Format("20060102_150405")
			}
			if currentSCAD == "" {
				fmt.Println("Nothing to save yet — generate a model first.")
				continue
			}
			if err := saveAndRender(currentSCAD, saveName, *outDir); err != nil {
				fmt.Fprintf(os.Stderr, "Save error: %v\n", err)
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
			fmt.Printf("\n📐 Generated SCAD (%d lines):\n", strings.Count(scad, "\n"))
			// Show first 15 lines as preview
			lines := strings.Split(scad, "\n")
			preview := lines
			if len(lines) > 15 {
				preview = lines[:15]
				fmt.Println(strings.Join(preview, "\n"))
				fmt.Printf("... (%d more lines) — say 'save <name>' to save\n", len(lines)-15)
			} else {
				fmt.Println(strings.Join(preview, "\n"))
				fmt.Println("Say 'save <name>' to write the file.")
			}
		} else {
			// No SCAD block — show the response as-is (explanation or question)
			fmt.Println()
			fmt.Println(resp)
		}
		fmt.Println()
	}
}
