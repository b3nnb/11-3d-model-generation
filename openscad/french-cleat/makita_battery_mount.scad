// ============================================================
// Makita 18V Battery Pack — French Cleat Wall Mount
// ============================================================
// Holds 2× Makita 18V BL18xx batteries on a French cleat wall.
// Batteries slot in from above and are held by the terminal tabs.
//
// French cleat: 19mm (3/4") ply, 45° standard
// Design: print flat, back plate on bed, no supports needed.
// ============================================================

// ---- FRENCH CLEAT (don't change) ----
cleat_material = 19;   // Wall strip thickness (mm)
hook_clearance = 1.5;  // Hook slide gap (mm)
wall_t         = 4;    // Structural wall thickness (mm)

// ---- MAKITA 18V BATTERY DIMENSIONS ----
// BL1830/BL1840/BL1850 — all share the same mounting rail
battery_w      = 75;   // Battery body width (mm)
battery_d      = 55;   // Battery body depth front-to-back (mm)
battery_slot_w = 20;   // Width of the battery retention tab slot (mm)
battery_gap    = 1.0;  // Clearance so battery slides in smoothly (mm)

// ---- HOLDER GEOMETRY ----
num_batteries  = 2;    // Batteries to hold side-by-side
pocket_gap     = 8;    // Gap between battery pockets (mm)
pocket_depth   = 35;   // How deep battery sits in pocket (mm)
pocket_h       = 40;   // Height of each battery pocket (mm)
plate_h        = 95;   // Total mount height (mm)

// ---- COMPUTED ----
bw             = battery_w + battery_gap * 2;
bd             = battery_d + battery_gap;
plate_w        = bw * num_batteries + pocket_gap * (num_batteries - 1) + wall_t * 2;
hook_h         = cleat_material + hook_clearance + wall_t;
hook_start_z   = plate_h - hook_h;

// ============================================================
// MODULES
// ============================================================

module cleat_hook() {
    difference() {
        cube([plate_w, cleat_material + wall_t + hook_clearance, hook_h]);
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

module battery_pocket(x_offset) {
    translate([x_offset, 0, 0])
    difference() {
        cube([bw, bd + wall_t, pocket_h]);
        // Open pocket (open at top for battery insertion)
        translate([battery_gap, wall_t, -0.1])
            cube([battery_w, battery_d, pocket_h + 0.2]);
    }
}

module holder_body() {
    union() {
        // Back plate — full width, full height
        cube([plate_w, wall_t, plate_h]);

        // Battery pockets
        for (i = [0 : num_batteries - 1]) {
            xpos = wall_t + i * (bw + pocket_gap);
            battery_pocket(xpos);
        }

        // Divider between batteries (and left/right end walls already in back plate)
        if (num_batteries > 1) {
            divider_x = wall_t + bw;
            translate([divider_x, 0, 0])
                cube([pocket_gap, bd + wall_t, pocket_h]);
        }
    }
}

// ============================================================
// ASSEMBLY
// ============================================================
union() {
    holder_body();
    translate([0, 0, hook_start_z])
        cleat_hook();
}

// ============================================================
// PRINT SETTINGS
// ============================================================
// Orientation : back plate flat on print bed (no supports)
// Material    : PLA+ or PETG
// Infill      : 25-30%
// Layer height: 0.2mm
// Batteries slide in from top, terminal tabs lock at pocket bottom
// ============================================================
