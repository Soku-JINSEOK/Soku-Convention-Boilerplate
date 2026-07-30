# Issue #163 Task Report — Gate-first CI/CD decision planning

## Goal and Background

Issue [#163](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/163)
will add a `soku`-owned decision layer that recommends whether and where CI
should run before any pipeline is generated.

This report approves the public contract and rollout boundaries only.
Implementation remains blocked until the external decision guide is merged and
reviewed public-safe decision, verification, and component contracts are
available on `main`.

## Proposed Approach

Expose separate read-only planning and explicitly confirmed initialization:

```text
soku ci-cd plan --config <path> [--json]
soku ci-cd init --config <path> --dry-run|--yes
```

Model each decision on two independent axes:

- mode: `local-only`, `ci-only`, or `undecided`;
- platform: `github-hosted`, `github-self-hosted`, `gcp-managed`,
  `hybrid-agents`, `cloud-native`, or `kubernetes-gitops`.

Initially, only `ci-only + github-hosted` is installable. All other platform
results are recommendations, and `init` must reject them as unpublished
mappings. Generated CI must be a thin caller of an explicit repository-owned
verification argument vector. Delivery remains disabled.

Future initialization must reuse the manifest-v2 component transaction
framework rather than add another lifecycle state store. The manifest records
only a portable component ID, catalog version, and project-owned configuration
path.

## Planned Implementation

- Merge a reviewed public-safe normative decision contract after its external
  source is approved.
- Add a strict versioned schema and deterministic human and JSON planner output.
- Validate missing or contradictory inputs and preserve `undecided` when a safe
  choice cannot be made.
- Add a read-only planner in a separate PR with personal, GCP, desktop, GitLab,
  and Kubernetes fixtures.
- Add only the GitHub-hosted CI-only installer in a later separate PR.
- Reuse manifest-v2 collision, journal, backup, manifest-last, rollback, and
  ownership behavior for initialization.
- Generate a thin CI caller for the configured repository-owned verification
  command without adding delivery behavior.
- Trial remaining recommendation classes downstream before publishing more
  mappings.

## Acceptance Criteria

- Planning is deterministic, repository-name independent, and performs no
  filesystem, repository-setting, cloud, secret, runner, or environment
  mutation.
- Human and JSON output distinguish the mode and platform axes and explain
  missing inputs, reasons, and rejected alternatives.
- Repeated planning produces byte-stable JSON and leaves the worktree unchanged.
- Initialization supports dry-run and requires explicit `--yes` to write.
- Only `ci-only + github-hosted` can be installed initially; every unpublished
  mapping is refused.
- Initialization refuses unmanaged collisions and completely restores files and
  the exact manifest after failure.
- Generated CI invokes an explicit repository-owned verification argument
  vector and does not duplicate repository verification logic.
- Lifecycle state contains only the portable component ID, catalog version, and
  project-owned configuration path.
- Delivery remains disabled, and no cloud resource, credential, secret, runner,
  environment, or repository setting is created or changed.
- Contract, planner, and installer changes are delivered in separate PRs, all
  marked `Related to #163`.
- Issue #163 remains open until all mappings, downstream trials, and
  delivery-disabled acceptance criteria are complete.

## Approval

- **Status:** `Approved`
- **Approved by:** `Soku-JINSEOK` (approved on 2026-07-31)

## Implementation Status

Task-report publication is in progress. Contract and CLI implementation are
blocked until the external decision guide is merged and the required public-safe
decision, verification, and manifest-v2 component contracts are available on
`main`.

After those gates clear, contract, read-only planner, and GitHub-hosted CI-only
installer work will proceed as separate PRs. No implementation or delivery
activation is authorized by this report PR.

## Verification

- Targeted Markdown lint for this report passed.
- `git diff --check` passed.
- Public disclosure review passed.
- Hosted Validation remains required on the task-report PR.

Later planner verification will cover strict schema fixtures, missing and
contradictory inputs, deterministic human and JSON output, repeated no-write
execution, repository-name independence, and the required personal, GCP,
desktop, GitLab, and Kubernetes scenarios.

Later installer verification will cover dry-run immutability, explicit `--yes`,
unmanaged collision refusal, rollback and exact manifest restoration,
unsupported mapping refusal, thin verification caller generation, and zero
delivery or cloud mutation.

Final implementation gates include `go test ./...`, race and lifecycle tests,
five-platform packaging, security scanning, and hosted cross-platform
validation. Docker-dependent local skips must remain explicitly distinguished
from passes.

## Public Disclosure Review

- [x] No credentials, tokens, private keys, or credential-bearing URLs
- [x] No private repository, project, or product names
- [x] No cloud project IDs, account numbers, service URLs, image URIs, or
      revision identifiers
- [x] No personal billing, subscription, budget, or payment-status information
- [x] No personal email, phone, address, or local absolute path
- [x] No private Issue, PR, Project, or control-plane identifiers

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex

---

## 목표 및 배경

Issue [#163](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/163)는
pipeline을 생성하기 전에 CI가 필요한지와 어디에서 실행할지를 추천하는
`soku` 소유 decision layer를 추가합니다.

이 보고서는 공개 contract와 rollout 경계만 승인합니다. 외부 decision guide가
병합되고 검토된 public-safe decision, verification, component contract가 `main`에
준비될 때까지 구현은 blocked 상태를 유지합니다.

## 제안하는 접근

읽기 전용 planning과 명시적으로 확인된 initialization을 분리합니다.

```text
soku ci-cd plan --config <path> [--json]
soku ci-cd init --config <path> --dry-run|--yes
```

각 decision은 독립된 두 축으로 구성합니다.

- mode: `local-only`, `ci-only`, `undecided`
- platform: `github-hosted`, `github-self-hosted`, `gcp-managed`,
  `hybrid-agents`, `cloud-native`, `kubernetes-gitops`

초기에는 `ci-only + github-hosted`만 설치할 수 있습니다. 다른 platform 결과는
recommendation 전용이며 `init`은 unpublished mapping으로 거부해야 합니다.
생성된 CI는 명시적인 repository-owned verification argument vector를 호출하는
thin caller여야 하며 delivery는 비활성 상태를 유지합니다.

향후 initialization은 별도 lifecycle state store를 추가하지 않고 manifest-v2
component transaction framework를 재사용해야 합니다. Manifest에는 portable
component ID, catalog version, project-owned config path만 기록합니다.

## 계획된 구현

- 외부 source 승인 후 검토된 public-safe normative decision contract를 병합합니다.
- Strict versioned schema와 deterministic human·JSON planner output을 추가합니다.
- 누락되거나 모순된 입력을 검증하고 안전한 선택이 불가능하면 `undecided`를
  유지합니다.
- Personal, GCP, desktop, GitLab, Kubernetes fixture를 포함한 read-only planner를
  별도 PR로 추가합니다.
- 이후 별도 PR에서는 GitHub-hosted CI-only installer만 추가합니다.
- Initialization에 manifest-v2 collision, journal, backup, manifest-last,
  rollback, ownership 동작을 재사용합니다.
- Delivery 동작 없이 설정된 repository-owned verification command를 호출하는
  thin CI caller를 생성합니다.
- 추가 mapping 공개 전 나머지 recommendation class를 downstream에서 시험합니다.

## 수용 기준

- Planning은 deterministic하고 repository 이름에 의존하지 않으며 filesystem,
  repository setting, cloud, secret, runner, environment를 변경하지 않습니다.
- Human·JSON output은 mode와 platform 축을 구분하고 누락 입력, 선택 근거,
  제외한 대안을 설명합니다.
- 반복 planning은 byte-stable JSON을 만들고 worktree를 변경하지 않습니다.
- Initialization은 dry-run을 지원하고 write에는 명시적인 `--yes`가 필요합니다.
- 초기에는 `ci-only + github-hosted`만 설치하며 unpublished mapping은 모두
  거부합니다.
- Initialization은 unmanaged collision을 거부하고 실패 시 file과 exact
  manifest를 완전히 복원합니다.
- 생성된 CI는 명시적인 repository-owned verification argument vector를 호출하고
  repository verification logic을 복제하지 않습니다.
- Lifecycle state에는 portable component ID, catalog version, project-owned
  config path만 포함합니다.
- Delivery는 비활성 상태이며 cloud resource, credential, secret, runner,
  environment, repository setting을 만들거나 변경하지 않습니다.
- Contract, planner, installer 변경은 각각 별도 PR로 제공하고 모두
  `Related to #163`으로 표시합니다.
- 모든 mapping, downstream trial, delivery-disabled 수용 기준을 완료할 때까지
  Issue #163을 open으로 유지합니다.

## 승인

- **상태:** `Approved`
- **승인자:** `Soku-JINSEOK` (2026-07-31 승인)

## 구현 현황

Task report 공개를 진행 중입니다. 외부 decision guide가 병합되고 필요한
public-safe decision, verification, manifest-v2 component contract가 `main`에
준비될 때까지 contract와 CLI 구현은 blocked 상태입니다.

Gate가 해제되면 contract, read-only planner, GitHub-hosted CI-only installer를
각각 별도 PR로 진행합니다. 이 task-report PR은 구현이나 delivery 활성화를
승인하지 않습니다.

## 검증

- 이 report 대상 Markdown lint를 통과했습니다.
- `git diff --check`를 통과했습니다.
- 공개 적합성 검토를 통과했습니다.
- Task-report PR의 hosted Validation은 필수 gate로 남아 있습니다.

향후 planner 검증은 strict schema fixture, 누락·모순 입력, deterministic
human·JSON output, 반복 no-write 실행, repository-name independence와 필수
personal, GCP, desktop, GitLab, Kubernetes scenario를 포함합니다.

향후 installer 검증은 dry-run immutability, explicit `--yes`, unmanaged
collision 거부, rollback과 exact manifest 복원, unsupported mapping 거부, thin
verification caller 생성, delivery·cloud mutation 0건을 포함합니다.

최종 구현 gate에는 `go test ./...`, race·lifecycle test, 5개 platform packaging,
security scan, hosted cross-platform validation이 포함됩니다. Docker 의존 local
skip은 pass와 명시적으로 구분해야 합니다.

## 공개 적합성 검토

- [x] credential, token, private key, credential이 포함된 URL이 없음
- [x] 비공개 저장소·프로젝트·제품 이름이 없음
- [x] cloud project ID, 계정 번호, service URL, image URI, revision 식별자가 없음
- [x] 개인 청구·구독·budget·결제 상태 정보가 없음
- [x] 개인 이메일·전화번호·주소·로컬 절대 경로가 없음
- [x] 비공개 Issue·PR·Project·control-plane 식별자가 없음

## AI 지원

- **계획/구현/초안 작성:** OpenAI Codex
