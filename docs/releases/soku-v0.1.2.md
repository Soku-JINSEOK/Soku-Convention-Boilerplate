# soku v0.1.2

Release axis: soku

Manifest schemas: manifest-v1 (unchanged)

Boilerplate compatibility: v1.0.0

Profiles: bootstrap, standard (default and legacy), and scaled

Provider compatibility: provider-v1 with optional deprecated legacy ref

Recovery and exit-code contract: unchanged from v0.1.1

Package matrix: linux amd64/arm64, macOS amd64/arm64, Windows amd64

Lifecycle conformance evidence: hermetic and pinned HTTPS provider gates required

Companion tag: none

This patch release makes the exact lowercase full commit supplied by
the CLI through `--integration-ref` and used for fetch the only authoritative
provider revision. Request artifacts and manifest-v1 integrations persist that
exact commit. Provider-supplied `ref` is optional and deprecated: bundles may
omit it, and a syntactically valid matching or mismatching legacy value does not
affect pending or connected decisions. A malformed present value remains a
provider compatibility failure.

Provider API v1 remains the supported major contract. Existing legacy bundles
with a well-formed `ref` require no immediate rewrite, while new bundles should
omit the field. Existing no-ref bundles become valid. Unknown fields, source or
configuration-hash mismatches, incompatible profiles, ownership conflicts,
mutable CLI refs, and executable provider content continue to fail closed.

Manifest-v1 is unchanged and requires no migration. Its provider API/schema,
lifecycle, ownership, and hash metadata retain their existing meanings. The
five-field pending request artifact remains limited to schema version, provider
ID, portable source, authoritative ref, and configuration hash. Sanitized or
raw configuration and secrets are never added to lifecycle state.

Boilerplate `v1.0.0` remains compatible. Verification requires Provider API
schema and decoder regressions, hermetic lifecycle and state-transition tests,
the public AI collaboration bundle fetched through HTTPS at immutable commit
`a81f7c91b0c9c8faa5ba2988fde29e9d17972a83`, the three-OS Go and lifecycle
matrix, quality/race/vet/format/import/lint checks, five-target package snapshot,
runtime templates, repository hygiene, sync parity, and the aggregate
`Validation Gate`.

Release validation uses `boilerplate-tag` set to `v1.0.0` and `cli-tag` set to
`soku/v0.1.2`. The manual dispatch is validation-only; delivery is triggered by
the separately approved, signed annotated `soku/v0.1.2` tag on the reviewed
`main` commit. `soku/v0.1.2` supersedes `soku/v0.1.1` as the current stable CLI
release without changing the public manifest-v1 or Provider API v1 contracts.
