# Issue #120 Task Report — Reduce public metadata exposure

## Goal and Background

Issue [#120](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/120)
requests reducing operational and account-level metadata published in this
repository beyond what a reusable public boilerplate needs, in CD deployment
evidence, example integration configuration, a handful of historical task
reports, and cross-references to a private upstream repository scattered
across several older Issues and pull requests. No credential, token, or
private key was found during the review that prompted this issue.

## Proposed Approach

Reduce the CD evidence schema to a minimal, non-identifying field set;
replace real-world identifiers in example downstream-integration
configuration with synthetic values (recomputing the affected hashes so the
existing provenance/integrity tests still pass); generalize historical task
reports that recorded more operational detail than necessary; generalize
the same class of private-repository cross-reference wherever it appeared
in older, already-closed Issues and pull requests (both their live GitHub
bodies and any tracked task-report files); extend `.gitignore` with common
secret-adjacent patterns; and add a disclosure-review checklist to the task
report template. Also remove the currently-active deployment-evidence
Actions artifacts as an immediate, independent step (not gated on this PR).

## Planned Implementation

- `scripts/cd-deploy.sh`: minimize `record_evidence()` and
  `write_step_summary()` output; update `.github/deploy-gcp.test.mjs`
  accordingly, including a regression guard on the minimal field set.
- `soku/providers/*-control-plane-v1/*`, `soku/providers/provenance/registered-downstream-v1.json`:
  replace real-world values with synthetic ones, keeping the file structure
  and internal hash chain self-consistent; update the one Go test
  (`registered_provider_provenance_test.go`) that hardcodes the previous
  commit reference.
- Historical task reports: replace specific figures/identifiers with
  generalized statements.
- Older, already-closed Issues and pull requests that carried the same
  private-repository cross-reference pattern: update their live bodies via
  the GitHub API, and update the matching task-report files tracked in this
  repository, to the same generalized phrasing.
- `.gitignore`: add common secret-adjacent patterns not previously covered.
- `docs/issues/TASK_REPORT_TEMPLATE.md`: add a Public Disclosure Review
  checklist.
- Delete the currently-active `deploy-evidence-*` Actions artifacts via the
  Actions API.

## Acceptance Criteria

- CD deployment evidence artifacts contain only a minimal, non-identifying
  field set.
- Example provider configuration files use synthetic values only.
- The referenced historical task reports, Issues, and pull requests no
  longer record more detail than necessary.
- `.gitignore` covers common secret-adjacent patterns.
- The task report template includes a disclosure-review checklist.
- The existing security workflow and full Go/Node regression suite continue
  to pass.
- No currently-active deployment-evidence Actions artifact remains.

## Approval

- **Status:** `Approved`
- **Approved by:** `Soku-JINSEOK` (requested resolving issue #120)

## Implementation Status

Complete. Active `deploy-evidence-*` artifacts were deleted directly via the
Actions API. All planned file changes landed across commits on this PR's
branch, and the affected older Issues/pull requests were updated directly
via the GitHub API (their edit history remains visible on GitHub, same as
any other post-hoc edit — this only reduces what a casual reader of the
current body sees).

## Verification

- `shellcheck` and `bash -n` over all edited shell scripts (0 findings).
- `node --test .github/deploy-gcp.test.mjs` (15/15 passing, including new
  minimal-evidence-schema assertions).
- `node --test .github/validation-workflow.test.mjs .github/usage-manual.test.mjs`
  (no regressions from unrelated edits).
- `cd soku && go build ./... && go vet ./... && go test ./...` (all
  packages pass, including the recomputed provenance hash chain).
- `gofmt -l .` reports no files needing formatting.
- `markdownlint-cli2` clean on all edited Markdown.
- Confirmed via the GitHub Actions API that no `deploy-evidence-*` artifact
  remains active.
- Confirmed via the GitHub search and issue/PR APIs that no live Issue or
  pull request body in this repository retains the specific
  private-repository cross-reference pattern this issue addresses.

## AI Assistance

- **Planning/implementation/drafting:** Claude Code (Sonnet 5) — the
  underlying findings were independently verified against the live
  repository (active artifact list, provider file contents, task report
  contents) before any change was made.
