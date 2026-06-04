// ============================================================
// Wall-Mounted Shelf Bracket — French Cleat + L-Bracket
// ============================================================
// Compatible with French cleat wall system (19mm plywood cleats).
// Can also be screwed directly to wall studs.

// ---- PARAMETERS ----
shelf_depth  = 200;   // How deep the shelf extends from wall (mm)
shelf_w      = 50;    // Width of one bracket (multiple = longer shelf)
shelf_load   = 5000;  // Expected load in grams — used in thickness calc
wall_t       = 5;     // Wall thickness of bracket
gusset_t     = 4;     // Diagonal gusset thickness

// French cleat dimensions (standard, don't change unless your cleats differ)
cleat_t      = 19;    // Cleat material thickness (mm)
cleat_angle  = 45;    // Cleat angle (degrees)
hook_clearance = 1;   // Clearance over cleat top

// ---- DERIVED ----
hook_h = cleat_t + hook_clearance + 3;

// ---- BACK PLATE ----
module back_plate() {
    cube([shelf_w, wall_t, shelf_depth + hook_h]);
}

// ---- SHELF ARM ----
module shelf_arm() {
    translate([0, wall_t, 0])
        cube([shelf_w, shelf_depth, wall_t]);
}

// ---- DIAGONAL GUSSET ----
// Triangular support under shelf arm
module gusset() {
    gusset_depth = shelf_depth * 0.6;
    translate([shelf_w/2 - gusset_t/2, wall_t, wall_t])
    linear_extrude(gusset_t)
    polygon([[0, 0], [gusset_depth, 0], [0, gusset_depth * 0.5]]);
}

// ---- FRENCH CLEAT HOOK ----
// Hook that slides over wall cleat
module cleat_hook() {
    translate([0, 0, shelf_depth])
    difference() {
        cube([shelf_w, cleat_t + wall_t, hook_h]);
        // 45° undercut to grab the cleat
        translate([0, wall_t, 0])
        rotate([0, 90, 0])
        linear_extrude(shelf_w)
        polygon([[0, 0], [cleat_t, 0], [cleat_t, hook_h], [0, hook_h * 0.6]]);
    }
}

// ---- ASSEMBLE ----
union() {
    back_plate();
    shelf_arm();
    gusset();
    cleat_hook();
}
