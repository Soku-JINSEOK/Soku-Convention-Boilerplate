# Issue #164 Task Report — Real-runtime user-manual capture

## Goal and Background

Issue [#164](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/164)
adds an opt-in workflow that runs actual frontend source in Chromium and
produces reproducible user-manual PNGs, stable capture relations, and
credential-redacted provenance instead of drawing an approximation.

The component is local/manual only. It must support ordinary web applications,
Google Apps Script HTML Service, deterministic maps, bounded Leaflet/OSM, and
restricted Google Maps JavaScript without authorizing production access,
customer data, cloud mutation, recurring hosted capture, or automatic release.

## Proposed Approach

Add nested read-only planner and doctor commands plus an explicit transactional
component installer. Preserve manifest v1 for existing repositories and migrate
to v2 only when `docs-manual` is installed. Ship an isolated Node.js 22+
TypeScript/Playwright runner with versioned configuration, report, component,
and provider-egress contracts.

Keep runner, schemas, example, and template core-managed. Keep actual
configuration, fixtures, hooks, scenarios, images, reports, manual prose, and
PDFs project-owned. Deny external network access by default, redact credential
values and URLs, preserve map attribution, and require explicit provider
readiness and request budgets.

## Planned Implementation

- Add the English normative standard and Korean/Japanese summaries.
- Extend lifecycle commands and manifest v2 component metadata.
- Implement strict YAML validation, deterministic planning, static doctor, and
  fixed-argv opt-in probe.
- Implement transactional component installation with exact v1 rollback.
- Add the isolated runner, web/GAS/backend/dialog/map adapters, provenance,
  output ownership checks, and atomic replacement.
- Add schemas, catalogs, examples, fixtures, unit tests, and synthetic browser
  validation.
- Run local downstream pilots; keep restricted live Google Maps evidence private
  and sanitized.

## Acceptance Criteria

- Existing v1 lifecycle reads remain supported and read-only commands do not
  migrate state.
- `docs manual init` reports and transactionally applies v1-to-v2 migration,
  refuses collisions/drift, and restores the exact v1 manifest on failure.
- Configuration rejects unknown/duplicate fields, secret literals, unsafe
  paths/HAR, unsupported actions/providers, hosted execution, missing provider
  budgets/readiness/attribution, and unsafe clips.
- Runner output records hashes, source/dirty state, browser settings, adapters,
  redacted egress, map readiness/attribution/budgets, and stable capture
  relations.
- Generated output replacement never overwrites an unowned or manually changed
  path and replaces the report last.
- GAS bridge preserves asynchronous chainable semantics and disclosed dialog
  overlays.
- Hosted checks use synthetic fixtures and deterministic providers only.
- No cloud project, billing, API, quota, budget, key, restriction, or production
  state is created or changed.

## Approval

- **Status:** `Approved`
- **Approved by:** `Soku-JINSEOK` (requested implementation of the approved plan
  on 2026-07-30)

## Implementation Status

The source implementation is complete in the local worktree. The general web
and synthetic GAS/browser pilots are complete. Live Leaflet/OSM and restricted
Google Maps JavaScript downstream pilots still require reviewed downstream
applications; the Google pilot additionally requires the separately owned,
restricted key and billing-owner declaration. No such key was present during
this pass.

Issue closure evidence remains pending those two live-provider pilots. Public
CLI, npm, or boilerplate release remains a separate approval-gated activity.

## Verification

- `go test ./...` and `go test -race ./...` passed.
- Runner `typecheck`, unit tests, build, report-schema validation, redaction,
  integrity, and transactional output tests passed.
- Pinned Chromium browser tests passed for 414 px and desktop web capture,
  HTTP-route fixtures, edge-clamped locator capture, repeated generated-output
  replacement, GAS asynchronous success/failure/user-object/mutation,
  disclosed dialog overlay, attribution refusal, missing-font refusal, and a
  credential-redacted Google provider test double.
- The fixed browser probe passed with Chromium 151 without emitting credential
  values.
- Five native package archives passed checksum and reproducibility checks.
- Markdown, YAML, JSON, shell, lifecycle, and repository security checks
  passed; the security pass found no leaked secrets or known dependency
  vulnerabilities.
- `scripts/verify.sh --profile full --skip-db` passed repository, lifecycle,
  race, package, and Node/Python/Go/Java template groups, then stopped at the
  unrelated gcloud template container build because the local Docker daemon was
  unavailable. DB, hosted cross-platform, and hosted provider-network gates are
  therefore not recorded as passes.
- Live Leaflet/OSM and live Google Maps JavaScript captures were not executed
  and no live-provider evidence is claimed.

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

Issue [#164](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/164)는
화면을 다시 그리는 대신 실제 frontend source를 Chromium에서 실행하고 재현
가능한 사용자 매뉴얼 PNG, stable capture 관계와 credential이 제거된 provenance를
만드는 opt-in workflow를 추가합니다.

Component는 local/manual 전용입니다. 일반 web app, Google Apps Script HTML
Service, deterministic map, 제한된 Leaflet/OSM, 제한 key를 사용하는 Google Maps
JavaScript를 지원하되 production 접근, 고객 data, cloud 변경, 반복 hosted
capture와 자동 release를 승인하지 않습니다.

## 제안하는 접근

읽기 전용 planner·doctor와 명시적인 transactional component installer를
추가합니다. 기존 repository의 manifest v1을 보존하고 `docs-manual` 설치 때만
v2로 전환합니다. 별도 lockfile을 가진 Node.js 22+ TypeScript/Playwright runner와
versioned config, report, component, provider egress contract를 제공합니다.

Runner·schema·example·template는 core-managed, 실제 config·fixture·hook·scenario,
image·report·manual prose·PDF는 project-owned로 둡니다. 외부 network는 기본
거부하고 credential 값과 URL을 제거하며 map attribution, readiness, request
budget를 강제합니다.

## 계획된 구현

- English normative standard와 Korean/Japanese summary를 추가합니다.
- Nested lifecycle command와 manifest v2 component metadata를 추가합니다.
- Strict YAML validation, deterministic plan, static doctor와 fixed-argv opt-in
  probe를 구현합니다.
- Exact v1 rollback을 포함한 transactional component init을 구현합니다.
- 격리 runner, web/GAS/backend/dialog/map adapter, provenance, output ownership와
  atomic replacement를 구현합니다.
- Schema, catalog, example, fixture, unit test와 synthetic browser validation을
  추가합니다.
- Local downstream pilot을 실행하고 제한된 Google Maps evidence는 private하고
  sanitized하게 유지합니다.

## 수용 기준

- 기존 v1 lifecycle read를 지원하고 read-only command는 state를 전환하지 않습니다.
- `docs manual init`은 v1→v2 plan/apply, collision/drift 거부, 실패 시 exact v1
  복원을 보장합니다.
- Config는 unknown/duplicate field, secret literal, unsafe path/HAR, unsupported
  action/provider, hosted mode, 누락된 map budget/readiness/attribution와 unsafe
  clip을 거부합니다.
- Runner output은 hash, source/dirty state, browser setting, adapter, redacted
  egress, map readiness/attribution/budget와 stable capture relation을 기록합니다.
- Generated output은 unowned 또는 수동 변경 path를 덮어쓰지 않고 report를
  마지막에 교체합니다.
- GAS bridge는 asynchronous chainable semantic과 공개된 dialog overlay를
  보존합니다.
- Hosted check는 synthetic fixture와 deterministic provider만 사용합니다.
- Cloud project, billing, API, quota, budget, key, restriction, production state를
  만들거나 변경하지 않습니다.

## 승인

- **상태:** `Approved`
- **승인자:** `Soku-JINSEOK` (2026-07-30 승인된 계획 구현 요청)

## 구현 현황

Local worktree의 source 구현은 완료했습니다. 일반 web과 synthetic GAS/browser
pilot도 완료했습니다. 실제 Leaflet/OSM과 제한된 Google Maps JavaScript downstream
pilot은 검토된 downstream app이 필요하며, Google pilot에는 별도 소유의 제한 key와
billing owner 선언도 필요합니다. 이번 pass에는 해당 key가 없었습니다.

따라서 Issue 종료 evidence는 두 live-provider pilot 전까지 pending입니다. 공개
CLI·npm·boilerplate release는 별도 승인 gate로 유지합니다.

## 검증

- `go test ./...`, `go test -race ./...`를 통과했습니다.
- Runner typecheck, unit test, build, report schema validation, redaction,
  integrity와 transactional output test를 통과했습니다.
- 고정 Chromium browser test에서 414 px·desktop web capture, HTTP route fixture,
  edge-clamped locator capture, generated output 재교체, GAS async
  success/failure/user-object/mutation, 공개된 dialog overlay, attribution crop
  거부, font 누락 거부와 credential이 제거된 Google provider test double을
  통과했습니다.
- Fixed browser probe는 credential 값을 출력하지 않고 Chromium 151에서
  통과했습니다.
- 5개 native package archive의 checksum과 reproducibility 검증을 통과했습니다.
- Markdown, YAML, JSON, shell, lifecycle과 repository security 검증을 통과했고
  secret leak 또는 알려진 dependency vulnerability가 발견되지 않았습니다.
- `scripts/verify.sh --profile full --skip-db`는 repository, lifecycle, race,
  package, Node/Python/Go/Java template group을 통과한 뒤 local Docker daemon이
  없어 관련 없는 gcloud template container build에서 중단됐습니다. 따라서 DB,
  hosted cross-platform과 hosted provider-network gate는 통과로 기록하지
  않습니다.
- 실제 Leaflet/OSM과 Google Maps JavaScript capture는 실행하지 않았으며 live
  provider evidence를 주장하지 않습니다.

## 공개 적합성 검토

- [x] credential, token, private key, credential이 포함된 URL이 없음
- [x] 비공개 저장소·프로젝트·제품 이름이 없음
- [x] cloud project ID, 계정 번호, service URL, image URI, revision 식별자가 없음
- [x] 개인 청구·구독·budget·결제 상태 정보가 없음
- [x] 개인 이메일·전화번호·주소·로컬 절대 경로가 없음
- [x] 비공개 Issue·PR·Project·control-plane 식별자가 없음

## AI 지원

- **계획/구현/초안 작성:** OpenAI Codex
