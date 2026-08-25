# Issue #208 Task Report — Strict Fail-Closed Shadow Correction

## State and authority

~~~text
TASK_ID=BOILERPLATE-208-STRICT-FAIL-CLOSED-LOCAL-CORRECTION-01
ISSUE_STATE=OPEN
ISSUE_STATUS=IN_PROGRESS
PROJECT_STATUS=IN_PROGRESS
ORIGINAL_BASE_SHA=76c739557d3919eb965d4de2792df1ee1ed2665f
PREVIOUS_CHECKPOINT=99ae819e5641b1fc7a585aa723c67883eef968f6
PREVIOUS_CHECKPOINT_REVIEW=BLOCKED
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
PUSH_STATUS=NOT_PUSHED
LIVE_GCP_EXECUTION_STATUS=NOT_RUN
~~

Issue #208 remains the approved shared implementation lane. Existing Actions sample history remains historical evidence at 3/10. This correction does not create a GCP sample, reinterpret local verification as hosted evidence, or implement a natural-sample attestor.

## Independent review finding and correction scope

The independent review of checkpoint `99ae819e5641b1fc7a585aa723c67883eef968f6` was blocked by six findings:

1. build-controlled or user-controlled values could be used to self-declare an otherwise manual result as natural;
2. static validation did not require effective runtime empty-substitution guards;
3. artifact output fields were not rejected;
4. nested shell construction could evade command-pattern checks;
5. unknown fields and duplicate JSON keys were accepted;
6. substitution `ALLOW_LOOSE`, empty values, and manual substitution semantics were not documented precisely.

This correction closes those local contract findings without implementing the trusted post-build attestor. The build config reports only `UNATTESTED_CANDIDATE`; natural classification remains `NOT_IMPLEMENTED` and requires independent Cloud Build API evidence.

## Exact six-file correction allowlist

Only these existing files are modified by this correction:

1. `cloudbuild/shadow-validation.yaml` — strict single-step config, explicit substitution-to-env mapping, and runtime guards.
2. `scripts/verify-cloud-build-shadow.mjs` — strict structural validator, duplicate-key parser, runtime identity validation, and unattested classification.
3. `scripts/verify-cloud-build-shadow.test.mjs` — existing regression matrix plus strict fail-closed tests.
4. `verification/provider-shadow-contract.json` — attestation, provenance, provider separation, and zero-sample contract.
5. `docs/guides/CLOUD_BUILD_SHADOW.md` — substitution, trust, security, and operational-boundary documentation.
6. `docs/issues/issue-208-task-report.md` — this correction record.

No dependency, package manifest, lockfile, existing validation profile, `scripts/verify.sh`, GitHub workflow, production Cloud Build config, or release file is changed. No new tracked fixture or generated source file is added.

## Strict local contract

The JSON-compatible Cloud Build config has an exact top-level and step schema, one approved immutable builder, one exact `ci-quick` invocation, one exact environment mapping, and no artifact, image, secret, deployment, publication, or callback fields. Any unknown key, additional step, additional arg/env entry, duplicate JSON key, malformed JSON, trailing data, nested shell, or altered guard fails closed.

The shell guard rejects empty, whitespace-only, malformed, or missing commit SHA, base SHA, PR number, head branch, base branch, head repository URL, build ID, and trigger values. `$$` escaping separates Cloud Build substitution from shell expansion. These checks establish input completeness only; they do not attest that a build was trigger-created or natural.

The provider contract requires a future trusted post-build attestor to verify server-issued build ID and trigger ID, expected trigger resource identity, resolved source provenance, expected repository and PR trigger, approved contract/version, status/conclusion, and duplicate attempt state. `$COMMIT_SHA`, `$REVISION_ID`, `$TRIGGER_NAME`, PR substitutions, user-defined substitutions, stdout, and source-generated results are insufficient alone. Manual and synthetic builds, unverified candidates, and local-only runs cannot enter the natural denominator.

## Verification status

