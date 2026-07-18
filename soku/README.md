# `soku` CLI

`soku` is the cross-platform command for the lifecycle contract in
[`SOKU_LIFECYCLE.md`](../docs/standards/SOKU_LIFECYCLE.md). This first release
provides stable parsing, output, safety validation, packaging boundaries, the
portable manifest-v1 record, and read-only `soku status` diagnostics. `init`,
`diff`, and `upgrade` intentionally continue to report `feature.unavailable`
until their roadmap issues implement planning and mutation.

## Manifest and Status

The durable record is `.soku/manifest.json`. Its JSON Schema Draft 2020-12
contract is [`schema/manifest-v1.schema.json`](./schema/manifest-v1.schema.json),
with representative [valid](./testdata/manifest-v1/valid/complete.json) and
[invalid](./testdata/manifest-v1/invalid/) fixtures. The record contains only
portable selections, immutable source identities, ownership metadata, and
canonical hashes. Raw configuration, secrets, credential-bearing URLs, and
machine-specific absolute paths are rejected.

Run `soku status` from the repository root. Human output includes a summary and
actionable diagnostics; `--quiet` suppresses that normal output, and `--json`
always emits exactly one ordered `{ok, command, error, data}` envelope. Status
never fetches, repairs, removes, or changes repository content.

| Exit | `status` meaning |
| --- | --- |
| `0` | The validated snapshot and current managed files are clean. |
| `1` | An unexpected handler or store failure occurred. |
| `2` | Manifest, path, hash, or readable-state validation failed. |
| `3` | State is uninitialized, recovery-required, pending, or drifted. |
| `5` | The manifest or recorded provider state is incompatible. |

Completed diagnostic results with exit `3` or `5` use `ok: true` in JSON.
Validation and internal failures use `ok: false`.

Manifest writes stage deterministic mode-`0600` JSON at
`.soku/manifest.json.pending`, synchronize it, and atomically replace the
durable manifest. If `status` reports `recovery-required`, preserve both files.
An explicit `Store.Recover` or a future mutation entrypoint may discard a valid
pending file beside a valid manifest, or promote a valid pending file when the
manifest is absent. Malformed or ambiguous evidence is preserved and recovery
stops with exit `2`.

## Build and Test

Go 1.26 or newer is required.

```bash
cd soku
go mod verify
go test ./...
go build -o ./bin/soku .
./bin/soku --help
./bin/soku --version
./bin/soku status
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
