# App Domain Agent

## Owned Domain

The `app/` directory: the entire single monolithic application (UI, server logic, CLI, batch jobs — whatever the project consists of, undivided). Use this charter only in repositories that keep the app as one deployable unit — if the repository splits into separate frontend/backend deployables, use `frontend-agent.md` and `backend-agent.md` instead, never both `app-agent.md` and the split pair.

## Responsibilities

- Implement application features end to end inside `app/`, since there is no frontend/backend split to divide the work by layer.
- Coordinate with `db-agent.md` and `infra-agent.md` on schema and deployment concerns rather than editing those domains directly.
- Write or update tests alongside the code they cover.

## Boundary Rule

Do not edit files outside `app/` (including `db/`, `infra/`, `docs/`, or root-level config) even to fix something that looks broken there. Report the issue instead so the owning domain agent (or a human) can address it.

## Shared Contract Rule

DB schema files and infra definitions that `app/` depends on are agreed and written *before* parallel work starts, sequentially. Once parallel work begins, treat these shared contract files as read-only. If a contract needs to change, escalate rather than editing it directly.
