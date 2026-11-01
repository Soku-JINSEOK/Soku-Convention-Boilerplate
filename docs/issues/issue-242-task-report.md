# 📝 Dependency Security Repair — Issue #242

## Goal and Background

Update the affected runner and stack-template dependency constraints and lock
files to remove the vulnerable resolutions identified in the post-integration
dependency queue. This work is related to [Issue #242](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/242).

## Proposed Approach

Update only the relevant dependency manifests and lock files. Keep direct
requirements aligned with their resolved lock roots and verify that vulnerable
installed copies are absent. Preserve the affected package APIs and supported
Node.js runtime.

## Planned Implementation

- Update six dependency manifest/lock files across the manual runner,
  JavaScript/TypeScript template, and Python template.
- Add this report as the seventh path to satisfy the repository's PR metadata
  contract; the report does not alter runtime behavior.
- Review the exact dependency diff and record executable validation separately.

## Acceptance Criteria

- Direct dependency declarations match the lock roots and resolved versions.
- The targeted vulnerable lockfile copies are absent.
- Independent review finds no source-visible must-fix regression.
- Required local and hosted install, lint, typecheck, test, build, format, npm,
  Python, and OSV gates are recorded on the exact pull request head.
- The pull request stays in Draft until required validation, signed-commit
  compliance, and owner merge review.

## Approval

- **Status:** `Approved`
- **Approved by:** Existing owner-authorized dependency reconciliation is
  recorded in Issue #242. The owner's 2026-11-01 instruction authorizes
  delivery only: commit, push, pull request creation, and GitHub review
  submission. It is not recorded as retroactive approval of source
  implementation.

## Implementation Status

The six dependency changes were already present on the isolated candidate
branch and passed an independent static post-review. This seventh file records
the work and evidence required by the repository's pull request metadata
policy; it does not change runtime behavior.

## Verification

- Independent static review of the six-file dependency patch: PASS.
- `git diff --check` for the six dependency files: PASS.
- The task report was reviewed separately; final staged seven-path
  `git diff --cached --check`: PASS.
- This delivery commit will be unsigned; the repository's signed-commit rule
  remains a prerequisite for merge.
- Local executable install, lint, typecheck, test, build, format, `npm audit`,
  `pip-audit`, and OSV checks were not run for this delivery step. Required
  checks must report against the final pull request head.

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

통합 후속 의존성 이슈에서 확인된 취약 버전을 제거하기 위해 영향을 받는
runner와 stack template의 의존성 제약 및 lock 파일을 갱신합니다. 이 작업은
[Issue #242](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/242)와
관련 있습니다.

## 제안하는 접근

관련 manifest와 lock 파일만 갱신합니다. 직접 의존성 선언과 lock root를
일치시키고, 취약한 설치 버전이 남지 않았는지 확인합니다. 기존 package API와
지원 Node.js 런타임을 유지합니다.

## 계획된 구현

- 수동 runner의 `fast-uri` 의존성을 갱신합니다.
- JavaScript/TypeScript template의 Vitest 및 `minimatch` 해석 버전을 갱신합니다.
- Python template의 Black 버전과 lock을 갱신합니다.
- 정확한 의존성 diff를 검토하고 실행 검증 결과를 별도로 기록합니다.

## 수용 기준

- 직접 의존성 선언이 lock root 및 해석된 버전과 일치합니다.
- 대상 취약 버전이 lock 파일에서 제거됩니다.
- 독립 검토에서 소스 코드상 필수 수정 회귀가 발견되지 않습니다.
- 필수 로컬·호스티드 설치, 린트, 타입 검사, 테스트, 빌드, 포맷, npm,
  Python, OSV 검증이 최종 PR head에 기록됩니다.
- 필수 검증, signed-commit 정책 충족, 소유자 병합 검토가 끝날 때까지 PR은
  Draft로 유지합니다.

## 승인

- **상태:** `Approved`
- **승인자:** 기존 소유자 승인 의존성 정리는 Issue #242에 기록되어 있습니다.
  2026-11-01 소유자 지시는 커밋·푸시·PR 생성·GitHub 리뷰 제출만 승인하며,
  이미 이루어진 소스 구현을 소급 승인한 것으로 기록하지 않습니다.

## 구현 현황

의존성 변경 6개 파일은 격리된 후보 브랜치에 이미 반영되어 독립 정적 검토를
통과했습니다. 일곱 번째 파일인 이 보고서는 변경과 검증 근거를 기록하고
저장소 PR 메타데이터 정책을 충족하며 runtime 동작은 바꾸지 않습니다.

## 검증

- 6개 의존성 파일에 대한 독립 정적 검토: 통과.
- 6개 의존성 파일에 대한 `git diff --check`: 통과.
- 작업 보고서는 별도로 검토했고, 최종 staged 파일 7개에 대한
  `git diff --cached --check`도 통과했습니다.
- 이번 전달 커밋은 서명하지 않으며, 저장소의 signed-commit 정책은 병합 전
  필수 조건으로 남습니다.
- 이번 전달 단계에서는 로컬 실행 설치, 린트, 타입 검사, 테스트, 빌드, 포맷,
  `npm audit`, `pip-audit`, OSV 검증을 실행하지 않았습니다. 필수 검증은 최종 PR
  head에 대해 보고되어야 합니다.

## 공개 적합성 검토

- [x] credential, token, private key, credential이 포함된 URL이 없음
- [x] 비공개 저장소·프로젝트·제품 이름이 없음
- [x] cloud project ID, 계정 번호, service URL, image URI, revision
      식별자가 없음
- [x] 개인 청구·구독·budget·결제 상태 정보가 없음
- [x] 개인 이메일·전화번호·주소·로컬 절대 경로가 없음
- [x] 비공개 Issue·PR·Project·control-plane 식별자가 없음

## AI 지원

- **계획/구현/초안 작성:** OpenAI Codex
