"""
setup_organic_base.py — Blender Python script for organic model base mesh setup

Run inside Blender via the Scripting tab, or:
  blender --background --python setup_organic_base.py -- --type sphere --output /tmp/organic_base.blend

Creates a base mesh ready for organic sculpting with:
- Correct metric units (mm) for 3D printing
- Subdivision Surface modifier for smooth organic forms
- 3D Print Toolbox addon check
- Named materials + sensible viewport colours

Usage inside Blender Scripting tab:
  1. Open this file
  2. Edit PRESET at the top if needed
  3. Run Script (Alt+P)
"""

import bpy
import sys
import os
import math


# ── Configuration ──────────────────────────────────────────────────────────────

PRESET = "sphere"  # sphere | cylinder | cube | human_head_rough

OUTPUT_PATH = ""   # optional: save .blend here after setup. "" = don't save.

# Dimensions (mm) — Blender scene scale is 0.001 (1 unit = 1mm)
PRESETS = {
    "sphere": {
        "description": "Generic organic blob — good starting point for props, characters",
        "primitive": "sphere",
        "size_mm": 80,        # diameter
        "subdiv_levels": 3,
        "sculpt_multires": 2,
    },
    "cylinder": {
        "description": "Grip handle / organic tube base",
        "primitive": "cylinder",
        "size_mm": 40,        # diameter; height = size_mm * 2.5
        "subdiv_levels": 3,
        "sculpt_multires": 2,
    },
    "cube": {
        "description": "Rounded organic cube — good for animals, chunky characters",
        "primitive": "cube",
        "size_mm": 60,
        "subdiv_levels": 3,
        "sculpt_multires": 2,
    },
    "human_head_rough": {
        "description": "Rough head proportion base — sphere with jaw-line hint",
        "primitive": "sphere",
        "size_mm": 200,       # ~20cm = close to real head scale
        "subdiv_levels": 4,
        "sculpt_multires": 3,
    },
}

# ── Unit helpers ───────────────────────────────────────────────────────────────

SCALE = 0.001  # 1 Blender unit = 1mm when scene scale is 0.001


def mm(value):
    """Convert mm to Blender units."""
    return value * SCALE


# ── Setup ──────────────────────────────────────────────────────────────────────

def configure_scene():
    """Set units to mm and enable 3D Print Toolbox."""
    scene = bpy.context.scene

    # Units
    scene.unit_settings.system = 'METRIC'
    scene.unit_settings.length_unit = 'MILLIMETERS'
    scene.unit_settings.scale_length = SCALE

    # Enable 3D Print Toolbox addon if not already enabled
    if not bpy.context.preferences.addons.get("object_print3d_utils"):
        try:
            bpy.ops.preferences.addon_enable(module="object_print3d_utils")
            print("✅ 3D Print Toolbox enabled")
        except Exception as e:
            print(f"⚠️  Could not enable 3D Print Toolbox: {e}")
            print("   Enable manually: Edit > Preferences > Add-ons > 3D Print Toolbox")

    # Viewport clipping for small objects
    for area in bpy.context.screen.areas:
        if area.type == 'VIEW_3D':
            for space in area.spaces:
                if space.type == 'VIEW_3D':
                    space.clip_start = mm(0.1)
                    space.clip_end = mm(5000)

    print(f"✅ Scene configured for mm printing (scale: {SCALE})")


def clear_scene():
    """Remove default objects."""
    bpy.ops.object.select_all(action='SELECT')
    bpy.ops.object.delete()


def create_base_mesh(preset_name):
    """Create the base primitive scaled to mm dimensions."""
    cfg = PRESETS[preset_name]
    prim = cfg["primitive"]
    size = mm(cfg["size_mm"])

    if prim == "sphere":
        bpy.ops.mesh.primitive_uv_sphere_add(
            segments=32, ring_count=16,
            radius=size / 2,
            location=(0, 0, 0)
        )
    elif prim == "cylinder":
        height = mm(cfg["size_mm"] * 2.5)
        bpy.ops.mesh.primitive_cylinder_add(
            vertices=32,
            radius=size / 2,
            depth=height,
            location=(0, 0, height / 2)
        )
    elif prim == "cube":
        bpy.ops.mesh.primitive_cube_add(
            size=size,
            location=(0, 0, size / 2)
        )

    obj = bpy.context.active_object
    obj.name = f"organic_{preset_name}_base"
    print(f"✅ Created {prim} — {cfg['size_mm']}mm")
    return obj


