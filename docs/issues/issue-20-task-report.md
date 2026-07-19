# 📝 Task Report: Transactional `soku diff` and `soku upgrade`

## Goal and Background

[Issue #20](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/20)
adds the ownership-aware transition workflow after the immutable `init` and
portable manifest foundations delivered by Issues #18 and #19. Downstream
repositories need to compare two pinned releases and move forward without
discarding project-owned content or managed-file customizations.

## Proposed Approach

`soku diff --boilerplate-release vX.Y.Z` and `soku upgrade
--boilerplate-release vX.Y.Z` use the current manifest source, verify both the
recorded and target tags as immutable commits, and render the base and target
trees from the stored portable selection. V1 supports exact forward releases
within one source. Downgrades, source changes, unsupported manifests, and moved
tags stop with compatibility exit `5` before mutation.

The planner compares base, current, and target content in path order and emits
`added`, `updated`, `removed`, `merged`, `unchanged`, `locally-modified`, or
`conflict`. Core-managed drift and ownership/path conflicts stop before writes.
Project-owned paths are never overwritten. `.gitignore` uses a line-set 3-way
merge and `.editorconfig` uses a section/key 3-way merge, preserving independent
local additions while rejecting changes to the same logical entry.

Upgrade applies creates, replacements, merges, and removals through the existing
journal/backup transaction, then atomically replaces the manifest last. Dry-run,
cancellation, and conflicts are zero-write. Successful rollback remains exit
`7`; rollback failure remains exit `8` with bounded recovery evidence.

## Planned Implementation

- Add command flags and stable human/JSON reports for `diff` and `upgrade`.
- Reuse the immutable source fetcher and expose injected fetchers for hermetic
  sequential-release tests.
- Reconstruct base and target desired trees from manifest selection plus each
  pinned source snapshot.
- Add deterministic 3-way planning, shared-file structural merges, compatibility
  checks, and first-write safety gates.
- Extend the transaction engine to remove obsolete files and restore deleted
  content during rollback while committing the target manifest last.
- Cover forward success, deletion, local customization, conflicts, moved tags,
  downgrade, cancellation, dry-run, rollback, no-op, and rerun.
- Document the public contract and migration boundary.

## Acceptance Criteria

| Criterion | Observable evidence |
| --- | --- |
| Immutable compatibility | Both releases resolve to recorded commits; source change, downgrade, moved tag, or unsupported input exits `5` without writes. |
| Truthful ordered plan | Human and JSON output expose all seven path states in deterministic order; `diff` exits `3` only for a non-empty transition. |
| Local intent preserved | Core drift conflicts, project-owned paths are untouched, and independent shared-file additions survive structural 3-way merge. |
| Transactional apply | Creates, updates, merges, and deletes are journaled and backed up; manifest replacement occurs last; rollback exits remain `7`/`8`. |
| Zero-write review | `diff`, dry-run, cancellation, compatibility refusal, and conflict do not change managed files or `.soku` state. |
| Sequential coverage | Hermetic releases verify success, deletion, conflict, rollback, rerun, and no-op behavior without real tags or network access. |

## Approval

- **Status:** `Approved`
- **Approved by:** `Soku-JINSEOK`
- **Approval record:** The repository owner's 2026-07-19 instruction to
  implement the previously approved complete roadmap plan.

## Implementation Status

In progress. The approved design is the implementation boundary for Issue #20.

## Verification

- Not started.

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex

---

## 목표 및 배경

[Issue #20](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/20)은
immutable `init`과 portable manifest 기반 위에 소유권을 인식하는 업데이트 흐름을
추가합니다. 다운스트림 저장소는 project-owned 내용과 managed-file 로컬 수정을
손상하지 않고 두 고정 release를 비교하고 앞으로 이동할 수 있어야 합니다.

## 제안하는 접근

`soku diff`와 `soku upgrade`는 현재 manifest의 source를 사용하고 현재·대상 tag를
immutable commit으로 검증한 뒤 저장된 portable selection으로 base/target tree를
다시 렌더링합니다. V1은 같은 source의 정확한 forward release만 지원합니다.

Planner는 base/current/target을 경로순으로 비교하고 7개 상태를 출력합니다.
core-managed drift와 ownership 충돌은 write 전에 중단하고, 두 공유 파일만 구조적인
3-way merge로 독립 로컬 항목을 보존합니다. 적용은 기존 journal/backup transaction에
생성·교체·merge·삭제를 포함하고 manifest를 마지막에 교체합니다.

## 계획된 구현

- `diff`/`upgrade` CLI flag, human/JSON report와 exit contract 구현
- 고정 source fetcher와 manifest selection 기반 양쪽 desired tree 재구성
- 경로순 3-way planner와 공유 파일 구조 merge 및 compatibility gate 구현
- 삭제와 rollback 복원을 포함하도록 transaction 확장
- 두 개 이상의 synthetic release로 성공·삭제·충돌·dry-run·취소·rollback·rerun 검증
- 사용자 문서와 migration 범위 갱신

## 수용 기준

| 기준 | 관찰 가능한 근거 |
| --- | --- |
| Immutable compatibility | 양쪽 commit 검증, source 변경·downgrade·moved tag 거부와 zero-write |
| 정확한 계획 | 7개 상태를 결정적 순서의 human/JSON으로 출력하고 변경 시 `diff` exit `3` |
| 로컬 의도 보존 | core drift 충돌, project-owned 무변경, 공유 파일 독립 로컬 항목 보존 |
| Transaction | 생성·갱신·merge·삭제 backup, manifest-last, rollback `7`/`8` 유지 |
| Zero-write 검토 | diff·dry-run·취소·compatibility·conflict에서 target과 `.soku` 무변경 |
| 연속 release 검증 | 실제 tag/network 없이 성공·삭제·충돌·rollback·rerun·no-op 검증 |

## 승인

- **상태:** `Approved`
- **승인자:** `Soku-JINSEOK`
- **승인 기록:** 2026-07-19 확정 계획 전체 구현 지시에 따라 승인함

## 구현 현황

진행 중입니다. 승인된 설계를 Issue #20 구현 경계로 사용합니다.

## 검증

- 시작 전

## AI 지원

- **계획/구현/초안 작성:** OpenAI Codex
