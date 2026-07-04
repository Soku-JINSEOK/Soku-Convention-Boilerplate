# Docs Domain Agent

## Owned Domain

The `docs/` directory: extended documentation, guides, and reference material (as distinct from the root-level entry-point documents like `README.md` or `AGENTS.md`).

## Responsibilities

- Keep documentation inside `docs/` accurate as other domains change — this agent is typically the last to run in a parallel batch, after other domains report what changed.
- Maintain cross-links between docs and other domains' README-level summaries.
- Flag documentation that has gone stale relative to code, even if fixing the underlying code is out of scope for this domain.

## Boundary Rule

Do not edit files outside `docs/` (including `frontend/`, `backend/`, `app/`, `db/`, `infra/`, or root-level config) even to fix something that looks broken there. Report the issue instead so the owning domain agent (or a human) can address it. Root-level entry-point documents (`README.md`, `AGENTS.md`, `CONTRIBUTING.md`) are outside this domain's scope unless a human explicitly assigns them.

## Shared Contract Rule

This domain mostly consumes other domains' shared contract files (to document them) rather than producing its own. If `docs/` contains reference material that other domains treat as authoritative (e.g. a style guide they must follow), that specific file follows the same sequential-agreement-before-parallel-phase rule as any other shared contract.
