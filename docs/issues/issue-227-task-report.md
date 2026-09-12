# 🔧 Issue 227 Task Report

## Goal and Background

Issue [#227](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/227)
requires one source candidate combining the exact runner Dependabot coverage
from PR #226 and fast-uri 3.1.7 from PR #224. Either PR alone is incomplete.

## Proposed Approach and Scope

The current user instruction authorizes completion of the original local
predeployment plan. Starting from main
`0f9ac36abe4255c02cc27487cc68dc34bd8aaeba`, the candidate preserves exactly five
functional changes: `.github/dependabot.yml`,
`scripts/verify-supply-chain.mjs`, `scripts/verify-supply-chain.test.mjs`,
`scripts/pull-request-policy.test.mjs`, and
`soku/internal/manual/assets/runner/package-lock.json`.
This report is the sixth path; the coverage-only Issue #225 report is excluded.

The runner path is exact, not a parent directory or wildcard. Manifest/lockfile
trust limits, major-update exclusion, existing workflow/settings, and existing
PR #224/#226 remain unchanged. No signed commit, remote push, PR publication,
manual hosted execution, Cloud, billing, Ready, merge, release or deployment
occurred during the earlier local preparation.

## Approval and Implementation Status

Local implementation and synthetic validation are authorized by the user's
current continuation request. The subsequent owner request to audit GitHub incorporation authorizes
a Draft source integration. Merge, release and Cloud execution remain pending.
Historical approvals and remote issue status are not rewritten by this report.
The combined local candidate is prepared; full acceptance remains blocked by
the required pinned browser and hosted checks.

## Actual Verification

- Focused pull-request policy, governance adapter and supply-chain tests passed.
- Supply-chain verification passed with the exact runner update target.
- Runner typecheck, build and ordinary Node tests passed; npm audit was clean.
- Go module tests, race checks, vet and govulncheck passed.
- An alternative Chromium 148 diagnostic attempted both browser tests; both
  stopped at the configured Japanese/emoji font readiness guard. The guard was
  preserved. This diagnostic is not the required pinned Chromium result.
- Pinned browser download was unavailable in this environment. Terraform's
  provider could not create its required Unix socket; Terraform execution is
  environment-blocked and not reported as passed.
- Exact file digests, patch replay tree, command logs and independent review
  are supplied with the final handoff. No historic PR success is inherited.

## Acceptance and Remaining Work

The five functional files must remain byte-equivalent to their reviewed source
PRs; the final handoff checks that invariant. Obtain the pinned browser and
configured fonts, rerun the two browser checks, then observe exact-candidate
hosted checks within the separate delivery boundary. Hosted OSV, PR policy and
main acceptance are not proven by a local dependency audit.

## 한국어 요약

#224의 fast-uri 보안 수정과 #226의 정확한 runner 검사 범위를 하나의 후보로
통합했습니다. #225 보고서는 포함하지 않고 이 #227 보고서를 작성했습니다.
로컬 검사는 기록한 범위에서 통과했지만 고정 브라우저·폰트·hosted 검증은
완료되지 않았습니다. 기존 PR과 원격 설정은 보존했습니다.

## 日本語の要約

#224 の fast-uri 修正と #226 の正確な runner 対象登録を統合しました。
#225 の報告書は含めず、この #227 報告書を追加しています。記録した範囲の
ローカル検証は合格しましたが、固定ブラウザー・フォント・hosted 検証は
未完了です。既存 PR とリモート設定は保持しています。

## AI Assistance

- Provider: OpenAI
- Model: Codex (exact backend model identifier not exposed)
- Usage Summary: Source integration, local validation, defect investigation and task-report drafting.

## GitHub incorporation review

The original PR #226 metadata gate fails because its commit subject lacks the
required Gitmoji/type pair. Editing only its PR title cannot repair that check.
The combined Issue #227 candidate uses a new correctly titled commit from the
observed main, preserving the original branches and their history. Its five
functional files still match the reviewed #224/#226 source. New-candidate
checks and any required commit signature must be verified before merge.

PR #222 is not another missing source patch: all three changed files match
current main exactly. Its old failed check does not justify applying it again.
PR #158/#159/#160/#179 are separate, unintegrated pipeline work and are not
claimed complete by this runner-only candidate.
