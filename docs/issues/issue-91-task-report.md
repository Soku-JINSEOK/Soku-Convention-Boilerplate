# Issue #91 Task Report — Resolve pushed image digest from Artifact Registry

## Goal and Background

Issue [#91](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/91) reports that `scripts/cd-plan.sh` fails on GitHub-hosted runners because `docker inspect` does not always expose `.RepoDigests` immediately after push. The script exits with code `11` and stops deployments in `cd-plan`.

## Proposed Approach

Keep `cd-plan` behavior and outputs unchanged where possible, and resolve the immutable image digest by querying Artifact Registry as the primary source after push. If that fails, fallback to the existing `docker inspect` logic so local/docker-only flows remain supported.

## Planned Implementation

- Update `scripts/cd-plan.sh` push path:
  - query `gcloud artifacts docker images describe ... --format='value(image_summary.fully_qualified_digest)'`
  - accept the digest only when it is a fully qualified digest for the pushed image tag
  - fallback to current `docker inspect` RepoDigests lookup when Registry query is unavailable
  - keep current exit code (`11`) when neither path returns a digest
- Update `.github/deploy-gcp.test.mjs` regression coverage:
  - Registry-first digest success path
  - Repository fallback success path
  - both unavailable (exit `11`) failure path

## Acceptance Criteria

- `cd-plan --push-image` resolves digest from Artifact Registry in normal GitHub runner paths.
- If Registry lookup is unavailable or malformed, deployment planning still works when `docker` RepoDigests are available.
- Plan failure mode remains unchanged when no digest is obtainable from either source.
- Relevant Node regression tests cover all three paths.

## Approval

- **Status:** `Approved`
- **Approved by:** User

## Implementation Status

Implementation prepared in the current branch for `main`-based workflow.

## Verification

- Not run in this session (per current instruction constraints).

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex

---

## 목표 및 배경

Issue [#91](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/91)은
`scripts/cd-plan.sh`에서 푸시 직후 `docker inspect`의 `.RepoDigests` 값이 비어
실패하는 GitHub-hosted 실행 문제를 다룹니다. 현재 스크립트는 `11` 코드로 종료되어
배포 파이프라인이 중단됩니다.

## 제안하는 접근

`cd-plan`의 기존 동작 형식은 유지하고, 푸시 후 다이제스트 해석은 Artifact
Registry를 1순위로 수행합니다. Registry 경로가 동작하지 않으면 기존 `docker
inspect` 경로로 폴백합니다.

## 계획된 구현

- `scripts/cd-plan.sh` 변경:
  - `gcloud artifacts docker images describe ... --format='value(image_summary.fully_qualified_digest)'`
  - 반환값이 푸시한 이미지의 fully-qualified digest인지 검증
  - 실패 시 기존 `docker inspect` RepoDigests 경로로 폴백
  - 두 경로 모두 실패 시 기존 종료 코드(`11`) 유지
- `.github/deploy-gcp.test.mjs` 회귀 테스트 보강:
  - Registry 우선 성공
  - RepoDigests 폴백 성공
  - 둘 다 실패한 경우(11 종료) 확인

## 수용 기준

- `cd-plan --push-image` 실행 시 GitHub runner에서 Artifact Registry 기반 digest 해석 우선 사용
- Registry 조회 실패/미지원 시에도 `docker` 메타데이터가 있으면 플랜 작성 성공
- 두 경로 모두 실패하면 기존 실패 모드 유지
- 세 경로를 커버하는 Node 회귀 테스트 반영

## 승인

- **상태:** `Approved`
- **승인자:** 사용자

## 구현 현황

현재 브랜치에서 구현 반영 완료 (실행 검증 대기).

## 검증

- 본 작업 세션에서는 실행 검증 미실시

## AI 지원

- **계획/구현/초안 작성:** OpenAI Codex
