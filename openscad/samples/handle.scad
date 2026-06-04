// ============================================================
// Parametric Handle — Replacement / Custom
// ============================================================
// Printable replacement handle for tools, drawers, doors.
// Screw-mount style, adjustable ergonomic profile.

// ---- PARAMETERS ----
handle_l    = 100;   // Handle length (mm)
handle_dia  = 25;    // Grip diameter at thickest point (ergonomic: 22–30mm)
grip_taper  = 0.8;   // Taper ratio at ends (1.0 = cylinder, 0.7 = teardrop grip)
mount_w     = 80;    // Distance between mounting holes (mm)
screw_d     = 4.5;   // M4 screw clearance = 4.5mm
head_counter = 4;    // Countersink depth for flush screw head
fillet      = 5;     // End rounding radius

// ---- GRIP BODY ----
// Hull between end circles creates the tapered grip shape
module grip_body() {
    end_r = (handle_dia / 2) * grip_taper;
    mid_r = handle_dia / 2;
    hull() {
        // End caps
        translate([0, 0, fillet]) sphere(r=end_r, $fn=32);
        translate([0, 0, handle_l - fillet]) sphere(r=end_r, $fn=32);
        // Middle body
        translate([0, 0, handle_l * 0.3]) cylinder(r=mid_r, h=handle_l*0.4, $fn=48);
    }
}

// ---- MOUNTING POSTS ----
// Flat mounting pads with screw holes under the handle
module mount_posts() {
    pad_h = 8;
    pad_w = 20;
    for (y = [-mount_w/2, mount_w/2]) {
        translate([-pad_w/2, y - pad_w/2, 0])
        difference() {
            cube([pad_w, pad_w, pad_h]);
            // Countersunk screw hole
            translate([pad_w/2, pad_w/2, -0.1])
                cylinder(d=screw_d, h=pad_h + 0.2, $fn=24);
            translate([pad_w/2, pad_w/2, pad_h - head_counter])
                cylinder(d=screw_d * 2, h=head_counter + 0.1, $fn=24);
        }
    }
}

// ---- ROTATE HANDLE HORIZONTAL ----
rotate([90, 0, 0])
translate([0, 0, -handle_l/2]) {
    grip_body();
    translate([0, 0, 0]) mount_posts();
}
