# Parameter Definitions & Adjustment Instructions
## 3D Model Generation — OpenSCAD Toolkit

---

## Why Parametric Design

Every model in this library is **fully parametric**: all dimensions live in named variables at the top of the file. To customize:

1. Open the `.scad` file in any text editor or OpenSCAD
2. Change the value(s) at the top
3. Re-render: `openscad -o output.stl model.scad`

---

## Common Parameter Names (used across all templates)

| Parameter | Type | Typical Range | Description |
|-----------|------|---------------|-------------|
| `wall` / `wall_t` | mm | 2–6mm | Wall thickness. 3mm is a solid default for most 3D prints. Use 4–5mm for structural parts. |
| `fillet` / `corner_r` | mm | 0–5mm | Corner rounding. 0 = sharp corners. Larger values look better and print better. |
| `clearance` | mm | 0.2–1.5mm | Tolerance gap for press-fit or snap-fit features. Adjust ±0.2mm per printer. |
| `base_t` | mm | 2–4mm | Bottom plate thickness. 3mm for standard prints. |
| `hole_d` / `screw_d` | mm | +0.2–0.5mm | Screw clearance holes. Add 0.2–0.5mm to nominal screw size. M3=3.5mm, M4=4.5mm. |
| `$fn` | integer | 16–64 | OpenSCAD circle facets. 24 is fine for most; 48+ for visible cylinders. |

---

## Template-Specific Parameters

### `box_parametric.scad` — General Box
| Parameter | Default | Notes |
|-----------|---------|-------|
| `width` / `depth` / `height` | 80/60/40mm | External or internal? These are **external** dims. |
| `mode` | `"print3d"` | Use `"cnc"` for flat-bottom CNC designs. |
| `include_lid_lip` | `true` | Set to `false` for open-top tray. |
| `lid_clearance` | 0.3mm | Reduce to 0.2mm for tight snap, increase to 0.5mm for loose fit. |

**Common adjustments:**
- Phone charging dock: `width=90, depth=120, height=20, include_lid_lip=false`
- Tool tray: `width=200, depth=150, height=50, fillet=0, include_lid_lip=false`

---

### `french_cleat_mount.scad` — Cleat Hook
| Parameter | Default | Notes |
|-----------|---------|-------|
| `mount_width` | 80mm | Width of your specific tool holder |
| `mount_depth` | 60mm | How far it sticks out from wall |
| `mount_height` | 120mm | Total height — depends on what you're hanging |
| `cleat_material` | 19mm | **Don't change** unless your cleats aren't standard 3/4" plywood |

**Common adjustments:**
- Pencil/marker holder: `mount_width=100, mount_depth=80, mount_height=150`
- Power drill cradle: `mount_width=180, mount_depth=100, mount_height=200`

---

### `flatpack_box.scad` — CNC Finger Joint Box
| Parameter | Default | Notes |
|-----------|---------|-------|
| `material_t` | 12mm | **Must match your actual plywood thickness** — measure with calipers |
| `bit_dia` | 3.175mm | Must match your CNC bit. Common: 3.175mm (1/8"), 6mm (1/4") |
| `kerf` | 0.1mm | Router kerf. May need ±0.05mm adjustment per machine |
| `dogbone` | `true` | Always true for 90° inside corners on CNC |
| `box_w/d/h` | 150/100/80mm | **Internal** dimensions |
| `finger_w` | 12mm | Finger joint tab width. Keep at 10–15mm for 12mm ply. |

**Common adjustments:**
- Seed/parts storage: `box_w=100, box_d=60, box_h=40, material_t=6`
- Tool organizer box: `box_w=300, box_d=200, box_h=100, finger_w=20`

---

### Sample Models

| File | Key Parameters | What to Adjust |
|------|---------------|----------------|
| `phone_stand.scad` | `phone_w`, `angle` | `phone_w` = your phone width + case; `angle` 60–75° for portrait, 70–80° for landscape |
| `cable_clip.scad` | `cable_dia` | Match to cable OD. Print a test at 4.5mm, 7mm, 10mm and keep the right sizes. |
| `pegboard_hook.scad` | `hook_reach`, `hook_h` | `hook_reach` = depth of tool. `hook_h` = height needed for tool to slide on. |
| `rpi_case.scad` | `top_clear`, `port_h` | Increase `top_clear` if you have tall components (heatsink, HAT). |
| `shelf_bracket.scad` | `shelf_depth`, `shelf_load` | Increase `wall_t` to 6–8mm for loads >5kg. |
| `spool_holder.scad` | `spool_hub_dia`, `rod_dia` | Measure your spool hub ID before printing. |
| `drawer_divider.scad` | `drawer_w/d/h` | Measure drawer interior. Print `div_t=2.5mm` dividers, cross-fit at 90°. |
| `electronics_enclosure.scad` | `pcb_w/d`, `top_clear` | Match PCB dims. Increase `top_clear` for tall caps/heatsinks. |
| `handle.scad` | `handle_dia`, `mount_w` | `handle_dia` 22–28mm is ergonomic. Measure hole spacing on original. |
| `drill_jig.scad` | `hole_dia`, `hole_y` | Match drill bits you'll use. `hole_y` = distance from edge where holes land. |

---

## Adjusting for Print Tolerances

Different printers print slightly differently. Run these calibration steps once:

1. **Hole calibration**: Print `drill_jig.scad` with `hole_dia=[5]`. Measure the printed hole with calipers. If it's 4.7mm instead of 5.0mm, add `+0.3` to all your hole diameters.

2. **Wall calibration**: Print a 20×20×10mm cube (use `box_parametric.scad` with `width=20, depth=20, height=10, include_lid_lip=false, fillet=0`). Measure wall thickness. Adjust `wall` parameter accordingly.

3. **Fit calibration**: Print two pieces that snap together (e.g. `box_parametric.scad` body + lid). If too tight, increase `lid_clearance` by 0.1mm increments. If too loose, decrease.

---

## Adjustment Workflow (via modelgen CLI)

```bash
# Start interactive session
modelgen

# Describe what you need
you> I need a wall bracket for my Makita drill, French cleat mount

# After it generates, iterate:
you> make the mount 20mm wider
you> add a slot for the battery pack

# Save when happy:
you> save makita-drill-mount
```

The CLI preserves context between turns — you never need to re-describe French cleat dimensions or your CNC specs.
