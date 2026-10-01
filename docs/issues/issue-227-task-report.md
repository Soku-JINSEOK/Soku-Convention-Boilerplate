# 🔧 Issue 227 Task Report

> **Document purpose:** Task decision record. Records the scope, decisions and verification for Issue #227: combining manual-runner dependency coverage and a fast-uri update.
>
> **Key point:** Read approval, implementation status and verification separately; this report is dated task evidence, not a fresh claim of current completion.

## Goal and Background

Issue [#227](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/227)
requires one source candidate combining the exact runner Dependabot coverage
from PR #226 and fast-uri 3.1.7 from PR #224. Either PR alone is incomplete.

## Proposed Approach and Scope

The current user instruction authorizes completion of the original local
predeployment plan. Starting from main
`0f9ac36abe4255c02cc27487cc68dc34bd8aaeba`, the original candidate preserves five
reviewed functional changes: `.github/dependabot.yml`,
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

PR #224의 fast-uri 보안 수정과 #226의 정확한 runner 검사 범위를 하나의 후보로
통합했습니다. #225 보고서는 포함하지 않고 이 #227 보고서를 작성했습니다.
로컬 검사는 기록한 범위에서 통과했지만 고정 브라우저·폰트·hosted 검증은
완료되지 않았습니다. 기존 PR과 원격 설정은 보존했습니다.

## 日本語の要約

PR #224 の fast-uri 修正と #226 の正確な runner 対象登録を統合しました。
Issue #225 の報告書は含めず、この #227 報告書を追加しています。記録した範囲の
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

## Exact-candidate hosted failure and correction

The new Draft #229 triggered real code validation at commit
`5d54940d80301096589333db6b1dccda47854b5d`. Metadata-only Validation runs
succeeded, but code run `34682376195` failed. Those green metadata runs are
not evidence that code validation passed.

The code run found three Markdown heading errors in this report and existing
JS template vulnerabilities: vitest/@vitest/mocker 4.1.10 and js-yaml 4.3.1.
The current owner request authorizes resolving these discovered validation
failures. The follow-up adds the template manifest/lock patch (Vitest and
coverage 4.1.11; compatible js-yaml 4.3.2) while preserving the original five
runner functional files. It neither changes workflow/settings nor enables
delivery. New code-triggered validation must succeed on the final head.

## 2026-09-27 signed successor candidate

This new candidate starts from remote `main`
`0f9ac36abe4255c02cc27487cc68dc34bd8aaeba` in a separate checkout.
It combines the #236 code and metadata gate separation, exact-head Hosted Full
contract, Vitest 4.1.11, and the #234 metadata concurrency correction with
the four runner Dependabot coverage files omitted from #236. The manual
runner manifest and lockfile now both specify fast-uri 3.1.7. The JavaScript
template manifest explicitly overrides js-yaml to 4.3.2, matching its lockfile.

The focused workflow, policy and supply-chain suites passed 43 tests.
Supply-chain validation checked 39 protected files and 12 update targets.
The runner and JavaScript template installed, tested, and passed high-severity
npm audits with zero reported vulnerabilities. The template also passed
format, lint and build checks. Pinned actionlint v1.7.12 reported no findings.
The uninterrupted `scripts/verify.sh --profile full` run passed with exit
code 0, including Docker-backed DB schemas, Terraform, Go vulnerability and
OSV scans. Neither infrastructure nor database checks were skipped. The
changed reports passed markdownlint-cli2 v0.22.1 with zero errors. Hosted
checks must be recorded separately against the signed head; earlier results
from #236 do not validate it.

GitHub's required-check evaluation on PR #237 exposed an additional workflow
issue: metadata-only Validation runs left `Validation Gate` expected and blocked
merge even after the signed code run and manual Hosted Full passed. The follow-up
workflow correction restricts Validation PR triggers to code-bearing events
and treats every `edited` event as code-bearing so base edits are revalidated.
PR metadata policy remains a separate workflow. The corrected head must pass
all hosted gates again before merge.
The correction also passed the full local profile without skip flags and the
three changed documentation files passed markdownlint-cli2 with zero errors.
