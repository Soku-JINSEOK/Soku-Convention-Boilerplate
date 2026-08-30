# Issue #163 Task Report — Gate-first CI/CD decision planning

## Goal and Background

Issue [#163](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/163)
adds a `soku`-owned decision layer that recommends whether and where CI should
run before any pipeline is generated.

The public contract and rollout boundaries were approved first. The contract,
immutable adapter catalog, and transactional installer are now delivered as
the three reviewable changes recorded in the implementation amendment below.

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

Exactly three mappings are installable: `ci-only + github-hosted`, `ci-only +
gcp-managed`, and `ci-only + github-self-hosted`. Other platform results remain
recommendations and `init` rejects them as unpublished mappings. Generated CI is
a thin caller of an explicit repository-owned verification argument vector.
Delivery remains disabled.

Initialization reuses the existing manifest component transaction framework
rather than adding another lifecycle state store. The manifest records the
portable component ID, binding catalog value, repository-relative configuration
path, and managed caller file state; the binding and baseline hashes are kept
in those existing component/file fields.

## Planned Implementation

- Merge a reviewed public-safe normative decision contract after its external
  source is approved.
- Add a strict versioned schema and deterministic human and JSON planner output.
- Validate missing or contradictory inputs and preserve `undecided` when a safe
  choice cannot be made.
- Add a read-only planner in a separate PR with personal, GCP, desktop, GitLab,
  and Kubernetes fixtures.
- Add the three reviewed validation-only mappings and their trusted renderers.
- Add the transactional installer in a later separate PR with no interactive
  fallback.
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
- Only the three catalog mappings can be installed; every unpublished mapping
  is refused.
- Initialization refuses unmanaged collisions and completely restores files and
  the exact manifest after failure.
- Generated CI invokes an explicit repository-owned verification argument
  vector and does not duplicate repository verification logic.
- Lifecycle state contains only portable component/configuration data and
  managed caller binding/baseline hashes; it contains no raw configuration,
  command output, inventory, credentials, endpoint, or absolute path.
- Delivery remains disabled, and no cloud resource, credential, secret, runner,
  environment, or repository setting is created or changed.
- Contract, planner, adapter conformance, and installer changes are delivered in
  separate reviewable PRs as recorded in the implementation amendment.
- Issue #163 remains open until all mappings, downstream trials, and
  delivery-disabled acceptance criteria are complete.

## Approval

- **Status:** `Approved`
- **Approved by:** `Soku-JINSEOK` (approved on 2026-07-31)

## Implementation Status

PR A (#212) merged the strict decision contract and deterministic planner. PR B
(#214) merged immutable engine provenance, the exact three validation-only
mappings, trusted renderers, and semantic conformance. PR C (#215) completes
transactional `soku ci-cd init`: exactly one of `--dry-run` or `--yes` is
required, the plan and all write identities are recomputed at the write
boundary, and the shared manifest-last transaction provides backup, journal,
atomic write, rollback, and recovery behavior.

The installer requires an existing initialized manifest, migrates v1 to v2
without downgrading v2 or v3, rejects unmanaged collisions and stale component
state, and records only repository-relative paths and immutable hashes. All
three mappings remain validation-only and delivery-disabled.

## Verification

- Targeted Markdown lint for this report passed.
- `git diff --check` passed.
- Public disclosure review passed.
- Planner, adapter conformance, installer, rollback, and manifest compatibility
  tests are covered by the implementation PRs.

Planner verification covers strict schema fixtures, missing and contradictory
inputs, deterministic human and JSON output, repeated no-write execution,
repository-name independence, and the required personal, GCP, desktop, GitLab,
and Kubernetes scenarios. Installer verification covers dry-run immutability,
the exact confirmation flags, unmanaged collision refusal, rollback and exact
manifest restoration, unsupported mapping refusal, thin caller generation,
stale identity, and zero delivery or Cloud mutation.

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
준비된 뒤 contract, planner, adapter conformance, transactional installer를
세 개의 reviewable change로 완료했습니다.

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

`ci-only + github-hosted`, `ci-only + gcp-managed`, `ci-only +
github-self-hosted` 세 mapping만 설치할 수 있습니다. 다른 platform 결과는
recommendation 전용이며 `init`은 unpublished mapping으로 거부해야 합니다.
생성된 CI는 명시적인 repository-owned verification argument vector를 호출하는
thin caller여야 하며 delivery는 비활성 상태를 유지합니다.

Initialization은 별도 lifecycle state store를 추가하지 않고 기존 manifest
component transaction framework를 재사용합니다. Manifest에는 portable
component/configuration 정보와 managed caller의 binding·baseline hash만
기존 component/file state로 기록합니다.

## 계획된 구현

- 외부 source 승인 후 검토된 public-safe normative decision contract를 병합합니다.
- Strict versioned schema와 deterministic human·JSON planner output을 추가합니다.
- 누락되거나 모순된 입력을 검증하고 안전한 선택이 불가능하면 `undecided`를
  유지합니다.
- Personal, GCP, desktop, GitLab, Kubernetes fixture를 포함한 read-only planner를
  별도 PR로 추가합니다.
- 세 개의 reviewed validation-only mapping과 trusted renderer를 추가합니다.
- 이후 별도 PR에서 interactive fallback 없는 transactional installer를
  추가합니다.
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
- catalog의 세 mapping만 설치하며 unpublished mapping은 모두 거부합니다.
- Initialization은 unmanaged collision을 거부하고 실패 시 file과 exact
  manifest를 완전히 복원합니다.
- 생성된 CI는 명시적인 repository-owned verification argument vector를 호출하고
  repository verification logic을 복제하지 않습니다.
- Lifecycle state에는 portable component/configuration 정보와 managed caller
  binding·baseline hash만 포함하며 raw config, command output, inventory,
  credential, endpoint, absolute path는 포함하지 않습니다.
- Delivery는 비활성 상태이며 cloud resource, credential, secret, runner,
  environment, repository setting을 만들거나 변경하지 않습니다.
- Contract, planner, adapter conformance, installer 변경은 별도 reviewable PR로
  제공하며 implementation amendment에 기록합니다.
- 모든 mapping, downstream trial, delivery-disabled 수용 기준을 완료할 때까지
  Issue #163을 open으로 유지합니다.

## 승인

- **상태:** `Approved`
- **승인자:** `Soku-JINSEOK` (2026-07-31 승인)

## 구현 현황

PR A (#212)가 strict decision contract와 deterministic planner를 병합했고,
PR B (#214)가 immutable provenance, 세 validation-only mapping, trusted
renderer와 semantic conformance를 병합했습니다. PR C (#215)는 정확히 하나의
`--dry-run` 또는 `--yes`를 요구하는 transactional `soku ci-cd init`을
완료합니다. Write 직전에 plan·worktree·catalog·renderer·manifest identity를
다시 계산하고 기존 manifest-last transaction의 backup, journal, atomic write,
rollback, recovery를 사용합니다.

Installer는 이미 초기화된 manifest를 요구하며 v1은 v2로 migration하고 v2와
v3는 downgrade하지 않습니다. Unmanaged collision과 stale component state를
거부하고 repository-relative path와 immutable hash만 기록합니다. 세 mapping은
모두 validation-only이며 delivery-disabled 상태입니다.

## 검증

- 이 report 대상 Markdown lint를 통과했습니다.
- `git diff --check`를 통과했습니다.
- 공개 적합성 검토를 통과했습니다.
- Implementation PR들의 planner, adapter conformance, installer, rollback,
  manifest compatibility test가 해당 기준을 검증합니다.

Planner 검증은 strict schema fixture, 누락·모순 입력, deterministic human·JSON
output, 반복 no-write 실행, repository-name independence와 필수 personal,
GCP, desktop, GitLab, Kubernetes scenario를 포함합니다. Installer 검증은
dry-run immutability, 정확한 confirmation flag, unmanaged collision 거부,
rollback과 exact manifest 복원, unsupported mapping 거부, thin caller 생성,
stale identity, delivery·Cloud mutation 0건을 포함합니다.

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
