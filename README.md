# 3D Model Generation

AI-assisted parametric 3D model generator for CNC and 3D printing workflows.

## What this is

A natural language CLI that talks to your local Ollama LLM to generate OpenSCAD parametric models. You describe what you want, it writes the `.scad` file and optionally renders the STL.

**Pre-loaded contexts (you never re-specify these):**
- French cleat dimensions (19mm plywood, 45° angle, 22mm hook depth)
- CNC machine profile (3.175mm bit, dogbone reliefs, flatpack finger joints)
- 3D print profile (45° overhang limit, 0.4mm nozzle)

## Usage

### Interactive mode
```bash
cd cli
./modelgen

you> I need a French cleat mount for a Makita drill
# → generates SCAD, shows preview
you> change the hook depth to 25mm
# → updates model
you> save makita-drill-mount
# → saves .scad and renders .stl (if openscad installed)
```

### One-shot mode
```bash
./modelgen --prompt "CNC flatpack box 200x150x100mm for my workbench" --name workbench-box --out ./models
```

### Flags
| Flag | Default | Description |
|------|---------|-------------|
| `--out` | `./models` | Output directory |
| `--name` | auto-timestamped | Model filename |
| `--prompt` | — | Non-interactive one-shot description |

## Templates included

### `/openscad/templates/box_parametric.scad`
General purpose parametric box. Works for 3D print (rounded corners, lid lip) or CNC (flat base).

Variables: `width`, `depth`, `height`, `wall`, `base`, `fillet`, `mode`, `include_lid_lip`

### `/openscad/french-cleat/french_cleat_mount.scad`
French cleat wall mount base template. The CLI extends this per request.

Variables: `cleat_width`, `mount_width`, `mount_depth`, `mount_height`, `wall_t`

### `/openscad/flatpack/flatpack_box.scad`
CNC flatpack box with finger joints and dogbone reliefs.

Variables: `material_t`, `bit_dia`, `box_w`, `box_d`, `box_h`, `finger_w`

## Installing binary

```bash
cp cli/modelgen ~/.local/bin/modelgen
# then from anywhere:
modelgen --prompt "shelf bracket for French cleat, 200mm wide" --name shelf-bracket
```

## Phase 2 (planned)

- Blender organic model generation via Python scripts
- NAS archiving of completed designs
- CTL 3D tab integration with per-idea design chat
