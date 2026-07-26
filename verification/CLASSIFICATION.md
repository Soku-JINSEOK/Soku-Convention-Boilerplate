# Verification Check Classification

This is the responsibility inventory for
[issue #112](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/112).
It tracks the implemented fast, sharded Quick, Hosted Full, release, and
digest-only deployment boundaries. Required-context migration remains gated by
the Issue #116 observation audit.

## Categories

- **local-capable** — a developer can run this on their own machine today (or
  after the tooling added in this phase) and get the same result CI would.
- **hosted-only** — needs something a laptop can't reliably provide: another
  OS, full repository history, live PR/GitHub API context, or a scheduled
  external service. Never claim this passed locally.
- **release-only** — only runs when cutting a tag.
- **deployment-only** — only runs during a Cloud Run deploy/rollback.

`scripts/verify.sh --profile full` (added later in this phase) runs every
`local-capable` row below. Rows marked `hosted-only` print an explicit
"hosted-only — skipped, not a pass" notice instead of being silently omitted.

## Fast Profile and Scope Selection

Issue #115 adds `scripts/verify.sh --profile fast` without changing any hosted
or required gate. The detector reads `verification/scopes.yml` and accepts
staged changes, an explicit base/head SHA range, or file/name-status input.

Always-on fast checks:

- diff whitespace;
- high-confidence secret patterns in added lines, with values redacted;
- repository baseline and changed-file syntax/lint.

Selected checks:

- affected Soku package formatting, unit test, vet, native build, and smoke;
- selected JavaScript/TypeScript, Python, Go, or Java template
  lint/test/build;
- selected MySQL/PostgreSQL schema load;
- selected GCP container or AWS/Azure syntax;
- selected GCP Terraform validation.

Fast deliberately excludes race, the complete lifecycle gate, network and OS
matrices, five-platform packaging, and the complete dependency/vulnerability
bundle. Shared verification/generator/sync/workflow/lint/provider/catalog/schema
paths and unknown paths select every scope.

## CI Quick Profile

Issue #116 adds `scripts/verify.sh --profile ci-quick` as the hosted,
fail-closed counterpart to `fast`. It requires an explicit base/head range,
does not allow DB or infrastructure skips, and runs on every event handled by
`validation.yml`. A detector-driven planner schedules independent always-on,
Soku, language-template, database, cloud, and infrastructure groups from
`verification/profiles.yml`. Its `CI Quick Gate` aggregate remains non-required
while the existing full gate runs in parallel for at least 14 days and 10
code-changing pull requests.

## `full-validation.yml` (Hosted Full)

Repository, template, and security workflows receive one exact `source-sha`.
Daily `02:41 UTC`, manual, and reusable invocations aggregate into
`Hosted Full Gate`; failure, cancellation, or an unexpected skip in any
component fails the aggregate.

## `ci.yml` (Repository CI)

| Check | Job | Command | Category |
| --- | --- | --- | --- |
| Baseline file existence | `repository-hygiene` | `test -f` loop over required files | local-capable |
| Contribution-title / PR-governance / npm-wrapper / provider-action / release-tag regression tests | `repository-hygiene` | `node --test ...`, `python3 ...`, `scripts/verify-release-tag_test.sh` | local-capable |
| Markdown lint | `repository-hygiene` | `npx markdownlint-cli2@0.22.1` | local-capable |
| YAML lint | `repository-hygiene` | `npx yaml-lint@1.7.0` | local-capable |
| GitHub Actions semantics | `repository-hygiene` | `actionlint@v1.7.10` | local-capable |
| Shell syntax + shellcheck | `repository-hygiene` | `bash -n`, `shellcheck` | local-capable (already in `ci-local.sh`) |
| PowerShell sync-script parse | `sync-parity` | `pwsh` `Parser::ParseFile` | local-capable, needs `pwsh` installed |
| sh/ps1 sync parity | `sync-parity` | `scripts/verify-sync-parity.sh` | local-capable |
| `soku` build/vet/test on Linux | `soku-cross-platform` (ubuntu leg) | `go mod verify`, `go test ./...`, `go vet ./...`, build+smoke | local-capable |
| `soku` build/vet/test on macOS/Windows | `soku-cross-platform` (macos/windows legs) | same as above | **hosted-only** (needs non-native OS) |
| `soku` race tests, gofmt, goimports, golangci-lint | `soku-quality` | `go test -race ./...`, `gofmt -l`, `goimports@v0.48.0`, `golangci-lint@v2.12.2` | local-capable — **not yet in `ci-local.sh`** |
| `soku` lifecycle conformance gate | `soku-core-lifecycle` (script itself) | `scripts/run_lifecycle_gate.sh` | local-capable |
| `soku` network-conformance fixture + 3-OS matrix | `soku-core-lifecycle` | `go test -run '^TestProviderNetworkConformance$'` w/ `GITHUB_TOKEN`, on 3 OSes | **hosted-only** (live network fixture + non-native OS) |
| `soku` 5-target package snapshot | `soku-package` | `soku/scripts/package_test.sh` | local-capable (pure Go cross-compile, no OS dependency) — **not yet in `ci-local.sh`** |

## `templates-ci.yml` (Templates CI)

| Check | Job | Category |
| --- | --- | --- |
| JS/TS lint, typecheck, test, build, format | `javascript-typescript-node` | local-capable (in `ci-local.sh`) |
| Python ruff/mypy/black/pytest | `python` | local-capable (in `ci-local.sh`) |
| Go goimports/golangci-lint/fmt/lint/test/build | `go` | local-capable (in `ci-local.sh`) |
| Java `mvn -B verify` | `java-spring` | local-capable (in `ci-local.sh`) |
| MySQL schema load | `mysql` | local-capable through `docker-compose.verify.yml` |
| PostgreSQL schema load | `postgresql` | local-capable through `docker-compose.verify.yml` |
| gcloud Dockerfile build | `gcloud` | local-capable (in `ci-local.sh`) |
| AWS/Azure placeholder YAML lint | `aws-azure-config` | local-capable, trivial |

## `security.yml` (Security)

| Check | Job | Category |
| --- | --- | --- |
| Gitleaks full-history secret scan | `secrets` | local-capable in principle (`gitleaks detect --source .`); treated as **hosted-only** for the weekly/scheduled full-history guarantee — local runs are a best-effort supplement, not a replacement |
| npm audit / pip-audit / license file checks | `dependencies` | local-capable — **drift**: `ci-local.sh` used `--audit-level=high`, this workflow uses `--audit-level=low` (resolved in this phase, see `tools.env`) |
| `govulncheck` (soku, templates/go) | `go-vulnerabilities` | local-capable — not yet in `ci-local.sh` |
| OSV scanner | `osv` | local-capable — not yet in `ci-local.sh` |

## `contribution-title-check.yml` / `pull-request-policy.yml`

| Check | Category |
| --- | --- |
| PR title/commit-title validation against live PR metadata | **hosted-only** (needs `gh api` PR context) |
| PR governance policy (labels, assignee, changed-files) against live PR metadata | **hosted-only** (needs `gh api` PR context) |

## `release.yml`

All jobs (exact-tag Hosted Full, tag/signature verification, GPG-signed notes
check, packaging + `gh release create`, `npm publish --provenance`) are
**release-only** — they only run on a tag push or release dry-run dispatch.

## `deploy-gcp.yml`

| Check | Operation | Category |
| --- | --- | --- |
| `bash -n` + `deploy-gcp.test.mjs` | `check` | local-capable |
| Successful canonical main Validation run + manifest verification | `deploy` | **deployment-only** |
| Digest-only plan, Cloud Run deploy, traffic and health verification | `deploy` | **deployment-only** |
| Rollback to previous revision | `rollback` | **deployment-only** |

## Known drift resolved by `verification/tools.env` (this phase)

- `npm audit --audit-level`: `ci-local.sh` used `high`, `security.yml` uses
  `low`. Unified to `low` (the stricter of the two) in `tools.env`.
- `goimports` version: local verification, `ci.yml`, and the generated
  `templates-ci.yml` workflow now agree on reviewed version `v0.48.0`.
- MySQL `8.4.10`, PostgreSQL `16.14`, and Alpine `3.21.7` image inputs use
  reviewed multi-architecture manifest digests. `verification/tools.env`
  remains authoritative for local database verification; the supply-chain
  verifier enforces parity where GitHub Actions must repeat service-image
  values before workflow steps can source that file.

## Pending operational transitions

- Required contexts remain `Validation Gate` and `PR Metadata Gate` until the
  complete Issue #116 comparison passes.
- The deployer's temporary Artifact Registry writer grant remains until a real
  CI-built digest deploy succeeds; only then may
  `grant_deployer_artifact_writer` be disabled.
- External `/gcbrun` evidence remains tracked separately by Issue #140.
