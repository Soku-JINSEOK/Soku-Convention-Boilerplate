# Issue #118 Task Report — Gate releases on exact-tag Hosted Full

## Goal and Background

Issue [#118](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/118)
requires release validation and every component checkout to use the exact tag
event source rather than a moving default branch.

## Implementation Status

Repository, template, and security reusable workflows accept `source-sha` and
pass it to every checkout. Release validation calls `full-validation.yml` with
`github.sha`; metadata, packaging, and npm publication checkouts use the same
SHA. All assets and the npm package are preflighted before a draft GitHub
Release is created. CLI releases become public only after tokenless npm
publication succeeds; manual dispatch remains validation-only because delivery
still requires a canonical tag push.

## Remaining Operational Gates

- Merge after Issue #117's code path is reviewed.
- Run a validation-only manual preflight on `main`.
- Retain exact-tag Hosted Full evidence from the later companion release.
- Configure the exact npm Trusted Publisher and preserve the draft Release if
  publication fails. If npm succeeds but finalization fails, retry only the
  GitHub Release finalization step.

## Verification

- Static regression tests require exact-SHA propagation and publication needs.
- Hosted Full aggregate rejects failure, cancellation, and unexpected skips.
- actionlint and supply-chain verification passed.

## Public Disclosure Review

- [x] No credentials or private identifiers
- [x] No cloud account identifiers or service endpoints
- [x] No billing or personal information

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex (GPT-5)
