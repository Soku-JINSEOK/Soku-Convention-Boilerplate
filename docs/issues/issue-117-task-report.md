# Issue #117 — code-validation preservation and current-main Hosted Full forward-port

## Purpose

This integration combines the narrow code-validation correction with the
additive current-main Hosted Full caller. It preserves the existing required
contexts and does not mutate branch protection, rulesets, delivery, or deploy
behavior.

## Code-validation correction

Code run 34682376195 on PR #229 failed while later metadata-only runs skipped
all code groups and reported success with the same required check names.
Concurrency isolation did not prevent that overwrite. Code events retain
`CI Quick Gate` and `Validation Gate`; metadata events use
`CI Quick Metadata Only` and `Validation Metadata Only`. Names do not depend
on job success or cancellation.

The integrated regression tests evaluate the actual workflow expressions for
code and metadata events, including a base edit, and execute the aggregate
shell. Failed, cancelled, or unexpectedly skipped code groups fail closed.
No extra trigger, permission, or heavy validation execution is added.

## Current-main Hosted Full contract

The current security workflow accepts two distinct inputs:

- `base-sha`: the trusted policy and comparison source;
- `head-sha`: the exact repository commit to scan.

Repository and template reusable workflows receive `head-sha` and pass it to
every checkout. Hosted Full passes both values to Security and aggregates
repository, template, and Security results into Hosted Full Gate. Manual and
scheduled Hosted Full runs default both values to `github.sha`; reusable
callers provide exact values when validating a separate source.

## Integrated scope

- Preserve the #230 metadata/code gate separation and fail-closed aggregate
  tests.
- Forward the #233 exact-head input through `ci.yml` and
  `templates-ci.yml`.
- Add `.github/workflows/full-validation.yml` with repository, template, and
  security aggregation.
- Keep release, deploy, IAM, ruleset, and Cloud Build resources unchanged.
- Retain both historical source reports in this single review record.

## Verification evidence

The source candidates separately recorded:

- #230: regression tests for metadata-only naming, cancellation isolation,
  exact head/base propagation, and aggregate shell failures.
- #233: 48 hosted-contract tests passed; PyYAML parsed the reusable workflows;
  39 protected files and 11 update targets were checked; exact-head checkout
  propagation and `git diff --check) passed.

The new integration commit requires fresh hosted validation against its exact
head. Earlier results are not reused as evidence for that commit.

## Remaining gates

This source change does not run Hosted Full, alter GitHub rulesets, create or
change Cloud Build triggers, publish a release, or deploy. Issue #116
comparison evidence, owner-approved hosted acceptance, and any ruleset
transition remain separate gates. The predecessor PR #158 remains preserved
because its source-SHA implementation targets an older workflow contract.
