# 3D Model Generation

AI-assisted parametric 3D model generator for CNC and 3D printing workflows.

## What this is

A natural language CLI (`modelgen`) + OpenSCAD template library that lets you describe what you need and get print-ready `.scad` / `.stl` files. French cleat dimensions, CNC profiles, and 3D print tolerances are pre-loaded — you never re-specify them.

## Quick Start

```bash
# Interactive REPL
cd cli
./modelgen

you> I need a French cleat mount for my Makita drill
# → generates SCAD, shows 15-line preview

you> make it 20mm wider and add a battery slot
# → updates model in-context

you> save makita-drill-mount
# → writes .scad and renders .stl (if openscad installed)
```

## One-Shot Mode

```bash
./modelgen --prompt "CNC flatpack box 200x150x100mm for my workbench" --name workbench-box
```

## Visualize Parameters

```bash
# Show all adjustable parameters in a model
python3 cli/vizparams.py openscad/samples/phone_stand.scad

# Scan all samples
python3 cli/vizparams.py openscad/samples/

# JSON output (for scripting)
python3 cli/vizparams.py openscad/samples/ --json
```

## Template Library

### `/openscad/templates/`
| File | What it is |
|------|-----------|
| `box_parametric.scad` | General purpose box — 3D print or CNC mode, lid lip |
| `french_cleat_mount.scad` | French cleat wall mount with 45° hook geometry |
| `flatpack_box.scad` | CNC finger-joint flatpack box with dogbone reliefs |

### `/openscad/samples/` — 10 ready-to-use models
| File | Description |
|------|-------------|
| `phone_stand.scad` | Adjustable desk phone stand (angle, width) |
| `cable_clip.scad` | Press-fit cable management clips (3 sizes) |
| `pegboard_hook.scad` | Standard 25.4mm pegboard hook |
| `rpi_case.scad` | Raspberry Pi 4B snap-fit case with lid |
| `shelf_bracket.scad` | French-cleat-compatible wall shelf bracket |
| `spool_holder.scad` | Filament spool wall mount (standard 200mm spools) |
| `drawer_divider.scad` | Cross-fit drawer divider strips |
| `electronics_enclosure.scad` | PCB enclosure with standoffs, cable glands, lid |
| `handle.scad` | Ergonomic replacement handle, screw-mount |
| `drill_jig.scad` | Drill guide jig with multiple hole sizes |

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--out` | `./models` | Output directory for .scad and .stl files |
| `--name` | auto-timestamped | Model filename |
| `--prompt` | — | Non-interactive one-shot description |

## Pre-loaded Contexts

The CLI pre-loads these so you never have to repeat them:

- **French cleat**: 19mm plywood, 45° angle, 22mm hook depth
- **CNC machine**: 3.175mm (1/8") bit, dogbone reliefs, finger joints, 18mm plywood default
- **3D printer**: 0.4mm nozzle, 0.2mm layers, 45° overhang limit

## Documentation

| Doc | Contents |
|-----|----------|
| [`docs/parameter-guide.md`](docs/parameter-guide.md) | Every parameter across all templates explained, with common adjustment recipes |
| [`docs/export-checklist.md`](docs/export-checklist.md) | Pre-print/pre-CNC validation checklist for STL and DXF export |
| [`docs/blender-organic-workflow.md`](docs/blender-organic-workflow.md) | Blender sculpt workflow for organic models + iterative adjustment format |

## Model Type Guide

| Need | Tool | Template |
|------|------|----------|
| Box / tray / enclosure | OpenSCAD CLI | `box_parametric.scad` |
| French cleat storage | OpenSCAD CLI | `french_cleat_mount.scad` |
| CNC flatpack | OpenSCAD CLI | `flatpack_box.scad` |
| Electronics case | OpenSCAD CLI | `electronics_enclosure.scad` |
| Custom tool holder | OpenSCAD CLI (generate) | — |
| Character / organic shape | Blender (manual + instructions) | see `blender-organic-workflow.md` |

## Requirements

- **Go** 1.21+ (for CLI build)
- **Ollama** running locally (`http://localhost:11434`) with a model (default: `qwen3:14b`)
- **OpenSCAD** (optional) — for auto-rendering `.stl` on `save`
- **Python 3.10+** — for `vizparams.py`

## Build CLI

```bash
cd cli
go build -o modelgen .
# Optional: install globally
cp modelgen ~/.local/bin/
```
