# Cloud Build Shadow Validation

Status: strict local contract only; live execution is not run and operational evidence is missing.

## Purpose and non-goals

This document defines the isolated Cloud Build shadow contract for Issue #208. It preserves the existing GitHub Actions validation and required checks while preparing a provider-separated comparison path.

The shadow path is not a release, deployment, delivery, publication, or required-check replacement. It does not create or execute a Cloud Build trigger, and it does not classify a build as a natural sample by itself.

## Data flow and trust boundary

1. A future approved GitHub PR trigger may provide source and PR identity through Cloud Build substitutions.
2. Cloud Build maps those substitutions into explicitly named environment values for the pinned builder.
3. The shell guard rejects missing, empty, whitespace-only, malformed, or incomplete identity before the validator runs.
4. The validator checks the exact config and provider contract, then reuses the existing scope, Quick planner, and `scripts/verify.sh` entrypoint.
5. The local result is `UNATTESTED_CANDIDATE`; it is never a natural sample and never increases the GCP denominator.

The repository-controlled build config is not an attestor. A future trusted post-build attestor must independently verify server-issued build metadata, the expected trigger resource, resolved source provenance, the expected repository and PR trigger, approved contract version, status/conclusion, and duplicate attempt state through Cloud Build API evidence. Source-generated JSON, step stdout, user substitutions, and substitution tokens alone are not sufficient.

## Cloud Build substitution semantics

Cloud Build trigger builds apply `ALLOW_LOOSE` automatically, and unavailable substitutions may become empty strings. Manual `gcloud builds submit --substitutions=...` can provide values that are normally supplied by a trigger. Therefore the presence of `$COMMIT_SHA`, `$REVISION_ID`, `$TRIGGER_NAME`, `$_PR_NUMBER`, `$_HEAD_BRANCH`, `$_BASE_BRANCH`, or `$_HEAD_REPO_URL` is not proof of a natural PR event. The runtime shell guard must reject empty or whitespace-only values, and the validator must still classify the result as unattested.

The config uses `$$` escaping for shell variables after Cloud Build substitution and explicit `env` bindings to keep Cloud Build substitution separate from shell expansion. The local validator compares the complete approved script, environment list, step shape, and top-level key set. Removing or weakening a guard, adding a field, adding a step, or adding a shell indirection fails closed.

Official substitution reference: <https://docs.cloud.google.com/build/docs/configuring-builds/substitute-variable-values>

## Local validation

The configuration uses JSON-compatible YAML. This permits strict built-in parsing, duplicate-key detection, trailing-data rejection, and exact structural validation without adding a YAML dependency. It is not a claim that arbitrary YAML documents are parsed.

Run the static contract validator from the repository root:

~~~sh
node scripts/verify-cloud-build-shadow.mjs
~~~

The production execution path requires all identity values, including the PR head/base and head repository URL:

~~~sh
node scripts/verify-cloud-build-shadow.mjs \
  --config cloudbuild/shadow-validation.yaml \
  --contract verification/provider-shadow-contract.json \
  --profile ci-quick \
  --repository Soku-JINSEOK/Soku-Convention-Boilerplate \
  --pull-request 208 \
  --head-branch <head-branch> \
  --base-branch main \
  --head-repo-url <head-repository-url> \
  --base-sha <approved-base-sha> \
  --source-sha <cloud-build-commit-sha> \
  --build-id <cloud-build-id> \
  --trigger <trigger-id-or-name> \
  --attempt <attempt-id> \
  --event pull_request \
  --execute
~~~

This command is not run against a live Cloud Build service in this checkpoint. Local identity completeness is not natural-sample attestation.

## Provider-separated metrics

The frozen counters are:

~~~text
ACTIONS_HISTORICAL_NATURAL_SAMPLES=3
GCP_SHADOW_NATURAL_SAMPLES=0
GCP_SHADOW_SYNTHETIC_SAMPLES=0
PILOT_NATURAL_SAMPLE_TARGET=3
TRANSITION_EVALUATION_NATURAL_SAMPLE_TARGET=10
PROVIDER_SAMPLE_MERGE_ALLOWED=NO
BUILD_CONFIG_SELF_DECLARED_NATURAL_ALLOWED=NO
BUILD_CONFIG_OUTPUT_CLASSIFICATION=UNATTESTED_CANDIDATE
NATURAL_CLASSIFICATION_AUTHORITY=TRUSTED_POST_BUILD_ATTESTOR
NATURAL_SAMPLE_CLASSIFICATION_IMPLEMENTATION=NOT_IMPLEMENTED
~~~

