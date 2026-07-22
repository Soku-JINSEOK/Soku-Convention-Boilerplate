# Issue 72 Task Report

## Goal and Background

Resolve central governance template drift in `Soku-Convention-Boilerplate` under
[`Soku-Convention-Boilerplate#72`](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/72).
This distribution is part of the approved repository-operations rollout tracked
by `ci-cd-control-plane#9` and `ci-cd-control-plane#21`.

## Proposed Approach

Apply the control plane's registered `boilerplate` profile output without
repository-name conditionals or delivery changes. Copy the shared PR policy and
contribution-title validators that the profile audit requires.

## Planned Implementation

- Write the rendered governance version, pull request template, and issue forms.
- Add the shared contribution-title and pull-request-policy workflows and scripts.
- Validate repository tests, lint, and whitespace before publishing a draft PR.

## Acceptance Criteria

- The control-plane repository audit reports no actionable template drift for
  `soku-convention-boilerplate`.
- Existing application, release, delivery, secrets, and custom labels remain
  unchanged.
- The change is delivered through a separately reviewable pull request.

## Approval

- **Status:** `Approved`
- **Approved by:** `Soku-JINSEOK` (conversation approval on 2026-07-22)

## Implementation Status

Implemented and published in PR #71 on branch
`agent/governance-template-distribution`; hosted validation passed.

## Verification

- `npx --yes markdownlint-cli2@0.22.1 --config .markdownlint.jsonc "**/*.md" "#**/node_modules/**"`: passed, 62 files and 0 errors.
- `npx --yes yaml-lint@1.7.0 .github/*.yml .github/**/*.yml templates/**/*.yml templates/**/*.yaml`: passed.
- `node --check scripts/contribution-title.mjs && node --check scripts/pull-request-policy.mjs`: passed.
- `node --test scripts/contribution-title.test.mjs scripts/pull-request-policy.test.mjs`: passed, 24 tests.
- Python `yaml.safe_load` over `.github/**/*.yml`: passed.
- `git diff --check`: passed before this report update; rerun before commit.
- `scripts/verify-sync-parity.sh`: unavailable because `pwsh` is not installed;
  hosted validation remains required.
- `actionlint`: unavailable locally; hosted validation remains required.
- GitHub Actions `Validation` run `29882460796`: passed, including the
  aggregate `Validation Gate` and hosted actionlint/sync parity coverage.
- `npm test` and `npm run lint`: not applicable because this repository has no
  root `package.json`; both commands returned `ENOENT` and are not claimed as
  passes.

## AI Assistance

- **Planning/implementation/drafting:** `OpenAI Codex`
