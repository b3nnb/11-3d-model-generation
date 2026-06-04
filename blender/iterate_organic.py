"""
iterate_organic.py — Apply iterative adjustment instructions to a Blender scene

Run inside Blender Scripting tab after opening your .blend file.
Paste your adjustment instructions at the top of this script and run.

This script provides helpers for common organic sculpt adjustments that can
be scripted (scale, move, mirror, add detail brushes, export).

For adjustments that require real sculpting (adding volume, refining features),
it prints step-by-step instructions using the right brushes/tools.

USAGE:
  1. Open your .blend file in Blender
  2. Open this script in the Scripting tab
  3. Set ADJUSTMENT below to describe your change
  4. Run Script (Alt+P)

The script will either:
  a) Apply the change automatically (for scriptable transforms)
  b) Print exact Blender instructions to follow manually
"""

import bpy
import math

# ── Your adjustment instruction here ──────────────────────────────────────────
ADJUSTMENT = """
The jaw area is too narrow — widen it by 15% and smooth the transition to the neck
"""
# ─────────────────────────────────────────────────────────────────────────────


def get_active_organic():
    """Return the active object, or the first mesh in scene."""
    obj = bpy.context.active_object
    if obj and obj.type == 'MESH':
        return obj
    for o in bpy.data.objects:
        if o.type == 'MESH':
            return o
    return None


def apply_scale_uniform(obj, factor, axis=None):
    """Scale object uniformly or on single axis."""
    if axis == 'X':
        obj.scale[0] *= factor
    elif axis == 'Y':
        obj.scale[1] *= factor
    elif axis == 'Z':
        obj.scale[2] *= factor
    else:
        obj.scale[0] *= factor
        obj.scale[1] *= factor
        obj.scale[2] *= factor
    bpy.ops.object.transform_apply(scale=True)
    print(f"✅ Scaled {'uniformly' if not axis else f'on {axis}'} by {factor:.2f}x")


def add_mirror_modifier(obj, axis='X'):
    """Add a Mirror modifier — useful for symmetric organic models."""
    idx = {'X': 0, 'Y': 1, 'Z': 2}[axis]
    mod = obj.modifiers.new(name=f"Mirror_{axis}", type='MIRROR')
    mod.use_axis[idx] = True
    mod.use_clip = True
    print(f"✅ Mirror modifier added on {axis} axis (clip on)")


def export_stl(obj, path):
    """Export the current object as STL."""
    bpy.context.view_layer.objects.active = obj
    bpy.ops.object.select_all(action='DESELECT')
    obj.select_set(True)
    # Apply scale before export
    bpy.ops.object.transform_apply(scale=True)
    bpy.ops.export_mesh.stl(
        filepath=path,
        use_selection=True,
        use_mesh_modifiers=True,
        global_scale=1000.0,  # Convert Blender units back to mm
        axis_forward='-Z',
        axis_up='Y',
    )
    print(f"✅ Exported STL: {path}")


