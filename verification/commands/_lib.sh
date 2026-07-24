#!/usr/bin/env bash
# Shared helpers for verification/commands/*.sh. Sourced, not executed
# directly. Mirrors the step/failure reporting scripts/ci-local.sh already
# used, so output stays familiar across both entry points.

: "${WORKSPACE:?WORKSPACE must be set by the caller before sourcing _lib.sh}"

# shellcheck source=verification/tools.env disable=SC1091
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/tools.env"

print_step() {
  echo "::group::$1"
}

print_step_end() {
  echo "::endgroup::"
}

action_fail() {
  local code="$1"
  local step="$2"
  shift 2
  echo "::error::[$step] $*" >&2
  exit "$code"
}

run_or_fail() {
  local step="$1"
  local code="$2"
  shift 2

  if ! "$@"; then
    action_fail "$code" "$step" "command failed: $*"
  fi
}

require_command() {
  local cmd="$1"
  local step="$2"
  local code="$3"

  if ! command -v "$cmd" >/dev/null 2>&1; then
    action_fail "$code" "$step" "required command '$cmd' is missing"
  fi
}

hosted_only() {
  local label="$1"
  echo "::notice::[$label] hosted-only — skipped, not a pass. See verification/CLASSIFICATION.md."
}
