# Complete documentation map

> **Document purpose:** Reading index. Find the purpose, authority and presentation of every Markdown document in this repository.
>
> **Key point:** Start from your question, then open the owning guide or contract; historical evidence is not a current operating instruction.

## Choose your reading route

| Your question | Start with | Expected result |
| --- | --- | --- |
| What is this project and how do I adopt it? | [README](../../README.md), then [usage manual](./USAGE_MANUAL.md) | Understand the baseline and preview the intended changes |
| Why are rules, files or technologies arranged this way? | [BLUEPRINT](../../BLUEPRINT.md), then the owning standard | Identify authority, constraints, alternatives and owners |
| What happens on success, failure or recovery? | The relevant lifecycle, CI, capture or delivery guide below | Trace the normal and exceptional path |
| How do I know the change was verified? | [Verification guide](../../VERIFICATION_GUIDE.md), then the task report | Separate planned checks from revision-specific results |
| What did an earlier release or task establish? | Historical records below | Read the evidence with its version, date and limitations |

## Presentation legend

- **Diagram + summary:** relationships or branching behavior, followed by reading guidance.
- **Summary:** purpose and key conclusion for descriptive, policy or example content.
- **Historical summary:** introductory explanation; the original task status and evidence remain unchanged.
- **Preserved record:** original release or append-only content; orientation is supplied by the [release index](../releases/README.md).

The inventory below covers 136 Markdown files: 134 pre-existing documents and
these two new indexes. Update it with document additions or moves. The
[presentation contract](./README_GUIDE.md) governs future updates.

## Start here and repository rules

| Document | Purpose | Presentation |
| --- | --- | --- |
| [AGENTS.md](../../AGENTS.md) | Defines how AI agents read authority, respect ownership and make reviewable repository changes. | Summary |
| [BLUEPRINT.md](../../BLUEPRINT.md) | Defines the repository's purpose, document authority and boundaries between shared conventions and project design. | Diagram + summary |
| [CONTRIBUTING.md](../../CONTRIBUTING.md) | Explains how to propose, implement, verify and review changes with explicit decision reasons. | Summary |
| [README.ja.md](../../README.ja.md) | 共通規約とSokuツールの役割、導入方法、詳細文書への入口を案内します。 | Diagram + summary |
| [README.ko.md](../../README.ko.md) | 공통 개발 규칙과 Soku 도구의 역할, 도입 방법, 상세 문서의 위치를 안내합니다. | Diagram + summary |
| [README.md](../../README.md) | Start here to understand the convention baseline, Soku lifecycle tooling and where to find setup and operating rules. | Diagram + summary |
| [SECURITY.md](../../SECURITY.md) | Explains how to report a suspected security issue privately and find the operating security policy. | Summary |
| [VERIFICATION_GUIDE.md](../../VERIFICATION_GUIDE.md) | Maps repository, runtime, governance and delivery claims to executable checks and retained evidence. | Diagram + summary |

## Guides

