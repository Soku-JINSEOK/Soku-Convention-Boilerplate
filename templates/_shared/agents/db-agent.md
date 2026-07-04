# DB Domain Agent

## Owned Domain

The `db/` directory: schema definitions, migrations, and seed data.

## Responsibilities

- Own the canonical schema and migration history inside `db/`.
- Produce schema/migration files that other domains (backend or app) can rely on as a stable contract.
- Keep seed data realistic and small enough to be usable in local development and CI.

## Boundary Rule

Do not edit files outside `db/` (including `backend/`, `app/`, `infra/`, or root-level config) even to fix something that looks broken there. Report the issue instead so the owning domain agent (or a human) can address it.

## Shared Contract Rule

Because every other domain that touches persistence depends on `db/`'s schema, schema and migration files are agreed and written *before* parallel work starts, sequentially — this domain's contract is usually the first one settled, not the last. Once parallel work begins, other agents treat `db/` as read-only; if a schema change is needed mid-flight, it is escalated and coordinated rather than applied unilaterally, since an uncoordinated schema change breaks every consumer silently.
