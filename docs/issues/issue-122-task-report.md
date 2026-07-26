# Issue #122 Task Report — Align npm package license and tarball contents

## Goal and Background

Issue [#122](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/122)
requires the public `@soku-jinseok/soku` package to declare one unambiguous
license and ship only the files that form its supported consumer contract.
The repository uses MIT, while the npm manifest previously declared
Apache-2.0 without including a package-scoped license.

## Proposed Approach

Use MIT for the npm wrapper, include a package-scoped copy of the repository
license, replace directory-wide package inclusions with an explicit runtime
allowlist, and compare both dry-run and real `npm pack` inventories against
that allowlist.

## Planned Implementation

- Change `soku/npm/package.json` to declare MIT and list only public runtime
  files.
- Add `soku/npm/LICENSE` and document the package license in its README.
- Add a Node regression test for license parity, `npm pack --dry-run`, and the
  generated tarball inventory.
- Register the new files in the repository hygiene check.

## Acceptance Criteria

- `package.json#license` and the packaged license both identify MIT.
- The package-scoped license is identical to the repository license.
- Dry-run and actual package inventories contain only the documented launcher
  contract.
- A future license or package inventory drift fails the npm package tests.
- Existing tags, GitHub Releases, and npm versions remain unchanged.

## Approval

- **Status:** `Approved`
- **Approved by:** `Soku-JINSEOK` (requested implementation of the approved
  Issue roadmap)

## Implementation Status

Implemented. Relevant local checks passed; hosted CI and final packed-artifact
evidence remain pending.

## Verification

- `cd soku/npm && npm test` (7/7 passing, including dry-run and actual
  tarball inventory checks).
- `npx --yes markdownlint-cli2@0.22.1 --config .markdownlint.jsonc
  soku/npm/README.md docs/issues/issue-122-task-report.md` (0 errors).
- `git diff --check origin/main...HEAD` (passing).

## Public Disclosure Review

- [x] No credentials, tokens, private keys, or credential-bearing URLs
- [x] No private repository, project, or product names
- [x] No cloud project IDs, account numbers, service URLs, image URIs, or
      revision identifiers
- [x] No personal billing, subscription, budget, or payment-status information
- [x] No personal email, phone, address, or local absolute path
- [x] No private Issue, PR, Project, or control-plane identifiers

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex (GPT-5)
