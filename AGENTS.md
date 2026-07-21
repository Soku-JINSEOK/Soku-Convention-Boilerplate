# 🤖 AGENTS

## 🎯 Purpose

This document provides stable operating guidance for AI agents working in repositories based on `Soku-Convention-Boilerplate`.

It should be read after `BLUEPRINT.md`, which defines the repository-level architecture and authority order.  
This file then acts as the AI-facing behavioral contract.  
If local project instructions exist, agents should follow the more specific rule as long as it does not conflict with higher-priority system constraints.

When editing release behavior, tag policy, or downstream sync logic, read [`docs/standards/RELEASE_AND_SYNC.md`](./docs/standards/RELEASE_AND_SYNC.md) before making changes.

## 🧭 Repository Intent

This repository prioritizes:

- readability over cleverness
- consistency over personal preference
- maintainability over short-term speed
- automation over subjective formatting debate
- documentation clarity over implicit tribal knowledge

## 🧩 Default Assumptions

Unless local documentation states otherwise, agents should assume the following:

1. `Google Style Guide` is the baseline convention.
2. Formatting and linting should be enforced by tools where possible.
3. Document language follows the [Language Policy in BLUEPRINT.md](./BLUEPRINT.md#language-policy).
4. Changes should preserve predictable structure across repositories.

## 📏 Agent Behavior Rules

Agents working in this repository should:

- make changes that are narrow, intentional, and easy to review
- preserve existing structure unless restructuring is necessary
- prefer explicit naming and straightforward logic
- avoid introducing one-off patterns that do not generalize
- update relevant documentation when behavior or structure changes
- optimize for future readability, not just immediate task completion

## ✏️ Editing Policy

When editing code or documentation:

- follow existing repository patterns first
- keep file responsibilities clear
- avoid mixing unrelated edits in one change
- prefer small diffs with obvious intent
- do not rewrite established conventions without a documented reason

## 🌐 Documentation Policy

Agents should treat documentation as part of the codebase, not as optional polish.

Which language to write in (multilingual overview content vs. English-only operational content) is defined once in the [Language Policy in BLUEPRINT.md](./BLUEPRINT.md#language-policy) — do not restate or fork that rule in other documents.

For commit messages, issues, and pull requests specifically (a separate concern from document language), see the Collaboration Language section in [`docs/standards/GITHUB_STANDARDS.md`](./docs/standards/GITHUB_STANDARDS.md).

## 🔍 Review Heuristics

When evaluating or generating changes, agents should prioritize:

- correctness
- readability
- consistency with repository standards
- maintainability
- testability

Style-only commentary should be minimized when tooling can enforce the rule automatically.

## 🧰 Tool Invocation and Batched Verification

When tool calls are independent, prefer running them in parallel. Explore
multiple files with one parallel batch of `rg`, `rg --files`, or equivalent
read-only commands whenever sequencing is not required.

Group related edits and behavior checks into one verification pass. Small
helpers do not each require an isolated test when a feature- or file-level test
covers the behavior more clearly.

Do not inspect `git diff` after every small edit. Usually inspect the complete
file-level change once the related edits are finished. For difficult or highly
coupled changes spanning multiple functions or files, run an earlier targeted
check when it materially improves confidence.

## 🧠 Decision Framework

When multiple implementations are possible, agents should prefer the option that:

1. is easier to understand on first read
2. matches existing repository patterns
3. creates the least policy ambiguity
4. scales better across multiple repositories

## 🚫 Anti-Patterns

Agents should avoid introducing:

- unnecessary abstraction
- naming shortcuts that reduce clarity
- formatting-only churn without operational value
- repository-specific conventions disguised as global standards
- undocumented deviations from the baseline style

## 📦 Expected Outputs

Good agent work in this repository should produce:

- readable code
- stable structure
- low-noise diffs
- clear rationale
- documentation that remains useful to the next contributor

## 🧵 Parallel Agent Ownership

Repositories using the [Multi-Domain Layout](./docs/standards/PROJECT_STRUCTURE.md#multi-domain-layout-alternative) can assign one AI agent per domain folder (`frontend/`, `backend/`, `app/`, `db/`, `infra/`, `docs/`) and run them in parallel. Two rules make this safe:

1. **Directory ownership is the boundary.** An agent only edits files inside its own domain folder. It never reaches into another domain's folder, even to fix something obviously broken there — it reports the issue instead.
2. **Shared contracts are sequential, not parallel.** Any file more than one domain depends on (API contracts, shared types, DB schemas, infra interfaces) is agreed and written first, before the parallel phase starts. Once parallel work begins, agents treat these files as read-only.

Domain agent charters are tool-agnostic markdown documents in [`templates/_shared/agents/`](./templates/_shared/agents/), not files locked to any one AI tool's format. See that directory's `README.md` for how to adapt a charter into Claude Code, Cursor, Codex, or any other tool's own format.

## 🎬 Summary

The repository is designed so that both humans and AI agents can work with shared expectations.  
Agents should contribute in ways that make the project easier to understand, easier to review, and easier to extend over time.
