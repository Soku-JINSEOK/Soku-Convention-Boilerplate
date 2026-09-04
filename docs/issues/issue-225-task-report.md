# 🔧 Issue 225 Task Report

## Goal and Background

Issue [#225](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/225)
registers the exact npm manifest directory
`/soku/internal/manual/assets/runner` as an explicit Dependabot update target.
The configuration closes the trusted-file-scope gap that blocked the
fast-uri security update in PR #224 from satisfying the repository policy.

## Proposed Approach

Permit only the exact runner directory, preserve the existing npm update
contract, and cover both required supply-chain coverage and the policy's
Dependabot file-scope behavior with regression tests.

## Implemented Scope

- `.github/dependabot.yml`: exact runner npm block using the existing monthly,
  one-open-PR, minor/patch and security grouping, major-ignore, assignee,
  label, and commit-message policy.
- `scripts/verify-supply-chain.mjs`: exact runner npm required update target.
- `scripts/verify-supply-chain.test.mjs` and
  `scripts/pull-request-policy.test.mjs`: regression coverage for the target,
  lockfile scope, and missing coverage finding.

## Acceptance Criteria

- The trust scope is only `/soku/internal/manual/assets/runner`; no parent
  path or wildcard is allowed.
- The existing manifest/lockfile-only Dependabot policy remains unchanged.
- Focused policy and supply-chain validation pass.
- A later, separately authorized PR can provide the exact Common Metadata task
  report path required by the PR gate.

## Approval

- **Status:** `Approved`
- **Approved by:** Soku-JINSEOK through D002/D005 scoped owner decisions.

## Implementation Status

The four-file candidate was reviewed, signed, and delivered to
`agent/issue-225-dependabot-runner-path` at commit
`a3c39a5d369c3e953196fd22808c044929841812`. No pull request, Cloud execution,
merge, or post-merge result exists yet.

## Verification

- Pull-request policy tests: 16 passed.
- Governance adapter tests: 2 passed.
- Supply-chain tests: 8 passed.
- Supply-chain verification: 38 protected files and 12 update targets.
- `git diff --check`: passed before signed delivery.
- Commit signature and remote branch SHA: verified.

## Public Disclosure Review

- [x] No credentials, tokens, private keys, or credential-bearing URLs
- [x] No cloud project IDs, account numbers, service URLs, or billing data
- [x] No personal email, phone, address, or local absolute paths
- [x] No unredacted secret or signing material

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex

---

## 목표 및 배경

Issue #225는 수동 runner의 정확한 npm 디렉터리만 Dependabot 신뢰 범위에
등록합니다. 이 변경은 PR #224의 fast-uri 보안 업데이트가 정책의 파일 범위
조건을 충족하지 못했던 문제를 해결합니다.

## 구현 범위와 비파괴 조건

Dependabot 설정, 공급망 대상, 두 회귀 테스트만 변경했습니다. 상위 경로,
와일드카드, credential, 권한, 워크플로우, Cloud, PR #224, 병합은 포함하지
않습니다.

## 검증과 후속 작업

focused policy·governance·supply-chain 검증이 통과했고 signed branch delivery가
완료되었습니다. PR 생성 및 hosted 검증은 별도 승인과 checkpoint가 필요합니다.

## AI 지원

- **계획/구현/초안 작성:** OpenAI Codex
