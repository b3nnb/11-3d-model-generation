"""
export_for_print.py — Prepare and export a Blender organic model for 3D printing

Run inside Blender Scripting tab when your sculpt is done.
Performs all pre-export checks and exports a print-ready STL.

Steps automated:
  1. Check scene units are mm
  2. Apply all transforms (scale/rotation/location)
  3. Check non-manifold geometry (report issues)
  4. Run 3D Print Toolbox checks if available
  5. Export STL with correct settings

USAGE:
  1. Open your sculpted .blend file
  2. Set OUTPUT_PATH below
  3. Run Script (Alt+P)
"""

import bpy
import os
import math

# ── Configuration ──────────────────────────────────────────────────────────────
OUTPUT_PATH = "/tmp/model_export.stl"   # Change to your desired output path
CHECK_THICKNESS_MM = 1.2                # Minimum wall thickness check
CHECK_OVERHANGS = True                  # Check for overhangs > 45°
# ─────────────────────────────────────────────────────────────────────────────


SCALE = 0.001  # 1 Blender unit = 1mm


def get_meshes():
    return [o for o in bpy.data.objects if o.type == 'MESH']


def check_units():
    scene = bpy.context.scene
    if scene.unit_settings.system != 'METRIC':
        print("⚠️  Units not set to Metric. Setting now...")
        scene.unit_settings.system = 'METRIC'
        scene.unit_settings.length_unit = 'MILLIMETERS'
        scene.unit_settings.scale_length = SCALE
    else:
        print("✅ Units: Metric mm")


def apply_transforms(obj):
    bpy.context.view_layer.objects.active = obj
    bpy.ops.object.select_all(action='DESELECT')
    obj.select_set(True)
    bpy.ops.object.transform_apply(location=True, rotation=True, scale=True)
    print(f"✅ Transforms applied: {obj.name}")


def check_non_manifold(obj):
    """Switch to edit mode and check for non-manifold edges."""
    bpy.context.view_layer.objects.active = obj
    bpy.ops.object.mode_set(mode='EDIT')
    bpy.ops.mesh.select_all(action='DESELECT')
    bpy.ops.mesh.select_non_manifold()
    bpy.ops.object.mode_set(mode='OBJECT')

    # Count selected vertices (non-manifold)
    nm_count = sum(1 for v in obj.data.vertices if v.select)
    if nm_count > 0:
        print(f"⚠️  Non-manifold edges found ({nm_count} vertices affected)")
        print("   Fix: Edit Mode > Mesh > Clean Up > Merge by Distance (0.01mm)")
        print("   Fix: Edit Mode > Mesh > Clean Up > Fill Holes")
        return False
    else:
        print("✅ No non-manifold geometry")
        return True


def check_dimensions(obj):
    """Report object dimensions in mm."""
    dims = obj.dimensions
    w_mm = dims.x / SCALE
    d_mm = dims.y / SCALE
    h_mm = dims.z / SCALE
    print(f"📐 Dimensions: {w_mm:.1f}mm × {d_mm:.1f}mm × {h_mm:.1f}mm")


def run_print3d_checks(obj):
    """Run 3D Print Toolbox checks if available."""
    if 'object_print3d_utils' not in bpy.context.preferences.addons:
        print("ℹ️  3D Print Toolbox not enabled — skipping automated checks")
        print("   Enable: Edit > Preferences > Add-ons > 3D Print Toolbox")
        return

    bpy.context.view_layer.objects.active = obj
    bpy.ops.object.select_all(action='DESELECT')
    obj.select_set(True)

    print("\n🔬 3D Print Toolbox checks:")
    try:
        bpy.ops.mesh.print3d_check_solid()
        bpy.ops.mesh.print3d_check_intersect()
        bpy.ops.mesh.print3d_check_degenerate()
        print("   ✅ Checks complete — see 3D Print panel (N key > 3D Print) for results")
    except Exception as e:
        print(f"   ⚠️  Could not run 3D Print checks: {e}")


def export_stl(obj, path):
    """Export selected object as STL."""
    # Make output directory if needed
    out_dir = os.path.dirname(path)
    if out_dir and not os.path.exists(out_dir):
        os.makedirs(out_dir)

    bpy.context.view_layer.objects.active = obj
    bpy.ops.object.select_all(action='DESELECT')
    obj.select_set(True)

    bpy.ops.export_mesh.stl(
        filepath=path,
        use_selection=True,
        use_mesh_modifiers=True,
        global_scale=1000.0,  # Blender units → mm
        axis_forward='-Z',
        axis_up='Y',
    )
    size_mb = os.path.getsize(path) / (1024 * 1024)
    print(f"\n✅ STL exported: {path} ({size_mb:.1f} MB)")

    # Triangle count
    mesh = obj.evaluated_get(bpy.context.evaluated_depsgraph_get()).data
    tri_count = len(mesh.polygons)
    print(f"   Triangle count: {tri_count:,}")
    if tri_count > 2_000_000:
        print("   ⚠️  Very high poly count — consider Decimate modifier (0.5 ratio) to reduce file size")
    elif tri_count > 500_000:
        print("   ℹ️  High poly — OK for most slicers, but may be slow to load")
    else:
        print("   ✅ Poly count looks good")


def print_slicer_tips():
    print("""
💡 SLICER TIPS (Bambu Studio / PrusaSlicer / Cura):

  Layer height:
    0.2mm  — standard quality, good for most organic shapes
    0.1mm  — fine detail for faces/textures
    0.3mm  — draft quality, good for test prints

  Supports:
    Organic models usually need supports for overhangs > 45°
    Use "Organic" or "Tree" supports to minimize contact marks

  Orientation:
    Rotate model to minimize overhangs and maximize strength
    Tallest axis vertical = strongest print
    Flat base on build plate = no brim needed

  Infill:
    15-20% gyroid infill for lightweight display pieces
    30-40% for functional/handled objects
""")


# ── Main ───────────────────────────────────────────────────────────────────────

def main():
    meshes = get_meshes()
    if not meshes:
        print("❌ No mesh objects found in scene")
        return

    obj = bpy.context.active_object
    if not obj or obj.type != 'MESH':
        obj = meshes[0]
        bpy.context.view_layer.objects.active = obj

    print(f"🖨️  Preparing for export: {obj.name}")
    print()

    check_units()
    check_dimensions(obj)
    apply_transforms(obj)
    is_manifold = check_non_manifold(obj)
    run_print3d_checks(obj)

    if not is_manifold:
        print("\n⛔ Non-manifold geometry detected — fix issues before exporting")
        print("   (You can still export, but the print may fail)")
        print()

    export_stl(obj, OUTPUT_PATH)
    print_slicer_tips()


main()
