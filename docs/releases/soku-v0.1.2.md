# soku v0.1.2

Release axis: soku

Manifest schemas: manifest-v1

Boilerplate compatibility: v1.0.0

Profiles: bootstrap, standard (default and legacy), and scaled

Provider compatibility: provider-v1 with optional deprecated bundle ref

Recovery and exit-code contract: unchanged from v0.1.1

Package matrix: linux amd64/arm64, macOS amd64/arm64, Windows amd64

Lifecycle conformance evidence: required, including an immutable external provider commit

Companion tag: none

This patch makes the exact lowercase commit supplied through
`--integration-ref` and used for the provider archive fetch the sole revision
authority. Provider API v1 no longer requires a bundle-declared `ref`. A
well-formed legacy `ref` remains readable but cannot select `pending` or
`connected`; a malformed present value remains incompatible.

Pending request artifacts remain hash-only for integration configuration. A
separate provider-controlled sanitized configuration review is documented for
cases where onboarding requires human-readable non-secret fields. There are no
manifest, catalog, profile, recovery, exit-code, or package format changes.
Existing manifest-v1 integrations remain compatible and require no migration.
