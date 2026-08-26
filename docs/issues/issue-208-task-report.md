# Issue #208 Task Report — Strict Fail-Closed Shadow Correction

## State and authority

~~~text
TASK_ID=BOILERPLATE-208-STRICT-FAIL-CLOSED-LOCAL-CORRECTION-01
CURRENT_GATE=BOILERPLATE-208-LOCAL-REPORT-CORRECTION-01
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
BRANCH_PUBLICATION_STATUS=PUBLISHED
DRAFT_PR_NUMBER=209
DRAFT_PR_STATUS=OPEN_DRAFT
PUBLISHED_HEAD_SHA=401a1a96023e725398a19d0242c53496a9677991
LIVE_GCP_EXECUTION_STATUS=NOT_RUN
~~

Issue #208 remains the approved shared implementation lane. The branch and
Draft PR are already published at the four-commit head recorded above.
Existing Actions sample history remains historical evidence at 3/10. This
correction does not create a GCP sample, reinterpret local verification as
hosted evidence, or implement a natural-sample attestor.

## Independent review finding and correction scope

The independent review of checkpoint `99ae819e5641b1fc7a585aa723c67883eef968f6` was blocked by six findings:

1. build-controlled or user-controlled values could be used to self-declare an otherwise manual result as natural;
2. static validation did not require effective runtime empty-substitution guards;
3. artifact output fields were not rejected;
4. nested shell construction could evade command-pattern checks;
5. unknown fields and duplicate JSON keys were accepted;
6. substitution `ALLOW_LOOSE`, empty values, and manual substitution semantics were not documented precisely.

This correction closes those local contract findings without implementing the trusted post-build attestor. The build config reports only `UNATTESTED_CANDIDATE`; natural classification remains `NOT_IMPLEMENTED` and requires independent Cloud Build API evidence.

## Exact six-file branch diff

The base-to-head branch diff contains exactly these six added files:

1. `cloudbuild/shadow-validation.yaml` — strict single-step config, explicit substitution-to-env mapping, and runtime guards.
2. `scripts/verify-cloud-build-shadow.mjs` — strict structural validator, duplicate-key parser, runtime identity validation, and unattested classification.
3. `scripts/verify-cloud-build-shadow.test.mjs` — existing regression matrix plus strict fail-closed tests.
4. `verification/provider-shadow-contract.json` — attestation, provenance, provider separation, and zero-sample contract.
5. `docs/guides/CLOUD_BUILD_SHADOW.md` — substitution, trust, security, and operational-boundary documentation.
6. `docs/issues/issue-208-task-report.md` — this correction record.

No dependency, package manifest, lockfile, existing validation profile,
`scripts/verify.sh`, GitHub workflow, production Cloud Build config, or release
file is changed. The six files above are additions relative to the frozen base;
no previously tracked base file is modified by the branch diff. No additional
tracked fixture or generated source file is added.

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
HISTORICAL_TARGETED_SHADOW_TESTS=53/53_PASS
HISTORICAL_TARGETED_TEST_HEAD=289b5691b171166050d3de9d3489ac4ee4e301fb
LOCAL_CONTRACT_STATUS=PASS
FULL_REPOSITORY_PROFILE=BLOCKED_EXTERNAL
FULL_PROFILE_SOURCE_FAILURE_OBSERVED=NO
FULL_PROFILE_INDEPENDENT_RERUN=NOT_RUN_ENVIRONMENT_UNAVAILABLE
LIVE_GCP_EXECUTION_STATUS=NOT_RUN
GCP_SHADOW_NATURAL_SAMPLES=0
GCP_SHADOW_SYNTHETIC_SAMPLES=0
OPERATIONAL_EVIDENCE=MISSING
FULL_CREDENTIAL_SCANNER=NOT_RUN
HISTORICAL_REVIEW_RESULT=RESOURCE_BOUND_CORRECTION_REQUIRED
~~

The full repository profile was not rerun because the read-only prerequisite probes found no Docker daemon, no running gpg-agent, and registry DNS unavailable. Previous attempts stopped at environment gates before a repository source failure could be established. This is an external limitation, not a full-profile pass.

The 53/53 result is historical local evidence for the strict-attestation
checkpoint identified above. It includes the original contract cases and
strict cases for self-attestation, runtime empty values, `ALLOW_LOOSE`, guard
removal/weakening, artifacts/images/secrets, nested shell, quoted or split
push, unknown fields, duplicate keys, malformed JSON, provenance, and zero
natural samples. It is not the current four-commit-head result.

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
CURRENT_TARGETED_SHADOW_TESTS=69/69_PASS
CURRENT_LOCAL_EVIDENCE_HEAD=401a1a96023e725398a19d0242c53496a9677991
NATURAL_SAMPLES=0
SYNTHETIC_SAMPLES=0
LIVE_GCP_EXECUTION_STATUS=NOT_RUN
FULL_PROFILE_STATUS=BLOCKED_EXTERNAL
FULL_PROFILE_SOURCE_FAILURE_OBSERVED=NO
FULL_PROFILE_INDEPENDENT_RERUN=NOT_RUN_ENVIRONMENT_UNAVAILABLE
LOCAL_VERIFICATION_ENTRYPOINT=BLOCKED_EXTERNAL_NPM_DNS
FOUR_COMMIT_HEAD_REVIEW_STATUS=REPORT_ACCOUNTING_CORRECTION_REQUIRED
~~

