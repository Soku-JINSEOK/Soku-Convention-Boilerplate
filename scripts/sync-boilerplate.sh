#!/usr/bin/env bash
set -euo pipefail

# Bash equivalent of sync-boilerplate.ps1 for Linux/macOS users.
# Keep the copied item list identical between sync-boilerplate.ps1 and sync-boilerplate.sh.

usage() {
  cat <<'EOF'
Usage: sync-boilerplate.sh --target <dir> [--force] [--include-readme]

  --target <dir>      Destination directory (created if it does not exist)
  --force             Overwrite existing files at the destination
  --include-readme    Also copy README.md (use only when bootstrapping from scratch)
EOF
}

target_root=""
force=0
include_readme=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --target)
      target_root="${2:-}"
      shift 2
      ;;
    --force)
      force=1
      shift
      ;;
    --include-readme)
      include_readme=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage
      exit 1
      ;;
  esac
done

if [[ -z "$target_root" ]]; then
  echo "Error: --target is required" >&2
  usage
  exit 1
fi

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source_root="$(cd "$script_dir/.." && pwd)"

mkdir -p "$target_root"
target_root="$(cd "$target_root" && pwd)"

# Keep this list identical to the $items array in sync-boilerplate.ps1.
items=(
  'BLUEPRINT.md'
  '.markdownlint.jsonc'
  'AGENTS.md'
  'CONTRIBUTING.md'
  'docs'
  'LICENSE'
  'SECURITY.md'
  '.editorconfig'
  '.gitignore'
  '.gitmessage'
  '.github'
  'templates'
  'scripts'
)

if [[ "$include_readme" -eq 1 ]]; then
  items=('README.md' 'README.ko.md' 'README.ja.md' "${items[@]}")
fi

copied=()

for relative_path in "${items[@]}"; do
  source_path="$source_root/$relative_path"
  if [[ ! -e "$source_path" ]]; then
    echo "Missing source item: $relative_path" >&2
    exit 1
  fi

  destination_path="$target_root/$relative_path"

  if [[ -d "$source_path" ]]; then
    if [[ "$force" -eq 0 && -e "$destination_path" ]]; then
      echo "Destination already exists (use --force to overwrite): $destination_path" >&2
      exit 1
    fi
    mkdir -p "$destination_path"
    cp -R "$source_path"/. "$destination_path"/
    copied+=("$relative_path")
    continue
  fi

  if [[ "$force" -eq 0 && -e "$destination_path" ]]; then
    echo "Destination already exists (use --force to overwrite): $destination_path" >&2
    exit 1
  fi

  mkdir -p "$(dirname "$destination_path")"
  cp "$source_path" "$destination_path"
  copied+=("$relative_path")
done

echo "Convention sync completed."
echo "Target root: $target_root"
echo "Copied items:"
for item in "${copied[@]}"; do
  echo " - $item"
done
