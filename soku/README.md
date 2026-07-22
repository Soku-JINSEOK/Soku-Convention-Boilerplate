# `soku` CLI

`soku` is the cross-platform command for the lifecycle contract in
[`SOKU_LIFECYCLE.md`](../docs/standards/SOKU_LIFECYCLE.md). It provides stable
parsing and output, transactional `init`, the portable manifest-v1 record, and
read-only `status` diagnostics, immutable release comparison, and transactional
core upgrades.

The recommended full-verification baseline is boilerplate `v1.0.3` with
`soku/v0.1.3` after the signed public `v1.0.3` Release exists. Existing
`v1.0.0`, `v1.0.1`, `soku/v0.1.1`, and `soku/v0.1.2` objects remain immutable
historical compatibility baselines.

## Transactional Init

`soku init` accepts only a public GitHub HTTPS source and an exact, non-prerelease
`vMAJOR.MINOR.PATCH`. It resolves the tag through the GitHub API to a full commit,
validates the bounded source archive and `catalog/core-v1.json`, renders the
complete plan, and writes the manifest last. A real non-interactive or JSON
mutation requires `--yes`; `--json --dry-run` emits one plan envelope and writes
nothing.

```bash
soku init \
  --boilerplate-source https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate \
  --boilerplate-release v1.0.3 \
  --stack javascript-typescript-node \
  --project-name example-service \
  --dry-run
```

The supported stack IDs are `javascript-typescript-node`, `python`, `go`,
`java-spring`, `mysql`, `postgresql`, `gcp`, `aws`, and `azure`. Repeat
`--stack` to select more than one; an explicit list replaces detection. Go requires `--module-path`, Java requires
`--java-group`, and Java/GCP service output accepts `--service-name`.

## Profiles

Catalog v2 composes three built-in profiles in one fixed order:

| Profile | Composition | Typical use |
| --- | --- | --- |
| `bootstrap` | `bootstrap` | Personal-minimal projects and early experiments. |
| `standard` | `bootstrap → standard` | Team-standard projects; this is the default and legacy-compatible ID. |
| `scaled` | `bootstrap → standard → scaled` | Scaled collaboration with core agent and ownership policy files. |

CLI flags override explicit YAML, and explicit YAML overrides manifest state.
An immutable source without `soku/catalog/index-v2.json` is interpreted as
legacy core-v1 and supports only `standard`. Profile changes are reviewable with
`diff --profile <id>` and apply through the same outer transaction with
`upgrade --profile <id>`; both commands still require an exact release.

AI collaboration is not a fourth profile. The declarative example under
`providers/ai-collaboration/` can combine with all three profiles.

## Bounded Integrations

Initialization, diff, and upgrade accept the generic provider inputs:

```bash
--integration-source github:<owner>/<repo>/<bundle-path>
--integration-ref <lowercase-40-character-commit>
--integration-config <yaml-path>
```

Provider API v1 permits only versioned metadata, a hashed configuration schema,
sorted compatible profiles, declared templates, and bounded text or binary
outputs. The exact lowercase full commit passed with `--integration-ref` and
used for fetch is the only authoritative revision in the request artifact,
manifest, and connection decision. A bundle may omit its deprecated legacy
`ref`; if present, that value must be well-formed but matching or mismatching it
has no effect on the fetched revision. Unknown fields, malformed legacy refs,
scripts, hooks, executable or dynamic-library paths, undeclared bundle files,
traversal, reserved state, secrets, and ownership collisions fail before
writes. Raw configuration is never stored.

If the exact source, ref, and configuration hash has no matching bundle, `soku`
creates only `.github/soku/integrations/<id>.json` and records `pending`. An
exact compatible bundle adds only its declared outputs and records `connected`.
Pending-to-connected and profile/provider changes use the same manifest-last
transaction and rollback boundary as core upgrades.

The public mirror includes exact registered bundles for `cutvi`, `archviz`,
`report-hub`, and `soku-pr-site` under `providers/<project>-control-plane-v1/`.
They share the generic loader and differ only through their reviewed metadata,
configuration schema, configuration bytes, and literal output. The provenance
ledger at `providers/provenance/registered-downstream-v1.json` binds the
control-plane merge and all public bytes. No caller is enabled automatically.

The pending artifact contains exactly `schema_version`, `id`, portable
`source`, authoritative `ref`, and `configuration_hash`. A sanitized
configuration can be submitted only through a provider-owned channel outside
the lifecycle: remove secrets and validate the schema locally, compare its
canonical hash with the pending artifact, submit it with the portable source,
exact requested commit, and hash, then wait for the provider to publish a new
immutable commit. The user must explicitly select that commit. Neither the
pending artifact nor `.soku/manifest.json` stores sanitized/raw configuration
or secrets.

The equivalent strict YAML file is a flat mapping. Unknown fields are rejected:

```yaml
schema_version: 1
boilerplate_source: https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate
boilerplate_release: v1.0.3
stacks:
  - go
  - postgresql
profile: standard
project_name: example-service
module_path: github.com/example/example-service
java_group: io.example
service_name: example-service
verify: false
```

