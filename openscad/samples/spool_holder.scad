// ============================================================
// Filament Spool Holder — Wall Mount
// ============================================================
// Mounts a filament spool on the wall or frame of a printer.
// Handles standard 200mm spools (Bambu, Prusa, eSun, etc.)

// ---- PARAMETERS ----
spool_hub_dia  = 52;   // Spool inner hub diameter (Bambu/standard ~52mm)
spool_w        = 68;   // Spool width (most standard spools: 55–70mm)
rod_dia        = 10;   // Steel rod diameter (mm) — standard Prusa: 8mm, use 10mm for heavier
wall_t         = 4;    // Wall thickness
bracket_h      = 60;   // Height of bracket arm
mount_hole_d   = 4.5;  // M4 mounting screw clearance
fillet         = 3;    // Corner rounding

// ---- DERIVED ----
bearing_od = spool_hub_dia + wall_t*2;
rod_r      = rod_dia / 2;

// ---- BRACKET ARM ----
module bracket_arm() {
    hull() {
        cylinder(d=bearing_od, h=wall_t, $fn=48);
        translate([0, -bracket_h, 0])
        hull() {
            for (x = [-20, 20]) translate([x, 0, 0])
                cylinder(r=fillet, h=wall_t, $fn=24);
        }
    }
}

// ---- ROD SOCKET (holds 8–10mm steel rod) ----
module rod_socket() {
    difference() {
        cylinder(d=rod_dia + wall_t*2, h=spool_w + wall_t*2, $fn=32);
        translate([0, 0, -0.1])
            cylinder(d=rod_dia + 0.4, h=spool_w + wall_t*2 + 0.2, $fn=32);
    }
}

// ---- SPOOL BEARING RING ----
module bearing_ring() {
    difference() {
        cylinder(d=bearing_od, h=wall_t*2, $fn=48);
        translate([0, 0, -0.1])
            cylinder(d=spool_hub_dia + 0.5, h=wall_t*2 + 0.2, $fn=48);
    }
}

// ---- WALL PLATE ----
module wall_plate() {
    translate([-40, -bracket_h - 30, 0])
    difference() {
        cube([80, 30, wall_t]);
        for (x = [10, 70])
            translate([x, 15, -0.1]) cylinder(d=mount_hole_d, h=wall_t + 0.2, $fn=24);
    }
}

// ---- ASSEMBLE ----
difference() {
    union() {
        bracket_arm();
        bearing_ring();
        wall_plate();
        // Rod holder at top
        translate([0, 0, wall_t]) rod_socket();
    }
    // Rod center clearance
    translate([0, 0, -0.1]) cylinder(d=rod_dia + 0.4, h=wall_t + 0.2, $fn=32);
}
