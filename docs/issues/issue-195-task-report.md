# Issue #195 Task Report: Downstream Project Sync Audit

## Goal and Background

Validate the optional `github-project-sync` Soku component as a downstream
installation and guarded runtime, without changing this boilerplate's Project
or inventing an external sandbox. The component already has unit and lifecycle
coverage; this task adds an exact-source disposable installation audit,
read-only Project #2 evidence, and an explicit decision at the sandbox apply
boundary.

## Proposed Approach

Build Soku from the exact verified `main` source commit
`9719ab75c54d7e75c99f4f00787b099644f5ae86`. Use isolated temporary Git
repositories and temporary output paths for all filesystem mutation tests.

The audit has three independent lanes:

1. **Disposable lifecycle fixture:** prove component dry-run non-mutation,
   apply, status, diff, repeat-install idempotency, project-owned configuration
   preservation, collision rejection, and rollback.
2. **Operational Project audit:** run `scripts/github-project-sync.mjs --mode audit`
   against the existing user-owned Project with a mode-`0600` redacted report;
   do not apply.
3. **Sandbox apply:** list existing user-owned Projects and proceed only if an
   unmistakably dedicated, empty sandbox already exists. Never create or empty a
   Project, add pull request items, create Issues, broaden a token, or use
   Project #2 for apply verification.

Existing fake-client tests remain the authoritative hermetic coverage for API
mutation sequencing, preservation, collision, stale-read, and partial failure.
Live apply is additional evidence, not permission to bypass those tests.

## Planned Implementation

### Disposable lifecycle fixture

- Record a complete pre-dry-run file listing, SHA-256 inventory, file modes, and
  Git status.
- Initialize a minimal downstream repository from immutable Boilerplate
  `v1.0.5` with a portable Soku profile.
- Install Project Sync with an explicit positive Project number in `--dry-run`
  mode and prove the fixture is unchanged.
- Apply the identical component plan locally with `--yes`; this writes only the
  fixture and performs no GitHub API request.
- Verify the four component outputs, manifest v2/v3 preservation, project-owned
  `.github/project-sync.yml`, clean status, same-release diff, and repeat
  installation no-op.
- Modify only the project-owned configuration and prove lifecycle operations
  preserve it.
- Use separate disposable fixtures for a pre-existing-output collision and an
  injected write failure; prove no partial component or manifest state remains.

### Remote audit and sandbox boundary

- Use the authenticated intended account and the existing narrow Project Sync
  credential source only for read-only audit.
- Store no raw body, token, Authorization header, credential-bearing URL,
  private identifier, or machine path in committed evidence.
- Inspect Project #2 only through audit mode and record an allowlisted redacted
  outcome.
- Use apply mode only against a pre-existing, dedicated, empty user-owned
  sandbox Project after separate approval. Missing, non-empty, ambiguous, or
  organization-owned candidates stop the lane.

## Acceptance Criteria

- [x] Component dry-run is byte-for-byte non-mutating in a disposable fixture.
- [x] Local apply installs only the declared component assets and durable
      manifest metadata.
- [x] Status, diff, repeat install, project-owned configuration preservation,
      collision, and rollback checks pass.
- [x] The generated configuration contains no repository name, historical
      relation, raw body, credential, token, or private identifier.
- [x] Operational Project audit mode passes with a redacted mode-`0600` report and no
      remote mutation.
- [ ] A pre-existing sandbox Project is confirmed user-owned, dedicated, and
      empty before any apply.
- [ ] Sandbox apply, idempotency, preservation, collision, rollback, and
      stale-read behavior are verified without creating Issues or pull request
      Project items.
- [x] Missing or unsuitable sandbox state leaves live apply Blocked and does not
      trigger Project creation or cleanup.
- [ ] Repository tests, Markdown, security, supply-chain, diff, and exact-head
      hosted checks pass.

## Approval

- **Status:** `Approved`
- **Approved by:** `@Soku-JINSEOK` through the repository-wide closeout
  instruction on 2026-08-08.

## Implementation Status

Repository-side verification is complete. Remote work remained read-only. The
only empty user-owned candidate lacked an unmistakable dedicated-sandbox
designation, so live apply was skipped and remains Blocked.

## Verification

### Exact-source lifecycle fixture

- Built Soku from exact source commit
  `9719ab75c54d7e75c99f4f00787b099644f5ae86` and installed the immutable
  Boilerplate `v1.0.5` bootstrap Go profile in a disposable Git repository.
- Component dry-run planned exactly the four declared Project Sync assets and
  left all existing bytes, modes, the manifest, and Git status unchanged.
