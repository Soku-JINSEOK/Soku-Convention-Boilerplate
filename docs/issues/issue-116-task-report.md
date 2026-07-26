# Issue #116 Task Report — Run CI Quick beside full validation

## Goal and Background

Issue [#116](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/116)
starts the measured parallel phase of the verification rollout. The new quick
gate must exercise the fail-closed scope detector from #115 on every pull
request and `main` push while the existing required `Validation Gate` remains
unchanged.

## Proposed Approach

Add a hosted-only `ci-quick` profile that requires an explicit commit range and
does not permit DB or infrastructure skips. A reusable workflow checks out the
exact head with full history and invokes that profile. `validation.yml` calls
it for code-bearing events in parallel with the existing full workflows and
uses a metadata-only no-op to preserve one stable aggregate job named
`CI Quick Gate` without duplicating expensive checks on label or assignment
changes.

## Acceptance Criteria

- `CI Quick Gate` runs for every event currently handled by `validation.yml`.
- Whitespace, scope detection, changed syntax, lockfile install, changed-scope
  test/build smoke, and changed-line secret scanning are fail-closed.
- The existing `Validation Gate`, `PR Metadata Gate`, and branch rules remain
  unchanged.
- Workflow syntax, immutable action references, fork-safe permissions, and
  aggregate failure propagation are regression-tested.
- Observation continues for at least 14 days and 10 code-changing pull
  requests before #117 may start.
- The final comparison records zero missed defects, zero unresolved flakes,
  all scope fixtures passing, median duration at most 50% of full, and Actions
  usage at least 40% lower than full.

## Approval

- **Status:** `Approved`
- **Approved by:** `Soku-JINSEOK` through the user-authorized implementation
  plan supplied for this continuation

## Implementation Status

Implementation in progress. The comparison window starts only after the
implementation pull request is merged. Its result will be published separately
under `docs/audits/`.

## Verification

- Workflow/profile regression tests: 28/28 passing.
- Bash syntax, ShellCheck, YAML/Markdown lint, and actionlint: passing.
- Supply-chain verification: passing for all protected workflow sources and
  immutable action/container references.
- Fail-closed changed-file run: all ten scopes selected for the workflow and
  profile changes; Soku and every runtime/config template check passed.
- MySQL and PostgreSQL services became healthy and both schemas loaded.
- Terraform ran credential-free in the digest-pinned container with an
  isolated data directory, read-only source mount, read-only lockfile, and
  successful `fmt`, `init`, and `validate`.

## Compatibility Boundary

The immutable `v1.0.5` catalog and released CLI behavior are not changed in the
initial instrumentation pull request. The downstream quick/full/security
catalog migration requires its own regression-backed compatibility change and
will be included in the later #116 adjustment pull request before the Issue can
close.

## Public Disclosure Review

- [x] No credentials, tokens, private keys, or credential-bearing URLs
- [x] No cloud account identifiers or private service endpoints
- [x] No private repository or project identifiers
- [x] No personal billing or payment-status information

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex (GPT-5.6)
