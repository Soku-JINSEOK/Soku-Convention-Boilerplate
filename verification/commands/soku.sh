#!/usr/bin/env bash
# soku CLI checks that are reproducible on a single developer machine:
# build, vet, unit tests, race tests, formatting/lint, packaging, and the
# hermetic lifecycle gate. Mirrors soku-cross-platform (ubuntu leg),
# soku-quality, soku-package, and the lifecycle-gate half of
# soku-core-lifecycle in .github/workflows/ci.yml. The macOS/Windows legs and
# the network-conformance fixture are hosted-only (see CLASSIFICATION.md) and
# only get a notice here, never a silent skip.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE="${WORKSPACE:?WORKSPACE must be set}"
# shellcheck source=verification/commands/_lib.sh
source "$SCRIPT_DIR/_lib.sh"

SOKU_DIR="$WORKSPACE/soku"
if [[ ! -f "$SOKU_DIR/go.mod" ]]; then
  echo "soku module not found at $SOKU_DIR — skipping"
  exit 0
fi

require_command go "soku" 80
cd "$SOKU_DIR"

print_step "soku build, vet, and unit tests"
run_or_fail "soku::mod-verify" 81 go mod verify
run_or_fail "soku::test" 82 go test ./...
run_or_fail "soku::vet" 83 go vet ./...
print_step_end

print_step "soku race tests and formatting"
run_or_fail "soku::race" 84 go test -race ./...
# shellcheck disable=SC2016 # single-quoted so bash -c expands it in the subshell, not here
run_or_fail "soku::gofmt" 85 bash -c 'test -z "$(gofmt -l .)"'
run_or_fail "soku::install-goimports" 86 go install "golang.org/x/tools/cmd/goimports@${GOIMPORTS_VERSION}"
# shellcheck disable=SC2016 # single-quoted so bash -c expands it in the subshell, not here
run_or_fail "soku::goimports" 86 bash -c 'test -z "$("$(go env GOPATH)/bin/goimports" -l .)"'
run_or_fail "soku::install-golangci-lint" 87 go install "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}"
run_or_fail "soku::golangci-lint" 87 "$(go env GOPATH)/bin/golangci-lint" run ./...
print_step_end

print_step "soku build and smoke test (native binary)"
binary_dir="$(mktemp -d)"
binary="$binary_dir/soku"
run_or_fail "soku::build" 88 go build -trimpath -o "$binary" .
run_or_fail "soku::smoke-help" 88 "$binary" --help
run_or_fail "soku::smoke-version" 88 "$binary" --version
set +e
"$binary" status
status_exit=$?
set -e
if [[ "$status_exit" -ne 3 ]]; then
  action_fail 88 "soku::smoke-status" "expected exit code 3, got $status_exit"
fi
rm -rf "$binary_dir"
print_step_end

print_step "soku hermetic lifecycle conformance gate"
lifecycle_artifact_dir="$(mktemp -d)"
run_or_fail "soku::lifecycle-gate" 89 scripts/run_lifecycle_gate.sh "$lifecycle_artifact_dir"
rm -rf "$lifecycle_artifact_dir"
print_step_end

print_step "soku five-target package snapshot"
run_or_fail "soku::package-test" 90 scripts/package_test.sh
print_step_end

hosted_only "soku::cross-platform (macOS/Windows build+test+smoke)"
hosted_only "soku::network-conformance (TestProviderNetworkConformance, 3-OS matrix)"
