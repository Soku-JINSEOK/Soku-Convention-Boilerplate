# 🔄 Release and Sync

> **Applies to:** Team (multi-repository) — see [`docs/guides/APPLICABILITY.md`](../guides/APPLICABILITY.md). If you maintain a single personal project off this boilerplate, you can skip tag-pinning discipline; this matters once you sync updates across more than one downstream repository.

## 🎯 Purpose

This document defines how `Soku-Convention-Boilerplate` is versioned, released, and synchronized across downstream repositories.

## 📍 Source Of Truth

- This repository is the canonical source for the convention baseline.
- Downstream repositories should treat convention-owned files as imported policy, not local invention.
- If a downstream project overrides a rule, the override should be documented in that project.

## 🔢 Versioning

Releases should use semantic-style tags in the form `vMAJOR.MINOR.PATCH`.

- `PATCH`: documentation clarifications, typo fixes, CI hygiene, and non-behavioral template cleanup.
- `MINOR`: new non-breaking conventions, new templates, additional docs, or tooling that does not invalidate the existing contract.
- `MAJOR`: breaking file layout changes, renamed required files, incompatible policy changes, or changes that invalidate older downstream assumptions.

Each release should include a short summary of what downstream repositories need to review after upgrading.

## 📥 Downstream Sync Rules

- Pin downstream projects to a specific boilerplate tag.
- Record the consumed tag in the downstream README or setup notes.
- Sync only convention-owned files:
  - root policy docs
  - `.github/`
  - `templates/`
  - shared editor and ignore settings
- Leave application code, product docs, and environment-specific secrets to the downstream repository.

## 🔁 Sync Workflow

1. Review the latest boilerplate tag.
2. Compare the downstream state with that tag.
3. Apply the import with `scripts/sync-boilerplate.ps1`.
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

## 🎬 Summary

The boilerplate stays reusable when releases are explicit and downstream sync is intentional.
