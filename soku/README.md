# `soku` CLI

[Terminal guide: English](../docs/guides/SOKU_TERMINAL_GUIDE.md) |
[한국어](../docs/guides/SOKU_TERMINAL_GUIDE.ko.md) |
[日本語](../docs/guides/SOKU_TERMINAL_GUIDE.ja.md)

`soku` is the cross-platform command for the lifecycle contract in
[`SOKU_LIFECYCLE.md`](../docs/standards/SOKU_LIFECYCLE.md). It provides stable
parsing and output, transactional `init`, portable manifest-v1/v2/v3 records, and
read-only `status` diagnostics, immutable release comparison, and transactional
core upgrades.

The recommended full-verification baseline is the published boilerplate
`v1.0.5` with CLI `soku/v0.1.4`. Existing boilerplate and CLI tags, including
`v1.0.4` and `soku/v0.1.3`, remain immutable historical compatibility
baselines. The current distribution release is `soku/v0.2.1`, which preserves
that lifecycle compatibility contract. Human adopters should start with the
[end-to-end usage manual](../docs/guides/USAGE_MANUAL.md).

CLI distribution is available in two equivalent paths:

- GitHub releases (current baseline and all compatible CLI tags): download the
  matching `soku/vX.Y.Z` archive from the release assets.
- npm (`@soku-jinseok/soku`) from `soku/v0.2.0` onward:

  ```bash
  npm install -g @soku-jinseok/soku@0.2.1
  soku --version
  ```

The npm wrapper verifies `checksums.txt` for the selected native release and caches
the binary for your platform in your user cache.

## Terminal Output and Shell Completion

Human reports use the same titles, status words, fields, lists, and next-action
lines in every environment. On a TTY, `--color=auto` adds restrained color;
pipes remain stable plain text. `NO_COLOR` and `TERM=dumb` disable automatic
color. Use `--color=always` or `--color=never` for an explicit override. JSON,
quiet output, prompts, and generated completion scripts never depend on color.

Load completion for only the current session:

```bash
# bash
source <(soku completion bash)

# zsh
source <(soku completion zsh)

# fish
soku completion fish | source
```

```powershell
# PowerShell
soku completion powershell | Out-String | Invoke-Expression
```

To keep a generated script in a user-owned location, generate it explicitly:

```bash
# bash
mkdir -p ~/.local/share/bash-completion/completions
soku completion bash > ~/.local/share/bash-completion/completions/soku

# zsh (add ~/.zfunc to fpath in your own shell configuration)
mkdir -p ~/.zfunc
soku completion zsh > ~/.zfunc/_soku

# fish
mkdir -p ~/.config/fish/completions
soku completion fish > ~/.config/fish/completions/soku.fish
```

```powershell
$completionDirectory = Join-Path $HOME ".config/soku/completions"
New-Item -ItemType Directory -Force $completionDirectory | Out-Null
soku completion powershell | Set-Content (Join-Path $completionDirectory "soku.ps1")
```

Soku does not edit shell profiles or invoke a shell plugin manager. Source the
saved file from your own profile only if you choose to make it persistent.

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
  --boilerplate-release v1.0.5 \
  --stack javascript-typescript-node \
  --project-name example-service \
  --dry-run
```

The supported stack IDs are `javascript-typescript-node`, `python`, `go`,
`java-spring`, `mysql`, `postgresql`, `gcp`, `aws`, and `azure`. Repeat
`--stack` to select more than one; an explicit list replaces detection. Go requires `--module-path`, Java requires
`--java-group`, and Java/GCP service output accepts `--service-name`.

## CI/CD Decision Planning

The CI/CD decision layer is read-only until its reviewed adapter mappings are
available:

```bash
soku ci-cd plan --config .soku/ci-cd-decision.yml --json
```

The strict `ci-cd-decision-v1` input binds CI-only verification to the fixed
`ci-quick` and `full` profiles and requires `delivery.enabled: false`. Planning
records only the Git remote host, immutable HEAD/tree, and content hashes; it
does not inspect or change Cloud, IAM, credentials, runners, or repository
settings. `soku ci-cd init` accepts exactly one of `--dry-run` or `--yes`, but
returns a safety refusal while no reviewed mapping is published.

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
boilerplate_release: v1.0.5
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

## Explicit Project Ownership Handoff

Use a handoff only after reviewing one intentional modification to a
core-managed file:

```bash
soku ownership handoff \
  --path .prettierignore \
  --expected-sha256 <lowercase-64-character-sha256> \
  --dry-run
