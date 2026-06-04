// ============================================================
// Parametric Enclosure — Electronics / Project Box
// ============================================================
// General-purpose electronics enclosure with lid, cable glands,
// and PCB mounting standoffs. Adjustable for any PCB.

// ---- PARAMETERS ----
// PCB dimensions
pcb_w       = 100;   // PCB width (mm)
pcb_d       = 80;    // PCB depth (mm)
pcb_standoff = 6;    // Standoff height (space under board for traces/connectors)

// Enclosure
wall        = 3;     // Wall thickness
lid_lip     = 5;     // Snap-fit lid overlap
clearance   = 5;     // Space around PCB on each side
top_clear   = 25;    // Clearance above PCB for components

// Cable glands
gland_count = 2;     // Number of cable entry holes on one side
gland_d     = 10;    // Cable gland hole diameter (M12 gland = 12mm, standard = 10mm)

// Mounting
mount_hole  = 4.5;   // M4 corner mounting holes
corner_r    = 3;     // Corner rounding

// ---- DERIVED ----
box_w = pcb_w + clearance*2 + wall*2;
box_d = pcb_d + clearance*2 + wall*2;
box_h = pcb_standoff + top_clear + wall*2;

// ---- HELPER: ROUNDED HULL ----
module rbox(w, d, h, r) {
    hull() {
        for (x = [r, w-r], y = [r, d-r])
            translate([x, y, 0]) cylinder(r=r, h=h, $fn=24);
    }
}

// ---- BOX BODY ----
module box_body() {
    difference() {
        rbox(box_w, box_d, box_h - lid_lip, corner_r);
        // Inner cavity
        translate([wall, wall, wall])
            rbox(box_w - wall*2, box_d - wall*2, box_h, max(0, corner_r - wall));
        // Cable gland holes
        for (i = [1 : gland_count])
            translate([box_w * i / (gland_count + 1), -0.1, box_h * 0.5])
            rotate([-90, 0, 0])
                cylinder(d=gland_d, h=wall + 0.2, $fn=24);
        // Corner mounting holes
        for (x = [wall + 4, box_w - wall - 4], y = [wall + 4, box_d - wall - 4])
            translate([x, y, -0.1]) cylinder(d=mount_hole, h=wall + 0.2, $fn=24);
    }
}

// ---- PCB STANDOFFS ----
module pcb_standoffs() {
    for (x = [wall + clearance + 3, wall + clearance + pcb_w - 3],
         y = [wall + clearance + 3, wall + clearance + pcb_d - 3])
    translate([x, y, wall])
    difference() {
        cylinder(d=6, h=pcb_standoff, $fn=24);
        translate([0, 0, -0.1]) cylinder(d=2.7, h=pcb_standoff + 0.2, $fn=24);
    }
}

// ---- LID ----
module lid() {
    translate([0, 0, box_h + 10]) // offset for display
    difference() {
        rbox(box_w, box_d, wall + lid_lip, corner_r);
        // Inner snap-fit lip
        translate([wall, wall, wall])
            rbox(box_w - wall*2, box_d - wall*2, lid_lip + 0.1, max(0, corner_r - wall));
        // Vent slots (optional — comment out if airtight needed)
        for (x = [1 : 3])
            translate([box_w * 0.25 * x - 5, box_d*0.4, -0.1])
                cube([6, box_d*0.2, wall + 0.2]);
    }
}

// ---- ASSEMBLE ----
box_body();
pcb_standoffs();
lid();
