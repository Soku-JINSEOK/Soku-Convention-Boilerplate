# 🧩 Soku-Convention-Boilerplate

> A reusable convention baseline for teams and AI agents who care about readable code, stable structure, and long-term maintainability.

This document is organized as one complete block per language: English first, then Korean, then Japanese. Each block stands on its own — read only the block for your language rather than jumping between them section by section.

## English

### 👋 Overview

`Soku-Convention-Boilerplate` is a reusable base template for maintaining consistent code style, structure, and collaboration standards across any project.  
It is designed not just as a starter, but as a repeatable foundation for building a development culture centered on readability and long-term maintainability.

### 🗺️ Master Blueprint

For the canonical operating design, start with [BLUEPRINT.md](./BLUEPRINT.md).

---

### 📌 At a Glance

| Area | Standard |
| --- | --- |
| Style baseline | `Google Style Guide` |
| Human-facing docs | Korean / English by default, Japanese when needed |
| Rule and governance docs | English only |
| Main objective | Consistency across projects |
| Review priority | Logic, clarity, maintainability |
| Enforcement model | Formatter + linter + documentation |

### 🤔 Why Google Style Guide

This project adopts the `Google Style Guide` as its baseline convention.  
The first reason is a strong agreement with the idea that code is read far more often than it is written. We value long-term clarity and maintainability more than short-term implementation speed.

The second reason is automation. Google-style conventions work well with formatters, linters, and static analysis tools, which helps reduce subjective style debates and allows teams to focus more on business logic and architecture.

---

### 💭 Philosophy

This boilerplate is built on the belief that a project should remain understandable, predictable, and maintainable regardless of its size, age, or contributors.

We do not treat conventions as cosmetic preferences. We treat them as operational guardrails that reduce ambiguity, support collaboration, and make both humans and AI agents more effective when working in the same codebase.

Our goal is to create a project foundation that can be reused across different domains without rewriting the team culture each time.

### ✅ Principles

1. Readability comes before cleverness.
2. Consistency is more valuable than personal preference.
3. Automation should enforce style whenever possible.
4. Project structure should be predictable across repositories.
5. Code review should focus on logic, behavior, and design rather than formatting disputes.
6. Documentation should explain intent, not just mechanics.
7. Every convention should help both current contributors and future maintainers.

### ⚙️ Operating Standards

#### 1. Style Baseline

All repositories based on this boilerplate should use `Google Style Guide` as the default style baseline unless there is a clearly documented reason to diverge.

#### 2. Formatting and Linting

Formatting and linting should be automated and treated as part of the development workflow, not as optional cleanup work.  
If a formatter or linter can enforce a rule reliably, the system should enforce it instead of leaving it to manual review.

#### 3. Repository Consistency

Projects should preserve a stable directory structure, naming strategy, and documentation pattern so that contributors can move between repositories with minimal cognitive overhead.

#### 4. Documentation Rules

