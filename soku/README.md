# `soku` CLI

`soku` is the cross-platform command shell for the lifecycle contract in
[`SOKU_LIFECYCLE.md`](../docs/standards/SOKU_LIFECYCLE.md). This first release
provides stable parsing, output, safety validation, and packaging boundaries.
The lifecycle handlers intentionally report `feature.unavailable` until their
own roadmap issues implement managed state.

## Build and Test

Go 1.26 or newer is required.

```bash
cd soku
go mod verify
go test ./...
go build -o ./bin/soku .
./bin/soku --help
./bin/soku --version
```

Use a temporary `GOBIN` to test local installation without changing a user-wide
Go configuration:

```bash
cd soku
temporary_gobin="$(mktemp -d)"
GOBIN="$temporary_gobin" go install .
"$temporary_gobin/soku" --version
```

For a published immutable release, Go understands the repository's submodule
tag and installs it by module version:

```bash
go install github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku@v0.1.0
```

## Verify a Release Download

Download the archive for the target platform together with `checksums.txt`,
then verify it before extraction. For example:

```bash
sha256sum --check --ignore-missing checksums.txt
tar -xzf soku_v0.1.0_linux_amd64.tar.gz
./soku --version
```

On macOS, replace `sha256sum` with `shasum -a 256`. On Windows, compare
`Get-FileHash -Algorithm SHA256` with the corresponding line in
`checksums.txt`.

## Package a Snapshot

The package script requires explicit build metadata and produces Linux amd64,
Linux arm64, macOS amd64, macOS arm64, and Windows amd64 archives:

```bash
cd soku
./scripts/package.sh \
  --version v0.1.0 \
  --commit 0123456789abcdef0123456789abcdef01234567 \
  --built-at 2026-07-18T00:00:00Z \
  --output-dir ./dist
```

Each archive contains the executable, the project `LICENSE`, and
`THIRD_PARTY_NOTICES.md`. `checksums.txt` lists the five archives in sorted
filename order.

## Release Procedure

The CLI and boilerplate use independent tags. Boilerplate policy releases use
`v*`; CLI releases use signed, annotated `soku/v*` tags. Before creating a CLI
tag:

1. Verify the version and supported Go toolchain.
2. Run the complete repository and package verification suite.
3. Create and verify a signed tag, for example
   `git tag -s soku/v0.1.0 -m "soku v0.1.0"` and
   `git tag -v soku/v0.1.0`.
4. Push the tag only after review. The guarded release job packages the tag's
   exact commit and creates the GitHub Release from the same script used in CI.

This workflow is designed for a public repository using standard GitHub-hosted
runners and GitHub Release assets. It does not require larger runners, a paid
package registry, GoReleaser, or a separate artifact service. Repository usage
and GitHub plan limits remain the operator's responsibility.