| Document | Purpose | Presentation |
| --- | --- | --- |
| [docs/guides/APPLICABILITY.md](APPLICABILITY.md) | Helps personal, connected and team projects select proportionate conventions and evidence. | Diagram + summary |
| [docs/guides/CLOUD_BUILD_SHADOW.md](CLOUD_BUILD_SHADOW.md) | Defines an isolated Cloud Build comparison candidate and the evidence required before operational use. | Diagram + summary |
| [docs/guides/CLOUD_RUN_CICD.md](CLOUD_RUN_CICD.md) | Explains GCP bootstrap, manual dev deployment, authenticated health checks and rollback. | Diagram + summary |
| [docs/guides/DOCUMENTATION_MAP.md](DOCUMENTATION_MAP.md) | Find every Markdown document by role, topic and presentation. | Summary |
| [docs/guides/GITHUB_PROJECT_SYNC.md](GITHUB_PROJECT_SYNC.md) | Explains installation, audit, guarded apply and recovery for Issue and Project metadata synchronization. | Diagram + summary |
| [docs/guides/INIT_GUIDE.md](INIT_GUIDE.md) | Provides the legacy manual fallback for detecting a downstream stack and applying the appropriate baseline. | Diagram + summary |
| [docs/guides/LANGUAGE_SELECTION.md](LANGUAGE_SELECTION.md) | Compares language choices through platform constraints, operating capability and representative workload evidence. | Diagram + summary |
| [docs/guides/NPM_TRUSTED_PUBLISHING.md](NPM_TRUSTED_PUBLISHING.md) | Defines the GitHub-to-npm trusted publisher identity and the tokenless release verification contract. | Summary |
| [docs/guides/PROJECT_SYNC_CREDENTIAL_RUNBOOK.md](PROJECT_SYNC_CREDENTIAL_RUNBOOK.md) | Defines scoped credential setup, replacement audits, secret rotation and revocation for Project Sync. | Summary |
| [docs/guides/README_GUIDE.md](README_GUIDE.md) | Defines what readers should learn immediately from a README and how to navigate deeper documentation. | Summary |
| [docs/guides/REAL_RUNTIME_MANUAL_CAPTURE.ja.md](REAL_RUNTIME_MANUAL_CAPTURE.ja.md) | 実画面をローカルで動かし、マニュアル用画像と出典の証拠を作成する手順を要約します。 | Summary |
| [docs/guides/REAL_RUNTIME_MANUAL_CAPTURE.ko.md](REAL_RUNTIME_MANUAL_CAPTURE.ko.md) | 실제 화면을 로컬에서 실행하여 사용자 매뉴얼용 이미지와 출처 근거를 만드는 절차를 요약합니다. | Summary |
| [docs/guides/SOKU_TERMINAL_GUIDE.ja.md](SOKU_TERMINAL_GUIDE.ja.md) | Sokuのターミナル表示、日常のコマンド、シェル補完設定を説明します。 | Summary |
| [docs/guides/SOKU_TERMINAL_GUIDE.ko.md](SOKU_TERMINAL_GUIDE.ko.md) | Soku 터미널 표시, 일상적인 명령 사용, 셸 자동완성 설정을 설명합니다. | Summary |
| [docs/guides/SOKU_TERMINAL_GUIDE.md](SOKU_TERMINAL_GUIDE.md) | Explains readable terminal output, safe everyday commands and shell completion. | Summary |
| [docs/guides/STACK_CONFIGS.md](STACK_CONFIGS.md) | Maps each supported stack to its copyable formatter, build, test and runtime configuration files. | Summary |
| [docs/guides/STACK_EXAMPLES.md](STACK_EXAMPLES.md) | Shows small examples of explicit interfaces and readable behavior across supported stacks. | Summary |
| [docs/guides/USAGE_MANUAL.md](USAGE_MANUAL.md) | Connects project review, installation, safe initialization, validation, governance and optional delivery. | Diagram + summary |

## Standards and policies

