#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage: scripts/ci-local.sh [--workspace <path>] [--skip-infra]

Run local-equivalent quality gates used by CI and deployment planning.

Options:
  --workspace <path>  Repository root path to run checks in.
  --skip-infra        Skip Terraform checks under infra/.
  --help              Show this help and exit.
USAGE
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE="$SCRIPT_DIR/.."
SKIP_INFRA=false

while ((${#})); do
  case "$1" in
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

WORKSPACE="$(cd "$WORKSPACE" && pwd)"

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

run_baseline_checks() {
  local file_glob
  shopt -s nullglob
  file_glob=("$WORKSPACE/scripts"/*.sh "$WORKSPACE/soku/scripts"/*.sh)
  shopt -u nullglob

  print_step "Repository baseline checks"
  if ((${#file_glob[@]})); then
    run_or_fail "shell::bash-syntax" 70 bash -n "${file_glob[@]}"
    require_command shellcheck "shell::shellcheck" 71
    run_or_fail "shell::shellcheck" 71 shellcheck "${file_glob[@]}"
  fi
  print_step_end
}

run_js_template() {
  local dir="$WORKSPACE/templates/javascript-typescript-node"
  [[ ! -f "$dir/package.json" ]] && return 0

  print_step "Node.js template quality gates"
  require_command npm "node-template" 10

  cd "$dir"
  run_or_fail "node-template::npm-ci" 11 npm ci
  run_or_fail "node-template::lint" 12 npm run lint
  run_or_fail "node-template::typecheck" 13 npm run typecheck
  run_or_fail "node-template::test" 14 npm test
  run_or_fail "node-template::build" 15 npm run build
  run_or_fail "node-template::format-check" 16 npm run format:check
  run_or_fail "node-template::security-audit" 17 npm audit --audit-level=high
  cd - >/dev/null
  print_step_end
}

run_python_template() {
  local dir="$WORKSPACE/templates/python"
  [[ ! -f "$dir/requirements-lock.txt" ]] && return 0
  [[ ! -f "$dir/pyproject.toml" ]] && return 0

  print_step "Python template quality gates"
  require_command python3 "python-template" 20

  local venv_dir="$WORKSPACE/.tmp-ci-local-python"
  local restore_dir
  restore_dir="$(pwd)"

  rm -rf "$venv_dir"
  python3 -m venv "$venv_dir"

  # shellcheck disable=SC1091
  source "$venv_dir/bin/activate"
  run_or_fail "python-template::venv-install" 21 python -m pip install --disable-pip-version-check -r "$dir/requirements-lock.txt" -e "${dir}[dev]"
  run_or_fail "python-template::ruff" 22 ruff check "$dir"
  run_or_fail "python-template::mypy" 23 mypy "$dir"
  run_or_fail "python-template::black" 24 black --check "$dir"
  run_or_fail "python-template::test" 25 pytest "$dir"

  if command -v pip-audit >/dev/null 2>&1; then
    run_or_fail "python-template::pip-audit" 26 pip-audit --strict -r "$dir/requirements-lock.txt"
  else
    run_or_fail "python-template::pip-audit-install" 27 python -m pip install --disable-pip-version-check pip-audit==2.10.0
    run_or_fail "python-template::pip-audit" 28 pip-audit --strict -r "$dir/requirements-lock.txt"
  fi

  deactivate
  rm -rf "$venv_dir"
  cd "$restore_dir"
  print_step_end
}

run_go_template() {
  local dir="$WORKSPACE/templates/go"
  [[ ! -f "$dir/go.mod" ]] && return 0

  print_step "Go template quality gates"
  require_command go "go-template" 30

  cd "$dir"
  run_or_fail "go-template::fmt-check" 31 test -z "$(gofmt -l .)"
  run_or_fail "go-template::build" 32 go build ./...
  run_or_fail "go-template::unit-test" 33 go test ./...
  run_or_fail "go-template::install-goimports" 34 go install golang.org/x/tools/cmd/goimports@v0.29.0
  run_or_fail "go-template::goimports" 35 test -z "$("$(go env GOPATH)/bin/goimports" -l .)"
  run_or_fail "go-template::install-golangci-lint" 36 go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
  run_or_fail "go-template::golangci-lint" 37 "$(go env GOPATH)/bin/golangci-lint" run ./...
  cd - >/dev/null
  print_step_end
}

run_java_template() {
  local dir="$WORKSPACE/templates/java-spring"
  [[ ! -f "$dir/pom.xml" ]] && return 0

  print_step "Java template quality gates"
  require_command mvn "java-template" 40
  cd "$dir"
  run_or_fail "java-template::verify" 41 mvn -B verify
  cd - >/dev/null
  print_step_end
}

run_gcloud_template() {
  local dir="$WORKSPACE/templates/gcloud"
  [[ ! -f "$dir/Dockerfile" ]] && return 0

  print_step "Container build gate"
  require_command docker "container-gateway" 50
  run_or_fail "docker::build-gcloud-template" 51 docker build -t "ci-local-gcloud:latest" "$dir"
  run_or_fail "docker::cleanup" 52 docker image rm -f "ci-local-gcloud:latest"
  print_step_end
}

run_infra_checks() {
  [[ "$SKIP_INFRA" == true ]] && return 0
  local dir="$WORKSPACE/infra/gcp"
  [[ ! -d "$dir" ]] && return 0

  print_step "Terraform checks"
  require_command terraform "terraform-gate" 60
  run_or_fail "terraform::fmt" 61 terraform -chdir="$dir" fmt -check -recursive
  run_or_fail "terraform::init" 62 terraform -chdir="$dir" init -backend=false -input=false
  run_or_fail "terraform::validate" 63 terraform -chdir="$dir" validate
  print_step_end
}

run_baseline_checks
run_js_template
run_python_template
run_go_template
run_java_template
run_gcloud_template
run_infra_checks

echo "Local CI checks passed for: $WORKSPACE"
