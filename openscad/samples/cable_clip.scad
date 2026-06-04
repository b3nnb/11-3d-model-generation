// ============================================================
// Cable Management Clip — Wall / Desk Adhesive Mount
// ============================================================
// Holds cables against a surface. Print multiple, press-fit cable.
// Sized for common cable diameters (USB-C, HDMI, power).

// ---- PARAMETERS ----
cable_dia   = 7;    // Cable outer diameter (mm). USB-C ~4.5mm, HDMI ~8mm, power ~7mm
clip_w      = 18;   // Width of the clip (mm)
base_t      = 3;    // Base plate thickness (mm)
base_l      = 24;   // Base plate length (mm)
arm_t       = 2.5;  // Clip arm thickness (mm)
arm_gap     = 0.4;  // Gap between arms for cable snap-in (press-fit clearance)
screw_d     = 3.5;  // Screw hole diameter (3M screw / M3 = 3.5mm clearance)

// ---- DERIVED ----
cr = cable_dia / 2;       // Cable radius
inner_r = cr + 0.3;       // Inner channel radius (slight clearance)
outer_r = cr + arm_t;     // Outer radius of the clip

// ---- BASE PLATE ----
module base_plate() {
    hull() {
        for (x = [3, base_l - 3])
            translate([x, clip_w/2, 0]) cylinder(r=3, h=base_t, $fn=24);
    }
}

// ---- SCREW HOLES ----
module screw_holes() {
    for (x = [5, base_l - 5])
        translate([x, clip_w/2, -0.1])
            cylinder(d=screw_d, h=base_t + 0.2, $fn=24);
}

// ---- CLIP ARMS ----
module clip_arms() {
    // Two arms form a C-shape to grip the cable
    // Bottom arm (solid)
    translate([base_l/2 - clip_w/2, 0, base_t])
    difference() {
        cylinder(r=outer_r, h=clip_w, $fn=48, center=false);
        // Inner channel
        translate([0, 0, -0.1]) cylinder(r=inner_r, h=clip_w + 0.2, $fn=48);
        // Open top gap for cable insertion
        translate([-outer_r - 0.1, -arm_gap/2, -0.1])
            cube([outer_r + 0.1, arm_gap, clip_w + 0.2]);
        // Cut lower half so it's a C not an O
        translate([-outer_r - 0.1, -outer_r - 0.1, -0.1])
            cube([outer_r * 2 + 0.2, outer_r + 0.1, clip_w + 0.2]);
    }
}

// ---- ASSEMBLE ----
difference() {
    union() {
        base_plate();
        clip_arms();
    }
    screw_holes();
}
