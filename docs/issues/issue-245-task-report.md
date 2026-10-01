# Task report: explainable decisions and CI workload

> **Document purpose:** Task decision record. Records the scope, decisions and verification for Issue #245: explainable decisions and CI workload.
>
> **Key point:** Read approval, implementation status and verification separately; this report is dated task evidence, not a fresh claim of current completion.

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

## Evidence-driven cache correction

The first follow-up hosted Repository Hygiene job passed but reported that Go
cache restoration could not find a dependency file at the repository root.
The module lives under soku. Configure its cache inputs explicitly using
soku/go.sum, verification/tools.env and the owning CI workflow, so module and
tool-pin changes invalidate the cache. Tool installation and validation steps
remain unconditional; no warm-cache speedup is claimed.

## Public Disclosure Review

Only public repository paths, source revisions, issue/run links and summarized
job timings are recorded. No credentials, private infrastructure identifiers,
personal contact details, local paths or billing information are included.

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex

## Approved first-screen documentation follow-up

The owner explicitly requested detailed, easier diagrams throughout the
documentation and top summaries for descriptive documents on 2026-10-01.
Approval: Soku-JINSEOK's direct request to implement this documentation change.

The review inventory contains 134 existing Markdown files. Add purpose and key
takeaway blocks to 117 editable files, improve 25 document diagrams, and add a
complete document map plus a release-record index. Preserve the 17 versioned
release and append-only records byte for byte; their adjacent index supplies
the reader summaries. Historical task-report status and evidence remain intact.

