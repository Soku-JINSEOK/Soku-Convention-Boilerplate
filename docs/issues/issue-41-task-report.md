# 📝 Issue 41 Task Report

## Goal and Background

[Issue #41](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/41)
requires a reviewed boilerplate successor to immutable `v1.0.0`. Current
`main` contains the generated JavaScript workflow formatting fix, the `tmp`
dependency override, the Jackson BOM update, the verification guidance, and the
aggregate Validation Gate, but downstream projects cannot consume those fixes
through an immutable boilerplate release yet.

### Corrective status update (2026-07-21)

Boilerplate `v1.0.1` and CLI `soku/v0.1.2` are now published, but public smoke
identified a remaining defect: the immutable CLI embeds a second CI renderer and
produces a JavaScript workflow whose quoted `node-version` fails the generated
project's Prettier check. This report therefore remains open and the completion
target is the companion corrective pair `v1.0.2` + `soku/v0.1.3`. Existing
`v1.0.0`, `v1.0.1`, and `soku` tags remain immutable.

The source-authoritative renderer, security Validation Gate, Dependabot/tag
protections, and the two-repository security work are tracked before public
publication. Boilerplate #54 and #56 and Control Plane #19, #26, and #27 are
complete; their public-provider and supply-chain evidence now form release
prerequisites rather than blockers.

GitHub Issue #41 was manually closed on 2026-07-22 before the corrective pair
was published. That repository state does not satisfy this report's completion
criteria. Repository-operation approval was granted to reopen the Issue and
return it to active Project status for the corrective release.

The published `soku/v0.1.2` CLI remains the current historical client. This task
prepares, validates, publishes, and publicly exercises the corrective
`v1.0.2`/`soku/v0.1.3` pair without moving or reusing any existing tag.

The original `v1.0.1` preparation and evidence sections below are retained as
the historical release record. Where they name `v1.0.1` as the completion
target, the corrective status update above supersedes that target.

## Proposed Approach

Use a PATCH release because the candidate fixes template and verification
defects without changing the public lifecycle contracts. Preserve `core-v1`,
`index-v2`, `manifest-v1`, `provider-v1`, profile composition, ownership rules,
and mergeable paths. Record `soku/v0.1.1` as core-lifecycle compatible and
`soku/v0.1.2` as the recommended fully supported client, including optional
legacy provider `ref`.

Document the exact six catalog-managed sources changed since `v1.0.0`: shared
downstream CI, Java `pom.xml`, JavaScript `package.json` and lockfile, and Python
`pyproject.toml` and lockfile. Existing downstream edits to core-managed files
continue to use the established lifecycle conflict rules. No files are removed,
no ownership transitions occur, and no mergeable path changes. Record
`Companion tag: none` because the CLI and boilerplate release axes remain
independent.

The release-preparation work merges with `Related to #41`; the issue remains
open until the signed tag, public Release, lifecycle smoke, security evidence,
and final evidence PR are complete. Creating or pushing `v1.0.1` requires a
separate explicit publication approval after the preflight succeeds.

## Planned Implementation

1. Add `docs/releases/v1.0.1.md` and register the task report and release record
   in the repository-hygiene required-file list.
2. Define source commit, catalog/profile/provider contracts, CLI compatibility,
   the six changed catalog-managed sources, unchanged ownership/removal/merge
   behavior, conflict handling, manual action, and `Companion tag: none` in the
   release record.
3. Update `README.md`, `README.ko.md`, `README.ja.md`, and
   `VERIFICATION_GUIDE.md` to distinguish immutable `v1.0.0` limitations from
   the proposed `v1.0.1` verified baseline and recommend `soku/v0.1.2`.
4. Reapply only the Issue #41 baseline intent from local commit `1436435`; do
   not merge `agent/preserve-current-work` or include its three unrelated
   commits.
5. Run Markdown, YAML, actionlint, release-note contract, release-tag
   regression, sync parity, whitespace, full Go/lifecycle/package, every
   runtime-template gate, dependency/vulnerability, secret, and license checks.
6. Merge the release-preparation PR with `Related to #41`, verify post-merge
   `main`, and manually dispatch the Release workflow with
   `boilerplate-tag=v1.0.1` and `cli-tag=soku/v0.1.2`. The publish job must skip.
7. After separate publication approval, create and verify the signed annotated
   `v1.0.1` tag from the reviewed release record and push only that tag.
8. Verify signed-tag identity, the full tag-triggered Validation Gate, and the
   GitHub Release. The boilerplate Release must contain no CLI archives.
9. Use published `soku/v0.1.2` to run public-source `init --yes --verify`,
   `status`, same-release `diff`, and no-op `upgrade` across JavaScript, Python,
   Go, and Java. Separately initialize `v1.0.0`, confirm `diff v1.0.1` reports
   changes, upgrade, and verify final status and generated stacks.
10. Record lifecycle and security evidence in a follow-up PR, then add final
    evidence to Issue #41, replace its active status with `status:done`, and
    close it as completed.

## Acceptance Criteria

- `v1.0.1` contains the reviewed `tmp`, Jackson, generated-workflow,
  verification-guide, and Validation Gate fixes.
- Release notes preserve `core-v1`, `index-v2`, `manifest-v1`, `provider-v1`,
  and all three profiles; they record both tested CLI versions and recommend
  `soku/v0.1.2`.
- Migration from `v1.0.0` names exactly six changed catalog-managed sources and
  records no removals, ownership transitions, mergeable-path changes, or
  companion tag.
- The release candidate passes every repository, lifecycle, package,
  runtime-template, security, dependency, secret, license, and aggregate gate.
- Manual preflight succeeds with its publish job skipped before publication is
  requested.
- The signed public tag resolves to the reviewed source commit, the
  tag-triggered Release succeeds, and the boilerplate Release has no CLI asset.
- Fresh public-source initialization and `v1.0.0` to `v1.0.1` upgrade smokes
  pass with `soku/v0.1.2` across the four application stacks.
- Existing `v1.0.0` and all existing `soku` tags and Releases remain immutable.

## Approval

- **Status:** `Approved`
- **Approved by:** `Soku-JINSEOK` (explicit implementation approval recorded
  on 2026-07-21 after Draft PR #53 was opened)
- **Approval boundary:** Only this report may be drafted, validated, and opened
  for review before explicit implementation approval. `v1.0.1` publication
  requires a second, separate explicit approval after release preflight.

## Implementation Status

Task report approved. Release-preparation implementation is complete;
boilerplate tag publication remains outside this approval boundary.

The `v1.0.1` release record, required-file registration, and trilingual
pre-publication baseline are implemented on Draft PR #53. The six changed
catalog-managed sources were confirmed directly against immutable `v1.0.0`.
The unrelated commits on `agent/preserve-current-work` were not merged; only
the Issue #41 limitation-warning intent from `1436435` was reapplied against
current `main` and `soku/v0.1.2`.

## Corrective implementation status

- The canonical downstream CI source now uses explicit job markers and
  formatter-compatible scalars; the CLI reads that source instead of embedding
  a second job definition. Legacy marker-free releases remain readable through
  a bounded compatibility parser.
- A public report-hub Provider adoption exposed another `soku/v0.1.2` defect:
  a same-release Provider transition preserved customized mergeable file bytes
  but replaced their manifest baselines with upstream hashes, causing false
  drift immediately after apply. The corrective client preserves the accepted
  downstream baseline for unchanged mergeable paths. A pending-to-connected
  Provider regression covers both `.editorconfig` and `.gitignore`.
- Boilerplate security automation is connected to the required Validation Gate,
  with scheduled scans, Dependabot configuration, CodeQL default setup, and an
  active immutable release-tag ruleset (`19336418`).
- Control Plane Issue #27 is delivered by merged PR #36. Python installs use a
  hash-locked requirements file, the dashboard audit covers development
  dependencies, and hosted Gitleaks, OSV, Action-pin, license, and Terraform
  lock checks passed before merge.
- The next implementation PR must use `Related to #41`; only the final PR after
  public lifecycle evidence may use a closing keyword.

## Verification

### Published smoke follow-up (2026-07-22)

- Published signed tags `v1.0.2` and `soku/v0.1.3` resolve to reviewed source
  commit `94df168861edb3ab3d136a4a0e5dd9b23d49fec5`. Both tag-triggered Release
  runs passed the full repository, runtime-template, security, three-OS
  lifecycle, five-target package, and aggregate gates. The boilerplate Release
  has no assets; the CLI Release has five platform archives plus
  `checksums.txt`, and every downloaded archive passed its published checksum.
- Failed: the required public four-application-stack `init --yes --verify`
  using the published macOS arm64 `soku/v0.1.3` archive and boilerplate
  `v1.0.2`. Go, Java, JavaScript tests, typecheck, lint, and build passed, but
  JavaScript `prettier --check .` rejected generated `.github/workflows/ci.yml`
  and `.golangci.yml`. The CLI correctly aborted before writing the manifest.
- Corrective action: keep both published tags immutable and release a
  boilerplate-only `v1.0.3` patch. Its JavaScript `.prettierignore` excludes
  only the two cross-stack YAML surfaces, which retain dedicated YAML, action,
  and Go lint validation. `soku/v0.1.3` remains the compatible CLI; no new CLI
  release is required.

### Corrective verification update (2026-07-21)

- The first `v1.0.2`/`soku/v0.1.3` validation-only preflight on final provider
  main commit `29cbaeb8daf34d69a09bd10b4e8b7046fc6373bc` produced startup failure
  before creating jobs. The Release caller granted only `contents: read`, while
  the reusable Validation workflow also declares `pull-requests: read`. The
  release-candidate fix passes that read-only permission, updates the manual
  defaults to the corrective pair, and adds a regression test that delivery
  remains restricted to tag pushes. A hosted rerun is required after merge.

- Passed: corrective PR #77 hosted run `29892490977`, including the aggregate
  Validation Gate, current PR Metadata Gate, five-language CodeQL, three-OS
  lifecycle, five-target package snapshot, runtime templates, databases,
  dependency/license/vulnerability checks, Gitleaks, OSV, and sync parity.

- Passed: `soku` and `templates/go` unit tests, vet, format/import checks,
  lifecycle conformance, five-target package reproducibility, release-tag
  regression, Markdown/YAML lint, actionlint, and `git diff --check`.
- Passed: Boilerplate Gitleaks full-history scan (78 commits), OSV Scanner
  v2.4.0 source scan, JavaScript `npm audit`, and the renderer regression suite.
  The repository allowlist covers only the historical synthetic secret fixture
  used by archive-rejection tests.
- Passed: Control Plane actionlint, YAML lint, Python bytecode compilation,
  Gitleaks, and whitespace checks. Hosted CI remains required for its locked
  Python registry tests and dashboard audit; local execution intentionally did
  not export private dependency metadata or install missing validator packages.
- Historical local limitation before publication: `shellcheck`,
  race/golangci-lint, hosted cloud/database gates, and the public lifecycle
  smoke were not yet run. The hosted gates later passed; the public smoke
  produced the `v1.0.3` corrective finding recorded above.

- Passed: targeted Markdown lint and `git diff --check` for the report-only
  Draft PR.
- Passed: changed-document Markdown lint, changed YAML lint, actionlint v1.7.12,
  release-note contract, release-tag regression, contribution-title tests, and
  `git diff --check`.
- Local limitation: PowerShell is unavailable, so sync parity must pass in the
  hosted PR matrix.
- Passed: soku module verification, unit tests, vet, gofmt, goimports, hermetic
  lifecycle conformance, and reproducible five-target package snapshot.
- Local limitation: the environment has no C compiler or locally installed
  golangci-lint, so race and golangci-lint must pass in hosted validation.
- Passed: JavaScript lint/typecheck/test/build/format and `npm audit` with zero
  vulnerabilities.
- Passed: Python ruff/mypy/Black/pytest and `pip-audit` with zero known
  vulnerabilities.
- Passed: `govulncheck` for `soku` and the Go template with no vulnerabilities.
- Passed: OSV-Scanner v2.4.0 offline scan of all five supported repository
  manifests/lockfiles with no issues. Only public vulnerability databases were
  downloaded; repository dependency metadata was not sent to OSV services.
- Passed: Gitleaks v8.30.1 full-history scan of 75 commits with no leaks.
- Passed: license inventory review. The four modules embedded in the CLI match
  `soku/THIRD_PARTY_NOTICES.md`; template dependency inventories introduced no
  new repository redistribution or license-policy blocker.
- Passed: hosted PR Validation run `29797658484`, including Java, MySQL,
  PostgreSQL, gcloud, AWS/Azure, three-OS Go and lifecycle,
  race/golangci-lint, sync parity, and the aggregate Validation Gate. The prior
  run was cancelled by the PR title update and did not identify a product or CI
  defect.
- Pending: merge the `v1.0.3` corrective candidate, run validation-only
  preflight, obtain separate signed-tag delivery approval, and repeat fresh and
  migration public lifecycle smoke against the immutable published Release.

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex

---

## 목표 및 배경

[Issue #41](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/41)은
불변 `v1.0.0`의 검증된 후속 보일러플레이트 Release를 요구합니다. 현재
`main`에는 generated JavaScript workflow 포맷 수정, `tmp` dependency override,
Jackson BOM 수정, 검증 가이드, aggregate Validation Gate가 있지만 downstream은
아직 immutable 보일러플레이트 Release로 이를 사용할 수 없습니다.

공개 `soku/v0.1.2`가 권장 lifecycle client입니다. 이 작업은 `v1.0.0`이나 기존
CLI tag를 이동·재사용하지 않고 보일러플레이트 `v1.0.1`을 준비·검증·발행하고
공개 lifecycle을 확인합니다.

## 제안하는 접근

Public lifecycle contract를 변경하지 않고 template/검증 결함을 수정하므로 PATCH
Release를 사용합니다. `core-v1`, `index-v2`, `manifest-v1`, `provider-v1`, profile
composition, ownership 규칙, mergeable path를 유지합니다. `soku/v0.1.1`은 core
lifecycle 호환, `soku/v0.1.2`는 optional legacy provider `ref`까지 완전히 지원하는
권장 client로 기록합니다.

`v1.0.0` 이후 변경된 catalog-managed source 6개를 정확히 기록합니다: shared
downstream CI, Java `pom.xml`, JavaScript `package.json`/lockfile, Python
`pyproject.toml`/lockfile. 기존 core-managed file의 downstream 수정은 현재 lifecycle
conflict 규칙을 따릅니다. 제거·소유권 전환·mergeable path 변경은 없으며, 두
release 축이 독립적이므로 `Companion tag: none`을 기록합니다.

Release-preparation PR은 `Related to #41`로 병합하고 signed tag, 공개 Release,
lifecycle smoke, security 증거, 최종 증거 PR까지 완료할 때 Issue를 열어 둡니다.
Preflight 성공 후 별도 명시적 발행 승인 없이는 `v1.0.1` tag를 생성하거나 push하지
않습니다.

## 계획된 구현

1. `docs/releases/v1.0.1.md`를 추가하고 task report와 release record를
   repository-hygiene required-file 목록에 등록합니다.
2. Release record에 source commit, catalog/profile/provider contract, CLI 호환성,
   변경된 catalog-managed source 6개, 유지되는 ownership/removal/merge 규칙,
   conflict 처리, manual action, `Companion tag: none`을 기록합니다.
3. `README.md`, `README.ko.md`, `README.ja.md`, `VERIFICATION_GUIDE.md`에서 불변
   `v1.0.0` 제한과 제안 `v1.0.1` verified baseline을 구분하고
   `soku/v0.1.2`를 권장합니다.
4. Local commit `1436435`의 Issue #41 baseline 의도만 재적용하고
   `agent/preserve-current-work`나 다른 세 commit을 병합하지 않습니다.
5. Markdown, YAML, actionlint, release-note contract, release-tag regression, sync
   parity, whitespace, 전체 Go/lifecycle/package, 모든 runtime-template,
   dependency/vulnerability, secret, license 검사를 실행합니다.
6. Release-preparation PR을 `Related to #41`로 병합하고 post-merge `main`을 확인한
   뒤 Release workflow를 `boilerplate-tag=v1.0.1`,
   `cli-tag=soku/v0.1.2`로 수동 실행합니다. Publish job은 반드시 skip돼야 합니다.
7. 별도 발행 승인 후 검토된 Release record로 signed annotated `v1.0.1` tag를
   생성·검증하고 해당 tag만 push합니다.
8. Signed-tag identity, 전체 tag-triggered Validation Gate, GitHub Release를
   확인합니다. Boilerplate Release에는 CLI archive가 없어야 합니다.
9. 공개 `soku/v0.1.2`로 JavaScript, Python, Go, Java를 포함해 public-source
   `init --yes --verify`, `status`, same-release `diff`, no-op `upgrade`를 실행합니다.
   별도 project를 `v1.0.0`으로 초기화하고 `diff v1.0.1`의 변경 보고, upgrade,
   최종 status와 generated stack을 검증합니다.
10. 후속 PR에 lifecycle/security 증거를 기록하고 Issue #41에 최종 증거를 남긴
    뒤 active status를 `status:done`으로 교체하고 completed 사유로 닫습니다.

## 수용 기준

- `v1.0.1`에 검토된 `tmp`, Jackson, generated-workflow, verification-guide,
  Validation Gate 수정이 포함됩니다.
- Release note는 `core-v1`, `index-v2`, `manifest-v1`, `provider-v1`, 세 profile을
  유지하고 검증한 두 CLI 버전을 기록하며 `soku/v0.1.2`를 권장합니다.
- `v1.0.0` migration은 변경된 catalog-managed source 6개를 정확히 명시하고
  제거·소유권 전환·mergeable-path 변경·companion tag가 없음을 기록합니다.
- Release candidate가 전체 repository, lifecycle, package, runtime-template,
  security, dependency, secret, license, aggregate gate를 통과합니다.
- 발행 요청 전에 manual preflight가 성공하고 publish job이 skip됩니다.
- Signed public tag가 검토 source commit을 가리키고 tag-triggered Release가
  성공하며 boilerplate Release에 CLI asset이 없습니다.
- `soku/v0.1.2`로 4개 application stack의 fresh public-source init과
  `v1.0.0` → `v1.0.1` upgrade smoke가 통과합니다.
- 기존 `v1.0.0` 및 모든 기존 `soku` tag/Release는 불변으로 유지됩니다.

## 승인

- **상태:** `Approved`
- **승인자:** `Soku-JINSEOK` (Draft PR #53 생성 후 2026-07-21 명시적 구현
  승인 기록)
- **승인 경계:** 명시적 구현 승인 전에는 이 report의 초안·검증·review 요청만
  허용합니다. `v1.0.1` 발행은 Release preflight 후 두 번째 별도 명시 승인이
  필요합니다.

## 구현 현황

Task report 승인을 기록했습니다. Release-preparation 구현을 완료했으며,
보일러플레이트 tag 발행은 이 승인 범위 밖에 있습니다.

Draft PR #53에 `v1.0.1` Release record, required-file 등록, 다국어 발행 전
baseline을 구현했습니다. Immutable `v1.0.0`과 직접 비교해 변경된
catalog-managed source 6개를 확인했습니다. `agent/preserve-current-work`의 관련 없는
commit은 병합하지 않았고, `1436435`의 Issue #41 제한 경고 의도만 current
`main`과 `soku/v0.1.2` 기준으로 재적용했습니다.

## 검증

- 통과: Report-only Draft PR의 targeted Markdown lint와 `git diff --check`.
- 통과: 변경 문서 Markdown lint, 변경 YAML lint, actionlint v1.7.12,
  release-note contract, release-tag regression, contribution-title test,
  `git diff --check`.
- 로컬 제한: PowerShell이 없어 sync parity는 hosted PR matrix에서 확인합니다.
- 통과: soku module verification, unit test, vet, gofmt, goimports, hermetic
  lifecycle conformance, reproducible 5-target package snapshot.
- 로컬 제한: C compiler와 local golangci-lint가 없어 race/golangci-lint는 hosted
  validation에서 확인합니다.
- 통과: JavaScript lint/typecheck/test/build/format 및 취약점 0건의 `npm audit`.
- 통과: Python ruff/mypy/Black/pytest 및 알려진 취약점 0건의 `pip-audit`.
- 통과: `soku`와 Go template의 `govulncheck` 취약점 0건.
- 통과: 지원되는 repository manifest/lockfile 5개의 OSV-Scanner v2.4.0 offline
  scan 문제 0건. 공개 vulnerability database만 내려받았고 repository dependency
  metadata는 OSV service에 전송하지 않았습니다.
- 통과: Gitleaks v8.30.1 full-history 75 commits scan leak 0건.
- 통과: License inventory review. CLI에 embed된 module 4개는
  `soku/THIRD_PARTY_NOTICES.md`와 일치하고 template dependency inventory에서 새
  repository redistribution/license-policy blocker가 발견되지 않았습니다.
- 통과: Hosted PR Validation run `29797658484`. Java, MySQL, PostgreSQL,
  gcloud, AWS/Azure, 3-OS Go/lifecycle, race/golangci-lint, sync parity,
  aggregate Validation Gate를 모두 포함합니다. 직전 run은 PR title 수정으로
  cancel되었으며 product/CI 결함은 확인되지 않았습니다.
- 대기: post-merge `main`, validation-only Release preflight, 별도 승인 signed-tag
  delivery, public lifecycle/security smoke.

## AI 지원

- **계획/구현/초안 작성:** OpenAI Codex
