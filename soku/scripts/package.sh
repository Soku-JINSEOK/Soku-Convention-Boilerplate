#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "Usage: $0 --version <vX.Y.Z> --commit <sha> --built-at <rfc3339> --output-dir <path>" >&2
}

version=""
commit=""
built_at=""
output_dir=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)
      version="${2:-}"
      shift 2
      ;;
    --commit)
      commit="${2:-}"
      shift 2
      ;;
    --built-at)
      built_at="${2:-}"
      shift 2
      ;;
    --output-dir)
      output_dir="${2:-}"
      shift 2
      ;;
    *)
      usage
      exit 2
      ;;
  esac
done

if [[ -z "$version" || -z "$commit" || -z "$built_at" || -z "$output_dir" ]]; then
  usage
  exit 2
fi
if [[ ! "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.-]+)?$ ]]; then
  echo "Error: --version must be an immutable vX.Y.Z-style version" >&2
  exit 2
fi
if [[ ! "$commit" =~ ^[0-9a-f]{40}$ ]]; then
  echo "Error: --commit must be a lowercase 40-character SHA" >&2
  exit 2
fi
if [[ ! "$built_at" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$ ]]; then
  echo "Error: --built-at must be an RFC 3339 UTC timestamp" >&2
  exit 2
fi

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
module_dir="$(cd "$script_dir/.." && pwd)"
repository_dir="$(cd "$module_dir/.." && pwd)"
mkdir -p "$output_dir"
output_dir="$(cd "$output_dir" && pwd)"
work_dir="$(mktemp -d)"
cleanup() {
  rm -rf "$work_dir"
}
trap cleanup EXIT

readonly metadata_package="github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/cli"
readonly ldflags="-s -w -X ${metadata_package}.Version=${version} -X ${metadata_package}.Commit=${commit} -X ${metadata_package}.BuiltAt=${built_at}"
readonly targets=(
  "linux amd64"
  "linux arm64"
  "darwin amd64"
  "darwin arm64"
  "windows amd64"
)

archives=()
for target in "${targets[@]}"; do
  read -r target_os target_arch <<<"$target"
  stage_dir="$work_dir/${target_os}_${target_arch}"
  mkdir -p "$stage_dir"
  binary_name="soku"
  if [[ "$target_os" == "windows" ]]; then
    binary_name="soku.exe"
  fi

  (
    cd "$module_dir"
    CGO_ENABLED=0 GOOS="$target_os" GOARCH="$target_arch" \
      go build -trimpath -ldflags "$ldflags" -o "$stage_dir/$binary_name" .
  )
  cp "$repository_dir/LICENSE" "$stage_dir/LICENSE"
  cp "$module_dir/THIRD_PARTY_NOTICES.md" "$stage_dir/THIRD_PARTY_NOTICES.md"
  touch -t 198001010000 "$stage_dir"/*

  archive_base="soku_${version}_${target_os}_${target_arch}"
  if [[ "$target_os" == "windows" ]]; then
    archive_name="${archive_base}.zip"
    (
      cd "$module_dir"
      go run ./internal/packagezip \
        --source "$stage_dir" \
        --output "$output_dir/$archive_name" \
        --binary "$binary_name"
    )
  else
    archive_name="${archive_base}.tar.gz"
    (
      cd "$stage_dir"
      tar --sort=name --mtime='@315532800' --owner=0 --group=0 --numeric-owner \
        -cf - LICENSE THIRD_PARTY_NOTICES.md "$binary_name" | gzip -n >"$output_dir/$archive_name"
    )
  fi
  archives+=("$archive_name")
done

(
  cd "$output_dir"
  printf '%s\n' "${archives[@]}" | LC_ALL=C sort | xargs sha256sum >checksums.txt
)
