# Issue #101 Task Report — Add npm wrapper package for CLI distribution

## Goal and Background

Issue [#101](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/101)
requests a non-archive installation path for `soku` by publishing a Node.js wrapper
package and wiring release automation so the same release source also updates npm.

## Proposed Approach

- Add a dedicated `soku/npm` package with a small Node.js launcher that downloads
  the matching `soku/vMAJOR.MINOR.PATCH` archive from GitHub Releases and executes
  the native binary.
- Gate wrapper execution on `soku/v0.2.0` and later, and verify release
  integrity with `checksums.txt`.
- Extend CI to test the launcher package unit tests.
- Extend the Release workflow to publish `@soku-jinseok/soku` (provenance-enabled)
  when a CLI tag is pushed.
- Update the release and installation documentation to describe the npm path without
  claiming npm availability prior to `0.2.0`.

## Planned Implementation

- Scaffold `soku/npm` with:
  - package manifest (`package.json`)
  - launch script (`bin/soku.js`)
  - reusable launcher helpers (`lib/launcher.mjs`) and launcher tests
  - package preparation validator (`scripts/prepare-package.mjs`)
  - package README
- Add release note `docs/releases/soku-v0.2.0.md` for the first npm-enabled CLI tag.
- Add Issue 101 task report to repository-hygiene checks.
- Extend `.github/workflows/ci.yml` to run `node --test` for `soku/npm`.
- Extend `.github/workflows/release.yml`:
  - keep GitHub Release publishing logic unchanged for native assets
  - add npm publish job with Node 22 and provenance
  - run package preparation validator before publish
- Update `soku/README.md`, `docs/guides/USAGE_MANUAL.md`, and
  `docs/standards/RELEASE_AND_SYNC.md` with npm install notes.
- Keep existing `soku/v0.1.4` statements intact; do not state any npm release
  before `0.2.0`.

## Acceptance Criteria

- The repository contains a functional `soku/npm` package that resolves to a valid
  native CLI binary for supported platforms (linux/darwin/windows) and validates checksums.
- `soku/npm` unit tests pass via `npm test` in `soku/npm`.
- Release workflow for CLI tags runs preparation validation and includes provenance-capable
  npm publish.
- Documentation reflects:
  - wrapper package purpose and install path,
  - release 0.2.0 note,
  - and no claim of npm releases before `0.2.0`.

## Implementation Status

Repository-local implementation is complete:

- Added and validated the full `soku/npm` wrapper package (runtime launcher,
  asset checksumming, and cache behavior contract).
- Added CLI-tag gated npm publishing to the Release workflow with preparation
  validation.
- Extended CI with `soku/npm` tests and repository-hygiene required file checks.
- Added npm install and release-path documentation updates to runtime docs.

Verification that still requires networked release execution remains pending until
the branch is run in remote CI context.

### Remote verification required

- Run the tagged release workflow for `soku/v0.2.0`.
- Confirm the `publish-npm` job succeeds and `npm view @soku-jinseok/soku version`
  returns `0.2.0`.
- Verify `soku --version` works after install (`npm install -g @soku-jinseok/soku@0.2.0`) and
  does not indicate npm availability for versions before `0.2.0`.
- Record any publish/verification restrictions (network, token, registry policy).

## Verification

- Repository-local verification completed:
  - `node --test soku/npm/lib/launcher.test.mjs`
  - `node --test .github/issue-form-order.test.mjs .github/usage-manual.test.mjs`
  - `cd soku/npm && npm test`
  - `node soku/npm/scripts/prepare-package.mjs --version 0.2.0`

- Runtime `npm publish` and post-publish release verification remain pending until
  remote CI execution.

### 남은 원격 검증 항목

- `soku/v0.2.0` 태그 기준 릴리스 워크플로우 실행.
- `publish-npm` Job 성공 확인.
- `npm view @soku-jinseok/soku version` 값이 `0.2.0`인지 확인.
- `npm install -g @soku-jinseok/soku@0.2.0 && soku --version` 동작 확인 및
  `0.2.0` 이전 버전에 대해 npm 경로가 노출되지 않음 확인.
- 발행/검증 제약(토큰/네트워크/레지스트리 정책)이 있으면 사유 기록.

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex

---

## 목표 및 배경

Issue [#101](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/101)은
`soku` CLI에 대해 GitHub 압축 파일 외에 npm 배포 경로를 추가하고, CLI 태그
기반 자동 배포에서 npm 게시를 함께 수행하라는 요청입니다.

## 제안하는 접근

- `soku/npm`에 Node.js 런처 패키지를 추가합니다. 이 패키지는 `soku/vMAJOR.MINOR.PATCH`
  태그의 GitHub Release 자산을 받아 실행 가능한 바이너리를 가져와 실행합니다.
- `soku/v0.2.0` 이상부터 지원되며, `checksums.txt`로 무결성을 검증합니다.
- CI에 `soku/npm` 단위 테스트 실행을 추가합니다.
- CLI 태그 publish 경로에 npm 배포(job)와 provenance를 추가합니다.
- 문서에 npm 설치 경로를 반영하되, `0.2.0` 이전 npm 릴리즈는 언급하지 않습니다.

## 계획된 구현

- `soku/npm` 하부 구조 추가:
  - 패키지 매니페스트(`package.json`)
  - 실행 엔트리(`bin/soku.js`)
  - 런처 보조 모듈(`lib/launcher.mjs`) 및 테스트
  - 준비 스크립트(`scripts/prepare-package.mjs`)
  - 패키지 `README`
- `docs/releases/soku-v0.2.0.md` 추가
- Issue #101 보고서(`docs/issues/issue-101-task-report.md`)를 repository-hygiene에 반영
- `.github/workflows/ci.yml`에 `soku/npm` 테스트 단계 추가
- `.github/workflows/release.yml`에 CLI 태그 시 npm publish 작업 추가:
  - GitHub Release는 기존 아티팩트 배포 유지
  - Node 22 + provenance 포함 publish 수행
  - publish 전 준비 스크립트 검증
- `soku/README.md`, `docs/guides/USAGE_MANUAL.md`, `docs/standards/RELEASE_AND_SYNC.md`
  npm 설치 안내 갱신
- 기존 `soku/v0.1.4` 기준은 유지하고 `0.2.0` 이전 npm 릴리즈는 주장하지 않음

## 수용 기준

- `soku/npm` 패키지가 동작 가능한 네이티브 바이너리를 지원 플랫폼에서 받아 실행하고
  `checksums.txt` 무결성 검증을 수행함
- `soku/npm` 단위 테스트가 `npm test`로 통과함
- CLI 태그 Release에서 게시 전 검증을 거쳐 프로비넌스 포함 npm publish가 동작
- 문서가 런처 패키지 목적, `soku-v0.2.0` 릴리즈 노트, `0.2.0` 이전 npm 릴리즈 비주장을
  반영함

## 승인

- **상태:** `Approved`
- **승인자:** User

## 구현 현황

로컬 범위에서 전체 구현을 반영했습니다:

- `soku/npm` 런처 패키지(런타임 실행기, 체크섬 검증, 캐시 동작 약속) 반영
- CLI 태그 조건의 npm 게시 경로를 Release 워크플로우에 추가하고 준비 검증 반영
- CI에 `soku/npm` 테스트와 repository-hygiene 필수 파일 목록 반영
- 런타임 문서(`soku/README.md`, `docs/guides/USAGE_MANUAL.md`,
  `docs/standards/RELEASE_AND_SYNC.md`) npm 설치 안내 반영

네트워크 기반 릴리스 실행 검증은 원격 CI 컨텍스트에서 추가 필요합니다.

## 검증

- 로컬 검증은 완료했습니다:
  - `node --test soku/npm/lib/launcher.test.mjs`
  - `node --test .github/issue-form-order.test.mjs .github/usage-manual.test.mjs`
  - `cd soku/npm && npm test`
  - `node soku/npm/scripts/prepare-package.mjs --version 0.2.0`

- 실제 `npm publish` 및 원격 배포 검증은 remote CI 환경에서 추가 확인해야 합니다.

## AI 지원

- **계획/구현/초안 작성:** OpenAI Codex
