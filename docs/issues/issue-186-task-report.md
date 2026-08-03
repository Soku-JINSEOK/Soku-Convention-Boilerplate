# 📝 Issue #186 Task Report

## Goal and Background

[Issue #186](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/186)
adds a restrained, terminal-friendly presentation policy and deterministic
completion scripts without changing Soku's automation contract.

## Proposed Approach

Keep every existing human report as the canonical plain-text representation.
A shared presentation package may add ANSI styling only after command execution
and only when the color policy permits it. Generate completion directly from
the pinned Cobra command tree with fixed local candidates and no state lookup.

## Planned Implementation

- Add `soku/internal/presentation` and `--color auto|always|never`.
- Detect the output TTY separately from the input TTY and honor `TERM=dumb` and
  `NO_COLOR`.
- Add `completion` for Bash, Zsh, Fish, and PowerShell, including JSON output.
- Add regression tests for output compatibility, color precedence, completion
  determinism, nested commands, and fixed candidates.
- Update the lifecycle contract and user-owned installation guidance.

## Acceptance Criteria

- Stripping ANSI from styled reports produces the existing plain report exactly.
- JSON, quiet output, errors, prompts, pipes, and completion scripts do not gain
  automatic ANSI output; exit codes and stream ownership remain unchanged.
- Four deterministic scripts are non-empty and expose the complete command tree.
- Soku tests, race tests, repository fast validation, and available shell smoke
  tests pass without changing Quick scope or shard configuration.

## Approval

- **Status:** `Approved`
- **Approved by:** `Soku-JINSEOK` (implementation instruction carrying the reviewed plan)

## Implementation Status

Implemented locally. The code is organized so presentation/color compatibility
and completion/documentation can be reviewed as two sequential commits or PRs.

## Verification

- `go test ./...`: passed.
- `go test -race ./...`: passed.
- `scripts/verify.sh --profile fast --files-from -`: passed for Soku scope.
- `soku/scripts/package_test.sh`: all five archives and checksums passed and
  were reproducible.
- Local shell paths and versions: `/bin/bash` 3.2.57,
  `/opt/homebrew/bin/bash` 5.3.15 with `bash-completion@2` 2.18.0,
  `/bin/zsh` 5.9, `/opt/homebrew/bin/fish` 4.8.1, and
  `/opt/homebrew/bin/pwsh` 7.6.4. Homebrew installed the missing verification
  shells without changing the login shell or a profile.
- Each script was generated twice from the local source build and checked with
  `test -s`, `cmp`, and an ANSI-byte scan. All four were non-empty,
  deterministic, and ANSI-free.
- Bash passed `bash -n`, session sourcing, `complete -p soku`, and live root,
  `docs manual`, `--color`, and `--profile` candidate checks.
- Zsh passed `zsh -n`, isolated `compinit -d`, script sourcing, and `_soku`
  registration. Fish passed `fish -n`, sourcing, and `complete -C` candidate
  checks. PowerShell passed parser validation, script evaluation, and
  `TabExpansion2` candidate checks.
- Bash, Fish, and PowerShell returned nested commands and fixed color/profile
  values directly. The shared Cobra completion endpoint used by the registered
  Zsh function returned the same `docs manual`, `completion`, color, and profile
  candidates.
- Real pipes and pseudo-TTYs verified `auto`, `always`, `never`, `NO_COLOR`, and
  `TERM=dumb`. JSON, quiet output, stderr errors, and completion scripts were
  ANSI-free; plain stdout/stderr ownership remained unchanged.
- `git diff --check`: passed.

## Public Disclosure Review

- [x] No credentials, tokens, private keys, or credential-bearing URLs
- [x] No private repository, project, or product names
- [x] No cloud project IDs, account numbers, service URLs, image URIs, or
      revision identifiers
- [x] No personal billing, subscription, budget, or payment-status information
- [x] No personal email, phone, address, or local absolute path
- [x] No private Issue, PR, Project, or control-plane identifiers

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex

---

## 목표 및 배경

[Issue #186](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/186)은
Soku 자동화 계약을 변경하지 않으면서 절제된 터미널 표시 정책과 결정적인
completion script를 추가합니다.

## 제안하는 접근

기존 human report를 canonical plain text로 유지합니다. 공통 presentation package는
명령 실행 후 color 정책이 허용할 때만 ANSI style을 추가합니다. Completion은 고정된
Cobra command tree와 로컬 후보에서 생성하며 상태 조회를 하지 않습니다.

## 계획된 구현

- `soku/internal/presentation`과 `--color auto|always|never` 추가
- 입력 TTY와 출력 TTY를 분리하고 `TERM=dumb`, `NO_COLOR` 반영
- Bash, Zsh, Fish, PowerShell completion 및 JSON 출력 추가
- 출력 호환성, color 우선순위, 결정성, 중첩 명령, 고정 후보 회귀 테스트 추가
- lifecycle 계약과 사용자 소유 설치 안내 갱신

## 수용 기준

- ANSI를 제거하면 기존 plain report와 정확히 일치합니다.
- JSON, quiet, error, prompt, pipe, completion script, exit code, stream 계약을
  보존합니다.
- 네 script가 결정적이고 비어 있지 않으며 전체 command tree를 노출합니다.
- Soku test, race, repository fast validation, 사용 가능한 shell smoke test를
  통과하고 Quick scope/shard 설정을 변경하지 않습니다.

## 승인

- **상태:** `Approved`
- **승인자:** `Soku-JINSEOK` (검토된 계획을 포함한 구현 지시)

## 구현 현황

로컬 구현을 완료했습니다. Presentation/color 호환성과 completion/documentation을
두 개의 순차 commit 또는 PR로 검토할 수 있도록 구성했습니다.

## 검증

- `go test ./...`: 통과
- `go test -race ./...`: 통과
- `scripts/verify.sh --profile fast --files-from -`: Soku scope 통과
- `soku/scripts/package_test.sh`: 다섯 archive와 checksum 검증 및 재현성 통과
- 로컬 shell 경로와 버전: `/bin/bash` 3.2.57,
  `/opt/homebrew/bin/bash` 5.3.15 및 `bash-completion@2` 2.18.0,
  `/bin/zsh` 5.9, `/opt/homebrew/bin/fish` 4.8.1,
  `/opt/homebrew/bin/pwsh` 7.6.4. 누락된 검증 shell은 Homebrew로
  설치했으며 login shell이나 profile은 변경하지 않음
- 로컬 source build에서 각 script를 두 번 생성하고 `test -s`, `cmp`, ANSI byte
  scan을 실행함. 네 script 모두 비어 있지 않고 결정적이며 ANSI가 없음
- Bash는 `bash -n`, session source, `complete -p soku`, root,
  `docs manual`, `--color`, `--profile` 실제 후보 검사를 통과함
- Zsh는 `zsh -n`, 격리된 `compinit -d`, script source, `_soku` 등록을
  통과함. Fish는 `fish -n`, source, `complete -C` 후보 검사를 통과함.
  PowerShell은 parser 검사, script 평가, `TabExpansion2` 후보 검사를 통과함
- Bash, Fish, PowerShell은 중첩 명령과 고정 color/profile 값을 직접 반환함.
  등록된 Zsh function이 사용하는 공통 Cobra completion endpoint도 같은
  `docs manual`, `completion`, color, profile 후보를 반환함
- 실제 pipe와 pseudo-TTY에서 `auto`, `always`, `never`, `NO_COLOR`,
  `TERM=dumb`를 확인함. JSON, quiet, stderr error, completion script에는
  ANSI가 없고 기존 stdout/stderr 소유권이 유지됨
- `git diff --check`: 통과

## 공개 적합성 검토

- [x] credential, token, private key, credential이 포함된 URL이 없음
- [x] 비공개 저장소·프로젝트·제품 이름이 없음
- [x] cloud project ID, 계정 번호, service URL, image URI, revision 식별자가 없음
- [x] 개인 청구·구독·budget·결제 상태 정보가 없음
- [x] 개인 이메일·전화번호·주소·로컬 절대 경로가 없음
- [x] 비공개 Issue·PR·Project·control-plane 식별자가 없음

## AI 지원

- **계획/구현/초안 작성:** OpenAI Codex
