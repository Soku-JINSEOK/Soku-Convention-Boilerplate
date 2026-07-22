# 🧩 Soku-Convention-Boilerplate

> A reusable convention baseline for teams and AI agents who care about readable code, stable structure, and long-term maintainability.

[한국어](./README.ko.md) | [日本語](./README.ja.md)

## 👋 Overview

`Soku-Convention-Boilerplate` is a reusable base template for maintaining consistent code style, structure, and collaboration standards across any project.  
It is designed not just as a starter, but as a repeatable foundation for building a development culture centered on readability and long-term maintainability.

## 🗺️ Master Blueprint

For the canonical operating design, start with [BLUEPRINT.md](./BLUEPRINT.md).

## 📦 Current Published Baseline

The current published releases are boilerplate `v1.0.3` and CLI
`soku/v0.1.4`. Their signed records, complete hosted gates, packages, and
checksums passed, but required public four-stack smoke found that Python Ruff
traverses a generated JavaScript `node_modules` tree. Issue #41 therefore
continues with the single-axis boilerplate `v1.0.4` candidate; keep existing
tags immutable and use the new baseline only after its separately approved
signed tag and Release are published. See
[VERIFICATION_GUIDE.md](./VERIFICATION_GUIDE.md) for the complete checks.

---

## 📌 At a Glance

| Area | Standard |
| --- | --- |
| Style baseline | `Google Style Guide` |
| Human-facing docs | English, Korean, and Japanese |
| Rule and governance docs | English only |
| Main objective | Consistency across projects |
| Review priority | Logic, clarity, maintainability |
| Enforcement model | Formatter + linter + documentation |
| Commit style | Gitmoji + Conventional Commits |
| Release tagging | Signed tags (`git tag -s`) for Verified status |

## 🤔 Why Google Style Guide

This project adopts the `Google Style Guide` as its baseline convention.  
The first reason is a strong agreement with the idea that code is read far more often than it is written. We value long-term clarity and maintainability more than short-term implementation speed.

The second reason is automation. Google-style conventions work well with formatters, linters, and static analysis tools, which helps reduce subjective style debates and allows teams to focus more on business logic and architecture.

---

## 💭 Philosophy

This boilerplate is built on the belief that a project should remain understandable, predictable, and maintainable regardless of its size, age, or contributors.

We do not treat conventions as cosmetic preferences. We treat them as operational guardrails that reduce ambiguity, support collaboration, and make both humans and AI agents more effective when working in the same codebase.

Our goal is to create a project foundation that can be reused across different domains without rewriting the team culture each time.

## ✅ Principles

1. Readability comes before cleverness.
2. Consistency is more valuable than personal preference.
3. Automation should enforce style whenever possible.
4. Project structure should be predictable across repositories.
5. Code review should focus on logic, behavior, and design rather than formatting disputes.
6. Documentation should explain intent, not just mechanics.
7. Every convention should help both current contributors and future maintainers.

## ⚙️ Operating Standards

### 1. Style Baseline

All repositories based on this boilerplate should use `Google Style Guide` as the default style baseline unless there is a clearly documented reason to diverge.

### 2. Formatting and Linting

Formatting and linting should be automated and treated as part of the development workflow, not as optional cleanup work.  
If a formatter or linter can enforce a rule reliably, the system should enforce it instead of leaving it to manual review.

### 3. Repository Consistency

Projects should preserve a stable directory structure, naming strategy, and documentation pattern so that contributors can move between repositories with minimal cognitive overhead.

### 4. Documentation Rules

