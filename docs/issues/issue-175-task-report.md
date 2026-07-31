# Issue #175 Task Report — Restore Go caching in CI Quick

## Goal and Background

Issue [#175](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/175)
addresses a cache configuration failure in the shared CI Quick Go setup.
Because the repository has no root `go.mod`, `actions/setup-go` cannot discover
a dependency file and the Soku shard repeatedly downloads and compiles modules.
This contributes to the #116 critical-duration median missing its target.

## Proposed Approach

Give the existing `Setup Go` step explicit dependency paths for both Go modules
that CI Quick can validate. Keep the current Quick command set, matrix planning,
toolchain versions, Full Validation, and required contexts unchanged.

## Planned Implementation

- Add `soku/go.sum` and `templates/go/go.mod` as cache dependency inputs.
- Add a workflow regression assertion for both paths.
- Validate the focused workflow test and repository checks.
- After merge, record the exact merge SHA and activation time in the #116 audit
  and reset its active natural-sample window to zero.

## Acceptance Criteria

- CI Quick restores or saves a Go module cache without the missing root module
  warning.
- Soku and Go template Quick shards retain their existing verification commands.
- The workflow regression test and hosted Quick/Full aggregates pass.
- The previous #116 13-sample result remains historical and a new observation
  window starts at this change's merge commit.

## Approval

- **Status:** `Approved`
- **Approved by:** `Soku-JINSEOK` (requested the next planned task on
  2026-07-31)

## Implementation Status

The workflow and regression-test changes are implemented and hosted validation
has passed. The exact #116 activation record remains a post-merge documentation
step because its merge SHA and time do not exist before this change lands.

## Verification

- `node --test .github/validation-workflow.test.mjs` passed all 10 tests.
- Ruby Psych parsed `.github/workflows/ci-quick.yml` successfully.
- `git diff --check` passed.
- `scripts/verify.sh --profile fast` passed Markdown, YAML, GitHub Actions,
  Soku, Node.js, Python, Go, and Java checks, then stopped at the unrelated
  gcloud template container build because the local Docker daemon was
  unavailable.
- Hosted Validation run `30599598716` passed CI Quick, Full Validation,
  security, metadata, CodeQL, and the external control-plane check.
- The Soku Quick job restored the explicit `actions/setup-go` cache key
  successfully and emitted no missing dependency-file warning. The Go template
  job used the same primary cache key without the warning.

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

Issue [#175](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/175)는
공용 CI Quick Go 설정의 cache 구성 실패를 해결합니다. Repository root에
`go.mod`가 없어 `actions/setup-go`가 dependency file을 찾지 못하고, Soku
shard가 module을 반복해서 내려받고 compile합니다. 이 지연은 #116의 critical
duration 중앙값이 목표를 충족하지 못한 원인 중 하나입니다.

## 제안하는 접근

기존 `Setup Go` step에 CI Quick이 검증하는 두 Go module의 dependency path를
명시합니다. 현재 Quick command set, matrix planning, toolchain version, Full
Validation과 required context는 변경하지 않습니다.

## 계획된 구현

- Cache dependency input에 `soku/go.sum`과 `templates/go/go.mod`를 추가합니다.
- 두 path를 고정하는 workflow regression assertion을 추가합니다.
- 집중 workflow test와 repository check를 검증합니다.
- 병합 후 정확한 merge SHA와 activation 시각을 #116 audit에 기록하고 활성
  natural sample 관측창을 0건으로 초기화합니다.

## 수용 기준

- CI Quick이 root module 누락 경고 없이 Go module cache를 복원하거나 저장합니다.
- Soku와 Go template Quick shard가 기존 검증 command를 유지합니다.
- Workflow regression test와 hosted Quick/Full aggregate를 통과합니다.
- 기존 #116 13-sample 결과는 history로 보존하고 이 변경의 merge commit부터 새
  관측창을 시작합니다.

## 승인

- **상태:** `Approved`
- **승인자:** `Soku-JINSEOK` (2026-07-31 다음 계획 작업 시작 요청)

## 구현 현황

Workflow와 regression test 변경을 구현했고 hosted validation을 통과했습니다.
정확한 #116 activation record는 이 변경의 merge SHA와 시각이 생기는 병합 후
문서 단계로 남겨 둡니다.

## 검증

- `node --test .github/validation-workflow.test.mjs`의 10개 test를 모두
  통과했습니다.
- Ruby Psych로 `.github/workflows/ci-quick.yml` parse를 통과했습니다.
- `git diff --check`를 통과했습니다.
- `scripts/verify.sh --profile fast`는 Markdown, YAML, GitHub Actions, Soku,
  Node.js, Python, Go, Java 검증을 통과한 뒤 local Docker daemon을 사용할 수
  없어 관련 없는 gcloud template container build에서 중단됐습니다.
- Hosted Validation run `30599598716`에서 CI Quick, Full Validation, security,
  metadata, CodeQL과 외부 control-plane check를 통과했습니다.
- Soku Quick job은 명시한 `actions/setup-go` cache key를 성공적으로 복원했고
  dependency file 누락 경고가 없었습니다. Go template job도 경고 없이 같은
  primary cache key를 사용했습니다.

## 공개 적합성 검토

- [x] credential, token, private key, credential이 포함된 URL이 없음
- [x] 비공개 저장소·프로젝트·제품 이름이 없음
- [x] cloud project ID, 계정 번호, service URL, image URI, revision 식별자가 없음
- [x] 개인 청구·구독·budget·결제 상태 정보가 없음
- [x] 개인 이메일·전화번호·주소·로컬 절대 경로가 없음
- [x] 비공개 Issue·PR·Project·control-plane 식별자가 없음

## AI 지원

- **계획/구현/초안 작성:** OpenAI Codex
