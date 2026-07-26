# Issue #124 Task Report — Establish immutable supply-chain inputs

## Goal and Background

Issue [#124](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/124)
requires protected verification and release paths to resolve only reviewed
dependency, tool, Action, and image versions. The first implementation unit
addresses CVE-2026-14257, which caused every current pull request's dependency
audit to fail through transitive `brace-expansion` versions.

## Proposed Approach

Deliver the Issue through individually reviewable supply-chain changes:

1. Pin the patched npm transitive dependency and restore a zero-vulnerability
   JS/TS template audit without forcing an unrelated ESLint/GTS migration.
2. Remove floating tool and image references and make hosted verification use
   the reviewed local version source.
3. Complete dependency-update coverage, add a floating-reference verifier,
   and document the reviewed update process.

## Planned Implementation

- Pin `brace-expansion` to reviewed version `5.0.8` in the JS/TS template and
  regenerate its npm lockfile.
- Replace remaining floating executable and container references with reviewed
  versions or immutable digests.
- Extend Dependabot or documented update coverage to every tracked ecosystem.
- Add regression checks that reject unapproved floating references.
- Document the inventory and update procedure.

## Acceptance Criteria

- The JS/TS template installs reproducibly and reports zero known
  vulnerabilities at the configured audit threshold.
- Protected paths contain no unapproved floating dependency or Action refs.
- Selected service and build images use reviewed immutable digests.
- Hosted and local verification consume the same reviewed version source.
- Every tracked dependency ecosystem has an active update path.
- A regression verifier rejects reintroduced floating references.
- Existing required gates and delivery behavior remain unchanged.

## Approval

- **Status:** `Approved`
- **Approved by:** `Soku-JINSEOK` (approved proceeding with the first focused
  #124 security fix)

## Implementation Status

Phase 1 implemented. The npm audit fix is ready for review; the remaining
tool, image, update-coverage, and verifier phases are not complete.

## Verification

- `npm install --package-lock-only --ignore-scripts` completed and reported
  zero vulnerabilities.
- `npm ci` completed and audited 279 packages with zero vulnerabilities.
- `npm audit --audit-level=low` reported zero vulnerabilities.
- `npm run lint` passed.
- `npm run typecheck` passed.
- `npm test` passed (1 test).
- `npm run build` passed.
- `npm run format:check` passed.
- Targeted Markdown lint completed with zero errors.
- `git diff --check origin/main...HEAD` passed.

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