~~~text
STRICT_FAIL_CLOSED_CORRECTION=PASS
BUILD_SELF_ATTESTED_NATURAL_CLASSIFICATION_ALLOWED=NO
RUNTIME_EMPTY_GUARD_STATUS=PASS
STRICT_STRUCTURE_STATUS=PASS
ARTIFACT_AND_PUBLISH_REJECTION=PASS
DUPLICATE_KEY_DETECTION=PASS
NESTED_SHELL_BYPASS_STATUS=PASS
TARGETED_SHADOW_TESTS=53/53_PASS
LOCAL_CONTRACT_STATUS=PASS
FULL_REPOSITORY_PROFILE=BLOCKED_EXTERNAL
FULL_PROFILE_SOURCE_FAILURE_OBSERVED=NO
FULL_PROFILE_INDEPENDENT_RERUN=NOT_RUN_ENVIRONMENT_UNAVAILABLE
LIVE_GCP_EXECUTION_STATUS=NOT_RUN
GCP_SHADOW_NATURAL_SAMPLES=0
GCP_SHADOW_SYNTHETIC_SAMPLES=0
OPERATIONAL_EVIDENCE=MISSING
FULL_CREDENTIAL_SCANNER=NOT_RUN
INDEPENDENT_REVIEW_STATUS=PENDING
~~

The full repository profile was not rerun because the read-only prerequisite probes found no Docker daemon, no running gpg-agent, and registry DNS unavailable. Previous attempts stopped at environment gates before a repository source failure could be established. This is an external limitation, not a full-profile pass.

The targeted test count is the actual Node test-runner count after correction. It includes the original contract cases and strict cases for self-attestation, runtime empty values, `ALLOW_LOOSE`, guard removal/weakening, artifacts/images/secrets, nested shell, quoted or split push, unknown fields, duplicate keys, malformed JSON, provenance, and zero natural samples.

## JSON resource-bound correction

The independent review of checkpoint `289b5691b171166050d3de9d3489ac4ee4e301fb`
reproduced a 5,000-level nested JSON input that caused
`RangeError: Maximum call stack size exceeded`. This was a source correction
requirement; a top-level exception catch was not used as the remediation.

The narrow correction is limited to these five existing files:

1. `scripts/verify-cloud-build-shadow.mjs`
2. `scripts/verify-cloud-build-shadow.test.mjs`
3. `verification/provider-shadow-contract.json`
4. `docs/guides/CLOUD_BUILD_SHADOW.md`
5. `docs/issues/issue-208-task-report.md`

The validator now applies a UTF-8 byte limit of `1048576` and an iterative
nesting limit of `64` before duplicate-key parsing, JSON parsing, strict schema
validation, or semantic validation. File inputs are checked for regular-file
identity and size before reading. Over-limit input fails closed with a concise
deterministic message and does not expose a stack trace or input payload.

~~~text
RESOURCE_BOUND_STATUS=IMPLEMENTED_LOCALLY
MAX_JSON_INPUT_BYTES=1048576
MAX_JSON_NESTING_DEPTH=64
STACK_EXHAUSTION_PREVENTION=PRE_PARSE_ITERATIVE_BOUND
CURRENT_SOURCE_CHANGED_FILES=5
~~

The implementation is pre-parse prevention, not a top-level `RangeError` catch.

Boundary and regression verification recorded for this correction:

~~~text
DEPTH_0_63_64=PASS
DEPTH_65_1000_5000=CONTROLLED_FAIL
MIXED_DEPTH_BOUNDARY=PASS
STRING_BRACKET_HANDLING=PASS
MALFORMED_STRUCTURE_HANDLING=PASS
ONE_MIB_BOUNDARY=PASS
MULTIBYTE_UTF8_BYTE_COUNT=PASS
CONFIG_FILE_PATH=PASS
CONTRACT_FILE_PATH=PASS
CLI_STACK_TRACE_SUPPRESSION=PASS
TARGETED_SHADOW_TESTS=69/69_PASS
NATURAL_SAMPLES=0
SYNTHETIC_SAMPLES=0
LIVE_GCP_EXECUTION_STATUS=NOT_RUN
FULL_PROFILE_STATUS=BLOCKED_EXTERNAL
FULL_PROFILE_SOURCE_FAILURE_OBSERVED=NO
FULL_PROFILE_INDEPENDENT_RERUN=NOT_RUN_ENVIRONMENT_UNAVAILABLE
LOCAL_VERIFICATION_ENTRYPOINT=BLOCKED_EXTERNAL_NPM_DNS
PUSH_STATUS=NOT_PUSHED
INDEPENDENT_REVIEW_STATUS=PENDING
~~

