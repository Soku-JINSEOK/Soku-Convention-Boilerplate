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

scope_selected() {
  local requested="$1"
  [[ "${VERIFICATION_SCOPES_SET-false}" != true ]] && return 0
  case " ${VERIFICATION_SCOPES} " in
    *" ${requested} "*) return 0 ;;
    *) return 1 ;;
  esac
}

read_changed_files() {
  : "${VERIFICATION_SCOPE_JSON:?VERIFICATION_SCOPE_JSON must be set}"
  # shellcheck disable=SC2016 # JavaScript template literals expand in Node.
  node -e '
    const {readFileSync} = require("node:fs");
    const result = JSON.parse(readFileSync(process.argv[1], "utf8"));
    for (const path of result.changedFiles) process.stdout.write(`${path}\0`);
  ' "$VERIFICATION_SCOPE_JSON"
}
