# Issue #93 Task Report — Make Dependabot governance deterministic

## Goal and Background

Issue [#93](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/93)
tracks duplicate PR workflow runs and human-only governance failures on
Dependabot PRs #89 and #90.

## Proposed Approach

Retain one event-subscribing Validation workflow, convert title and PR policy
workflows into reusable components, and converge duplicate JavaScript rules on
shared modules. Grant a body/title exception only to an exact, file-scoped
Dependabot identity while preserving every required check and metadata field
that a bot can satisfy.

## Planned Implementation

- Make component workflows `workflow_call` only and invoke them from Validation.
- Preserve independent full and metadata concurrency domains and only the two
  required aggregate gate contexts.
- Re-export one contribution-title implementation and test repository/template
  parity.
- Make the GitHub governance validator a thin adapter over the shared PR policy.
- Parse configured Dependabot ecosystems/directories and restrict bot changes
  to their manifest, lock, or workflow paths.
- Require exact bot login/ref, `type:chore`, `area:tooling`, assignee, and full
  validation while exempting human Issue/task-report/multilingual/AI fields.
- Document the exception and regression-test valid and adversarial scenarios.

## Acceptance Criteria

- A PR event produces no independent component workflow runs.
- Human PR validation behavior remains unchanged.
- Configured npm, pip, gomod, and GitHub Actions bot updates pass.
- Impersonation, wrong refs, unsupported paths, labels, or assignment fail.
- Title behavior is identical in repository and downstream modules.
- #89 and #90 can be refreshed and deterministically revalidated without being
  merged by this task.

## Approval

- **Status:** `Approved`
- **Approved by:** User (approved the implementation plan)

## Implementation Status

Implementation is complete locally on the dedicated Step 2 branch and awaits
hosted validation after the preceding deployment PR is merged.

## Verification

- Shared title/policy/workflow Node regression tests — 43 tests passed
- Full repository Node governance/deployment suite — passed
- `actionlint` for all workflows — passed
- `yaml-lint` for repository GitHub YAML — passed
- Markdown lint for all tracked docs and this report — passed
- `git diff --check` — passed
- Hosted Validation and #89/#90 refresh — pending

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex

---

## 목표 및 배경

Issue [#93](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/93)은
Dependabot PR #89와 #90에서 발생하는 중복 PR workflow 실행과 인간 전용 governance
필드 실패를 다룹니다.

## 제안하는 접근

이벤트를 구독하는 Validation workflow는 하나만 유지하고 title·PR policy workflow는
재사용 구성요소로 전환합니다. 중복 JavaScript 규칙은 공통 모듈로 수렴시키며, bot이
충족할 수 있는 필수 check와 metadata는 유지한 채 정확하고 파일 범위가 제한된
Dependabot에만 body/title 예외를 적용합니다.

## 계획된 구현

- 구성요소 workflow를 `workflow_call` 전용으로 바꾸고 Validation에서 호출
- full/metadata concurrency 영역과 필수 aggregate gate 두 개 유지
- contribution-title 구현 재사용 및 repository/template parity 테스트
- GitHub governance validator를 공통 PR policy adapter로 축소
- Dependabot ecosystem/directory를 읽어 manifest·lock·workflow 경로만 허용
- 정확한 bot login/ref, `type:chore`, `area:tooling`, 담당자, 전체 validation 필수
- 정상 및 공격적 시나리오 회귀 테스트와 정책 문서화

## 수용 기준

- PR 이벤트가 독립 component workflow를 생성하지 않습니다.
- 일반 사용자 PR 검증은 기존 동작을 유지합니다.
- 등록된 npm, pip, gomod, GitHub Actions update는 통과합니다.
- 위장 계정, 잘못된 ref·경로·라벨·담당자는 실패합니다.
- repository와 downstream title 동작이 같습니다.
- #89/#90은 병합 없이 최신 기준으로 결정적으로 재검증할 수 있습니다.

## 승인

- **상태:** `Approved`
- **승인자:** 사용자 (구현 계획 승인)

## 구현 현황

Step 2 전용 브랜치의 로컬 구현을 완료했으며 선행 배포 PR 병합 후 hosted
validation을 기다립니다.

## 검증

- 공통 title/policy/workflow Node 회귀 테스트 — 43개 통과
- 전체 repository Node governance/deployment suite — 통과
- 전체 workflow `actionlint` — 통과
- repository GitHub YAML `yaml-lint` — 통과
- 전체 tracked 문서와 본 보고서 Markdown lint — 통과
- `git diff --check` — 통과
- hosted Validation과 #89/#90 갱신 — 대기

## AI 지원

- **계획/구현/초안 작성:** OpenAI Codex
