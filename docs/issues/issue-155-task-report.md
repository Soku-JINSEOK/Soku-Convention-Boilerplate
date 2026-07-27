# Issue #155 Task Report — Add low-cost GCP sandbox guardrails

## Goal and Background

Issue [#155](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/155)
adds bounded cost controls for the opt-in GCP sandbox without changing the
active CI observation structure or applying cloud changes.

The implementation was prepared from the explicitly authorized remediation
plan before the tracking Issue existed. This report records that sequencing
exception transparently. Normal review, hosted validation, dry-run observation,
and live-apply approval remain required.

## Proposed Approach

Keep reusable defaults general while exposing explicit low-cost sandbox inputs.
Create budget alerts only when opted in, begin Artifact Registry cleanup in
dry-run mode, retain recent and protected images, and constrain state cleanup to
old noncurrent versions. Keep billing identity private and cloud identifiers
out of public evidence.

## Planned Implementation

- Add sensitive, opt-in budget inputs and a project-scoped budget with 50%, 80%,
  and 100% current-spend thresholds.
- Add configurable Artifact Registry delete and keep policies with dry-run
  enabled by default.
- Apply a state-bucket lifecycle that retains at least 30 days and 10 newer
  noncurrent versions.
- Add sandbox bootstrap inputs for one maximum instance and reviewed cleanup
  activation.
- Limit validation-only Cloud Build to 15 minutes on the default machine.
- Document latest-head-SHA failure classification.
- Add Terraform, bootstrap, and Cloud Build regression tests.

## Acceptance Criteria

- Budget creation is opt-in, uses a sensitive billing account input, and never
  disables billing automatically.
- Artifact cleanup remains dry-run until seven days of logs are reviewed.
- Untagged images retain seven days, ordinary commit images retain 30 days, the
  newest five versions remain, and release/protected tags remain.
- Current state objects are never lifecycle deletion candidates; noncurrent
  objects need both 30 days of age and at least 10 newer versions.
- The reusable Cloud Run default remains three maximum instances, while the
  sandbox apply can explicitly select one.
- Cloud Build uses its default machine with a 15-minute overall timeout.
- Terraform validation and mock tests pass.
- Current-tree and full-history secret scans report zero findings.
- No runtime, IAM, bucket deletion, ruleset, or CI observation change is applied
  from this task.

## Approval

- **Status:** `Approved`
- **Approved by:** `Soku-JINSEOK` (authorized plan implementation and
  continuation)

## Implementation Status

Implementation is published in Draft PR
[#156](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/pull/156).
Live GCP discovery and apply are blocked by expired authentication and remain
intentionally unperformed. Cleanup activation also remains blocked on the
required seven-day dry-run review.

## Verification

- Terraform 1.15.3 `fmt -check`, `validate`, and `test` with Google provider
  6.50.0 (4/4 mock plans passing).
- `node --test .github/cloudbuild-validation.test.mjs
  .github/deploy-gcp.test.mjs` (passing).
- ShellCheck 0.11.0 for `scripts/gcp-bootstrap.sh` (passing).
- Gitleaks 8.30.1 current-tree scan (0 findings).
- Gitleaks 8.30.1 full-history scan (0 findings).
- `git diff --check` (passing).
- Full profile attempted with `--skip-db`; repository regression tests that
  invoke the scope detector from temporary workspaces fail under local Node
  26.5.0 before this change's focused checks.
- Draft PR #156 hosted Node 24 Quick, Full, runtime-template, security, CodeQL,
  Terraform, metadata, and validation-only Cloud Build checks (passing).

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

Issue [#155](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/155)는
현재 CI 관찰 구조를 바꾸거나 cloud 변경을 적용하지 않으면서 opt-in GCP
sandbox에 제한된 비용 제어를 추가합니다.

전용 Issue가 만들어지기 전에 명시적으로 승인된 정비 계획에 따라 구현 초안이
작성되었습니다. 이 보고서는 그 순서 예외를 투명하게 기록합니다. 일반 review,
hosted validation, dry-run 관찰과 live apply 승인은 그대로 필요합니다.

## 제안하는 접근

재사용 기본값은 범용으로 유지하고 저비용 sandbox 입력은 명시적으로 제공합니다.
예산 알림은 opt-in으로만 생성하고 Artifact Registry 정리는 dry-run으로 시작하며
최근·보호 이미지를 보존합니다. state 정리는 오래된 noncurrent version으로
제한하고 billing 식별자와 cloud 식별자는 공개 증거에서 제외합니다.

## 계획된 구현

- 민감한 opt-in budget 입력과 현재 지출 50%, 80%, 100% 임계값을 추가합니다.
- 기본 dry-run인 Artifact Registry 삭제·보존 정책을 추가합니다.
- 최소 30일과 newer version 10개를 함께 보장하는 state lifecycle을 적용합니다.
- sandbox 최대 instance 1개와 검토된 cleanup 활성화 입력을 추가합니다.
- 기본 머신의 validation-only Cloud Build timeout을 15분으로 제한합니다.
- 최신 head SHA 기반 실패 분류를 문서화합니다.
- Terraform, bootstrap, Cloud Build 회귀 테스트를 추가합니다.

## 수용 기준

- Budget은 opt-in이며 민감 billing account 입력을 사용하고 billing을 자동으로
  비활성화하지 않습니다.
- Artifact cleanup은 7일간 로그를 검토하기 전까지 dry-run입니다.
- Untagged image 7일, 일반 commit image 30일, 최신 5개와 release/protected tag를
  보존합니다.
- Current state object는 삭제 대상이 아니며 noncurrent object는 30일과 newer
  version 10개 조건을 모두 충족해야 합니다.
- 범용 Cloud Run 기본값은 최대 3개를 유지하고 sandbox apply는 1개를 명시합니다.
- Cloud Build는 기본 머신과 전체 timeout 15분을 사용합니다.
- Terraform validation과 mock test가 통과합니다.
- 현재 tree와 전체 history secret scan 결과가 0건입니다.
- Runtime, IAM, bucket 삭제, ruleset, CI 관찰 변경을 이 작업에서 적용하지 않습니다.

## 승인

- **상태:** `Approved`
- **승인자:** `Soku-JINSEOK` (계획 구현과 작업 계속 진행 승인)

## 구현 현황

구현은 Draft PR
[#156](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/pull/156)에
게시했습니다. 실제 GCP 조회와 apply는 만료된 인증으로 차단되어 의도적으로
수행하지 않았습니다. Cleanup 활성화도 필수 7일 dry-run 검토 전까지 차단됩니다.

## 검증

- Terraform 1.15.3 `fmt -check`, `validate`, `test`, Google provider 6.50.0
  (mock plan 4/4 통과).
- `node --test .github/cloudbuild-validation.test.mjs
  .github/deploy-gcp.test.mjs` 통과.
- ShellCheck 0.11.0 `scripts/gcp-bootstrap.sh` 통과.
- Gitleaks 8.30.1 현재 tree 0건.
- Gitleaks 8.30.1 전체 history 0건.
- `git diff --check` 통과.
- `--skip-db` full profile을 시도했으며 임시 workspace에서 scope detector를
  호출하는 기존 repository regression test가 로컬 Node 26.5.0에서 focused
  check 전에 실패했습니다.
- Draft PR #156 hosted Node 24 Quick, Full, runtime template, security, CodeQL,
  Terraform, metadata, validation-only Cloud Build check 통과.

## 공개 적합성 검토

- [x] credential, token, private key, credential이 포함된 URL이 없음
- [x] 비공개 저장소·프로젝트·제품 이름이 없음
- [x] cloud project ID, 계정 번호, service URL, image URI, revision 식별자가 없음
- [x] 개인 청구·구독·budget·결제 상태 정보가 없음
- [x] 개인 이메일·전화번호·주소·로컬 절대 경로가 없음
- [x] 비공개 Issue·PR·Project·control-plane 식별자가 없음

## AI 지원

- **계획/구현/초안 작성:** OpenAI Codex
