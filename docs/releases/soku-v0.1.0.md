# soku v0.1.0

Release axis: soku

Manifest schemas: manifest-v1

Boilerplate compatibility: v1.0.0

Profiles: bootstrap, standard (default and legacy), and scaled

Provider compatibility: provider-v1

Recovery and exit-code contract: documented

Package matrix: linux amd64/arm64, macOS amd64/arm64, Windows amd64

Lifecycle conformance evidence: required

Companion tag: v1.0.0

This first CLI release inspects and mutates manifest-v1 projects pinned to the
boilerplate v1.0.0 catalog contracts. It supports all three built-in profiles,
legacy `standard` interpretation, and provider-v1 integrations.

Mutations are transactional. Recovery-required states and the stable exit-code
boundaries documented in `soku/README.md` are part of this release contract.
The release contains five archives—Linux amd64/arm64, macOS amd64/arm64, and
Windows amd64—and one `checksums.txt` file. Lifecycle conformance, package
snapshot, quality/race, sync parity, and runtime template gates must all pass
before publication. Compatibility outside the versions named here is not
implied.
