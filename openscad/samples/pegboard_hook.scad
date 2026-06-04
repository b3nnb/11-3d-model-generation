// ============================================================
// Pegboard Hook — Standard 1" (25.4mm) pegboard pattern
// ============================================================
// Parametric hook for standard pegboard storage walls.
// Adjust hook geometry for different tool types.

// ---- PARAMETERS ----
// Pegboard
hole_dia    = 6.5;   // Peg hole diameter (standard 1/4" = 6.35mm, use 6.5mm clearance)
hole_pitch  = 25.4;  // Hole spacing (standard 1" = 25.4mm)
peg_rows    = 2;     // How many holes to engage (1 = single, 2 = more stable)

// Hook geometry
hook_reach  = 60;    // How far hook extends from wall (mm)
hook_h      = 40;    // Vertical height of hook opening (mm)
hook_t      = 5;     // Hook arm thickness (mm)
hook_tip    = 10;    // Upward tip at end to prevent items falling off (mm)
load_cap    = 500;   // Expected load in grams (used in export checklist)

// ---- PEG PINS ----
// Cylindrical pins that insert into pegboard holes
module peg_pin(row) {
    peg_len = 22;    // Standard pegboard depth clearance
    translate([hole_dia/2, row * hole_pitch, 0])
    rotate([0, -90, 0])
    cylinder(d=hole_dia - 0.4, h=peg_len, $fn=24);
}

// ---- BACK PLATE ----
module back_plate() {
    plate_w = hole_dia * 2 + 4;
    plate_h = (peg_rows - 1) * hole_pitch + hole_dia * 3;
    hull() {
        for (y = [hole_dia, plate_h - hole_dia])
            translate([hole_dia, y, 0]) cylinder(r=hole_dia, h=hook_t, $fn=24);
    }
}

// ---- HORIZONTAL ARM ----
module arm() {
    translate([0, (peg_rows - 1) * hole_pitch / 2, hook_t])
    rotate([90, 0, 0])
    linear_extrude(hook_t)
    polygon([
        [0, 0],
        [hook_reach, 0],
        [hook_reach, hook_h],
        [hook_reach - hook_t, hook_h + hook_tip],  // tip curve
        [hook_reach - hook_t, hook_h],
        [hook_t, hook_h],
        [hook_t, hook_t],
        [0, hook_t]
    ]);
}

// ---- ASSEMBLE ----
union() {
    back_plate();
    for (row = [0 : peg_rows - 1]) {
        peg_pin(row);
    }
    arm();
}

// ---- ECHO LOAD INFO ----
echo(str("Expected load: ", load_cap, "g — inspect wall anchoring for loads >1kg"));
