# Issue #208 Task Report — GCP Shadow Validation Contract

## State and authority

~~~text
TASK_ID=BOILERPLATE-208-OPTION-B-ACTIVATION-AND-LOCAL-IMPLEMENTATION-01
ISSUE_STATE=OPEN
ISSUE_STATUS=IN_PROGRESS
PROJECT_STATUS=IN_PROGRESS
BASE_SHA=76c739557d3919eb965d4de2792df1ee1ed2665f
BRANCH=agent/issue-208-gcp-shadow-validation
SHARED_WIP=1
PRODUCT_WIP=1
GLOBAL_WIP=2
ACTIVE_PRODUCT_LANE=CutVi #111
ACTIONS_HISTORICAL_NATURAL_SAMPLES=3
GCP_SHADOW_NATURAL_SAMPLES=0
GCP_SHADOW_SYNTHETIC_SAMPLES=0
PILOT_NATURAL_SAMPLE_TARGET=3
TRANSITION_EVALUATION_NATURAL_SAMPLE_TARGET=10
PROVIDER_SAMPLE_MERGE_ALLOWED=NO
~~~

Issue #208 was activated as the approved shared implementation lane. Existing Actions sample history remains historical evidence at 3/10. The local implementation does not create a GCP sample and does not reinterpret local verification as hosted evidence.

## Goal and acceptance boundary

Issue #208 is the provider-separated Cloud Build shadow successor for the Boilerplate validation experiment: <https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/208>

Acceptance is limited to:

- preserving the three historical Actions natural samples;
- keeping GCP natural and synthetic denominators separate and at zero in this checkpoint;
- binding a future natural sample to its repository, PR event, exact source SHA, base SHA, build, trigger, attempt, profile, timestamps, conclusion, and evidence reference;
- reusing the existing local verification entrypoint without duplicating its command matrix;
- rejecting mutable builders, deployment/publication, credential material, sample merging, duplicate profiles, and required-check or Actions transitions;
- requiring separate owner approval for any live Cloud Build execution.

## Exact six-file allowlist

Only these files are authorized on this local branch:

1. cloudbuild/shadow-validation.yaml — create the isolated JSON-compatible YAML Cloud Build contract.
2. scripts/verify-cloud-build-shadow.mjs — create the fail-closed config, provider, identity, and shared-entrypoint validator.
3. scripts/verify-cloud-build-shadow.test.mjs — create contract regression tests.
4. verification/provider-shadow-contract.json — create the provider-separated sample and security contract.
5. docs/guides/CLOUD_BUILD_SHADOW.md — create the shadow boundary and operating guide.
6. docs/issues/issue-208-task-report.md — create this task report.

No dependency, package manifest, lockfile, existing validation profile, scripts/verify.sh, GitHub workflow, production Cloud Build config, or release file is modified.

## Implementation contract

The Cloud Build config selects exactly one ci-quick profile and invokes the new validator once. The validator delegates scope detection, Quick planning, and group execution to the existing repository entrypoints. It rejects mutable builder references, missing or unbound source SHA values, provider/sample mixing, synthetic-natural reclassification, duplicated Quick/Full invocation, deployment or registry publication commands, inline secret material, missing timeout, required-check transition declarations, and existing Actions disablement.

Cloud Build built-in substitution semantics are documented by Google at <https://docs.cloud.google.com/build/docs/configuring-builds/substitute-variable-values?authuser=5&hl=en>. The local contract requires COMMIT_SHA for the exact source and the approved base-SHA trigger input; no implicit or guessed base is accepted.

The builder identity is copied from the existing repository configuration and validated as an immutable digest:

~~~text
node:24.17.0-bookworm@sha256:733e1c06ada118ed9f6133a31aa1290be6929664026fb28821500437c61f2c6f
~~~

This is local configuration evidence only. It is not a live registry attestation or hosted execution result.

## Verification status

~~~text
LOCAL_CONTRACT_STATUS=PASS
TARGETED_SHADOW_TESTS=18/18_PASS
FULL_REPOSITORY_PROFILE=BLOCKED_EXTERNAL
FULL_REPOSITORY_PROFILE_EXIT=51
FULL_REPOSITORY_PROFILE_BLOCKER=Docker daemon unavailable at gcloud template container step
FORMAT_MARKDOWN_YAML_JSON_JS=PASS
LIVE_GCP_EXECUTION_STATUS=NOT_RUN
OPERATIONAL_EVIDENCE=MISSING
FULL_CREDENTIAL_SCANNER=NOT_RUN
INDEPENDENT_REVIEW_STATUS=PENDING
~~~

The targeted shadow contract and regression tests passed. The canonical full entrypoint executed its available repository suites and stopped at the gcloud template container step because no Docker daemon was available; this is an environment limitation, not a Cloud Build or source pass. It must not be claimed as a full-profile pass. The implementation must not claim Cloud Build verified, hosted validation passed, migration complete, required-check replacement ready, or Actions replacement complete.

## Reusable evidence and remaining gaps

Natural sample eligibility, failed/cancelled denominator treatment, duplicate callback handling, rerun attempt handling, source SHA binding, build identity, trigger identity, and provider-specific evidence are encoded in verification/provider-shadow-contract.json.

Control-plane #74, #75, and #76 remain open gaps. This local contract does not close them. Native macOS/Windows release execution, signing, delivery, deployment, publication, IAM, Secret Manager, billing, Terraform, trigger execution, and required-check transition are outside this task.

## Mutation accounting

~~~text
AUTHORIZED_ISSUE_STATUS_TRANSITION_COUNT=1
AUTHORIZED_ISSUE_208_COMMENT_COUNT=1
AUTHORIZED_ISSUE_87_COMMENT_UPDATE_COUNT=1
AUTHORIZED_LOCAL_BRANCH_CREATE_COUNT=1
AUTHORIZED_LOCAL_COMMIT_COUNT=PENDING
DIRECT_AGENT_PROJECT_MUTATION_COUNT=0
AUTOMATIC_PROJECT_SYNC_MUTATION_COUNT=1
WORKFLOW_RERUN_COUNT=0
WORKFLOW_DISPATCH_COUNT=0
PUSH_COUNT=0
LIVE_GCP_BUILD_COUNT=0
CLOUD_BUILD_TRIGGER_MUTATION_COUNT=0
GCP_IAM_MUTATION_COUNT=0
BILLING_MUTATION_COUNT=0
READY_FOR_REVIEW_COUNT=0
MERGE_COUNT=0
DELIVERY_COUNT=0
PUBLICATION_COUNT=0
~~~

## AI Assistance

Planning, contract design, implementation, test drafting, and verification reporting were assisted by OpenAI Codex. No credential, secret, service-account key, or live GCP operation was used or recorded.
