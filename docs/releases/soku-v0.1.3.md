# soku v0.1.3

Release axis: soku

Manifest schemas: manifest-v1 (unchanged)

Boilerplate compatibility: v1.0.0 and v1.0.1 remain readable; v1.0.2 is the
recommended companion baseline

Profiles: bootstrap, standard (default and legacy), and scaled

Provider compatibility: provider-v1 with optional deprecated legacy ref

Recovery and exit-code contract: unchanged from v0.1.2

Package matrix: Linux amd64/arm64, macOS amd64/arm64, and Windows amd64 (five archives)

Lifecycle conformance evidence: hermetic and pinned HTTPS provider gates
required, plus source-authoritative downstream CI rendering tests

Companion tag: v1.0.2

This corrective PATCH release removes the CLI's second source of truth for the
downstream workflow. It renders the fetched canonical
`templates/_shared/ci/downstream-ci.yml`, activates only selected job blocks,
rejects malformed or ambiguous markers, and retains a deterministic legacy
parser for older boilerplate sources. No manifest, catalog, profile, provider,
ownership, secret, or delivery contract changes.

The release supports existing lifecycle state from `v1.0.0` and `v1.0.1` and
is the full-verification client for companion boilerplate `v1.0.2`. Existing
`soku/v0.1.1` and `soku/v0.1.2` tags remain immutable and are not rewritten.

Verification requires renderer regression tests for every application stack,
formatter-compatible generated workflows, manifest selection reproduction,
hermetic lifecycle and state-transition tests, the public provider conformance
fixture at an immutable revision, the three-OS Go and lifecycle matrix,
quality/race/vet/format/import/lint checks, the five-target package snapshot,
runtime templates, repository security workflow, sync parity, and aggregate
`Validation Gate`.

Release preflight uses `boilerplate-tag` set to `v1.0.2` and `cli-tag` set to
`soku/v0.1.3`. The manual dispatch is validation-only; delivery is triggered by
the separately approved signed annotated `soku/v0.1.3` tag on the reviewed
source commit. The CLI Release receives five platform archives and
`checksums.txt`.
