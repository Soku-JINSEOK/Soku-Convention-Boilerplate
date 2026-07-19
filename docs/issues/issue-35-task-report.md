# Issue #35 Task Report

## Goal and Background

Repair the release gate and publish the first boilerplate and CLI releases from
one reviewed commit without paid runners, external paid services, or persistent
test infrastructure. This report tracks
[Issue #35](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/35).

## Proposed Approach

Reuse the full repository and runtime-template CI workflows from one integrated
Release workflow. Keep manual dispatch validation-only. On tag events, require
signed annotated tags, GitHub-verified signatures, and complete compatibility
records. Paired releases must resolve to an identical source commit; a
single-axis patch explicitly records no companion. Build CLI assets only for
the CLI release and perform all post-publication smoke tests in temporary
directories.

## Planned Implementation

- Prepare compatibility and migration records for boilerplate `v1.0.0` and CLI
  `soku/v0.1.0`.
- Make CI and Templates CI reusable and remove their direct tag delivery paths.
- Add an integrated Release workflow that gates delivery on every repository,
  CLI, lifecycle, package, and runtime-template check.
- Extend the local tag helper with `--tag`, `--notes-file`, and `--dry-run` while
  retaining its no-argument interactive mode and never pushing automatically.
- Add regression coverage for malformed, lightweight, and unsigned tags.
- Merge the preparation PR, run manual preflight on `main`, create both signed
  tags at the same commit, and push them atomically.
- Verify both Releases, the exact six CLI assets, checksums, archive metadata,
  module installation, and temporary downstream lifecycle behavior.
- Record Actions, tag, commit, asset, checksum, and smoke evidence here before
  closing the Issue.

## Acceptance Criteria

- YAML and Markdown lint, contribution-title tests, sync parity, Go unit/race/
  vet/lint, lifecycle, package snapshot, and every runtime template pass.
- Manual preflight never executes delivery; invalid or unverified tags fail.
- `v1.0.0` and `soku/v0.1.0` are signed annotated tags on the same reviewed
  commit and each produces a GitHub Release without automatic latest selection.
- The CLI Release contains exactly five archives and `checksums.txt`.
- Download, checksum, archive execution, `go install`, metadata, and temporary
  `init -> status -> diff -> same-version upgrade` smoke tests pass.
- Failed gates create no Release. Published tags remain immutable and any fix is
  issued as the affected release axis's next patch.

## Approval

- **Status:** `Approved`
- **Approved by:** `Soku-JINSEOK` (the supplied implementation plan explicitly
  directs implementation and publication)

## Implementation Status

Complete. The preparation and remediation PRs, hosted preflights, initial atomic
tag push, and all three GitHub Releases completed. Post-release smoke testing
found that `soku/v0.1.0` rejected GitHub's PAX global archive header as a second
root. The public tags remained immutable, and the corrected CLI was published
as `soku/v0.1.1` against the existing boilerplate `v1.0.0`.

## Verification

- YAML lint — passed.
- Markdown lint — passed.
- Release notes contract checks — passed for both release axes.
- Malformed, lightweight, and unsigned tag regression tests — passed.
- Go module verification, unit tests, race tests, vet, gofmt, goimports, and
  golangci-lint — passed.
- Hermetic lifecycle conformance — passed locally.
- Five-target package snapshot, checksums, native execution, build metadata, and
  reproducibility — passed locally.
- JavaScript/TypeScript, Python, Go, and Java runtime templates — passed locally.
- Sync parity — deferred to hosted CI because PowerShell 7 is not installed in
  the local environment.
- MySQL, PostgreSQL, gcloud, and AWS/Azure configuration checks — deferred to
  hosted Templates CI because they depend on CI services or Docker; passed.
- Hosted Actions, release asset, installation, and downstream lifecycle evidence
  — passed for the corrected CLI patch.

## Release Evidence

- Preparation: [PR #36](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/pull/36),
  initial [preflight](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/29691038900),
  and initial tag workflows for
  [`v1.0.0`](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/29691730897)
  and
  [`soku/v0.1.0`](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/29691730744).
- Remediation: [PR #37](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/pull/37),
  post-merge [CI](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/29692998393),
  [Templates CI](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/29692998389),
  validation-only [preflight](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/29693067312),
  and gated
  [`soku/v0.1.1` delivery](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/29693174230).
- Releases:
  [`v1.0.0`](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/releases/tag/v1.0.0),
  [`soku/v0.1.0`](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/releases/tag/soku/v0.1.0),
  and corrected
  [`soku/v0.1.1`](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/releases/tag/soku/v0.1.1).
- GitHub reports valid signatures for all three annotated tags. The initial pair
  resolves to `c13a5e7614f242b6e1a1ccbe12072394479a112b`; the CLI patch resolves to
  `5b5949107a9e051570e964745fdccadedbe69fab`.
- The `soku/v0.1.1` Release has exactly five archives plus `checksums.txt`.
  SHA-256 values are `b9f367737de42ab9fdad597c686fcceb97bf52a21efc6725b87ac45a8e97915a`
  (darwin amd64), `aeb70f52a0885cd31c5a6ca85fdedf4f3a9c8f64117803e5de47d588f47e1a5a`
  (darwin arm64), `7acd35053bbfbfbca9d21708aeeb8c863a1666a2b1f6b6688de521b5795656cd`
  (linux amd64), `1d45181461b27f4a2443eb8133ee4e3ba8e69ab73a50ade6863987b34ea4f9ac`
  (linux arm64), and `9f09e656f5515cec2f501a7d75008136a092ef7585c181ae54abc2797f989a4e`
  (Windows amd64).
- A downloaded darwin arm64 archive reported version `v0.1.1`, commit
  `5b5949107a9e051570e964745fdccadedbe69fab`, and hosted build timestamp
  `2026-07-19T15:30:57Z`. `go install` resolved module `v0.1.1` through the Go
  proxy; source installations intentionally report unknown linker metadata.
- In a fresh temporary directory, the released binary initialized boilerplate
  `v1.0.0` at `c13a5e7614f242b6e1a1ccbe12072394479a112b`, reported eight clean managed
  files, and returned no-op results for both `diff` and same-version `upgrade`.

## AI Assistance

- Planning and implementation: OpenAI Codex
