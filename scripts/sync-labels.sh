#!/usr/bin/env bash
set -euo pipefail

# Creates/updates GitHub labels from .github/labels.yml using the gh CLI.
# Idempotent: safe to re-run any time the catalog changes.
#
# Usage: scripts/sync-labels.sh [--repo <owner/repo>]
#
# Requires: gh CLI authenticated (gh auth status), python3.

repo_arg=()
if [[ "${1:-}" == "--repo" ]]; then
  repo_arg=(--repo "$2")
fi

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
labels_file="$script_dir/../.github/labels.yml"

if [[ ! -f "$labels_file" ]]; then
  echo "Missing label catalog: $labels_file" >&2
  exit 1
fi

if ! command -v gh >/dev/null 2>&1; then
  echo "gh CLI is required" >&2
  exit 1
fi

# Parse and sync labels using Python to avoid string splitting issues and empty description bugs.
python3 - "$labels_file" "${repo_arg[@]}" <<'PYEOF'
import sys
import subprocess

path = sys.argv[1]
repo_args = sys.argv[2:]
name = color = description = None

def sync_label(n, c, d):
    if not n or not c:
        return
    desc = d if d is not None else ""
    print(f"Syncing label: {n}")
    cmd = ["gh", "label", "create", n, "--color", c, "--description", desc, "--force"] + repo_args
    subprocess.run(cmd, check=True)

with open(path, encoding="utf-8") as f:
    for line in f:
        stripped = line.strip()
        if not stripped or stripped.startswith("#"):
            continue
        if stripped.startswith("- name:"):
            sync_label(name, color, description)
            name = stripped.split(":", 1)[1].strip().strip('"')
            color = description = None
        elif stripped.startswith("color:"):
            color = stripped.split(":", 1)[1].strip().strip('"')
        elif stripped.startswith("description:"):
            description = stripped.split(":", 1)[1].strip().strip('"')

sync_label(name, color, description)
PYEOF

echo "Label sync completed."