Document language follows the [Language Policy in BLUEPRINT.md](./BLUEPRINT.md#language-policy): human-facing overview content defaults to Korean and English, with Japanese added as needed, while agent-facing rules, project philosophy, governance, and operational standards stay English-only. When a document mixes languages, each language's content is grouped into one contiguous block (English block, then the next language's block, and so on) rather than interleaved section by section.

#### 5. Review Discipline

Pull requests and code reviews should prioritize:

- correctness
- maintainability
- architectural clarity
- testability
- explicit tradeoffs

Style issues should be handled by tooling whenever possible.

#### 6. Scalability of Conventions

Any rule added to this boilerplate should be reusable across multiple repositories.  
If a convention only works for one project, it should be treated as a project-specific rule rather than a boilerplate standard.

#### 7. AI Agent Compatibility

This repository should be organized so that AI agents can quickly infer:

- project intent
- code ownership boundaries
- structural conventions
- documentation expectations
- execution and validation workflows

To support that goal, global rules should be explicit, stable, and written in direct English.

### 🎯 Intended Use

This boilerplate is intended to serve as:

- a standard starting point for new repositories
- a shared convention layer across personal and team projects
- a training ground for writing clean, readable, maintainable code
- a base environment where automation reduces style friction
- a repository structure that remains understandable to both humans and AI agents

### 📚 Documents

- [README.md](./README.md): multilingual overview and project positioning
- [BLUEPRINT.md](./BLUEPRINT.md): canonical repository design and authority map
- [CONTRIBUTING.md](./CONTRIBUTING.md): contributor workflow and review expectations
- [AGENTS.md](./AGENTS.md): English-only operating guidance for AI agents
- [LICENSE](./LICENSE): default boilerplate license
- [SECURITY.md](./SECURITY.md): security reporting entrypoint

#### 📏 `docs/standards/` — normative structural and process rules

- [CODE_STYLE.md](./docs/standards/CODE_STYLE.md): style baseline and code-writing rules
- [PROJECT_STRUCTURE.md](./docs/standards/PROJECT_STRUCTURE.md): repository folder organization and structural rules
- [GITHUB_STANDARDS.md](./docs/standards/GITHUB_STANDARDS.md): issue, PR, review, and template governance
- [RELEASE_AND_SYNC.md](./docs/standards/RELEASE_AND_SYNC.md): release tagging and downstream sync model
- [CICD_STANDARDS.md](./docs/standards/CICD_STANDARDS.md): continuous integration and delivery expectations

#### 🛡️ `docs/policy/` — declared policy positions

- [LICENSE_POLICY.md](./docs/policy/LICENSE_POLICY.md): how repositories should choose and declare licenses
- [SECURITY_POLICY.md](./docs/policy/SECURITY_POLICY.md): baseline security expectations for source, secrets, and delivery
- [CLOUD_POLICY.md](./docs/policy/CLOUD_POLICY.md): practical cloud-provider decision rules for GCP, AWS, and Azure

#### 🧭 `docs/guides/` — reference material and walkthroughs

- [STACK_EXAMPLES.md](./docs/guides/STACK_EXAMPLES.md): practical examples across common languages, frameworks, databases, and cloud workflows
- [STACK_CONFIGS.md](./docs/guides/STACK_CONFIGS.md): copyable starter configuration sets by stack
- [README_GUIDE.md](./docs/guides/README_GUIDE.md): how repository README files should be written and maintained
- [INIT_GUIDE.md](./docs/guides/INIT_GUIDE.md): stack-detection and setup checklist for AI agents bootstrapping a downstream repository
- [APPLICABILITY.md](./docs/guides/APPLICABILITY.md): which parts of this boilerplate apply to personal projects vs. teams

### 🧱 Starter Stack Coverage

This boilerplate is prepared to grow into a multi-stack standard base.  
Example guidance is included for:

- `JavaScript`
- `TypeScript`
- `Node.js`
- `Python`
- `Go`
- `Java`
- `Spring`
- `MySQL`
- `PostgreSQL`
- `gcloud`
- `AWS`
- `Azure`

### 🛠️ Configuration Sets

For copyable starter configuration files, see [STACK_CONFIGS.md](./docs/guides/STACK_CONFIGS.md) and the `templates/` directory.

### 🎬 Summary

This boilerplate is more than a project starter. It is a reusable operational baseline for keeping development standards stable across projects.

## 🇰🇷 한국어

> 상세 운영 규칙, 정책, 레퍼런스 문서는 영어로만 제공됩니다 — 위 [English](#english) 섹션과 [BLUEPRINT.md](./BLUEPRINT.md)를 참고하세요. 아래는 개요·채택 이유·요약만 한국어로 제공합니다.

### 👋 개요

`Soku-Convention-Boilerplate`는 어떤 프로젝트에서도 일관된 코드 스타일, 구조, 협업 기준을 유지하기 위한 공통 베이스 템플릿입니다.  
이 보일러플레이트는 단순한 시작점이 아니라, 읽기 쉬운 코드와 장기적인 유지보수성을 중심에 두는 개발 문화를 반복 가능하게 만드는 것을 목표로 합니다.

### 🤔 Google Style Guide를 채택한 이유

이 프로젝트는 기본 스타일 가이드로 `Google Style Guide`를 채택합니다.  
첫 번째 이유는, 코드는 작성되는 횟수보다 읽히는 횟수가 훨씬 많다는 철학에 공감하기 때문입니다. 우리는 구현 속도보다, 시간이 지나도 쉽게 이해되고 수정할 수 있는 코드를 더 중요한 가치로 둡니다.

두 번째 이유는 자동화의 용이성입니다. Google 컨벤션은 포맷터, 린터, 정적 분석 도구와 잘 맞물리며, 주관적인 스타일 논쟁을 줄이고 비즈니스 로직과 아키텍처에 더 집중할 수 있는 환경을 만들어 줍니다.

### 🎬 요약

이 보일러플레이트는 단순한 프로젝트 시작 템플릿을 넘어, 어떤 프로젝트에서도 흔들리지 않는 개발 기준을 유지하기 위한 재사용 가능한 운영 기반입니다.

## 🎌 日本語

日本語のコンテンツは、プロジェクトの読者や共同作業者が必要とする場合に追加できます。現時点では概要・要約を含め、日本語の翻訳は用意されていません。詳細は上記の [English](#english) セクションを参照してください。
