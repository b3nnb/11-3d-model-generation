#!/usr/bin/env python3
"""
vizparams.py — Parameter Visualization Tool for OpenSCAD files
Reads a .scad file and outputs a formatted table of all parameters.

Usage:
    python3 vizparams.py model.scad
    python3 vizparams.py model.scad --json
    python3 vizparams.py openscad/samples/  # scan a directory
"""

import re
import sys
import json
import os
import argparse
from pathlib import Path


PARAM_PATTERN = re.compile(
    r'^(?P<name>[a-zA-Z_][a-zA-Z0-9_]*)\s*=\s*(?P<value>[^;]+);'
    r'(?:\s*//\s*(?P<comment>.*))?',
    re.MULTILINE
)

# Detect parameter type from value string
def infer_type(value: str) -> str:
    value = value.strip()
    if value.startswith('"') or value.startswith("'"):
        return "string"
    if re.match(r'^\[.*\]$', value):
        return "list"
    if re.match(r'^true|false$', value, re.IGNORECASE):
        return "bool"
    if re.match(r'^-?[\d.]+$', value):
        return "number"
    return "expression"


def parse_scad_params(path: Path) -> list[dict]:
    """Extract top-level parameters from a .scad file."""
    content = path.read_text()

    # Only look at lines before the first module/function definition
    first_module = re.search(r'^(module|function)\s', content, re.MULTILINE)
    if first_module:
        content = content[:first_module.start()]

    params = []
    seen = set()

    for m in PARAM_PATTERN.finditer(content):
        name = m.group("name")
        value = m.group("value").strip()
        comment = (m.group("comment") or "").strip()

        # Skip OpenSCAD built-ins and all-caps constants
        if name.startswith("$") or name in ("true", "false", "undef"):
            continue
        if name in seen:
            continue
        seen.add(name)

        params.append({
            "name": name,
            "value": value,
            "type": infer_type(value),
            "description": comment,
        })

    return params


def render_table(params: list[dict], filename: str) -> str:
    """Render params as a human-readable table."""
    if not params:
        return f"  (no top-level parameters found in {filename})"

    col_name = max(len(p["name"]) for p in params)
    col_val  = max(len(p["value"]) for p in params)
    col_type = max(len(p["type"]) for p in params)
    col_desc = 40

    col_name = max(col_name, 10)
    col_val  = max(col_val, 8)
    col_type = max(col_type, 6)

    header = (
        f"  {'Parameter':<{col_name}}  {'Value':<{col_val}}  {'Type':<{col_type}}  Description\n"
        f"  {'-'*col_name}  {'-'*col_val}  {'-'*col_type}  {'-'*col_desc}"
    )
    rows = []
    for p in params:
        desc = p["description"]
        if len(desc) > col_desc:
            desc = desc[:col_desc-3] + "..."
        rows.append(
            f"  {p['name']:<{col_name}}  {p['value']:<{col_val}}  {p['type']:<{col_type}}  {desc}"
        )
    return header + "\n" + "\n".join(rows)


def scan_file(path: Path, as_json: bool = False):
    params = parse_scad_params(path)
    if as_json:
        return {"file": str(path), "params": params}
    lines = [f"\n📐 {path.name} ({len(params)} parameters)"]
    lines.append(render_table(params, path.name))
    return "\n".join(lines)


def main():
    parser = argparse.ArgumentParser(description="Visualize OpenSCAD model parameters")
    parser.add_argument("target", help="Path to .scad file or directory of .scad files")
    parser.add_argument("--json", action="store_true", help="Output JSON instead of table")
    args = parser.parse_args()

    target = Path(args.target)

    if target.is_dir():
        scad_files = sorted(target.glob("*.scad"))
        if not scad_files:
            print(f"No .scad files found in {target}", file=sys.stderr)
            sys.exit(1)
        if args.json:
            results = [scan_file(f, as_json=True) for f in scad_files]
            print(json.dumps(results, indent=2))
        else:
            print(f"🗂️  Scanning {len(scad_files)} files in {target}/")
            for f in scad_files:
                print(scan_file(f))
    elif target.is_file():
        if args.json:
            result = scan_file(target, as_json=True)
            print(json.dumps(result, indent=2))
        else:
            print(scan_file(target))
    else:
        print(f"Error: {target} not found", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
