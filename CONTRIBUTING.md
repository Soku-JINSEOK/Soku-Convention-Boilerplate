# Contributing

## Overview

This repository is designed to keep development standards consistent across projects.  
Contributing is not only about adding code. It is also about preserving readability, predictability, and maintainability for future contributors.

## Core Expectations

All contributions should align with the following expectations:

- prefer clarity over clever shortcuts
- follow repository conventions before personal style
- keep changes scoped and intentional
- document important decisions clearly
- rely on automation for formatting and linting whenever possible

## Contribution Workflow

### 1. Understand the Existing Structure

Before making changes, review the current repository structure, naming patterns, and documentation.  
New code should extend the existing system instead of introducing a parallel style.

### 2. Keep Changes Small and Focused

A single change should solve a single problem whenever possible.  
Avoid mixing refactors, feature work, formatting-only changes, and unrelated cleanup in one contribution unless there is a clear operational reason.

### 3. Follow the Style Baseline

This repository uses `Google Style Guide` as the default convention baseline.  
Language-specific tooling may differ by stack, but the principles of readability, consistency, and explicitness should remain the same.

### 4. Let Tooling Enforce Style

Run the formatter, linter, and test suite relevant to the project before requesting review.  
If a rule can be enforced automatically, prefer automation over manual policing in review comments.

### 5. Write Reviewable Changes

Contributions should be easy to review. That means:

- clear naming
- intentional file organization
- minimal noise
- meaningful commit boundaries
- concise documentation when context is not obvious

### 6. Explain Why

Comments, pull request descriptions, and documentation should explain intent and tradeoffs, not restate obvious syntax.

## Commit Message Standard

This repository uses **Conventional Commits** combined with **gitmoji**.

Reference: [conventionalcommits.org](https://www.conventionalcommits.org/en/v1.0.0/) · [gitmoji.dev](https://gitmoji.dev)

### Format

```text
<gitmoji> <type>(<scope>): <short description>

[optional body — explain WHY, not WHAT]

[optional footer(s): Closes #123, BREAKING CHANGE: ...]
```

### Type + Gitmoji Map

| Gitmoji | Type | When to use |
| ------- | ---- | ----------- |
| ✨ | `feat` | New feature |
| 🐛 | `fix` | Bug fix |
| 📝 | `docs` | Documentation only |
| 💄 | `style` | Formatting, whitespace (no logic change) |
| ♻️ | `refactor` | Code restructure without behavior change |
| ⚡️ | `perf` | Performance improvement |
| ✅ | `test` | Add or update tests |
| 🔧 | `chore` | Config, tooling, maintenance |
| 👷 | `ci` | CI/CD pipeline changes |
| 🔒️ | `security` | Security fix or hardening |
| ⏪️ | `revert` | Revert a previous commit |
| 💥 | `feat!` / `fix!` | Breaking change |
| 🌐 | `i18n` | Internationalization or translation |
| 🚀 | `deploy` | Deployment or release |

### Rules

- Subject line: imperative mood, no trailing period, ≤ 72 characters
- Body: explain WHY, not WHAT — wrap at 72 characters
- Breaking change: append `!` after type (`feat!:`) and add footer `BREAKING CHANGE: <description>`

### Setup

A `.gitmessage` template is provided at the repository root.  
Activate it with:

```bash
git config commit.template .gitmessage
```

## Pull Request Standards

Every pull request should make it easy for reviewers to answer these questions:

- What changed?
- Why was the change necessary?
- What assumptions were made?
- How was the change validated?
- Are there follow-up tasks or known limitations?

## Review Standards

Code review should prioritize:

- correctness
- maintainability
- architectural coherence
- testing impact
- readability

Avoid spending review energy on formatting issues that should be handled by tools.

## Documentation Policy

User-facing or onboarding-oriented content should default to Korean and English.  
Japanese may be added when needed for collaboration or audience fit.  
Project rules, governance, conventions, and AI-facing operational guidance should be written in English.

## Decision Rule

If a proposed change improves short-term speed but weakens long-term consistency, prefer consistency unless there is a documented reason not to.

## Summary

Contributing to this repository means protecting a shared standard.  
The goal is not just to ship code, but to leave behind a codebase that the next human or AI contributor can understand quickly and extend safely.
