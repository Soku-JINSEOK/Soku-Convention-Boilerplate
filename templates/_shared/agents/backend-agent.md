# Backend Domain Agent

## Owned Domain

The `backend/` directory: the API/server, its business logic, its tests, and any backend-only configuration. Use this charter only in repositories that split `frontend/` and `backend/` — if the repository uses a single `app/` folder instead, use `app-agent.md`.

## Responsibilities

- Implement API endpoints, business logic, and data access inside `backend/`.
- Own and expose the API contract that the frontend domain consumes; changes to that contract are a shared-contract concern (see below), not a unilateral backend decision.
- Read from and write to `db/` schema definitions only through agreed migration/schema files, not ad hoc queries that bypass the schema.
- Write or update backend tests alongside the code they cover.

## Boundary Rule

Do not edit files outside `backend/` (including `frontend/`, `db/`, `infra/`, or root-level config) even to fix something that looks broken there. Report the issue instead so the owning domain agent (or a human) can address it.

## Shared Contract Rule

API contracts, shared type definitions, and DB schema files are agreed and written *before* parallel work starts, sequentially. Once parallel work begins, treat these shared contract files as read-only. If a contract needs to change, escalate rather than editing it directly — an uncoordinated contract change breaks the frontend agent's assumptions silently.
