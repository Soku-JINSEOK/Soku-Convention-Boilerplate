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

The first Quick Gate implementation merged in PR #146, but its single serial
job did not provide the runner topology required by the approved completion
plan. The follow-up implementation uses a detector-driven dynamic matrix for
always-on, Soku, language-template, database, cloud-template, and GCP
infrastructure shards. `verification/profiles.yml` is the single source for
group-to-scope and group-to-toolchain mapping, while
`scripts/plan-ci-quick.mjs` converts detector output to the hosted matrix.

The sharded implementation merged in
[#149](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/pull/149)
as commit `9539413a08e545a2e0d91383709cacc7b8a385de` at
`2026-07-26T14:13:58Z`. The earlier comparison window beginning at
`2026-07-26T12:09:01Z` is retained only as serial-implementation history. The
13-sample sharded window missed the critical-duration median target at 68.85%.

Issue #175 identified that CI Quick's shared Go setup could not discover a root
dependency file. PR
[#176](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/pull/176)
added explicit cache dependency paths and merged as commit
`db50c14781067d74ce05e689d13f951239d92f34` at `2026-07-31T02:47:21Z`.
Because this changes Quick behavior, the previous 13 samples remain historical
and the authoritative 14-day and 10-PR window restarts from that merge
timestamp, as recorded in
[`docs/audits/ci-quick-comparison.md`](../audits/ci-quick-comparison.md).

PR #181 later removed automatic Quick and Full validation from pull request and
`main` events while restoring trusted policy and security checks. The previous
activation is therefore invalid for closeout and the authoritative sample is
reset to 0/10. The 2026-08-08 recovery restores Validation as the single
automatic caller for Quick, Full repository, runtime templates, and trusted
Security. Security accepts explicit base/head commit inputs, while its schedule
and manual entrypoints remain independent. Pull Request Policy retains only the
trusted `PR Metadata Gate`; its temporary compatibility `Validation Gate` is
removed.

The recovery merged in
[#199](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/pull/199)
as commit `44a4106cab9e51f1568665e27d022a4d293e89c2` at
`2026-08-08T02:16:45Z`. That merge starts the new measurement epoch. The
14-day elapsed criterion was satisfied on `2026-08-22T02:16:45Z` and is no
longer a pending time gate as of `2026-08-30`. The recovery pull
request and its separate activation-record pull request are excluded from the
sample.

The current epoch now contains three qualifying natural code pull requests:
[#202](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/pull/202),
[#203](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/pull/203),
and [#204](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/pull/204).
Their exact-head Validation evidence is recorded in
[`docs/audits/ci-quick-comparison.md`](../audits/ci-quick-comparison.md).
The authoritative count is **3/10**; there are no observed Quick-pass /
Full-fail misses or unresolved flakes. The current median Quick/Full critical
duration ratio is **67.6%**, which does not meet the required maximum of 50%,
while aggregate runner-seconds are **987 seconds for Quick** and **2,149
seconds for Full**. The resulting aggregate reduction is **54.1%**, which meets
the required minimum of 40% for this small sample. The per-sample median
runner-second reduction is **55.6%** and is descriptive only, not the acceptance
metric.

This documentation reconciliation does not change Quick or Full behavior and
is excluded from the natural-sample counter. The remaining natural samples and
fixture checks remain mandatory before #117; the 14-day elapsed criterion is
already satisfied.

## Verification

### 2026-08-08 Recovery Verification

- Validation workflow contract and supply-chain regression tests pass.
- YAML lint and actionlint pass for Validation, Security, and Pull Request
  Policy.
- The Node 24 full profile passed the repository regression suites (75 tests),
  GitHub/workflow suites (53 tests), release-tag and historical-baseline tests,
  npm wrapper package tests, Markdown/YAML/actionlint checks, and Soku build,
  vet, and unit tests.
- The local full profile then stopped at `go test -race ./...` because this
  environment has `CGO_ENABLED=0` and no C compiler. Docker is also unavailable,
  so service/container stages remain for hosted validation. Neither missing
  local capability is represented as a passing full run.
- `git diff --check` and the immutable supply-chain verifier pass.
- The first hosted recovery run exposed a pre-existing shallow-checkout defect
  in `Repository Hygiene`: the job could not read the pinned historical
  baseline commit. Its checkout now fetches full history, with a workflow
  regression assertion covering that requirement.
- Restoring trusted Security also surfaced newly disclosed high-severity
  advisories in the JavaScript template lockfile. The lockfile now resolves
  `js-yaml` 4.3.1 and `nanoid` 3.3.18; `npm ci`, the template test/build/lint
  suite, and `npm audit --audit-level=high` pass locally with zero findings.
- The final signed PR head `540d4d19ac1d0b084efe74c22b98bb5729df3516`
  passed automatic Quick, Full repository, runtime-template, trusted Security,
  CodeQL, PR Metadata, and external Cloud Build validation. The Cloud Build PR
  trigger was reconciled from requiring `/gcbrun` for every contributor to the
  repository contract that requires it only for external contributors; the
  final head then started automatically without another approval comment.
- The signed merge commit `44a4106cab9e51f1568665e27d022a4d293e89c2`
  passed the automatic `main` Validation run
  [31234601365](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/31234601365),
  including `CI Quick Gate`, `Validation Gate`, and trusted Security. Its
  CodeQL and external Cloud Build validations also passed.

- Workflow/profile regression tests cover planner fail-closed behavior,
  matrix/toolchain propagation, known-group enforcement, and profile isolation.
- Current five-file and released three-file catalog shapes pass decoder
  regression coverage; incomplete four-file shapes fail both schema and runtime
  validation.
- All three downstream workflows render for multi-stack selection, with quick
  and full responsibilities separated and reviewed security tools retained.
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

The immutable `v1.0.5` catalog and released CLI behavior are not changed.
Current catalog sources add quick, full, and security workflow outputs while
the decoder continues to accept the released three-shared-file, single-workflow
shape. Catalog schema version 1, profile-index version 2, and manifest version 1
remain unchanged.

## Public Disclosure Review

- [x] No credentials, tokens, private keys, or credential-bearing URLs
- [x] No cloud account identifiers or private service endpoints
- [x] No private repository or project identifiers
- [x] No personal billing or payment-status information

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex (GPT-5.6)
