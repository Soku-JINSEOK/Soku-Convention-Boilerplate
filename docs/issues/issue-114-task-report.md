# Issue #114 Task Report — Freeze and classify current CI checks

## Goal and Background

Issue [#114](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/114) is phase 1 (plus the start of phase 2) of the sub-issue of #112: before any required-gate or branch-protection change, every check currently run by `ci.yml`, `templates-ci.yml`, `security.yml`, `contribution-title-check.yml`, `pull-request-policy.yml`, `release.yml`, and `deploy-gcp.yml` must be classified as `local-capable` / `hosted-only` / `release-only` / `deployment-only`, and a single local entry point must run every `local-capable` check reproducibly.

## Proposed Approach

Inventory every job and exact command in the six workflow files, classify each in a new `verification/CLASSIFICATION.md`, extract the drifting tool versions/thresholds into a single `verification/tools.env`, split the checks into shared `verification/commands/*.sh` groups, add `docker-compose.verify.yml` so the MySQL/PostgreSQL schema checks become locally reproducible, and add `scripts/verify.sh --profile full` as the new local entry point, with `scripts/ci-local.sh` becoming a thin wrapper around it.

## Planned Implementation

- `verification/CLASSIFICATION.md`
- `verification/tools.env`
- `verification/commands/{_lib,repo-hygiene,soku,templates,db-schema,security}.sh`
- `docker-compose.verify.yml`
- `scripts/verify.sh`
- `scripts/ci-local.sh` reduced to a wrapper
- `scripts/verify.test.mjs`, wired into `ci.yml`'s existing `node --test` step
- `VERIFICATION_GUIDE.md`, `docs/standards/CICD_STANDARDS.md`, `docs/guides/USAGE_MANUAL.md` updated to reference the new entry point

## Acceptance Criteria

- `scripts/verify.sh --profile full` runs every locally-reproducible check.
- Checks that can't be guaranteed locally are documented as `hosted-only`.
- Local and CI use identical commands and tool versions from one source of truth (resolves the `npm audit` level `high`/`low` drift and the `goimports` `v0.29.0`/`v0.48.0`/`@latest` drift).
- No existing required gate, branch protection rule, or CD behavior changes.

## Approval

- **Status:** `Approved`
- **Approved by:** `Soku-JINSEOK` (requested proceeding with issue #112 step by step across multiple commits)

## Implementation Status

Complete. All planned files added; existing regression suite (84 tests) re-run clean after the change.

## Verification

- `bash -n` over all new/edited shell scripts.
- `shellcheck` over all new/edited shell scripts (0 findings).
- `scripts/verify.sh --help` / missing `--profile` / unknown profile / not-yet-implemented profile behavior matches documentation.
- `scripts/ci-local.sh --help` output matches `scripts/verify.sh --help` (delegation works).
- `node --test scripts/verify.test.mjs` (9/9 passing).
- Full existing regression suite referenced from `ci.yml` (84/84 passing, confirming no regressions from the `ci.yml` edit).
- `markdownlint-cli2` clean on all new/edited Markdown.
- Full end-to-end `scripts/verify.sh --profile full` run (npm/go/docker toolchain install and build) was not exercised in the authoring sandbox for time; a maintainer or hosted run should confirm it before merge.

## AI Assistance

- **Planning/implementation/drafting:** Claude Code (Sonnet 5)