def add_subdivision(obj, levels):
    """Add a Subdivision Surface modifier for smooth sculpting."""
    bpy.context.view_layer.objects.active = obj
    mod = obj.modifiers.new(name="Subdivision", type='SUBSURF')
    mod.levels = levels
    mod.render_levels = levels
    mod.subdivision_type = 'CATMULL_CLARK'
    print(f"✅ Subdivision Surface — {levels} levels")
    return mod


def add_material(obj, preset_name):
    """Add a simple clay-like material for sculpt viewport."""
    mat = bpy.data.materials.new(name=f"clay_{preset_name}")
    mat.use_nodes = True
    nodes = mat.node_tree.nodes
    bsdf = nodes.get("Principled BSDF")
    if bsdf:
        # Warm clay colour
        bsdf.inputs["Base Color"].default_value = (0.85, 0.72, 0.58, 1.0)
        bsdf.inputs["Roughness"].default_value = 0.9
        bsdf.inputs["Specular IOR Level"].default_value = 0.1

    obj.data.materials.append(mat)
    print(f"✅ Clay material applied")


def add_smooth_shading(obj):
    """Apply smooth shading for a nicer sculpt viewport."""
    bpy.context.view_layer.objects.active = obj
    bpy.ops.object.shade_smooth()


def print_next_steps(preset_name):
    cfg = PRESETS[preset_name]
    size = cfg["size_mm"]
    print()
    print("=" * 60)
    print(f"  Organic base ready — {preset_name} ({size}mm)")
    print(f"  {cfg['description']}")
    print("=" * 60)
    print()
    print("NEXT STEPS:")
    print("  1. Switch to Sculpt Mode (Ctrl+Tab > Sculpt)")
    print("  2. Enable Dyntopo (N panel > Dyntopo > toggle ON)")
    print("     — Set detail size to 12px for rough blocking")
    print("     — Reduce to 6px for detail work")
    print()
    print("  3. Rough blocking brushes:")
    print("     Draw (X) — add volume")
    print("     Grab (G)  — move large areas")
    print("     Smooth (Shift+draw) — blend")
    print()
    print("  4. 3D Print check (when done sculpting):")
    print("     Sidebar (N) > 3D Print > Check All")
    print("     Fix non-manifold: Mesh > Clean Up > Merge by Distance (0.01mm)")
    print()
    print("  5. Export:")
    print("     Apply Scale first: Object > Apply > Scale")
    print("     File > Export > STL (Apply Modifiers checked)")
    print()
    print("  TIP: Use modelgen CLI to generate Blender Python scripts:")
    print("       modelgen --prompt 'blender script: [your adjustment description]'")
    print()


# ── Main ───────────────────────────────────────────────────────────────────────

def main():
    preset_name = PRESET

    # Parse command-line args if running from CLI
    argv = sys.argv
    if "--" in argv:
        args = argv[argv.index("--") + 1:]
        for i, a in enumerate(args):
            if a == "--type" and i + 1 < len(args):
                preset_name = args[i + 1]
            if a == "--output" and i + 1 < len(args):
                global OUTPUT_PATH
                OUTPUT_PATH = args[i + 1]

    if preset_name not in PRESETS:
        print(f"Unknown preset '{preset_name}'. Choose: {list(PRESETS.keys())}")
        return

    configure_scene()
    clear_scene()
    obj = create_base_mesh(preset_name)
    add_subdivision(obj, PRESETS[preset_name]["subdiv_levels"])
    add_smooth_shading(obj)
    add_material(obj, preset_name)

    if OUTPUT_PATH:
        bpy.ops.wm.save_as_mainfile(filepath=OUTPUT_PATH)
        print(f"✅ Saved: {OUTPUT_PATH}")

    print_next_steps(preset_name)


main()
