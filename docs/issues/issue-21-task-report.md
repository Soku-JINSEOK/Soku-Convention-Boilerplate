# 📝 Task Report: Core Lifecycle End-to-End Release Gate

## Goal and Background

[Issue #21](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/21)
requires release-level evidence for the complete downstream lifecycle rather
than isolated command and template tests. This first phase establishes the core
gate before provider conformance is added after Issue #22.

## Proposed Approach

Add a hermetic Go end-to-end package that injects immutable synthetic release
snapshots and runs initialization, status inspection, local customization,
diff, upgrade, and final status without real tags or network access. Execute the
same package on Linux, macOS, and Windows, while keeping the existing Linux
template jobs responsible for runtime-specific install, lint, typecheck, test,
and build commands.

The OS matrix covers empty, existing, single-stack, and multi-stack fixtures,
canonical line endings, case collisions, symlink boundaries where supported,
atomic manifest replacement, deletion rollback, rerun, and no-op behavior.
Failure artifacts contain only sanitized Go test JSON with workspace, temporary,
and home paths removed.

## Planned Implementation

- Add hermetic core lifecycle tests using injected source snapshots.
- Add filesystem-risk tests that run proportionately on each supported OS.
- Add a sanitized failure-log runner and bounded artifact retention.
- Run the gate for pull requests, `main`, `v*`, and `soku/v*` events.
- Make the CLI release job depend on the core lifecycle matrix.
- Trigger runtime template validation when lifecycle rendering changes.
- Record actual local and pull request verification results.

## Acceptance Criteria

| Criterion | Observable evidence |
| --- | --- |
| Full core flow | Fixtures execute `init → status → local edit → diff → upgrade → status`. |
| Hermetic releases | Tests inject exact synthetic commits and perform no network or tag operations. |
| Stack coverage | Single and multi-stack trees contain no unresolved supported placeholders; Linux template jobs run documented runtime checks. |
| OS filesystem risk | Linux, macOS, and Windows execute one lifecycle package with platform-aware symlink and line-ending assertions. |
| Release gate | PR, `main`, `v*`, and `soku/v*` events run the matrix; CLI publishing depends on it. |
| Safe diagnostics | Failure artifacts are path-sanitized, bounded logs with short retention. |

## Approval

- **Status:** `Approved`
- **Approved by:** `Soku-JINSEOK`
- **Approval record:** The repository owner's 2026-07-19 instruction to
  implement the approved Issue #20–#23 roadmap plan.

## Implementation Status

In progress. This report covers only the first `Related to #21` core gate. The
provider conformance phase remains a separate final PR after Issue #22.

## Verification

- Not started.

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex

---

## 목표 및 배경

[Issue #21](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/21)의
첫 단계로 개별 command 테스트를 넘어 다운스트림 core lifecycle 전체를 검증하는
release gate를 구축합니다. provider conformance는 Issue #22 이후 두 번째 PR에서
연결합니다.

## 제안하는 접근

실제 tag와 network 대신 immutable synthetic snapshot fetcher를 주입해
`init → status → local edit → diff → upgrade → status`를 실행합니다. 동일 package를
Linux, macOS, Windows에서 실행하고, 기존 Linux template job은 각 runtime의 install,
lint, typecheck, test, build를 담당합니다.

## 계획된 구현

- empty/existing/single/multi-stack hermetic fixture와 placeholder 검사
- line ending, path case, symlink, atomic replacement, rollback 위험 matrix
- 경로를 제거한 제한적 실패 artifact와 짧은 보존 기간
- PR, `main`, `v*`, `soku/v*` event의 동일 gate와 CLI release 의존성
- 실제 local/PR 검증 결과 기록

## 수용 기준

| 기준 | 관찰 가능한 근거 |
| --- | --- |
| 전체 흐름 | 초기화부터 최종 clean status까지 실제 순서 실행 |
| Hermetic | network/tag 없이 exact synthetic commit 주입 |
| Stack | single/multi 결과의 placeholder 부재와 Linux runtime 검증 |
| OS 위험 | 세 OS에서 동일 lifecycle package와 platform-aware assertion 실행 |
| Release gate | 네 event 범위와 CLI publish 의존성 |
| 안전한 진단 | 경로 제거, 짧은 보존 기간의 제한적 실패 로그 |

## 승인

- **상태:** `Approved`
- **승인자:** `Soku-JINSEOK`
- **승인 기록:** 2026-07-19 승인된 Issue #20–#23 roadmap 전체 구현 지시

## 구현 현황

진행 중입니다. 이 보고서의 첫 PR은 `Related to #21` core gate만 다루며,
provider conformance는 Issue #22 이후 별도 최종 PR로 유지합니다.

## 검증

- 시작 전

## AI 지원

- **계획/구현/초안 작성:** OpenAI Codex
