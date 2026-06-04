// ============================================================
// Parametric Drawer Divider
// ============================================================
// Snap-fit divider strips for tool drawers, desk organizers.
// Set drawer dimensions, print as many dividers as needed.

// ---- PARAMETERS ----
// Drawer interior dimensions
drawer_w     = 300;   // Drawer internal width (mm)
drawer_d     = 400;   // Drawer internal depth (mm)
drawer_h     = 50;    // Drawer internal height (mm)

// Divider geometry
div_t        = 2.5;   // Divider thickness (mm)
div_h        = 48;    // Divider height (should be ≤ drawer_h - 1mm clearance)
slot_w       = 3.0;   // Slot width for cross-fit notching (div_t + 0.5mm)
slot_depth   = drawer_h * 0.4;  // How deep interlocking slots cut

// ---- HORIZONTAL DIVIDER (spans full drawer width) ----
module h_divider(position_y) {
    // Slots every 50mm for vertical dividers to cross
    translate([0, position_y, 0])
    difference() {
        cube([drawer_w, div_t, div_h]);
        // Top slots (for vertical divider cross-fit)
        for (x = [50 : 50 : drawer_w - 50])
            translate([x - slot_w/2, -0.1, div_h - slot_depth])
                cube([slot_w, div_t + 0.2, slot_depth + 0.1]);
    }
}

// ---- VERTICAL DIVIDER (spans full drawer depth) ----
module v_divider(position_x) {
    translate([position_x, 0, 0])
    difference() {
        cube([div_t, drawer_d, div_h]);
        // Bottom slots (cross-fit with horizontal dividers)
        for (y = [50 : 50 : drawer_d - 50])
            translate([-0.1, y - slot_w/2, 0])
                cube([div_t + 0.2, slot_w, slot_depth + 0.1]);
    }
}

// ---- SAMPLE LAYOUT ----
// This renders an example 2-section drawer layout.
// In real use, print individual dividers and assemble.
color("LightBlue", 0.8) {
    h_divider(drawer_d * 0.35);
    v_divider(drawer_w * 0.5);
}

// ---- DRAWER OUTLINE (ghost reference, remove for print) ----
color("Gray", 0.1)
difference() {
    cube([drawer_w, drawer_d, drawer_h]);
    translate([div_t, div_t, div_t])
        cube([drawer_w - div_t*2, drawer_d - div_t*2, drawer_h]);
}
