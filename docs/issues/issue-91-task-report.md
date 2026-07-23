# Issue #91 Task Report — Restore authenticated Cloud Run deployment evidence

## Goal and Background

Issue [#91](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/91)
tracks two consecutive failures in the first `dev` deployment. The first run
could not resolve the pushed container digest from Docker's local metadata. PR
[#92](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/pull/92)
fixed that planning failure by querying Artifact Registry first and was merged.
The next run reached the private Cloud Run health check but could not mint an
audience-bound ID token from the WIF credential without explicitly
impersonating the configured deployer service account. Its failure evidence was
also written under hidden `.cd/`, which `upload-artifact` does not include by
default.

The two hosted failures are preserved as
[run 29937165884](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/29937165884)
and
[run 29969347852](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/29969347852).

## Proposed Approach

Keep the existing WIF and least-privilege IAM model. Add an optional deployer
identity argument to `cd-deploy.sh`, use it only for service-account
impersonation when minting the private health-check token, and retain the active
account behavior when the argument is absent. Move sanitized JSON evidence to a
non-hidden directory and require both deploy and rollback workflow paths to
upload it.

## Planned Implementation

- Add `--identity-service-account <email>` with Google service-account email
  validation and backward-compatible omission behavior.
- Mint the ID token with both the configured service account and the exact
  Cloud Run service URL as its audience.
- Pass `GCP_WIF_SERVICE_ACCOUNT` explicitly from deploy and rollback jobs.
- Write evidence under `deploy-evidence/`, ignore local evidence in Git, and
  fail artifact upload when the expected evidence is missing.
- Cover successful deploy, automatic recovery, manual rollback, token failure,
  invalid identity, and workflow evidence behavior with regression tests.
- Do not change Terraform, IAM roles, staging or production environments,
  release tags, or credential persistence.

## Acceptance Criteria

- A private `/health` request uses an impersonated, audience-bound identity
  token when the service-account option is supplied.
- Omitting the new option preserves the existing active-account path.
- Token acquisition failure makes the new revision unhealthy, restores traffic
  to the exact pre-deploy revision, and records the final recovery result.
- Successful deploy, automatic recovery, manual rollback, and recovery failure
  all create sanitized JSON evidence.
- Deploy and rollback upload non-hidden evidence and fail when it is absent.
- The hosted `check` operation and full pull-request validation pass before a
  reviewed `dev` deployment is attempted.

## Approval

- **Status:** `Approved`
- **Approved by:** User (approved the implementation plan)

## Implementation Status

PR #92 resolved the initial digest failure. The follow-up authentication and
evidence changes are implemented locally and await hosted validation and the
explicit `dev` deployment verification.

## Verification

- `bash -n scripts/gcp-bootstrap.sh scripts/cd-plan.sh scripts/cd-deploy.sh scripts/ci-local.sh` — passed
- `node --test .github/deploy-gcp.test.mjs` — passed
- Full Node governance/deployment regression suite — 80 tests passed
- `git diff --check` — passed
- Hosted `operation=check`, pull-request gates, and `dev` deployment — pending

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex

---

## 목표 및 배경

Issue [#91](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/91)은
첫 `dev` 배포에서 연속으로 발생한 두 실패를 다룹니다. 첫 실행은 Docker 로컬
메타데이터에서 푸시된 이미지 digest를 찾지 못했습니다. PR
[#92](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/pull/92)는 Artifact
Registry 우선 조회로 이 문제를 수정해 병합되었습니다. 다음 실행은 비공개 Cloud
Run 상태 확인까지 도달했지만, WIF credential만 활성화된 상태에서 구성된 deployer
service account를 명시적으로 impersonate하지 않아 audience-bound ID token을 발급하지
못했습니다. 또한 실패 증거가 숨김 `.cd/` 아래에 있어 `upload-artifact`의 기본
업로드 대상에서 제외되었습니다.

두 hosted 실패 실행은
[run 29937165884](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/29937165884)와
[run 29969347852](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/29969347852)에
보존되어 있습니다.

## 제안하는 접근

기존 WIF와 최소 권한 IAM 구성은 유지합니다. `cd-deploy.sh`에 선택적 deployer
identity 인자를 추가해 비공개 상태 확인 토큰 발급에만 service-account
impersonation을 사용하고, 인자를 생략하면 기존 활성 계정 동작을 유지합니다. 민감
정보가 없는 JSON 증거는 비숨김 디렉터리로 옮기고 deploy와 rollback 양쪽이 반드시
업로드하도록 합니다.

## 계획된 구현

- Google service-account email을 검증하는
  `--identity-service-account <email>` 추가 및 생략 시 하위 호환성 유지
- 구성된 service account와 정확한 Cloud Run URL audience로 ID token 발급
- deploy와 rollback job에서 `GCP_WIF_SERVICE_ACCOUNT` 명시 전달
- 증거를 `deploy-evidence/`에 기록하고 로컬 파일은 Git에서 제외하며, 누락 시 artifact
  업로드 실패
- 성공 배포, 자동 복구, 수동 rollback, token 실패, 잘못된 identity, workflow 증거
  동작 회귀 테스트
- Terraform, IAM role, staging/prod, release tag, credential 저장은 변경하지 않음

## 수용 기준

- service-account 옵션 사용 시 비공개 `/health` 요청이 impersonation과 정확한
  audience를 사용하는 ID token으로 인증됩니다.
- 옵션을 생략하면 기존 활성 계정 경로가 유지됩니다.
- token 발급 실패는 새 revision을 비정상으로 처리하고 정확한 배포 전 revision으로
  traffic을 복구한 뒤 최종 결과를 기록합니다.
- 성공, 자동 복구, 수동 rollback, 복구 실패가 모두 민감 정보 없는 JSON 증거를
  생성합니다.
- deploy와 rollback이 비숨김 증거를 업로드하며 누락은 실패 처리됩니다.
- 검토된 `dev` 배포 전에 hosted `check`와 전체 PR validation이 통과합니다.

## 승인

- **상태:** `Approved`
- **승인자:** 사용자 (구현 계획 승인)

## 구현 현황

PR #92는 최초 digest 실패를 해결했습니다. 후속 인증·증거 수정은 로컬 구현을
마쳤으며 hosted validation과 명시적 `dev` 배포 검증을 기다립니다.

## 검증

- `bash -n scripts/gcp-bootstrap.sh scripts/cd-plan.sh scripts/cd-deploy.sh scripts/ci-local.sh` — 통과
- `node --test .github/deploy-gcp.test.mjs` — 통과
- 전체 Node governance/deployment 회귀 테스트 — 80개 통과
- `git diff --check` — 통과
- hosted `operation=check`, PR gate, `dev` 배포 — 대기

## AI 지원

- **계획/구현/초안 작성:** OpenAI Codex
