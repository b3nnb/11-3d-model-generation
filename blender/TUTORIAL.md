# Blender Organic Model Workflow — Tutorial

A practical guide for iterative organic sculpting in Blender, focused on 3D printing. This covers the full loop: setup → block out → sculpt → adjust → export.

---

## Contents

1. [Setup](#setup)
2. [Scripts in this folder](#scripts)
3. [Phase 1 — Block out base mesh](#phase-1)
4. [Phase 2 — Sculpt mode basics](#phase-2)
5. [Phase 3 — Iterative adjustment loop](#phase-3)
6. [Phase 4 — 3D print prep](#phase-4)
7. [Phase 5 — Export and slice](#phase-5)
8. [Keyboard shortcut reference](#shortcuts)
9. [Troubleshooting](#troubleshooting)

---

## 1. Setup <a name="setup"></a>

### Units — always do this first

```
Scene Properties (camera icon) > Units:
  Unit System: Metric
  Length: Millimeters
  Scale: 0.001
```

This makes 1 Blender unit = 1mm. **Everything else depends on this.**

### Enable 3D Print Toolbox

```
Edit > Preferences > Add-ons > search "3D Print" > enable "Mesh: 3D-Print Toolbox"
```

Gives you wall-thickness check, non-manifold detection, and overhang analysis.

### Enable X Mirror (for symmetric sculpting)

In Sculpt Mode, look in the header bar for the butterfly icon — click it. Now both sides of your model are sculpted symmetrically.

---

## 2. Scripts in this folder <a name="scripts"></a>

| Script | What it does |
|--------|-------------|
| `setup_organic_base.py` | Creates a pre-configured base mesh (sphere/cylinder/cube/head) with correct mm units and Subdivision Surface modifier |
| `iterate_organic.py` | Translates natural language adjustment descriptions into Blender step-by-step instructions |
| `export_for_print.py` | Pre-flight checks + export STL with correct settings |

**How to run:**

1. Open Blender
2. Click **Scripting** tab at the top
3. Click **Open** (or **New** and paste)
4. Load the script
5. Press **Run Script** (Alt+P) or the ▶ button

Or from command line:
```bash
blender --background --python setup_organic_base.py -- --type sphere --output /tmp/base.blend
```

---

## 3. Phase 1 — Block out base mesh <a name="phase-1"></a>

Start with `setup_organic_base.py`. It creates:
- A primitive (sphere/cylinder/cube) at correct mm scale
- Subdivision Surface modifier pre-applied for smooth organic forms
- Clay viewport material
- Correct scene units

### Preset choices

| Preset | Size | Good for |
|--------|------|----------|
| `sphere` | 80mm | Generic organic props, blobs, creatures |
| `cylinder` | 40mm × 100mm | Grip handles, vases, tubes |
| `cube` | 60mm | Chunky animals, square characters |
| `human_head_rough` | 200mm | Portrait busts, face studies |

Edit `PRESET = "sphere"` at the top of the script to change it.

### After setup_organic_base.py runs

Your scene is ready for Sculpt Mode. The base mesh is smooth (subdivision applied) and at print scale.

---

## 4. Phase 2 — Sculpt mode basics <a name="phase-2"></a>

**Enter Sculpt Mode:** `Ctrl+Tab` > Sculpt (or dropdown in top left)

### Essential brushes (memorize these 5)

| Shortcut | Brush | Use it for |
|----------|-------|-----------|
| `X` | **Draw** | Add or remove volume; basic sculpting |
| `G` | **Grab** | Move large areas; reposition features |
| `C` | **Crease** | Sharp edges, folds, facial features |
| `I` | **Inflate** | Swell outward; lips, bulges, cheeks |
| `Shift+draw` | **Smooth** | Blend and soften (works with any brush) |

### Dyntopo — enable for organic work

`N` > Dyntopo > toggle **Enable Dyntopo**

- **Detail Size:** 12px for rough blocking, 6px for detail work, 3px for fine features
- Dyntopo adds/removes mesh polygons as you sculpt so you're never fighting the mesh

### Workflow rhythm

1. **Rough block** — big strokes, large radius, low detail size (12px)
2. **Shape** — medium strokes, define major planes and features
3. **Smooth constantly** — `Shift+draw` after every 3–4 sculpt strokes
4. **Detail** — small radius, high detail size (4–6px)

---

## 5. Phase 3 — Iterative adjustment loop <a name="phase-3"></a>

This is where `iterate_organic.py` helps. Describe your change in plain English, run the script, follow the instructions.

### Format for adjustment descriptions

```
"The [feature] is [problem] — [fix description]"
```

**Examples:**

```
"The jaw is too narrow — widen it by 15% and smooth the transition to the neck"
"The nose bridge sits too high — lower it by 10mm"
"The shoulder is too sharp — smooth and blend into the upper arm"
"The ear needs more detail — sharpen the outer rim"
```

### Save versions as you go

Use `File > Save Copy` to create versioned saves:
- `model_v1_rough.blend`
- `model_v2_jaw_adjusted.blend`
- `model_v3_detail.blend`

This is your version history — there's no undo past a session.

### Reference photos

`N` > View > Background Images — add reference photos to sculpt against.

---

## 6. Phase 4 — 3D print prep <a name="phase-4"></a>

Run `export_for_print.py` — it handles most of this automatically. But here's what it checks:

### Apply all transforms first

```
Object Mode > Object > Apply > All Transforms (Ctrl+A > All Transforms)
```

**Why:** If you scaled in Object Mode without applying, the STL will export at wrong size.

### Check non-manifold geometry

```
Edit Mode > Select All (A) > Select > Select All by Trait > Non-Manifold
```

Non-manifold = holes, disconnected faces, zero-area faces. Slicers can't print these.

**Fix:**
```
Mesh > Clean Up > Merge by Distance (0.01mm threshold)
Mesh > Clean Up > Fill Holes
```

### Wall thickness

```
3D Print Toolbox (N panel) > Check All > Wall Thickness
```

Minimum for PLA/PETG: **1.2mm**. Thin walls may fail to print.

### Overhangs

Overhangs > 45° need supports. The toolbox shows them in red.

---

## 7. Phase 5 — Export and slice <a name="phase-5"></a>

### Export STL

```
File > Export > STL

Settings:
  ✅ Selection Only (if exporting one object)
  ✅ Apply Modifiers
  Scale: 1000 (converts Blender's 0.001 scale back to mm)
  Forward: -Z
  Up: Y
```

Or run `export_for_print.py` — it handles all of this.

### Slicer tips

**Layer height:**
- 0.2mm — standard, good for organic shapes
- 0.1mm — fine detail (faces, textures)
- 0.3mm — fast draft print

**Supports:**
- Use "Tree" or "Organic" supports for organic models
- Minimize overhang areas when orienting your model

**Infill:**
- 15–20% gyroid for display pieces
- 30–40% for handled objects

---

## 8. Keyboard shortcut reference <a name="shortcuts"></a>

### Object Mode
| Key | Action |
|-----|--------|
| `G` | Grab (move) |
| `S` | Scale |
| `R` | Rotate |
| `X/Y/Z` after G/S/R | Constrain to axis |
| `Ctrl+A` | Apply transforms |
| `Shift+A` | Add object |

### Sculpt Mode
| Key | Action |
|-----|--------|
| `X` | Draw brush |
| `G` | Grab brush |
| `C` | Crease brush |
| `I` | Inflate brush |
| `Shift+draw` | Smooth brush |
| `F` | Resize brush radius |
| `Shift+F` | Change brush strength |
| `Ctrl+Z` | Undo |
| `N` | Toggle N panel (settings) |

### Edit Mode
| Key | Action |
|-----|--------|
| `A` | Select all |
| `Alt+click` | Select loop |
| `Ctrl+R` | Loop cut |
| `S Z 0 Enter` | Flatten selection to Z plane |

---

## 9. Troubleshooting <a name="troubleshooting"></a>

| Problem | Fix |
|---------|-----|
| Model prints at wrong size | Apply transforms (Ctrl+A > All) before export |
| STL file is huge | Add Decimate modifier (ratio 0.5) to reduce poly count |
| Print keeps failing at thin areas | Check wall thickness with 3D Print Toolbox (min 1.2mm) |
| Model looks lumpy | Too high Dyntopo detail — smooth more; reduce strength |
| Can't sculpt smoothly | Dyntopo off, or mesh too low-poly — enable Dyntopo or add Subdivision |
| Mirror not working | X Mirror must be enabled in Sculpt Mode header, not just Object Mode |
| Supports everywhere | Rotate model to minimize overhangs; aim for flat base on build plate |

---

## Connecting to the modelgen CLI

The `modelgen` CLI in `../cli/` can generate Blender Python scripts:

```bash
cd ../cli
modelgen --prompt "blender script: setup a 150mm sphere base with 4 levels subdivision for sculpting a character head"
modelgen --prompt "blender script: export active object as STL to /tmp/my_model.stl"
```

This bridges text prompts → Blender automation for the parametric parts of organic workflows.

---

*Tutorial by Friday — part of the 3D Model Generation pipeline*
