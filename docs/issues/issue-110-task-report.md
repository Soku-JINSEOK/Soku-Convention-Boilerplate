# Issue #110 Task Report — Silence CD plan and summary logs in node tests

## Goal and Background

Issue [#110](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/110) requires suppressing verbose `CD plan` and `Cloud Run deployment evidence` test output logs from dumping into the terminal console during `node --test .github/*.test.mjs`.

## Proposed Approach

Set a temporary `GITHUB_STEP_SUMMARY` file path in `.github/deploy-gcp.test.mjs` test helper `run()`. This redirects the step summary output from standard stdout to a temporary file, silencing console spam while keeping test assertions intact.

## Planned Implementation

- Update `run()` in `.github/deploy-gcp.test.mjs` to specify `GITHUB_STEP_SUMMARY: summaryFile` in the environment options.

## Acceptance Criteria

- `node --test .github/deploy-gcp.test.mjs` runs silently without console dump.
- All 41 node tests pass cleanly.

## Approval

- User requested creating Issue #110 and implementing the silence test output fix.
