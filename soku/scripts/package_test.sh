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

archives=()
while IFS= read -r archive; do
  archives+=("$(basename "$archive")")
done < <(find "$output_dir" -maxdepth 1 -type f \( -name '*.tar.gz' -o -name '*.zip' \) | LC_ALL=C sort)
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
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum --check checksums.txt
  else
    shasum -a 256 --check checksums.txt
  fi
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

host_os="$(uname -s)"
host_arch="$(uname -m)"
case "$host_arch" in
  x86_64) host_arch="amd64" ;;
  arm64 | aarch64) host_arch="arm64" ;;
  *) echo "Unsupported package smoke architecture: $host_arch" >&2; exit 1 ;;
esac
case "$host_os" in
  Linux) host_os="linux"; host_binary="soku" ;;
  Darwin) host_os="darwin"; host_binary="soku" ;;
  MINGW* | MSYS* | CYGWIN*) host_os="windows"; host_arch="amd64"; host_binary="soku.exe" ;;
  *) echo "Unsupported package smoke OS: $host_os" >&2; exit 1 ;;
esac
host_archive="$output_dir/soku_v0.1.0-test_${host_os}_${host_arch}.tar.gz"
if [[ "$host_os" == "windows" ]]; then
  host_archive="$output_dir/soku_v0.1.0-test_windows_amd64.zip"
  tar -xf "$host_archive" -C "$extract_dir"
else
  tar -xzf "$host_archive" -C "$extract_dir"
fi
if [[ ! -x "$extract_dir/$host_binary" ]]; then
  echo "Packaged native binary is not executable" >&2
  exit 1
fi
version_json="$($extract_dir/$host_binary --json --version)"
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
