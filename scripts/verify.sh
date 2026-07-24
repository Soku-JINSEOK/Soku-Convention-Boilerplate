#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage: scripts/verify.sh --profile <name> [--workspace <path>] [--skip-infra] [--skip-db]

Run the verification profile that matches the moment in the development
lifecycle you're at. See verification/CLASSIFICATION.md for what each group
of checks covers and why.

Profiles:
  full    Every locally-reproducible check: repo hygiene, soku, templates,
          db schema (via docker-compose.verify.yml), security, and infra.
          This is the only profile implemented so far (issue #112, phase
          1-2). Other profile names are recognized but not yet implemented —
          they fail loudly instead of silently running a subset.

Planned, not yet implemented: fast, ci-quick, hosted-full, release, deploy.

Options:
  --profile <name>    Verification profile to run. Required.
  --workspace <path>  Repository root path to run checks in. Defaults to the
                      repository containing this script.
  --skip-infra        Skip Terraform checks under infra/.
  --skip-db           Skip the docker-compose-backed MySQL/PostgreSQL schema
                      checks.
  --help              Show this help and exit.
USAGE
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE="$SCRIPT_DIR/.."
PROFILE=""
SKIP_INFRA=false
SKIP_DB=false

while ((${#})); do
  case "$1" in
    --profile)
      if [[ "${2-}" == "" ]]; then
        echo "Missing value for --profile" >&2
        usage
        exit 2
      fi
      PROFILE="$2"
      shift 2
      ;;
    --workspace)
      if [[ "${2-}" == "" ]]; then
        echo "Missing value for --workspace" >&2
        usage
        exit 2
      fi
      WORKSPACE="$2"
      shift 2
      ;;
    --skip-infra)
      SKIP_INFRA=true
      shift
      ;;
    --skip-db)
      SKIP_DB=true
      shift
      ;;
    --help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage
      exit 2
      ;;
  esac
done

if [[ -z "$PROFILE" ]]; then
  echo "::error::--profile is required" >&2
  usage
  exit 2
fi

WORKSPACE="$(cd "$WORKSPACE" && pwd)"
export WORKSPACE
COMMANDS_DIR="$SCRIPT_DIR/../verification/commands"

run_infra_checks() {
  [[ "$SKIP_INFRA" == true ]] && return 0
  local dir="$WORKSPACE/infra/gcp"
  [[ ! -d "$dir" ]] && return 0
  command -v terraform >/dev/null 2>&1 || return 0

  echo "::group::Terraform checks"
  terraform -chdir="$dir" fmt -check -recursive
  terraform -chdir="$dir" init -backend=false -input=false
  terraform -chdir="$dir" validate
  echo "::endgroup::"
}

run_full_profile() {
  "$COMMANDS_DIR/repo-hygiene.sh"
  "$COMMANDS_DIR/soku.sh"
  "$COMMANDS_DIR/templates.sh"
  if [[ "$SKIP_DB" == true ]]; then
    echo "::notice::db-schema — skipped via --skip-db"
  else
    "$COMMANDS_DIR/db-schema.sh"
  fi
  "$COMMANDS_DIR/security.sh"
  run_infra_checks
}

case "$PROFILE" in
  full)
    run_full_profile
    ;;
  fast | ci-quick | hosted-full | release | deploy)
    echo "::error::profile '$PROFILE' is not yet implemented (see verification/CLASSIFICATION.md and issue #112's phased rollout)" >&2
    exit 3
    ;;
  *)
    echo "::error::unknown profile '$PROFILE'" >&2
    usage
    exit 2
    ;;
esac

echo "Verification profile '$PROFILE' passed for: $WORKSPACE"