| Document | Purpose | Presentation |
| --- | --- | --- |
| [docs/policy/CLOUD_POLICY.md](../policy/CLOUD_POLICY.md) | Defines provider and service selection through workload, identity, recovery and operating cost. | Diagram + summary |
| [docs/policy/LICENSE_POLICY.md](../policy/LICENSE_POLICY.md) | Explains how projects select, declare and maintain licensing and dependency obligations. | Summary |
| [docs/policy/SECURITY_POLICY.md](../policy/SECURITY_POLICY.md) | Defines secret hygiene, identity, resource authorization, dependency review and response responsibilities. | Diagram + summary |
| [docs/standards/CICD_STANDARDS.md](../standards/CICD_STANDARDS.md) | Explains validation responsibilities, required gates, execution cost and controlled delivery. | Diagram + summary |
| [docs/standards/CODE_STYLE.md](../standards/CODE_STYLE.md) | Defines readable modules, explicit inputs and effects, naming, formatting and behavior checks. | Diagram + summary |
| [docs/standards/GITHUB_STANDARDS.md](../standards/GITHUB_STANDARDS.md) | Defines issue, task-report, pull-request, label, signature and review requirements. | Diagram + summary |
| [docs/standards/PROJECT_STRUCTURE.md](../standards/PROJECT_STRUCTURE.md) | Maps directories to responsibilities, deployable units, data ownership and shared interfaces. | Diagram + summary |
| [docs/standards/REAL_RUNTIME_MANUAL_CAPTURE.md](../standards/REAL_RUNTIME_MANUAL_CAPTURE.md) | Defines local real-runtime screenshots, ownership, adapters, provenance and redacted reports. | Diagram + summary |
| [docs/standards/RELEASE_AND_SYNC.md](../standards/RELEASE_AND_SYNC.md) | Defines independent convention and CLI releases, immutable identities and scoped downstream updates. | Diagram + summary |
| [docs/standards/SOKU_LIFECYCLE.md](../standards/SOKU_LIFECYCLE.md) | Defines Soku ownership, compatibility, planning, transactional writes and recovery. | Diagram + summary |
| [docs/standards/SUPPLY_CHAIN.md](../standards/SUPPLY_CHAIN.md) | Defines authoritative versions, lockfiles, pinned tools and generated validation outputs. | Diagram + summary |

## Operating and distributed documentation

| Document | Purpose | Presentation |
| --- | --- | --- |
| [.github/COMMENT_TEMPLATES.md](../../.github/COMMENT_TEMPLATES.md) | Copyable review and progress comments with findings, blockers and next actions. | Summary |
| [.github/PULL_REQUEST_TEMPLATE.md](../../.github/PULL_REQUEST_TEMPLATE.md) | Collect the linked issue, decision rationale, verification and remaining risks for a reviewable change. | Summary |
| [infra/gcp/README.md](../../infra/gcp/README.md) | Explains Terraform foundation, runtime, optional validation resources and isolated state responsibilities. | Diagram + summary |
| [infra/gcp/cloud-build-logging/README.md](../../infra/gcp/cloud-build-logging/README.md) | Describes the isolated validation log bucket, sink and disabled rollout exclusion. | Summary |
| [soku/README.md](../../soku/README.md) | Explains Soku installation, command behavior, supported lifecycle features and distribution. | Summary |
| [soku/THIRD_PARTY_NOTICES.md](../../soku/THIRD_PARTY_NOTICES.md) | Lists bundled Go modules, pinned versions and license references for the Soku binary. | Summary |
| [soku/internal/manual/assets/templates/MANUAL_TEMPLATE.md](../../soku/internal/manual/assets/templates/MANUAL_TEMPLATE.md) | Provides the structure for project-owned instructions backed by stable capture identifiers. | Summary |
| [soku/npm/README.md](../../soku/npm/README.md) | Explains the npm launcher that downloads, verifies and runs the matching native Soku executable. | Diagram + summary |
| [templates/_shared/agents/README.md](../../templates/_shared/agents/README.md) | Explains how tool-neutral domain ownership rules are connected to an AI coding tool. | Summary |
| [templates/_shared/agents/app-agent.md](../../templates/_shared/agents/app-agent.md) | Defines responsibilities and editing boundaries for the app/ domain. | Summary |
| [templates/_shared/agents/backend-agent.md](../../templates/_shared/agents/backend-agent.md) | Defines responsibilities and editing boundaries for the backend/ domain. | Summary |
| [templates/_shared/agents/db-agent.md](../../templates/_shared/agents/db-agent.md) | Defines responsibilities and editing boundaries for the db/ domain. | Summary |
| [templates/_shared/agents/docs-agent.md](../../templates/_shared/agents/docs-agent.md) | Defines responsibilities and editing boundaries for the docs/ domain. | Summary |
| [templates/_shared/agents/frontend-agent.md](../../templates/_shared/agents/frontend-agent.md) | Defines responsibilities and editing boundaries for the frontend/ domain. | Summary |
| [templates/_shared/agents/infra-agent.md](../../templates/_shared/agents/infra-agent.md) | Defines responsibilities and editing boundaries for the infra/ domain. | Summary |
| [verification/CLASSIFICATION.md](../../verification/CLASSIFICATION.md) | Classifies checks as local-capable, hosted-only, release-only or deployment-only. | Summary |

