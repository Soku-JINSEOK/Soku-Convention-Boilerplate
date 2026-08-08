# soku v0.3.0 release candidate

Release axis: soku

Publication status: source candidate only; tag, GitHub Release, archives,
checksums, npm publication, and downstream adoption are not authorized by
Issue #201

Manifest schemas: manifest-v1 and manifest-v2 remain byte-compatible and
readable; manifest-v3 adds explicit project-owned core-rendering overrides

Boilerplate compatibility: v1.0.1 through v1.0.5 ownership-handoff reproducer,
with the existing declared lifecycle compatibility retained for other paths

Profiles: bootstrap, standard (default and legacy), and scaled

Provider compatibility: provider-v1 with optional deprecated legacy ref;
delivery remains disabled unless separately approved

Recovery and exit-code contract: unchanged, with manifest-only handoff apply
using the existing manifest-last transaction and exact previous-manifest
rollback

Package matrix: Linux amd64/arm64, macOS amd64/arm64, and Windows amd64 (five
archives required before any publication approval)

Lifecycle conformance evidence: manifest-v1/v2 regression, manifest-v3 schema
and fixtures, single-path handoff refusal matrix, dry-run byte immutability,
selected-file byte/mode preservation, rollback/recovery injection,
same-release and future-release suppression, component no-downgrade, and the
existing three-OS lifecycle suites

Package and distribution: candidate metadata only; the coordinated native and
`@soku-jinseok/soku` 0.3.0 publication remains a later approval boundary

Any breaking behavior, read-only compatibility state, or manual recovery
requirement: older CLIs treat manifest v3 as unsupported/read-only; repositories
remaining on v1/v2 retain their existing bytes and behavior

Companion tag: none

This candidate adds
`soku ownership handoff --path <path> --expected-sha256 <sha256>` for one
reviewed, intentionally modified current core-managed file. Dry-run is
non-mutating. Apply requires `--yes` or interactive confirmation, never writes
the selected file, and changes only the manifest after revalidating exact bytes,
mode, and normalized hash.

The successful transition records the canonical path in
`selection.project_owned_overrides`, changes its file record to
`project-owned` / `unmanaged-expected`, and explicitly migrates the manifest to
v3. Same-release diff and future core upgrades suppress rendering for that
path. Provider and component ownership collisions remain fail-closed; reclaim
requires a separately reviewed explicit operation.

No tag, Release, package, checksum artifact, Provider apply, downstream merge,
cloud mutation, ruleset change, credential change, or delivery activation is
created by this candidate record.
