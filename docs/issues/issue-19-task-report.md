# 📝 Task Report: Portable Manifest and `soku status`

## Goal and Background

[Issue #19](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/19)
implements the durable, portable state record required by the lifecycle contract
and replaces the `status` placeholder added by Issue #17. Issues #16 and #17 are
complete, so the manifest can now be implemented without changing the approved
command surface or mutation safety contract.

## Proposed Approach

Publish a JSON Schema Draft 2020-12 manifest-v1 contract and representative
fixtures under `soku/`. Implement a validating manifest package that owns
canonical paths, secret-safe portable fields, deterministic serialization,
content hashing, interrupted-write detection, atomic replacement, and explicit
recovery. Implement `status` as a read-only comparison between the validated
snapshot and the current repository, with stable human and JSON diagnostics.

The manifest records immutable source revisions and portable selections, never
raw configuration, credentials, or machine-specific absolute paths. `status`
does not fetch desired state and never repairs or removes an interrupted write;
it reports the recovery requirement for a later mutation or explicit recovery.

## Planned Implementation

- Add manifest-v1 Go types, validation, canonical JSON, text/binary hashing,
  atomic storage, pending-file inspection, and recovery behavior.
- Publish the Draft 2020-12 schema plus valid and invalid fixtures.
- Extend the handler result boundary so `status` can return human text, ordered
  JSON data, and exit codes while preserving the existing envelope and parser.
- Detect clean, missing, changed, obsolete, unmanaged-expected, pending,
  incompatible, malformed, symlink-escape, type-mismatch, and unreadable state.
- Add table-driven unit and command tests for schema, security, storage,
  recovery, filesystem diagnostics, output modes, handler failures, and panic.
- Update lifecycle and CLI documentation, repository hygiene, and the
  cross-platform native smoke contract without adding `soku/` to manual sync.

## Acceptance Criteria

| Criterion | Observable evidence |
| --- | --- |
| Manifest v1 is portable, versioned, deterministic, and secret-safe. | Schema, fixtures, Go validation, and serialization tests agree; forbidden paths, raw configuration, credential-bearing fields, and collisions are rejected. |
| Baselines are stable across platforms. | Text hashing validates UTF-8 and normalizes CRLF/CR only; binary hashing preserves bytes. |
| Manifest replacement and recovery are durable and explicit. | Store tests cover pending writes, atomic replacement, damaged states, promotion rules, and preservation of ambiguous evidence. |
| `status` is read-only and actionable. | Tests cover all file and integration states, recovery-required diagnostics, and verify that filesystem contents do not change. |
| Output and exit semantics remain stable. | Human, quiet, and single-envelope JSON tests cover exit codes `0`, `1`, `2`, `3`, and `5`; diagnostic `3`/`5` results use `ok: true`. |
| Deferred lifecycle commands stay deferred. | `init`, `diff`, and `upgrade` continue returning `feature.unavailable` with exit code `5`. |
| CI and documentation expose the implemented contract. | Native smoke expects uninitialized `status` exit `3`; documentation and hygiene link the schema, fixtures, status meanings, and recovery procedure. |

## Approval

- **Status:** `Approved`
- **Approved by:** `Soku-JINSEOK` (the repository owner supplied the reviewed
  Issue #19 implementation plan and explicitly requested implementation)

## Implementation Status

Implemented and locally validated on `agent/implement-soku-manifest`. The
approved task report was added before implementation. The repository owner
subsequently authorized commit, push, and Draft pull request publication.
Merge, tagging, and release remain outside this approval.

## Verification

- `go mod verify`, `go test ./...`, and `go vet ./...` — passed.
- Manifest tests compile the published Draft 2020-12 schema and verify every
  valid and invalid fixture against both the schema and semantic validator.
- `gofmt` and `goimports@v0.48.0` — no files required changes.
- `golangci-lint@v2.12.2 run ./...` — 0 issues.
- Windows amd64 and macOS arm64 cross-builds — passed; native store tests cover
  platform-specific replacement in CI.
- `soku/scripts/package_test.sh` — all five archives, checksums, packaged
  execution, and deterministic rerun passed.
- Native uninitialized `status --json` smoke — one `ok: true` envelope and exit
  `3`, with no repository mutation.
- Contribution-title tests, Markdown lint, YAML lint, shell syntax, repository
  hygiene, sync parity, and `git diff --check` — passed.
- `go test -race ./...` could not run locally because this environment has
  `CGO_ENABLED=0` and no C compiler; the existing Linux CI race job remains the
  required native race verification.

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex

---

## 목표 및 배경

[Issue #19](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/19)는
lifecycle 계약에 필요한 영속적이고 portable한 상태 기록을 구현하고
Issue #17에서 추가한 `status` placeholder를 교체합니다. Issue #16과 Issue #17이 완료되어
승인된 command surface와 mutation safety 계약을 바꾸지 않고 manifest를 구현할
수 있습니다.

## 제안하는 접근

JSON Schema Draft 2020-12 기반 manifest-v1 계약과 대표 fixture를 `soku/` 아래에
게시합니다. Canonical path, secret-safe portable field, 결정적 직렬화, content
hash, 중단된 write 탐지, atomic replacement와 명시적 recovery를 담당하는 manifest
package를 구현합니다. `status`는 검증된 snapshot과 현재 저장소를 읽기 전용으로
비교하고 안정된 human·JSON 진단을 반환합니다.

Manifest에는 immutable source revision과 portable selection만 기록하며 raw
configuration, credential, machine별 absolute path는 저장하지 않습니다.
`status`는 desired state를 fetch하지 않고 중단된 write를 복구하거나 삭제하지
않으며 후속 mutation 또는 명시적 recovery가 필요하다고만 보고합니다.

## 계획된 구현

- manifest-v1 Go type, validation, canonical JSON, text/binary hashing, atomic
  store, pending 상태 검사와 recovery 동작을 추가합니다.
- Draft 2020-12 schema와 valid·invalid fixture를 게시합니다.
- 기존 envelope와 parser를 유지하면서 `status`가 human text, 정렬된 JSON data,
  exit code를 반환하도록 handler result 경계를 확장합니다.
- clean, missing, changed, obsolete, unmanaged-expected, pending, incompatible,
  malformed, symlink escape, type mismatch, unreadable 상태를 탐지합니다.
- schema, 보안, storage, recovery, filesystem 진단, 출력 mode, handler 오류·panic을
  table-driven unit·command test로 검증합니다.
- `soku/`를 manual sync에 추가하지 않고 lifecycle·CLI 문서, repository hygiene,
  cross-platform native smoke 계약을 갱신합니다.

## 수용 기준

| 기준 | 관찰 가능한 근거 |
| --- | --- |
| Manifest v1이 portable하고 versioned·deterministic·secret-safe함 | Schema, fixture, Go validation과 serialization test가 일치하고 금지 path, raw configuration, credential field, collision을 거부 |
| Baseline이 platform 간 안정적임 | Text hash는 UTF-8을 검증하고 CRLF/CR만 정규화하며 binary hash는 byte를 그대로 보존 |
| Manifest 교체와 recovery가 durable하고 명시적임 | Store test가 pending write, atomic replacement, 손상 상태, 승격 규칙과 모호한 증거 보존을 검증 |
| `status`가 read-only이고 actionable함 | 모든 file·integration 상태와 recovery-required 진단을 검증하고 filesystem이 바뀌지 않음을 확인 |
| 출력과 exit 의미가 안정적임 | Human, quiet, single-envelope JSON test가 exit `0`, `1`, `2`, `3`, `5`를 검증하고 진단 결과 `3`/`5`는 `ok: true` 사용 |
| 연기된 lifecycle command가 계속 연기됨 | `init`, `diff`, `upgrade`는 exit `5`의 `feature.unavailable`을 계속 반환 |
| CI·문서가 구현 계약을 공개함 | Native smoke가 미초기화 `status` exit `3`을 기대하고 문서·hygiene이 schema, fixture, status 의미, recovery 절차를 연결 |

## 승인

- **상태:** `Approved`
- **승인자:** `Soku-JINSEOK` (저장소 소유자가 검토된 Issue #19 구현 계획을
  제공하고 구현을 명시적으로 요청함)

## 구현 현황

`agent/implement-soku-manifest`에서 구현과 로컬 검증을 완료했습니다. 승인된 task
report를 구현보다 먼저 추가했습니다. 저장소 소유자가 이후 commit, push, Draft
PR 게시를 승인했습니다. Merge, tag, release는 이번 승인 범위 밖에 유지합니다.

## 검증

- `go mod verify`, `go test ./...`, `go vet ./...` — 통과
- Manifest test가 공개된 Draft 2020-12 schema를 compile하고 모든 valid·invalid
  fixture를 schema와 semantic validator 양쪽으로 검증 — 통과
- `gofmt`, `goimports@v0.48.0` — 변경 필요 파일 없음
- `golangci-lint@v2.12.2 run ./...` — issue 0건
- Windows amd64·macOS arm64 cross-build — 통과, platform별 replacement는 CI의
  native store test에서 검증
- `soku/scripts/package_test.sh` — archive 5개, checksum, package 실행, 결정적
  재실행 통과
- 미초기화 native `status --json` smoke — repository 변경 없이 `ok: true`
  envelope 하나와 exit `3` 확인
- Contribution title, Markdown, YAML, shell syntax, repository hygiene, sync
  parity, `git diff --check` — 통과
- `go test -race ./...`는 이 환경이 `CGO_ENABLED=0`이고 C compiler가 없어 로컬
  실행 불가. 기존 Linux CI race job에서 native race 검증 필요

## AI 지원

- **계획/구현/초안 작성:** OpenAI Codex
