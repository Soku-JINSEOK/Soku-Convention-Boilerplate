#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "Usage: $0 --tag <tag> --notes-file <path> [--check-notes-only]" >&2
}

tag=""
notes_file=""
check_notes_only=false
while [[ $# -gt 0 ]]; do
  case "$1" in
    --tag) tag="${2:-}"; shift 2 ;;
    --notes-file) notes_file="${2:-}"; shift 2 ;;
    --check-notes-only) check_notes_only=true; shift ;;
    *) usage; exit 2 ;;
  esac
done
if [[ -z "$tag" || -z "$notes_file" || ! -f "$notes_file" ]]; then
  usage
  exit 2
fi

require_line() {
  local file="$1"
  local text="$2"
  if ! grep -Fqx "$text" "$file"; then
    echo "Error: required release record is missing: $text" >&2
    exit 1
  fi
}

capture_line() {
  local prefix="$1"
  local line
  line="$(grep -F "$prefix" "$notes_file" | head -n 1 || true)"
  if [[ -z "$line" || "$line" == "$prefix" ]]; then
    echo "Error: required release record is missing a value: $prefix" >&2
    exit 1
  fi
  required_lines+=("$line")
}

required_lines=()
if [[ "$tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  axis="boilerplate"
  required_lines+=("Release axis: boilerplate")
  capture_line "Catalog contracts: "
  capture_line "Profiles: "
  capture_line "Compatible soku: "
  capture_line "Migration from previous release: "
  capture_line "Provider compatibility: "
  capture_line "Companion tag: "
  companion_tag="$(sed -n 's/^Companion tag: //p' "$notes_file")"
  if [[ "$companion_tag" != "none" && ! "$companion_tag" =~ ^soku/v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: boilerplate companion must be a soku/vMAJOR.MINOR.PATCH tag." >&2
    exit 1
  fi
elif [[ "$tag" =~ ^soku/v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  axis="soku"
  required_lines+=("Release axis: soku")
  capture_line "Manifest schemas: "
  capture_line "Boilerplate compatibility: "
  capture_line "Profiles: "
  capture_line "Provider compatibility: "
  capture_line "Recovery and exit-code contract: "
  capture_line "Package matrix: "
  capture_line "Lifecycle conformance evidence: "
  capture_line "Companion tag: "
  companion_tag="$(sed -n 's/^Companion tag: //p' "$notes_file")"
  if [[ "$companion_tag" != "none" && ! "$companion_tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: soku companion must be a vMAJOR.MINOR.PATCH tag." >&2
    exit 1
  fi
else
  echo "Error: unsupported release tag format: $tag" >&2
  exit 2
fi

for line in "${required_lines[@]}"; do require_line "$notes_file" "$line"; done
if [[ "$check_notes_only" == true ]]; then
  echo "Release notes contract passed for $tag ($axis)."
  exit 0
fi

object_type="$(git cat-file -t "refs/tags/$tag" 2>/dev/null || true)"
if [[ "$object_type" != "tag" ]]; then
  echo "Error: $tag must be an annotated tag." >&2
  exit 1
fi
git verify-tag "$tag"
resolved_commit="$(git rev-list -n 1 "$tag")"
annotation="$(git for-each-ref --format='%(contents)' "refs/tags/$tag")"
for line in "${required_lines[@]}"; do
  if ! grep -Fqx "$line" <<<"$annotation"; then
    echo "Error: signed annotation is missing: $line" >&2
    exit 1
  fi
done
if ! grep -Fqx "Source commit: $resolved_commit" <<<"$annotation"; then
  echo "Error: signed annotation does not record its resolved source commit." >&2
  exit 1
fi
echo "Signed annotated tag contract passed for $tag at $resolved_commit."
