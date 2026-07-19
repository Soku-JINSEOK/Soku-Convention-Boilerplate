# 📝 Issue 39 Task Report

## Goal and Background

[Issue #39](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/39)
requests a repository-wide documentation, code, GitHub governance, release
integrity, and personal-account cost audit. The work must improve validation and
correct reviewable defects without changing the public `soku` API, rewriting
history, moving published tags or releases, or publishing a new release.

## Proposed Approach

Use two reviewed pull requests. The implementation pull request adds the
verification contract, aligns public documentation, consolidates CI behind one
required gate, enforces contribution titles, and pins third-party actions. After
that pull request passes and merges, enable the supported repository ruleset and
open a second evidence pull request containing the final GitHub, history, and
read-only billing findings.

Historical records remain immutable. Metadata defects may be corrected in
place, while incompatible or release-requiring defects become follow-up issues.

## Planned Implementation

1. Add `VERIFICATION_GUIDE.md` and link it from the multilingual READMEs.
2. Document pull-request, completion-metadata, hosted-resource, and cost-audit
   requirements in the GitHub and CI/CD standards.
3. Make one top-level Validation workflow call repository and runtime-template
   validation, enforce PR and commit titles, cancel stale runs, use read-only
   permissions by default, and pin external actions to full commit SHAs.
4. Run the documented local, hosted, security, dependency, release, history,
   repository-setting, and cost checks.
5. Merge the implementation pull request, activate the supported `main`
   ruleset, then merge a final evidence pull request through that ruleset.
6. Recheck final state and close Issue #39 with completion metadata.

## Acceptance Criteria

- All applicable local and hosted checks pass, or a precise environment limit is
  recorded with no false success claim.
- One `Validation Gate` is the required status, contribution titles are blocked
  when invalid, and every external action reference uses a full commit SHA.
- The active `main` ruleset requires pull requests, signed commits, resolved
  conversations, and the latest successful gate, while blocking deletion and
  force-push without routine bypass.
- Historical issue/PR metadata corrections and exceptions are recorded without
  rewriting bodies, commits, tags, or releases.
- Cost evidence separates repository-attributable metered cost, pre-existing
  personal-account usage/subscriptions, and future quota risk without recording
  payment identifiers.
- The final evidence pull request passes the protected gate, Issue #39 is closed
  with `status:done`, and the final local worktree is clean.

## Approval

- **Status:** `Approved`
- **Approved by:** `User (provided the plan and explicitly requested implementation)`

## Implementation Status

Implementation changes are prepared on the audit branch. The protected hosted
gate, repository settings, authenticated billing evidence, and final evidence
pull request remain pending.

## Verification

Preliminary local verification completed on 2026-07-20 (JST):

- Markdown, YAML, JSON, link, Action semantics, contribution-title, release-tag,
  Bash syntax, Go unit/race/vet/format/import/lint, lifecycle/provider, five-target
  packaging, Java, JavaScript, Python, and secret checks passed where the local
  environment supported them.
- The JavaScript dependency audit is clean after pinning the repaired `tmp`
  release, and the Java OSV finding is resolved by the current Jackson BOM.
- PowerShell, ShellCheck, database services, and cloud/container validation are
  delegated to the hosted gate because the required local runtimes are absent.
- [Issue #40](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/40)
  tracks the Pyink-pinned vulnerable Black release. [Issue #41](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/41)
  tracks defects reproduced from immutable boilerplate release `v1.0.0` that
  require a successor release.
- Public `soku/v0.1.1` assets, checksums, metadata, native execution, and isolated
  `go install` passed. Published tags and releases were not changed.

Final command versions, hosted results, historical exceptions, repository
settings, resource usage, and billing evidence will be recorded after the
implementation pull request merges.

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex

---

## 목표 및 배경

[Issue #39](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/39)는
저장소 전체 문서, 코드, GitHub 거버넌스, 릴리스 무결성, 개인 계정 비용 감사를
요청합니다. 공개 `soku` API, 과거 기록, 공개 태그·Release를 변경하거나 새
Release를 발행하지 않고 검증 체계와 발견된 결함을 개선해야 합니다.

## 제안하는 접근

두 개의 검토 가능한 PR로 진행합니다. 첫 PR에서 검증 계약, 공개 문서 정합성,
단일 필수 gate, 제목 규칙 차단, 외부 Action SHA 고정을 구현합니다. 해당 PR이
통과·병합된 뒤 지원되는 저장소 ruleset을 활성화하고, 최종 GitHub·이력·읽기
전용 비용 증거를 두 번째 PR로 기록합니다.

과거 기록은 불변으로 유지합니다. 메타데이터 결함만 제자리에서 수정하고,
호환성 변경이나 신규 릴리스가 필요한 결함은 후속 Issue로 분리합니다.

## 계획된 구현

1. `VERIFICATION_GUIDE.md`를 추가하고 다국어 README에서 연결합니다.
2. GitHub/CI 표준에 PR, 완료 메타데이터, hosted resource, 비용 감사 규칙을
   기록합니다.
3. 상위 Validation workflow에서 저장소 및 runtime template 검증을 호출하고,
   제목 규칙·concurrency·read-only 권한·외부 Action SHA 고정을 적용합니다.
4. 문서화한 로컬, hosted, 보안, 의존성, 릴리스, 이력, 설정, 비용 검사를
   실행합니다.
5. 구현 PR 병합 후 `main` ruleset을 활성화하고 최종 증거 PR을 해당 ruleset
   아래에서 병합합니다.
6. 최종 상태를 재확인하고 Issue #39를 완료 메타데이터와 함께 닫습니다.

## 수용 기준

- 적용 가능한 로컬·hosted 검사가 통과하며, 환경 제한은 거짓 성공 없이 정확히
  기록됩니다.
- 단일 `Validation Gate`가 필수 상태가 되고, 제목 규칙이 실제 차단되며, 외부
  Action은 모두 전체 commit SHA로 고정됩니다.
- 활성 `main` ruleset이 PR, signed commit, conversation resolution, 최신 gate를
  요구하고 deletion/force-push를 routine bypass 없이 차단합니다.
- 과거 본문·commit·tag·Release를 재작성하지 않고 Issue/PR 메타데이터 수정과
  예외를 기록합니다.
- 결제 식별자를 저장하지 않고 저장소 유발 비용, 개인 계정 기존 사용/구독,
  향후 quota 위험을 분리합니다.
- 최종 증거 PR이 보호 gate를 통과하고 Issue #39가 `status:done`으로 닫히며
  최종 로컬 작업 트리가 깨끗합니다.

## 승인

- **상태:** `Approved`
- **승인자:** `User (계획을 제공하고 구현을 명시적으로 요청함)`

## 구현 현황

감사 브랜치의 구현 변경은 준비되었습니다. 보호된 hosted gate, 저장소 설정,
인증된 비용 증거, 최종 증거 PR은 아직 남아 있습니다.

## 검증

2026-07-20(JST) 기준 1차 로컬 검증을 완료했습니다.

- 로컬 환경이 지원하는 Markdown, YAML, JSON, 링크, Action 의미 검증, 제목 규칙,
  릴리스 태그, Bash 문법, Go unit/race/vet/format/import/lint,
  lifecycle/provider, 5개 대상 패키징, Java, JavaScript, Python, secret 검사가
  통과했습니다.
- 수정된 `tmp` 버전을 고정해 JavaScript 의존성 감사를 정리했고, 현재 Jackson
  BOM으로 Java OSV 발견 사항을 해결했습니다.
- 로컬에 없는 PowerShell, ShellCheck, DB service, cloud/container 검증은 hosted
  gate가 수행합니다.
- [Issue #40](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/40)은
  Pyink가 고정한 취약 Black 버전을 추적하고, [Issue #41](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/41)은
  불변인 boilerplate `v1.0.0`에서 재현되어 후속 Release가 필요한 결함을
  추적합니다.
- 공개 `soku/v0.1.1` asset, checksum, metadata, native 실행, 격리된
  `go install` 검증은 통과했고 공개 tag·Release는 변경하지 않았습니다.

최종 명령 버전, hosted 결과, 과거 예외, 저장소 설정, resource 사용량, 비용
증거는 구현 PR 병합 후 기록합니다.

## AI 지원

- **계획/구현/초안 작성:** OpenAI Codex
