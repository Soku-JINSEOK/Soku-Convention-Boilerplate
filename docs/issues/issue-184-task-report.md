# 🐛 Issue 184 Task Report

## Goal and Background

Issue [#184](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/184)
removes a stale Dependabot Docker target. Post-merge update job `1498791196`
failed because the configured repository root contains no Dockerfile or
Kubernetes manifest.

## Proposed Approach

Remove only the invalid root Docker entry, preserve both valid Docker targets,
and add a regression test that requires a local manifest for every configured
Docker update directory.

## Planned Implementation

- Remove `docker:/` from Dependabot and supply-chain target contracts.
- Preserve `docker:/.devcontainer` and `docker:/templates/gcloud`.
- Verify configured Docker directories contain a Dockerfile or YAML manifest.

## Acceptance Criteria

- No Docker update targets a directory without a supported manifest.
- Existing ecosystem coverage and immutable supply-chain checks pass.
- No dependency, permission, ruleset, Terraform, IAM, cloud, delivery, or
  deployment mutation is included.

## Approval

- **Status:** `Approved`
- **Approved by:** Soku-JINSEOK on 2026-08-01

## Implementation Status

Implementation is complete locally on the focused Issue #184 branch and awaits
hosted review.

## Verification

- Full Node repository/workflow/policy/supply-chain suites: 140 passed.
- Focused Dependabot and supply-chain suites: 10 passed.
- Supply-chain verification: 36 protected files and 11 valid update targets.
- YAML lint, Markdown lint, and `git diff --check`: passed.
- Hosted verification is pending the Draft pull request.

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

Issue #184는 저장소 root에 Dockerfile 또는 Kubernetes manifest가 없는데도
남아 있던 Dependabot Docker 대상을 제거합니다.

## 제안하는 접근

잘못된 root 항목만 제거하고 두 유효 Docker 대상을 보존하며, 각 Docker update
directory에 실제 manifest가 존재하는지 회귀 테스트로 고정합니다.

## 계획된 구현

- Dependabot 및 supply-chain 계약에서 `docker:/` 제거
- `/.devcontainer`, `/templates/gcloud` 대상 보존
- Docker 대상별 manifest 존재 검사 추가

## 수용 기준

manifest가 없는 Docker update 대상이 없고 기존 공급망 검증이 통과해야 합니다.
dependency, permission, ruleset, Terraform, IAM, cloud, delivery, deployment는
변경하지 않습니다.

## 승인

- **상태:** `Approved`
- **승인자:** Soku-JINSEOK, 2026-08-01

## 구현 현황

Issue #184 전용 branch의 local 구현을 완료했고 hosted review를 기다립니다.

## 검증

- 전체 Node repository/workflow/policy/supply-chain test 140개 통과
- Dependabot 및 supply-chain 집중 test 10개 통과
- supply-chain 검증: 보호 파일 36개, 유효 update target 11개
- YAML lint, Markdown lint, `git diff --check` 통과
- Draft PR의 hosted 검증 대기 중

## 공개 적합성 검토

- [x] credential, token, private key, credential이 포함된 URL이 없음
- [x] 비공개 저장소·프로젝트·제품 이름이 없음
- [x] cloud project ID, 계정 번호, service URL, image URI, revision
      식별자가 없음
- [x] 개인 청구·구독·budget·결제 상태 정보가 없음
- [x] 개인 이메일·전화번호·주소·로컬 절대 경로가 없음
- [x] 비공개 Issue·PR·Project·control-plane 식별자가 없음

## AI 지원

- **계획/구현/초안 작성:** OpenAI Codex
