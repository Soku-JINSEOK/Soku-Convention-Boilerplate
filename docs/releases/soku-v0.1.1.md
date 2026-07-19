# soku v0.1.1

Release axis: soku

Manifest schemas: manifest-v1

Boilerplate compatibility: v1.0.0

Profiles: bootstrap, standard (default and legacy), and scaled

Provider compatibility: provider-v1

Recovery and exit-code contract: unchanged from v0.1.0

Package matrix: linux amd64/arm64, macOS amd64/arm64, Windows amd64

Lifecycle conformance evidence: required

Companion tag: none

This patch supersedes `soku/v0.1.0` for lifecycle operations. The initial CLI
release rejected real GitHub source archives because their POSIX PAX global
metadata entry was incorrectly treated as a second file root. It also treated
the repository's deliberately invalid manifest fixtures as real credentials,
and a same-release transition reformatted an unchanged mergeable file instead
of reporting a no-op. The patch ignores archive metadata, exempts only the
non-rendered `soku/testdata/` fixture tree from secret scanning, and detects an
already converged baseline before running a three-way merge. Traversal, link,
collision, size, portability, and rendered-source secret checks remain active.

There are no manifest, catalog, profile, provider, recovery, exit-code, or
package format changes. Projects created by other means with `soku/v0.1.0`
remain manifest-v1 compatible. Use `soku/v0.1.1` for `init`, `diff`, and
`upgrade` against boilerplate `v1.0.0`.
