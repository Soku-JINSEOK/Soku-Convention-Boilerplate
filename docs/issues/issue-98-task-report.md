# Issue #98 Task Report — Add an end-to-end boilerplate manual

## Goal and Background

Issue [#98](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/98)
asks for one human entrypoint that connects adoption level, safe initialization,
governance, optional GCP dev delivery, and lifecycle operation.

## Proposed Approach

Write one English operational manual under the repository Language Policy. Keep
existing lifecycle, initialization, release/sync, governance, verification, and
Cloud Run documents authoritative and link to them for edge cases. Make the new
entrypoint discoverable from all overview languages and protect its public
identifiers with a repository test.

## Planned Implementation

- Add `docs/guides/USAGE_MANUAL.md` with the twelve approved adoption stages.
- Use the published boilerplate `v1.0.5` and CLI `soku/v0.1.4` baselines.
- Match profile and stack IDs to catalog v2/core v1 and command names to the
  actual CLI surface.
- Explain checksum verification, collision refusal, manifest operation,
  rollback/recovery, and secret prohibitions.
- Link README EN/KO/JA, Blueprint, and Applicability to the manual.
- Correct the stale `soku/README.md` recommended baseline and examples.
- Add a documentation contract regression test and include it in repository
  hygiene.

## Acceptance Criteria

- A first-time adopter can progress from profile selection through upgrades from
  one page.
- Commands, releases, profiles, and stack IDs match implementation contracts.
- Manual and automated paths remain distinct and non-destructive.
- Optional cloud delivery is limited to GCP `dev` and sanitized evidence.
- All requested entrypoints link to the manual.
- Documentation tests, Markdown, repository hygiene, and hosted Validation pass.

## Approval

- **Status:** `Approved`
- **Approved by:** User (approved the implementation plan)

## Implementation Status

The dedicated usage-manual branch now contains the manual, discovery links,
baseline correction, documentation contract test, and repository-hygiene
integration. Hosted review remains pending.

## Verification

- Usage-manual contract — 5 tests passed
- Repository contribution-title and PR policy Node suites — passed
- Full GitHub governance/workflow Node suite — passed
- Markdown lint for all changed documentation — passed
- YAML lint for the changed workflow — passed
- `actionlint` for all workflows — passed
- `git diff --check` — passed
- `scripts/ci-local.sh --workspace .` — environment preflight stopped because
  local `shellcheck` is not installed; targeted checks above passed and hosted
  Validation remains required

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex

---

## 목표 및 배경

Issue [#98](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/98)은
적용 수준, 안전한 초기화, governance, 선택적 GCP dev 배포와 lifecycle 운영을 하나로
연결하는 인간 사용자용 시작점을 요구합니다.

## 제안하는 접근

Language Policy에 따라 영어 운영 매뉴얼 하나를 작성합니다. 기존 lifecycle, 초기화,
release/sync, governance, verification, Cloud Run 문서를 권위 문서로 유지하고 예외
상황은 연결합니다. 세 README 언어판에서 매뉴얼을 발견할 수 있게 하고 공개 식별자는
repository 테스트로 보호합니다.

## 계획된 구현

- 승인된 12단계의 `docs/guides/USAGE_MANUAL.md` 추가
- 공개 baseline `v1.0.5`와 `soku/v0.1.4` 사용
- catalog와 실제 CLI의 profile, stack, command 식별자 일치
- checksum, 충돌 거부, manifest 운영, rollback/recovery, secret 금지 설명
- README EN/KO/JA, Blueprint, Applicability 연결
- `soku/README.md`의 오래된 baseline과 예시 수정
- 문서 계약 회귀 테스트를 repository hygiene에 포함

## 수용 기준

- 신규 사용자가 한 페이지에서 profile 선택부터 upgrade까지 수행할 수 있습니다.
- 명령, release, profile, stack ID가 구현 계약과 일치합니다.
- 수동 및 자동 경로가 분리되고 비파괴 조건을 유지합니다.
- 선택적 cloud delivery는 GCP `dev`와 sanitized evidence로 제한됩니다.
- 요청된 모든 진입점이 매뉴얼을 연결합니다.
- 문서 테스트, Markdown, repository hygiene, hosted Validation이 통과합니다.

## 승인

- **상태:** `Approved`
- **승인자:** 사용자 (구현 계획 승인)

## 구현 현황

전용 usage-manual branch에 매뉴얼, 발견 링크, baseline 수정, 문서 계약 테스트와
repository-hygiene 연결을 구현했습니다. Hosted review가 남았습니다.

## 검증

- usage-manual 계약 — 5개 테스트 통과
- repository contribution-title 및 PR policy Node suite — 통과
- 전체 GitHub governance/workflow Node suite — 통과
- 변경 문서 Markdown lint — 통과
- 변경 workflow YAML lint — 통과
- 전체 workflow `actionlint` — 통과
- `git diff --check` — 통과
- `scripts/ci-local.sh --workspace .` — 로컬 `shellcheck` 미설치로 환경 preflight에서
  중단; 위 targeted 검사는 통과했으며 hosted Validation은 계속 필수

## AI 지원

- **계획/구현/초안 작성:** OpenAI Codex
