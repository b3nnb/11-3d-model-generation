# Export Validation Checklist
## 3D Model Generation — Pre-Print / Pre-Cut Checklist

Run through this before sending any file to a printer or CNC machine.

---

## 3D Print Checklist (STL / OBJ output)

### ✅ Geometry Checks (in OpenSCAD or slicer)

- [ ] **No open edges** — Model is watertight (manifold). OpenSCAD auto-handles this; verify in slicer (PrusaSlicer / Bambu Studio shows warnings).
- [ ] **No zero-thickness walls** — Every wall is ≥ 1 nozzle width (0.4mm). Walls thinner than 0.8mm may not print.
- [ ] **Overhangs ≤ 45°** — Beyond 45° = needs supports. Check in slicer's overhang preview (red highlighting).
- [ ] **Minimum feature size ≥ 2mm** — Anything smaller may not print cleanly at 0.4mm nozzle / 0.2mm layers.
- [ ] **Correct scale** — Open in slicer, verify dimensions match expected. Common issue: `.stl` exported in inches vs mm.
- [ ] **Single body** — Unless intentionally multi-part, all geometry should be unioned. OpenSCAD `union()` ensures this.

### ✅ STL Export Settings (OpenSCAD)

- [ ] **Units = mm** — OpenSCAD defaults to mm. Never change this unless you know why.
- [ ] **`$fn` ≥ 24 on visible cylinders** — Lower `$fn` = faceted circles. 48 for show surfaces, 24 for hidden.
- [ ] **STL file size < 50MB** — If larger, reduce `$fn` values or split model.
- [ ] **No degenerate geometry** — Check for warnings in OpenSCAD render log. Fix any `CGAL errors`.

### ✅ Slicer Settings Confirmation

- [ ] Layer height matches material/tolerance needs (0.2mm default)
- [ ] Infill ≥ 20% for structural parts, 10-15% for cosmetic
- [ ] Supports enabled if overhangs >45°
- [ ] First layer calibrated (Z-offset set for your printer)
- [ ] Print time estimated and reasonable (flag if >8 hours for review)

---

## CNC Checklist (DXF / SVG output)

### ✅ Geometry Checks

- [ ] **All geometry is flat (2D panels)** — CNC routes flat sheets. Z-axis only for depth of cut.
- [ ] **No overlapping lines** — Duplicate lines cause double-passes and dimensional errors.
- [ ] **Closed loops** — All cut paths are closed polygons. Open paths = incomplete cuts.
- [ ] **Dogbone reliefs present** — All inside 90° corners have dogbone circles for full-depth square corners.
- [ ] **Tab/bridge placement** — For through-cuts, tabs hold parts in place. Add 3–5mm tabs every 150mm on long cuts.
- [ ] **Grain direction noted** — Mark plywood grain direction on design when it matters for strength.

### ✅ DXF Export Settings

- [ ] **Units = mm** (not inches)
- [ ] **DXF version = R14 or R2000** — Most CNC controllers accept these. Avoid newer formats.
- [ ] **One layer per operation** — Cut lines on one layer, engrave on another. Simplifies CAM setup.
- [ ] **No construction geometry in export** — Remove guide lines, dimension lines, center marks before export.

### ✅ Material & Machine Config

- [ ] Material thickness measured with calipers (not assumed) and matched to `material_t` parameter
- [ ] Bit diameter matched to `bit_dia` parameter
- [ ] Feeds/speeds set in CAM (not in OpenSCAD — these are CAM settings)
- [ ] Toolpath preview run in CAM software before machine start
- [ ] Workpiece secured — clamps, tabs, or vacuum

---

## File Naming Convention

Use consistent naming to track iterations:

```
{model-name}-v{version}-{mode}-{date}.{ext}
```

Examples:
```
makita-drill-mount-v1-print3d-20260603.stl
workbench-box-v3-cnc-20260603.dxf
rpi-case-body-v2-print3d-20260603.stl
rpi-case-lid-v2-print3d-20260603.stl
```

**Version bump triggers:**
- Any dimension change
- Mode change (print3d ↔ cnc)
- Structural change to geometry

---

## Load/Stress Spot Checks

For functional parts bearing loads, quickly verify:

| Use Case | Minimum Wall | Infill | Material |
|----------|-------------|--------|----------|
| Cosmetic/display | 2mm | 10% | PLA |
| Light duty (phone stand, organizer) | 3mm | 15% | PLA/PETG |
| Medium load (shelf bracket, tool holder) | 4–5mm | 30% | PETG |
| Heavy load (>5kg, clamps, jigs) | 6mm+ | 40–50% | PETG/ABS |
| CNC jigs / machining fixtures | 5mm+ | 40%+ | PETG/ABS |

---

## Quick Render Test (run before final export)

```bash
# Render and check for errors
openscad -o /tmp/test.stl your_model.scad 2>&1 | grep -E "(WARNING|ERROR|CGAL)"

# Check STL is valid and get stats
openscad --info /tmp/test.stl 2>/dev/null || python3 -c "
import struct, sys
with open('/tmp/test.stl', 'rb') as f:
    f.read(80)  # header
    n = struct.unpack('<I', f.read(4))[0]
    print(f'Triangles: {n}')
    print(f'File size: {50*n/1024:.1f} KB estimated')
"
```

---

## modelgen CLI — Export Commands

```bash
# Save current model
you> save my-model-name

# One-shot with direct save
modelgen --prompt "parametric box 100x80x50" --name storage-box --out ./output

# Batch render all templates
for f in openscad/samples/*.scad; do
    name=$(basename "$f" .scad)
    openscad -o "output/stl/${name}.stl" "$f" && echo "✅ $name"
done
```
