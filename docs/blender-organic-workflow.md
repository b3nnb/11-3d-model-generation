# Blender Organic Model Workflow
## Iterative Adjustment Instructions for Organic 3D Models

This guide covers the iterative workflow for creating organic/artistic models in Blender — the complement to the parametric OpenSCAD path.

---

## When to Use Blender vs OpenSCAD

| | OpenSCAD | Blender |
|---|---|---|
| **Best for** | Mechanical parts, boxes, jigs, brackets | Characters, organic shapes, artistic props |
| **Workflow** | Edit parameters → render | Sculpt → adjust → export |
| **Repeatability** | Exact — change a number, get a new model | Iterative — describe changes, apply manually |
| **Output** | Precise dimensions | Artistic fidelity |

---

## Blender Setup for 3D Printing

### Required Addons
- **3D Print Toolbox** (bundled with Blender) — `Edit > Preferences > Add-ons > 3D Print Toolbox`
- **Mesh: 3D-Print Toolbox** enables: wall thickness check, overhangs, non-manifold edges

### Recommended Units
```
Scene Properties > Units:
  Unit System: Metric
  Length: Millimeters
  Scale: 0.001
```

### Export Settings for STL
```
File > Export > STL:
  ✅ Selection Only (if exporting one object)
  ✅ Apply Modifiers
  Scale: 1.0
  Forward: -Z
  Up: Y
```

---

## Organic Model Workflow (Iterative)

### Phase 1: Block Out
1. Start with a primitive (sphere, cylinder, cube) sized appropriately
2. Apply **Subdivision Surface** modifier (2–3 levels) for smooth base
3. Use **Proportional Editing** (O key) for organic shaping
4. Keep geometry low at this stage — subdivide later

### Phase 2: Sculpt Mode Refinement
1. Switch to Sculpt mode
2. **Brushes for organic work:**
   - `Draw` (X) — add volume
   - `Smooth` (Shift+draw) — smooth transitions
   - `Grab` (G) — move large areas
   - `Crease` (C) — add sharp edges/folds
   - `Inflate` (I) — swell areas
3. Enable **Dyntopo** (dynamic topology) for adaptive detail during sculpting

### Phase 3: Adjustment Instructions Format

When iterating on a model with AI assistance, use this format:

```
Current state: [brief description or photo]
Change: [specific adjustment needed]
Reference: [optional photo or sketch]
```

**Example instructions:**
- "The nose bridge is too narrow — widen it by ~20% and smooth the transition to the forehead"
- "The ear sits too high relative to the eye line — move it down about 15mm"  
- "The jaw angle is too soft — sharpen the jawline definition"
- "The shoulder connection is too abrupt — blend the deltoid into the upper arm more gradually"

### Phase 4: 3D Print Prep
1. **Check thickness**: 3D Print Toolbox > Thickness (set minimum to 1.2mm for typical prints)
2. **Fix non-manifold**: Select All → Mesh > Clean Up > Merge by Distance (0.01mm threshold)
3. **Check overhangs**: 3D Print Toolbox > Overhangs (>45° shown in red)
4. **Hollow if needed**: Solidify modifier (1.5–3mm) for large objects, saves material
5. **Add flat base**: Cut flat with a boolean intersect + cube if model needs to stand

### Phase 5: Export
1. **Scale check**: Measure in Edit mode (N panel > Item > Dimensions)
2. **Apply Scale**: Object > Apply > Scale (important before STL export)
3. **Export STL**: File > Export > STL with settings above

---

## Common Organic Adjustments Cheatsheet

| Issue | Solution |
|-------|----------|
| Surface too lumpy | Smooth brush, or Laplacian Smooth modifier |
| Feature too sharp | Crease brush in Smooth mode, or reduce crease weight |
| Proportions wrong | Grab brush in Sculpt, or Scale tool in Object mode |
| Overhangs everywhere | Rotate model for best print orientation; add supports in slicer |
| Thin walls failing | Solidify modifier (1.5mm min for PLA/PETG) |
| Non-manifold mesh | Mesh > Clean Up > Fill Holes + Merge by Distance |
| STL too large | Decimate modifier (0.5 ratio keeps most detail, halves file size) |

---

## Sample Organic Models to Build

These are good exercises that grow in complexity:

1. **Rounded teardrop hook** — minimal organic shape, 3D-printable
2. **Grip handle** — ergonomic cylinder with finger grooves
3. **Bird / animal silhouette** — low-poly display piece
4. **Abstract wall art tile** — flat-ish organic relief, CNC friendly
5. **Character bust** — face/head with iterative sculpt refinement

---

## Connecting Blender to the modelgen Workflow

The `modelgen` CLI handles OpenSCAD generation. For Blender:

```bash
# Generate a Blender Python script for organic base mesh setup
modelgen --prompt "blender python script: create a subdivision surface sphere 
  60mm diameter with 3 levels subdivision and apply 3D print toolbox"
```

The CLI can generate Blender Python (bpy) scripts that auto-setup base meshes, modifiers, and export settings. This bridges the parametric CLI to organic workflows.
