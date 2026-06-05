// ============================================================
// French Cleat — Blank Holder Strip (Hook Rail)
// ============================================================
// A flat mounting strip that sits on French cleat wall strips.
// Add your own hooks, bins, or attachments in OpenSCAD, or
// use it as a baseplate for custom tool holders.
//
// Also generates a simple J-hook variant for hanging items
// by a hole, loop, or trigger guard.
//
// French cleat: 19mm (3/4") ply, 45° standard
// ============================================================

// ---- FRENCH CLEAT (don't change) ----
cleat_material  = 19;   // Wall strip thickness (mm)
hook_clearance  = 1.5;  // Hook slide gap (mm)
wall_t          = 4;    // Structural wall thickness (mm)

// ---- STRIP DIMENSIONS ----
strip_w         = 150;  // Width of the holder strip (mm) — covers one or more cleat slots
strip_h         = 80;   // Height of the strip (mm)
strip_d         = 20;   // Depth from wall to front face (mm)

// ---- ATTACHMENT POINTS (screw holes for adding custom holders) ----
add_screw_holes = true; // true = add M4 counterbore holes for attaching hooks/bins
screw_dia       = 4.5;  // M4 clearance hole (mm)
cbore_dia       = 8.0;  // M4 counterbore diameter (mm)
cbore_depth     = 4.0;  // Counterbore depth (mm)
screw_rows      = 2;    // Rows of screw holes
screw_cols      = 3;    // Columns of screw holes

// ---- COMPUTED ----
hook_h          = cleat_material + hook_clearance + wall_t;
hook_start_z    = strip_h - hook_h;
screw_margin_x  = strip_w / (screw_cols + 1);
screw_margin_z  = (strip_h - hook_h - 10) / (screw_rows + 1);

// ============================================================
// MODULES
// ============================================================

module cleat_hook() {
    difference() {
        cube([strip_w, cleat_material + wall_t + hook_clearance, hook_h]);
        translate([-0.1, wall_t, -0.1])
        rotate([0, 90, 0])
        linear_extrude(strip_w + 0.2)
        polygon([
            [0, 0],
            [cleat_material + hook_clearance, 0],
            [cleat_material + hook_clearance, hook_h + 0.2],
            [0, hook_h * 0.45]
        ]);
    }
}

module strip_body() {
    difference() {
        union() {
            // Back plate
            cube([strip_w, wall_t, strip_h]);
            // Shelf — flat surface to attach bins/hooks to
            translate([0, wall_t, 0])
                cube([strip_w, strip_d - wall_t, wall_t]);
            // Bottom reinforcement rib
            translate([0, wall_t, 0])
                cube([strip_w, wall_t, strip_h * 0.5]);
        }

        // Screw holes for attaching add-on bins/hooks
        if (add_screw_holes) {
            for (col = [1 : screw_cols]) {
                for (row = [1 : screw_rows]) {
                    sx = screw_margin_x * col;
                    sz = 5 + screw_margin_z * row;
                    // Through hole
                    translate([sx, -0.1, sz])
                    rotate([-90, 0, 0])
                        cylinder(d=screw_dia, h=wall_t + 0.2, $fn=16);
                    // Counterbore (from back)
                    translate([sx, -0.1, sz])
                    rotate([-90, 0, 0])
                        cylinder(d=cbore_dia, h=cbore_depth, $fn=16);
                }
            }
        }
    }
}

// ============================================================
// ASSEMBLY
// ============================================================
union() {
    strip_body();
    translate([0, 0, hook_start_z])
        cleat_hook();
}

// ============================================================
// PRINT SETTINGS
// ============================================================
// Orientation : back plate flat on bed
// Material    : PLA+ or PETG
// Infill      : 20-25%  (mostly flat plate, doesn't need much)
// Supports    : none
// Use M4×12 screws + hex nuts to attach bins/hooks to the strip
// ============================================================
