#!/usr/bin/env bash
# Focused syntax and lint checks for files selected by the fast profile.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE="${WORKSPACE:?WORKSPACE must be set}"
# shellcheck source=verification/commands/_lib.sh
source "$SCRIPT_DIR/_lib.sh"

cd "$WORKSPACE"

markdown_files=()
yaml_files=()
json_files=()
shell_files=()
javascript_files=()
workflow_changed=false
detector_changed=false

while IFS= read -r -d '' path; do
  [[ -e "$path" ]] || continue
  case "$path" in
    *.md) markdown_files+=("$path") ;;
    *.yml | *.yaml) yaml_files+=("$path") ;;
    *.json) json_files+=("$path") ;;
    *.sh) shell_files+=("$path") ;;
    *.js | *.mjs | *.cjs) javascript_files+=("$path") ;;
  esac
  case "$path" in
    .github/workflows/*) workflow_changed=true ;;
  esac
  case "$path" in
    verification/scopes.yml | scripts/detect-verification-scope.mjs | scripts/detect-verification-scope.test.mjs)
      detector_changed=true
      ;;
  esac
done < <(read_changed_files)

print_step "Repository baseline"
for path in README.md BLUEPRINT.md AGENTS.md CONTRIBUTING.md LICENSE; do
  run_or_fail "fast::baseline:$path" 60 test -s "$path"
done
print_step_end

if ((${#markdown_files[@]})); then
  print_step "Changed Markdown lint"
  require_command npx "fast::markdownlint" 61
  run_or_fail "fast::markdownlint" 61 \
    npx --yes "markdownlint-cli2@${MARKDOWNLINT_CLI2_VERSION}" \
    --config .markdownlint.jsonc "${markdown_files[@]}"
  print_step_end
fi

if ((${#yaml_files[@]})); then
  print_step "Changed YAML lint"
  require_command npx "fast::yaml" 62
  run_or_fail "fast::yaml" 62 \
    npx --yes "yaml-lint@${YAML_LINT_VERSION}" "${yaml_files[@]}"
  print_step_end
fi

if ((${#json_files[@]})); then
  print_step "Changed JSON syntax"
  require_command node "fast::json" 63
  for path in "${json_files[@]}"; do
    run_or_fail "fast::json:$path" 63 node -e \
      'JSON.parse(require("node:fs").readFileSync(process.argv[1], "utf8"))' "$path"
  done
  print_step_end
fi

if ((${#shell_files[@]})); then
  print_step "Changed shell syntax"
  run_or_fail "fast::bash-syntax" 64 bash -n "${shell_files[@]}"
  print_step_end
fi

if ((${#javascript_files[@]})); then
  print_step "Changed JavaScript syntax"
  require_command node "fast::javascript" 65
  for path in "${javascript_files[@]}"; do
    run_or_fail "fast::javascript:$path" 65 node --check "$path"
  done
  print_step_end
fi

if [[ "$workflow_changed" == true ]]; then
  print_step "Changed GitHub Actions syntax"
  require_command go "fast::actionlint" 66
  run_or_fail "fast::actionlint" 66 \
    go run "github.com/rhysd/actionlint/cmd/actionlint@${ACTIONLINT_VERSION}" \
    .github/workflows/*.yml
  print_step_end
fi

if [[ "$detector_changed" == true ]]; then
  print_step "Scope-detector regression tests"
  run_or_fail "fast::scope-detector" 67 \
    node --test scripts/detect-verification-scope.test.mjs
  print_step_end
fi
