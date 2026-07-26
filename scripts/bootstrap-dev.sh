#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage: scripts/bootstrap-dev.sh [--check-only] [--strict]

Configure repository Git hooks and report the local tools used by verification.
This script does not download tools or mutate cloud/package credentials.

Options:
  --check-only  Inspect tools without changing core.hooksPath.
  --strict      Fail when an optional full-profile tool is missing.
  --help        Show this help.
USAGE
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE="$(cd "$SCRIPT_DIR/.." && pwd)"
CHECK_ONLY=false
STRICT=false

while ((${#})); do
  case "$1" in
    --check-only) CHECK_ONLY=true ;;
    --strict) STRICT=true ;;
    --help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage
      exit 2
      ;;
  esac
  shift
done

for required in git bash node npm; do
  if ! command -v "$required" >/dev/null 2>&1; then
    echo "Required developer tool is missing: $required" >&2
    exit 1
  fi
done

missing=()
for optional in go python3 mvn docker terraform pwsh shellcheck; do
  if command -v "$optional" >/dev/null 2>&1; then
    echo "available: $optional"
  else
    echo "unavailable: $optional (some full/scoped checks cannot run)"
    missing+=("$optional")
  fi
done

if [[ "$CHECK_ONLY" == false ]]; then
  git -C "$WORKSPACE" config core.hooksPath .githooks
  echo "Configured core.hooksPath=.githooks"
fi

if [[ "$STRICT" == true && ${#missing[@]} -gt 0 ]]; then
  echo "Strict bootstrap failed; missing: ${missing[*]}" >&2
  exit 1
fi

echo "Developer bootstrap complete. No credentials or global packages changed."
