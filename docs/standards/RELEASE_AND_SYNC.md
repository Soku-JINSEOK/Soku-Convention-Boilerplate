# 🔄 Release and Sync

> **Applies to:** Team (multi-repository) — see [`docs/guides/APPLICABILITY.md`](../guides/APPLICABILITY.md). If you maintain a single personal project off this boilerplate, you can skip tag-pinning discipline; this matters once you sync updates across more than one downstream repository.

## 🎯 Purpose

This document defines how `Soku-Convention-Boilerplate` is versioned, released, and synchronized across downstream repositories.

This document governs release tags and the existing manual sync scripts. The
automated `soku` lifecycle is governed by
[`SOKU_LIFECYCLE.md`](./SOKU_LIFECYCLE.md), which additionally requires an
immutable source version, its resolved commit, and explicit compatibility and
migration checks. The scripts described here remain supported; they do not
implement or weaken that lifecycle contract.

## 📍 Source Of Truth

- This repository is the canonical source for the convention baseline.
- Downstream repositories should treat convention-owned files as imported policy, not local invention.
- If a downstream project overrides a rule, the override should be documented in that project.

## 🔢 Independent Release Axes

This repository publishes two independent version lines:

| Artifact | Tag form | Meaning |
| --- | --- | --- |
| Boilerplate convention package | `vMAJOR.MINOR.PATCH` | A versioned convention baseline for downstream synchronization and lifecycle source selection. |
| `soku` CLI | `soku/vMAJOR.MINOR.PATCH` | A signed source tag used to build the cross-platform CLI archives and GitHub Release. |

A tag on one axis never creates, advances, or implies a release on the other
axis. The versions may differ because CLI compatibility and boilerplate content
evolve independently. The CLI release procedure is documented in
[`soku/README.md`](../../soku/README.md).

Boilerplate releases use semantic-style tags in the form
`vMAJOR.MINOR.PATCH`.

- `PATCH`: documentation clarifications, typo fixes, CI hygiene, and non-behavioral template cleanup.
- `MINOR`: new non-breaking conventions, new templates, additional docs, or tooling that does not invalidate the existing contract.
- `MAJOR`: breaking file layout changes, renamed required files, incompatible policy changes, or changes that invalidate older downstream assumptions.

Each release must include a compatibility and migration record in its release
notes. A boilerplate `v*` release records:

- the immutable tag and resolved source commit
- the catalog contract versions it publishes, including the default and legacy
  profile interpretation
- the tested `soku` CLI compatibility range, or an explicit statement that no
  automated lifecycle migration is supported
- the supported forward migration from the previous release, including file
  ownership, removals, mergeable paths, and any manual action
- profile and provider compatibility changes

A CLI `soku/v*` release records:

- the immutable signed tag and source commit used for all archives
- the manifest schema versions it can inspect and mutate
- the boilerplate catalog and forward-migration ranges it supports
- the provider API/schema versions and profile rules it supports
- the operating-system package matrix, checksum artifact, and lifecycle
  conformance evidence
- any breaking behavior, read-only compatibility state, or manual recovery
  requirement

Unknown compatibility is unsupported; release notes must not infer support from
matching version numbers. A release is blocked if its required lifecycle
conformance, runtime template, quality/race, sync parity, or package snapshot
gate fails. These records document releases when they are intentionally made;
they do not authorize one release axis to create or advance the other.

The integrated Release workflow is the only automation that publishes GitHub
Releases. It reuses the full repository and runtime-template workflows for tag
events. Manual dispatch is a preflight: it validates both release records and
runs every gate, but its delivery job is structurally disabled. Tag delivery
requires a signed annotated tag, GitHub-verified signature, a matching source
commit in the annotation, and any declared companion tag to resolve to the same
commit. CLI releases alone receive the five platform archives and
`checksums.txt`; neither release axis is marked latest automatically.

Release tag helpers create and verify local tags but never push. When two axes
are intentionally released from one reviewed commit, publish them with one
atomic push. A public tag is immutable: never move, delete, or reuse it after a
failure. Correct the defect and publish the affected axis's next patch version.

## 📥 Downstream Sync Rules

- Pin downstream projects to a specific boilerplate tag.
- Record the consumed tag in the downstream README or setup notes.
- Sync only convention-owned files:
  - root policy docs
  - `.github/`
  - `templates/`
  - shared editor and ignore settings
- Do not copy the repository's `soku/` source directory through the manual sync
  scripts. Install the CLI from its independently versioned source or release
  archive instead.
- Leave application code, product docs, and environment-specific secrets to the downstream repository.

## 🔁 Sync Workflow

1. Review the latest boilerplate tag.
2. Compare the downstream state with that tag.
3. Apply the import with `scripts/sync-boilerplate.sh` (Linux/macOS) or `scripts/sync-boilerplate.ps1` (Windows).
4. Re-run CI and replace placeholder values where the downstream project requires them.
5. Commit the sync as a focused change.

## 💻 Recommended Command

### 🪟 Windows (PowerShell)

```powershell
pwsh ./scripts/sync-boilerplate.ps1 -TargetRoot C:\path\to\downstream -Force
```

Use `-IncludeReadme` only when the downstream repository is being bootstrapped from scratch.

### 🐧 Linux / macOS (bash)

```bash
./scripts/sync-boilerplate.sh --target /path/to/downstream --force
```

Use `--include-readme` only when the downstream repository is being bootstrapped from scratch.

Both scripts only copy git-tracked files from the source checkout (via `git ls-files`), so local build artifacts that happen to sit in the working tree (`node_modules/`, `dist/`, `__pycache__/`, and anything else covered by `.gitignore`) are never included even if present on disk. Both require the source to be a git repository. Preview what would be copied without touching the filesystem with `--dry-run` (bash) or `-WhatIf` (PowerShell's built-in dry-run mechanism). `scripts/verify-sync-parity.sh` runs both scripts against fresh temporary directories and diffs the output — it runs in CI on every change and can be run locally too.

## 🎬 Summary

The boilerplate stays reusable when releases are explicit and downstream sync is intentional.
