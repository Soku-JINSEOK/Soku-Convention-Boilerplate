# 🤝 Contributing

## 👋 Overview

This repository is designed to keep development standards consistent across projects.  
Contributing is not only about adding code. It is also about preserving readability, predictability, and maintainability for future contributors.

## ✅ Core Expectations

All contributions should align with the following expectations:

- prefer clarity over clever shortcuts
- follow repository conventions before personal style
- keep changes scoped and intentional
- document important decisions clearly
- rely on automation for formatting and linting whenever possible

## 🔄 Contribution Workflow

### 1️⃣ Understand the Existing Structure

Before making changes, review the current repository structure, naming patterns, and documentation.  
New code should extend the existing system instead of introducing a parallel style.

### 2️⃣ Keep Changes Small and Focused

A single change should solve a single problem whenever possible.  
Avoid mixing refactors, feature work, formatting-only changes, and unrelated cleanup in one contribution unless there is a clear operational reason.

### 3️⃣ Follow the Style Baseline

This repository uses `Google Style Guide` as the default convention baseline.  
Language-specific tooling may differ by stack, but the principles of readability, consistency, and explicitness should remain the same.

### 4️⃣ Let Tooling Enforce Style

Run the formatter, linter, and test suite relevant to the project before requesting review.  
If a rule can be enforced automatically, prefer automation over manual policing in review comments.

### 5️⃣ Write Reviewable Changes

Contributions should be easy to review. That means:

- clear naming
- intentional file organization
- minimal noise
- meaningful commit boundaries
- concise documentation when context is not obvious

### 6️⃣ Explain Why

Comments, pull request descriptions, and documentation should explain intent and tradeoffs, not restate obvious syntax.

## 📝 Commit Message Standard

This repository uses **Conventional Commits** combined with **gitmoji**. This is a structural convention (type, scope, gitmoji), independent of what human language the description text is written in — see the Collaboration Language section in [`docs/standards/GITHUB_STANDARDS.md`](./docs/standards/GITHUB_STANDARDS.md) for that decision.

Reference: [conventionalcommits.org](https://www.conventionalcommits.org/en/v1.0.0/) · [gitmoji.dev](https://gitmoji.dev)

### 📐 Format

```text
<gitmoji> <type>(<scope>): <English subject>

[optional body — explain WHY, not WHAT]

[optional footer(s): Closes #123, BREAKING CHANGE: ...]
```

### 🎨 Type + Gitmoji Map

| Gitmoji | Type | When to use |
| ------- | ---- | ----------- |
| ✨ | `feat` | New feature |
| 🐛 | `fix` | Bug fix |
| ♻️ | `refactor` | Code restructure without behavior change |
| 🎨 | `style` | Formatting, whitespace, UI styling (no logic change) |
| 📚 | `docs` | Documentation only |
| ✅ | `test` | Add or update tests |
| 🔧 | `chore` | Config, tooling, maintenance |
| 🚀 | `perf` | Performance improvement |
| 📦 | `build` | Build system, compile settings, package managers |
| 👷 | `ci` | CI/CD pipeline changes |
| 🔥 | `remove` | Remove files or features |
| 🚑 | `hotfix` | Critical production bug fix |
| 🔖 | `release` | Release tagging or version update |
| 🔄 | `sync` | Sync changes from boilerplate or upstream |
| 🔒️ | `security` | Security fix or hardening |
| ⏪️ | `revert` | Revert a previous commit |
| 💥 | `feat!` / `fix!` | Breaking change |

### 📋 Rules

- **Scope is required** and must use lowercase kebab-case (e.g., `(sync-script)`, `(ts-template)`).
- **English subject is required** and must use standard ASCII/English text. Imperative mood, no trailing period, ≤ 72 characters.
- Body: explain WHY, not WHAT — wrap at 72 characters
- Breaking change: append `!` after type (`feat!:`) and add footer `BREAKING CHANGE: <description>`

### 🔧 Setup

A `.gitmessage` template is provided at the repository root.  
Activate it with:

```bash
git config commit.template .gitmessage
```

Additionally, automated commit linting is available as a shared template. You can copy the configuration files from [templates/_shared/commitlint/](file:///home/seok_jinseok/CodeBase/Soku-Convention-Boilerplate/templates/_shared/commitlint/) to your project root to run automatic title checks on Git hooks or CI.

## 🔀 Pull Request Standards

Every pull request should make it easy for reviewers to answer these questions:

- What changed?
- Why was the change necessary?
- What assumptions were made?
- How was the change validated?
- Are there follow-up tasks or known limitations?

## 🔍 Review Standards

Code review should prioritize:

- correctness
- maintainability
- architectural coherence
- testing impact
- readability

Avoid spending review energy on formatting issues that should be handled by tools.

## 🌐 Documentation Policy

Document language (multilingual overview content vs. English-only operational content) follows the single canonical rule in the [Language Policy in BLUEPRINT.md](./BLUEPRINT.md#language-policy).

The language used for commit messages, issues, and pull requests is a separate, per-project decision — see the Collaboration Language section in [`docs/standards/GITHUB_STANDARDS.md`](./docs/standards/GITHUB_STANDARDS.md).

## ⚖️ Decision Rule

If a proposed change improves short-term speed but weakens long-term consistency, prefer consistency unless there is a documented reason not to.

## 🎬 Summary

Contributing to this repository means protecting a shared standard.  
The goal is not just to ship code, but to leave behind a codebase that the next human or AI contributor can understand quickly and extend safely.
