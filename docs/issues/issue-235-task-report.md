# Issue #235 task report — Archify repository flow maps

## Authority and scope

Issue #235 tracks the documentation-only Archify map publication for
Soku-Convention-Boilerplate. The work is limited to evidence-linked
architecture, workflow, and sequence maps under `docs/architecture/archify/`.
It does not authorize application, release, deployment, cloud, IAM, billing,
or publication changes.

## Implementation

- Preserve the three Archify JSON maps and regenerate their standalone HTML
  views with the official Archify v2.16.0 CLI.
- Pin Architecture evidence to merged source revision
  `0c51a1bf3fee0d704c95c7720b2e6b3d6ec26315`.
- Keep this PR limited to six map files and this task report. The runner and
  JavaScript dependency changes were merged separately in PR #237.

## Verification and disposition

- Archify v2.16.0 `doctor` passed.
- Architecture, workflow, and sequence maps each passed official showcase
  validation and atomic `deliver`, with 9/9 artifact checks and no warnings.
- Architecture evidence verified five source references against the pinned
  repository revision.
- PR #237 passed both high-severity npm audits, OSV, CodeQL, Cloud Build, and
  manual Hosted Full on its exact head before it was merged. The current
  documentation head still requires its own PR policy and hosted checks.

## AI assistance

- Provider: OpenAI
- Model: Codex (exact backend model identifier not exposed)
- Usage Summary: repository inspection, Archify map revision refresh,
  showcase validation, and task-report correction.
