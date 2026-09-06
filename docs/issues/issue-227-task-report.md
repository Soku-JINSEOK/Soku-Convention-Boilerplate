# Issue #227 Task Report — Manual runner security integration

## Goal and Background

[Issue #227](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/227)
combines manual-runner dependency update coverage with the fast-uri3.1.7 fix.
The coverage-only Issue225 scope excludes the lock repair; this dedicated Issue
owns the combination and preserves both existing PR224 and PR226.

## Proposed Approach

Start from main0f9ac36abe4255c02cc27487cc68dc34bd8aaeba. Retain exactly the four
functional changes from PR226 at189b71e7ee40ae366c86f2e2ae3be54add3f5119 and the
three lock values from PR224 at177dba61d7c2f58e2a1ec7cc55730ab992e565bd.
Do not carry the Issue225 report or old commit history into this candidate.

## Planned Implementation

- Register the manual runner directory in .github/dependabot.yml.
- Require that update target in scripts/verify-supply-chain.mjs and its test.
- Retain the scoped Dependabot path regressions in scripts/pull-request-policy.test.mjs.
- Update only version, resolved URL and integrity for fast-uri in
  soku/internal/manual/assets/runner/package-lock.json.
- Add this accurate Issue227 report. No other tracked path changes.

## Acceptance Criteria

- [x] Four functional blobs equal the reviewed source and the lock delta is exact.
- [x] Focused supply-chain, policy and adapter tests pass.
- [x] Manual runner dependency audit, compilation, tests and browser checks pass.
- [x] Applicable repository validation, diff, disclosure and final review pass.
- [ ] Final committed-candidate and hosted acceptance remain separately evidenced.

## Approval

- **Status:** Approved
- **Approved by:** Soku-JINSEOK
- **Scope:** The owner approved the dedicated Issue and six-file local integration
  plan, then separately approved its initial automatic metadata effects before
  the single Issue creation. Local implementation proceeds after independent
  scope and effect review. This phase does not execute commit signing, push,
  PR publication, Cloud, workflow dispatch, merge, release or deployment.

## Implementation Status

The five frozen functional changes and this report are prepared as an uncommitted
candidate from the exact main above. No workflow, permissions, matcher, secret,
sync behavior, required check, existing PR or Issue225 content changes.

## Verification

Initial Issue metadata converged to Open, assigned owner and canonical
type:chore/area:tooling/status:triage. Its single Project item converged to
Inbox/P1/S/Security without a Target date.

Automatic runs34010866159 and34010885722 succeeded. Initial run34010865506 failed
from GitHub's secondary read-rate limit before synchronization writes; the same
event group's pending run34010865676 was automatically cancelled. No manual
retry, cancellation or metadata correction was performed. These are mixed
automatic run outcomes with verified final convergence, not an all-runs-PASS claim.

Local candidate results:

- Four functional files cmp-match the reviewed PR226 bytes; all modes preserved.
- Lockfile semantics change only the three reviewed fast-uri values, with one
  fast-uri entry; SHA-256
  b5e0ae48a2bc671980c8007e4d325a9a468605cfe85cf2d4a0e30e1e02869f89.
- Policy/supply-chain tests:24PASS; Dependabot/governance adapter tests:4PASS.
- Supply-chain command:38protected files/12update targets PASS.
- Runner npm ci --ignore-scripts and npm audit --audit-level=high:PASS,
  zero reported vulnerabilities.
- Runner npm test:TypeScript compilation and8tests PASS;2opt-in browser cases
  skipped in that command, then explicitly executed with SOKU_BROWSER_E2E=1:
  both PASS. Total focused successful cases across these suites:38.
- Pinned Markdown and YAML lint and git diff --check:PASS.
- Six-file changed-line secret scan and report placeholder/local-path scan:PASS.

Independent POST review passed this local uncommitted candidate, including
the report-only template correction. These are uncommitted-worktree results,
not exact committed-candidate, hosted OSV, full CI, live GAS or real user evidence.
Synthetic browser scenarios do not supply the separate product acceptance gates.

## Public Disclosure Review

- [x] No credentials, tokens, private keys, or credential-bearing URLs.
- [x] No private repository, project, or product names.
- [x] No cloud project IDs, account numbers, service URLs, image URIs, or
  revision identifiers.
- [x] No personal billing, subscription, budget, or payment-status information.
- [x] No personal email, phone, address, or local absolute path.
- [x] No private Issue, PR, Project, or control-plane identifiers.

Only this public repository's task and validation references are included.

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex.
- **Independent review:** OpenAI Codex, read-only reviewer.

---

## 목표 및 배경

[Issue #227](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/227)은
manual runner의 의존성 업데이트 검사 범위와 fast-uri3.1.7 수정을 통합합니다.
Issue225는 잠금파일 수정을 제외하므로 별도 Issue가 결합 범위를 소유합니다.
기존 PR224와 PR226은 변경하지 않습니다. 위 영문이 규범 원문입니다.

## 제안하는 접근

영문에 기록한 정확한 main을 기준으로 PR226의 기능 파일 4개와 PR224의
fast-uri version, resolved URL, integrity 값만 결합합니다. Issue225 보고서와
이전 커밋 이력은 복사하지 않습니다.

## 계획된 구현

- .github/dependabot.yml에 runner 디렉터리를 등록합니다.
- scripts/verify-supply-chain.mjs와 해당 테스트에 필수 검사 대상을 추가합니다.
- scripts/pull-request-policy.test.mjs의 승인된 Dependabot 경로 테스트를 유지합니다.
- soku/internal/manual/assets/runner/package-lock.json의 세 값만 변경합니다.
- 이 Issue227 보고서를 추가하며 다른 추적 파일은 변경하지 않습니다.

## 수용 기준

- [x] 기능 파일 4개가 원본과 같고 잠금파일 변경이 정확합니다.
- [x] 공급망·정책·adapter 테스트가 통과했습니다.
- [x] runner audit·컴파일·단위·브라우저 검사가 통과했습니다.
- [x] 적용되는 로컬 검사와 공개 적합성 및 최종 독립 검토가 완료되었습니다.
- [ ] 최종 커밋과 hosted 검증은 별도 증거가 필요합니다.

## 승인

- **상태:** Approved
- **승인자:** Soku-JINSEOK
- **범위:** 전용 Issue와 6개 파일 로컬 통합 계획을 승인했으며, 단일 Issue 생성
  전에 초기 자동 메타데이터 효과를 별도로 승인했습니다. 독립 범위·효과 검토
  후 로컬 구현을 진행합니다. 이 단계는 서명·커밋·push·PR 공개·Cloud·수동 실행·
  병합·release·배포를 수행하지 않습니다.

## 구현 현황

정확한 main 위에 기능 변경 5개와 보고서로 구성된 미커밋 후보를 준비했습니다.
workflow, 권한, matcher, secret, 동기화 동작, 필수 검사, 기존 PR 및 Issue225는
변경하지 않았습니다.

## 검증

Issue는 Open, owner 담당, type:chore/area:tooling/status:triage로 확인됐고,
Project 항목 1개가 Inbox/P1/S/Security로 수렴했습니다. Target date는 없습니다.
영문에 명시한 자동 실행 2개가 성공했습니다. 초기 실행 1개는 쓰기 전 GitHub
조회 제한으로 실패했고 대기 실행 1개는 같은 동시성 그룹에서 자동 취소됐습니다.
수동 재시도·취소·보정은 하지 않았으며 모든 실행이 성공했다고 주장하지 않습니다.

- 기능 파일 원본 비교, 세 값만 바뀐 잠금파일 의미·해시 비교가 통과했습니다.
- 정책·공급망 24개와 Dependabot·adapter 4개 테스트가 통과했습니다.
- 공급망 검사는 보호 파일 38개와 업데이트 대상 12개에서 통과했습니다.
- runner npm ci --ignore-scripts와 audit가 통과했고 보고된 취약점은 0개입니다.
- 컴파일과 단위 테스트 8개가 통과했습니다. 최초 건너뛴 opt-in 브라우저 2개는
  별도 SOKU_BROWSER_E2E=1 실행에서 모두 통과했습니다. 총 성공 사례는 38개입니다.
- 고정 버전 Markdown·YAML lint, diff, 변경 줄 secret 검사와 보고서의
  placeholder·로컬 경로 검사가 통과했습니다.

보고서 템플릿 보완을 포함한 이 미커밋 로컬 후보의 독립 POST 검토가 통과했습니다.
미커밋 로컬 결과는 최종 커밋·hosted OSV·전체
CI·실제 GAS·사용자 환경 증거가 아니며 합성 브라우저 사례로 제품 검증을 대체하지 않습니다.

## 공개 적합성 검토

- [x] credential, token, private key, credential이 포함된 URL이 없습니다.
- [x] 비공개 저장소·프로젝트·제품 이름이 없습니다.
- [x] cloud project ID, 계정 번호, service URL, image URI, revision 식별자가 없습니다.
- [x] 개인 청구·구독·budget·결제 상태 정보가 없습니다.
- [x] 개인 이메일·전화번호·주소·로컬 절대 경로가 없습니다.
- [x] 비공개 Issue·PR·Project·control-plane 식별자가 없습니다.

## AI 지원

- **계획/구현/초안 작성:** OpenAI Codex.
- **독립 검토:** OpenAI Codex, 읽기 전용 검토자.
