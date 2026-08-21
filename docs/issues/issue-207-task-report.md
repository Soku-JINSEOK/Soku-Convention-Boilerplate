# Issue #207 Task Report — Reposition README to clarify Soku lifecycle and sync translations

## Goal and Background

Issue [#207](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/207) addresses the first-screen presentation of `Soku-Convention-Boilerplate`. The previous top-of-README framed the project primarily as a static starter template rather than declarative convention baseline and lifecycle tooling powered by the `soku` CLI. Furthermore, the Quick Start was collapsed inside an accordion, and terminology required formal alignment with the normative `.soku/manifest.json` contract.

This task restructures the top 30% of `README.md` to establish a 30-second product comprehension hook, exposes an immediate and visible 60-second Quick Start, and synchronizes technical parity across English (`README.md`), Korean (`README.ko.md`), and Japanese (`README.ja.md`).

## Proposed Approach

1. **Product Identity & Problem Statement**: Clarify the core value proposition (declarative multi-stack conventions, managed ownership boundaries, and reproducible CLI workflows) to resolve template drift without manual diffing risks.
2. **Visible Quick Start**: Expose full, verified commands for preview (`soku init ... --dry-run`), apply (`soku init ... --yes`), and standard verification (`npm run ...` / `./scripts/verify.sh`).
3. **Contract Alignment**: Strictly adhere to `.soku/manifest.json` as the portable lifecycle record and use canonical ownership terminology (`managed file`, `core-managed`, `project-owned`).
4. **Lifecycle Architecture Flow**: Add a clean, theme-neutral Mermaid `flowchart TB` illustrating the flow from Boilerplate Source (`v1.0.5`) through `soku` CLI (`soku/v0.2.1`) to the Target Downstream Repository.
5. **Multilingual Parity**: Ensure 100% technical, command, version, and terminology parity across `README.md`, `README.ko.md`, and `README.ja.md` while preserving all lower-section governance documents.

## Planned Implementation

- Update `README.md` top section with the revised product definition, visible Quick Start, and Mermaid flowchart.
- Synchronize `README.ko.md` with identical technical meaning, command bytes, versions, and manifest path.
- Synchronize `README.ja.md` with identical technical meaning, command bytes, versions, and manifest path.
- Validate formatting using `markdownlint-cli2@0.22.1` with `.markdownlint.jsonc`.
- Validate PR title against `scripts/contribution-title.mjs`.

## Acceptance Criteria

- [x] Top 30% clearly conveys `soku` CLI lifecycle tooling within 30 seconds.
- [x] Quick Start preview and apply commands provide identical immutable input arguments.
- [x] Portable lifecycle record is consistently referenced as `.soku/manifest.json`.
- [x] Exact command bytes and version pairs (`v1.0.5` / `soku/v0.2.1`) are identical across EN, KO, and JA.
- [x] All lower sections (`At a Glance`, `Philosophy`, `Operating Standards`, `Documents`) are preserved.
- [x] Markdown linting passes with 0 errors.

## Approval

- **Status:** `Approved`
- **Approved by:** `Soku-JINSEOK`

## Implementation Status

- [x] Phase P1-A: English `README.md` top-30% restructuring completed.
- [x] Phase P1-B: `README.ko.md` and `README.ja.md` technical parity synchronization completed.
- [x] Markdown lint normalization via `markdownlint-cli2` completed.
- [x] Documentation tracking issue [#207](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/207) and task report created.

## Verification

- `npx --yes markdownlint-cli2@0.22.1 --config .markdownlint.jsonc "README*.md"` (0 errors)
- `node -e 'import("./scripts/contribution-title.mjs").then(m => console.log(m.validateContributionTitle("📚 docs(readme): clarify Soku lifecycle and sync translations")))'` (Valid)
- `git diff --check` (0 whitespace errors)
- Parity audit between `README.md`, `README.ko.md`, and `README.ja.md` (100% matched commands, versions, and manifest paths)

## Public Disclosure Review

- [x] No credentials, tokens, private keys, or credential-bearing URLs
- [x] No private repository, project, or product names
- [x] No cloud project IDs, account numbers, service URLs, image URIs, or revision identifiers
- [x] No personal billing, subscription, budget, or payment-status information
- [x] No personal email, phone, address, or local absolute path
- [x] No private Issue, PR, Project, or control-plane identifiers

## AI Assistance

- **Planning/implementation/drafting:** `Antigravity`

---

## 목표 및 배경

이슈 [#207](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/207)은 `Soku-Convention-Boilerplate`의 첫 화면 정보 전달력을 개선하기 위한 작업입니다. 기존 README 상단은 정적 템플릿의 성격이 강조되어 `soku` CLI 기반의 선언적 컨벤션 베이스라인 및 수명주기 툴체인이라는 제품 정체성이 충분히 드러나지 않았으며, 빠른 시작이 아코디언 내부에 숨겨져 있었습니다.

본 작업은 `README.md` 상단 30% 영역을 재구성하여 30초 내 제품 정체성을 전달하고, 60초 빠른 시작 명령어를 즉시 노출하며, 영어(`README.md`), 한국어(`README.ko.md`), 일본어(`README.ja.md`) 3개 문서 간 기술적 동등성(Parity)을 확보합니다.

## 제안하는 접근

1. **제품 정체성 및 문제의식 명시**: 템플릿 복사 후 발생하는 템플릿 드리프트(Drift) 문제를 해결하는 선언적 프로필, 소유권 경계, 재현 가능한 CLI 워크플로를 명시합니다.
2. **빠른 시작 전면 노출**: Preview(`soku init ... --dry-run`)와 확정 적용(`soku init ... --yes`), 표준 검증 명령어를 실행 가능한 형태로 배치합니다.
3. **규범 계약 일치**: 휴대용 수명주기 기록 파일로 `.soku/manifest.json`을 명시하고 정규 소유권 용어(`managed file`, `core-managed`, `project-owned`)를 사용합니다.
4. **수명주기 아키텍처 흐름도**: 테마 중립적 Mermaid `flowchart TB`를 적용하여 릴리스 소스(`v1.0.5`)부터 CLI(`soku/v0.2.1`), 대상 저장소까지의 제어 흐름을 시각화합니다.
5. **다국어 기술 동등성**: 하단 기존 거버넌스 문서를 온전히 보존하면서 3개 언어 간 명령어 바이트, 버전, 용어를 100% 일치시킵니다.

## 계획된 구현

- `README.md` 상단 30% 영역 재구성 및 Mermaid 다이어그램 추가.
- `README.ko.md` 한국어 기술 동등성 동기화.
- `README.ja.md` 일본어 기술 동등성 동기화.
- `markdownlint-cli2` 기반 마크다운 포맷팅 정규화.
- `contribution-title.mjs` 기반 PR 제목 및 커밋 규칙 검증.

## 수용 기준

- [x] 상단 30%에서 30초 내에 `soku` CLI 수명주기 도구임을 직관적으로 이해할 수 있음.
- [x] 빠른 시작의 dry-run 및 apply 명령어가 동일한 불변 인자 구조를 가짐.
- [x] 수명주기 기록 파일이 `.soku/manifest.json`으로 일관되게 표기됨.
- [x] EN, KO, JA 3개 문서 간 명령어 바이트와 버전(`v1.0.5` / `soku/v0.2.1`)이 100% 일치함.
- [x] `At a Glance` 이하 하단 거버넌스 및 철학 문서가 원형 보존됨.
- [x] 마크다운 린트 검사가 0 에러로 통과함.

## 승인

- **상태:** `Approved`
- **승인자:** `Soku-JINSEOK`

## 구현 현황

- [x] Phase P1-A: 영문 `README.md` 상단 개편 완료.
- [x] Phase P1-B: `README.ko.md` 및 `README.ja.md` 다국어 기술 동등성 동기화 완료.
- [x] `markdownlint-cli2` 포맷팅 정규화 완료.
- [x] 문서화 추적 이슈 [#207](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/207) 및 태스크 리포트 작성 완료.

## 검증

- `npx --yes markdownlint-cli2@0.22.1 --config .markdownlint.jsonc "README*.md"` (0 errors)
- `node -e 'import("./scripts/contribution-title.mjs").then(m => console.log(m.validateContributionTitle("📚 docs(readme): clarify Soku lifecycle and sync translations")))'` (Valid)
- `git diff --check` (0 whitespace errors)
- `README.md`, `README.ko.md`, `README.ja.md` 간 명령어, 버전, 매니페스트 경로 100% 일치 검증

## 공개 적합성 검토

- [x] credential, token, private key, credential이 포함된 URL이 없음
- [x] 비공개 저장소·프로젝트·제품 이름이 없음
- [x] cloud project ID, 계정 번호, service URL, image URI, revision 식별자가 없음
- [x] 개인 청구·구독·budget·결제 상태 정보가 없음
- [x] 개인 이메일·전화번호·주소·로컬 절대 경로가 없음
- [x] 비공개 Issue·PR·Project·control-plane 식별자가 없음

## AI 지원

- **계획/구현/초안 작성:** `Antigravity`
