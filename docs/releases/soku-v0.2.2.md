# soku v0.2.2

Release axis: soku

Manifest schemas: manifest-v1 (unchanged)

Boilerplate compatibility: v1.0.5

Profiles: bootstrap, standard (default and legacy), and scaled

Provider compatibility: provider-v1 with optional deprecated legacy ref

Recovery and exit-code contract: unchanged from soku/v0.1.4

Package matrix: Linux amd64/arm64, macOS amd64/arm64, and Windows amd64 (five archives)

Lifecycle conformance evidence: unchanged native archive smoke profile, three-OS CLI
runtime gates, and lifecycle smoke migration matrix over supported upgrade paths

Package and distribution: unchanged package matrix and checksum artifact, plus a
new npm wrapper package `@soku-jinseok/soku` for launch-time installation and
execution of `soku/v0.2.2`

Any breaking behavior, read-only compatibility state, or manual recovery requirement: unchanged

Companion tag: none

This release keeps the same native behavior as `soku/v0.1.4` while adding a
distribution pathway for users who prefer npm tooling. The CLI wrapper lives
under `soku/npm`, validates download integrity from `checksums.txt`, caches a local
copy of the native executable for each target platform, and executes it
directly for each user request. No behavioral compatibility promises are changed:
manifest schema, provider API, profile behavior, lifecycle contracts, and
migration semantics remain the same.

Verification requires existing `soku/v0.2.2` artifacts for all five platform
archives, checksum validation, and end-to-end CLI smoke against release download
and launcher execution.
