// ============================================================
// Makita Cordless Drill — French Cleat Wall Mount
// ============================================================
// Stores a Makita 18V compact drill by the handle. Drill hangs
// trigger-side down, grip resting on a contoured cradle.
//
// French cleat: 19mm (3/4") plywood, 45° angle
// Designed for 3D print (PLA or PETG). No supports needed.
// Print orientation: back plate flat on bed.
//
// Default values match a standard Makita DHP/DDF 18V compact drill.
// Adjust the handle dimensions section if your drill differs.
// ============================================================

// ---- FRENCH CLEAT SYSTEM (don't change) ----
cleat_material = 19;   // Wall cleat plywood thickness (mm)
hook_clearance = 1.5;  // Gap so hook slides onto cleat easily
wall_t         = 4;    // Structural wall/plate thickness (mm)

// ---- DRILL HANDLE DIMENSIONS (*** measure with calipers to confirm) ----
// These match typical Makita 18V compact drill handle geometry
handle_w       = 52;   // Handle grip width at narrowest grab zone (mm)
handle_d       = 37;   // Handle depth front-to-back (mm)

// ---- CRADLE / MOUNT ----
plate_w        = 90;   // Width of the backing plate (mm)
plate_h        = 130;  // Total mount height — cradle bottom to hook top (mm)
cradle_h       = 45;   // Height of the handle pocket section (mm)
cradle_depth   = 50;   // Depth (front-to-back) of the cradle pocket (mm)
cradle_gap     = 1.0;  // Clearance on each side of the handle for easy insert (mm)
lip_h          = 8;    // Retention lip height — keeps drill from sliding up (mm)
lip_inset      = 3;    // How far the lip narrows the opening (mm)

// ---- COMPUTED ----
pocket_w       = handle_w + cradle_gap * 2;
pocket_d       = handle_d + cradle_gap;
hook_h         = cleat_material + hook_clearance + wall_t;
hook_start_z   = plate_h - hook_h;

// ============================================================
// MODULES
// ============================================================

// French cleat hook — slides over the 45° wall cleat strip
module cleat_hook() {
    difference() {
        cube([plate_w, cleat_material + wall_t + hook_clearance, hook_h]);
        // Angled undercut so hook seats on the 45° cleat face
        translate([-0.1, wall_t, -0.1])
        rotate([0, 90, 0])
        linear_extrude(plate_w + 0.2)
        polygon([
            [0, 0],
            [cleat_material + hook_clearance, 0],
            [cleat_material + hook_clearance, hook_h + 0.2],
            [0, hook_h * 0.45]
        ]);
    }
}

// Main mount body — back plate + cradle pocket + retention lips
module mount_body() {
    difference() {
        union() {
            // Back plate — full height, wall_t deep
            cube([plate_w, wall_t, plate_h]);

            // Cradle block — where the drill handle sits
            cube([plate_w, cradle_depth + wall_t, cradle_h]);
        }

        // Handle pocket — open at front and top for easy insert/remove
        // Slightly wider at top so drill drops in, narrows at bottom to hold it
        translate([(plate_w - pocket_w) / 2, wall_t, -0.1])
            cube([pocket_w, pocket_d, cradle_h - lip_h + 0.2]);

        // Widen the top of the pocket (funnel opening) for easy insertion
        translate([(plate_w - pocket_w) / 2 - 4, wall_t, cradle_h - lip_h - 5])
            cube([pocket_w + 8, pocket_d, cradle_h]);
    }

    // Retention lips — narrow the pocket slightly near the bottom
    // so the drill handle is gripped and can't fall out
    lip_x_left  = (plate_w - pocket_w) / 2;
    lip_x_right = (plate_w + pocket_w) / 2;
    // Left lip
    translate([lip_x_left - lip_inset, wall_t, 0])
        cube([lip_inset, min(pocket_d - 5, 25), lip_h]);
    // Right lip
    translate([lip_x_right, wall_t, 0])
        cube([lip_inset, min(pocket_d - 5, 25), lip_h]);
}

// Gussets — reinforce the cradle-to-backplate junction
module gussets() {
    gusset_t = 3;
    gusset_d = 25;
    gusset_h = cradle_h * 0.6;
    // Place gussets near outer edges, clear of pocket
    for (gx = [
        (plate_w - pocket_w) / 2 - gusset_t - 3,
        (plate_w + pocket_w) / 2 + 3
    ]) {
        translate([gx, wall_t, 0])
        linear_extrude(gusset_h)
        polygon([
            [0, 0],
            [gusset_t, 0],
            [gusset_t, gusset_d],
            [0, gusset_d * 0.5]
        ]);
    }
}

// ============================================================
// ASSEMBLY — unioned into one solid printable part
// ============================================================
union() {
    mount_body();
    gussets();
    translate([0, 0, hook_start_z])
        cleat_hook();
}

// ============================================================
// PRINT SETTINGS
// ============================================================
// Orientation : back plate flat on print bed (no supports needed)
// Material    : PETG or PLA+  (PETG recommended for garage/workshop)
// Infill      : 30–40%
// Perimeters  : 3 walls minimum (especially around hook section)
// Layer height: 0.2mm
// ============================================================