## Audit and task evidence

| Document | Purpose | Presentation |
| --- | --- | --- |
| [docs/audits/ci-quick-comparison.md](../audits/ci-quick-comparison.md) | Records the Quick-versus-Full observation criteria and measurements for Issue #116. | Historical summary |
| [docs/audits/github-governance-2026-07-23.md](../audits/github-governance-2026-07-23.md) | Records the read-only governance inventory of 33 issues and 59 pull requests at the stated cutoff. | Historical summary |
| [docs/issues/TASK_REPORT_TEMPLATE.md](../issues/TASK_REPORT_TEMPLATE.md) | Captures a task's problem, alternatives, approval, implementation and actual verification. | Diagram + summary |
| [docs/issues/issue-10-task-report.md](../issues/issue-10-task-report.md) | Records the scope, decisions and verification for Issue #10: validate templates and harden sync scripts. | Historical summary |
| [docs/issues/issue-100-task-report.md](../issues/issue-100-task-report.md) | Records the scope, decisions and verification for Issue #100: Audit complete Issue and pull-request history. | Historical summary |
| [docs/issues/issue-101-task-report.md](../issues/issue-101-task-report.md) | Records the scope, decisions and verification for Issue #101: Add npm wrapper package for CLI distribution. | Historical summary |
| [docs/issues/issue-102-task-report.md](../issues/issue-102-task-report.md) | Records the scope, decisions and verification for Issue #102: Simplify issue templates for operator efficiency. | Historical summary |
| [docs/issues/issue-110-task-report.md](../issues/issue-110-task-report.md) | Records the scope, decisions and verification for Issue #110: Silence CD plan and summary logs in node tests. | Historical summary |
| [docs/issues/issue-114-task-report.md](../issues/issue-114-task-report.md) | Records the scope, decisions and verification for Issue #114: Freeze and classify current CI checks. | Historical summary |
| [docs/issues/issue-115-task-report.md](../issues/issue-115-task-report.md) | Records the scope, decisions and verification for Issue #115: Add fast verification and scope detection. | Historical summary |
| [docs/issues/issue-116-task-report.md](../issues/issue-116-task-report.md) | Records the scope, decisions and verification for Issue #116: Run CI Quick beside full validation. | Historical summary |
| [docs/issues/issue-117-task-report.md](../issues/issue-117-task-report.md) | Records the scope, decisions and verification for Issue #117: preserving required code validation and adding current-source Hosted Full. | Historical summary |
| [docs/issues/issue-120-task-report.md](../issues/issue-120-task-report.md) | Records the scope, decisions and verification for Issue #120: Reduce public metadata exposure. | Historical summary |
| [docs/issues/issue-122-task-report.md](../issues/issue-122-task-report.md) | Records the scope, decisions and verification for Issue #122: Align npm package license and tarball contents. | Historical summary |
| [docs/issues/issue-123-task-report.md](../issues/issue-123-task-report.md) | Records the scope, decisions and verification for Issue #123: Unify version metadata and publication identity. | Historical summary |
| [docs/issues/issue-124-task-report.md](../issues/issue-124-task-report.md) | Records the scope, decisions and verification for Issue #124: Establish immutable supply-chain inputs. | Historical summary |
| [docs/issues/issue-125-task-report.md](../issues/issue-125-task-report.md) | Records the scope, decisions and verification for Issue #125: Isolate fork PRs from token-backed tests. | Historical summary |
| [docs/issues/issue-136-task-report.md](../issues/issue-136-task-report.md) | Records the scope, decisions and verification for Issue #136: Add validation-only Cloud Build checks. | Historical summary |
| [docs/issues/issue-14-task-report.md](../issues/issue-14-task-report.md) | Records the scope, decisions and verification for Issue #14: Language Selection Guide. | Historical summary |
| [docs/issues/issue-155-task-report.md](../issues/issue-155-task-report.md) | Records the scope, decisions and verification for Issue #155: Add low-cost GCP sandbox guardrails. | Historical summary |
| [docs/issues/issue-16-task-report.md](../issues/issue-16-task-report.md) | Records the scope, decisions and verification for Issue #16: `soku` Lifecycle Contract. | Historical summary |
| [docs/issues/issue-163-implementation-amendment.md](../issues/issue-163-implementation-amendment.md) | Records the scope, decisions and verification for Issue #163: the CI/CD decision planner and its implementation amendment. | Historical summary |
| [docs/issues/issue-163-task-report.md](../issues/issue-163-task-report.md) | Records the scope, decisions and verification for Issue #163: the CI/CD decision planner and its implementation amendment. | Historical summary |
| [docs/issues/issue-164-task-report.md](../issues/issue-164-task-report.md) | Records the scope, decisions and verification for Issue #164: Real-runtime user-manual capture. | Historical summary |
| [docs/issues/issue-17-task-report.md](../issues/issue-17-task-report.md) | Records the scope, decisions and verification for Issue #17: `soku` CLI Shell and Distribution. | Historical summary |
| [docs/issues/issue-175-task-report.md](../issues/issue-175-task-report.md) | Records the scope, decisions and verification for Issue #175: Restore Go caching in CI Quick. | Historical summary |
| [docs/issues/issue-18-task-report.md](../issues/issue-18-task-report.md) | Records the scope, decisions and verification for Issue #18: Transactional `soku init`. | Historical summary |
| [docs/issues/issue-180-task-report.md](../issues/issue-180-task-report.md) | Records the scope, decisions and verification for Issue #180: isolating validation logging state and restoring authenticated checks. | Historical summary |
| [docs/issues/issue-184-task-report.md](../issues/issue-184-task-report.md) | Records the scope, decisions and verification for Issue #184: removing a stale Dependabot Docker target. | Historical summary |
| [docs/issues/issue-186-task-report.md](../issues/issue-186-task-report.md) | Records the scope, decisions and verification for Issue #186: terminal presentation and deterministic shell completion. | Historical summary |
| [docs/issues/issue-188-task-report.md](../issues/issue-188-task-report.md) | Records the scope, decisions and verification for Issue #188: remediating npm lockfile findings. | Historical summary |
| [docs/issues/issue-19-task-report.md](../issues/issue-19-task-report.md) | Records the scope, decisions and verification for Issue #19: Portable Manifest and `soku status`. | Historical summary |
| [docs/issues/issue-192-task-report.md](../issues/issue-192-task-report.md) | Records the scope, decisions and verification for Issue #192: the optional downstream Project Sync component. | Historical summary |
| [docs/issues/issue-195-task-report.md](../issues/issue-195-task-report.md) | Records the scope, decisions and verification for Issue #195: Downstream Project Sync Audit. | Historical summary |
| [docs/issues/issue-197-task-report.md](../issues/issue-197-task-report.md) | Records the scope, decisions and verification for Issue #197: Project Sync Credential Rotation. | Historical summary |
| [docs/issues/issue-198-task-report.md](../issues/issue-198-task-report.md) | Records the scope, decisions and verification for Issue #198: Immutable CI/CD adapter conformance. | Historical summary |
| [docs/issues/issue-20-task-report.md](../issues/issue-20-task-report.md) | Records the scope, decisions and verification for Issue #20: Transactional `soku diff` and `soku upgrade`. | Historical summary |
| [docs/issues/issue-201-task-report.md](../issues/issue-201-task-report.md) | Records the scope, decisions and verification for Issue #201: Explicit project ownership handoff. | Historical summary |
| [docs/issues/issue-207-task-report.md](../issues/issue-207-task-report.md) | Records the scope, decisions and verification for Issue #207: Reposition README to clarify Soku lifecycle and sync translations. | Historical summary |
| [docs/issues/issue-208-task-report.md](../issues/issue-208-task-report.md) | Records the scope, decisions and verification for Issue #208: Strict Fail-Closed Shadow Correction. | Historical summary |
| [docs/issues/issue-21-task-report.md](../issues/issue-21-task-report.md) | Records the scope, decisions and verification for Issue #21: Lifecycle End-to-End Release Gate. | Historical summary |
| [docs/issues/issue-22-task-report.md](../issues/issue-22-task-report.md) | Records the scope, decisions and verification for Issue #22: Profiles and Bounded Declarative Extensions. | Historical summary |
| [docs/issues/issue-227-task-report.md](../issues/issue-227-task-report.md) | Records the scope, decisions and verification for Issue #227: combining manual-runner dependency coverage and a fast-uri update. | Historical summary |
| [docs/issues/issue-23-task-report.md](../issues/issue-23-task-report.md) | Records the scope, decisions and verification for Issue #23: Soku Lifecycle Roadmap Closure. | Historical summary |
| [docs/issues/issue-235-task-report.md](../issues/issue-235-task-report.md) | Records the scope, decisions and verification for Issue #235: Archify repository flow maps. | Historical summary |
| [docs/issues/issue-24-task-report.md](../issues/issue-24-task-report.md) | Records the scope, decisions and verification for Issue #24: aligning issue and PR authoring contracts. | Historical summary |
| [docs/issues/issue-243-task-report.md](../issues/issue-243-task-report.md) | Records the scope, decisions and verification for Issue #243: design and readiness documentation. | Historical summary |
| [docs/issues/issue-245-task-report.md](../issues/issue-245-task-report.md) | Records the scope, decisions and verification for Issue #245: explainable decisions and CI workload. | Historical summary |
| [docs/issues/issue-35-task-report.md](../issues/issue-35-task-report.md) | Records the scope, decisions and verification for Issue #35: release-gate repair and first release preparation. | Historical summary |
| [docs/issues/issue-39-task-report.md](../issues/issue-39-task-report.md) | Records the scope, decisions and verification for Issue #39: repository governance, integrity and cost audit. | Historical summary |
| [docs/issues/issue-40-task-report.md](../issues/issue-40-task-report.md) | Records the scope, decisions and verification for Issue #40: Replace Vulnerable Pyink with Black. | Historical summary |
| [docs/issues/issue-41-task-report.md](../issues/issue-41-task-report.md) | Records the scope, decisions and verification for Issue #41: a corrective convention release and compatibility evidence. | Historical summary |
| [docs/issues/issue-44-task-report.md](../issues/issue-44-task-report.md) | Records the scope, decisions and verification for Issue #44: using the fetched immutable provider revision as authority. | Historical summary |
| [docs/issues/issue-54-task-report.md](../issues/issue-54-task-report.md) | Records the scope, decisions and verification for Issue #54: the public Provider API mirror. | Historical summary |
| [docs/issues/issue-55-task-report.md](../issues/issue-55-task-report.md) | Records the scope, decisions and verification for Issue #55: GitHub governance hardening. | Historical summary |
| [docs/issues/issue-56-task-report.md](../issues/issue-56-task-report.md) | Records the scope, decisions and verification for Issue #56: dependency, secret, license and signed-tag gates. | Historical summary |
| [docs/issues/issue-69-task-report.md](../issues/issue-69-task-report.md) | Records the scope, decisions and verification for Issue #69: the July 2026 dependency update batch. | Historical summary |
| [docs/issues/issue-72-task-report.md](../issues/issue-72-task-report.md) | Records the scope, decisions and verification for Issue #72: central governance template alignment. | Historical summary |
| [docs/issues/issue-78-task-report.md](../issues/issue-78-task-report.md) | Records the scope, decisions and verification for Issue #78: the registered provider public mirror. | Historical summary |
| [docs/issues/issue-8-task-report.md](../issues/issue-8-task-report.md) | Records the scope, decisions and verification for Issue #8: add task-report and title-check templates. | Historical summary |
| [docs/issues/issue-85-task-report.md](../issues/issue-85-task-report.md) | Records the scope, decisions and verification for Issue #85: Dependabot schema repair. | Historical summary |
| [docs/issues/issue-87-task-report.md](../issues/issue-87-task-report.md) | Records the scope, decisions and verification for Issue #87: confirmed GCP bootstrap. | Historical summary |
| [docs/issues/issue-91-task-report.md](../issues/issue-91-task-report.md) | Records the scope, decisions and verification for Issue #91: authenticated private Cloud Run health evidence. | Historical summary |
| [docs/issues/issue-93-task-report.md](../issues/issue-93-task-report.md) | Records the scope, decisions and verification for Issue #93: Make Dependabot governance deterministic. | Historical summary |
| [docs/issues/issue-98-task-report.md](../issues/issue-98-task-report.md) | Records the scope, decisions and verification for Issue #98: Add an end-to-end boilerplate manual. | Historical summary |

