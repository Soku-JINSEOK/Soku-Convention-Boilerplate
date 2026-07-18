#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
module_dir="$(cd "$script_dir/.." && pwd)"
output_dir="$(mktemp -d)"
extract_dir="$(mktemp -d)"
cleanup() {
  rm -rf "$output_dir" "$extract_dir"
}
trap cleanup EXIT

set +e
"$script_dir/package.sh" >/dev/null 2>&1
missing_argument_code=$?
set -e
if [[ "$missing_argument_code" -ne 2 ]]; then
  echo "Package script must reject missing required arguments with exit code 2" >&2
  exit 1
fi

before_diff="$(git -C "$module_dir" diff -- .)"
package_args=(
  --version v0.1.0-test
  --commit 0123456789abcdef0123456789abcdef01234567
  --built-at 2026-07-18T00:00:00Z
  --output-dir "$output_dir"
)
"$script_dir/package.sh" "${package_args[@]}"

mapfile -t archives < <(find "$output_dir" -maxdepth 1 -type f \( -name '*.tar.gz' -o -name '*.zip' \) -printf '%f\n' | LC_ALL=C sort)
expected=(
  soku_v0.1.0-test_darwin_amd64.tar.gz
  soku_v0.1.0-test_darwin_arm64.tar.gz
  soku_v0.1.0-test_linux_amd64.tar.gz
  soku_v0.1.0-test_linux_arm64.tar.gz
  soku_v0.1.0-test_windows_amd64.zip
)
if [[ "${archives[*]}" != "${expected[*]}" ]]; then
  echo "Unexpected archive set: ${archives[*]}" >&2
  exit 1
fi

for archive in "${expected[@]}"; do
  if [[ "$archive" == *.zip ]]; then
    contents="$(go -C "$module_dir" run ./internal/packagezip --list "$output_dir/$archive" | LC_ALL=C sort)"
    expected_contents=$'LICENSE\nTHIRD_PARTY_NOTICES.md\nsoku.exe'
  else
    contents="$(tar -tzf "$output_dir/$archive" | LC_ALL=C sort)"
    expected_contents=$'LICENSE\nTHIRD_PARTY_NOTICES.md\nsoku'
  fi
  if [[ "$contents" != "$expected_contents" ]]; then
    echo "Unexpected contents in $archive:" >&2
    echo "$contents" >&2
    exit 1
  fi
done

(
  cd "$output_dir"
  sha256sum --check checksums.txt
)
if [[ "$(wc -l <"$output_dir/checksums.txt")" -ne 5 ]]; then
  echo "checksums.txt must contain exactly five entries" >&2
  exit 1
fi
checksum_names="$(awk '{print $2}' "$output_dir/checksums.txt")"
if [[ "$checksum_names" != "$(printf '%s\n' "${expected[@]}")" ]]; then
  echo "checksums.txt is not sorted by archive filename" >&2
  exit 1
fi

tar -xzf "$output_dir/soku_v0.1.0-test_linux_amd64.tar.gz" -C "$extract_dir"
if [[ ! -x "$extract_dir/soku" ]]; then
  echo "Packaged Linux binary is not executable" >&2
  exit 1
fi
version_json="$($extract_dir/soku --json --version)"
if [[ "$version_json" != *'"version":"v0.1.0-test"'* || "$version_json" != *'"commit":"0123456789abcdef0123456789abcdef01234567"'* ]]; then
  echo "Packaged binary metadata is incorrect: $version_json" >&2
  exit 1
fi

first_checksums="$(cat "$output_dir/checksums.txt")"
"$script_dir/package.sh" "${package_args[@]}"
if [[ "$first_checksums" != "$(cat "$output_dir/checksums.txt")" ]]; then
  echo "Packaging is not reproducible" >&2
  exit 1
fi
if [[ "$before_diff" != "$(git -C "$module_dir" diff -- .)" ]]; then
  echo "Packaging changed tracked files" >&2
  exit 1
fi

echo "All five soku archives and checksums are valid and reproducible."
