# Verification Check Classification

Phase 1 of [issue #112](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/112):
freeze every check currently run by CI and classify it before changing any
required gate. No workflow, branch protection, or CD behavior changes yet —
this is inventory only.

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
| MySQL schema load | `mysql` | **hosted-only today** — becomes local-capable once `docker-compose.verify.yml` (this phase) is used |
| PostgreSQL schema load | `postgresql` | **hosted-only today** — same as above |
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

All jobs (`validation` re-run, tag/signature verification, GPG-signed notes
check, packaging + `gh release create`, `npm publish --provenance`) are
**release-only** — they only run on a tag push or release dry-run dispatch.

## `deploy-gcp.yml`

| Check | Operation | Category |
| --- | --- | --- |
| `bash -n` + `deploy-gcp.test.mjs` | `check` | local-capable |
| WIF auth, image build+push+digest resolve, Cloud Run deploy, health check | `deploy` | **deployment-only** |
| Rollback to previous revision | `rollback` | **deployment-only** |

## Known drift resolved by `verification/tools.env` (this phase)

- `npm audit --audit-level`: `ci-local.sh` used `high`, `security.yml` uses
  `low`. Unified to `low` (the stricter of the two) in `tools.env`.
- `goimports` version: `ci-local.sh`'s Go template check pinned `v0.29.0`,
  `ci.yml`'s `soku-quality` job pinned `v0.48.0`, and `templates-ci.yml`'s `go`
  job used unpinned `@latest`. Unified to `v0.48.0` in `tools.env`.
  `templates-ci.yml` itself still installs `@latest` — updating that
  hosted-only-required workflow is out of scope for this phase (no CI
  reduction/behavior change yet) and is left as explicit follow-up once
  `verify.sh`/`tools.env` are wired into CI directly.

## Explicitly out of scope for this phase

Per the issue's phased rollout, none of the following change yet: the `fast`
profile or path-based scope detection, a `CI Quick Gate` workflow, branch
protection required contexts, `hosted-full` scheduling, release gating on
`hosted-full`, or CD restructuring. This document and the `verify.sh --profile
full` entry point it backs only reproduce today's checks locally — they do not
remove or replace any existing required check.
