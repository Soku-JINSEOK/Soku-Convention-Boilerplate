# 🔁 `soku` Lifecycle Contract

## Status and Authority

- **Status:** Accepted
- **Decision owner:** `Soku-JINSEOK`
- **Decision record:** Issue #16 and its approved task report
- **Scope:** the public lifecycle contract for the `soku` convention-management
  system

This document is the normative architecture decision record for `soku`.
Implementation details may evolve, but implementations must preserve this
contract unless a later approved decision explicitly replaces it.

## Context

The repository already provides manual bootstrap and synchronization
procedures. Automation adds risks that those procedures do not need to model:
file ownership, local customization, compatibility across independently
versioned components, provider output, partial failure, and rollback.

The lifecycle therefore needs a stable contract before the CLI, manifest, and
mutation engine are implemented. The contract favors explicit immutable inputs,
reviewable plans, and non-destructive failure over convenience that makes an
upgrade difficult to reproduce.

## Decision Summary

`soku` will be a single cross-platform binary implemented in Go. Go provides a
practical path to self-contained binaries for Linux, macOS, and Windows without
requiring a language runtime in each downstream project. Public repositories
can build release artifacts on standard GitHub-hosted runners without a
separate paid runner requirement, and GitHub Releases supports distributing
binary assets. See GitHub's documentation for
[Actions billing](https://docs.github.com/en/billing/concepts/product-billing/github-actions)
and [release assets](https://docs.github.com/en/repositories/releasing-projects-on-github/about-releases).

The first public command surface is:

```text
soku init
soku status
soku diff
soku upgrade
soku --help
soku --version
```

The portable lifecycle record is `.soku/manifest.json`. Every mutation is
planned and validated before any managed state changes, then applied as one
transaction with the manifest replaced last.

## Lifecycle Terminology

| Term | Meaning |
| --- | --- |
| Source release | An immutable, explicitly selected boilerplate release together with the exact resolved source commit. |
| Downstream project | The repository into which `soku` plans or applies convention files. |
| Managed file | A repository-relative file whose owner, class, baseline hash, and lifecycle state are recorded in `.soku/manifest.json`. |
| `core-managed` file | A file rendered and owned exclusively by the boilerplate core. |
| `provider-managed` file | A file rendered and owned exclusively by one declared integration provider. |
| `mergeable` file | A managed file for which a declared deterministic merge strategy may preserve compatible local changes. |
| `project-owned` file | A file owned by the downstream project. It may appear in a plan for context but is never an automatic mutation target. |
| Baseline | The normalized content hash of a managed file immediately after the last successful transaction. |
| Local customization | A managed file whose current normalized hash differs from its recorded baseline. |
| Integration | A declarative, executable-free provider bundle pinned to an immutable source revision. |
| Lifecycle state | The status derived from the manifest, current files, selected source, and integration compatibility. |

## Command Contract

### Public Commands and Responsibilities

| Command | Responsibility | Mutation |
| --- | --- | --- |
| `init` | Detect project context, require an explicit boilerplate source and version, construct the initial ownership-aware plan, and create managed state. | Yes |
| `status` | Diagnose manifest compatibility, integration state, local customization, pending work, and drift without changing the repository. | No |
| `diff` | Render and compare the desired state with the current managed state without applying it. | No |
| `upgrade` | Require an explicit new boilerplate version or integration revision, plan the compatible transition, and apply it transactionally. | Yes |
| `--help` | Describe the supported command and option surface. | No |
| `--version` | Print the CLI version. | No |

Issue #17 may add subcommand-specific flags, but it must implement these names
and responsibilities without weakening their safety boundaries.

The core release-transition implementation delivered for Issue #20 accepts
`--boilerplate-release` on `diff` and `upgrade`, obtains the source only from the
durable manifest, and supports the manifest-v1 `standard` profile. Profile and
integration transitions remain within Issue #22's separately approved wire
format and compatibility boundary.

### Common Options

All commands use the following common option names where applicable:

| Option | Contract |
| --- | --- |
| `--config <yaml-path>` | Load an explicit portable YAML configuration file. |
| `--json` | Emit machine-readable output. Issue #17 defines the output schema. |
| `--quiet` | Suppress non-essential human-readable output. |
| `--non-interactive` | Forbid prompts. Missing required decisions are validation failures. |
| `--dry-run` | On a mutation command, perform fetch and all read-side validation and produce the complete plan without writing lifecycle state. |
| `--yes` | On a mutation command, approve application of the already validated plan without an interactive confirmation. |

Without `--yes`, a real mutation requires an interactive confirmation. In
non-interactive mode, a real mutation without `--yes` fails validation and does
not write. `--dry-run` never requires confirmation because it cannot mutate
managed files, the manifest, backups, or a transaction journal.

### Explicit Selection

`init` requires an explicit boilerplate source and immutable release version.
`upgrade` requires at least one explicit target: a new boilerplate version or a
new integration SHA. The following selection modes are prohibited:

- `latest` or any equivalent floating version
- branch tracking
- an implicit provider refresh
- resolving a mutable tag without recording its immutable resolved commit

The generic integration selection interface is reserved as follows:

```text
--integration-source github:<owner>/<repo>/<bundle-path>
--integration-ref <lowercase-40-character-sha>
--integration-config <yaml-path>
```

An integration ref must match `^[0-9a-f]{40}$`. A branch, tag, uppercase SHA,
abbreviated SHA, or overlong SHA is invalid. The source and ref are persisted;
raw integration configuration is not.

### Configuration Precedence

Lifecycle values are resolved in this order, from highest to lowest priority:

1. CLI flags
2. explicit configuration YAML
3. existing manifest state
4. project detection
5. built-in defaults

Environment variables are limited to output formatting and ambient credential
discovery. They must not select a boilerplate release, integration source,
integration revision, ownership rule, or other lifecycle behavior. Secrets and
credentials must not be copied from the environment into configuration or the
manifest.

### Exit Codes

| Code | Meaning |
| --- | --- |
| `0` | Success, clean state, or successful plan generation. |
| `1` | Internal error not represented by a more specific code. |
| `2` | Invocation, configuration, schema, ref, or path validation failed. |
| `3` | `diff` found changes, or `status` found pending or drifted state. |
| `4` | Conflict or safety refusal before mutation. |
| `5` | Compatibility failure. |
| `6` | Source, authentication, or fetch failure. |
| `7` | Apply failed and rollback restored the previous state. |
| `8` | Rollback failed; manual recovery is required. |

A dry-run that successfully produces a plan exits `0`, including when the plan
contains changes. `diff` uses `3` for a non-empty comparison so it can act as a
read-only automation gate.

## Independent Compatibility Axes

The following versions are independent and must not be treated as one shared
version:

1. `soku` CLI version
2. `.soku/manifest.json` schema version
3. boilerplate or template release version and its resolved commit
4. provider API version

The CLI must declare the manifest schema range, boilerplate migration range,
and provider API range it supports. Boilerplate and provider inputs must also
declare their compatibility requirements. Compatibility is checked before
rendering a mutable plan and again before application if any fetched input has
changed.

An unsupported manifest major version or provider API is read-only. `status`
must report `incompatible`; `diff` may report compatible diagnostics, but no
command may rewrite, migrate, or otherwise mutate lifecycle state until a
supported CLI and an explicit migration path are available. A mutation stops
with exit code `5`.

Migrations must be explicit, deterministic, and covered by rollback. A source
release is always recorded as both the immutable release identifier and the
resolved commit so the desired state remains reproducible even if external
metadata later changes.

## Portable Manifest Contract

The manifest path is fixed at `.soku/manifest.json`. The published JSON Schema
Draft 2020-12 wire contract is
[`soku/schema/manifest-v1.schema.json`](../../soku/schema/manifest-v1.schema.json),
with representative fixtures under
[`soku/testdata/manifest-v1/`](../../soku/testdata/manifest-v1/). Manifest v1
preserves the following meanings:

- manifest schema version
- `soku` version that completed the last successful apply
- boilerplate release identifier and resolved commit
- portable selections used to render the desired core state
- hash of canonical, non-secret lifecycle inputs
- for each managed file: canonical path, owner, class, and baseline hash
- for each integration: stable ID, source, full ref, provider API version,
  provider schema version, configuration hash, lifecycle state, and managed
  files

At minimum, integration state must preserve these meanings:

| State | Meaning |
| --- | --- |
| `pending` | A request exists, but compatible provider data for the exact request and ref is not yet available. |
| `connected` | Compatible provider data matches the exact request and its declared connected outputs are current. |
| `drifted` | Current managed content or provider data differs from the recorded baseline or desired state. |
| `incompatible` | The CLI, schema, API, source, or request cannot safely participate in mutation. |

Manifest v1 must not merge or remove these four semantic states.

The manifest must never contain:

- a token, password, credential, or other secret
- a machine-specific absolute path
- raw integration configuration
- an ambient credential reference that changes lifecycle selection

Only portable selections, immutable source identity, and canonical hashes are
stored for sensitive or environment-dependent inputs.

Files use the classes `core-managed`, `provider-managed`, `mergeable`, and
`project-owned`, and the lifecycle states `current`, `obsolete`, and
`unmanaged-expected`. Managed files require a `text` or `binary` content mode
and a baseline SHA-256. Project-owned files prohibit baseline fields. Text
baselines validate UTF-8 and normalize CRLF and bare CR to LF; they do not
change whitespace, a BOM, or the final newline. Binary baselines hash bytes
without normalization.

Manifest serialization sorts stacks, file paths, integration IDs, and managed
file references. It rejects absolute or traversing paths, backslash bypasses,
`.git` and `.soku` paths, Windows-incompatible components, case-insensitive
collisions, ambiguous owners, inconsistent integration references, secrets,
credential-bearing URLs, raw configuration, and environment-specific paths.

### Status and Recovery

`status` reads the last recorded snapshot and the current filesystem only. It
does not fetch a source or desired state, and it never creates, removes,
repairs, or replaces a file. Human output reports a summary and actionable
details. JSON output uses the existing single envelope and places ordered
manifest, boilerplate, count, path-sorted file, ID-sorted integration, and
guidance fields in `data`. A completed drift result (`3`) or compatibility
result (`5`) uses `ok: true`; validation or internal failure uses `ok: false`.

Writers stage mode-`0600` deterministic JSON in
`.soku/manifest.json.pending`, synchronize it, atomically replace the manifest
on the same filesystem, and synchronize the state directory where the platform
supports it. Windows replacement uses replace-existing and write-through
semantics. `status` reports a valid pending file as `recovery-required` with
exit `3` and leaves it untouched.

Explicit recovery applies only these unambiguous rules:

1. If a valid manifest and valid pending file both exist, retain the manifest
   and discard the pending file.
2. If only a valid pending file exists, promote it to the manifest.
3. Preserve malformed or ambiguous state and stop with exit `2`.

## Ownership and Drift Rules

Each managed output has exactly one owner and one class. Core output paths and
all provider output paths must be globally disjoint. A provider must not
declare or modify:

- `.soku/manifest.json`
- a core-owned path
- another provider's path
- `.git/` or any other reserved repository state
- a project-owned path

If the current hash of an existing managed file differs from its recorded
baseline, `soku` treats the difference as local customization. It must not
overwrite that file unless the file is `mergeable` and its declared
deterministic merge completes without conflict. Every other case stops before
mutation with exit code `4`.

Project-owned files may be referenced in diagnostics or shown in a plan, but
they are never automatically created, replaced, deleted, or merged.

## Canonical Paths and Content

Managed paths are stored as repository-relative POSIX paths using `/`. Before
planning, every source and output path is normalized and validated. `soku`
must reject:

- an absolute path, empty path, or `..` traversal
- a backslash-based traversal or separator bypass
- a path that enters `.git/`, `.soku/`, or other reserved state except for the
  core's final manifest replacement
- a symlink that escapes the repository or changes the resolved ownership
  boundary
- case-insensitive path collisions
- a component invalid on supported Windows filesystems
- a collision with a core, provider, or project-owned path

Text templates use UTF-8 and LF as the canonical hash representation. The
renderer normalizes supported text input before calculating its baseline hash
so line-ending differences across Linux, macOS, and Windows do not create false
drift. Binary content is hashed byte-for-byte.

## Plan and Confirmation Contract

Every mutation must complete all read-side work before the first write:

1. Resolve and fetch each immutable source.
2. Validate compatibility, schemas, refs, configuration, and canonical paths.
3. Resolve global ownership and reserved-path rules.
4. Detect local customization and merge conflicts.
5. Render the entire core and provider plan.
6. Present a reviewable plan with creates, updates, merges, deletions, ownership,
   integration-state changes, and manifest changes.
7. Obtain confirmation unless `--yes` was supplied.

A conflict is a planning result, not a recoverable apply event. The command
must stop before writing any managed file, backup, journal, or manifest.

`--dry-run` performs network access, authentication discovery, fetches, parsing,
rendering, compatibility checks, and every read-side validation needed to make
the plan truthful. It writes no managed file, manifest, backup, or transaction
journal.

## Transaction and Rollback Contract

A confirmed apply is one outer transaction across core and provider output:

1. Preserve the previous manifest and all managed paths that the plan may touch.
2. Stage and apply core file changes.
3. Render and apply each provider only within its validated, bounded outputs.
4. Revalidate the complete result and its hashes.
5. Atomically replace `.soku/manifest.json` last.
6. Remove transient transaction state only after the manifest replacement is
   durable.

Any failure before the final commit restores the previous manifest and every
touched managed path. An apply failure with successful rollback exits `7` and
must leave the repository at its pre-apply lifecycle state. Only a rollback
failure exits `8` and may leave a recovery-required record with bounded manual
recovery instructions. A normal conflict or validation failure must never
create recovery state.

## Executable-Free Provider Contract

A provider bundle is declarative data. It may contain only declared manifests,
schemas, text or binary templates, and bounded output declarations. The core
must not execute or dynamically load a remote script, hook, executable, shared
library, or binary field from a provider bundle.

The core owns:

- provider source and exact-ref validation
- fetch, parsing, and schema validation
- deterministic rendering
- path normalization, ownership checks, and conflict detection
- the outer transaction and rollback
- manifest, `status`, `diff`, and `upgrade` behavior

The provider owns:

- onboarding schema content
- registry-match rules expressed as declarative data
- pipeline and governance compatibility declarations
- its templates and conformance fixtures

Neither core nor provider may mutate a central registry as a side effect of a
downstream lifecycle command.

## Two-Stage Provider Lifecycle

A two-stage provider has a request phase and a connected phase:

1. The request phase may create only its declared request artifact and records
   the integration as `pending`.
2. The connected phase may create only declared connected outputs after
   provider data matches the request, schema/configuration hash, and exact full
   integration SHA.
3. If new provider data does not match the exact request or ref, the integration
   remains `pending`; `soku` must not render CI or delivery output.

This rule prevents a request for one governance or pipeline configuration from
being silently connected to data produced for another revision.

## Consumer Example and Deferred Wire Format

`ci-cd-control-plane-v1` is a consumer example, not a special core adapter. The
core contract must not contain repository-name conditions, provider-ID
branches, built-in control-plane adapters, or central-registry interpretation
logic.

Issue #22 will define the loader and provider API wire format. That work may
choose field names and serialization, but it must remain generic,
exact-SHA-pinned, executable-free, ownership-bounded, and compatible with the
transaction contract in this document.

The control-plane follow-up may coordinate only these provider-owned details:

- provider ID and API version
- compatible `soku` version range
- exact-SHA provider source
- phase-scoped outputs
- schema and configuration hash
- declared templates and output paths
- request and connected state
- conformance fixtures

## Implementation Handoff

| Issue | Required boundary |
| --- | --- |
| #17 | Implement the Go CLI shell, command/options surface, exit codes, and human/JSON output contract. |
| #19 | Define the portable manifest and status wire schemas without reducing the required semantic fields or states. |
| #18 and #20 | Implement complete planning, confirmation, mutation, conflict handling, the outer transaction, and rollback. |
| #21 | Test the supported operating-system, release, manifest, and provider compatibility matrix. |
| #22 | Implement the generic executable-free provider loader and API wire format without consumer-specific branching. |

## Conformance Scenarios

Implementations must cover at least these scenarios:

| Scenario | Required result |
| --- | --- |
| Integration ref is a lowercase 40-character SHA | Accepted if source and compatibility validation also pass. |
| Integration ref is a branch, tag, uppercase SHA, or the wrong length | Rejected before mutation with exit code `2`. |
| Provider output uses traversal, reserved state, or an owned path | Rejected before mutation with exit code `2` or `4`, according to whether the failure is path validity or ownership conflict. |
| Configuration or manifest would persist a secret | Rejected before mutation with exit code `2`. |
| Exact provider data is not available for a valid request | State remains `pending`; connected CI or delivery output is absent. |
| Exact compatible provider data arrives | A reviewed transaction may move state to `connected` and create only declared connected outputs. |
| A managed file differs from its baseline without a successful declared merge | Conflict exits `4` before any write. |
| Provider application fails after core staging and rollback succeeds | All touched files and the prior manifest are restored; exit code is `7`. |
| Provider application and rollback both fail | Exit code is `8`, and bounded recovery information remains for manual action. |

## Acceptance-Criteria Traceability

| Issue #16 acceptance criterion | Normative section |
| --- | --- |
| CLI packaging, command boundaries, configuration precedence, and exit codes | Decision Summary; Command Contract |
| Lifecycle terminology | Lifecycle Terminology |
| Initial command responsibilities | Public Commands and Responsibilities |
| Compatibility rules | Independent Compatibility Axes; Portable Manifest Contract |
| Non-destructive defaults, confirmation, dry-run, and rollback | Ownership and Drift Rules; Plan and Confirmation Contract; Transaction and Rollback Contract |
| Align existing authority without duplication | Status and Authority; Consumer Example and Deferred Wire Format; linked updates in `BLUEPRINT.md`, `INIT_GUIDE.md`, and `RELEASE_AND_SYNC.md` |
| Approved task report before implementation | `docs/issues/issue-16-task-report.md` |

## Consequences

The decision adds upfront schema, planning, and compatibility work, and it
forbids convenient floating upgrades or executable provider hooks. In return,
an operation is reproducible from immutable inputs, local customization is not
silently destroyed, providers cannot escape declared ownership, and partial
failure has one defined rollback boundary across supported operating systems.
