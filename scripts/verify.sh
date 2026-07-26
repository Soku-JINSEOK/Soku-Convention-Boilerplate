#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage: scripts/verify.sh --profile <name> [options]

Run the verification profile that matches the moment in the development
lifecycle you're at. See verification/CLASSIFICATION.md for what each group
of checks covers and why.

Profiles:
  fast    Staged or explicitly supplied changed-file checks: whitespace,
          changed-line secrets, focused repository hygiene, and selected
          Soku/template/DB/config/infra smoke checks.
  full    Every locally-reproducible check: repo hygiene, soku, templates,
          db schema (via docker-compose.verify.yml), security, and infra.

Planned, not yet implemented: ci-quick, hosted-full, release, deploy.

Options:
  --profile <name>    Verification profile to run. Required.
  --workspace <path>  Repository root path to run checks in. Defaults to the
                      repository containing this script.
  --staged            For fast, inspect staged changes (default).
  --base <sha>        For fast, base commit SHA. Requires --head.
  --head <sha>        For fast, head commit SHA. Requires --base.
  --files-from <path|->  For fast, read changed paths from a file or stdin.
  --skip-infra        Skip Terraform checks under infra/.
  --skip-db           Skip the docker-compose-backed MySQL/PostgreSQL schema
                      checks.
  --write-local-report  After a passing full run, write the ignored,
                      non-authoritative .soku/verification/local-full.json.
  --help              Show this help and exit.
USAGE
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE="$SCRIPT_DIR/.."
PROFILE=""
SKIP_INFRA=false
SKIP_DB=false
WRITE_LOCAL_REPORT=false
INPUT_MODE="staged"
BASE_SHA=""
HEAD_SHA=""
FILES_FROM=""
EXPLICIT_INPUT=false

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
    --staged)
      if [[ "$EXPLICIT_INPUT" == true ]]; then
        echo "Changed-file input modes are mutually exclusive" >&2
        exit 2
      fi
      EXPLICIT_INPUT=true
      INPUT_MODE="staged"
      shift
      ;;
    --base)
      if [[ "${2-}" == "" || "$FILES_FROM" != "" ]]; then
        echo "Invalid or conflicting --base" >&2
        exit 2
      fi
      EXPLICIT_INPUT=true
      INPUT_MODE="range"
      BASE_SHA="$2"
      shift 2
      ;;
    --head)
      if [[ "${2-}" == "" || "$FILES_FROM" != "" ]]; then
        echo "Invalid or conflicting --head" >&2
        exit 2
      fi
      EXPLICIT_INPUT=true
      INPUT_MODE="range"
      HEAD_SHA="$2"
      shift 2
      ;;
    --files-from)
      if [[ "${2-}" == "" || "$EXPLICIT_INPUT" == true ]]; then
        echo "Invalid or conflicting --files-from" >&2
        exit 2
      fi
      EXPLICIT_INPUT=true
      INPUT_MODE="files"
      FILES_FROM="$2"
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
    --write-local-report)
      WRITE_LOCAL_REPORT=true
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
PROFILES_FILE="$SCRIPT_DIR/../verification/profiles.yml"
# shellcheck source=verification/commands/_lib.sh
source "$COMMANDS_DIR/_lib.sh"

if [[ "$INPUT_MODE" == "range" && ( -z "$BASE_SHA" || -z "$HEAD_SHA" ) ]]; then
  echo "::error::--base and --head must be provided together" >&2
  exit 2
fi

if [[ "$PROFILE" == "fast" || "$PROFILE" == "full" ]]; then
  # shellcheck disable=SC2016 # JavaScript template literals expand in Node.
  node -e '
    const {readFileSync} = require("node:fs");
    const config = JSON.parse(readFileSync(process.argv[1], "utf8"));
    if (config.schemaVersion !== 1 || config.format !== "json-compatible-yaml") {
      throw new Error("unsupported verification profile schema");
    }
    if (!config.profiles[process.argv[2]]) {
      throw new Error(`profile is missing from verification/profiles.yml: ${process.argv[2]}`);
    }
  ' "$PROFILES_FILE" "$PROFILE"
fi

run_infra_checks() {
  local required="${1-false}"
  if [[ "$SKIP_INFRA" == true ]]; then
    echo "::notice::Terraform checks skipped via --skip-infra"
    return 0
  fi
  local dir="$WORKSPACE/infra/gcp"
  [[ ! -d "$dir" ]] && return 0
  if ! command -v terraform >/dev/null 2>&1; then
    if [[ "$required" == true ]]; then
      echo "::error::Terraform is required for the selected infra-gcp scope" >&2
      return 70
    fi
    echo "::notice::Terraform is unavailable — skipped, not a pass"
    return 0
  fi

  echo "::group::Terraform checks"
  terraform -chdir="$dir" fmt -check -recursive
  terraform -chdir="$dir" init -backend=false -input=false
  terraform -chdir="$dir" validate
  echo "::endgroup::"
}

