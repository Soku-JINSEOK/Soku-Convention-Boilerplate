# Issue #100 Task Report — Audit complete Issue and pull-request history

## Goal and Background

Issue [#100](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/100)
requires a reproducible, read-only governance audit of the repository's complete
Issue and pull-request history. The approved plan expected 33 Issues and 48
pull requests. An API inventory at the agreed cutoff found 33 Issues and 59 pull
requests, so omitting the additional 11 pull requests would make the result
incomplete.

## Proposed Approach

Use creation timestamps and the policy-changing Git commits to select the
contract applicable when each item was created. Read current metadata and
historical evidence with GET-only `gh api` calls, process raw bodies only in
memory, and persist only body SHA-256 values, evidence counts, classifications,
and a separately approved mutation manifest.

## Planned Implementation

- Add `scripts/github-governance-audit.mjs` with `--repo`, `--as-of`, and
  `--output` interfaces.
- Inventory every Issue and PR created by `2026-07-23T01:00:19Z`.
- Read commits, check runs, reviews, comments, changed files, labels,
  assignees, merge state, and signature verification through GET requests.
- Apply the legacy, multilingual, strengthened, and strict policy epochs from
  their source Git commits without retroactive enforcement.
- Apply the requested current contracts only to #89, #90, #91, and #92.
- Write the complete result to
  `docs/audits/github-governance-2026-07-23.md`.
- Add hermetic tests for argument validation, GET-only collection, epoch
  boundaries, current Issue and Dependabot contracts, and raw-body exclusion.
- Add only unambiguous metadata corrections to the follow-up manifest; do not
  apply any GitHub mutation in this PR.

## Acceptance Criteria

- All 33 Issues and all 59 PRs before the cutoff receive exactly one judgment.
- Every judgment uses one of the five approved classifications.
- Issue/PR raw bodies do not appear in repository artifacts.
- The report reconciles the planned 81-item count with the API's 92-item
  inventory.
- Issue #91 remains normalized with one consolidated comment, and #89/#90 use
  the exact current Dependabot contract.
- The tool and report are deterministic, reviewable, and read-only.
- Any metadata change remains pending separate approval and a fresh read.

## Approval

- **Status:** `Approved`
- **Approved by:** User (approved the implementation plan)

## Implementation Status

Implementation is complete on the audit branch. The GET-only inventory found
and judged 92 items: 37 compliant, 3 correctable metadata, 48 historical
exceptions, 4 blocked/missing evidence, and 0 not applicable. Issue #91 is
compliant and has exactly one consolidated comment. The mutation manifest is
review-only and has not been applied.

## Verification

- Audit regression suite — 6 tests passed
- Live GET-only audit — 33 Issues + 59 PRs = 92 judgments written
- Body privacy check — raw marker excluded; SHA-256 retained
- Issue #91 — compliant; one comment; linked PRs #92, #94, #95, and #96
- PRs #89/#90 — current Dependabot contract compliant
- GitHub mutation — none
- Full repository validation — pending PR execution

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex

---

## 목표 및 배경

Issue [#100](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/100)은
저장소의 전체 Issue 및 pull request 이력을 재현 가능하고 읽기 전용으로 감사하는
작업입니다. 승인된 계획은 Issue 33건과 PR 48건을 예상했지만, 합의한 cutoff의 API
inventory에는 Issue 33건과 PR 59건이 존재했습니다. 추가 11건을 제외하면 전수 감사가
아니므로 모두 포함합니다.

## 제안하는 접근

생성 시각과 정책 변경 Git commit으로 각 항목의 당시 계약을 선택합니다. 현재
metadata와 역사 증거는 GET 전용 `gh api`로 읽고, raw body는 메모리에서만 처리하며,
저장소에는 body SHA-256, 증거 개수, 판정, 별도 승인이 필요한 mutation manifest만
남깁니다.

## 계획된 구현

- `--repo`, `--as-of`, `--output`을 지원하는
  `scripts/github-governance-audit.mjs` 추가
- `2026-07-23T01:00:19Z`까지 생성된 모든 Issue와 PR inventory
- GET 요청으로 commit, check run, review, comment, 변경 파일, label, assignee, merge
  state와 signature 검증 정보 조회
- 정책 원본 Git commit에 따라 legacy, multilingual, strengthened, strict 시점 적용
- 요청된 #89, #90, #91, #92에만 현재 계약 적용
- `docs/audits/github-governance-2026-07-23.md`에 전체 결과 기록
- 인자, GET 강제, 정책 경계, 현재 Issue/Dependabot 계약, raw-body 비저장을 검증하는
  hermetic test 추가
- 명확한 metadata 수정만 후속 manifest에 기록하고 이번 PR에서는 GitHub 변경 금지

## 수용 기준

- cutoff 이전 Issue 33건과 PR 59건이 각각 하나의 판정을 가집니다.
- 모든 판정이 승인된 다섯 분류 중 하나를 사용합니다.
- Issue/PR raw body가 저장소 산출물에 포함되지 않습니다.
- 계획의 81건과 API의 92건 차이가 보고서에서 조정됩니다.
- Issue #91은 하나의 통합 댓글을 유지하고 #89/#90은 정확한 현재 Dependabot 계약으로
  검사됩니다.
- 도구와 보고서는 결정적이고 검토 가능하며 읽기 전용입니다.
- metadata 변경은 별도 승인과 fresh read 전까지 보류합니다.

## 승인

- **상태:** `Approved`
- **승인자:** 사용자 (구현 계획 승인)

## 구현 현황

감사 branch 구현을 완료했습니다. GET 전용 inventory에서 92건을 판정했으며 compliant
37건, correctable metadata 3건, historical exception 48건, blocked/missing evidence
4건, not applicable 0건입니다. Issue #91은 compliant이며 통합 댓글 1건만 존재합니다.
mutation manifest는 검토 전용이고 적용하지 않았습니다.

## 검증

- 감사 회귀 suite — 6개 테스트 통과
- live GET 전용 감사 — Issue 33건 + PR 59건 = 판정 92건 기록
- body privacy 검사 — raw marker 제외, SHA-256 보존
- Issue #91 — compliant, 댓글 1건, 연결 PR #92, #94, #95, #96
- PR #89/#90 — 현재 Dependabot 계약 compliant
- GitHub mutation — 없음
- 전체 repository validation — PR 실행 대기

## AI 지원

- **계획/구현/초안 작성:** OpenAI Codex
