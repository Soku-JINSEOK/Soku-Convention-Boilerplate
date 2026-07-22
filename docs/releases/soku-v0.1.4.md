# soku v0.1.4

Release axis: soku

Manifest schemas: manifest-v1 (unchanged)

Boilerplate compatibility: v1.0.0 through v1.0.3; v1.0.3 is the recommended
companion baseline

Profiles: bootstrap, standard (default and legacy), and scaled

Provider compatibility: provider-v1 with optional deprecated legacy ref

Recovery and exit-code contract: unchanged from v0.1.3

Package matrix: Linux amd64/arm64, macOS amd64/arm64, and Windows amd64 (five archives)

Lifecycle conformance evidence: hermetic and pinned HTTPS provider gates
required, plus source-authoritative downstream CI rendering tests

Companion tag: v1.0.3

This corrective PATCH release preserves the accepted downstream baseline when
an unchanged mergeable file is carried through a lifecycle transition. The
published `soku/v0.1.3` can preserve customized `.editorconfig` or `.gitignore`
bytes while replacing their recorded baseline with the upstream hash, which
causes false drift after a same-release provider transition. The immutable
`soku/v0.1.3` tag and archives remain unchanged.

The correction keeps both the customized bytes and their accepted baseline
when the mergeable path is unchanged. Integration coverage exercises the
pending-to-connected transition for `.editorconfig` and `.gitignore`. Manifest
schema, catalog, profile, provider, ownership, recovery, exit-code, and delivery
contracts do not change, and no state migration or manual recovery is required.

Verification requires manifest and mergeable-baseline regressions, hermetic
lifecycle and provider state-transition tests, the three-OS Go and lifecycle
matrix, quality/race/vet/format/import/lint checks, the five-target package
snapshot, runtime templates, repository security, sync parity, and the
aggregate `Validation Gate`. Public smoke must cover fresh four-application-
stack `init --yes --verify`, status, same-release diff, no-op upgrade, and
migrations from `v1.0.0`, `v1.0.1`, and `v1.0.2` to companion `v1.0.3`.

Release preflight uses `boilerplate-tag` set to `v1.0.3` and `cli-tag` set to
`soku/v0.1.4`. Manual dispatch is validation-only and its publish job must be
skipped. Delivery requires separately approved signed annotated companion tags
on the same reviewed source commit. The CLI Release receives five platform
archives and `checksums.txt`.
