// ============================================================
// CNC Flatpack Box — Finger Joint / Tab-Slot Design
// ============================================================
// Designed for plywood CNC routing. All pieces are flat.
// Parts: Bottom, Front/Back, Left/Right, optional Lid
// Output each panel as a separate piece for CNC cutting.

// ---- MATERIAL / MACHINE PARAMETERS ----
material_t = 12;     // Plywood thickness (mm) — 1/2" = 12.7mm
bit_dia    = 3.175;  // CNC end mill diameter (mm) — 1/8" = 3.175mm
kerf       = 0.1;    // Laser/router kerf adjust (mm)
dogbone    = true;   // Add dogbone reliefs for inside corners?

// ---- BOX DIMENSIONS (internal) ----
box_w = 150;   // Internal width (X)
box_d = 100;   // Internal depth (Y)
box_h = 80;    // Internal height (Z)

// ---- FINGER JOINT PARAMETERS ----
finger_w = 12;  // Width of each finger/tab (mm)

// ---- DERIVED ----
T = material_t + kerf;
DB = dogbone ? bit_dia/2 : 0;  // Dogbone radius

// ---- DOGBONE RELIEF ----
module dogbone_corner(r) {
    if (dogbone) {
        translate([r, r, -0.1]) cylinder(r=r, h=T + 0.2, $fn=16);
    }
}

// ---- FINGER JOINT EDGE (generates tabs/slots for one edge) ----
// len = edge length, male = true → tabs, false → slots
module finger_edge(len, male) {
    n = floor(len / finger_w);
    step = len / n;
    start = male ? 0 : 1;
    for (i = [start : 2 : n-1]) {
        translate([i * step, -0.1, 0])
            cube([step, T + 0.2, T + 0.2]);
    }
}

// ---- BOTTOM PANEL ----
module panel_bottom() {
    difference() {
        cube([box_w + T*2, box_d + T*2, T]);
        // Corner notches
        cube([T, T, T + 0.2]);
        translate([box_w + T, 0, 0]) cube([T, T, T + 0.2]);
        translate([0, box_d + T, 0]) cube([T, T, T + 0.2]);
        translate([box_w + T, box_d + T, 0]) cube([T, T, T + 0.2]);
    }
}

// ---- FRONT / BACK PANEL ----
module panel_front() {
    difference() {
        cube([box_w + T*2, box_h, T]);
        // Bottom finger slots
        translate([T, 0, 0]) finger_edge(box_w, false);
        // Side finger slots (vertical)
        // (simplified — full implementation would intersect side tabs)
    }
}

// ---- SIDE PANEL ----
module panel_side() {
    difference() {
        cube([box_d, box_h, T]);
        // Bottom slots
        finger_edge(box_d, false);
    }
}

// ---- LAYOUT (flat on build plate for CNC preview) ----
// Bottom
panel_bottom();
// Front (offset for viewing)
translate([0, box_d + T*2 + 20, 0]) panel_front();
// Back
translate([0, box_d + T*2 + box_h + 40, 0]) panel_front();
// Left side
translate([box_w + T*2 + 20, 0, 0]) panel_side();
// Right side
translate([box_w + T*2 + box_d + 40, 0, 0]) panel_side();
