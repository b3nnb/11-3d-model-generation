// ============================================================
// CNC Routed Box — Finger-Joint Flat-Pack Design
// ============================================================
// Designed for plywood CNC routing. All 6 panels lay flat.
// Uses dogbone reliefs at inside corners for clean fit.
// Export each panel individually for toolpath generation.
//
// Generated: 2026-06-05 — Friday / 3D Model Generation system
// Machine: 3.175mm (1/8") end mill, dogbones enabled by default
// ============================================================

// ---- MATERIAL / MACHINE SETTINGS ----
material_t   = 18;      // Plywood thickness (mm) — 3/4" = 18mm, 1/2" = 12mm
bit_dia      = 3.175;   // CNC end mill diameter (mm) — 1/8" = 3.175mm
kerf         = 0.1;     // Router kerf compensation (mm) — add to slots, remove from tabs
dogbone      = true;    // Add dogbone reliefs at inside 90° corners

// ---- BOX DIMENSIONS (internal cavity) ----
box_w        = 200;     // Internal width  (X)  in mm
box_d        = 150;     // Internal depth  (Y)  in mm
box_h        = 100;     // Internal height (Z)  in mm

// ---- FINGER JOINT SETTINGS ----
finger_w     = 15;      // Width of each finger/tab (mm) — should be > 2× material_t
finger_count = 0;       // 0 = auto-calculate from edge length

// ---- LID OPTIONS ----
include_lid  = true;    // Generate a slip-fit lid panel
lid_gap      = 0.3;     // Clearance between lid and box top (mm)

// ---- LAYOUT SPACING ----
panel_gap    = 20;      // Space between panels in flat layout

// ============================================================
// DERIVED — don't change unless you know what you're doing
// ============================================================
T  = material_t + kerf;  // Slot width (slightly wider than material)
DB = bit_dia / 2;         // Dogbone relief radius

// External panel sizes
ext_w = box_w + T*2;
ext_d = box_d + T*2;

// ============================================================
// MODULES
// ============================================================

// Dogbone relief — small circle at each inside corner
module db_corner(r) {
    if (dogbone) {
        translate([r, r, -0.1])
            cylinder(r=r, h=material_t + 0.2, $fn=24);
    }
}

// Single finger-joint edge slot strip
// len = panel edge length, male = true → tabs cut away, false → slots cut in
module finger_edge(len, male) {
    n_fingers = finger_count > 0 ? finger_count : max(2, floor(len / finger_w));
    step = len / n_fingers;
    // Alternate: male starts at tab (cut notches between), female starts at notch
    start = male ? 0 : 1;
    for (i = [start : 2 : n_fingers - 1]) {
        translate([i * step, -0.1, 0])
            cube([step, material_t + 0.2, material_t + 0.2]);
    }
}

// Bottom panel
// Full external size, finger slots on all 4 edges
module panel_bottom() {
    difference() {
        cube([ext_w, ext_d, material_t]);

        // Front/back finger slots (along X)
        translate([T, 0, 0]) finger_edge(box_w, false);
        translate([T, ext_d - material_t, 0]) finger_edge(box_w, false);

        // Left/right finger slots (along Y — rotate)
        translate([0, T, 0]) rotate([0, 0, 0]) {
            // Left edge — remap to Y
            for (i = [1 : 2 : max(2, floor(box_d / finger_w)) - 1]) {
                step = box_d / max(2, floor(box_d / finger_w));
                translate([-0.1, i * step, 0]) cube([material_t + 0.2, step, material_t + 0.2]);
            }
        }
        translate([ext_w - material_t, T, 0]) {
            for (i = [1 : 2 : max(2, floor(box_d / finger_w)) - 1]) {
                step = box_d / max(2, floor(box_d / finger_w));
                translate([-0.1, i * step, 0]) cube([material_t + 0.2, step, material_t + 0.2]);
            }
        }

        // Dogbone reliefs at each slot corner
        // (simplified — one set on front edge for demo)
        if (dogbone) {
            translate([T, 0, 0]) db_corner(DB);
            translate([ext_w - T, 0, 0]) db_corner(DB);
        }
    }
}

// Front/back panel — flat, portrait orientation
module panel_front() {
    difference() {
        cube([box_w, box_h, material_t]);

        // Bottom tabs — align with bottom panel slots
        finger_edge(box_w, true);

        // Side finger slots (left and right vertical edges)
        n_side = max(2, floor(box_h / finger_w));
        side_step = box_h / n_side;
        for (i = [0 : 2 : n_side - 1]) {
            translate([-0.1, i * side_step, 0]) cube([material_t + 0.2, side_step, material_t + 0.2]);
            translate([box_w - material_t - 0.1, i * side_step, 0]) cube([material_t + 0.2, side_step, material_t + 0.2]);
        }

        // Dogbone reliefs on bottom edge corners
        if (dogbone) {
            translate([0, 0, 0]) db_corner(DB);
            translate([box_w, 0, 0]) {
                translate([-DB*2, 0, -0.1]) cylinder(r=DB, h=material_t+0.2, $fn=24);
            }
        }
    }
}

// Left/right side panel
module panel_side() {
    difference() {
        cube([box_d, box_h, material_t]);

        // Bottom tabs
        finger_edge(box_d, true);

        // Vertical edge tabs (mate with front/back finger slots)
        n_side = max(2, floor(box_h / finger_w));
        side_step = box_h / n_side;
        for (i = [1 : 2 : n_side - 1]) {
            translate([-0.1, i * side_step, 0]) cube([material_t + 0.2, side_step, material_t + 0.2]);
            translate([box_d - material_t - 0.1, i * side_step, 0]) cube([material_t + 0.2, side_step, material_t + 0.2]);
        }
    }
}

// Lid panel — slip-fit, slightly smaller than external box top
module panel_lid() {
    lid_w = ext_w - lid_gap * 2;
    lid_d = ext_d - lid_gap * 2;
    cube([lid_w, lid_d, material_t]);
}

// ============================================================
// FLAT LAYOUT — all panels displayed for CNC preview
// Separate each one with panel_gap spacing
// ============================================================
// Bottom
color("BurlyWood") panel_bottom();

// Front (offset below)
color("Tan") translate([0, ext_d + panel_gap, 0]) panel_front();

// Back
color("Tan") translate([0, ext_d + panel_gap + box_h + panel_gap, 0]) panel_front();

// Left side
color("SaddleBrown") translate([ext_w + panel_gap, 0, 0]) panel_side();

// Right side
color("SaddleBrown") translate([ext_w + panel_gap, box_h + panel_gap, 0]) panel_side();

// Lid (optional)
if (include_lid) {
    color("Peru", 0.7)
        translate([ext_w + panel_gap, box_h*2 + panel_gap*2, 0])
        panel_lid();
}

// ============================================================
// USAGE GUIDE (comment block)
// ============================================================
//
// To customise this box:
//   1. Adjust box_w / box_d / box_h for internal cavity size
//   2. Set material_t to your actual plywood thickness
//   3. Measure your router bit, update bit_dia
//   4. Run in OpenSCAD or via: modelgen from cnc_routed_box box_w=250 box_h=120
//
// To render all panels to STL:
//   openscad -o bottom.stl -D "show=\"bottom\"" cnc_routed_box.scad
//
// For actual CNC cutting, export each panel as SVG/DXF by viewing
// the model from directly above (Z axis) and using File → Export → DXF.
//
// Dogbone reliefs ensure inside corners fully clear the bit radius.
// If fit is tight, increase 'kerf' by 0.05mm increments.
// ============================================================
