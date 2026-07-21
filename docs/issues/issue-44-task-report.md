# 📝 Issue 44 Task Report

## Goal and Background

[Issue #44](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/44)
requires Provider API v1 to trust the exact immutable
`--integration-ref` fetched by the CLI as the only authoritative provider
revision. The optional provider-supplied `ref` remains accepted only as
deprecated legacy metadata and must not decide which revision was fetched or
whether an integration becomes connected.

## Proposed Approach

Preserve the Provider API major version, manifest schema, CLI flags, and
`IntegrationFetcher` interface. Make `ref` optional in the Provider API v1
schema and decoder while retaining its lowercase 40-character SHA validation
when present. Remove it from connection decisions, and continue to persist the
exact CLI-requested commit in request artifacts and manifests.

Document a separate provider-owned submission channel for sanitized
configuration. Lifecycle state remains hash-only and never stores sanitized or
raw configuration, credentials, or secrets. Add hermetic regression coverage
and an opt-in real-Git conformance test pinned to a public immutable commit.

Prepare, but do not publish, the `soku/v0.1.2` compatibility and release
preflight record. Existing tags and Releases remain immutable.

## Planned Implementation

1. Remove `ref` from the Provider API v1 required fields, retain its deprecated
   annotation and format constraint, and remove it from the default provider
   example.
2. Keep Go decoding compatible with an omitted or well-formed legacy `ref`,
   reject malformed values and unknown fields, and stop comparing bundle
   metadata with the fetched revision during integration planning.
3. Extend unit and lifecycle tests for no-ref and legacy-ref bundles, fetched
   revision persistence, pending and connected state, mismatches, ownership,
   unsupported profiles, and no-write failures.
4. Add a separately runnable network conformance test that fetches the public
   AI collaboration provider at commit
   `a81f7c91b0c9c8faa5ba2988fde29e9d17972a83`, with CI coverage on Linux,
   macOS, and Windows using read-only credentials and a timeout.
5. Clarify the lifecycle trust boundary and sanitized configuration submission
   procedure without expanding the five-field pending artifact.
6. Add `docs/releases/soku-v0.1.2.md` with compatibility, migration, and
   validation-only release readiness evidence while keeping the published
   stable version at `soku/v0.1.1`.
7. Run the complete local and hosted validation gates, merge through the
   protected merge-commit workflow, verify `main`, and record evidence on
   Issue #44 without closing it or publishing `soku/v0.1.2`.

## Acceptance Criteria

- Bundles without `ref` and bundles with a well-formed matching or mismatching
  legacy `ref` decode successfully; malformed `ref` and unknown fields fail
  with validation exit code `5`.
- Matching or mismatching legacy metadata cannot alter pending or connected
  decisions. Request artifacts and manifests persist exactly the lowercase
  full SHA supplied through `--integration-ref` and used for fetch.
- Configuration-hash or source mismatches remain pending. Unsupported profiles
  fail with exit code `5`, ownership conflicts fail with exit code `4`, and
  these failures do not write lifecycle state or output.
- Pending artifacts contain only schema version, provider ID, portable source,
  authoritative ref, and configuration hash; no raw or sanitized configuration
  or secret is persisted.
- Hermetic lifecycle coverage remains independent, while required three-OS CI
  conformance fetches the pinned external provider through HTTPS with a
  read-only token and timeout.
- Provider-v1 compatibility, no-ref and legacy migration, unchanged
  manifest-v1, and boilerplate `v1.0.0` compatibility are documented for the
  proposed `soku/v0.1.2` patch.
- The strict aggregate `Validation Gate`, post-merge `main` gate, and manual
  validation-only release preflight pass.
- No `soku/v0.1.2` tag or GitHub Release is created, and Issue #44 remains open
  with `status:in-progress` pending separate publication approval.

## Approval

- **Status:** `Approved`
- **Approved by:** `Soku-JINSEOK` (explicit implementation instruction recorded
  on Draft PR #50 on 2026-07-21)
- **Approval boundary:** Implementation must not begin until explicit approval
  of this report is recorded on Issue #44 or its Draft PR. Publishing a tag or
  Release and closing Issue #44 require separate later approval.

## Implementation Status

Provider schema, decoder, revision planning, examples, hermetic and pinned HTTPS
conformance, lifecycle documentation, CI, and the proposed `soku/v0.1.2` record
were implemented and merged in pull request #50. PR Validation run
`29755588302`, post-merge `main` Validation run `29756161515`, and validation-only
release preflight `29756307978` passed. The `soku/v0.1.2` tag and GitHub Release
remain absent pending separate publication approval, so Issue #44 remains open.

## Verification

- Passed: Markdown lint for this task report.
- Passed: YAML lint for the required-file registration.
- Passed: `git diff --check` for the initial report-only change.
- Passed: all Go unit tests, race tests, vet, gofmt, goimports, and
  golangci-lint.
- Passed: hermetic lifecycle conformance and the opt-in HTTPS provider test at
  commit `a81f7c91b0c9c8faa5ba2988fde29e9d17972a83`.
- Passed: Markdown, YAML, actionlint, contribution-title, release-note contract,
  release-tag regression, Bash syntax, and final whitespace checks.
- Local environment limit: PowerShell and ShellCheck are unavailable; the
  Homebrew Go 1.26.5 installation lacks cross-compile standard-library packages
  required by the five-target package snapshot. Hosted validation covers all
  three checks.
- Passed: PR run `29755390115`, including PowerShell/shell parity, ShellCheck,
  the five-target package snapshot, three-OS native Go and hermetic/network
  lifecycle gates, runtime templates, contribution titles, and the aggregate
  `Validation Gate`.
- Passed: the three preceding implementation-branch commits are GitHub-verified
  signed commits after the title-only rewrite required by the contribution
  gate.
- Passed: final PR run `29755588302`, protected merge commit
  `8995a029ca5100b9594c16cd5f877a8791616d78`, post-merge `main` run
  `29756161515`, and validation-only release preflight `29756307978`.
- Verified: the preflight publish job was skipped and no `soku/v0.1.2` tag or
  GitHub Release exists.

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex

---

## 목표 및 배경

[Issue #44](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/44)는
CLI가 불변 `--integration-ref`로 실제 fetch한 commit만 Provider API v1의
authoritative provider revision으로 신뢰하도록 요청합니다. Provider가
제공하는 `ref`는 deprecated legacy metadata로만 optional 수용하며, fetch한
revision이나 integration의 connected 전환을 결정하지 않아야 합니다.

## 제안하는 접근

Provider API major version, manifest schema, CLI flag, `IntegrationFetcher`
interface를 유지합니다. Provider API v1 schema와 decoder에서 `ref`를
optional로 바꾸되, 존재하면 lowercase 40-character SHA 형식은 계속
검증합니다. 연결 결정에서 해당 metadata를 제외하고 CLI가 요청한
정확한 commit을 request artifact와 manifest에 계속 저장합니다.

Sanitized configuration은 lifecycle 밖의 provider-owned 제출 채널로만
전달합니다. Lifecycle state는 hash-only를 유지하고 sanitized/raw
configuration, credential, secret을 저장하지 않습니다. Hermetic regression
커버리지와 공개 immutable commit에 고정한 opt-in real-Git conformance
test를 추가합니다.

`soku/v0.1.2` compatibility 및 release preflight 기록은 준비하되
발행하지 않습니다. 기존 tag와 Release는 불변으로 유지합니다.

## 계획된 구현

1. Provider API v1 required field에서 `ref`를 제거하고 deprecated annotation과
   형식 제약은 유지하며 기본 provider 예제에서 `ref`를 제거합니다.
2. Go decoder는 생략되거나 형식이 올바른 legacy `ref`를 수용하고,
   malformed value와 unknown field를 거부하며 integration plan에서 bundle
   metadata와 fetched revision을 비교하지 않습니다.
3. No-ref/legacy-ref bundle, fetched revision 저장, pending/connected state,
   mismatch, ownership, unsupported profile, no-write failure를 unit/lifecycle
   test로 검증합니다.
4. 공개 AI collaboration provider를 commit
   `a81f7c91b0c9c8faa5ba2988fde29e9d17972a83`에서 fetch하는 별도 network
   conformance test를 추가하고 read-only credential과 timeout으로
   Linux/macOS/Windows CI에서 실행합니다.
5. 5-field pending artifact를 늘리지 않고 lifecycle trust boundary와 sanitized
   configuration 제출 절차를 명확히 문서화합니다.
6. `docs/releases/soku-v0.1.2.md`에 compatibility, migration,
   validation-only release readiness 증거를 기록하며 공개 stable 버전은
   `soku/v0.1.1`로 유지합니다.
7. 전체 local/hosted validation gate를 통과하고 protected merge-commit
   workflow로 병합한 뒤 `main`과 Issue #44에 증거를 기록하되,
   Issue를 닫거나 `soku/v0.1.2`를 발행하지 않습니다.

## 수용 기준

- `ref`가 없거나 형식이 올바른 matching/mismatching legacy `ref`가 있는
  bundle은 decoding에 성공하고 malformed `ref`와 unknown field는 validation
  exit code `5`로 실패합니다.
- Legacy metadata 일치 여부는 pending/connected 결정을 바꾸지 않으며,
  request artifact와 manifest는 fetch에 사용한 lowercase full SHA를 정확히
  저장합니다.
- Configuration hash/source mismatch는 pending을 유지하고 unsupported
  profile은 exit `5`, ownership conflict는 exit `4`로 state/output을 쓰지 않고
  실패합니다.
- Pending artifact에는 schema version, provider ID, portable source,
  authoritative ref, configuration hash만 있으며 raw/sanitized configuration이나
  secret은 저장되지 않습니다.
- Hermetic lifecycle suite는 독립적으로 유지하고, 필수 3-OS CI
  conformance는 read-only token과 timeout으로 pinned external provider를 HTTPS
  fetch합니다.
- Provider-v1 compatibility, no-ref/legacy migration, manifest-v1 무변경,
  boilerplate `v1.0.0` compatibility를 `soku/v0.1.2` patch 기록에 명시합니다.
- Strict aggregate `Validation Gate`, 병합 후 `main` gate, manual
  validation-only release preflight가 통과합니다.
- `soku/v0.1.2` tag/GitHub Release를 생성하지 않고 Issue #44와
  `status:in-progress`를 별도 발행 승인 전까지 유지합니다.

## 승인

- **상태:** `Approved`
- **승인자:** `Soku-JINSEOK` (2026-07-21 Draft PR #50에 명시적 구현 지시
  기록)
- **승인 경계:** Issue #44 또는 Draft PR에 본 보고서의 명시적 승인이
  기록되기 전에는 구현을 시작하지 않습니다. Tag/Release 발행과 Issue #44
  종료는 후속 별도 승인이 필요합니다.

## 구현 현황

Provider schema, decoder, revision planning, 예제, hermetic/pinned HTTPS
conformance, lifecycle 문서, CI, `soku/v0.1.2` 제안 기록을 PR #50으로 구현해
병합했습니다. PR Validation run `29755588302`, post-merge `main` Validation run
`29756161515`, validation-only release preflight `29756307978`이 통과했습니다.
별도 발행 승인을 기다리므로 `soku/v0.1.2` tag와 GitHub Release는 없고
Issue #44는 열린 상태를 유지합니다.

## 검증

- 통과: task report Markdown lint.
- 통과: required-file 등록 YAML lint.
- 통과: 초기 report-only 변경의 `git diff --check`.
- 통과: 전체 Go unit/race/vet/gofmt/goimports/golangci-lint.
- 통과: hermetic lifecycle conformance 및 commit
  `a81f7c91b0c9c8faa5ba2988fde29e9d17972a83`의 opt-in HTTPS provider test.
- 통과: Markdown, YAML, actionlint, contribution-title, release-note
  contract, release-tag regression, Bash syntax, whitespace 검사.
- 로컬 환경 제한: PowerShell과 ShellCheck가 없고 Homebrew Go 1.26.5에
  5-target package snapshot용 cross-compile standard-library package가 없습니다.
  Hosted validation에서 세 검사를 모두 수행합니다.
- 통과: PR run `29755390115`의 PowerShell/shell parity, ShellCheck,
  5-target package snapshot, 3-OS native Go 및 hermetic/network lifecycle,
  runtime template, contribution title, aggregate `Validation Gate`.
- 통과: contribution gate가 요구한 title-only rewrite 후 앞선 구현 브랜치
  commit 3개 모두 GitHub-verified signed 상태.
- 통과: 최종 PR run `29755588302`, 보호된 merge commit
  `8995a029ca5100b9594c16cd5f877a8791616d78`, post-merge `main` run
  `29756161515`, validation-only release preflight `29756307978`.
- 확인: preflight publish job은 skip됐고 `soku/v0.1.2` tag와 GitHub Release는
  존재하지 않습니다.

## AI 지원

- **계획/구현/초안 작성:** OpenAI Codex
