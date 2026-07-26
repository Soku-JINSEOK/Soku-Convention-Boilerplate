#!/usr/bin/env bash
set -euo pipefail

script_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repository_directory="$(cd "$script_directory/.." && pwd)"
test_repository="$(mktemp -d)"
bare_repository="$(mktemp -d)"
test_gnupg_home="$(mktemp -d)"
cleanup() {
  rm -rf "$test_repository" "$bare_repository" "$test_gnupg_home"
}
trap cleanup EXIT
chmod 700 "$test_gnupg_home"
export GNUPGHOME="$test_gnupg_home"

git -C "$test_repository" init --quiet --initial-branch=main
git -C "$test_repository" config user.name "Release Gate Test"
git -C "$test_repository" config user.email "release-gate@example.invalid"
git -C "$test_repository" config commit.gpgsign false
git -C "$test_repository" config tag.gpgsign false
touch "$test_repository/tracked"
git -C "$test_repository" add tracked
git -C "$test_repository" commit --quiet -m "test: initialize repository"

notes="$repository_directory/docs/releases/v1.0.0.md"
verifier="$script_directory/verify-release-tag.sh"
creator="$script_directory/create-release-tag.sh"
placeholder_fingerprint="0000000000000000000000000000000000000000"

if "$verifier" \
  --tag invalid/v1.0.0 \
  --notes-file "$notes" \
  --expected-fingerprint "$placeholder_fingerprint" \
  --check-notes-only >/dev/null 2>&1; then
  echo "Malformed release tags must be rejected." >&2
  exit 1
fi

git -C "$test_repository" tag v1.0.0
if (
  cd "$test_repository" &&
    "$verifier" \
      --tag v1.0.0 \
      --notes-file "$notes" \
      --expected-fingerprint "$placeholder_fingerprint"
) \
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
if (
  cd "$test_repository" &&
    "$verifier" \
      --tag v1.0.0 \
      --notes-file "$notes" \
      --expected-fingerprint "$placeholder_fingerprint"
) \
  >/dev/null 2>&1; then
  echo "Unsigned annotated release tags must be rejected." >&2
  exit 1
fi

git -C "$test_repository" tag -d v1.0.0 >/dev/null

gpg --batch --pinentry-mode loopback --passphrase '' \
  --quick-generate-key \
  "Approved Release Signer <approved@example.invalid>" rsa2048 sign 0
gpg --batch --pinentry-mode loopback --passphrase '' \
  --quick-generate-key \
  "Unapproved Release Signer <unapproved@example.invalid>" rsa2048 sign 0

approved_fingerprint="$(
  gpg --batch --with-colons --fingerprint \
    "Approved Release Signer <approved@example.invalid>" |
    awk -F: '$1 == "fpr" {print toupper($10); exit}'
)"
unapproved_fingerprint="$(
  gpg --batch --with-colons --fingerprint \
    "Unapproved Release Signer <unapproved@example.invalid>" |
    awk -F: '$1 == "fpr" {print toupper($10); exit}'
)"

git -C "$test_repository" config user.signingkey "$approved_fingerprint"
git -C "$test_repository" tag -s -F "$annotation" v1.0.0
(
  cd "$test_repository"
  "$verifier" \
    --tag v1.0.0 \
    --notes-file "$notes" \
    --expected-fingerprint "$approved_fingerprint"
) >/dev/null

git -C "$test_repository" config user.signingkey "$unapproved_fingerprint"
git -C "$test_repository" tag -s -F "$annotation" v1.0.1
git -C "$test_repository" verify-tag v1.0.1 >/dev/null 2>&1
if (
  cd "$test_repository" &&
    "$verifier" \
      --tag v1.0.1 \
      --notes-file "$notes" \
      --expected-fingerprint "$approved_fingerprint"
) \
  >/dev/null 2>&1; then
  echo "A valid signature from an unapproved signer must be rejected." >&2
  exit 1
fi

rm -f "$annotation"
printf '%s\n' \
  '{' \
  '  "signing": {' \
  "    \"activeFingerprint\": \"$approved_fingerprint\"" \
  '  }' \
  '}' >"$test_repository/release-identity.json"
git -C "$test_repository" add release-identity.json
git -C "$test_repository" commit --quiet -m "test: add release identity"
git -C "$bare_repository" init --quiet --bare
git -C "$test_repository" remote add origin "$bare_repository"
git -C "$test_repository" push --quiet --set-upstream origin main

git -C "$test_repository" config user.signingkey "$approved_fingerprint"
(
  cd "$test_repository"
  "$creator" --tag v9.9.9 --dry-run
) >/dev/null

git -C "$test_repository" config user.signingkey "$unapproved_fingerprint"
if (
  cd "$test_repository" &&
    "$creator" --tag v9.9.9 --dry-run
) >/dev/null 2>&1; then
  echo "Tag creation must reject an unapproved configured signer." >&2
  exit 1
fi

echo "Malformed, unsigned, and valid-but-unapproved release tags are rejected."
