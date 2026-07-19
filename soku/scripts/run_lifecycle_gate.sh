#!/usr/bin/env bash

set -euo pipefail

artifact_directory="${1:?usage: run_lifecycle_gate.sh <artifact-directory>}"
script_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$script_directory/.."
mkdir -p "$artifact_directory"
raw_log="$artifact_directory/raw.json"
sanitized_log="$artifact_directory/core-lifecycle.log"

set +e
go test -json ./internal/lifecyclee2e >"$raw_log" 2>&1
test_exit=$?
set -e

sed_arguments=()
add_sanitized_root() {
  local value="$1"
  local label="$2"
  if [[ -z "$value" ]]; then
    return
  fi
  value="${value//\\//}"
  value="${value//#/\\#}"
  value="${value//&/\\&}"
  sed_arguments+=("-e" "s#$value#<$label>#g")
}

add_sanitized_root "${GITHUB_WORKSPACE:-}" workspace
add_sanitized_root "${RUNNER_TEMP:-}" temporary
add_sanitized_root "${HOME:-}" home
sed_arguments+=("-E" "-e" 's#[A-Za-z]:/[^"[:space:]]+#<path>#g')
sed_arguments+=("-e" 's#(/[^/"[:space:]]+)+#<path>#g')

tr '\\' '/' <"$raw_log" | sed "${sed_arguments[@]}" >"$sanitized_log"
rm -f "$raw_log"

if [[ "$test_exit" -ne 0 ]]; then
  tail -n 200 "$sanitized_log"
  exit "$test_exit"
fi

echo "Core lifecycle gate passed."
