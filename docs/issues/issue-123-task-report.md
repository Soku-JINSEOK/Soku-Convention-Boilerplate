# Issue #123 Task Report — Unify version metadata and publication identity

## Goal and Background

Issue [#123](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/123)
requires every public release surface to derive from one reviewed identity
record while publication credentials and signer trust are strengthened.

The published state on 2026-07-26 is boilerplate `v1.0.5`, native CLI
`soku/v0.2.1`, and npm launcher `0.2.1`. The verified lifecycle compatibility
baseline remains `v1.0.5` with `soku/v0.1.4`; it is intentionally distinct from
the current distribution version.

## Approved Implementation Sequence

1. Add a machine-readable release identity and reject metadata drift.
2. Migrate npm publication from `NPM_TOKEN` to Trusted Publishing/OIDC after a
   verified publication path exists.
3. Enforce an approved signer fingerprint set and document reviewed rotation.

## Phase 1 Implementation

- Add `release-identity.json` as the reviewed current-version authority.
- Validate package metadata, workflow defaults, installation documentation,
  compatibility notes, and release records against the manifest.
- Register the verifier and its regression tests in repository hygiene.
- Align current-release examples without modifying any published tag, Release,
  npm version, or historical release note.

## Remaining Gates

- Trusted Publishing/OIDC must succeed with provenance and without
  `NPM_TOKEN`.
- The previous long-lived npm token may be retired only after that success.
- Tag verification must reject a valid signature outside the approved signer
  set.
- A reviewed signer-rotation procedure and tests must be added.
- The exact-tag `hosted-full` release gate must remain intact.

## Non-destructive Boundary

No existing public tag, GitHub Release, npm version, or historical release note
is moved, deleted, reused, or republished.

## Approval

- **Status:** `Approved`
- **Approved by:** `Soku-JINSEOK` through the ordered Issue roadmap

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex (GPT-5)
