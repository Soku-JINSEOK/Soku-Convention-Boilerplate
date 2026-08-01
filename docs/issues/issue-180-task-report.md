# 🔒️ Issue 180 Task Report

## Goal and Background

Issue [#180](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/180)
isolates Cloud Build validation logging from foundation state and restores
authenticated, read-only PR and security validation. The supplied tracking
plan named #73–#76 as Issues, but those repository numbers are historical pull
requests; they are not mutated or repurposed. Issue #178 remains the live
Cloud Build rollout tracker.

## Proposed Approach

Use one Terraform root and one matching backend prefix for the three logging
resources. Enforce the plan boundary with an allowlist-based JSON verifier.
Make only Pull Request Policy and Security event-driven, authenticate current
PR API metadata against trusted event identity, and run history-aware security
checks with immutable scanner images.

## Planned Implementation

- Add `infra/gcp/cloud-build-logging` with a regional 30-day bucket, sink, and
  disabled exclusion.
- Normalize both Google provider lockfiles to 6.50.0 Linux/macOS checksums.
- Add exact-plan and historical-baseline verifiers with adversarial tests.
- Integrate contribution-title enforcement into Pull Request Policy and remove
  its duplicate repository workflow.
- Execute PR policy and historical scan controls from a base-SHA trusted
  checkout while inspecting the authenticated PR head in a separate checkout.
- Make general CI manual/reusable and keep delivery reusable-only.
- Update Cloud Build, local verification, Dependabot, supply-chain, and
  operational documentation for the second Terraform root.

## Acceptance Criteria

- Only the three reviewed logging creates are accepted.
- IAM, updates, deletes, `_Required`, enabled exclusions, and unexpected
  resources fail closed.
- PR repository, number, head repository, and head SHA must match both the
  trusted event and current API response.
- API files are mode 0600, removed on exit, and checkout credentials are not
  persisted.
- Full-history Gitleaks and OSV scanners are digest-pinned; historical baseline
  ancestry and raw bytes are verified.
- A PR cannot replace the policy scripts, historical baseline verifier, or
  Gitleaks configuration that evaluates its own head.
- No apply, import, trigger cutover, IAM, delivery, or merge is performed
  without a later explicit approval.

## Approval

- **Status:** `Approved`
- **Approved by:** User-provided implementation plan

## Implementation Status

Implementation is in review on Ready PR #181. Review hardening is prepared
locally to execute policy and historical security controls from the trusted
base SHA while inspecting the PR head separately. The updated head must be
pushed and its hosted Policy and Security checks must pass before merge. Live
Terraform plan, apply, and merge remain out of scope.

## Verification

- Review-hardened local tree: Node policy, workflow, verification, and
  supply-chain suites passed, 126 tests.
- Review-hardened local tree: Python Terraform-plan and historical-baseline
  suites passed, 7 tests.
- The plan verifier now rejects cross-project sinks, redirected destinations,
  and altered build-log filters in addition to the existing action/resource
  boundary.
- YAML lint, actionlint, shell syntax, supply-chain verification, historical
  baseline verification, and `git diff --check` passed after review hardening.
- The trusted-checkout workflow regression test proves that policy code,
  baseline verification, and Gitleaks configuration resolve from the base SHA,
  not from the PR head being evaluated.
- The pre-review remote-head evidence below remains historical until the
  review-hardened head is pushed and rerun.
- Node PR identity, policy, supply-chain, workflow, Dependabot, and Cloud Build
  suites: passed, 43 tests.
- Python Terraform-plan and historical-baseline suites: passed, 6 tests.
- Historical baseline verification and supply-chain verifier: passed.
- Terraform 1.15.3 format/init/validate for both roots: passed.
- Existing Terraform test suite: passed, 4 tests.
- `git diff --check`: passed.
- Draft hosted Security: all history, dependency, Go vulnerability, and OSV
  jobs passed.
- Global Cloud Build validation: passed on commit `b503b30`.
- Controlled invalid PR relation: policy run `30627794387` failed; the body was
  immediately restored and run `30627838025` passed.
- Latest Draft SHA `0e871ec`: Policy, Security, CodeQL, and global Cloud Build
  passed before the Ready transition.
- Ready event: Policy run `30630685703` and all jobs in Security run
  `30630685730` passed.

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

Issue #180은 Cloud Build validation logging을 foundation state에서 분리하고
인증된 read-only PR/security 검증을 복구합니다. 제공된 계획의 #73–#76은
Issue가 아니라 과거 PR이므로 변경하지 않으며, #178은 실제 rollout 추적
항목으로 유지합니다.

## 제안하는 접근

logging 3개 리소스 전용 Terraform root/backend prefix와 allowlist 기반 plan
검사기를 사용합니다. Pull Request Policy와 Security만 event-driven으로
유지하고 PR API/event identity 및 full-history 보안 검사를 인증합니다.

## 계획된 구현

- 전용 logging root와 provider 6.50.0 lockfile 추가
- 정확한 3-create plan 및 historical baseline 검사기와 회귀 테스트 추가
- contribution-title을 PR policy에 통합하고 중복 workflow 제거
- CI, Cloud Build, Dependabot, supply-chain 및 운영 문서 갱신

## 수용 기준

허용된 3개 create 외의 plan은 실패하고, PR identity/SHA 불일치와 baseline
변조가 실패하며, credential·임시 파일 경계가 검증되어야 합니다.

## 승인

- **상태:** `Approved`
- **승인자:** 사용자가 제공한 구현 계획

## 구현 현황

Ready PR #181은 review 중입니다. PR 자체가 검증 script와 Gitleaks 설정을
교체할 수 없도록 base SHA trusted checkout과 PR head checkout을 분리하는
보강을 로컬에서 준비했습니다. 보강된 head의 push 및 hosted 재검증 전에는
merge하지 않습니다. live Terraform plan, apply, merge는 범위 밖입니다.

## 검증

review 보강 tree에서 Node 126개와 Python 7개 테스트, YAML/actionlint,
supply-chain, historical baseline 및 diff 검사가 통과했습니다. 기존 remote
head의 Terraform 4개 테스트와 두 root init/validate, hosted
Security/Policy/global Cloud Build 증거는 보강 head가 push되어 hosted checks를
다시 통과할 때까지 역사적 증거로만 유지합니다.

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