document language follows the [Language Policy in BLUEPRINT.md](./BLUEPRINT.md#language-policy): human-facing overview content defaults to English, Korean, and Japanese, while agent-facing rules, project philosophy, governance, and operational standards stay English-only. When a document mixes languages, each language's content is grouped into one contiguous block (English block, then the next language's block, and so on) rather than interleaved section by section.

### 5. Review Discipline

Pull requests and code reviews should prioritize:

- correctness
- maintainability
- architectural clarity
- testability
- explicit tradeoffs

Style issues should be handled by tooling whenever possible.

### 6. Scalability of Conventions

Any rule added to this boilerplate should be reusable across multiple repositories.  
If a convention only works for one project, it should be treated as a project-specific rule rather than a boilerplate standard.

### 7. AI Agent Compatibility

This repository should be organized so that AI agents can quickly infer:

- project intent
- code ownership boundaries
- structural conventions
- documentation expectations
- execution and validation workflows

To support that goal, global rules should be explicit, stable, and written in direct English.

## 🎯 Intended Use

This boilerplate is intended to serve as:

- a standard starting point for new repositories
- a shared convention layer across personal and team projects
- a training ground for writing clean, readable, maintainable code
- a base environment where automation reduces style friction
- a repository structure that remains understandable to both humans and AI agents

## 📚 Documents

- [README.md](./README.md): multilingual overview and project positioning
- [BLUEPRINT.md](./BLUEPRINT.md): canonical repository design and authority map
- [CONTRIBUTING.md](./CONTRIBUTING.md): contributor workflow and review expectations
- [AGENTS.md](./AGENTS.md): English-only operating guidance for AI agents
- [LICENSE](./LICENSE): default boilerplate license
- [SECURITY.md](./SECURITY.md): security reporting entrypoint
- [`soku` CLI](./soku/README.md): build, install, verification, packaging, and release operation
- [VERIFICATION_GUIDE.md](./VERIFICATION_GUIDE.md): complete local, hosted, governance, artifact, security, and cost checks

### 📏 `docs/standards/` — normative structural and process rules

- [CODE_STYLE.md](./docs/standards/CODE_STYLE.md): style baseline and code-writing rules
- [PROJECT_STRUCTURE.md](./docs/standards/PROJECT_STRUCTURE.md): repository folder organization and structural rules
- [GITHUB_STANDARDS.md](./docs/standards/GITHUB_STANDARDS.md): issue, PR, review, and template governance
- [RELEASE_AND_SYNC.md](./docs/standards/RELEASE_AND_SYNC.md): release tagging and downstream sync model
- [SOKU_LIFECYCLE.md](./docs/standards/SOKU_LIFECYCLE.md): normative CLI, manifest, ownership, provider, and transactional lifecycle contract for `soku`
- [CICD_STANDARDS.md](./docs/standards/CICD_STANDARDS.md): continuous integration and delivery expectations

### 🛡️ `docs/policy/` — declared policy positions

- [LICENSE_POLICY.md](./docs/policy/LICENSE_POLICY.md): how repositories should choose and declare licenses
- [SECURITY_POLICY.md](./docs/policy/SECURITY_POLICY.md): baseline security expectations for source, secrets, and delivery
- [CLOUD_POLICY.md](./docs/policy/CLOUD_POLICY.md): practical cloud-provider decision rules for GCP, AWS, and Azure

### 🧭 `docs/guides/` — reference material and walkthroughs

- [STACK_EXAMPLES.md](./docs/guides/STACK_EXAMPLES.md): practical examples across common languages, frameworks, databases, and cloud workflows
- [STACK_CONFIGS.md](./docs/guides/STACK_CONFIGS.md): copyable starter configuration sets by stack
- [README_GUIDE.md](./docs/guides/README_GUIDE.md): how repository README files should be written and maintained
- [INIT_GUIDE.md](./docs/guides/INIT_GUIDE.md): stack-detection and setup checklist for AI agents bootstrapping a downstream repository
- [CLOUD_RUN_CICD.md](./docs/guides/CLOUD_RUN_CICD.md): local parity and Cloud Run CD pipeline guide (OIDC/WIF, plan/deploy, rollback, evidence)
- [APPLICABILITY.md](./docs/guides/APPLICABILITY.md): which parts of this boilerplate apply to personal projects vs. teams
- [LANGUAGE_SELECTION.md](./docs/guides/LANGUAGE_SELECTION.md): how to choose a programming language for a new project or feature

### 📝 `docs/issues/` — task report artifacts

- [TASK_REPORT_TEMPLATE.md](./docs/issues/TASK_REPORT_TEMPLATE.md): template for a written, approved plan before implementation starts

## 🧱 Starter Stack Coverage

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

## 🛠️ Configuration Sets

For copyable starter configuration files, see [STACK_CONFIGS.md](./docs/guides/STACK_CONFIGS.md) and the `templates/` directory.

## 🎬 Summary

This boilerplate is more than a project starter. It is a reusable operational baseline for keeping development standards stable across projects.