The shared presentation contract lives in README_GUIDE and is referenced by
CONTRIBUTING and AGENTS. Every diagram explains how to read it and what to check.
Reference patterns, implemented behavior and proposed shadow behavior are
distinguished. The Korean lifecycle overview is also available in the existing
[FigJam board](https://www.figma.com/board/SJgcvEV1HZqYwHM5Nt5HWE).

Verification: inspect summary coverage, links and diagram structure; use hosted
repository validation because local process provisioning is unavailable.
Final source and hosted results are recorded in Issue #245.

## Approved diagram fidelity follow-up

The owner's 2026-10-01 request authorizes English-only diagram content, detailed
source comparison, Figma corrections and additional commits on PR #244.
Merge remains the owner's decision. Review baseline:
`99a3b7aa7d1c533db6d9936f48f43bb81975dc20`.

### Findings and implementation reasons

| Finding | Correction and reason | Evidence owner |
| --- | --- | --- |
| Localized diagram labels diverged | Use one English vocabulary; keep surrounding localized prose | BLUEPRINT language policy and all three READMEs |
| Lifecycle arrow implied writes to application files | Represent project ownership as a planning constraint | initcmd engine and upgrade ownership checks |
| Unconfirmed apply looked like dry-run | Separate explicit dry-run, cancellation and confirmation error | engine.go and upgrade.go |
| Manifest/cleanup failures were hidden | Show manifest inside rollback boundary; distinguish committed cleanup failure | transaction.go |
| Cached launcher appeared freshly verified | State existence-only reuse; checksum covers newly downloaded archive | npm/lib/launcher.mjs |
| Cloud recovery appeared unconditional | Show absent prior revision, early failure and unsuccessful rollback; recovered deploy still fails | cd-deploy.sh |
| Public deployment artifact described absent fields | List actual sanitized fields; keep revision inspection with authorized operators | record_evidence in cd-deploy.sh |
| DNS looked like an HTTP transit component | Separate name resolution from HTTPS traffic and record ownership | CLOUD_POLICY reference model |
| Project Sync looked atomic | Explain fresh in-memory audit, per-operation conflicts, field-batch checks and continued independent operations | github-project-sync.mjs |
| Retry arrow implied requeue | Keep retry inside active serialized slot; disclose GET-only budget | GitHubApiClient.requestWithRetry |
| CI diagrams conflated revisions | State PR head versus merge-ref checkout and independent gate boundaries | validation.yml and called workflows |
| Bootstrap picture confused state with sequence | Show explicit apply decision and separate state-ownership edges | gcp-bootstrap.sh and Terraform roots |

### Complete diagram coverage

Review all 27 inline Mermaid diagrams across 25 documents, plus the standalone
Mermaid source and three SVG reference illustrations. Inline figures now state
their scope, provide reading guidance and reader checks. Implemented mechanisms
link to owning sources; normative/reference figures make no deployed-system
claim. The extra inline figure separates lifecycle consent from recovery.

| Document | Inline figures | Classification |
| --- | --- | --- |
| [BLUEPRINT.md](../../BLUEPRINT.md) | 1 | Normative authority map |
| [README.ja.md](../../README.ja.md) | 1 | Implemented lifecycle overview |
| [README.ko.md](../../README.ko.md) | 1 | Implemented lifecycle overview |
| [README.md](../../README.md) | 1 | Implemented lifecycle overview |
| [VERIFICATION_GUIDE.md](../../VERIFICATION_GUIDE.md) | 1 | Reference verification model |
| [docs/guides/APPLICABILITY.md](../../docs/guides/APPLICABILITY.md) | 1 | Reference decision aid |
| [docs/guides/CLOUD_BUILD_SHADOW.md](../../docs/guides/CLOUD_BUILD_SHADOW.md) | 1 | Proposed operational path with an implemented local contract check |
| [docs/guides/CLOUD_RUN_CICD.md](../../docs/guides/CLOUD_RUN_CICD.md) | 1 | Implemented deployment and recovery |
| [docs/guides/GITHUB_PROJECT_SYNC.md](../../docs/guides/GITHUB_PROJECT_SYNC.md) | 2 | Implemented metadata mutation and API scheduling |
| [docs/guides/INIT_GUIDE.md](../../docs/guides/INIT_GUIDE.md) | 1 | Adoption procedure |
| [docs/guides/LANGUAGE_SELECTION.md](../../docs/guides/LANGUAGE_SELECTION.md) | 1 | Reference decision aid |
| [docs/guides/USAGE_MANUAL.md](../../docs/guides/USAGE_MANUAL.md) | 1 | Adoption procedure |
| [docs/issues/TASK_REPORT_TEMPLATE.md](../../docs/issues/TASK_REPORT_TEMPLATE.md) | 1 | Normative evidence lifecycle |
| [docs/policy/CLOUD_POLICY.md](../../docs/policy/CLOUD_POLICY.md) | 1 | Reference cloud responsibility model |
| [docs/policy/SECURITY_POLICY.md](../../docs/policy/SECURITY_POLICY.md) | 1 | Reference request-security model |
| [docs/standards/CICD_STANDARDS.md](../../docs/standards/CICD_STANDARDS.md) | 1 | Implemented workflow topology |
| [docs/standards/CODE_STYLE.md](../../docs/standards/CODE_STYLE.md) | 1 | Reference module contract |
| [docs/standards/GITHUB_STANDARDS.md](../../docs/standards/GITHUB_STANDARDS.md) | 1 | Normative contribution lifecycle |
| [docs/standards/PROJECT_STRUCTURE.md](../../docs/standards/PROJECT_STRUCTURE.md) | 1 | Reference ownership and dependency model |
| [docs/standards/REAL_RUNTIME_MANUAL_CAPTURE.md](../../docs/standards/REAL_RUNTIME_MANUAL_CAPTURE.md) | 1 | Implemented capture path with human review |
| [docs/standards/RELEASE_AND_SYNC.md](../../docs/standards/RELEASE_AND_SYNC.md) | 1 | Normative release and adoption lifecycle |
| [docs/standards/SOKU_LIFECYCLE.md](../../docs/standards/SOKU_LIFECYCLE.md) | 2 | Implemented consent and transaction paths |
| [docs/standards/SUPPLY_CHAIN.md](../../docs/standards/SUPPLY_CHAIN.md) | 1 | Normative dependency-review model |
| [infra/gcp/README.md](../../infra/gcp/README.md) | 1 | Implemented bootstrap and declared state boundaries |
| [soku/npm/README.md](../../soku/npm/README.md) | 1 | Implemented launcher |

The standalone `docs/assets/review-evidence.mmd` and
`docs/assets/review-evidence.svg` describe a reference review process.
`runtime-boundaries.svg` compares possible service and offline arrangements;
`delivery-decision.svg` is a reference readiness/recovery model. Their English
labels describe design questions, not resources discovered in this repository.
Their owning usage/policy/verification sections explain applicability.

### Enforcement and CI cost

Add a dependency-free tracked-file text guard to the existing Repository Hygiene
job and local equivalent. It examines Markdown Mermaid fences, standalone
Mermaid and SVG text, rejecting non-ASCII letters while allowing punctuation.
Focused fixtures cover localized prose, tilde fences, nested Markdown examples,
unclosed Mermaid and encoded SVG text. The guard cannot prove English grammar,
Mermaid rendering or semantic parity; source review remains mandatory.

No new job, browser installation, renderer service or package dependency is
introduced. Preserve the previous manual full-tree deduplication, npm/Go caches,
cancellation behavior and fail-closed full/Quick/metadata gates. Batch metadata
edits before the final source push; record final results in the issue rather
than repeatedly editing the PR and retriggering Validation. No measured speedup
is claimed for this additional lightweight check.

### Figma and verification limits

The existing editable board is updated to English and identifies authority and
security as normative/reference maps, and CI/lifecycle as implemented overviews.
Correct ownership wiring and missing security denial paths. Read back all text
nodes and inspect a screenshot for legibility and routing.

Local command execution is unavailable because process provisioning fails before
execution. Static prepared-content checks and final hosted tests provide the
available evidence. This review does not claim live Cloud Run deployment,
rollback drills, manual workflow dispatch or browser rendering of every GitHub
Mermaid figure. Final commit, run and check results belong in Issue #245.
