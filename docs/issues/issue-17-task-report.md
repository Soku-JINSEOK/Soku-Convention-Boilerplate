# 📝 Task Report: `soku` CLI Shell and Distribution

## Goal and Background

[Issue #17](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/17)
implements the first executable shell for the `soku` lifecycle contract defined
by Issue #16. The command and output layer must be stable before later issues
add manifests, planning, mutation, rollback, or provider behavior.

This task starts from merge commit `14bd3ff` (PR #26). Issue #16 remains a
separate pull request and commit history. Issue #17 will remain as uncommitted
local work on its dedicated branch until the repository owner separately asks
for publication.

## Proposed Approach

Add a Go 1.26 submodule under `soku/` with a deliberately small executable
entrypoint and an injectable `internal/cli` package. Cobra owns parsing, while
runtime adapters own filesystem and terminal observations and lifecycle
handlers own future behavior. The four lifecycle handlers initially return a
stable `feature.unavailable` compatibility error so later issues can replace
behavior without redesigning parsing or output.

Use one deterministic output contract. Human mode separates normal output and
diagnostics between stdout and stderr. JSON mode is detected from raw arguments
before parsing and emits exactly one ordered envelope on stdout with empty
stderr, including for parse errors. Package the same binary reproducibly for
five supported OS/architecture targets and exercise that path in CI.

## Planned Implementation

- Add the Go module, Cobra dependencies, CLI runtime and handler boundaries,
  build metadata resolution, typed exit errors, commands, flags, and output.
- Add table-driven unit and integration coverage for parsing, streams, safety
  flags, terminal behavior, configuration validation, handler failures, panic
  recovery, JSON determinism, local installation, and packaged execution.
- Add a deterministic packaging script, dependency notices, and documentation
  for build, test, install, checksum, and signed CLI release operation.
- Extend CI with cross-platform tests and smoke checks, Linux static analysis,
  pull-request packaging verification, and guarded `soku/v*` releases.
- Register `soku/` in the repository architecture, public document indexes,
  applicability matrix, and hygiene checks without adding it to manual sync.
- Add an explicit sync-parity assertion that downstream output excludes
  `soku/`.

## Acceptance Criteria

| Criterion | Observable evidence |
| --- | --- |
| The public surface contains only `init`, `status`, `diff`, `upgrade`, `--help`, and `--version`. | Command tests reject completion and a `help` command while help lists only the supported surface. |
| Common and mutation flags preserve the approved safety rules. | Table-driven tests cover TTY and non-TTY execution, explicit non-interactive mode, `--yes`, `--dry-run`, invalid combinations, and configuration paths. |
| Exit codes and streams are stable in human and JSON modes. | Tests cover codes `1`, `2`, and `5`, deterministic envelopes, raw `--json` detection, quiet output, and stdout/stderr separation. |
| Future lifecycle behavior is replaceable without parser changes. | Four injected handlers are independently tested and default to `feature.unavailable` with exit code `5`. |
| Build metadata works in development and release builds. | Unit and smoke tests cover ldflags, build-info fallback, and `dev`/`unknown` fallbacks. |
| Five release archives are reproducible and complete. | Packaging tests validate names, binary modes, license files, checksums, reruns, and the Linux amd64 packaged binary. |
| CI and documentation describe the independent CLI release axis. | CI includes cross-platform and guarded release jobs; repository indexes and release/sync policy identify `soku/` and `soku/v*`. |
| Manual boilerplate sync still excludes the CLI source. | Sync parity explicitly fails if either downstream result contains `soku/`. |

## Approval

- **Status:** `Approved`
- **Approved by:** `Soku-JINSEOK` (the repository owner supplied and requested
  implementation of the reviewed Issue #16/#17 execution plan before changes
  began)

## Implementation Status

Complete on `agent/implement-soku-cli`. The approved task report was added
before implementation files. The repository owner subsequently authorized
final validation, commit, push, and Draft pull request publication. Draft PR #27
contains the implementation, and its initial CI run passed. No CLI tag or GitHub
Release is part of this work.

## Verification

- `go mod verify` — all modules verified.
- `go test ./...` — all CLI, metadata, stream, runtime, handler, JSON, and
  temporary-`GOBIN` installation tests passed.
- `go vet ./...` — passed.
- `gofmt` and `goimports@v0.48.0` — no files required changes.
- `golangci-lint@v2.12.2 run ./...` — 0 issues.
- `scripts/package_test.sh` — five expected archives, archive contents,
  executable mode, sorted SHA-256 entries, checksum recalculation, packaged
  Linux amd64 execution, deterministic rerun, and tracked-tree preservation
  passed.
- Native Linux smoke — human help, development version, and JSON parse-error
  stream separation passed.
- TTY safety regression — a null device and pipe are rejected as terminals;
  `soku init </dev/null` now stops with validation exit code `2`.
- `markdownlint-cli2@0.22.1` and `yaml-lint@1.7.0` — passed.
- Contribution-title regression tests — passed.
- `bash -n scripts/*.sh soku/scripts/*.sh` and `git diff --check` — passed.
- Repository-hygiene required-file reconstruction — every registered path
  exists.
- Sync parity — passed with checksum-verified portable PowerShell 7.6.3. Both
  current and post-commit-simulated tracked trees produced identical bash and
  PowerShell output, no leaked artifacts, and no downstream `soku/` directory.
- `go test -race ./...` could not run locally because this environment has
  `CGO_ENABLED=0` and no C compiler.
- Draft PR #27 CI — repository hygiene, sync parity, Linux/macOS/Windows native
  test/vet/build/smoke, Ubuntu race/static analysis, and five-target packaging
  all passed. The release job was correctly skipped for the non-tag pull
  request.

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex

---

## 목표 및 배경

[Issue #17](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/17)은
Issue #16에서 정의한 `soku` lifecycle 계약의 첫 실행 가능한 CLI shell을
구현합니다. 후속 Issue가 manifest, plan, mutation, rollback, provider 동작을
추가하기 전에 command와 출력 계층을 안정적인 계약으로 고정해야 합니다.

이 작업은 PR #26의 merge commit `14bd3ff`에서 시작합니다. Issue #16은 별도
PR과 commit 이력으로 유지합니다. Issue #17은 저장소 소유자가 별도로 게시를
요청하기 전까지 전용 branch의 commit하지 않은 로컬 변경으로 남깁니다.

## 제안하는 접근

`soku/` 아래에 Go 1.26 submodule을 만들고 작은 실행 진입점과 주입 가능한
`internal/cli` package를 둡니다. Cobra는 parsing, runtime adapter는 filesystem과
terminal 관찰, lifecycle handler는 후속 동작을 담당합니다. 네 lifecycle
handler는 처음에는 안정적인 `feature.unavailable` compatibility 오류를 반환하여
후속 Issue가 parsing·출력 계층을 바꾸지 않고 동작만 교체할 수 있게 합니다.

출력 계약은 하나로 고정합니다. Human mode는 정상 출력과 진단을 stdout과
stderr로 나눕니다. JSON mode는 parsing 전에 raw argument에서 탐지하며 parse
오류를 포함해 stdout에 순서가 고정된 envelope 하나만 출력하고 stderr는
비웁니다. 동일 binary를 지원하는 OS/architecture 5개 조합으로 재현 가능하게
package하고 CI에서 전체 경로를 검증합니다.

## 계획된 구현

- Go module, Cobra dependency, CLI runtime·handler 경계, build metadata 해석,
  typed exit error, command, flag, 출력을 추가합니다.
- parsing, stream, safety flag, terminal 동작, config 검증, handler 오류·panic,
  JSON 결정성, local install, package 실행을 table-driven unit·integration test로
  검증합니다.
- 재현 가능한 packaging script, dependency notice와 build·test·install·checksum·
  signed CLI release 운영 문서를 추가합니다.
- CI에 cross-platform test·smoke, Linux 정적 분석, PR package 검증, 보호된
  `soku/v*` release job을 추가합니다.
- manual sync 대상에는 포함하지 않은 채 repository 구조, 다국어 문서 색인,
  applicability matrix, hygiene 검사에 `soku/`를 등록합니다.
- downstream 결과에 `soku/`가 없음을 sync parity가 명시적으로 검사하게 합니다.

## 수용 기준

| 기준 | 관찰 가능한 근거 |
| --- | --- |
| 공개 surface는 `init`, `status`, `diff`, `upgrade`, `--help`, `--version`만 포함 | help는 지원 surface만 표시하고 completion과 `help` command 호출은 test에서 거부 |
| 공통·mutation flag가 승인된 safety rule을 보존 | TTY/non-TTY, 명시적 non-interactive, `--yes`, `--dry-run`, 잘못된 조합, config path를 table-driven test로 검증 |
| Human·JSON mode의 exit code와 stream이 안정적 | code `1`, `2`, `5`, deterministic envelope, raw `--json` 탐지, quiet, stdout/stderr 분리 검증 |
| lifecycle 동작을 parser 변경 없이 교체 가능 | 주입된 handler 4개를 독립적으로 검증하고 기본값은 exit `5`의 `feature.unavailable` 반환 |
| 개발·release build metadata 동작 | ldflags, build-info fallback, `dev`/`unknown` fallback unit·smoke test |
| release archive 5개가 재현 가능하고 완전함 | 이름, binary mode, license, checksum, 재실행, Linux amd64 packaged binary 검증 |
| CI·문서가 독립 CLI release axis를 설명 | cross-platform·guarded release job과 `soku/`, `soku/v*` 문서화 |
| manual boilerplate sync가 CLI source를 계속 제외 | 두 downstream 결과에 `soku/`가 있으면 sync parity 실패 |

## 승인

- **상태:** `Approved`
- **승인자:** `Soku-JINSEOK` (저장소 소유자가 변경 시작 전에 검토된 Issue
  #16/#17 실행 계획을 제공하고 구현을 요청함)

## 구현 현황

`agent/implement-soku-cli`에서 완료했습니다. 구현 파일보다 먼저 승인된 task
report를 추가했습니다. 저장소 소유자가 이후 최종 검증, commit, push, Draft PR
게시를 승인했습니다. Draft PR #27에 구현을 게시했고 초기 CI가 통과했습니다.
CLI tag와 GitHub Release는 이 작업 범위에 포함하지 않습니다.

## 검증

- `go mod verify` — 전체 module 검증 완료
- `go test ./...` — CLI, metadata, stream, runtime, handler, JSON, 임시 `GOBIN`
  설치 test 통과
- `go vet ./...` — 통과
- `gofmt`, `goimports@v0.48.0` — 변경 필요 파일 없음
- `golangci-lint@v2.12.2 run ./...` — issue 0건
- `scripts/package_test.sh` — archive 5개, 내용, 실행 mode, 정렬된 SHA-256,
  checksum 재계산, Linux amd64 package 실행, 결정적 재실행, tracked tree 보존
  통과
- Native Linux smoke — human help, 개발 version, JSON parse error의 stream 분리
  통과
- TTY safety regression — null device와 pipe를 terminal로 인식하지 않으며
  `soku init </dev/null`은 validation exit code `2`로 중단
- `markdownlint-cli2@0.22.1`, `yaml-lint@1.7.0` — 통과
- contribution-title regression test — 통과
- `bash -n scripts/*.sh soku/scripts/*.sh`, `git diff --check` — 통과
- repository-hygiene required file 재구성 — 등록된 모든 경로 존재
- Sync parity — 공식 checksum을 검증한 portable PowerShell 7.6.3으로 통과.
  현재 tree와 commit 이후를 모사한 tracked tree에서 bash·PowerShell 결과 일치,
  artifact 누출 없음, downstream `soku/` 없음 확인
- `go test -race ./...`는 이 환경이 `CGO_ENABLED=0`이고 C compiler가 없어 로컬
  실행할 수 없었습니다.
- Draft PR #27 CI — repository hygiene, sync parity, Linux/macOS/Windows native
  test·vet·build·smoke, Ubuntu race·정적 분석, 5-target packaging 모두 통과.
  Non-tag PR이므로 release job은 의도대로 skipped

## AI 지원

- **계획/구현/초안 작성:** OpenAI Codex
