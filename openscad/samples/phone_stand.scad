// ============================================================
// Phone Stand — Parametric / 3D Print
// ============================================================
// Sample model: desk phone holder, adjustable angle + width
// Parameters are defined at top for easy customization.

// ---- PARAMETERS ----
phone_w    = 80;    // Phone width (mm) — iPhone 15: 71mm, S24: 70mm, +cases: up to 85mm
phone_t    = 12;    // Phone thickness (mm) — typical 10–15mm with case
base_d     = 80;    // Base depth (mm)
base_t     = 4;     // Base thickness (mm)
angle      = 70;    // Viewing angle from vertical (degrees) — 60–80 feels natural
wall       = 3;     // Wall/support thickness
lip_h      = 15;    // Bottom lip height to retain phone (mm)
fillet     = 2;     // Corner fillet radius

// ---- DERIVED ----
slot_w     = phone_w + 2;    // Slot slightly wider than phone
slot_h     = phone_t + 1.5;  // Slot slightly deeper than phone

// ---- BASE PLATE ----
module base() {
    hull() {
        for (x = [fillet, slot_w + wall*2 - fillet], y = [fillet, base_d - fillet]) {
            translate([x, y, 0]) cylinder(r=fillet, h=base_t, $fn=24);
        }
    }
}

// ---- REAR SUPPORT ----
// Angled back that the phone rests against
module rear_support() {
    support_h = 60;
    rotate([angle, 0, 0])
    translate([0, 0, 0])
    cube([slot_w + wall*2, wall, support_h]);
}

// ---- SIDE WALLS ----
module side_walls() {
    cube([wall, base_d, lip_h + base_t]);
    translate([slot_w + wall, 0, 0])
        cube([wall, base_d, lip_h + base_t]);
}

// ---- BOTTOM LIP ----
module bottom_lip() {
    translate([wall, 0, base_t])
        cube([slot_w, wall*2, lip_h]);
}

// ---- ASSEMBLE ----
union() {
    base();
    translate([0, 0, base_t]) {
        side_walls();
        bottom_lip();
        translate([0, base_d * 0.3, lip_h])
            rear_support();
    }
}