The 69/69 result is current local evidence for the frozen four-commit head. The
full repository profile remains subject to its external Docker, gpg-agent, and
npm registry prerequisites. Hosted readback observes Issue #208 as
Open/`status:in-progress` and PR #209 as Open/Draft at that exact head. The
previous checkpoint remains preserved as its parent.

## Evidence classification at the published four-commit head

~~~text
OWNER_RATIFICATION_EVIDENCE=RECORDED_SEPARATELY
INDEPENDENT_REVIEW_EVIDENCE=REPORT_ACCOUNTING_CORRECTION_REQUIRED
LOCAL_EVIDENCE=69/69_PASS_AT_401a1a96023e725398a19d0242c53496a9677991
HOSTED_VALIDATION_GATE=SUCCESS_AT_PUBLISHED_HEAD
HOSTED_PR_METADATA_GATE=FAILURE_AT_PUBLISHED_HEAD
EXTERNAL_LEGACY_REQUIRED_CONTEXT=SUCCESS_AT_PUBLISHED_HEAD
PROVIDER_BLOCKED_METADATA_SYNC=GITHUB_SECONDARY_RATE_LIMIT
FUTURE_FIFTH_HEAD_HOSTED_EVIDENCE=NOT_YET_AVAILABLE
~~

Owner ratification, independent review, local verification, hosted checks, and
provider limitations are distinct evidence classes. The public Issue records
the initial local implementation authorization; later ratification and review
records are not converted into test or hosted evidence here.

At the published four-commit head, the required Validation Gate and external
legacy context succeeded. The PR Metadata Gate failed because the Draft PR's
Task report and Governance profile values did not use the required Markdown
code formatting. The separate metadata synchronization job was blocked by a
GitHub secondary rate limit; it is provider-blocked evidence and is not a
required status context.

These hosted results apply only to the published four-commit head. They must
not be reused as evidence for the future fifth commit. That commit's SHA is
intentionally absent from this self-referential report and must be recorded as
external post-commit evidence.

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
OWNER_POLICY_SET_A_TO_J=FOLLOW_ON_LIVE_SHADOW_INPUT_ONLY
~~

Control-plane #74 remains a full-history secret-scan gap, #75 remains a
historical baseline-byte verification gap, and #76 remains a Tokyo
logging/retention gap. Native release execution, signing, IAM, Secret Manager,
billing, Terraform, trigger execution, required-check transition, delivery,
and publication remain outside this checkpoint.

The Owner policy set A–J is preserved only as input to a later, separately
authorized live-shadow implementation. This public report does not reproduce
private cloud-project identifiers, personal budget information, or private
Project identifiers.

## Published history accounting

~~~text
BRANCH_DIFF_ADDED_FILES=6
SOURCE_IMPLEMENTATION_COMMIT_COUNT=THREE_LINEAR_NON_MERGE_COMMITS
SOURCE_IMPLEMENTATION_COMMIT_CHAIN=99ae819e5641b1fc7a585aa723c67883eef968f6 -> 289b5691b171166050d3de9d3489ac4ee4e301fb -> cff7d7b3cdb4a8e96e4f3a8c7ce1b33ae58a13dd
EXISTING_REPORT_ACCOUNTING_COMMIT_COUNT=ONE_DIRECT_NON_MERGE_CHILD
PUBLISHED_LINEAR_NON_MERGE_COMMIT_COUNT=FOUR
PUBLISHED_HEAD=401a1a96023e725398a19d0242c53496a9677991
PREVIOUS_CHECKPOINT_PRESERVED=YES
BRANCH_PUBLICATION_STATUS=PUBLISHED
DRAFT_PR_STATUS=OPEN_DRAFT
HISTORICAL_GITHUB_API_OR_MUTATION_COUNT=NOT_INFERRED
~~

The three source implementation commits are recorded in
SOURCE_IMPLEMENTATION_COMMIT_CHAIN above. The fourth commit already exists as
one direct non-merge child of
`cff7d7b3cdb4a8e96e4f3a8c7ce1b33ae58a13dd`, and the branch and Draft PR are
published at that four-commit head. Historical GitHub API or mutation counts
are not estimated from current state.

The next report-only correction, if separately committed, will be the fifth
linear non-merge commit. Its final SHA belongs in external post-commit evidence
and is not self-embedded in this report.

## AI assistance

Planning, contract design, implementation, test drafting, and verification
reporting were assisted by OpenAI Codex. No credential, secret, service-account
key, or live GCP operation is recorded. Existing branch and Draft PR
publication are reported as observable GitHub state; historical API-call
counts are not inferred.