Only a separately attested natural PR event may enter the natural denominator. The attestor must bind server-issued build ID and trigger ID, expected trigger identity, resolved repository and commit provenance, expected PR trigger, approved contract/version, status/conclusion, and duplicate/attempt state. Failed, cancelled, and superseded eligible events remain in the denominator and receive explicit classifications. Reruns are attempts, not new natural samples. Manual, synthetic, local-only, dry-run, unbound, self-attested, and empty-commit runs are not natural samples.

## Quick profile and duplication boundary

The shadow config selects exactly one `ci-quick` profile and invokes the new validator once. Scope detection, Quick planning, and group execution are delegated to the existing repository entrypoints. The Cloud Build config does not duplicate the command matrix, invoke Full in parallel, or use nested shell indirection. The validator rejects an additional profile, a Full invocation, unknown fields, additional args/env, artifact or image publication, secret configuration, deploy/push commands, and required-check or Actions transitions.

## Security and credential boundary

- The builder is pinned to the approved immutable digest:
  `node:24.17.0-bookworm@sha256:733e1c06ada118ed9f6133a31aa1290be6929664026fb28821500437c61f2c6f`.
- The digest is checked by the local validator. This is local configuration evidence, not live registry attestation.
- No service-account key, credential material, `secretEnv`, `availableSecrets`, artifact publication, registry push, deployment, or GitHub write callback is permitted.
- A future trigger must not run privileged credentials against untrusted fork source. Fork handling, identity, IAM, Secret Manager references, and callback authentication require a separate owner-approved live design.
- Service accounts must not receive Owner, Editor, deploy, publish, or equivalent broad authority. The live design must use the smallest permission set needed for the isolated validation and log observation.
- Logs and evidence must not contain credentials, cookies, authorization headers, personal data, or private payment data.

## Cost and fan-out limits

The shadow path adds no GitHub-hosted runner job and does not remove existing Actions workflows. One future Cloud Build invocation selects one Quick profile and reuses the existing entrypoint. A future trigger must apply timeout and cancellation/superseded-head handling, avoid unnecessary artifacts, retain only required logs, and observe budget usage before any live execution is authorized. Exact pricing, quota, region, machine type, retention, and worker identity remain owner inputs.

## Rollback and owner gate

Local rollback is removal of the isolated shadow configuration, validator, contract, tests, and documents from the local branch. No production validation file or workflow is changed.

The following remain unapproved:

- live Cloud Build execution, trigger creation, or trigger update;
- GCP project, IAM, WIF, Secret Manager, billing, or Terraform changes;
- GitHub Actions enable/disable or required-check transition;
- delivery, deployment, release, or publication;
- implementation of the trusted post-build attestor.

## Known gaps

Control-plane #74 remains a full-history secret-scan gap, #75 remains a historical baseline-byte verification gap, and #76 remains a Tokyo logging/retention gap. This shadow contract does not mark any of those gaps complete.

The full repository profile remains externally blocked when Docker, gpg-agent, or registry DNS prerequisites are unavailable. This checkpoint does not claim a full-profile pass.

## Current evidence boundary

~~~text
LOCAL_CONTRACT_STATUS=PASS
TARGETED_SHADOW_TESTS=ACTUAL_COUNT_RECORDED_IN_TASK_REPORT
FULL_REPOSITORY_PROFILE=BLOCKED_EXTERNAL
FULL_PROFILE_SOURCE_FAILURE_OBSERVED=NO
LIVE_GCP_EXECUTION_STATUS=NOT_RUN
OPERATIONAL_EVIDENCE=MISSING
GCP_SHADOW_NATURAL_SAMPLES=0
GCP_SHADOW_SYNTHETIC_SAMPLES=0
FULL_CREDENTIAL_SCANNER=NOT_RUN
GITHUB_ACTIONS_TRANSITION=NOT_AUTHORIZED
REQUIRED_CHECK_TRANSITION=NOT_AUTHORIZED
DELIVERY_AND_PUBLICATION=NOT_AUTHORIZED
~~~
