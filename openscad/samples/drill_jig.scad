// ============================================================
// Parametric Jig / Drill Guide
// ============================================================
// Drilling guide for consistent hole placement.
// Clamps or presses against workpiece edge.

// ---- PARAMETERS ----
jig_w        = 60;    // Jig body width (mm)
jig_d        = 50;    // Jig body depth (mm)
jig_h        = 20;    // Jig body height (mm)
wall         = 4;     // Wall thickness

// Guide holes
hole_dia     = [5, 8, 10];   // Drill guide diameters (mm) — one column per size
hole_spacing = 18;           // Spacing between guide hole columns (mm)
hole_y       = 25;           // Y-position of guide holes (depth into workpiece)

// Fence (edge registration)
fence_h      = 15;    // Height of fence that registers against workpiece edge
fence_t      = 4;     // Fence thickness
fence_l      = jig_w; // Fence length

// ---- JIG BODY ----
module jig_body() {
    difference() {
        cube([jig_w, jig_d, jig_h]);
        // Remove material inside (weight saving, not structural)
        translate([wall, wall, wall])
            cube([jig_w - wall*2, jig_d - wall*2 - fence_t, jig_h]);
    }
}

// ---- GUIDE BUSHINGS ----
module guide_holes() {
    start_x = (jig_w - (len(hole_dia) - 1) * hole_spacing) / 2;
    for (i = [0 : len(hole_dia) - 1]) {
        x = start_x + i * hole_spacing;
        translate([x, hole_y, -0.1]) {
            // Guide hole through jig
            cylinder(d=hole_dia[i] + 0.2, h=jig_h + 0.2, $fn=24);
            // Chamfer entry
            translate([0, 0, jig_h - 1])
                cylinder(d1=hole_dia[i] + 0.2, d2=hole_dia[i] + 3, h=2, $fn=24);
        }
    }
}

// ---- FENCE ----
module fence() {
    translate([0, jig_d, 0])
        cube([fence_l, fence_t, fence_h]);
}

// ---- ASSEMBLE ----
difference() {
    union() {
        jig_body();
        fence();
    }
    guide_holes();
}

// ---- LABELS (echo to console) ----
for (i = [0 : len(hole_dia) - 1])
    echo(str("Guide ", i+1, ": ", hole_dia[i], "mm dia @ X=", (jig_w - (len(hole_dia)-1)*hole_spacing)/2 + i*hole_spacing, "mm"));
