# 📝 Task Report: Language Selection Guide

## Goal and Background

`docs/guides/STACK_EXAMPLES.md` and `docs/guides/STACK_CONFIGS.md` explain how to configure a stack once chosen, but nothing in this boilerplate explains how to choose one — `docs/guides/INIT_GUIDE.md` tells an AI agent to "ask the user which stack(s) to bootstrap" with no criteria to reason from. Closes #14.

The source material was a personal Korean-language language-selection guide the user maintains, which needed to be adapted rather than copied verbatim into this boilerplate.

## Proposed Approach

Add `docs/guides/LANGUAGE_SELECTION.md`, English only, following this repository's Language Policy (`BLUEPRINT.md`): trilingual treatment is reserved for public-facing overview content (`README.*`), and all 16 existing `docs/` files are English-only. Only the three README index entries are translated.

Two content decisions, confirmed with the user before implementation:

- Drop personal/portfolio content from the source (career planning, a specific Discord-bot project, a personal stack-tracking log) — `scripts/sync-boilerplate.sh` copies all of `docs/` to every downstream repository, so anything kept here needed to be reusable, not autobiographical.
- Keep the source's ADR-style decision template as one section inside the guide rather than introducing a repository-wide ADR convention — none exists today (verified: zero hits for "ADR" / "decision record" / `docs/adr/` across the repo).

Structural constraint discovered during research: `.markdownlint.jsonc` disables only `MD013`; `MD024` (no-duplicate-heading, `siblings_only: false`) is on. The source's per-language subheading pattern (14 languages × 4 repeated subheadings) would fail it and would also run to roughly 700 lines against this repo's 130–245-line peer docs. Both problems share one fix: collapse the per-language breakdown into a single table, matching this repository's existing table-first style (`CODE_STYLE.md`, `APPLICABILITY.md`).

## Planned Implementation

- `docs/guides/LANGUAGE_SELECTION.md` (new): constraint categories, a pre-decision checklist, a quick selection table, a per-language fit table, an infrastructure-language table, a trimmed cross-link to `CODE_STYLE.md` instead of restating it, a 7-step selection procedure, an inline ADR-style template, a pre-addition checklist, and anti-patterns.
- Registration, per `docs/guides/APPLICABILITY.md`'s Maintenance Rule:
  - `BLUEPRINT.md` — Authority Model → Reference list, and Repository Shape → `docs/guides/` listing.
  - `README.md` / `README.ko.md` / `README.ja.md` — one index bullet each, in that file's own language.
  - `.github/workflows/ci.yml` — `required_files`.
  - `docs/guides/APPLICABILITY.md` — new table row, classified `Both`.
  - `docs/guides/INIT_GUIDE.md` and `docs/guides/STACK_CONFIGS.md` — one cross-link sentence each.

## Acceptance Criteria

- `npx markdownlint-cli2` reports 0 errors across the repository.
- `npx yaml-lint` passes on the modified `ci.yml`.
- Every relative link inside the new document resolves to an existing file.
- The `repository-hygiene` `required_files` existence check passes for the new path.
- The file is git-tracked and `scripts/sync-boilerplate.sh` copies it to a fresh target directory (proving it reaches downstream repositories).

## Approval

- **Status:** `Approved`
- **Approved by:** repository owner (plan reviewed and approved via the plan-mode workflow before implementation started)

## Implementation Status

Complete. All seven registration points and the new document were written, staged, and verified locally before commit.

## Verification

- `npx --yes markdownlint-cli2@0.22.1 --config .markdownlint.jsonc "**/*.md" "#node_modules"` — 38 files linted, 0 errors.
- `npx --yes yaml-lint@1.7.0 .github/workflows/*.yml` — pass.
- Manually traced all four relative links in the new document (`./APPLICABILITY.md`, `./INIT_GUIDE.md`, `../standards/CODE_STYLE.md`, `../issues/TASK_REPORT_TEMPLATE.md`) and the cross-links added to `INIT_GUIDE.md`/`STACK_CONFIGS.md` — all resolve. (CI has no markdown link-checker, so this was checked by hand rather than by an automated gate.)
- Reconstructed the `repository-hygiene` `required_files` existence loop locally against all 81 entries — all present.
- `git add`-ed the new file, then ran `scripts/sync-boilerplate.sh --target <tmp>` (this branch predates PR #11's `--dry-run` flag, so a real sync to a scratch temp directory was used instead) and confirmed `docs/guides/LANGUAGE_SELECTION.md` appeared at the target path.
- Not run: `templates-ci.yml` / `sync-parity` (no template or sync-script files changed in this branch); this is left to CI on the resulting PR.

## AI Assistance

- **Planning/implementation/drafting:** `Claude Code`
