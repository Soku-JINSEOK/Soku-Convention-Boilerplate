# Issue #35 Task Report

## Goal and Background

Repair the release gate and publish the first boilerplate and CLI releases from
one reviewed commit without paid runners, external paid services, or persistent
test infrastructure. This report tracks
[Issue #35](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/35).

## Proposed Approach

Reuse the full repository and runtime-template CI workflows from one integrated
Release workflow. Keep manual dispatch validation-only. On tag events, require
paired signed annotated tags, GitHub-verified signatures, complete compatibility
records, and an identical source commit before delivery. Build CLI assets only
for the CLI release and perform all post-publication smoke tests in temporary
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

Remediation in progress. The preparation PR, hosted preflight, atomic tag push,
and both initial GitHub Releases completed. Post-release smoke testing found
that `soku/v0.1.0` rejects GitHub's PAX global archive header as a second root.
The public tags remain immutable; the CLI-only fix is being issued as
`soku/v0.1.1` against the existing boilerplate `v1.0.0`.

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
  hosted Templates CI because they depend on CI services or Docker.
- Hosted Actions, release asset, installation, and downstream lifecycle evidence
  — initial Actions, assets, checksums, archive execution, and `go install`
  passed; public-source lifecycle failed on the PAX global header and is pending
  the CLI patch release.

## AI Assistance

- Planning and implementation: OpenAI Codex
