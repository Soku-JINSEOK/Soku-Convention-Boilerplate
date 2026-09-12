# Issue #117 Current-main Hosted Full Forward-port

## Goal

Provide an additive Hosted Full workflow from the current main contract
without removing the existing per-PR Validation Gate, changing required
contexts, or mutating rulesets.

## Current-main contract

The current security.yml workflow accepts two distinct inputs:

- base-sha: the trusted policy and comparison source;
- head-sha: the exact repository commit to scan.

This forward-port preserves that distinction. Repository and template reusable
workflows receive head-sha and pass it to every checkout. Hosted Full passes
both values to Security and aggregates repository, template, and Security
results into Hosted Full Gate.

Manual and scheduled Hosted Full runs default both values to github.sha.
Reusable callers must provide exact values when validating a separate source.

## Scope

- Add .github/workflows/full-validation.yml.
- Add optional head-sha to ci.yml and templates-ci.yml.
- Pin every checkout in those reusable workflows to the selected head.
- Add focused regression checks for exact checkout propagation and fail-closed
  aggregation.
- Register the new workflow and task report in the boilerplate hygiene
  inventory, and document the Hosted Full classification and CI contract.
- Preserve the existing validation.yml, release, deploy, IAM, ruleset, and
  Cloud Build resources.

## Verification

- `node --test .github/validation-workflow.test.mjs scripts/verify-release-identity.test.mjs scripts/pull-request-policy.test.mjs scripts/verify-supply-chain.test.mjs` — 48 passed.
- PyYAML parsed `ci.yml`, `templates-ci.yml`, `full-validation.yml`, and
  `security.yml` successfully.
- `node scripts/verify-supply-chain.mjs` — 39 protected files and 11 update
  targets verified.
- Every checkout in `ci.yml` and `templates-ci.yml` has the exact-head ref.
- `git diff --check` — passed.
- `actionlint` was not available in the local environment; hosted workflow
  checks remain an explicit acceptance gate.

## Remaining gates

This source change does not run Hosted Full, alter GitHub rulesets, create or
change Cloud Build triggers, publish a release, or deploy. Issue #116's
comparison evidence, an owner-approved hosted run, and the required ruleset
transition remain separate gates. The predecessor PR #158 remains preserved
because its source-sha implementation targets an older workflow contract.
