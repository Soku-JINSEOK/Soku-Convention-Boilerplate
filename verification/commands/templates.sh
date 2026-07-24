#!/usr/bin/env bash
# Runtime template checks: lint/typecheck/test/build for each stack template,
# plus the gcloud Dockerfile build and the AWS/Azure placeholder YAML lint.
# Mirrors the javascript-typescript-node/python/go/java-spring/gcloud/
# aws-azure-config jobs in .github/workflows/templates-ci.yml. The mysql/
# postgresql jobs live in verification/commands/db-schema.sh instead, since
# they need docker-compose.verify.yml services running.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE="${WORKSPACE:?WORKSPACE must be set}"
# shellcheck source=verification/commands/_lib.sh
source "$SCRIPT_DIR/_lib.sh"

run_js_template() {
  local dir="$WORKSPACE/templates/javascript-typescript-node"
  [[ ! -f "$dir/package.json" ]] && return 0

  print_step "Node.js template quality gates"
  require_command npm "templates::node" 10
  (
    cd "$dir"
    run_or_fail "templates::node-ci" 11 npm ci
    run_or_fail "templates::node-lint" 12 npm run lint
    run_or_fail "templates::node-typecheck" 13 npm run typecheck
    run_or_fail "templates::node-test" 14 npm test
    run_or_fail "templates::node-build" 15 npm run build
    run_or_fail "templates::node-format-check" 16 npm run format:check
  )
  print_step_end
}

run_python_template() {
  local dir="$WORKSPACE/templates/python"
  [[ ! -f "$dir/requirements-lock.txt" ]] && return 0
  [[ ! -f "$dir/pyproject.toml" ]] && return 0

  print_step "Python template quality gates"
  require_command python3 "templates::python" 20

  local venv_dir
  venv_dir="$(mktemp -d)"
  rm -rf "$venv_dir"
  python3 -m venv "$venv_dir"

  # shellcheck disable=SC1091
  source "$venv_dir/bin/activate"
  run_or_fail "templates::python-install" 21 python -m pip install --disable-pip-version-check \
    -r "$dir/requirements-lock.txt" -e "${dir}[dev]"
  run_or_fail "templates::python-ruff" 22 ruff check "$dir"
  run_or_fail "templates::python-mypy" 23 mypy "$dir"
  run_or_fail "templates::python-black" 24 black --check "$dir"
  run_or_fail "templates::python-test" 25 pytest "$dir"
  deactivate
  rm -rf "$venv_dir"
  print_step_end
}

run_go_template() {
  local dir="$WORKSPACE/templates/go"
  [[ ! -f "$dir/go.mod" ]] && return 0

  print_step "Go template quality gates"
  require_command go "templates::go" 30
  (
    cd "$dir"
    # shellcheck disable=SC2016 # single-quoted so bash -c expands it in the subshell, not here
    run_or_fail "templates::go-fmt-check" 31 bash -c 'test -z "$(gofmt -l .)"'
    run_or_fail "templates::go-build" 32 go build ./...
    run_or_fail "templates::go-test" 33 go test ./...
    run_or_fail "templates::go-install-goimports" 34 go install "golang.org/x/tools/cmd/goimports@${GOIMPORTS_VERSION}"
    # shellcheck disable=SC2016 # single-quoted so bash -c expands it in the subshell, not here
    run_or_fail "templates::go-goimports" 35 bash -c 'test -z "$("$(go env GOPATH)/bin/goimports" -l .)"'
    run_or_fail "templates::go-install-golangci-lint" 36 go install "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}"
    run_or_fail "templates::go-golangci-lint" 37 "$(go env GOPATH)/bin/golangci-lint" run ./...
  )
  print_step_end
}

run_java_template() {
  local dir="$WORKSPACE/templates/java-spring"
  [[ ! -f "$dir/pom.xml" ]] && return 0

  print_step "Java template quality gates"
  require_command mvn "templates::java" 40
  (cd "$dir" && run_or_fail "templates::java-verify" 41 mvn -B verify)
  print_step_end
}

run_gcloud_template() {
  local dir="$WORKSPACE/templates/gcloud"
  [[ ! -f "$dir/Dockerfile" ]] && return 0

  print_step "gcloud template container build"
  require_command docker "templates::gcloud" 50
  run_or_fail "templates::gcloud-build" 51 docker build -t "verify-local-gcloud:latest" "$dir"
  run_or_fail "templates::gcloud-cleanup" 52 docker image rm -f "verify-local-gcloud:latest"
  print_step_end
}

run_aws_azure_config() {
  local aws_file="$WORKSPACE/templates/aws/buildspec.yml"
  local azure_file="$WORKSPACE/templates/azure/azure-pipelines.yml"
  [[ ! -f "$aws_file" || ! -f "$azure_file" ]] && return 0

  print_step "AWS/Azure placeholder YAML lint"
  require_command npx "templates::aws-azure" 53
  run_or_fail "templates::aws-azure-lint" 53 npx --yes "yaml-lint@${YAML_LINT_VERSION}" "$aws_file" "$azure_file"
  print_step_end
}

run_js_template
run_python_template
run_go_template
run_java_template
run_gcloud_template
run_aws_azure_config
