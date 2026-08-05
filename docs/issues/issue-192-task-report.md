# Issue #192 Task Report

## Goal and Background

[Issue #192](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/192)
defines an opt-in GitHub Project and metadata synchronization component for the
Soku lifecycle CLI. Plain `soku init` must remain unchanged, while downstream
repositories can install the reviewed component without copying
repository-specific configuration from this boilerplate.

The component contract is portable and does not select a production runtime.
Any GCP or Cloud Build execution is outside this task's mutation scope; this
task verifies the CLI contract and the generated component assets.

## Proposed Approach

- Add `github-project-sync` as a first-party Soku core component with catalog
  version `1` and a project-owned `.github/project-sync.yml` configuration.
- Expose the opt-in `--project-sync` lifecycle option and require a positive
  Project number for non-interactive installation.
- Generate portable assets that resolve the repository at runtime, use an
  authenticated user-owned Project, and contain no historical Issue/PR
  mappings, credentials, tokens, or raw metadata.
- Keep the component inactive by default, use audit mode by default, preserve
  project-owned configuration, and apply the existing journal, collision,
  backup, rollback, and manifest-last lifecycle rules.

## Planned Implementation

- Add the component catalog, generated workflow, configuration, runtime, and
  focused test assets.
- Preserve plain `soku init` behavior and support dry-run, transactional
  `--yes`, idempotent repeated installation, and actionable validation errors.
- Migrate manifest v1 to v2 and preserve component metadata through `status`,
  `diff`, and `upgrade`.
- Detect unmanaged or repository-specific Project Sync files as collisions
  instead of adopting or overwriting them.
- Validate portability, lifecycle behavior, security boundaries, and existing
  CLI/provider/completion/manual-component regressions.

## Acceptance Criteria

1. Plain `soku init` remains behaviorally unchanged.
2. `--project-sync --dry-run` writes nothing and lists every planned file.
3. `--yes` installs the component transactionally.
4. Non-interactive use without a Project number fails with an actionable error.
5. Manifest v1-to-v2 migration and component metadata validate correctly.
6. `status`, `diff`, and `upgrade` detect drift and preserve project-owned
   configuration.
7. Generated files contain no repository-specific Issue numbers or
   credentials.
8. Workflow execution is disabled by default and audit is the default mode.
9. Existing Project items and custom labels remain untouched.
10. No GCP, Cloud Build, Cloud Run, Storage, or Artifact Registry resource is
    created by this feature.
11. Existing CLI, provider, completion, and manual-component tests remain
    green.
12. Fresh-repository and collision fixtures cover downstream installation.

## Approval

- **Status:** `Pending`
- **Approved by:** `None`

## Implementation Status

The implementation is present in PR #193 and is undergoing contract and
governance verification. Approval remains pending because no explicit approval
record was available when this report was prepared.

## Verification

- Project Sync Node tests: 13 passed.
- Soku Go tests: `GOCACHE=/tmp/soku-go-cache go test ./...` passed.
- Full local Node regression set: 105 passed, 0 failed.
- Repository policy and governance regression selection: 102 passed, 0 failed.
- NPM wrapper tests: 7 passed with an isolated cache.
- `soku/scripts/package_test.sh`: all five archives and checksums are valid and
  reproducible.
- `go mod verify`: all modules verified.
- `git diff --check`: passed.

Hosted Cloud Build results are reported as external evidence only; this task
does not change cloud configuration.

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

[Issue #192](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/192)은
Soku lifecycle CLI에 선택형 GitHub Project 및 metadata 동기화 component를
추가하기 위한 작업입니다. 일반 `soku init` 동작은 유지하고, downstream
repository가 boilerplate의 repository-specific 설정을 복사하지 않아도
검토된 component를 설치할 수 있어야 합니다.

이 component 계약은 portable하며 production runtime을 선택하지 않습니다.
GCP 또는 Cloud Build 실행은 이번 변경의 mutation 범위 밖이며, CLI 계약과
생성되는 component asset을 검증합니다.

## 제안하는 접근

- `github-project-sync`를 catalog version `1` 및 project-owned
  `.github/project-sync.yml` 설정을 갖는 Soku first-party core component로
  추가합니다.
- 선택형 `--project-sync` lifecycle option을 제공하고, non-interactive
  설치에는 양의 Project 번호를 요구합니다.
- runtime에서 repository를 해석하고 authenticated user-owned Project를
  사용하며, historical Issue/PR mapping·credential·token·raw metadata를
  포함하지 않는 portable asset을 생성합니다.
- 기본값은 비활성 및 audit mode로 유지하고, project-owned 설정과 기존
  journal·collision·backup·rollback·manifest-last lifecycle 규칙을
  보존합니다.

## 계획된 구현

- component catalog, generated workflow, 설정, runtime, focused test asset을
  추가합니다.
- plain `soku init`을 보존하고 dry-run, transactional `--yes`, 반복 설치의
  idempotency, 실행 가능한 validation error를 지원합니다.
- manifest v1을 v2로 migration하고 `status`, `diff`, `upgrade`에서 component
  metadata를 보존합니다.
- unmanaged 또는 repository-specific Project Sync file은 adopt/overwrite하지
  않고 collision으로 보고합니다.
- portability, lifecycle 동작, security boundary 및 기존 CLI/provider/
  completion/manual-component regression을 검증합니다.

## 수용 기준

1. 일반 `soku init` 동작이 바뀌지 않습니다.
2. `--project-sync --dry-run`은 파일을 쓰지 않고 모든 예정 파일을 표시합니다.
3. `--yes`는 component를 transactionally 설치합니다.
4. Project 번호가 없는 non-interactive 실행은 실행 가능한 오류를 반환합니다.
5. manifest v1-to-v2 migration과 component metadata가 올바르게 검증됩니다.
6. `status`, `diff`, `upgrade`가 drift를 감지하고 project-owned 설정을
   보존합니다.
7. 생성 파일에 repository-specific Issue 번호나 credential이 없습니다.
8. workflow 실행은 기본 비활성이고 audit이 기본 mode입니다.
9. 기존 Project item과 custom label을 변경하지 않습니다.
10. 이 기능으로 GCP, Cloud Build, Cloud Run, Storage, Artifact Registry
    resource를 만들지 않습니다.
11. 기존 CLI, provider, completion, manual-component test가 통과합니다.
12. fresh repository 및 collision fixture가 downstream 설치를 검증합니다.

## 승인

- **상태:** `Pending`
- **승인자:** `None`

## 구현 현황

구현은 PR #193에 포함되어 있으며 contract 및 governance 검증을 진행 중입니다.
이 보고서 작성 시점에 명시적인 승인 기록을 확인하지 못했으므로 승인은 계속
pending으로 유지합니다.

## 검증

- Project Sync Node test 13건 통과
- `GOCACHE=/tmp/soku-go-cache go test ./...` 통과
- 전체 로컬 Node regression 105건 통과, 실패 0건
- repository policy 및 governance regression 102건 통과, 실패 0건
- NPM wrapper test 7건 통과(격리 cache 사용)
- `soku/scripts/package_test.sh`의 5개 archive 및 checksum 재현성 통과
- `go mod verify` 전체 module 검증 통과
- `git diff --check` 통과

Hosted Cloud Build 결과는 외부 증거로만 보고하며, cloud 설정은 변경하지
않습니다.

## 공개 적합성 검토

- [x] credential, token, private key, credential이 포함된 URL이 없음
- [x] 비공개 저장소·프로젝트·제품 이름이 없음
- [x] cloud project ID, 계정 번호, service URL, image URI, revision 식별자가 없음
- [x] 개인 청구·구독·budget·결제 상태 정보가 없음
- [x] 개인 이메일·전화번호·주소·로컬 절대 경로가 없음
- [x] 비공개 Issue·PR·Project·control-plane 식별자가 없음

## AI 지원

- **계획/구현/초안 작성:** OpenAI Codex

---

## 目標と背景

[Issue #192](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/192)は、
Soku lifecycle CLIにオプトイン式のGitHub Projectおよびmetadata同期componentを
追加する作業です。通常の `soku init` の動作を変えず、downstream repositoryが
boilerplate固有の設定をコピーせずにレビュー済みcomponentをインストールできる
ようにします。

このcomponent契約はportableであり、本番runtimeを選択しません。GCPまたは
Cloud Buildでの実行は今回のmutation範囲外で、CLI契約と生成されるcomponent
assetを検証します。

## 提案するアプローチ

- catalog version `1` と project-owned `.github/project-sync.yml` 設定を持つ
  `github-project-sync` をSoku first-party core componentとして追加します。
- オプトインの `--project-sync` lifecycle optionを提供し、non-interactiveな
  インストールには正のProject番号を要求します。
- runtimeでrepositoryを解決し、authenticated user-owned Projectを利用する
  portable assetを生成します。historical Issue/PR mapping、credential、token、
  raw metadataは含めません。
- 初期状態を無効、audit modeを既定値とし、project-owned設定と既存の
  journal・collision・backup・rollback・manifest-last lifecycle規則を維持します。

## 計画された実装

- component catalog、generated workflow、設定、runtime、focused test assetを
  追加します。
- 通常の `soku init` を維持し、dry-run、transactionalな `--yes`、繰り返し
  インストールのidempotency、明確なvalidation errorを実装します。
- manifest v1をv2へmigrationし、`status`、`diff`、`upgrade`でcomponent metadataを
  維持します。
- unmanagedまたはrepository-specificなProject Sync fileは採用・上書きせず、
  collisionとして報告します。
- portability、lifecycle動作、security boundary、および既存CLI/provider/
  completion/manual-component regressionを検証します。

## 受け入れ基準

1. 通常の `soku init` の動作が変わりません。
2. `--project-sync --dry-run` は書き込みを行わず、予定される全ファイルを表示します。
3. `--yes` はcomponentをtransactionallyにインストールします。
4. Project番号なしのnon-interactive実行は明確なエラーになります。
5. manifest v1-to-v2 migrationとcomponent metadataが正しく検証されます。
6. `status`、`diff`、`upgrade` はdriftを検出し、project-owned設定を維持します。
7. 生成ファイルにrepository固有のIssue番号やcredentialがありません。
8. workflow実行は既定で無効、auditが既定modeです。
9. 既存のProject itemとcustom labelを変更しません。
10. この機能によってGCP、Cloud Build、Cloud Run、Storage、Artifact Registryの
    resourceを作成しません。
11. 既存のCLI、provider、completion、manual-component testが成功します。
12. fresh repositoryとcollision fixtureがdownstreamインストールを検証します。

## 承認

- **状態:** `Pending`
- **承認者:** `None`

## 実装状況

実装はPR #193に含まれており、contractおよびgovernance検証を進めています。
この報告書の作成時点で明示的な承認記録を確認できなかったため、承認は
pendingのまま維持します。

## 検証

- Project Sync Node test 13件成功
- `GOCACHE=/tmp/soku-go-cache go test ./...` 成功
- 全ローカルNode regression 105件成功、失敗0件
- repository policyおよびgovernance regression 102件成功、失敗0件
- NPM wrapper test 7件成功（隔離cacheを使用）
- `soku/scripts/package_test.sh` の5 archiveとchecksumの再現性を確認
- `go mod verify` の全module検証成功
- `git diff --check` 成功

Hosted Cloud Buildの結果は外部証拠としてのみ報告し、cloud設定は変更
しません。

## 公開適合性レビュー

- [x] credentials、tokens、private keys、credential-bearing URLsがない
- [x] 非公開repository、project、product名がない
- [x] cloud project ID、account number、service URL、image URI、revision identifierがない
- [x] 個人のbilling、subscription、budget、payment status情報がない
- [x] 個人のemail、電話番号、住所、local absolute pathがない
- [x] 非公開Issue、PR、Project、control-plane identifierがない

## AI支援

- **計画/実装/草案作成:** OpenAI Codex