- Local `--yes` installed only those four assets and manifest metadata. The
  generated configuration contained only the portable owner/Project selector,
  canonical field names, and empty optional backfill values.
- Status reported all managed assets clean, same-release diff was empty, and
  repeated dry-run/apply were no-ops. A project-owned configuration edit was
  preserved byte-for-byte by later lifecycle operations.
- A pre-existing-output fixture was rejected before write with every original
  hash unchanged. An injected second-write permission failure returned the
  rollback exit and restored the original manifest with no partial assets.

### Read-only remote audit

- Audit output was created with mode `0600`, body storage disabled, body hashes
  enabled, raw metadata excluded, zero failures, and zero remote writes.
- The audit identified one public Dependabot relation already represented by
  neighboring dependency updates. The canonical configuration now explicitly
  maps PR `#143` to tracking Issue `#69`; audit mode reports the intended
  reconciliation but it was not applied.
- Twenty-four closed Issues not present in the Project remained warnings; no
  items were added and no Project fields were changed.
- Retrofitting the merged PR body was rejected at the historical-mutation
  safety boundary. No alternate mutation path was attempted.

No Project, Issue, credential, repository secret, release, cloud resource, or
delivery setting was created, deleted, or remotely changed. Repository and
hosted final-head verification are recorded after the documentation commit.

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

optional `github-project-sync` Soku component를 disposable downstream과
read-only remote audit로 검증합니다. Boilerplate Project에 apply하거나 외부 sandbox를
새로 만들지 않습니다.

## 제안하는 접근

검증된 exact `main` source commit으로 Soku를 build합니다. disposable lifecycle
fixture, 운영 Project audit, 기존 empty user-owned sandbox 확인의 세 lane을 분리합니다.
sandbox가 없거나 모호하면 만들거나 비우지 않고 live apply lane을 중단합니다.

## 계획된 구현

- dry-run byte 불변, local component apply, status/diff/repeat-install을 검증합니다.
- project-owned config 보존, collision, rollback을 별도 temporary fixture에서
  검증합니다.
- 운영 Project는 redacted audit만 실행합니다.
- 기존 dedicated empty sandbox에서만 별도 승인된 apply를 허용합니다.
- Issue나 PR Project item, credential, Project를 만들지 않습니다.

## 수용 기준

- [x] disposable dry-run이 byte-for-byte non-mutating임
- [x] local apply가 component asset과 manifest metadata만 설치함
- [x] status/diff/idempotency/config preservation/collision/rollback이 성공함
- [x] generated config에 repository-specific 또는 secret data가 없음
- [x] 운영 Project redacted audit가 remote mutation 없이 성공함
- [ ] 기존 user-owned dedicated empty sandbox가 확인됨
- [ ] sandbox apply와 preservation/stale-read가 Issue·PR item 생성 없이 성공함
- [x] sandbox가 없거나 부적합하면 live apply를 Blocked로 유지함
- [ ] repository와 hosted final-head 검증이 성공함

## 승인

- **상태:** `Approved`
- **승인자:** 2026-08-08 repository-wide closeout 지시의 `@Soku-JINSEOK`

## 구현 현황

repository-side 검증은 완료했습니다. empty user-owned candidate는 dedicated sandbox로
명확히 식별되지 않아 live apply를 수행하지 않았고 Blocked로 유지합니다.

## 검증

exact source build와 disposable dry-run/apply, status/diff/repeat install,
project-owned config 보존, collision rejection, injected-failure rollback이 모두
성공했습니다. remote audit report는 mode `0600`, body 저장 비활성화, body hash
활성화, raw metadata 제외, failure 0, remote write 0을 확인했습니다. 공개 dependency
관계 `PR #143 → Issue #69`는 canonical mapping으로 기록했지만 audit가 제시한 remote
reconciliation은 apply하지 않았습니다. 과거 merged PR 본문 수정은 safety boundary에서
거부됐고 우회하지 않았습니다. Project item·Issue·credential·secret·release·cloud
resource·delivery setting을 만들거나 변경하지 않았습니다.

## 공개 적합성 검토

- [x] credential, token, private key, credential이 포함된 URL이 없음
- [x] 비공개 저장소·프로젝트·제품 이름이 없음
- [x] cloud project ID, 계정 번호, service URL, image URI, revision 식별자가 없음
- [x] 개인 청구·구독·budget·결제 상태 정보가 없음
- [x] 개인 이메일·전화번호·주소·로컬 절대 path가 없음
- [x] 비공개 Issue·PR·Project·control-plane 식별자가 없음

## AI 지원

- **계획/구현/초안 작성:** OpenAI Codex (GPT-5.6)
