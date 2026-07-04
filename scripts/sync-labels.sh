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

# Parse the constrained labels.yml format (list of {name, color, description})
# without requiring a YAML library dependency.
python3 - "$labels_file" <<'PYEOF' | while IFS=$'\t' read -r name color description; do
import sys

path = sys.argv[1]
name = color = description = None

with open(path, encoding="utf-8") as f:
    for line in f:
        stripped = line.strip()
        if not stripped or stripped.startswith("#"):
            continue
        if stripped.startswith("- name:"):
            if name is not None:
                print(f"{name}\t{color}\t{description}")
            name = stripped.split(":", 1)[1].strip().strip('"')
            color = description = None
        elif stripped.startswith("color:"):
            color = stripped.split(":", 1)[1].strip().strip('"')
        elif stripped.startswith("description:"):
            description = stripped.split(":", 1)[1].strip().strip('"')

if name is not None:
    print(f"{name}\t{color}\t{description}")
PYEOF
  echo "Syncing label: $name"
  gh label create "$name" --color "$color" --description "$description" --force "${repo_arg[@]}"
done

echo "Label sync completed."
