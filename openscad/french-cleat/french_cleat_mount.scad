// ============================================================
// French Cleat System — Mount Plate + Tool Holder
// ============================================================
// Standard French cleat: 45° angle, typically 3/4" (19mm) thick plywood
// Adjust variables below. Works for wall-mounted storage.

// ---- CLEAT PARAMETERS (store these, reuse every mount) ----
cleat_width    = 200;  // Width of the cleat (mm) — your wall strip
cleat_height   = 40;   // Height of the cleat strip (mm)
cleat_angle    = 45;   // Always 45° for French cleat standard
cleat_material = 19;   // Plywood thickness (mm) — 3/4" = 19mm

// ---- MOUNT PARAMETERS ----
mount_width    = 80;   // Width of this specific mount/holder (mm)
mount_depth    = 60;   // How far it sticks out from wall (mm)
mount_height   = 120;  // Total height of mount (mm)
wall_t         = 4;    // Wall thickness of mount (mm)

// ---- CLEAT HOOK (the part that grabs the wall cleat) ----
module cleat_hook() {
    // Hook profile that slides onto 45° wall cleat
    hook_h = cleat_material + 3; // Slight clearance
    translate([0, 0, mount_height - hook_h])
    difference() {
        cube([mount_width, cleat_material + wall_t, hook_h]);
        // Cut 45° hook underside
        translate([0, wall_t, 0])
        rotate([0, 90, 0])
        linear_extrude(mount_width)
        polygon([[0,0],[cleat_material,0],[cleat_material, hook_h],[0, hook_h * 0.5]]);
    }
}

// ---- MOUNT BODY ----
module mount_body() {
    // Back plate
    cube([mount_width, wall_t, mount_height]);
    // Bottom shelf (for resting tools)
    translate([0, wall_t, 0])
        cube([mount_width, mount_depth - wall_t, wall_t]);
    // Side walls
    cube([wall_t, mount_depth, mount_height * 0.4]);
    translate([mount_width - wall_t, 0, 0])
        cube([wall_t, mount_depth, mount_height * 0.4]);
}

// ---- ASSEMBLE ----
mount_body();
cleat_hook();

// ---- WALL CLEAT (for reference / render separately) ----
module wall_cleat() {
    color("SaddleBrown")
    difference() {
        cube([cleat_width, cleat_material, cleat_height]);
        // 45° cut on top
        translate([-1, -0.1, cleat_height - cleat_material])
        rotate([0, 90, 0])
        linear_extrude(cleat_width + 2)
        polygon([[0,0],[cleat_material + 0.1, 0],[cleat_material + 0.1, cleat_material + 0.1]]);
    }
}

// Uncomment to visualize wall cleat:
// translate([0, -cleat_material - 10, 0]) wall_cleat();
