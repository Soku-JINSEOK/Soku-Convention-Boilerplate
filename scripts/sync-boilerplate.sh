#!/usr/bin/env bash
set -euo pipefail

# Bash equivalent of sync-boilerplate.ps1 for Linux/macOS users.
# Keep the copied item list identical between sync-boilerplate.ps1 and sync-boilerplate.sh.

usage() {
  cat <<'EOF'
Usage: sync-boilerplate.sh --target <dir> [--force] [--include-readme] [--dry-run]

  --target <dir>      Destination directory (created if it does not exist)
  --force             Overwrite existing files at the destination
  --include-readme    Also copy README.md (use only when bootstrapping from scratch)
  --dry-run           Print the files that would be copied without copying them
EOF
}

target_root=""
force=0
include_readme=0
dry_run=0

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
    --dry-run)
      dry_run=1
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

if ! git -C "$source_root" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "Error: source root is not a git checkout: $source_root" >&2
  echo "This script only copies git-tracked files, so it requires a git repository as its source." >&2
  exit 1
fi

if [[ "$dry_run" -eq 0 ]]; then
  mkdir -p "$target_root"
  target_root="$(cd "$target_root" && pwd)"
fi

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

# Copies only git-tracked files under $1 (relative to $source_root), so build
# artifacts and other .gitignore'd content sitting in a local checkout (e.g.
# node_modules/, dist/, __pycache__/) never leak into the sync target.
copy_tracked_dir() {
  local relative_path="$1"
  local destination_path="$target_root/$relative_path"

  if [[ "$dry_run" -eq 0 ]]; then
    if [[ "$force" -eq 0 && -e "$destination_path" ]]; then
      echo "Destination already exists (use --force to overwrite): $destination_path" >&2
      exit 1
    fi
    mkdir -p "$destination_path"
  fi

  local file
  while IFS= read -r -d '' file; do
    if [[ "$dry_run" -eq 1 ]]; then
      echo "$file"
      continue
    fi
    mkdir -p "$target_root/$(dirname "$file")"
    cp "$source_root/$file" "$target_root/$file"
  done < <(git -C "$source_root" ls-files -z -- "$relative_path")
}

for relative_path in "${items[@]}"; do
  source_path="$source_root/$relative_path"
  if [[ ! -e "$source_path" ]]; then
    echo "Missing source item: $relative_path" >&2
    exit 1
  fi

  if [[ -d "$source_path" ]]; then
    copy_tracked_dir "$relative_path"
    copied+=("$relative_path")
    continue
  fi

  destination_path="$target_root/$relative_path"

  if [[ "$dry_run" -eq 1 ]]; then
    echo "$relative_path"
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

if [[ "$dry_run" -eq 1 ]]; then
  exit 0
fi

echo "Convention sync completed."
echo "Target root: $target_root"
echo "Copied items:"
for item in "${copied[@]}"; do
  echo " - $item"
done
