# Issue #117 Task Report — Stage Hosted Full and required-gate migration

## Goal and Background

Issue [#117](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/117)
separates the daily/manual complete hosted suite from the pull-request critical
path, but required contexts may change only after Issue #116's comparison
criteria pass.

## Implementation Status

`full-validation.yml` now supports reusable exact-SHA calls, manual runs, and a
daily `02:41 UTC` schedule. Repository, template, and security results aggregate
into `Hosted Full Gate`. The current `Validation Gate` and `PR Metadata Gate`
required path remains intact; no ruleset mutation is authorized yet.

This report covers the additive #117-A merge unit only. The checkpoint in Draft
PR #154 is not a merge candidate. Removing the existing per-PR Full path is a
separate #117-B change after the observation and ruleset gates below pass.

## Remaining Operational Gates

- Complete and publish the Issue #116 comparison. The authoritative audit is
  `docs/audits/ci-quick-comparison.md`; it currently records 5/10 qualifying
  samples and an earliest completion of `2026-08-09T14:13:58Z`.
- Run Hosted Full manually and retain the Actions link.
- Change the ruleset epoch and required contexts.
- Only then remove PR/main full jobs from `validation.yml` and update governance
  audit epochs.

## Verification

- Workflow regression tests passed.
- actionlint passed for every workflow.
- Supply-chain verification passed for every protected workflow source.

## Public Disclosure Review

- [x] No credentials or private identifiers
- [x] No cloud account identifiers or service endpoints
- [x] No billing or personal information

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex (GPT-5)
