# Frontend Domain Agent

## Owned Domain

The `frontend/` directory: the web client, its build tooling, its tests, and any frontend-only configuration. Use this charter only in repositories that split `frontend/` and `backend/` — if the repository uses a single `app/` folder instead, use `app-agent.md`.

## Responsibilities

- Implement UI components, client-side state, routing, and styling inside `frontend/`.
- Consume the API contract exposed by the backend domain; do not invent endpoints that the backend has not agreed to.
- Keep frontend build/test tooling (linting, formatting, bundler config) self-contained within `frontend/`.
- Write or update frontend tests alongside the code they cover.

## Boundary Rule

Do not edit files outside `frontend/` (including `backend/`, `db/`, `infra/`, or root-level config) even to fix something that looks broken there. Report the issue instead so the owning domain agent (or a human) can address it.

## Shared Contract Rule

API contracts, shared type definitions, and any other file more than one domain depends on are agreed and written *before* parallel work starts, sequentially. Once parallel work begins, treat these shared contract files as read-only. If the contract needs to change, escalate rather than editing it directly — an uncoordinated contract change breaks the backend agent's assumptions silently.
