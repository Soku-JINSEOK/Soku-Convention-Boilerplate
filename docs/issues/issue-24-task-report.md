# 📝 Issue #24 Task Report

## Goal and Background

Align the repository's Issue and pull request authoring formats with the operational structure demonstrated by `Soku-JINSEOK/ci-cd-control-plane` Issue #16 and pull request #8. This report tracks [Issue #24](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/24).

## Proposed Approach

Use GitHub Issue Forms for the five supported Issue types, but prefill each language textarea with the nested operational hierarchy used by the reference Issue. Align the Markdown pull request template with the reference PR hierarchy while retaining repository-specific metadata, verification commands, Gitmoji conventions, and downstream-safety language. Keep priority and area classification in GitHub labels rather than duplicating it in Issue bodies.

## Planned Implementation

- Replace `.github/ISSUE_TEMPLATE/*.md` with corresponding `.yml` Issue Forms.
- Update `.github/PULL_REQUEST_TEMPLATE.md` with common metadata, the English normative source, Korean and Japanese summaries, verification, security boundaries, risks, Gitmoji, and AI disclosure.
- Update `.github/workflows/ci.yml` to require and lint the Issue Form YAML files.
- Update `docs/standards/GITHUB_STANDARDS.md` to document the shared Issue and PR contract.
- Reformat roadmap Issue #23 without changing its title, roadmap links, labels, assignee, or Project status.

## Acceptance Criteria

- All five Issue Forms are valid YAML and provide the standardized multilingual hierarchy.
- The PR template follows the structure demonstrated by the reference PR.
- Repository-specific labels, checks, and non-destructive boundaries are preserved.
- Issue #23 uses the standardized body structure and retains its GitHub metadata.
- Relevant local checks pass, with unavailable checks reported explicitly.

## Approval

- **Status:** `Approved`
- **Approved by:** `Soku-JINSEOK` (explicit request to create the Issue and publish the PR)

## Implementation Status

Complete. The local templates, CI file list, standards documentation, and Issue #23 body have been updated. Publication is tracked by the pull request linked from Issue #24.

## Verification

- `npx --yes yaml-lint@1.7.0 .github/**/*.yml` — passed.
- `npx --yes markdownlint-cli2@0.22.1 --config .markdownlint.jsonc "**/*.md" "#node_modules"` — passed, 36 files and 0 errors.
- `node --test templates/_shared/commitlint/*.test.mjs` — passed, 1 test.
- Issue and pull request template contract assertions — passed.
- `git diff --check` — passed.
- `scripts/verify-sync-parity.sh` — not run successfully because `pwsh` is unavailable in the current environment.
- GitHub Issue #23 metadata verification — title, seven labels, assignee, and Project #2 `Ready` status preserved.

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex

---

## 목표 및 배경

이 저장소의 Issue 및 PR 작성 형식을 `Soku-JINSEOK/ci-cd-control-plane` Issue #16과 PR #8에서 사용한 실제 운영 구조에 맞춥니다. 이 보고서는 [Issue #24](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/24)를 추적합니다.

## 제안하는 접근

지원하는 Issue 유형 5종은 GitHub Issue Form을 사용하되, 각 언어 textarea에 기준 Issue의 중첩된 운영 계층을 미리 제공합니다. Markdown PR 템플릿은 기준 PR 구조에 맞추면서 이 저장소의 메타데이터, 검증 명령, Gitmoji 규칙과 다운스트림 안전 조건을 유지합니다. 우선순위와 영역은 Issue 본문에 중복하지 않고 GitHub 라벨로 관리합니다.

## 계획된 구현

- `.github/ISSUE_TEMPLATE/*.md`를 대응하는 `.yml` Issue Form으로 교체합니다.
- `.github/PULL_REQUEST_TEMPLATE.md`에 공통 메타데이터, 영문 원문, 한·일 요약, 검증, 보안 경계, 위험, Gitmoji와 AI 사용 내역을 반영합니다.
- `.github/workflows/ci.yml`이 Issue Form YAML 파일을 요구하고 검사하도록 갱신합니다.
- `docs/standards/GITHUB_STANDARDS.md`에 공통 Issue/PR 계약을 문서화합니다.
- Issue #23의 제목, 로드맵 링크, 라벨, 담당자와 Project 상태를 유지하면서 본문 형식을 정렬합니다.

## 수용 기준

- Issue Form 5종이 유효한 YAML이며 표준 다국어 계층을 제공합니다.
- PR 템플릿이 기준 PR의 구조를 따릅니다.
- 저장소별 라벨, 검사와 비파괴 조건을 보존합니다.
- Issue #23이 표준 본문 구조를 사용하고 GitHub 메타데이터를 유지합니다.
- 실행 가능한 로컬 검사가 통과하고 실행 불가능한 검사는 명확히 보고됩니다.

## 승인

- **상태:** `Approved`
- **승인자:** `Soku-JINSEOK` (Issue 작성 및 PR 게시를 명시적으로 요청)

## 구현 현황

완료했습니다. 로컬 템플릿, CI 파일 목록, 표준 문서와 Issue #23 본문을 갱신했으며 게시는 Issue #24에 연결되는 PR에서 추적합니다.

## 검증

- YAML lint — 통과.
- Markdown lint — 36개 파일, 오류 0건으로 통과.
- contribution-title 테스트 — 1개 통과.
- Issue 및 PR 템플릿 계약 검사 — 통과.
- `git diff --check` — 통과.
- sync parity — 현재 환경에 `pwsh`가 없어 성공적으로 실행하지 못했습니다.
- Issue #23의 제목, 라벨 7개, 담당자와 Project #2 `Ready` 상태 보존을 확인했습니다.

## AI 지원

- **계획/구현/초안 작성:** OpenAI Codex
