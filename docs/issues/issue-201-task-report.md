# Issue #201 Task Report — Explicit project ownership handoff

## Goal and Background

Issue [#201](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/201)
requires a truthful, non-overwriting way to reconcile an intentionally modified
core-managed file. The current lifecycle correctly fails closed on local drift,
but it cannot record a reviewed downstream decision without overwriting the file
or fabricating its baseline.

The approved model is **project ownership handoff**. A handed-off path becomes
project-owned, and versioned manifest selection state suppresses future core
rendering until a separately designed explicit reclaim operation exists.

## Proposed Approach

Publish manifest schema v3 with the sorted
`selection.project_owned_overrides` path list and add this explicit command:

```text
soku ownership handoff \
  --path <canonical-relative-path> \
  --expected-sha256 <lowercase-64-hex> \
  --dry-run
```

The command accepts exactly one path and one expected normalized content hash.
It requires `--dry-run`, `--yes`, or interactive confirmation. It rejects stale
hashes and ineligible ownership classes before writing, never mutates the
selected file, and applies only a staged, validated, manifest-last transaction.

Manifest v1 and v2 remain readable and retain their exact serialization and
configuration-hash behavior. A successful handoff explicitly migrates either
schema to v3. Component installation preserves v3 instead of downgrading it.

## Planned Implementation

- Add the manifest-v3 JSON Schema, fixtures, validation, canonical hashing, and
  compatibility documentation.
- Add the `ownership handoff` parser, plan/result surface, confirmation flow,
  and transactional manifest-only mutation.
- Reject missing, symlinked, unreadable, non-regular, clean, obsolete,
  mergeable, provider-managed, already project-owned, stale-hash, reserved,
  non-canonical, case-mismatched, and repeated path input.
- Suppress v3 project-owned overrides during same-release diff and future core
  upgrades while preserving provider ownership conflict checks.
- Preserve v3 during docs-manual and GitHub Project Sync component installs.
- Add regression coverage for dry-run byte immutability, file bytes and mode,
  rollback/recovery, deterministic ordering, legacy manifests, same-release
  transitions, future releases, and a governance-ignore reproducer.
- Prepare the source-only `soku/v0.3.0` candidate note without changing the
  published release identity or creating a tag,
  Release, archive, checksum file, npm publication, or downstream mutation.

## Acceptance Criteria

- A single-field baseline edit cannot represent a reviewed handoff.
- Every invalid or stale request fails before writes.
- Dry-run is byte-for-byte non-mutating.
- Apply preserves the selected file's bytes and permission mode.
- Manifest replacement is deterministic, atomic, validated, and rollback-safe.
- Status and same-release/provider-qualified diff treat the handed-off path as
  intentionally unmanaged rather than drift.
- Future core releases cannot overwrite a handed-off path.
- Existing manifest v1/v2 fixtures and configuration hashes remain unchanged.
- Component installation never downgrades manifest v3.
- The full local lifecycle, schema, quality, security, and diff checks pass;
  hosted final-head checks remain mandatory before merge.

## Approval

- **Status:** `Approved`
- **Approved by:** `Soku-JINSEOK` through the repository closeout execution plan
  supplied on 2026-08-08

## Implementation Status

The source implementation and local acceptance fixture are complete on the
dedicated Issue #201 branch. Final-head hosted Linux/macOS/Windows, Security,
policy, and supply-chain checks remain required before merge. Tag, Release,
package publication, downstream merge, Provider apply, cloud, credential,
ruleset, and delivery mutations remain outside this task.

## Verification

Passed locally on 2026-08-08:

- `go test ./...` and `go vet ./...`
- focused ownership, manifest, status, manual, and CLI test packages
- `goimports -l .` with no output and `golangci-lint v2.12.2` with `0 issues`
- `scripts/run_lifecycle_gate.sh` with the hermetic lifecycle gate passing
- `scripts/package_test.sh` with all five archives and checksums reproducible
- working-tree Gitleaks `v8.24.2` with no leaks and `govulncheck v1.6.0` with no
  vulnerabilities
- `git diff --check`
- Node.js, Python, and Go template quality gates
- an isolated report-hub reproduction: dry-run changed no bytes, confirmed
  apply changed only `.soku/manifest.json`, `.prettierignore` bytes and mode
  remained exact, `status` was clean, and the exact registered Provider diff
  was a no-op with delivery disabled

Local race execution is unavailable because this runner has no C compiler, and
the Java template gate is unavailable because Maven is not installed. The
repository-wide regression runner also retains failures reproduced unchanged
from clean `main` in `detect-verification-scope.test.mjs` and `verify.test.mjs`;
they are not caused by this branch. Hosted final-head checks are authoritative
for those environment-dependent and cross-platform gates.

## Public Disclosure Review

- [x] No credentials, tokens, private keys, or credential-bearing URLs
- [x] No new private repository, project, or product identifiers beyond the
      approved reproducer already recorded in Issue #201
- [x] No cloud project IDs, account numbers, service URLs, image URIs, or
      revision identifiers
- [x] No personal billing, subscription, budget, or payment-status information
- [x] No personal email, phone, address, or local absolute path
- [x] No new private Issue, PR, Project, or control-plane identifiers

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex

---

## 목표 및 배경

Issue [#201](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/201)은
의도적으로 변경된 core-managed 파일을 덮어쓰지 않고 진실하게 조정하는 방법을
요구합니다. 현재 lifecycle은 local drift에 올바르게 fail closed하지만 파일을
덮어쓰거나 baseline을 조작하지 않고 검토된 downstream 결정을 기록할 수 없습니다.

승인된 모델은 **project ownership handoff**입니다. Handoff된 경로는 project-owned가
되고 versioned manifest selection state가 향후 core rendering을 억제합니다. Core
ownership reclaim은 별도의 명시적 operation이 설계되기 전까지 허용하지 않습니다.

## 제안하는 접근

정렬된 `selection.project_owned_overrides` path 목록을 갖는 manifest schema v3를
공개하고 다음 명시적 command를 추가합니다.

```text
soku ownership handoff \
  --path <canonical-relative-path> \
  --expected-sha256 <lowercase-64-hex> \
  --dry-run
```

Command는 정확히 한 path와 예상 normalized content hash 하나만 허용합니다.
`--dry-run`, `--yes`, interactive confirmation 중 하나를 요구하며 stale hash와
부적합 ownership class를 write 전에 거부합니다. 선택된 파일은 전혀 변경하지 않고
staged·validated·manifest-last transaction으로 manifest만 적용합니다.

Manifest v1/v2는 계속 읽을 수 있고 기존 serialization과 configuration hash 동작을
그대로 유지합니다. 성공한 handoff만 schema v3로 명시적으로 전환합니다. Component
설치는 v3를 v2로 downgrade하지 않습니다.

## 계획된 구현

- Manifest-v3 JSON Schema, fixture, validation, canonical hashing, compatibility
  문서를 추가합니다.
- `ownership handoff` parser, plan/result, confirmation, manifest-only transaction을
  추가합니다.
- missing, symlink, unreadable, non-regular, clean, obsolete, mergeable,
  provider-managed, already project-owned, stale hash, reserved, non-canonical,
  case mismatch, repeated path를 거부합니다.
- Same-release diff와 future core upgrade에서 v3 override rendering을 억제하면서
  provider ownership conflict 검사를 유지합니다.
- docs-manual 및 GitHub Project Sync component 설치가 v3를 보존하게 합니다.
- Dry-run byte 불변, file bytes/mode, rollback/recovery, deterministic ordering,
  legacy manifest, same-release transition, future release, governance-ignore
  reproducer 회귀 검증을 추가합니다.
- Tag, Release, archive, checksum, npm publication, downstream 변경 없이
  `soku/v0.3.0` release identity와 notes를 준비합니다.

## 수용 기준

- 단일 baseline field 수정으로 검토된 handoff를 표현할 수 없습니다.
- Invalid 또는 stale 요청은 모두 write 전에 실패합니다.
- Dry-run은 byte-for-byte non-mutating입니다.
- Apply는 선택 파일 bytes와 permission mode를 보존합니다.
- Manifest replacement는 deterministic, atomic, validated, rollback-safe입니다.
- Status와 same-release/provider-qualified diff가 handoff path를 drift가 아니라
  의도적으로 unmanaged된 상태로 표현합니다.
- 향후 core release가 handoff path를 덮어쓸 수 없습니다.
- 기존 manifest v1/v2 fixture와 configuration hash가 변하지 않습니다.
- Component 설치가 manifest v3를 downgrade하지 않습니다.
- 전체 local lifecycle, schema, quality, security, diff 검사가 통과하며 merge 전
  hosted final-head check를 반드시 통과합니다.

## 승인

- **상태:** `Approved`
- **승인자:** 2026-08-08 제공된 전체 저장소 closeout 실행 계획을 통한
  `Soku-JINSEOK`

## 구현 현황

전용 Issue #201 branch에서 source 구현과 local acceptance fixture를 완료했습니다.
Merge 전 최종 head의 hosted Linux/macOS/Windows, Security, policy, supply-chain
check가 필요합니다. Tag, Release, package publication, downstream merge, Provider
apply, cloud, credential, ruleset, delivery mutation은 이 작업 범위 밖에 유지합니다.

## 검증

2026-08-08 local 검증에서 전체 Go test와 vet, ownership 관련 package test,
goimports, golangci-lint, hermetic lifecycle gate, 5개 archive 재현성, working-tree
Gitleaks, govulncheck, diff check, Node.js/Python/Go template gate가 통과했습니다.
격리된 report-hub 재현에서는 dry-run이 byte 불변이었고 apply가 manifest만
변경했으며 `.prettierignore` bytes/mode를 보존했습니다. 이후 status는 clean,
등록된 정확한 Provider diff는 delivery disabled 상태에서 no-op이었습니다.

현재 runner에는 C compiler와 Maven이 없어 race 및 Java template 검증을 실행할 수
없습니다. 또한 repository-wide regression runner의 두 실패는 clean `main`에서도
동일하게 재현됩니다. 이 환경 및 cross-platform 항목은 최종 head hosted check를
권위 있는 결과로 사용합니다.

## 공개 적합성 검토

- [x] credential, token, private key, credential이 포함된 URL이 없음
- [x] Issue #201에 이미 승인되어 기록된 reproducer 외 새로운 비공개
      저장소·프로젝트·제품 식별자가 없음
- [x] cloud project ID, 계정 번호, service URL, image URI, revision 식별자가 없음
- [x] 개인 청구·구독·budget·결제 상태 정보가 없음
- [x] 개인 이메일·전화번호·주소·로컬 절대 경로가 없음
- [x] 새로운 비공개 Issue·PR·Project·control-plane 식별자가 없음

## AI 지원

- **계획/구현/초안 작성:** OpenAI Codex
