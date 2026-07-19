#!/usr/bin/env bash
# Creates one locally signed release tag after validating release readiness.
set -euo pipefail

usage() {
  cat >&2 <<'EOF'
Usage: scripts/create-release-tag.sh [--tag <tag>] [--notes-file <path>] [--dry-run]

With no arguments, the script keeps its interactive tag prompt. It never pushes.
EOF
}

tag=""
notes_file=""
dry_run=false
while [[ $# -gt 0 ]]; do
  case "$1" in
    --tag)
      tag="${2:-}"
      shift 2
      ;;
    --notes-file)
      notes_file="${2:-}"
      shift 2
      ;;
    --dry-run)
      dry_run=true
      shift
      ;;
    -h | --help)
      usage
      exit 0
      ;;
    *)
      usage
      exit 2
      ;;
  esac
done

echo "Verified Release Tag Creator"

if [[ -z "$tag" ]]; then
  echo "Recent release tags:"
  git tag -l --sort=-v:refname 'v*' 'soku/v*' | head -n 5 || true
  read -rp "Enter next release tag (vX.Y.Z or soku/vX.Y.Z): " tag
fi

if [[ ! "$tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ && ! "$tag" =~ ^soku/v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Error: tag must be vMAJOR.MINOR.PATCH or soku/vMAJOR.MINOR.PATCH." >&2
  exit 2
fi
if [[ -n "$notes_file" && ! -f "$notes_file" ]]; then
  echo "Error: notes file does not exist: $notes_file" >&2
  exit 2
fi
if [[ "$(git branch --show-current)" != "main" ]]; then
  echo "Error: release tags may only be created from main." >&2
  exit 1
fi
if [[ -n "$(git status --porcelain)" ]]; then
  echo "Error: the working tree must be clean." >&2
  exit 1
fi

git fetch --quiet origin main
head_commit="$(git rev-parse HEAD)"
remote_main="$(git rev-parse refs/remotes/origin/main)"
if [[ "$head_commit" != "$remote_main" ]]; then
  echo "Error: main must exactly match origin/main." >&2
  exit 1
fi
if git show-ref --verify --quiet "refs/tags/$tag"; then
  echo "Error: local tag already exists: $tag" >&2
  exit 1
fi
if [[ -n "$(git ls-remote --tags origin "refs/tags/$tag")" ]]; then
  echo "Error: remote tag already exists: $tag" >&2
  exit 1
fi

signing_key="$(git config --get user.signingkey || true)"
if [[ -z "$signing_key" ]]; then
  echo "Error: user.signingkey is not configured." >&2
  exit 1
fi

if [[ "$dry_run" == true ]]; then
  echo "Dry run passed for $tag at $head_commit; no tag was created."
  exit 0
fi

annotation_file="$(mktemp)"
cleanup() {
  rm -f "$annotation_file"
}
trap cleanup EXIT
if [[ -n "$notes_file" ]]; then
  cp "$notes_file" "$annotation_file"
else
  printf 'Release %s\n' "$tag" >"$annotation_file"
fi
printf '\nSource commit: %s\n' "$head_commit" >>"$annotation_file"

git tag -s -F "$annotation_file" "$tag"
git verify-tag "$tag"
echo "Created and verified signed tag $tag. This script did not push it."
echo "Push explicitly after all companion tags are verified."
