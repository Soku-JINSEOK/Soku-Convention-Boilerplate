#!/usr/bin/env bash
# Repository hygiene checks: shell syntax/lint, Markdown/YAML/Actions
# linting, and the repo's own node/python regression test suites. Mirrors
# the repository-hygiene job in .github/workflows/ci.yml (excluding the
# baseline-file-existence check, which is specific to this boilerplate's own
# layout and adds no value as a "verification" concept).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE="${WORKSPACE:?WORKSPACE must be set}"
# shellcheck source=verification/commands/_lib.sh
source "$SCRIPT_DIR/_lib.sh"

cd "$WORKSPACE"

print_step "Shell syntax and shellcheck"
shopt -s nullglob
shell_files=(scripts/*.sh soku/scripts/*.sh verification/commands/*.sh)
shopt -u nullglob
if ((${#shell_files[@]})); then
  run_or_fail "shell::bash-syntax" 70 bash -n "${shell_files[@]}"
  require_command shellcheck "shell::shellcheck" 71
  run_or_fail "shell::shellcheck" 71 shellcheck "${shell_files[@]}"
fi
print_step_end

print_step "Regression test suites"
require_command node "regression::node" 72
run_or_fail "regression::contribution-title" 72 node --test \
  templates/_shared/commitlint/contribution-title.test.mjs \
  scripts/contribution-title.test.mjs \
  scripts/pull-request-policy.test.mjs
run_or_fail "regression::pr-governance" 72 node --test \
  .github/dependabot.test.mjs \
  .github/validate-pr-governance.test.mjs \
  .github/issue-form-order.test.mjs \
  .github/validation-workflow.test.mjs \
  .github/deploy-gcp.test.mjs \
  .github/usage-manual.test.mjs \
  .github/github-governance-audit.test.mjs
run_or_fail "regression::npm-wrapper-unit" 72 node --test soku/npm/lib/launcher.test.mjs
run_or_fail "regression::release-tag-gate" 72 scripts/verify-release-tag_test.sh

if command -v python3 >/dev/null 2>&1; then
  run_or_fail "regression::public-provider-action" 72 python3 \
    soku/actions/ci-cd-control-plane-v1/test_validate_config.py
fi
print_step_end

print_step "npm wrapper package tests"
require_command npm "regression::npm-wrapper-package" 72
(cd soku/npm && run_or_fail "regression::npm-wrapper-package" 72 npm test)
print_step_end

print_step "Markdown, YAML, and GitHub Actions linting"
require_command npx "lint::npx" 73
run_or_fail "lint::markdownlint" 73 npx --yes "markdownlint-cli2@${MARKDOWNLINT_CLI2_VERSION}" \
  --config .markdownlint.jsonc "**/*.md" "#**/node_modules/**"
run_or_fail "lint::yaml" 73 npx --yes "yaml-lint@${YAML_LINT_VERSION}" \
  .github/*.yml .github/**/*.yml docs/callers/*.yml \
  soku/actions/**/*.yml soku/providers/**/*.yml
require_command go "lint::actionlint" 73
run_or_fail "lint::actionlint" 73 go run "github.com/rhysd/actionlint/cmd/actionlint@${ACTIONLINT_VERSION}" \
  .github/workflows/*.yml
print_step_end
