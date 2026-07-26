# Issue #136 Task Report — Add validation-only Cloud Build checks

## Goal and Background

Issue [#136](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/136)
adds an opt-in Cloud Build path for GCP-specific validation without moving
deployment authority away from the existing manual GitHub Actions workflow.

## Proposed Approach

Use digest-pinned builders and a dedicated Logs Writer identity for
validation-only PR and `main` triggers. Keep the integration disabled by
default, isolate its Terraform state from Cloud Run, and stop the bootstrap
before Docker or runtime work when validation is enabled.

## Planned Implementation

- Add `cloudbuild/validation.yaml` with Node 24 regression tests, Terraform
  validation, and an amd64 container build.
- Add opt-in Terraform resources for the Cloud Build API, dedicated identity,
  IAM binding, and two first-generation GitHub App triggers.
- Add bootstrap preview/apply support with an isolated validation state prefix.
- Add policy, bootstrap, Terraform mock-plan, and supply-chain regression tests.
- Document activation, external `/gcbrun` approval, evidence collection,
  required-check promotion, and rollback.

## Acceptance Criteria

- Cloud Build validation is disabled by default and creates exactly two
  reviewed triggers when enabled.
- The dedicated identity receives only `roles/logging.logWriter`.
- Builder images are pinned by exact version and digest.
- Validation cannot publish images, access secrets, or deploy Cloud Run.
- Missing first-generation connectivity fails before image work.
- Cloud Run state and deployment behavior remain isolated and unchanged.
- Existing GitHub Actions deployment and rollback regression tests pass.

## Approval

- **Status:** `Approved`
- **Approved by:** `Soku-JINSEOK` (authorized continuation and publication)

## Implementation Status

Implemented locally. GitHub Actions and live Cloud Build trigger evidence remain
pending until the Draft pull request is opened and merged in the documented
order.

## Verification

- `node --test .github/*.test.mjs scripts/contribution-title.test.mjs
  scripts/pull-request-policy.test.mjs scripts/verify.test.mjs
  scripts/verify-supply-chain.test.mjs` (81/81 passing).
- `node --test .github/cloudbuild-validation.test.mjs
  .github/deploy-gcp.test.mjs scripts/verify-supply-chain.test.mjs` (28/28
  passing after final bootstrap and state-isolation changes).
- Terraform 1.15.3 `fmt -check`, `validate`, and `test` (2/2 mock plans
  passing).
- `node scripts/verify-supply-chain.mjs` (passing).
- `npx --yes yaml-lint@1.7.0 cloudbuild/validation.yaml` (passing).
- `npx --yes markdownlint-cli2@0.22.1 --config .markdownlint.jsonc
  docs/guides/CLOUD_RUN_CICD.md infra/gcp/README.md` (0 errors).
- `bash -n scripts/gcp-bootstrap.sh
  verification/commands/repo-hygiene.sh` (passing).
- `git diff --check` (passing).

## Public Disclosure Review

- [x] No credentials, tokens, private keys, or credential-bearing URLs
- [x] No private repository, project, or product names
- [x] No cloud project IDs, account numbers, service URLs, image URIs, or
      revision identifiers
- [x] No personal billing, subscription, budget, or payment-status information
- [x] No personal email, phone, address, or local absolute path
- [x] No private Issue, PR, Project, or control-plane identifiers

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex (GPT-5.6)
