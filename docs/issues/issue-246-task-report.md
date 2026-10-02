# Operational follow-up for Issue 246

> **Document purpose:** Records the implemented follow-up to the existing integration tracker.
>
> **Key point:** Profiles and validation measurement are implemented; signed integration and a new natural observation window remain owner actions.

## Goal and Background

Reduce downstream setup work and make CI cost/coverage observable, using the
existing [integration tracker](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/246).
This branch builds on PR 244 at `0734f737bda0c986a44832aa48863c85f5e674a5`;
it retains that PR's dependency repairs, diagrams, and single Full tree.

## Approval

Implementation was authorized by the user's instruction in this Codex chat:
“다음 작업을 실제로 운용한다.” The user additionally requested research into
commit signing from Codex cloud. This records implementation/review preparation,
not permission to bypass signature requirements, reduce required checks,
publish a release, or deploy infrastructure.

## Decisions and Implementation

| Decision | Reason, alternative, and accepted cost |
| --- | --- |
| Layer on PR 244 | Retain already validated common security fixes and avoid a conflicting duplicate Full-tree repair; integration depends on its signed handoff. |
| Complete stacks in bootstrap | A file-list prefix cannot provide a runnable starter. Reuse the same tested stack payload and Quick/Security workflows; add Full and concise forms only in standard. |
| Keep catalog v2 compatible | Existing readers already support complete stacks and layer files; keep legacy positional reads and fix shared-catalog mutation rather than add another schema. |
| Match support declarations to evidence | Python and Node receive explicit Full matrices; Go/Java declarations and Terraform's tested minor are checked for drift. Wider ranges require new evidence. |
| Reuse policy and audit documents | Generate mutable tool values and collect attempt-level JSON; avoid a parallel policy source or a fabricated success-only sample set. |

## Acceptance and Verification

- Pass: 231 repository Node tests (`scripts/*.test.mjs`, `.github/*.test.mjs`).
- Pass: all Soku Go packages with `go test ./...`, including every published
  stack/profile output combination, legacy isolation, transitions, and rollback.
- Pass: Node 22.12.0, 24.21.0, and 26.10.0 lint, typecheck, tests, build, format.
- Pass: Python 3.11–3.14 lint, typecheck, format, and tests in the priority phase.
- Pass: Go template tests; generated Full workflow parity; policy/support parity;
  changed Markdown/YAML/Actions lint and diagram text guard.
- Pass: Terraform 1.15.3 format and both modules' backend-disabled initialization
  and validation. No cloud plan/apply, state migration, or infrastructure change.
- Local limitation: Java 21 runtime lacks the full JDK `ct.sym`; Maven resolves
  dependencies through the session proxy but local `--release 21` compilation
  cannot complete. Hosted `setup-java` validation remains required.
- Actual observation: run 36806174238 attempt 1 has all 29 Jobs API records.
  Quick passes (4 jobs, 49 job-seconds); Full fails (24 jobs, 691 job-seconds).
  One unclassified job remains explicit. See the existing
  [comparison record](../audits/ci-quick-comparison.md); this is not an eligible
  sample from a new observation epoch.

## Cloud Signing Findings

A disposable SSH signing probe created and locally verified a commit in this
cloud environment. Its key and repository were deleted without upload.
GitHub's official GraphQL schema documents server signing through
`createCommitOnBranch` when supported. The current connector exposes unsigned
Git Database commits, and the injected `gh` credential is invalid; account-level
server signing was not exercised. No user signing material was accessed.
The [signing handoff](../standards/GITHUB_STANDARDS.md#cloud-signing-options)
distinguishes these paths and requires GitHub verification plus final-head CI.

## Integration and Remaining Work

- Keep the source PR Draft until hosted verification and the owner signing path
  are resolved. Retain PR 244 as the stack base until its integration is reviewed.
- Signing a new child does not sign existing unsigned parents. Compare the
  replacement tree, verify introduced commits, and revalidate the new head.
- The metrics completion workflow activates on the default branch after merge.
  Record that merge and freeze coverage before collecting natural samples.
- Keep Issue 117 blocked on Issue 116's coverage and observation criteria.
- Reconcile dependency PRs only after their common fixes integrate; do not close
  implementation/integration issues merely because local tests passed.
- Profile adoption remains an explicit source release/transition with dry-run
  and collision review. No downstream project was mutated by this work.

## AI Assistance

Planning, implementation, research, tests, and drafting: OpenAI Codex.
