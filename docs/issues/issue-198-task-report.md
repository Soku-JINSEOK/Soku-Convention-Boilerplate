# Issue #198 Task Report — Immutable CI/CD adapter conformance

## Goal and Background

Issue [#198](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/198)
requires a repository-owned conformance boundary for the portable CI/CD planner.
The public engine contracts are useful provenance inputs, but they must not be
downloaded, executed, or treated as delivery authority by Soku-core.

## Proposed Approach

Vendor the reviewed engine schemas and synthetic fixtures as immutable bytes,
record their hashes and the public engine merge in a strict mapping catalog,
and expose only three validation-only mappings. Trusted renderers are embedded
in the CLI, use fixed verification argv, and are checked for immutable action or
image references, narrow permissions, absent secrets, and no delivery behavior.

## Planned Implementation

- Add `ci-cd-adapter-mapping-v1` schema and catalog with exact engine and
  descriptor provenance.
- Vendor the versioned engine schemas, Local/GCP/Jenkins descriptors, and
  portable fixtures for hash-only conformance.
- Add trusted renderers for GitHub-hosted, GCP-managed, and GitHub self-hosted
  validation callers.
- Resolve the catalog from `soku ci-cd plan` without filesystem writes,
  inventory inspection, remote access, or adapter execution.
- Add positive, deterministic, provenance, profile, capability, and semantic
  adversarial tests with stable error IDs.

## Acceptance Criteria

- Only the three reviewed `ci-only` platform mappings are installable.
- Every mapping binds immutable engine/ref/descriptor/implementation/template
  hashes, fixed `ci-quick` and `full` argv, runner capabilities, and
  `delivery_authority: none`.
- Vendored upstream files match their recorded byte hashes and are never run.
- Mutable references, arbitrary commands, download-and-execute, broad
  permissions, undeclared secrets, delivery behavior, stale provenance,
  profile drift, disabled adapters, and fallback behavior fail closed with
  stable error IDs.
- Repeated human and JSON plans remain byte-stable and write zero files.

## Approval

- **Status:** `Approved`
- **Approved by:** `Soku-JINSEOK` through the repository closeout execution plan

## Implementation Status

The immutable catalog, vendored provenance, trusted renderers, planner
resolution, and semantic conformance tests are implemented on the dedicated
Issue #198 branch. Hosted final-head validation remains required before merge.

## Verification

Passed locally so far:

- `go test ./internal/cicd`
- `go test ./...`
- JSON syntax and adapter-catalog schema validation
- `git diff --check`

Environment-dependent lifecycle, race, packaging, and full hosted checks remain
to be run on the exact pull-request head.

## Public Disclosure Review

- [x] No credentials, tokens, private keys, or credential-bearing URLs
- [x] No private repository, project, or product names
- [x] No cloud project IDs, account numbers, service URLs, image URIs, or
      private revision identifiers
- [x] No personal billing, subscription, budget, or payment-status information
- [x] No personal email, phone, address, or local absolute path
- [x] No private Issue, PR, Project, or control-plane identifiers

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex
