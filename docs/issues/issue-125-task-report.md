# Issue #125 Task Report — Isolate fork PRs from token-backed tests

## Goal and Background

Issue [#125](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/125)
requires public fork pull requests to retain the complete hermetic lifecycle
test surface without executing contributor-controlled Go tests with an
explicit `GITHUB_TOKEN`.

## Proposed Approach

Keep the hermetic lifecycle gate unconditional and restrict only the external
network-conformance step. The authenticated step runs for same-repository pull
requests and non-pull-request trusted events, while fork pull requests skip it
without receiving the token.

## Planned Implementation

- Add a trusted-event condition to the token-backed network-conformance step
  in the reusable repository CI workflow.
- Add regression coverage for fork PR, same-repository PR, `main`, schedule,
  and manual-dispatch contexts.
- Assert that no relevant workflow subscribes to `pull_request_target`.
- Document the hermetic and network-provider verification boundaries.

## Acceptance Criteria

- Fork pull requests run the hermetic lifecycle gate without an explicit
  repository token.
- Token-backed provider network conformance runs only in trusted contexts.
- Release and trusted-branch conformance coverage remains available.
- Event-selection tests cover every documented context.
- No pull-request workflow executes an untrusted head through
  `pull_request_target`.

## Approval

- **Status:** `Approved`
- **Approved by:** `Soku-JINSEOK` (requested implementation of the approved
  Issue roadmap)

## Implementation Status

Implemented. Relevant local checks passed; hosted event-selection evidence
remains pending.

## Verification

- `node --test .github/validation-workflow.test.mjs` (8/8 passing).
- `npx --yes yaml-lint@1.7.0 .github/workflows/ci.yml` (passing).
- `npx --yes markdownlint-cli2@0.22.1 --config .markdownlint.jsonc
  VERIFICATION_GUIDE.md docs/issues/issue-125-task-report.md` (0 errors).
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