The full repository profile remains subject to its external Docker,
gpg-agent, and npm registry prerequisites. GitHub readback in this task
observed Issue #208 as Open/`status:in-progress`; no GitHub mutation was
performed. The previous checkpoint remains preserved as its parent.

## Provider and release boundary

~~~text
BUILD_CONFIG_SELF_DECLARED_NATURAL_ALLOWED=NO
BUILD_CONFIG_OUTPUT_CLASSIFICATION=UNATTESTED_CANDIDATE
NATURAL_CLASSIFICATION_AUTHORITY=TRUSTED_POST_BUILD_ATTESTOR
NATURAL_SAMPLE_CLASSIFICATION_IMPLEMENTATION=NOT_IMPLEMENTED
ACTIONS_HISTORICAL_NATURAL_SAMPLES=3
GCP_SHADOW_NATURAL_SAMPLES=0
GCP_SHADOW_SYNTHETIC_SAMPLES=0
PROVIDER_SAMPLE_MERGE_ALLOWED=NO
REQUIRED_CHECK_TRANSITION=NOT_AUTHORIZED
GITHUB_ACTIONS_TRANSITION=NOT_AUTHORIZED
DELIVERY=NOT_AUTHORIZED
PUBLICATION=NOT_AUTHORIZED
~~

Control-plane #74 remains a full-history secret-scan gap, #75 remains a historical baseline-byte verification gap, and #76 remains a Tokyo logging/retention gap. Native release execution, signing, IAM, Secret Manager, billing, Terraform, trigger execution, required-check transition, delivery, and publication remain outside this checkpoint.

## Mutation accounting

~~~text
SOURCE_CHANGED_FILES=6
NEW_TRACKED_FILES=6
SOURCE_IMPLEMENTATION_COMMIT_COUNT=THREE_LINEAR_NON_MERGE_COMMITS
SOURCE_IMPLEMENTATION_COMMIT_CHAIN=99ae819e5641b1fc7a585aa723c67883eef968f6 -> 289b5691b171166050d3de9d3489ac4ee4e301fb -> cff7d7b3cdb4a8e96e4f3a8c7ce1b33ae58a13dd
REPORT_ACCOUNTING_CORRECTION_COMMIT_COUNT=ONE_DIRECT_NON_MERGE_CHILD
TOTAL_LOCAL_COMMIT_COUNT_AFTER_REPORT_CORRECTION=FOUR_LINEAR_NON_MERGE_COMMITS
PREVIOUS_CHECKPOINT_PRESERVED=YES
PUSH_COUNT=0
GITHUB_MUTATION_COUNT=0
PROJECT_MUTATION_COUNT=0
WORKFLOW_RERUN_COUNT=0
WORKFLOW_DISPATCH_COUNT=0
LIVE_GCP_BUILD_COUNT=0
CLOUD_BUILD_TRIGGER_MUTATION_COUNT=0
GCP_IAM_MUTATION_COUNT=0
BILLING_MUTATION_COUNT=0
READY_FOR_REVIEW_COUNT=0
MERGE_COUNT=0
DELIVERY_COUNT=0
PUBLICATION_COUNT=0
~~

The three source implementation commits are recorded in SOURCE_IMPLEMENTATION_COMMIT_CHAIN above. When committed, this report-only correction becomes the fourth linear local commit and one direct non-merge child of cff7d7b3cdb4a8e96e4f3a8c7ce1b33ae58a13dd; its final SHA is recorded in post-commit evidence rather than self-embedded.

## AI assistance

Planning, contract design, implementation, test drafting, and verification reporting were assisted by OpenAI Codex. No credential, secret, service-account key, live GCP operation, or GitHub mutation was used or recorded.
