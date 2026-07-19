# 📝 Task Report: Transactional `soku init`

## Goal and Background

[Issue #18](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/18)
implements the first end-to-end mutation workflow in the accepted `soku`
lifecycle. Issues #16, #17, and #19 now provide the lifecycle contract, CLI
boundary, portable manifest, atomic manifest storage, and read-only status
diagnostics that initialization depends on.

The goal is to initialize a new or existing downstream repository from one
explicit immutable boilerplate release. The command must detect or accept the
selected stacks, render only declared convention files, preserve project-owned
content, present the complete plan before writing, and either commit the whole
result with the manifest last or restore the exact pre-apply state.

## Proposed Approach

Add reusable source, catalog, selection, rendering, planning, verification, and
transaction packages beneath `soku/internal/`, then connect them to the existing
`init` handler. Keep source acquisition behind an interface so production code
can resolve a requested `vMAJOR.MINOR.PATCH` tag to a full commit and tests can
use a hermetic synthetic `v1.0.0` fixture without publishing a real release.

Each source release will contain a declarative `soku/catalog/core-v1.json`.
Catalog entries may declare only a supported stack ID, detection markers,
source and output paths, ownership class, content mode, merge strategy, and
bounded placeholders. They cannot declare commands, hooks, executables,
provider behavior, or writes outside the target repository.

Selection follows `CLI > explicit YAML config > detection > defaults`. Repeated
`--stack` flags replace detection completely. The v1 stack IDs are
`javascript-typescript-node`, `python`, `go`, `java-spring`, `mysql`,
`postgresql`, `gcp`, `aws`, and `azure`; the only profile is `standard`.
Configuration is validated and canonicalized before hashing. The manifest
stores only portable rendering values actually used by the selected stacks
(`project_name`, `module_path`, `java_group`, and `service_name`) so the desired
tree can be reproduced from the pinned source snapshot. Raw config,
verification options, credentials, and machine-local paths are never stored.

Rendering is deterministic and complete before mutation. Existing paths remain
project-owned by default. Only `.gitignore` and `.editorconfig` may use declared
deterministic line-based merge strategies; every other existing-path collision,
ownership conflict, unsafe path, symlink escape, case collision, Windows
reserved name, invalid placeholder, or secret-bearing source fails before the
first write with exit `4` or the more specific validation, compatibility, or
fetch exit code.

A confirmed apply uses one journaled transaction. It backs up every touched
path, applies the already validated plan, revalidates output hashes, and
atomically replaces `.soku/manifest.json` last through the manifest store. Any
failure restores all touched files and the previous manifest. Successful
rollback exits `7`; rollback failure exits `8` and retains bounded recovery
instructions. A conflict or cancelled confirmation creates no backup, journal,
manifest, or managed-file write.

`--verify` is disabled by default. When selected, the CLI chooses commands only
from a built-in stack allowlist and runs them against a temporary staging tree;
the source catalog cannot introduce executable content. `--dry-run` performs
source fetch, resolution, validation, rendering, conflict checks, optional
staging verification, and plan output without writing the target, backup,
journal, or manifest.

## Planned Implementation

- Extend `soku init` parsing with required `--boilerplate-source` and
  `--boilerplate-release`, repeatable `--stack`, `--profile standard`,
  `--project-name`, `--module-path`, `--java-group`, `--service-name`, and
  `--verify`, while preserving the common output and confirmation contract.
- Define and validate portable YAML configuration for the same values, reject
  unknown fields and secret-bearing inputs, and calculate one canonical
  configuration SHA-256 for the manifest.
- Implement immutable source resolution, fetch limits, archive validation, and
  an injected hermetic source fixture. Record both the requested release and
  resolved lowercase 40-character commit.
- Publish and validate `soku/catalog/core-v1.json` with the fixed stack IDs,
  marker rules, bounded template paths, ownership, content modes, merge rules,
  and placeholders.
- Add explicit-over-detected selection, deterministic template rendering,
  placeholder validation, stack-appropriate CI generation, and stable plan
  ordering.
- Add safe `.gitignore` and `.editorconfig` merge functions; treat all other
  existing paths as project-owned conflicts unless their desired content is an
  exact no-op.
- Add preflight checks for traversal, reserved state, symlink escape, file type,
  case-insensitive collisions, Windows-invalid components, ownership overlap,
  source secrets, compatibility, existing manifest selection, and reruns.
- Implement interactive confirmation, non-interactive `--yes`, dry-run, journal,
  backups, rollback, recovery guidance, manifest-last commit, and cleanup.
- Implement optional allowlisted verification in an isolated staging tree and
  report each command and result without accepting source-defined commands.
- Add unit, integration, packaging-smoke, and CI coverage across supported
  operating systems and stack fixtures; update CLI and lifecycle documentation.

## Acceptance Criteria

| Criterion | Observable evidence |
| --- | --- |
| Immutable, reproducible input | Init requires an HTTPS source plus `vMAJOR.MINOR.PATCH`, resolves and records a full commit, rejects floating or changed input, and uses hermetic fixtures before the #21 release gate. |
| Stable selection and configuration | Explicit stacks fully replace detection; config precedence is deterministic; only the nine stack IDs and `standard` profile are accepted; placeholder requirements are stack-specific; only used portable rendering values and their reproducible canonical hash are persisted. |
| Truthful plan before writes | Fetch, compatibility, schema, path, placeholder, ownership, rendering, merge, conflict, and optional verification checks finish before confirmation or any target/journal/backup/manifest write. |
| Non-destructive existing-repository behavior | Existing files are project-owned by default; only `.gitignore` and `.editorconfig` are deterministically mergeable; every other collision stops with exit `4` and zero writes. |
| Transactional apply and rollback | A confirmed apply journals and backs up bounded targets, writes the manifest last, returns `7` after successful rollback, and returns `8` with recovery evidence only when rollback itself fails. |
| Safe dry-run and confirmation | Dry-run performs the complete read-side workflow with zero target writes; cancellation is zero-write; non-interactive mutation requires `--yes`. |
| Idempotent lifecycle state | Repeating the same source and selection is a no-op success; a different selection in an existing manifest is refused with guidance to use `upgrade`. |
| Cross-platform and security coverage | Tests cover empty/existing/multi-stack repositories, every stack, placeholders, CI output, symlink and traversal attacks, case collisions, Windows paths, secret-bearing sources, cancellation, reruns, rollback, and packaging on Linux, macOS, and Windows. |
| Scope remains bounded | No real boilerplate tag or GitHub Release is created; `diff`, `upgrade`, provider loading, profiles beyond `standard`, and downstream release-gate E2E remain in Issues #20, #22, and #21. |

## Approval

- **Status:** `Approved`
- **Approved by:** `Soku-JINSEOK`
- **Approval record:** The repository owner's 2026-07-18 implementation
  instruction approved the decisions fixed in this report.

## Implementation Status

Complete. The approved report was implemented as `soku init` production code,
the core-v1 catalog and schema, portable manifest rendering inputs, hermetic
integration/security/failure tests, lifecycle documentation, and portable
packaging smoke coverage. Real boilerplate tags/Releases, `diff`, `upgrade`,
providers, and downstream release E2E remain in the follow-up issues.

## Verification

- `go mod verify` — passed.
- `go test ./...`, `go test -race ./...`, `go vet ./...` — passed.
- `golangci-lint` v2.12.2, `gofmt`, and `goimports` v0.48.0 — passed.
- Linux amd64/arm64, macOS amd64/arm64, and Windows amd64 cross-build — passed.
- `soku/scripts/package_test.sh` deterministic five-target package smoke — passed.
- JS/TS, Python, Go, and Java/Spring template lint/typecheck/test/build — passed.
- Core catalog and manifest Draft 2020-12 schema fixtures plus hermetic
  source/archive/transaction tests — passed.
- Manifest and pinned source snapshot desired-tree reproduction regression —
  passed.
- Full Markdown lint, GitHub/template YAML lint, and `git diff --check` —
  passed.
- `scripts/verify-sync-parity.sh` — local `pwsh` unavailable; Draft PR #29 CI
  passed.
- Draft PR #29 repository hygiene, Linux/macOS/Windows native, race/lint, and
  five-target package jobs — passed; the Release job was skipped by policy.

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex

---

## 목표 및 배경

[Issue #18](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/18)은
승인된 `soku` lifecycle의 첫 end-to-end mutation workflow를 구현합니다. 선행 Issues #16, #17과 #19에서는 lifecycle 계약, CLI 경계, portable manifest, atomic manifest 저장과
read-only status 진단을 마련했으므로 이제 명시적인 immutable boilerplate release로
신규 또는 기존 downstream 저장소를 안전하게 초기화할 수 있습니다.

목표는 선택된 stack만 결정적으로 렌더링하고 project-owned 내용을 보존하며, 모든
변경 계획을 쓰기 전에 제시한 뒤 manifest를 마지막에 commit하거나 적용 전 상태로
전체 복구하는 것입니다.

## 제안하는 접근

`soku/internal/` 아래에 source, catalog, selection, rendering, planning,
verification, transaction 계층을 공용 package로 추가하고 기존 `init` handler에
연결합니다. Production source는 요청한 `vMAJOR.MINOR.PATCH` tag를 full commit으로
해석하며, 테스트는 실제 release를 만들지 않는 synthetic `v1.0.0` hermetic fixture를
사용합니다.

Source release의 `soku/catalog/core-v1.json`은 고정 stack ID, marker, source/output
path, ownership, content mode, merge strategy와 제한된 placeholder만 선언합니다.
명령, hook, executable, provider 동작은 선언할 수 없습니다. 선택 우선순위는
`CLI > YAML config > detection > defaults`이며 명시적인 `--stack`은 detection을
완전히 대체합니다. V1 profile은 `standard` 하나뿐입니다.

기존 경로는 기본적으로 project-owned입니다. `.gitignore`와 `.editorconfig`만
결정적 line merge를 허용하며 다른 충돌은 첫 write 전에 exit `4`로 중단합니다.
확정된 적용은 journal과 backup을 만들고 계획을 적용·재검증한 뒤 manifest를 마지막에
atomic replace합니다. 실패 시 전체 rollback하며 rollback 성공은 exit `7`, rollback
실패만 recovery 정보와 exit `8`을 남깁니다.

`--verify`는 기본 비활성화이며 CLI 내장 stack allowlist의 명령만 임시 staging tree에서
실행합니다. `--dry-run`은 fetch와 전체 검증·rendering·plan·선택적 staging 검증까지
수행하지만 target, backup, journal, manifest를 쓰지 않습니다.

## 계획된 구현

- `soku init`에 source/release, stack/profile, placeholder, verify option과 portable
  YAML config를 추가하고 선택 stack이 사용한 portable rendering 값과 재현 가능한
  canonical configuration hash만 저장합니다.
- Immutable source 해석·제한된 fetch, core catalog schema, fixed stack detection,
  deterministic renderer와 stack별 CI 생성을 구현합니다.
- `.gitignore`·`.editorconfig` 전용 merge와 모든 path·ownership·secret·compatibility
  preflight를 구현합니다.
- interactive 취소, non-interactive `--yes`, dry-run, journal, backup, rollback,
  manifest-last commit, recovery guidance를 구현합니다.
- 동일 입력 rerun no-op과 다른 selection의 upgrade 안내 거부를 구현합니다.
- 모든 stack과 Linux/macOS/Windows에서 unit·integration·packaging smoke를 추가하고
  CLI/lifecycle 문서를 갱신합니다.

## 수용 기준

| 기준 | 관찰 가능한 근거 |
| --- | --- |
| Immutable input | HTTPS source와 semver release를 필수로 받아 full commit을 기록하고 floating input을 거부하며 #21 전에는 hermetic fixture만 사용 |
| 결정적 선택 | 명시 stack이 detection을 완전히 대체하고 고정 ID 9개와 `standard`만 허용하며 raw config 대신 사용한 portable rendering 값과 재현 가능한 canonical hash만 저장 |
| Zero-write preflight | Fetch부터 compatibility, path, placeholder, ownership, render, merge, conflict, 선택 검증까지 write 전에 완료 |
| 기존 저장소 보존 | 기본 project-owned, 두 공유 파일만 deterministic merge, 나머지 충돌은 exit `4`와 zero-write |
| Transaction/rollback | Manifest-last commit, 전체 rollback, rollback 성공 exit `7`, rollback 실패만 exit `8`과 recovery evidence |
| Dry-run/confirmation | Dry-run과 취소는 target 무변경, non-interactive mutation은 `--yes` 필수 |
| Idempotency | 같은 source·selection rerun은 no-op, 다른 selection은 `upgrade` 안내와 함께 거부 |
| 검증 범위 | 빈/기존/multi-stack 저장소, 모든 stack, placeholder, CI, symlink/traversal/case/Windows/secret 공격, 취소, rerun, rollback, 3개 OS package 검증 |
| 범위 제한 | 실제 tag/Release, `diff`, `upgrade`, provider, 조합형 profile, downstream E2E gate는 후속 Issue에 유지 |

## 승인

- **상태:** `Approved`
- **승인자:** `Soku-JINSEOK`
- **승인 기록:** 2026-07-18 구현 지시에 따라 본 report의 확정된 결정을 승인함

## 구현 현황

승인된 report를 기준으로 `soku init` production code, core-v1 catalog/schema,
hermetic integration/security/failure test, 문서와 portable packaging smoke를
완료했습니다. 실제 boilerplate tag/Release 생성, `diff`, `upgrade`, provider와
downstream release E2E는 후속 Issue 범위로 유지합니다.

## 검증

- `go mod verify` — 통과
- `go test ./...`, `go test -race ./...`, `go vet ./...` — 통과
- `golangci-lint v2.12.2`, `gofmt`, `goimports v0.48.0` — 통과
- Linux amd64/arm64, macOS amd64/arm64, Windows amd64 cross-build — 통과
- `soku/scripts/package_test.sh` five-target deterministic package smoke — 통과
- JS/TS, Python, Go, Java/Spring template lint/typecheck/test/build — 통과
- catalog Draft 2020-12 schema와 hermetic source/archive/transaction test — 통과
- 전체 Markdown lint, GitHub/template YAML lint, `git diff --check` — 통과
- `scripts/verify-sync-parity.sh` — 로컬 `pwsh` 부재; Draft PR #29 CI에서 통과
- Draft PR #29 repository hygiene, Linux/macOS/Windows native, race/lint,
  five-target package job — 통과 (Release job은 정책에 따라 skip)

## AI 지원

- **계획/구현/초안 작성:** OpenAI Codex
