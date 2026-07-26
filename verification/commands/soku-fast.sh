#!/usr/bin/env bash
# Affected-package Soku checks for the fast profile. Deliberately excludes
# race, lifecycle, network, lint installation, and packaging.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE="${WORKSPACE:?WORKSPACE must be set}"
# shellcheck source=verification/commands/_lib.sh
source "$SCRIPT_DIR/_lib.sh"

SOKU_DIR="$WORKSPACE/soku"
[[ -f "$SOKU_DIR/go.mod" ]] || exit 0
require_command go "soku-fast" 68

go_files=()
packages=()
all_packages=false
while IFS= read -r -d '' path; do
  case "$path" in
    soku/**/*.go)
      [[ -e "$WORKSPACE/$path" ]] && go_files+=("${path#soku/}")
      package="./$(dirname "${path#soku/}")"
      [[ " ${packages[*]-} " == *" $package "* ]] || packages+=("$package")
      ;;
    soku/*.go | soku/go.mod | soku/go.sum)
      all_packages=true
      ;;
    soku/**)
      all_packages=true
      ;;
  esac
done < <(read_changed_files)

if [[ "$all_packages" == true ]] || ((${#packages[@]} == 0)); then
  packages=("./...")
fi

cd "$SOKU_DIR"
print_step "Soku affected-package formatting, test, and vet"
run_or_fail "soku-fast::mod-verify" 68 go mod verify
if ((${#go_files[@]})); then
  run_or_fail "soku-fast::gofmt" 68 gofmt -d "${go_files[@]}"
  if [[ -n "$(gofmt -l "${go_files[@]}")" ]]; then
    action_fail 68 "soku-fast::gofmt" "changed Go files require gofmt"
  fi
fi
run_or_fail "soku-fast::test" 68 go test "${packages[@]}"
run_or_fail "soku-fast::vet" 68 go vet "${packages[@]}"
print_step_end

print_step "Soku native build and smoke"
binary_dir="$(mktemp -d)"
trap 'rm -rf "$binary_dir"' EXIT
binary="$binary_dir/soku"
run_or_fail "soku-fast::build" 69 go build -trimpath -o "$binary" .
run_or_fail "soku-fast::smoke-help" 69 "$binary" --help
run_or_fail "soku-fast::smoke-version" 69 "$binary" --version
set +e
"$binary" status
status_exit=$?
set -e
if [[ "$status_exit" -ne 3 ]]; then
  action_fail 69 "soku-fast::smoke-status" "expected exit code 3, got $status_exit"
fi
print_step_end
