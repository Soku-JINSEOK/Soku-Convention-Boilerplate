# 🧑‍💻 Applicability

## 🎯 Purpose

This boilerplate was originally written with team-scale collaboration in mind, so some of its documents assume things that do not hold for a solo/personal project: a second reviewer, a shared release cadence, multi-account cloud governance. This document audits which parts of the boilerplate apply to a **Personal** project, which assume a **Team**, and which apply to **Both** — so an individual adopting this boilerplate can tell what to keep, what to skip, and what to defer.

This maps onto the existing [Maturity Levels in BLUEPRINT.md](../../BLUEPRINT.md#maturity-levels): a Personal project typically stays at **Bootstrap**, a small team moves into **Standard**, and multi-team or regulated environments reach **Scaled**.

## 📋 How to Read This Table

- **Personal** — applies and is worth keeping even solo; skipping it has a real cost.
- **Team** — assumes multiple contributors, a review process, or shared infrastructure; adopt it only once that context exists.
- **Both** — applies regardless of team size, just at different intensity.

| Document / Artifact | Applies to | Notes |
| --- | --- | --- |
| `README.md`, `BLUEPRINT.md`, `CONTRIBUTING.md`, `AGENTS.md` | Both | Entry points and AI operating rules are useful at any scale. |
| `docs/standards/CODE_STYLE.md` | Both | Formatter/linter enforcement benefits a solo project as much as a team — it removes decisions, not just review friction. |
| `docs/standards/PROJECT_STRUCTURE.md` | Both | A predictable layout helps future-you and AI agents even with one contributor. |
| `docs/standards/GITHUB_STANDARDS.md` — Issue/PR/Review sections | Team | The multi-step review discipline assumes a second person reading the PR. Solo, keep the issue template as a personal scratchpad; the "Review Standards" section does not apply until a reviewer exists. |
| `docs/standards/GITHUB_STANDARDS.md` — Collaboration Language | Both | Even solo, deciding your own commit/issue language once avoids inconsistency later. |
| `docs/standards/GITHUB_STANDARDS.md` — Labels (`type/priority/status/area`) | Both (lighter for Personal) | A solo project can usually operate with `type:*` alone; `priority:`/`status:`/`area:` exist to coordinate across people and can be skipped until multiple contributors are involved. |
| `docs/standards/CICD_STANDARDS.md` | Both | CI validates the repo either way; Layer 3/4 (delivery/production gating) in `BLUEPRINT.md`'s CI/CD Model is Team/Scaled territory. |
| `docs/standards/RELEASE_AND_SYNC.md` | Team | This document assumes multiple downstream repositories pinning to boilerplate tags. A solo user maintaining one project does not need tag discipline for their own repo — only relevant if you maintain several projects off this boilerplate or distribute it to others. |
| `.github/CODEOWNERS` | Both (trivial for Personal) | A solo repo's `* @you` is a no-op for review gating but is still useful as an explicit ownership record; delete or leave as-is, low cost either way. |
| `docs/policy/SECURITY_POLICY.md`, `SECURITY.md` | Both (lighter for Personal) | Secret hygiene and dependency review matter solo too. The "private security reporting channel" language in `SECURITY.md` assumes external reporters — for a personal project with no other users, this can be a one-line "contact me directly" instead. |
| `docs/policy/CLOUD_POLICY.md` | Team/Scaled | Multi-account governance, organizational tooling fit, and "team-capability-aware" tradeoffs assume an organization. A personal project on a single cloud account only needs the workload-fit reasoning, not the governance framing. |
| `docs/policy/LICENSE_POLICY.md` | Both | Every repository should declare a license regardless of size. |
| `templates/*` (stack starters) | Both | Directly reusable regardless of team size. |
| `templates/_shared/ci/downstream-ci.yml` | Both | Useful even solo — CI does not require a second contributor to add value. |
| `scripts/sync-boilerplate.{sh,ps1}`, `scripts/sync-labels.sh` | Team (useful, not required, for Personal) | These exist to keep multiple downstream repositories in sync with this boilerplate. A solo user with one project can just copy files once and skip re-syncing. |
| `docs/guides/INIT_GUIDE.md` | Both | The stack-detection and bootstrap checklist applies regardless of team size; the collaboration-language step still matters solo (it is still a decision, just one you make alone). |

## ⏭️ What a Personal Project Can Skip Entirely

- `docs/standards/RELEASE_AND_SYNC.md`'s tag-pinning workflow (unless you maintain multiple repositories from this boilerplate).
- The `priority:`/`status:`/`area:` label axes (keep `type:` only).
- `docs/standards/GITHUB_STANDARDS.md`'s Review Standards section (no reviewer to apply it).
- Multi-account/organizational framing in `docs/policy/CLOUD_POLICY.md`.

## 🔒 What a Personal Project Should Still Keep

- Style/lint/format enforcement (`docs/standards/CODE_STYLE.md`, the `templates/*` tool configs) — it pays off even without a reviewer.
- `docs/policy/SECURITY_POLICY.md`'s baseline (secrets hygiene, dependency review) — scaled down, not skipped.
- `docs/policy/LICENSE_POLICY.md` — every repo needs a declared license.
- `templates/_shared/ci/downstream-ci.yml` — CI catches regressions with zero reviewers.

## 🔁 Maintenance Rule

When a new document or policy is added to this boilerplate, add a row here classifying it as Personal, Team, or Both, so this audit does not go stale.
