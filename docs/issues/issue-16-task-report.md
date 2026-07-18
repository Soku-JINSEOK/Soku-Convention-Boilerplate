# 📝 Task Report: `soku` Lifecycle Contract

## Goal and Background

[Issue #16](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/16)
requires a normative architecture and lifecycle contract before implementation
of the `soku` CLI, portable manifest, safe upgrade engine, or provider loader.
The existing bootstrap and synchronization documents describe manual
operations, but they do not define automated ownership, compatibility,
non-destructive planning, transaction, or rollback behavior.

This task is based on the latest `main` merge commit `c13e698` (PR #25). It is
documentation and CI-hygiene work only: no Issue or Project mutation, commit,
push, or pull request is part of the approved scope.

## Proposed Approach

Add `docs/standards/SOKU_LIFECYCLE.md` as one normative architecture decision
record containing both the decision rationale and the enforceable contract.
Choose a Go cross-platform binary and an immutable GitHub Release distribution
model, then define the public commands, configuration precedence, exit codes,
independent compatibility axes, portable manifest meanings, ownership and drift
rules, provider boundaries, and one outer transaction.

Keep wire formats and implementation choices with their roadmap issues. The ADR
reserves semantic boundaries for Issues #17–#22 without defining a
consumer-specific provider adapter or treating `ci-cd-control-plane-v1` as core
logic.

## Planned Implementation

- Add `docs/standards/SOKU_LIFECYCLE.md` with the complete lifecycle contract,
  conformance scenarios, roadmap handoff, and acceptance-criteria traceability.
- Register the ADR as normative authority in `BLUEPRINT.md`, all three README
  document indexes, the applicability matrix, and repository-hygiene CI.
- Mark `INIT_GUIDE.md` as the legacy manual workflow and link the future
  ownership-aware `soku init` boundary.
- Distinguish manual release/sync scripts from the immutable, compatibility-aware
  lifecycle in `RELEASE_AND_SYNC.md`.
- Add this approved bilingual task report, then record only checks actually run.

## Acceptance Criteria

| Issue #16 criterion | Evidence |
| --- | --- |
| Record CLI packaging, command boundaries, configuration precedence, and exit codes. | `SOKU_LIFECYCLE.md`: Decision Summary and Command Contract. |
| Define lifecycle terminology. | `SOKU_LIFECYCLE.md`: Lifecycle Terminology. |
| Define `init`, `status`, `diff`, and `upgrade`. | `SOKU_LIFECYCLE.md`: Public Commands and Responsibilities. |
| Define CLI, manifest, and boilerplate compatibility. | `SOKU_LIFECYCLE.md`: Independent Compatibility Axes and Portable Manifest Contract. |
| Define non-destructive defaults, confirmation, dry-run, and rollback. | `SOKU_LIFECYCLE.md`: Ownership and Drift Rules, Plan and Confirmation Contract, and Transaction and Rollback Contract. |
| Align `BLUEPRINT.md`, `INIT_GUIDE.md`, and `RELEASE_AND_SYNC.md` without duplicating authority. | Each document links to the ADR and states only its own authority boundary. |
| Add an approved task report before implementation. | This report records approval before lifecycle implementation begins. |

## Approval

- **Status:** `Approved`
- **Approved by:** `Soku-JINSEOK` (the repository owner explicitly requested
  implementation of the reviewed plan in this session before changes began)

## Implementation Status

Complete. The normative ADR and authority-boundary updates are implemented and
verified locally. The repository owner subsequently authorized the agent to
commit, publish, and merge this work as a separate pull request before Issue
#17 implementation begins.

## Verification

- `npx --yes markdownlint-cli2@0.22.1 --config .markdownlint.jsonc
  "**/*.md" "#node_modules"` — 38 files linted, 0 errors.
- `npx --yes yaml-lint@1.7.0 .github/*.yml .github/**/*.yml` — pass.
- `node --test templates/_shared/commitlint/*.test.mjs` — 1 test passed.
- `bash -n scripts/*.sh` — pass.
- Lifecycle contract assertions — pass. The assertions checked all commands and
  common flags, exit codes `0`–`8`, manifest and ownership terms, four required
  integration states, provider execution prohibition, rollback, authority
  links, CI registration, and both acceptance-criteria traceability tables.
- SHA scenario check — the lowercase 40-character SHA passed; a branch, tag,
  uppercase SHA, 39-character SHA, and 41-character SHA were rejected.
- The path, secret-persistence, pending/connected, pre-mutation conflict, and
  provider-failure rollback scenarios are present in the ADR's Conformance
  Scenarios table and covered by the contract assertions.
- `git diff --check` — pass.
- GitHub's current official documentation was checked for the ADR rationale:
  standard GitHub-hosted runners are free for public repositories, and Releases
  support distributing binary files.
- Sync parity — pass with checksum-verified PowerShell 7.6.3 portable for Linux
  x64. `scripts/verify-sync-parity.sh` confirmed identical bash and PowerShell
  output with no leaked artifacts. A second run used a temporary clone where
  all current changes, including both new documents, were tracked; it also
  passed, confirming the post-commit sync behavior without changing this work
  tree's index.

## AI Assistance

- **Planning/analysis/drafting:** OpenAI Codex
- **Implementation/verification:** OpenAI Codex

---

## 목표 및 배경

[Issue #16](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/16)은
`soku` CLI, portable manifest, 안전한 upgrade engine, provider loader를 구현하기
전에 하나의 규범 lifecycle 계약을 요구합니다. 기존 bootstrap·sync 문서는 수동
절차를 설명하지만 자동화에 필요한 소유권, 호환성, 비파괴 plan, transaction,
rollback 계약은 정의하지 않습니다.

이 작업은 최신 `main`의 PR #25 merge commit `c13e698`을 기준으로 진행합니다.
문서와 CI hygiene만 변경하며 Issue/Project 변경, commit, push, PR 생성은 승인
범위에 포함되지 않습니다.

## 제안하는 접근

`docs/standards/SOKU_LIFECYCLE.md`를 결정 배경과 강제 계약을 함께 담는 단일
normative ADR로 추가합니다. Go 기반 cross-platform binary, immutable release,
공개 command, 설정 우선순위, exit code, 독립적인 compatibility axis, portable
manifest 의미, ownership·drift, provider 경계, outer transaction을 정의합니다.

Wire format과 구현 세부 사항은 후속 Issue #17–#22에 남깁니다.
`ci-cd-control-plane-v1`은 consumer 사례로만 다루며 core에 특수 adapter나
provider별 분기를 넣지 않습니다.

## 계획된 구현

- 전체 lifecycle 계약, conformance scenario, 후속 Issue handoff, 수용 기준
  추적표를 포함하는 `SOKU_LIFECYCLE.md`를 추가합니다.
- `BLUEPRINT.md`, README 3개 언어 색인, applicability matrix, repository-hygiene
  CI에 ADR을 등록합니다.
- `INIT_GUIDE.md`를 legacy manual workflow로 구분하고 미래 `soku init` 경계를
  연결합니다.
- `RELEASE_AND_SYNC.md`에서 기존 수동 sync script와 immutable source 및
  compatibility를 요구하는 lifecycle 계약을 구분합니다.
- 승인된 이중 언어 task report를 추가하고 실제 실행한 검증만 기록합니다.

## 수용 기준

| Issue #16 기준 | 근거 |
| --- | --- |
| CLI packaging, command 경계, 설정 우선순위, exit code 정의 | ADR의 Decision Summary와 Command Contract |
| lifecycle 용어 정의 | ADR의 Lifecycle Terminology |
| `init`, `status`, `diff`, `upgrade` 정의 | ADR의 Public Commands and Responsibilities |
| CLI, manifest, boilerplate compatibility 정의 | ADR의 Independent Compatibility Axes와 Portable Manifest Contract |
| 비파괴 기본값, confirmation, dry-run, rollback 정의 | ADR의 Ownership and Drift Rules, Plan and Confirmation Contract, Transaction and Rollback Contract |
| 기존 권한 문서 정렬 | 세 문서가 ADR을 링크하고 각 문서 고유 경계만 설명 |
| 구현 전 승인된 task report 추가 | 이 문서의 승인 기록 |

## 승인

- **상태:** `Approved`
- **승인자:** `Soku-JINSEOK` (저장소 소유자가 이 세션에서 검토된 계획의
  구현을 변경 시작 전에 명시적으로 요청함)

## 구현 현황

규범 ADR과 권한 경계 정렬을 구현하고 로컬 검증을 완료했습니다. 저장소
소유자가 이후 Issue #17 구현 전에 이 작업을 별도 PR로 commit, publish,
merge하도록 agent에게 승인했습니다.

## 검증

- `markdownlint-cli2@0.22.1` — Markdown 38개 파일, 오류 0건
- `yaml-lint@1.7.0` — 통과
- commit-title regression test — 1개 통과
- `bash -n scripts/*.sh` — 통과
- lifecycle 계약 assertion — command·공통 flag, exit code `0`–`8`, manifest,
  ownership, integration 상태 4개, provider 실행 금지, rollback, 권한 링크, CI
  등록, 수용 기준 추적표 확인 완료
- SHA scenario — lowercase 40자는 허용하고 branch, tag, uppercase, 39자,
  41자는 거부함을 확인
- path, secret 저장 금지, pending/connected, conflict 사전 중단, provider 실패
  rollback scenario가 ADR에 있으며 assertion 대상임을 확인
- `git diff --check` — 통과
- GitHub 공식 문서에서 public repository의 standard runner 무료 사용과
  Releases의 binary 배포 지원 근거 확인
- 공식 release checksum을 확인한 Linux x64 portable PowerShell 7.6.3으로 sync
  parity 통과. Bash와 PowerShell 출력이 동일하고 artifact 누출이 없음을 확인함.
  현재 변경과 새 문서 2개를 모두 track한 임시 clone에서도 다시 통과하여 실제
  작업 트리 index를 변경하지 않고 commit 이후 sync 동작까지 확인함.

## AI 지원

- **계획/분석/초안 작성:** OpenAI Codex
- **구현/검증:** OpenAI Codex