write_local_report() {
  local report="$WORKSPACE/.soku/verification/local-full.json"
  local commit_sha
  commit_sha="$(git -C "$WORKSPACE" rev-parse HEAD)"
  # shellcheck disable=SC2016 # JavaScript template literals expand in Node.
  node -e '
    const {mkdirSync, writeFileSync} = require("node:fs");
    const {dirname} = require("node:path");
    const [path, commitSha] = process.argv.slice(1);
    mkdirSync(dirname(path), {recursive: true});
    writeFileSync(path, `${JSON.stringify({
      schemaVersion: 1,
      profile: "full",
      status: "passed",
      authoritative: false,
      commitSha,
      createdAt: new Date().toISOString(),
    }, null, 2)}\n`);
  ' "$report" "$commit_sha"
  echo "::notice::Wrote non-authoritative local report: $report"
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
  if [[ "$WRITE_LOCAL_REPORT" == true ]]; then
    write_local_report
  fi
}

run_fast_profile() {
  local temporary_dir scope_json diff_file
  temporary_dir="$(mktemp -d)"
  scope_json="$temporary_dir/scope.json"
  diff_file="$temporary_dir/changes.diff"
  trap 'rm -rf "$temporary_dir"' RETURN

  local scope_arguments=(--workspace "$WORKSPACE" --json)
  case "$INPUT_MODE" in
    staged)
      scope_arguments+=(--staged)
      ;;
    range)
      scope_arguments+=(--base "$BASE_SHA" --head "$HEAD_SHA")
      ;;
    files)
      scope_arguments+=(--files-from "$FILES_FROM")
      ;;
  esac

  if ! node "$SCRIPT_DIR/detect-verification-scope.mjs" \
    "${scope_arguments[@]}" >"$scope_json"; then
    echo "::error::Changed-file scope detection failed" >&2
    return 58
  fi
  # shellcheck disable=SC2016 # JavaScript template literals expand in Node.
  node -e '
    const {readFileSync} = require("node:fs");
    const profiles = JSON.parse(readFileSync(process.argv[1], "utf8"));
    const result = JSON.parse(readFileSync(process.argv[2], "utf8"));
    const allowed = new Set(profiles.profiles.fast.scopes);
    const unexpected = result.scopes.filter((scope) => !allowed.has(scope));
    if (unexpected.length > 0) {
      throw new Error(`detector selected scopes outside the fast profile: ${unexpected.join(", ")}`);
    }
  ' "$PROFILES_FILE" "$scope_json"

  VERIFICATION_SCOPES="$(node -e '
    const {readFileSync} = require("node:fs");
    const result = JSON.parse(readFileSync(process.argv[1], "utf8"));
    process.stdout.write(result.scopes.join(" "));
  ' "$scope_json")"
  export VERIFICATION_SCOPES
  export VERIFICATION_SCOPES_SET=true
  export VERIFICATION_SCOPE_JSON="$scope_json"

  # shellcheck disable=SC2016 # JavaScript template literals expand in Node.
  node -e '
    const {readFileSync} = require("node:fs");
    const result = JSON.parse(readFileSync(process.argv[1], "utf8"));
    console.log(`Changed files: ${result.changedFiles.length}`);
    console.log(`Selected scopes: ${result.scopes.join(", ") || "(always-on only)"}`);
    console.log(`All selected: ${result.allSelected}`);
  ' "$scope_json"

  echo "::group::Diff whitespace"
  case "$INPUT_MODE" in
    staged)
      git -C "$WORKSPACE" diff --cached --check --
      git -C "$WORKSPACE" diff --cached --no-ext-diff --unified=0 -- >"$diff_file"
      ;;
    range)
      git -C "$WORKSPACE" diff --check "$BASE_SHA" "$HEAD_SHA" --
      git -C "$WORKSPACE" diff --no-ext-diff --unified=0 \
        "$BASE_SHA" "$HEAD_SHA" -- >"$diff_file"
      ;;
    files)
      git -C "$WORKSPACE" diff --check HEAD --
      git -C "$WORKSPACE" diff --no-ext-diff --unified=0 HEAD -- >"$diff_file"
      ;;
  esac
  echo "::endgroup::"

  echo "::group::Changed-line secret scan"
  node "$SCRIPT_DIR/scan-diff-secrets.mjs" --diff-file "$diff_file"
  echo "::endgroup::"

  "$COMMANDS_DIR/fast-repository.sh"

  if scope_selected "soku"; then
    "$COMMANDS_DIR/soku-fast.sh"
  fi

  if [[ " $VERIFICATION_SCOPES " == *" javascript-typescript-node "* ||
        " $VERIFICATION_SCOPES " == *" python "* ||
        " $VERIFICATION_SCOPES " == *" go "* ||
        " $VERIFICATION_SCOPES " == *" java-spring "* ||
        " $VERIFICATION_SCOPES " == *" gcloud "* ||
        " $VERIFICATION_SCOPES " == *" cloud-config "* ]]; then
    "$COMMANDS_DIR/templates.sh"
  fi

  if [[ " $VERIFICATION_SCOPES " == *" mysql "* ||
        " $VERIFICATION_SCOPES " == *" postgresql "* ]]; then
    if [[ "$SKIP_DB" == true ]]; then
      echo "::notice::DB schema checks skipped via --skip-db"
    else
      "$COMMANDS_DIR/db-schema.sh"
    fi
  fi

  if scope_selected "infra-gcp"; then
    run_infra_checks true
  fi
}

case "$PROFILE" in
  fast)
    run_fast_profile
    ;;
  full)
    if [[ "$EXPLICIT_INPUT" == true ]]; then
      echo "::error::changed-file inputs are only valid with --profile fast" >&2
      exit 2
    fi
    run_full_profile
    ;;
  ci-quick | hosted-full | release | deploy)
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
