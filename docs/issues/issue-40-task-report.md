# 📝 Task Report: Replace Vulnerable Pyink with Black

## Goal and Background

[Issue #40](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/40)
tracks the vulnerable Black version forced by the latest supported Pyink
release. The approved migration removes Pyink and adopts a supported Black
release directly without suppressing security advisories.

## Proposed Approach

Use Black `>=26.5.1,<27`, pin `black==26.5.1` in the reproducible lockfile, and
retain the Python template's 80-column and Python 3.11 formatting contract.
Keep Ruff lint-only and limit formatting changes to the Python template.

## Planned Implementation

- Replace Pyink configuration and commands with Black.
- Regenerate and verify the Python dependency lock.
- Run formatting, lint, type, test, dependency, and repository validation.

## Acceptance Criteria

- Pyink and Black versions below 26.5.1 are absent from the active template.
- Black, Ruff, mypy, pytest, pip-audit, and OSV checks pass.
- Runtime-template and aggregate validation gates pass.

## Approval

- **Status:** `Approved`
- **Approved by:** User on 2026-07-20
- **Approval boundary:** Local implementation and verification for Issue #40.
  Publishing a PR, merging, or closing the issue remains a separate action.

## Implementation Status

Implementation is complete on a dedicated worktree and published draft PR
[#45](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/pull/45).
Hosted Validation passed. Merge and issue closure remain pending their approval
boundary.

## Verification

- Passed: Python 3.14 install, Black 26.5.1 check, Ruff 0.15.21, mypy 2.3.0,
  and pytest 9.1.1.
- Passed: Python 3.12 binary-wheel resolution dry-run for the complete lockfile.
- Passed: byte-identical `pip-compile` regeneration; SHA-256
  `40fd8b6c47757be0c9554b8a2401494add1217b101c57799914d30cfbc6271a2`.
- Passed: pip-audit 2.10.0 and OSV Scanner 2.3.8 with no known issues.
- Passed: all `soku` Go tests with `GOFLAGS=-buildvcs=false`; this flag only
  works around unavailable VCS stamping in the isolated worktree sandbox.
- Passed: gofmt, Markdown, YAML, actionlint, Bash syntax, and `git diff --check`.
- Passed: hosted Validation run
  [29717172603](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/29717172603)
  on signed commit `0cc7265`, including shellcheck, sync parity, all runtime
  templates, three-OS lifecycle and CLI checks, package snapshots, and the
  aggregate `Validation Gate`.
- A prior identical-tree run exposed a pre-existing MySQL service-readiness
  race; the isolated job retry and the signed-commit run both passed without a
  code change.

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex

---

## 목표 및 배경

[Issue #40](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/40)은
최신 지원 Pyink가 강제하는 취약한 Black 버전을 추적합니다. 승인된 이전은 보안
advisory를 억제하지 않고 Pyink를 제거한 뒤 지원되는 Black을 직접 사용합니다.

## 제안하는 접근

Black `>=26.5.1,<27`을 사용하고 재현 가능한 lockfile에는 `black==26.5.1`을
고정합니다. Python template의 80열 및 Python 3.11 formatting 계약은 유지하고,
Ruff는 lint 전용으로 둡니다.

## 계획된 구현

- Pyink 설정과 명령을 Black으로 교체합니다.
- Python dependency lock을 재생성하고 검증합니다.
- format, lint, type, test, dependency 및 저장소 validation을 실행합니다.

## 수용 기준

- 활성 template에서 Pyink와 Black 26.5.1 미만 버전이 제거됩니다.
- Black, Ruff, mypy, pytest, pip-audit 및 OSV 검사가 통과합니다.
- runtime-template 및 aggregate validation gate가 통과합니다.

## 승인

- **상태:** `Approved`
- **승인자:** 사용자 (2026-07-20)
- **승인 범위:** Issue #40의 로컬 구현과 검증입니다. PR 게시, 병합, Issue 종료는
  별도 작업으로 남습니다.

## 구현 현황

전용 worktree와 draft PR
[#45](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/pull/45)에서
구현을 완료했고 hosted Validation도 통과했습니다. 병합과 Issue 종료는 승인 경계를
기다립니다.

## 검증

- 통과: Python 3.14 설치, Black 26.5.1, Ruff 0.15.21, mypy 2.3.0 및
  pytest 9.1.1.
- 통과: 전체 lockfile의 Python 3.12 binary wheel 해석 dry-run.
- 통과: `pip-compile` byte 단위 재생성; SHA-256
  `40fd8b6c47757be0c9554b8a2401494add1217b101c57799914d30cfbc6271a2`.
- 통과: pip-audit 2.10.0 및 OSV Scanner 2.3.8, 알려진 문제 0건.
- 통과: `GOFLAGS=-buildvcs=false`를 사용한 전체 `soku` Go test. 이 flag는 격리된
  worktree sandbox에서 사용할 수 없는 VCS stamping만 우회합니다.
- 통과: gofmt, Markdown, YAML, actionlint, Bash syntax 및 `git diff --check`.
- 통과: signed commit `0cc7265`의 hosted Validation run
  [29717172603](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/29717172603).
  shellcheck, sync parity, 모든 runtime template, 3개 OS lifecycle/CLI, package
  snapshot 및 aggregate `Validation Gate`를 포함합니다.
- 동일 tree의 이전 run에서 기존 MySQL service readiness race가 한 번 나타났지만,
  해당 job 재실행과 signed commit run은 코드 변경 없이 모두 통과했습니다.

## AI 지원

- **계획/구현/초안 작성:** OpenAI Codex
