# Issue #87 Task Report — Confirmed GCP Bootstrap

## Goal and Background

Issue [#87](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/87)
tracks a reproducible path from one GCP project ID to foundation resources,
GitHub OIDC variables, and the first manual Cloud Run deployment.

## Proposed Approach

Keep every default path cloud-free. Require both `--apply` and an exact project
ID confirmation for bootstrap mutations, separate image-independent foundation
Terraform from digest-required runtime creation, and expose deployment only
through an explicitly selected manual workflow operation.

## Planned Implementation

- Add a dry-run-first `scripts/gcp-bootstrap.sh` command.
- Use partial GCS backend configuration and two Terraform stages.
- Replace automatic delivery with manual `check`, `deploy`, and `rollback` jobs.
- Add mock/static regressions and document bootstrap, deployment, and recovery.

## Acceptance Criteria

- [x] Dry-run invokes no cloud, container, Terraform, or GitHub command.
- [x] `GCP_PROJECT_ID` in the CLI environment is accepted, while an explicit
  `--project-id` takes precedence.
- [x] Apply refuses missing or mismatched project confirmation.
- [x] Foundation validates without an image and runtime requires a digest URI.
- [x] State bucket creation and six GitHub Variable writes are repeatable.
- [x] The default workflow check has no authentication or cloud mutation path.
- [x] The first manual dev deployment and rollback are documented.

## Approval

- **Status:** `Approved`
- **Approved by:** User-provided implementation plan

## Implementation Status

Implemented and applied to `ci-cd-control-plane` with the documented defaults.
The GCS backend, foundation resources, bootstrap image, private Cloud Run
runtime, WIF connection, and six Repository Variables are active.

## Verification

- [x] Bash syntax and eight Node mock/workflow regression tests pass, including
  the complete mocked apply sequence.
- [x] Terraform 1.8.5 format check and provider-backed validation pass.
- [x] Repository Node tests, Go tests, release-tag tests, Python provider action
  tests, and whitespace checks pass.
- [x] Live bootstrap completed; Cloud Run revision
  `soku-convention-boilerplate-00001-ltw` is `Ready=True` and authenticated
  `/health` returns `ok`.

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex

---

## 목표 및 배경

Issue #87은 하나의 GCP 프로젝트 ID에서 foundation 리소스, GitHub OIDC
변수, 최초 수동 Cloud Run 배포까지 재현 가능한 경로를 추적합니다.

## 제안하는 접근

기본 경로는 항상 클라우드 변경 없이 유지합니다. 실제 bootstrap은
`--apply`와 정확한 프로젝트 ID 재확인을 모두 요구하고, Terraform은 이미지가
필요 없는 foundation과 digest가 필수인 runtime 단계로 분리합니다.

## 계획된 구현

- dry-run 기본 bootstrap 스크립트 추가
- 부분 GCS backend 설정과 2단계 Terraform 적용
- 수동 `check`, `deploy`, `rollback` 워크플로 구성
- 모의/정적 회귀 테스트와 운영 문서 추가

## 수용 기준

- [x] dry-run은 외부 명령을 호출하지 않습니다.
- [x] CLI 환경의 `GCP_PROJECT_ID`가 반영되고 명시적 인자가 우선합니다.
- [x] 실제 적용은 프로젝트 ID 확인이 일치해야 합니다.
- [x] foundation과 runtime 이미지 계약이 분리됩니다.
- [x] 최초 dev 배포 및 복구 절차가 문서화됩니다.

## 승인

- **상태:** `Approved`
- **승인자:** 사용자가 제공한 구현 계획

## 구현 현황

구현과 `ci-cd-control-plane` 실제 적용을 완료했습니다. GCS state,
foundation, 비공개 Cloud Run, WIF, GitHub Variables 6개가 활성화됐습니다.

## 검증

- [x] 셸, Node, Terraform, Go, Python 및 공백 검사를 통과했습니다.
- [x] Cloud Run `Ready=True`와 인증된 `/health` 응답 `ok`를 확인했습니다.

## AI 지원

- **계획/구현/초안 작성:** OpenAI Codex
