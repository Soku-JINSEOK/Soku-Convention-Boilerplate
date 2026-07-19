#!/usr/bin/env bash
set -euo pipefail

script_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repository_directory="$(cd "$script_directory/.." && pwd)"
test_repository="$(mktemp -d)"
cleanup() {
  rm -rf "$test_repository"
}
trap cleanup EXIT

git -C "$test_repository" init --quiet
git -C "$test_repository" config user.name "Release Gate Test"
git -C "$test_repository" config user.email "release-gate@example.invalid"
git -C "$test_repository" config commit.gpgsign false
git -C "$test_repository" config tag.gpgsign false
touch "$test_repository/tracked"
git -C "$test_repository" add tracked
git -C "$test_repository" commit --quiet -m "test: initialize repository"

notes="$repository_directory/docs/releases/v1.0.0.md"
verifier="$script_directory/verify-release-tag.sh"

if "$verifier" \
  --tag invalid/v1.0.0 \
  --notes-file "$notes" \
  --check-notes-only >/dev/null 2>&1; then
  echo "Malformed release tags must be rejected." >&2
  exit 1
fi

git -C "$test_repository" tag v1.0.0
if (cd "$test_repository" && "$verifier" --tag v1.0.0 --notes-file "$notes") \
  >/dev/null 2>&1; then
  echo "Lightweight release tags must be rejected." >&2
  exit 1
fi

git -C "$test_repository" tag -d v1.0.0 >/dev/null
commit="$(git -C "$test_repository" rev-parse HEAD)"
annotation="$test_repository/annotation.md"
cp "$notes" "$annotation"
printf '\nSource commit: %s\n' "$commit" >>"$annotation"
git -C "$test_repository" tag -a -F "$annotation" v1.0.0
if (cd "$test_repository" && "$verifier" --tag v1.0.0 --notes-file "$notes") \
  >/dev/null 2>&1; then
  echo "Unsigned annotated release tags must be rejected." >&2
  exit 1
fi

echo "Malformed, lightweight, and unsigned release tags are rejected."