Only `.gitignore` and `.editorconfig` are mergeable on first initialization.
Any other existing selected output is treated as project-owned and stops with
exit `4` before a journal, backup, managed file, or manifest is written. Optional
`--verify` runs only built-in argv sequences against an isolated staging tree.
Apply failure with complete rollback exits `7`; incomplete rollback retains the
mode-restricted journal and exits `8` with recovery data.

## Diff and Upgrade

Run release transitions from an initialized project with the manifest's
recorded source. A transition cannot select a different source, track a branch,
or downgrade:

```bash
soku diff --boilerplate-release v1.1.0
soku upgrade --boilerplate-release v1.1.0 --dry-run
soku upgrade --boilerplate-release v1.1.0 --yes
```

Both the recorded release and target tag must resolve to their immutable
40-character commits. `diff` writes nothing and exits `3` when either managed
content or the release identity would change; it exits `0` for an exact no-op.
An upgrade dry-run performs the same complete read-side validation but always
exits `0` after producing a valid plan.

Plans list paths in order as `added`, `updated`, `removed`, `merged`,
`unchanged`, `locally-modified`, or `conflict`. Core-managed drift and
project-owned collisions stop with exit `4`. `.gitignore` is merged as a line
set and `.editorconfig` by section and key so independent local entries survive
a compatible forward transition. Creates, replacements, merges, removals, and
the prior manifest share one backup journal; the target manifest is replaced
last. A clean upgrade to the already recorded release is a no-op.

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

## Lifecycle Conformance Release Gate

The hermetic package under `internal/lifecyclee2e` injects synthetic immutable
source releases and verifies empty, existing, single-stack, and multi-stack
repositories through initialization, status, local customization, diff,
upgrade, rollback, rerun, and final clean status. It performs no real tag or
network operation.

CI runs this package on Linux, macOS, and Windows for pull requests and `main`.
The integrated Release workflow reuses the complete CI matrix for boilerplate
`v*` and CLI `soku/v*` tags before either GitHub Release can be published.
Platform-aware cases cover canonical line endings,
case-insensitive collisions, symlink boundaries where available, atomic
manifest replacement, and deletion rollback. A failure retains a path-sanitized
log for three days; successful runs retain no lifecycle artifact.

Linux template jobs remain the runtime gate for generated JavaScript/TypeScript,
Python, Go, and Java projects and run whenever template or `soku` rendering code
changes. The same three-OS package covers all profile/provider combinations,
pending-to-connected state, combined release/profile/provider upgrades,
ownership conflicts, and unsupported provider or manifest compatibility.

For a published immutable release, Go understands the repository's submodule
tag and installs it by module version:

```bash
go install github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku@v0.1.2
```

## Verify a Release Download

Download the archive for the target platform together with `checksums.txt`,
then verify it before extraction. For example:

```bash
sha256sum --check --ignore-missing checksums.txt
tar -xzf soku_v0.1.2_linux_amd64.tar.gz
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
  --version v0.1.2 \
  --commit 0123456789abcdef0123456789abcdef01234567 \
  --built-at 2026-07-18T00:00:00Z \
  --output-dir ./dist
```

Each archive contains the executable, the project `LICENSE`, and
`THIRD_PARTY_NOTICES.md`. `checksums.txt` lists the five archives in sorted
filename order.

## Release Procedure

The CLI and boilerplate use independent signed, annotated tags. Boilerplate
policy releases use `v*`; CLI releases use `soku/v*`. A manual Release workflow
dispatch is a validation-only preflight and never creates a tag or GitHub
Release. Before creating release tags:

1. Prepare the CLI compatibility and migration record required by
   [`RELEASE_AND_SYNC.md`](../docs/standards/RELEASE_AND_SYNC.md), including
   manifest, catalog, provider API, profile, and recovery boundaries.
2. Verify the version and supported Go toolchain.
3. Run the complete repository and package verification suite.
4. Run `scripts/create-release-tag.sh --tag <tag> --notes-file <path>` for each
   release axis. The helper verifies clean, up-to-date `main`, creates the local
   signed annotated tag, verifies it, and never pushes it.
5. Verify companion tags resolve to the same reviewed commit, then publish them
   together with `git push --atomic origin <boilerplate-tag> <cli-tag>`.
6. The guarded Release workflow reuses full repository and runtime-template CI,
   verifies both Git and GitHub signature status, and creates one GitHub Release
   for each tag. Only the CLI release receives the five archives and checksum
   file, built from the exact tagged commit.

Published tags are immutable. If a gate fails after publication, do not move,
delete, or reuse a public tag; fix the defect and issue the affected axis's next
patch version.

This workflow is designed for a public repository using standard GitHub-hosted
runners and GitHub Release assets. It does not require larger runners, a paid
package registry, GoReleaser, or a separate artifact service. Repository usage
and GitHub plan limits remain the operator's responsibility.
