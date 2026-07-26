#!/usr/bin/env bash
# Dependency, vulnerability, secret, and license checks. Mirrors the
# secrets/dependencies/go-vulnerabilities/osv jobs in
# .github/workflows/security.yml, using the versions and thresholds pinned
# in verification/tools.env (this closes the npm-audit-level drift recorded
# in CLASSIFICATION.md). The gitleaks scan here is best-effort against the
# current checkout; the weekly full-history scheduled scan remains the
# authoritative hosted-only check.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE="${WORKSPACE:?WORKSPACE must be set}"
# shellcheck source=verification/commands/_lib.sh
source "$SCRIPT_DIR/_lib.sh"

cd "$WORKSPACE"
require_command go "security" 110

PIP_AUDIT_VENV=""
cleanup() {
  [[ -z "$PIP_AUDIT_VENV" ]] || rm -rf "$PIP_AUDIT_VENV"
}
trap cleanup EXIT

print_step "Gitleaks secret scan (working tree, best-effort)"
run_or_fail "security::gitleaks" 111 go run "github.com/zricethezav/gitleaks/v8@${GITLEAKS_VERSION}" \
  detect --source . --redact --no-banner
print_step_end
hosted_only "security::gitleaks-full-history (weekly scheduled scan)"

js_dir="$WORKSPACE/templates/javascript-typescript-node"
if [[ -f "$js_dir/package.json" ]]; then
  print_step "JavaScript template dependency audit"
  require_command npm "security::npm-audit" 112
  (cd "$js_dir" && run_or_fail "security::npm-audit" 112 npm audit "--audit-level=${NPM_AUDIT_LEVEL}")
  print_step_end
fi

python_dir="$WORKSPACE/templates/python"
if [[ -f "$python_dir/requirements-lock.txt" ]]; then
  print_step "Python template dependency audit"
  pip_audit_command=()
  if ! command -v pip-audit >/dev/null 2>&1; then
    require_command python3 "security::pip-audit-install" 113
    PIP_AUDIT_VENV="$(mktemp -d)"
    rm -rf "$PIP_AUDIT_VENV"
    python3 -m venv "$PIP_AUDIT_VENV"
    if [[ -x "$PIP_AUDIT_VENV/bin/python" ]]; then
      pip_audit_python="$PIP_AUDIT_VENV/bin/python"
      pip_audit_command=("$PIP_AUDIT_VENV/bin/pip-audit")
    else
      pip_audit_python="$PIP_AUDIT_VENV/Scripts/python.exe"
      pip_audit_command=("$PIP_AUDIT_VENV/Scripts/pip-audit.exe")
    fi
    run_or_fail "security::pip-audit-install" 113 \
      "$pip_audit_python" -m pip install --disable-pip-version-check \
      "pip-audit==${PIP_AUDIT_VERSION}"
  else
    pip_audit_command=(pip-audit)
  fi
  run_or_fail "security::pip-audit" 113 \
    "${pip_audit_command[@]}" --strict -r "$python_dir/requirements-lock.txt"
  cleanup
  PIP_AUDIT_VENV=""
  print_step_end
fi

print_step "Declared license and notice audit"
run_or_fail "security::license-file" 114 test -s LICENSE
run_or_fail "security::third-party-notices" 114 test -s soku/THIRD_PARTY_NOTICES.md
run_or_fail "security::third-party-notices-content" 114 grep -Eq \
  'MIT|Apache|BSD|third[- ]party' soku/THIRD_PARTY_NOTICES.md
print_step_end

print_step "Go vulnerability checks (govulncheck)"
for module in soku templates/go; do
  [[ -f "$WORKSPACE/$module/go.mod" ]] || continue
  (cd "$WORKSPACE/$module" && run_or_fail "security::govulncheck:$module" 115 \
    go run "golang.org/x/vuln/cmd/govulncheck@${GOVULNCHECK_VERSION}" ./...)
done
print_step_end

print_step "OSV dependency scan"
run_or_fail "security::osv-scanner" 116 go run "github.com/google/osv-scanner/v2/cmd/osv-scanner@${OSV_SCANNER_VERSION}" \
  scan source -r . --experimental-exclude .git --no-resolve
print_step_end
