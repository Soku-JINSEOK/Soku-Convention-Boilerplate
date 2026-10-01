# Task report: explainable decisions and CI workload

## Goal and Background

Issue #245 extends PR #244. The owner requested explicit answers to why an
implementation was chosen and fewer unnecessary CI executions.
Baseline source: `8c24bdc6162c67837c54f5947bb973a2e714c15c`.

## Proposed Approach

Extend existing contribution, PR and task-report guidance; add concrete
reasoning beside the structure, language, security, cloud and verification
conventions. Remove the duplicate full tree from manual Validation and cache
the manual runner's npm download store by lockfile.

## Decision Rationale

| Decision | Constraint and alternatives | Reason and accepted cost | Evidence | Owner and revisit trigger |
| --- | --- | --- | --- | --- |
| Keep rationale in owning documents | Separate learning material would duplicate boilerplate policy | Link one decision contract; maintain concrete domain reasoning locally | CONTRIBUTING and updated owning documents | Repository maintainer; revise when ownership or constraints change |
| Remove nested manual Hosted Full call | Manual Validation already calls the same full components | Keep the direct graph and required gate; standalone Hosted Full remains independent | Workflow call graph and regression tests | Repository maintainer; revisit if Hosted Full adds distinct checks |
| Cache npm downloads | Runner installs a committed lock on each run | Lock-keyed cache can reduce download work but consumes storage; never skip npm ci or checks | Hosted install/type/unit job; warm-cache benefit not yet measured | Repository maintainer; revisit if restore/storage overhead outweighs savings |
| Keep edited events and full PR checks | Base retargeting can invalidate earlier evidence; Quick transition is still under observation | Accept body-edit and Quick/full overlap until a safe transition is proven | Existing event and fail-closed gate regressions; Issue #116 | Repository maintainer; revisit with retarget/context tests and comparison evidence |

## Planned Implementation

- Update existing templates and domain documents without new learning-guide hierarchy.
- Remove one manual-only nested full call; preserve automatic events and gates.
- Add npm download caching for the manual runner with its committed lock key.
- Update regression assertions and correct stale CI topology prose.
- Link this issue/report from the existing PR and original report.

## Acceptance Criteria

- Constraints, alternatives, accepted costs, evidence and revisit conditions are explicit.
- Manual Validation calls each full component once; both required gates retain their names.
- Failure, cancellation and unexpected skipping still fail aggregate checks.
- PR/main, standalone Hosted Full, trusted base security inputs and locked installs remain.
- Newest hosted results identify the tested source revision.

## Verification and Delivery Plan

Inspect workflow call counts, event coverage, exact revision fallback, npm cache
configuration and unconditional installation. Run the existing workflow
regressions, including aggregate shell failure cases, in hosted repository CI.
Observe all newest required checks. No merge or deployment is requested.

## Approval

- **Status:** Approved
- **Approved by:** Soku-JINSEOK
- **Basis:** Explicit 2026-10-01 request for Issue creation, rationale improvements,
  CI optimization and additional commits to the existing PR.

## Implementation Status

Implemented in the additional PR #244 commit: workflow deduplication, lock-keyed
download caching, regression updates, decision contract and domain rationale.
Integration remains pending review and commit signing.

## Verification

- Inspected baseline [run 36807006720](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/36807006720):
  24 full component jobs, 760 summed job-seconds; not a billing measurement.
- Expected structural reduction: manual Validation full-component invocations
  fall from two trees to one. No claim of equivalent PR runtime reduction.
- Local execution unavailable: process provisioning failed before command execution.
- Hosted tests and final revision/run evidence are recorded in
  [Issue #245](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/245).
  Pending results are not passes.
- Cache speedup and live manual dispatch behavior are not measured in this report;
  workflow regressions establish the graph and retain fail-closed behavior.

## Public Disclosure Review

Only public repository paths, source revisions, issue/run links and summarized
job timings are recorded. No credentials, private infrastructure identifiers,
personal contact details, local paths or billing information are included.

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex
