# 📝 Task Report: Bind Provider Bundles to Fetched Revisions

## Goal and Background

[Issue #44](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/44)
corrects Provider API v1 so that an untrusted bundle cannot claim which Git
revision supplied it. The exact `--integration-ref` fetched by the CLI must be
the revision authority, while the legacy bundle `ref` remains readable for
backward compatibility without deciding whether an integration is connected.

The work also needs executable evidence against an external provider commit,
a privacy-preserving pending artifact contract, and an independently reviewed
`soku/v0.1.2` patch release.

## Proposed Approach

Keep the current exact-revision CLI input and validate it before any network or
filesystem mutation. Fetch the provider archive using that value and carry the
validated fetched revision as trusted transition context. Make Provider API v1
`ref` optional and mark it deprecated in the published schema. When a legacy
bundle includes `ref`, accept only the existing lowercase full-commit form, but
do not compare it with the fetched revision and do not use it to select
`pending` or `connected`.

Continue to require exact provider ID, source, configuration hash, schema,
compatibility, ownership, and transaction checks. The pending artifact will
remain hash-only for configuration: it may identify the provider source and
requested commit, but it must not contain raw configuration or secrets. The
lifecycle guide will define a separate provider-controlled submission path for
a deliberately sanitized configuration document, including review, redaction,
and retention expectations.

Add a minimal declarative provider fixture to the existing external
`Soku-JINSEOK/ci-cd-control-plane` repository through its own reviewed change.
Pin the conformance test to that fixture's immutable commit and fetch its real
Git archive. Do not use a branch, tag, bundle-declared `ref`, or local in-memory
substitute as revision authority. This cross-repository fixture is part of the
approval boundary and will not be created or changed before approval.

Prepare code and documentation in a reviewable pull request first. After it is
merged and the full release gate succeeds, present the resolved commit, signed
tag command, release record, artifacts, and checks for a final publication
confirmation. Only then publish the new immutable `soku/v0.1.2` tag and release;
existing tags and releases remain unchanged.

## Planned Implementation

- Update `soku/schema/provider-v1.schema.json` so `ref` is optional,
  deprecated, and strictly validated only when present.
- Refactor provider validation and transition planning so the fetched CLI
  revision is authoritative and a valid legacy bundle `ref` is ignored for
  connection-state decisions.
- Add unit and lifecycle coverage for omitted, matching, mismatching, and
  malformed legacy refs, including rollback and no-write failure behavior.
- Add an immutable external provider fixture in `ci-cd-control-plane` and a
  conformance test that fetches its actual Git commit archive.
- Document the hash-only pending artifact and a separate sanitized
  configuration submission lifecycle in `docs/standards/SOKU_LIFECYCLE.md`.
- Update CLI/provider documentation and verification guidance where the
  authority model or conformance command changes.
- Add the `soku/v0.1.2` compatibility and migration release record, then run
  the complete CLI, lifecycle, dependency, runtime-template, packaging, and
  aggregate validation gates before requesting publication confirmation.

## Acceptance Criteria

| Criterion | Observable evidence |
| --- | --- |
| Schema compatibility | Provider API v1 accepts an omitted `ref`; a present malformed legacy `ref` is rejected. |
| Revision authority | The CLI fetches only the validated `--integration-ref` and records that fetched revision as authoritative. |
| Legacy behavior | A well-formed mismatching bundle `ref` does not change `pending` or `connected`. |
| Connection checks | Provider ID, source, configuration hash, schema, compatibility, ownership, and transaction checks remain enforced. |
| External conformance | A test fetches and validates a provider bundle from an immutable commit in an external Git repository. |
| Privacy contract | Pending state stores only the configuration hash; sanitized configuration follows a separate documented submission process. |
| Release | The reviewed resolved commit passes all gates and is published as a new signed `soku/v0.1.2` release only after final confirmation. |
| Immutability | Existing `soku/v0.1.0`, `soku/v0.1.1`, and boilerplate `v1.0.0` tags/releases are unchanged. |

## Approval

- **Status:** `Approved`
- **Approved by:** User on 2026-07-20
- **Approval boundary:** Provider API implementation, the reviewed external
  fixture change in `ci-cd-control-plane`, and release preparation. Publishing
  `soku/v0.1.2` requires a final confirmation after all evidence is available.

## Implementation Status

In progress. Issue #44 is assigned, belongs to Project #2, and is set to
P1 / M / Engineering / In progress. The user approved the implementation,
external fixture, and release-preparation boundary on 2026-07-20. Publishing
the tag and release remains a separate confirmation.

## Verification

- Confirmed the current Provider API v1 schema requires `ref`.
- Confirmed current connection planning compares bundle `ref` with the CLI
  request ref.
- Confirmed existing provider lifecycle tests use an in-memory fetcher rather
  than a real external Git commit.
- Confirmed the current `soku/v0.1.1` release record declares provider-v1 and
  boilerplate `v1.0.0` compatibility.
- Not run: implementation, conformance, release, or publication gates.

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex

---

## 목표 및 배경

[Issue #44](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/44)는
신뢰할 수 없는 provider bundle이 자신을 제공한 Git revision을 주장하지 못하도록
Provider API v1의 권한 경계를 수정합니다. CLI가 실제 fetch한 정확한
`--integration-ref`만 revision의 권위 있는 값으로 사용하고, 기존 bundle `ref`는
하위 호환성을 위해 읽을 수만 있으며 연결 상태 결정에는 사용하지 않습니다.

또한 외부 provider 실제 커밋을 사용하는 실행 가능한 증거, 개인정보를 보존하는
pending artifact 계약, 독립적으로 검토되는 `soku/v0.1.2` patch release가 필요합니다.

## 제안하는 접근

현재 exact-revision CLI 입력을 유지하고 네트워크 또는 파일 변경 전에 검증합니다.
그 값으로 provider archive를 fetch하고 검증한 fetched revision을 신뢰할 수 있는
transition context로 전달합니다. Provider API v1의 `ref`는 optional 및 deprecated로
변경합니다. legacy bundle에 `ref`가 있으면 기존 lowercase full-commit 형식만
허용하지만 fetched revision과 비교하지 않고 `pending`/`connected` 결정에도
사용하지 않습니다.

Provider ID, source, configuration hash, schema, compatibility, ownership,
transaction 검사는 계속 정확히 적용합니다. pending artifact에는 provider source,
요청 commit과 configuration hash만 허용하며 raw configuration이나 secret을 넣지
않습니다. lifecycle guide에는 의도적으로 sanitize한 configuration 문서를 provider가
관리하는 별도 경로로 제출하는 절차와 검토, redaction, retention 기준을 명시합니다.

외부 `Soku-JINSEOK/ci-cd-control-plane` 저장소에 최소 선언형 provider fixture를
별도 검토 변경으로 추가합니다. conformance test는 해당 fixture의 immutable commit을
pin하고 실제 Git archive를 fetch합니다. branch, tag, bundle의 `ref`, 로컬 in-memory
대체물은 revision authority로 사용하지 않습니다. 이 cross-repository 변경은 이번
승인 범위에 포함하며 승인 전에는 생성하거나 수정하지 않습니다.

먼저 코드와 문서를 검토 가능한 PR로 준비합니다. merge 및 전체 release gate 통과
후 resolved commit, signed tag 명령, release record, artifact와 검사 결과를 제시하여
최종 발행 확인을 받습니다. 그 후에만 새로운 immutable `soku/v0.1.2` tag/release를
발행하며 기존 tag와 release는 변경하지 않습니다.

## 계획된 구현

- `soku/schema/provider-v1.schema.json`에서 `ref`를 optional/deprecated로 변경하고,
  존재할 때만 strict validation을 적용합니다.
- provider validation과 transition planning을 수정하여 fetched CLI revision만
  authoritative하게 사용하고 올바른 legacy bundle `ref`는 연결 판단에서 무시합니다.
- ref 생략, 일치, 불일치, malformed 사례와 rollback/no-write failure를 unit 및
  lifecycle test로 검증합니다.
- `ci-cd-control-plane`에 immutable 외부 provider fixture를 추가하고 실제 Git commit
  archive를 fetch하는 conformance test를 추가합니다.
- `docs/standards/SOKU_LIFECYCLE.md`에 hash-only pending artifact와 별도의 sanitized
  configuration 제출 lifecycle을 문서화합니다.
- authority model 또는 conformance 명령이 바뀌는 CLI/provider 문서와 Verification
  Guide를 갱신합니다.
- `soku/v0.1.2` compatibility/migration release record를 추가하고 CLI, lifecycle,
  dependency, runtime-template, package, aggregate gate를 모두 실행한 후 발행 확인을
  요청합니다.

## 수용 기준

| 기준 | 관찰 가능한 근거 |
| --- | --- |
| Schema 호환성 | Provider API v1은 `ref` 생략을 허용하고, 존재하는 malformed legacy `ref`는 거부합니다. |
| Revision 권위 | CLI는 검증된 `--integration-ref`만 fetch하고 fetched revision을 authoritative하게 기록합니다. |
| Legacy 동작 | 형식이 올바르지만 불일치하는 bundle `ref`는 `pending`/`connected`를 바꾸지 않습니다. |
| 연결 검사 | Provider ID, source, configuration hash, schema, compatibility, ownership, transaction 검사를 유지합니다. |
| 외부 conformance | 외부 Git 저장소의 immutable commit에서 provider bundle을 fetch하여 검증합니다. |
| 개인정보 계약 | Pending에는 configuration hash만 저장하고 sanitized configuration은 별도 문서화된 절차로 제출합니다. |
| Release | 검토된 resolved commit이 모든 gate를 통과하고 최종 확인 후 새 signed `soku/v0.1.2`로 발행됩니다. |
| 불변성 | 기존 `soku/v0.1.0`, `soku/v0.1.1`, boilerplate `v1.0.0` tag/release는 변경하지 않습니다. |

## 승인

- **상태:** `Approved`
- **승인자:** 사용자 (2026-07-20)
- **승인 범위:** Provider API 구현, `ci-cd-control-plane`의 검토된 외부 fixture 변경,
  release 준비입니다. `soku/v0.1.2` 실제 발행은 모든 근거가 준비된 뒤 최종 확인이
  필요합니다.

## 구현 현황

구현 중입니다. Issue #44는 담당자가 지정되었고 Project #2에서
P1 / M / Engineering / In progress입니다. 사용자가 2026-07-20에 구현, 외부 fixture,
release 준비 범위를 승인했습니다. tag와 release 실제 발행은 별도 최종 확인이
필요합니다.

## 검증

- 현재 Provider API v1 schema가 `ref`를 필수로 요구함을 확인했습니다.
- 현재 연결 planning이 bundle `ref`와 CLI request ref를 비교함을 확인했습니다.
- 기존 provider lifecycle test가 실제 외부 Git commit이 아닌 in-memory fetcher를
  사용함을 확인했습니다.
- 현재 `soku/v0.1.1` release record가 provider-v1과 boilerplate `v1.0.0` 호환성을
  선언함을 확인했습니다.
- 미실행: 구현, conformance, release, publication gate.

## AI 지원

- **계획/구현/초안 작성:** OpenAI Codex
