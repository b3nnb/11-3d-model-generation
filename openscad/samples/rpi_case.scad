// ============================================================
// Raspberry Pi 4 Case — Parametric Board Mount
// ============================================================
// Snap-fit case for Raspberry Pi 4B.
// Adjust for other Pi variants by changing board dimensions.

// ---- BOARD DIMENSIONS (Pi 4B) ----
board_w     = 85;    // PCB width (mm)
board_d     = 56;    // PCB depth (mm)
board_t     = 1.6;   // PCB thickness (mm)
board_standoff_h = 5; // Height above bottom of standoff support

// ---- MOUNTING HOLES (Pi 4B — 58mm x 49mm pattern) ----
hole_x1 = 3.5;   hole_x2 = 61.5;
hole_y1 = 3.5;   hole_y2 = 52.5;
hole_d  = 2.7;   // M2.5 clearance = 2.7mm

// ---- CASE GEOMETRY ----
wall        = 2.5;   // Case wall thickness
base_t      = 2;     // Bottom plate thickness
lid_lip     = 3;     // Snap-fit lid overlap height
clearance   = 1.5;   // Clearance on each side around board
port_h      = 20;    // Port cutout height (covers all ports generously)

case_w = board_w + clearance*2 + wall*2;
case_d = board_d + clearance*2 + wall*2;
case_h = 30;  // Total case height — adjust for your board + components

// ---- STANDOFFS ----
module standoff(x, y) {
    translate([x + wall + clearance, y + wall + clearance, base_t])
    difference() {
        cylinder(d=5, h=board_standoff_h, $fn=24);
        translate([0, 0, -0.1]) cylinder(d=hole_d, h=board_standoff_h + 0.2, $fn=24);
    }
}

// ---- CASE BODY ----
module case_body() {
    difference() {
        cube([case_w, case_d, case_h]);
        // Inner cavity
        translate([wall, wall, base_t])
            cube([case_w - wall*2, case_d - wall*2, case_h]);
        // Port cutouts (adjust positions for real use)
        // USB side (right)
        translate([case_w - wall - 0.1, wall + 5, base_t + board_t + board_standoff_h])
            cube([wall + 0.2, 45, port_h]);
        // HDMI/USB-C side (left)
        translate([-0.1, wall + 5, base_t + board_t + board_standoff_h])
            cube([wall + 0.2, 45, port_h]);
    }
}

// ---- LID ----
module lid() {
    translate([0, 0, case_h + 5]) // offset for display — remove in final assembly
    difference() {
        cube([case_w, case_d, wall + lid_lip]);
        // Snap-fit inner lip
        translate([wall, wall, wall])
            cube([case_w - wall*2, case_d - wall*2, lid_lip + 0.1]);
        // Vent holes (3x3 grid)
        for (x = [1:3], y = [1:3])
            translate([case_w/4 * x - 3, case_d/4 * y - 3, -0.1])
                cylinder(d=4, h=wall + 0.2, $fn=24);
    }
}

// ---- ASSEMBLE ----
case_body();
standoff(hole_x1, hole_y1);
standoff(hole_x2, hole_y1);
standoff(hole_x1, hole_y2);
standoff(hole_x2, hole_y2);
lid();
