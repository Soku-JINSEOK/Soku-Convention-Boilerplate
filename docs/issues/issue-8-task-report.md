# 📝 Task Report: add task-report and title-check templates

## Goal and Background

Closes #8. `PULL_REQUEST_TEMPLATE.md` requires a `docs/issues/issue-<n>-task-report.md` file that had no template, `contribution-title.mjs` had no CI enforcement or regression tests, and there was no bilingual template for GitHub comments — see issue #8 for the full comparison.

## Proposed Approach

Add the four missing artifacts (task report template, contribution-title CI check, regression tests, comment templates) and wire each into the existing standards docs (`BLUEPRINT.md`, `GITHUB_STANDARDS.md`, `APPLICABILITY.md`, `INIT_GUIDE.md`, README document indexes, `ci.yml`'s required-file check) following the repository's existing Authority Model and Language Policy rather than inventing new patterns.

## Planned Implementation

- `docs/issues/TASK_REPORT_TEMPLATE.md`, `.github/COMMENT_TEMPLATES.md`
- `templates/_shared/ci/contribution-title-check.yml`, `templates/_shared/commitlint/contribution-title.test.mjs`
- Doc updates: `BLUEPRINT.md`, `docs/standards/GITHUB_STANDARDS.md`, `docs/standards/PROJECT_STRUCTURE.md`, `docs/guides/APPLICABILITY.md`, `docs/guides/INIT_GUIDE.md`, `README.md`/`.ko`/`.ja`, `.github/workflows/ci.yml`

## Acceptance Criteria

- Every new file exists and is referenced from the relevant standards doc, so no template is a dangling reference.
- `node --test` actually runs in CI (this repo's own hygiene job and the downstream starter workflow), not just a file-existence check.
- `markdownlint-cli2` and `yaml-lint` pass repo-wide; the commit-title regression suite passes.

## Approval

- **Status:** `Approved`
- **Approved by:** `Soku-JINSEOK` (plan approved via Claude Code plan mode prior to implementation)

## Implementation Status

Complete. All planned files were added and all planned doc updates were applied.

An independent review pass (workflow-backed code review) additionally found and fixed, before this was ever committed:

- `contribution-title-check.yml` was missing `types: [opened, synchronize, reopened, edited]` on its `pull_request` trigger, so a title-only edit (no new commit) would not re-run the check despite the workflow's own header comment claiming it catches GitHub UI title edits.
- Its `git log` range had no `--no-merges`, so a PR containing a merge commit (e.g. merging `main` into a feature branch) would always fail the check on that commit's auto-generated subject.
- `docs/issues/` was added to `BLUEPRINT.md` but not to `README.md`'s (and `.ko`/`.ja`'s) document index or `docs/standards/PROJECT_STRUCTURE.md`'s summary of the `docs/` grouping.
- `contribution-title.test.mjs` was listed in `required_files` (existence only) but nothing in CI ever executed it — added a `node --test` step to both `.github/workflows/ci.yml` and `templates/_shared/ci/contribution-title-check.yml`.

## Verification

- `npx --yes markdownlint-cli2@0.22.1 --config .markdownlint.jsonc "**/*.md" "#**/node_modules/**"` — pass, 35 files, 0 errors.
- `npx --yes yaml-lint@1.7.0 .github/workflows/*.yml templates/_shared/ci/*.yml` — pass.
- `node --test templates/_shared/commitlint/*.test.mjs` — pass, 10/10.
- Manually confirmed every path in `ci.yml`'s `required_files` array exists on disk.
- `grep -rniI "cutvi"` repo-wide (excluding `.git`, `node_modules`) — no matches.

## AI Assistance

- **Planning/implementation/drafting:** `Claude Code`
