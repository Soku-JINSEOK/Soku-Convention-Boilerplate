# Cloud Build Shadow Validation

Status: local contract implementation only.

## Purpose

This document defines an isolated Cloud Build shadow validator for Issue #208. It preserves the existing GitHub Actions validation and required checks while preparing a provider-separated comparison path.

The shadow path is not a release pipeline, deployment pipeline, delivery mechanism, or required-check replacement. This checkpoint does not create or execute a Cloud Build trigger.

## Data flow

1. An approved future GitHub PR trigger supplies the source commit through the Cloud Build built-in COMMIT_SHA substitution and supplies an explicitly bound base SHA through the approved base-SHA trigger input.
2. The pinned Node builder invokes scripts/verify-cloud-build-shadow.mjs once with provider, repository, PR, source, base, build, trigger, attempt, and event identity.
3. The validator checks the contract, plans the affected Quick groups with the existing scope/planner scripts, and calls the existing scripts/verify.sh entrypoint with the ci-quick profile.
4. Evidence is classified as gcp-cloud-build-shadow; it is never merged into the historical Actions denominator.

Cloud Build substitution semantics must be checked against the official documentation for the selected trigger type: <https://docs.cloud.google.com/build/docs/configuring-builds/substitute-variable-values?authuser=5&hl=en>

## Local validation

The configuration uses JSON-compatible YAML so the repository can validate its complete structure with the built-in Node JSON parser without adding a YAML dependency. This is full structural validation of this JSON-compatible configuration, not a claim that the validator parses arbitrary YAML documents.

Run the static contract validator from the repository root:

~~~sh
node scripts/verify-cloud-build-shadow.mjs
~~~

The production execution path is explicit and requires all identity values:

~~~sh
node scripts/verify-cloud-build-shadow.mjs \
  --profile ci-quick \
  --repository Soku-JINSEOK/Soku-Convention-Boilerplate \
  --pull-request 208 \
  --base-sha <approved-base-sha> \
  --source-sha <cloud-build-commit-sha> \
  --build-id <cloud-build-id> \
  --trigger <trigger-id-or-name> \
  --attempt <attempt-id> \
  --event pull_request \
  --execute
~~~

The command is not run against a live Cloud Build service in this checkpoint.

## Provider-separated metrics

The frozen counters are:

~~~text
ACTIONS_HISTORICAL_NATURAL_SAMPLES=3
GCP_SHADOW_NATURAL_SAMPLES=0
GCP_SHADOW_SYNTHETIC_SAMPLES=0
PILOT_NATURAL_SAMPLE_TARGET=3
TRANSITION_EVALUATION_NATURAL_SAMPLE_TARGET=10
PROVIDER_SAMPLE_MERGE_ALLOWED=NO
~~~

Only a natural PR event with source SHA, base SHA, repository, PR, build, trigger, attempt, profile, timestamps, conclusion, and evidence reference is eligible. Failed and cancelled natural builds remain in the denominator and receive an explicit classification. Duplicate callbacks count once. A rerun is a new attempt record, not a new natural sample. Manual, synthetic, local-only, dry-run, unbound-SHA, and empty-commit runs are excluded from the natural sample count.

## Quick profile and duplication boundary

The shadow config selects exactly one ci-quick profile. It does not embed the verification command matrix. Scope detection and Quick planning are delegated to the existing repository scripts, and the resulting group IDs are passed to scripts/verify.sh. The validator rejects a second profile, a Full-profile invocation, or direct validation-command duplication in the Cloud Build config.

## Security and credential boundary

- The builder is pinned to the digest already used by the repository's existing Cloud Build validation configuration:
  node:24.17.0-bookworm@sha256:733e1c06ada118ed9f6133a31aa1290be6929664026fb28821500437c61f2c6f.
- The digest is checked by the local fail-closed validator and by the existing repository Cloud Build configuration test. Registry re-fetch or operational provenance is not claimed here.
- The service account is an explicit future trigger input. The contract does not grant or verify deployment privileges.
- Secret values, service-account key JSON, GitHub write callbacks, secretEnv, artifact publication, registry push, and deployment commands are prohibited.
- A future trigger must not run privileged credentials against untrusted fork source. Fork handling, identity, IAM, Secret Manager references, and callback authentication require a separate owner-approved live design.
- Logs and evidence must not contain credentials, cookies, authorization headers, personal data, or private payment data.

## Cost and fan-out limits

The shadow path adds no GitHub-hosted runner job and does not remove the existing Actions workflows. One future Cloud Build invocation selects one Quick profile and invokes the existing entrypoint for its selected groups. A future trigger must apply timeout and cancellation/superseded-head handling, avoid unnecessary artifacts, retain only required logs, and observe budget usage before any live execution is authorized. Exact pricing, quota, region, machine type, retention, and worker identity remain owner inputs.

## Rollback and owner gate

Local rollback is removal of the isolated shadow configuration, validator, contract, tests, and documents from the local branch. No production validation file or workflow is changed.

The following remain unapproved:

- live Cloud Build execution;
- trigger creation or update;
- GCP project, IAM, WIF, Secret Manager, billing, or Terraform changes;
- GitHub Actions enable/disable or required-check transition;
- delivery, deployment, release, or publication.

## Known gaps

Control-plane #74 remains a full-history secret-scan gap, #75 remains a historical baseline-byte verification gap, and #76 remains a Tokyo logging/retention gap. This shadow contract does not mark any of those gaps complete.

## Current evidence boundary

~~~text
LOCAL_CONTRACT_STATUS=PASS
TARGETED_SHADOW_TESTS=18/18_PASS
FULL_REPOSITORY_PROFILE=BLOCKED_EXTERNAL
LIVE_GCP_EXECUTION_STATUS=NOT_RUN
OPERATIONAL_EVIDENCE=MISSING
FULL_CREDENTIAL_SCANNER=NOT_RUN
GITHUB_ACTIONS_TRANSITION=NOT_AUTHORIZED
REQUIRED_CHECK_TRANSITION=NOT_AUTHORIZED
DELIVERY_AND_PUBLICATION=NOT_AUTHORIZED
~~~
