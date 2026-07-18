#!/usr/bin/env bash
set -euo pipefail

# Verifies that sync-boilerplate.sh and sync-boilerplate.ps1 produce identical
# output, and that neither leaks build artifacts (node_modules/, dist/,
# __pycache__/, etc.) that may exist locally but are not git-tracked.
#
# Usage: scripts/verify-sync-parity.sh
#
# Requires: git, pwsh (PowerShell 7+, preinstalled on GitHub-hosted ubuntu runners).

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if ! command -v pwsh >/dev/null 2>&1; then
  echo "pwsh (PowerShell 7+) is required to verify sync-boilerplate.ps1" >&2
  exit 1
fi

sh_target="$(mktemp -d)"
ps_target="$(mktemp -d)"
cleanup() {
  rm -rf "$sh_target" "$ps_target"
}
trap cleanup EXIT

echo "Running sync-boilerplate.sh -> $sh_target"
"$script_dir/sync-boilerplate.sh" --target "$sh_target" --include-readme

echo "Running sync-boilerplate.ps1 -> $ps_target"
pwsh -NoProfile -NonInteractive -Command \
  "& '$script_dir/sync-boilerplate.ps1' -TargetRoot '$ps_target' -IncludeReadme"

echo "Diffing outputs..."
if ! diff -rq "$sh_target" "$ps_target"; then
  echo "Mismatch: sync-boilerplate.sh and sync-boilerplate.ps1 produced different output." >&2
  exit 1
fi

echo "Checking that the independently released soku source is excluded..."
if [[ -e "$sh_target/soku" || -e "$ps_target/soku" ]]; then
  echo "Unexpected soku/ directory in manual sync output." >&2
  exit 1
fi

echo "Checking for leaked build artifacts..."
leaked=0
for pattern in node_modules dist __pycache__ .venv coverage .pytest_cache .mypy_cache .ruff_cache; do
  if find "$sh_target" "$ps_target" -iname "$pattern" -print -quit | grep -q .; then
    echo "Found unexpected artifact matching '$pattern' in sync output." >&2
    leaked=1
  fi
done

if [[ "$leaked" -eq 1 ]]; then
  exit 1
fi

echo "sync-boilerplate.sh and sync-boilerplate.ps1 are in parity, no leaked artifacts."
