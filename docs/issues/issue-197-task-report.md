# Issue #197 Task Report: Project Sync Credential Rotation

## Goal and Background

Issue #197 requires an operational, least-privilege procedure for the dedicated
`PROJECT_SYNC_TOKEN` used by GitHub Project Sync. The existing guide listed the
permission set and a short rotation outline, but it did not make each mutation,
abort condition, rollback path, or durable redaction boundary independently
reviewable.

## Proposed Approach

Keep credential material and credential lifecycle mutations outside repository
automation. Add one authoritative English-only operational runbook that:

- fixes the exact four-part permission matrix;
- models old and replacement credentials as an overlap window;
- requires direct and post-secret-replacement audits before revocation;
- separates optional apply from credential verification;
- defines fail-closed abort and rollback behavior; and
- allowlists only a UTC date and redacted outcomes for durable evidence.

Add a regression test so future documentation changes cannot silently drop the
permission boundary, reorder rotation phases, permit a broad fallback, or embed
a token-shaped value.

## Planned Implementation

1. Add `docs/guides/PROJECT_SYNC_CREDENTIAL_RUNBOOK.md`.
2. Link it as the authoritative procedure from
   `docs/guides/GITHUB_PROJECT_SYNC.md`.
3. Extend `scripts/github-project-sync.test.mjs` with static runbook contract
   checks.
4. Run Markdown, Node, Project Sync, supply-chain, and secret-scanning checks.
5. Leave live credential creation, Actions secret replacement, audit with a
   replacement credential, apply, and revocation unperformed until the owner
   supplies the external credential and separate mutation approvals.

## Acceptance Criteria

- [x] The permission matrix names Metadata read, Issues read/write, Pull
      requests read/write, and authenticated-user Projects read/write.
- [x] Broader repository, Actions, workflow, release, organization, and billing
      permissions are explicitly rejected.
- [x] Replacement audit, secret replacement, post-replacement audit, optional
      apply, and revocation are ordered separate phases.
- [x] The old credential remains valid until replacement verification succeeds.
- [x] Abort, rollback, stale-read, mutation-error, and revocation-failure paths
      are documented.
- [x] Durable evidence is restricted to a UTC date and redacted outcomes.
- [x] Static regression coverage fixes the contract and rejects token-shaped
      literals.
- [ ] A replacement credential passes the direct audit.
- [ ] The repository secret is replaced and the post-replacement audit passes.
- [ ] Any separately approved apply is verified, and the old credential is
      revoked.

## Approval

- **Status:** `Approved`
- **Approved by:** `@Soku-JINSEOK` through the repository-wide closeout
  instruction on 2026-08-08.

## Implementation Status

The repository-owned runbook, authoritative link, and static contract are
implemented. Live credential operations remain externally blocked and are not
claimed as complete.

## Verification

- `node --test scripts/github-project-sync.test.mjs` — passed, including the
  permission, phase-order, broad-fallback, link, and token-literal assertions.
- `node --test scripts/pull-request-policy.test.mjs
  scripts/authenticated-pr-metadata.test.mjs` — passed.
- `node scripts/verify-supply-chain.mjs` — passed; 37 protected files and 11
  update targets.
- `markdownlint-cli2@0.22.1` across 124 Markdown files — passed with zero
  errors.
- `gitleaks@v8.30.0 dir . --redact --no-banner` — passed; 2.13 MB scanned and no
  leaks found.
- `git diff --check` — passed.

No live credential, repository secret, Project apply, or revocation operation
is part of repository-local verification.

## Public Disclosure Review

- [x] No credentials, tokens, private keys, or credential-bearing URLs
- [x] No private repository, project, or product names
- [x] No cloud project IDs, account numbers, service URLs, image URIs, or
      revision identifiers
- [x] No personal billing, subscription, budget, or payment-status information
- [x] No personal email, phone, address, or local absolute path
- [x] No private Issue, PR, Project, or control-plane identifiers

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex (GPT-5.6)

---

## 목표 및 배경

Issue #197은 GitHub Project Sync 전용 `PROJECT_SYNC_TOKEN`의 최소 권한 설정과
교체·검증·폐기 절차를 요구합니다. 기존 guide에는 권한과 짧은 순서만 있어 각
mutation, 중단 조건, rollback, 영구 evidence redaction 경계를 독립적으로 검토하기
어려웠습니다.

## 제안하는 접근

credential 값과 lifecycle mutation은 repository automation 밖에 유지합니다. 정확한
4개 권한, old/replacement credential 공존 구간, 교체 전후 audit, 별도 apply 승인,
fail-closed rollback, UTC 날짜와 redacted outcome만 남기는 evidence 계약을 하나의
영문 운영 runbook으로 고정합니다.

## 계획된 구현

1. Project Sync credential 운영 runbook을 추가합니다.
2. 기존 Project Sync guide에서 이를 authoritative procedure로 연결합니다.
3. 권한·단계 순서·broad-token 거부·token-shaped literal 부재를 정적 test로
   고정합니다.
4. Markdown, Node, Project Sync, supply-chain, secret scan을 실행합니다.
5. 실제 credential 생성·secret 교체·audit/apply·기존 credential 폐기는 외부
   credential과 별도 mutation 승인이 있을 때까지 수행하지 않습니다.

## 수용 기준

- [x] 정확한 최소 권한 matrix와 거부 권한이 문서화됨
- [x] replacement audit부터 old credential revoke까지 순서가 분리됨
- [x] abort, rollback, stale-read, mutation error가 문서화됨
- [x] 영구 evidence가 UTC 날짜와 redacted outcome으로 제한됨
- [x] 정적 regression test가 계약을 고정함
- [ ] replacement credential을 사용한 실제 audit가 성공함
- [ ] repository secret 교체 후 audit가 성공함
- [ ] 필요한 apply 검증 후 기존 credential이 폐기됨

## 승인

- **상태:** `Approved`
- **승인자:** 2026-08-08 repository-wide closeout 지시의 `@Soku-JINSEOK`

## 구현 현황

repository-owned runbook, authoritative link, static contract 구현은 완료했습니다.
실제 credential 작업은 외부 적용 대기로 남기며 완료로 주장하지 않습니다.

## 검증

- Project Sync contract test, PR policy/authenticated metadata test,
  supply-chain verifier가 성공했습니다.
- Markdown 124개 file은 오류 0건, working-tree Gitleaks는 leak 0건으로
  성공했습니다.
- `git diff --check`가 성공했습니다.

repository-local 검증은 credential, Actions secret, Project apply, revocation을
변경하지 않습니다.

## 공개 적합성 검토

- [x] credential, token, private key, credential이 포함된 URL이 없음
- [x] 비공개 저장소·프로젝트·제품 이름이 없음
- [x] cloud project ID, 계정 번호, service URL, image URI, revision 식별자가 없음
- [x] 개인 청구·구독·budget·결제 상태 정보가 없음
- [x] 개인 이메일·전화번호·주소·로컬 절대 경로가 없음
- [x] 비공개 Issue·PR·Project·control-plane 식별자가 없음

## AI 지원

- **계획/구현/초안 작성:** OpenAI Codex (GPT-5.6)