```

Apply the exact validated plan with `--yes` or interactive confirmation. The
command refuses clean, stale, missing, symlinked, obsolete, mergeable,
provider-managed, already project-owned, non-canonical, case-mismatched,
repeated, and batch paths. It never writes or changes the mode of the selected
file. A confirmed handoff migrates the manifest to v3, records the path in
`selection.project_owned_overrides`, and changes its file record to
`project-owned` / `unmanaged-expected`.

Same-release diff and future upgrades suppress core rendering for the recorded
path. Providers and components cannot claim it, and no implicit reclaim exists.
Dry-run writes no manifest, pending file, backup, or transaction journal.

## Optional GitHub Project Sync Component

Project synchronization is an opt-in first-party component. Plain `soku init`
does not install any Project Sync file. On a fresh repository, include the
component in the same complete plan:

```bash
soku init \
  --boilerplate-source https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate \
  --boilerplate-release v1.0.5 \
  --stack javascript-typescript-node \
  --project-name example-service \
  --project-sync \
  --project-sync-project-number 2 \
  --dry-run
```

On an already initialized repository, `--project-sync` installs only the
component and does not fetch or re-render the boilerplate:

```bash
soku init --project-sync --project-sync-project-number 2 --yes
```

Non-interactive use requires a positive Project number; interactive use may
enter it at the prompt. The component installs the guarded workflow, the
runtime and its focused test under core ownership, plus
`.github/project-sync.yml` as project-owned configuration. Its generated
configuration uses `owner: "@me"`, the selected number, canonical field names,
and no repository name, historical mappings, token, or Issue body. Existing
Project Sync files are reported as collisions and are never silently adopted.

The workflow is inactive until the repository variable
`PROJECT_SYNC_ENABLED=true` is set. Audit is the default runtime mode; apply
requires `PROJECT_SYNC_MODE=apply` or the explicit manual-dispatch choice.
Create `PROJECT_SYNC_TOKEN` manually with only repository Metadata read,
Issues read/write, Pull Requests read/write, and authenticated-user Projects
read/write permission. The CLI makes no GitHub API calls, creates no
credentials or Projects, and creates no cloud resources. Version 1 supports
authenticated user-owned Projects only; organization-owned Projects are a
follow-up scope. `status`, `diff`, and `upgrade` preserve the component while
leaving the project-owned configuration untouched.

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
contracts are [`schema/manifest-v1.schema.json`](./schema/manifest-v1.schema.json),
[`schema/manifest-v2.schema.json`](./schema/manifest-v2.schema.json), and
[`schema/manifest-v3.schema.json`](./schema/manifest-v3.schema.json), with
representative fixtures under the corresponding `testdata/manifest-v*/`
directories. Base initialization emits v1. An explicit opt-in
component installation migrates to v2 and adds only portable component ID,
catalog version, and configuration-path metadata. An ownership handoff
explicitly migrates v1/v2 to v3; later component installation preserves v3.
The record contains only
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

## Real-Runtime Manual Capture Component

The source tree includes the unreleased opt-in `docs-manual` component governed
by the
[real-runtime capture standard](../docs/standards/REAL_RUNTIME_MANUAL_CAPTURE.md).
It is not enabled by base initialization or CI.

```bash
soku docs manual plan --config docs/manual/capture.yml --json
soku docs manual doctor --config docs/manual/capture.yml
soku docs manual init --config docs/manual/capture.yml --dry-run
soku docs manual init --config docs/manual/capture.yml --yes
node tools/manual-capture/dist/cli.js capture \
  --config docs/manual/capture.yml
```

Initialization installs only the locked Node.js 22+ TypeScript/Playwright
runner, schemas, example configuration, and manual template. It does not create
the actual `capture.yml`, fixtures, prose, PNGs, report, or PDF, and it does not
run `npm ci` or install browsers, operating-system packages, or fonts.

The runner supports existing static builds, reviewed argv dev servers,
HTTP-route or sanitized-HAR replay, a GAS HTML Service harness with chainable
asynchronous `google.script.run`, disclosed dialog overlays, deterministic
maps, bounded local/manual Leaflet/OSM, and restricted/budgeted Google Maps
JavaScript. External egress is allowlisted, attribution is required, and only
the name `GOOGLE_MAPS_API_KEY` may enter configuration or reports.

This source feature does not itself authorize a public CLI/npm/boilerplate
release, live provider key use, recurring hosted capture, or manual/PDF
publication.

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
go install github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku@v0.1.4
```

## Verify a Release Download

Download the archive for the target platform together with `checksums.txt`,
then verify it before extraction. For example:

```bash
sha256sum --check --ignore-missing checksums.txt
tar -xzf soku_v0.1.4_linux_amd64.tar.gz
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
  --version v0.1.4 \
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
   release axis. The helper verifies clean, up-to-date `main`, requires the
   configured GPG key's full primary fingerprint to match
   `release-identity.json`, creates the local signed annotated tag, verifies it,
   and never pushes it.
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
