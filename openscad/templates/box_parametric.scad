// ============================================================
// Parametric Box Template — 3D Printable / CNC
// ============================================================
// All dimensions in mm. Adjust variables below.
// Use: openscad -o box.stl box_parametric.scad

// ---- PARAMETERS ----
width  = 80;   // X dimension (mm)
depth  = 60;   // Y dimension (mm)
height = 40;   // Z dimension (mm)
wall   = 3;    // Wall thickness (mm)
base   = 3;    // Base thickness (mm)
fillet = 2;    // Corner fillet radius (0 = sharp)

// Print mode: "print3d" considers overhangs, "cnc" makes flat-bottom design
mode = "print3d"; // "print3d" | "cnc"

// Lid: include a lip for a snap-fit lid?
include_lid_lip = true;
lid_lip_height  = 4;    // Height of the lid lip (mm)
lid_clearance   = 0.3;  // Clearance between lid and box (mm)

// ---- BOX BODY ----
module box_body() {
    difference() {
        // Outer shell
        hull_box(width, depth, height, fillet);
        // Inner cavity
        translate([wall, wall, base])
            hull_box(width - wall*2, depth - wall*2, height - base + 0.1, max(0, fillet - wall));
    }
}

// ---- LID LIP (optional) ----
module lid_lip() {
    if (include_lid_lip) {
        translate([wall + lid_clearance, wall + lid_clearance, height - lid_lip_height])
            difference() {
                hull_box(width - (wall+lid_clearance)*2, depth - (wall+lid_clearance)*2, lid_lip_height + 0.1, max(0, fillet - wall - lid_clearance));
                translate([wall, wall, -0.1])
                    hull_box(width - (wall+lid_clearance)*2 - wall*2, depth - (wall+lid_clearance)*2 - wall*2, lid_lip_height + 0.3, max(0, fillet - wall*2 - lid_clearance));
            }
    }
}

// ---- ROUNDED BOX HELPER ----
module hull_box(w, d, h, r) {
    if (r <= 0) {
        cube([w, d, h]);
    } else {
        hull() {
            for (x = [r, w-r], y = [r, d-r]) {
                translate([x, y, 0]) cylinder(r=r, h=h, $fn=32);
            }
        }
    }
}

// ---- ASSEMBLE ----
box_body();
lid_lip();