## Versioned release records

| Document | Purpose | Presentation |
| --- | --- | --- |
| [docs/releases/README.md](../releases/README.md) | Locate each version's compatibility/migration record and the append-only signer history. | Summary |
| [docs/releases/SIGNER_ROTATIONS.md](../releases/SIGNER_ROTATIONS.md) | Signer rotation procedure and historical trust record. | Preserved record |
| [docs/releases/soku-v0.1.0.md](../releases/soku-v0.1.0.md) | Version-specific compatibility, migration and changes: soku v0.1.0. | Preserved record |
| [docs/releases/soku-v0.1.1.md](../releases/soku-v0.1.1.md) | Version-specific compatibility, migration and changes: soku v0.1.1. | Preserved record |
| [docs/releases/soku-v0.1.2.md](../releases/soku-v0.1.2.md) | Version-specific compatibility, migration and changes: soku v0.1.2. | Preserved record |
| [docs/releases/soku-v0.1.3.md](../releases/soku-v0.1.3.md) | Version-specific compatibility, migration and changes: soku v0.1.3. | Preserved record |
| [docs/releases/soku-v0.1.4.md](../releases/soku-v0.1.4.md) | Version-specific compatibility, migration and changes: soku v0.1.4. | Preserved record |
| [docs/releases/soku-v0.2.0.md](../releases/soku-v0.2.0.md) | Version-specific compatibility, migration and changes: soku v0.2.0. | Preserved record |
| [docs/releases/soku-v0.2.1.md](../releases/soku-v0.2.1.md) | Version-specific compatibility, migration and changes: soku v0.2.1. | Preserved record |
| [docs/releases/soku-v0.2.2.md](../releases/soku-v0.2.2.md) | Version-specific compatibility, migration and changes: soku v0.2.2. | Preserved record |
| [docs/releases/soku-v0.2.3.md](../releases/soku-v0.2.3.md) | Version-specific compatibility, migration and changes: soku v0.2.3. | Preserved record |
| [docs/releases/soku-v0.3.0.md](../releases/soku-v0.3.0.md) | Version-specific compatibility, migration and changes: soku v0.3.0 release candidate. | Preserved record |
| [docs/releases/v1.0.0.md](../releases/v1.0.0.md) | Version-specific compatibility, migration and changes: Boilerplate v1.0.0. | Preserved record |
| [docs/releases/v1.0.1.md](../releases/v1.0.1.md) | Version-specific compatibility, migration and changes: Boilerplate v1.0.1. | Preserved record |
| [docs/releases/v1.0.2.md](../releases/v1.0.2.md) | Version-specific compatibility, migration and changes: Boilerplate v1.0.2. | Preserved record |
| [docs/releases/v1.0.3.md](../releases/v1.0.3.md) | Version-specific compatibility, migration and changes: Boilerplate v1.0.3. | Preserved record |
| [docs/releases/v1.0.4.md](../releases/v1.0.4.md) | Version-specific compatibility, migration and changes: Boilerplate v1.0.4. | Preserved record |
| [docs/releases/v1.0.5.md](../releases/v1.0.5.md) | Version-specific compatibility, migration and changes: Boilerplate v1.0.5. | Preserved record |
