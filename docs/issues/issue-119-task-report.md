# Issue #119 Task Report — Promote verified image digests

## Goal and Background

Issue [#119](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/119)
separates image publication from deployment so Cloud Run receives only an image
already built and verified by canonical `main` Validation.

## Implementation Status

- Terraform defines a dedicated CI builder constrained by immutable repository
  and owner IDs, `refs/heads/main`, and the exact Validation workflow. It has
  only repository-scoped Artifact Registry writer access.
- `gcp-bootstrap.sh --ci-builder-only` applies that identity and two repository
  variables without building or deploying.
- Canonical main Validation builds `linux/amd64`, checks `/health`, pushes the
  commit tag, resolves the registry digest, and uploads the schema-v1 manifest.
- Deployment accepts `source_run_id`, verifies canonical successful run
  metadata and every manifest binding, then creates a digest-only plan.
- `cd-plan.sh` no longer builds, tests, installs dependencies, pushes images,
  or runs Terraform. Rollback-only behavior remains supported.

## Remaining Operational Gates

- Apply the CI builder identity and variables.
- Produce a successful main artifact and deploy it to `dev`.
- Verify health, traffic, evidence, and rollback.
- Only after that success, set `grant_deployer_artifact_writer=false`, apply,
  and repeat deployment verification.

## Verification

- Promotion mismatch and digest-only static tests passed.
- Deployment health, traffic, rollback, and evidence tests passed.
- Terraform fmt, init, validate, and mock tests passed in the pinned container.
- actionlint and supply-chain verification passed.

## Public Disclosure Review

- [x] No credentials or private identifiers
- [x] No cloud account identifiers or service endpoints
- [x] No billing or personal information

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex (GPT-5)