def print_sculpt_instructions(adjustment):
    """Parse a natural language adjustment and print step-by-step Blender instructions."""
    adj = adjustment.lower()

    print()
    print("=" * 60)
    print("  SCULPT ADJUSTMENT INSTRUCTIONS")
    print("  Instruction:", adjustment.strip())
    print("=" * 60)

    # Widen / narrow instructions
    if any(w in adj for w in ['widen', 'wider', 'broaden', 'expand']):
        print("""
📐 To widen an area:
  1. Enter Sculpt Mode (Ctrl+Tab > Sculpt)
  2. Select brush: Grab (shortcut: G)
  3. Set Radius to cover the area you want to move
  4. Click and drag outward on both sides
  5. OR use Scale (S) in Edit Mode with vertices selected

  For symmetrical widening:
  1. Enable X Mirror in the sculpt header (butterfly icon)
  2. Grab brush will mirror automatically
""")

    if any(w in adj for w in ['narrow', 'thinner', 'compress', 'reduce width']):
        print("""
📐 To narrow an area:
  1. Enter Sculpt Mode (Ctrl+Tab > Sculpt)
  2. Select brush: Grab (shortcut: G) with X Mirror enabled
  3. Drag inward from both sides
  4. OR: Edit Mode > select region vertices > Scale on X axis (S, X, value, Enter)
""")

    if any(w in adj for w in ['smooth', 'blend', 'transition']):
        print("""
🌊 To smooth a transition:
  1. Enter Sculpt Mode
  2. Hold Shift + Draw anywhere to use the Smooth brush
  3. Set strength to 0.5 for gradual smoothing
  4. Work along the transition boundary with medium-large radius
  5. Use Relax brush (no default shortcut — find in brush list) for mesh flow
""")

    if any(w in adj for w in ['move', 'shift', 'reposition', 'down', 'up', 'lower', 'raise', 'higher']):
        print("""
↕️  To move/reposition a feature:
  1. Enter Sculpt Mode
  2. Grab brush (G) — set large radius to encompass the feature
  3. Drag to new position
  4. Clean up edges with Smooth (Shift+draw)
  5. For precise movement: Edit Mode > select feature vertices > G to grab
""")

    if any(w in adj for w in ['sharp', 'sharpen', 'crease', 'define', 'angular']):
        print("""
✏️  To sharpen/crease a feature:
  1. Enter Sculpt Mode
  2. Crease brush (C) — drag along the edge you want to sharpen
  3. Set strength to 0.6–0.8
  4. Use small radius for fine creases
  5. Inverse crease (Ctrl+C) to flatten/remove creases
""")

    if any(w in adj for w in ['inflate', 'puff', 'swell', 'volume', 'bulge']):
        print("""
💨 To add volume/inflate an area:
  1. Enter Sculpt Mode
  2. Inflate brush (I) — click and hold to swell outward
  3. Draw brush (X) for more directional volume
  4. Use medium strength (0.3–0.5) to avoid over-inflating
""")

    if any(w in adj for w in ['flatten', 'flat base', 'bottom', 'stand']):
        print("""
📐 To add a flat base (for 3D printing stability):
  1. Object Mode — position model so flat side faces down
  2. Add a cube: Shift+A > Mesh > Cube, size it larger than the base
  3. Move cube to intersect bottom of model
  4. Add Boolean modifier: Object > Modifiers > Boolean > Intersect
  5. OR: Edit Mode > select bottom vertices > S Z 0 Enter (flatten to plane)
""")

    # Generic fallback
    all_keywords = ['widen', 'wider', 'narrow', 'thinner', 'smooth', 'blend',
                    'move', 'shift', 'sharp', 'crease', 'inflate', 'flatten', 'flat']
    if not any(w in adj for w in all_keywords):
        print(f"""
🎨 General sculpting workflow for: "{adjustment.strip()}"

  1. Enter Sculpt Mode (Ctrl+Tab > Sculpt)
  2. Choose appropriate brush:
     Draw (X)           — add/remove volume
     Grab (G)           — move large areas
     Smooth (Shift+any) — soften and blend
     Crease (C)         — add definition/sharp edges
     Inflate (I)        — swell outward
     Flatten (Shift+T)  — flatten to average plane

  3. Enable X Mirror for symmetric models (header butterfly icon)
  4. Start with large radius, refine with smaller
  5. Smooth frequently to keep mesh clean
""")

    print("─" * 60)
    print("After changes: File > Export > STL (Apply Modifiers: on)")
    print()


# ── Main ───────────────────────────────────────────────────────────────────────

def main():
    obj = get_active_organic()
    if not obj:
        print("❌ No mesh object found. Open your .blend file first.")
        return

    print(f"🔧 Working on: {obj.name}")
    print_sculpt_instructions(ADJUSTMENT)


main()
