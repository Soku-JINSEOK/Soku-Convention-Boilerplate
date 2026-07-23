# Issue #102 Task Report — Simplify issue templates for operator efficiency

## Goal and Background

Issue [#102](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/102) requires issue forms to stop forcing users to fill full multilingual and metadata fields for every task.

## Proposed Approach

Keep English task framing as the contract, and reduce required inputs for
issue forms to the requested core:

- core task fields (`goal`, `scope`-style content, or equivalent)
- `acceptance`
- `safety`

Optional metadata remains in each form:

- `priority`
- `area`
- `summary_ko`
- `summary_ja`
- `ai_assistance`

Implement this together with a corresponding policy update in
`docs/standards/GITHUB_STANDARDS.md`, and update
`.github/issue-form-order.test.mjs` to assert exact required/optional order.

## Planned Implementation

- Update all five issue forms in `.github/ISSUE_TEMPLATE/*.yml` to mark
  optional fields as not required.
- Update `docs/standards/GITHUB_STANDARDS.md`:
  - PRs require full bilingual contract.
  - Issues require only English core task framing + safety/acceptance.
- Update `.github/issue-form-order.test.mjs` to validate required and optional
  groups.

## Acceptance Criteria

- Every issue form requires only its core English task fields plus safety and
  acceptance.
- `summary_ko`, `summary_ja`, `priority`, `area`, and `ai_assistance` are optional.
- Governance standards and issue-form-order test reflect the same policy.

## Approval

- **Status:** `Approved`
- **Approved by:** User

## Implementation Status

Implemented in local Issue #102 branch work: issue templates and standards/test policy updated.

## Verification

- Completed locally:
  - `node --test .github/issue-form-order.test.mjs`
  - `node --test .github/usage-manual.test.mjs`

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex

---

## 목표 및 배경

Issue [#102](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/102)는
모든 이슈에 대해 다국어/우선순위/영역/AI 메타데이터를 강제 입력하지 않도록
요청합니다.

## 제안하는 접근

이슈 템플릿의 핵심 계약은 English 중심으로 유지하고, 필수 필드는 아래로
축소합니다.

- 핵심 English 작업 필드 (`goal`, `scope`/`work`/`content`/`rationale` 등)
- `acceptance`
- `safety`

선택 필드는 각 폼에 유지합니다:

- `priority`
- `area`
- `summary_ko`
- `summary_ja`
- `ai_assistance`

동시에 `docs/standards/GITHUB_STANDARDS.md`의 규약 문구를 정렬하고,
`.github/issue-form-order.test.mjs`에서 필수/선택 집합을 검증하도록
업데이트합니다.

## 계획된 구현

- 5개 이슈 템플릿에서 선택 항목을 `required: false`로 변경
- `docs/standards/GITHUB_STANDARDS.md`에서 PR/이슈 계약을 분리 정리
- `.github/issue-form-order.test.mjs`의 required/optional 순서를 반영한 검증 로직 반영

## 수용 기준

- 모든 이슈 템플릿에서 핵심 English 작업 필드 + `safety` + `acceptance`가 필수
- `summary_ko`, `summary_ja`, `priority`, `area`, `ai_assistance`가 선택
- 표준 문서와 이슈 폼 순서 테스트가 동일 정책을 반영

## 승인

- **상태:** `Approved`
- **승인자:** 사용자

## 구현 현황

Issue #102 분기에서 이슈 템플릿, 표준 문서, 테스트 정책 변경 작업 반영 완료.

## 검증

- 로컬 회귀 검증을 완료했습니다:
  - `node --test .github/issue-form-order.test.mjs`
  - `node --test .github/usage-manual.test.mjs`

## AI 지원

- **계획/구현/초안 작성:** OpenAI Codex
