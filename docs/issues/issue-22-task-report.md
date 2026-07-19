# 📝 Task Report: Profiles and Bounded Declarative Extensions

## Goal and Background

[Issue #22](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/22)
adds deterministic maturity profiles and executable-free integrations without
weakening the ownership, compatibility, or transaction boundaries established
by Issues #18–#21.

## Proposed Approach

Publish a catalog-v2 profile index that composes `bootstrap → standard → scaled`
while treating a source without the index as legacy core-v1 `standard`. Keep AI
collaboration as a declarative example provider that can combine with any
profile.

Accept only `github:<owner>/<repo>/<bundle-path>` sources, lowercase full commit
refs, and explicit YAML configuration. Provider API v1 contains declarative
metadata, compatibility, schema hashes, bounded templates, and outputs; it
forbids scripts, hooks, executables, dynamic libraries, traversal, reserved
state, ownership collisions, raw configuration, and secrets. Missing exact
provider data records only a portable request artifact and `pending`; exact
matching data produces declared output and `connected`. Core and provider
changes share the existing outer transaction.

## Planned Implementation

- Add catalog-v2 profile index schema, published index, composition, and legacy
  fallback.
- Enable profile selection during init, diff, and upgrade.
- Add integration CLI flags, strict request/config validation, and stable hashes.
- Add executable-free provider API v1 bundle decoding and bounded rendering.
- Add pending request and exact-match connected lifecycle states.
- Enforce global core/provider/project ownership in one transition planner.
- Cover profile composition, legacy migration, malicious input, secrets,
  pending-to-connected, rollback, and ownership conflict.
- Publish an AI-collaboration example provider and update lifecycle docs.

## Acceptance Criteria

| Criterion | Observable evidence |
| --- | --- |
| Profile model | Default and legacy IDs resolve to `standard`; layers compose in fixed order. |
| Compatibility | Missing index is legacy standard; unsupported profile/provider combinations exit `5`. |
| Safe provider API | Only declarative bounded data is accepted; executable and escaping fields fail before writes. |
| Exact lifecycle | Missing exact data records `pending`; exact source/ref/configuration data records `connected`. |
| Ownership | Core, each provider, and project paths are globally disjoint. |
| Transaction | Profile/provider plans use diff and apply through the same manifest-last rollback boundary. |

## Approval

- **Status:** `Approved`
- **Approved by:** `Soku-JINSEOK`
- **Approval record:** The repository owner's 2026-07-19 instruction to
  implement the approved Issue #20–#23 roadmap plan.

## Implementation Status

Implemented. Catalog-v2 profile composition, legacy fallback, profile
transitions, public integration inputs, provider API v1 fetching and decoding,
pending/connected state, global ownership planning, AI-collaboration example,
schemas, rollback coverage, CI inventory, and user documentation are complete.
Pull request CI remains the merge gate.

## Verification

- Passed: `go test ./...` and `go test -race ./...` from `soku/`.
- Passed: `go vet ./...`, gofmt, goimports v0.48.0, and golangci-lint v2.12.2.
- Passed: profile/provider contract and published JSON Schema validation.
- Passed: Markdown lint, GitHub YAML lint, title tests, and `git diff --check`.
- Passed: pending-to-connected, provider rollback, ownership, secret,
  executable, traversal, legacy migration, and profile composition tests.

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex

---

## 목표 및 배경

[Issue #22](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/22)는
기존 ownership, compatibility, transaction 경계를 약화하지 않는 결정적 maturity
profile과 executable-free integration을 추가합니다.

## 제안하는 접근

catalog-v2 index에서 `bootstrap → standard → scaled`를 선형 합성하고 index가 없는
source는 legacy core-v1 `standard`로 해석합니다. AI collaboration은 profile이 아닌
세 profile 모두와 조합 가능한 declarative example provider로 제공합니다.

provider는 exact source/ref/configuration hash가 일치할 때만 declared output을 만들고,
그 외에는 portable request artifact와 `pending`만 기록합니다. script, hook,
executable, traversal, reserved path, secret과 ownership 충돌은 write 전에 거부합니다.

## 계획된 구현

- catalog-v2 index/schema, 합성, legacy fallback과 profile transition
- integration CLI 입력, strict validation과 hash
- provider API v1 bundle decode/render와 pending/connected state
- core/provider/project 전역 ownership과 동일 outer transaction
- 조합, migration, malicious input, secret, pending 연결, rollback 테스트
- AI collaboration example과 lifecycle 문서

## 수용 기준

| 기준 | 관찰 가능한 근거 |
| --- | --- |
| Profile | 기본/legacy `standard`와 고정된 선형 layer 순서 |
| 호환성 | index 없는 standard fallback과 unsupported 조합 거부 |
| Provider | 선언적 bounded data만 허용하고 실행/escape 입력 거부 |
| Exact state | 불일치 `pending`, exact match `connected` |
| Ownership | core/provider/project 경로 전역 분리 |
| Transaction | profile/provider를 manifest-last rollback 경계에서 처리 |

## 승인

- **상태:** `Approved`
- **승인자:** `Soku-JINSEOK`
- **승인 기록:** 2026-07-19 승인된 Issue #20–#23 roadmap 전체 구현 지시

## 구현 현황

구현을 완료했습니다. catalog-v2 profile 합성, legacy fallback, profile transition,
integration 공개 입력, provider API v1 fetch/decode, pending/connected, 전역 ownership,
AI example, schema, rollback 테스트, CI 목록과 문서를 반영했습니다. PR CI를 최종
merge gate로 유지합니다.

## 검증

- 통과: `go test ./...`, `go test -race ./...`, `go vet ./...`
- 통과: gofmt, goimports v0.48.0, golangci-lint v2.12.2
- 통과: profile/provider contract와 published JSON Schema 검증
- 통과: Markdown, GitHub YAML, title test, `git diff --check`
- 통과: pending 연결, provider rollback, ownership, secret, executable, traversal,
  legacy migration과 profile 합성 테스트

## AI 지원

- **계획/구현/초안 작성:** OpenAI Codex
