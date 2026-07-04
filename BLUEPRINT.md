# 🗺️ Blueprint

## 🎯 Purpose

`Soku-Convention-Boilerplate` is a design-first repository blueprint for creating projects that stay readable, consistent, and maintainable across time, teams, and stacks.

This document is the canonical source of truth for the repository architecture.  
All other top-level documents either define narrower policies, provide operational templates, or serve as reference material.

## 🎨 Design Goals

The boilerplate is optimized for:

- readability for humans
- predictability for contributors
- low ambiguity for AI agents
- automation-friendly conventions
- portability across multiple repositories

## 📖 Reading Order

When approaching a repository built on this boilerplate, use this order:

1. [README.md](./README.md) for the public overview
2. [BLUEPRINT.md](./BLUEPRINT.md) for the canonical architecture
3. [AGENTS.md](./AGENTS.md) for AI operating behavior
4. policy documents for the relevant domain
5. stack examples and project-specific docs for implementation details

## ⚖️ Authority Model

Not all documents carry the same weight.

### 📏 Normative

These documents define expected behavior:

- [BLUEPRINT.md](./BLUEPRINT.md)
- [AGENTS.md](./AGENTS.md)
- [CONTRIBUTING.md](./CONTRIBUTING.md)
- [CODE_STYLE.md](./docs/standards/CODE_STYLE.md)
- [PROJECT_STRUCTURE.md](./docs/standards/PROJECT_STRUCTURE.md)
- [GITHUB_STANDARDS.md](./docs/standards/GITHUB_STANDARDS.md)
- [CICD_STANDARDS.md](./docs/standards/CICD_STANDARDS.md)
- [RELEASE_AND_SYNC.md](./docs/standards/RELEASE_AND_SYNC.md)
- [LICENSE_POLICY.md](./docs/policy/LICENSE_POLICY.md)
- [SECURITY_POLICY.md](./docs/policy/SECURITY_POLICY.md)
- [CLOUD_POLICY.md](./docs/policy/CLOUD_POLICY.md)

### 🔌 Interface

These documents are the public or operational entrypoints:

- [README.md](./README.md)
- [.github/PULL_REQUEST_TEMPLATE.md](./.github/PULL_REQUEST_TEMPLATE.md)
- [.github/ISSUE_TEMPLATE/*](./.github/ISSUE_TEMPLATE/)
- [.github/workflows/ci.yml](./.github/workflows/ci.yml)
- [.github/CODEOWNERS](./.github/CODEOWNERS)
- [LICENSE](./LICENSE)
- [SECURITY.md](./SECURITY.md)

### 📚 Reference

These documents explain examples and implementation patterns:

- [STACK_EXAMPLES.md](./docs/guides/STACK_EXAMPLES.md)
- [STACK_CONFIGS.md](./docs/guides/STACK_CONFIGS.md)
- [README_GUIDE.md](./docs/guides/README_GUIDE.md)
- [INIT_GUIDE.md](./docs/guides/INIT_GUIDE.md)
- [APPLICABILITY.md](./docs/guides/APPLICABILITY.md)

If a document conflicts with this blueprint, the blueprint wins unless a downstream project explicitly overrides it in a documented way.

## Language Policy

The repository uses a layered language strategy.

- Human-facing overview content defaults to Korean and English.
- Japanese may be added when the audience or collaboration context makes it useful.
- Operational rules, governance, policy, and AI instructions are written in English only.

This keeps the public-facing docs approachable while making the operating rules easy for AI agents and humans to parse consistently.

### 🧱 Multi-Language Block Ordering

When a document contains more than one language, group each language's content into a single contiguous block instead of interleaving languages section by section or paragraph by paragraph. Order the blocks English first, followed by each additional language in the order it was added (for example: English, then Korean, then Japanese). A reader should encounter at most one language switch per additional language in the document, not once per section.

`README.md` is the reference implementation of this rule — see its `## English`, `## 한국어`, and `## 日本語` blocks.

## 🏗️ Repository Shape

The boilerplate assumes a structure that is easy to navigate without hidden conventions.

### 📂 Expected Top-Level Files

- [.gitignore](./.gitignore)
- [.editorconfig](./.editorconfig)
- [.gitmessage](./.gitmessage)
- [README.md](./README.md)
- [BLUEPRINT.md](./BLUEPRINT.md)
- [CONTRIBUTING.md](./CONTRIBUTING.md)
- [AGENTS.md](./AGENTS.md)
- [LICENSE](./LICENSE)
- [SECURITY.md](./SECURITY.md)

Deeper reference material lives under `docs/`, grouped by category:

- `docs/standards/`: [CODE_STYLE.md](./docs/standards/CODE_STYLE.md), [PROJECT_STRUCTURE.md](./docs/standards/PROJECT_STRUCTURE.md), [GITHUB_STANDARDS.md](./docs/standards/GITHUB_STANDARDS.md), [CICD_STANDARDS.md](./docs/standards/CICD_STANDARDS.md), [RELEASE_AND_SYNC.md](./docs/standards/RELEASE_AND_SYNC.md)
- `docs/policy/`: [LICENSE_POLICY.md](./docs/policy/LICENSE_POLICY.md), [SECURITY_POLICY.md](./docs/policy/SECURITY_POLICY.md), [CLOUD_POLICY.md](./docs/policy/CLOUD_POLICY.md)
- `docs/guides/`: [STACK_EXAMPLES.md](./docs/guides/STACK_EXAMPLES.md), [STACK_CONFIGS.md](./docs/guides/STACK_CONFIGS.md), [README_GUIDE.md](./docs/guides/README_GUIDE.md), [INIT_GUIDE.md](./docs/guides/INIT_GUIDE.md), [APPLICABILITY.md](./docs/guides/APPLICABILITY.md)

### 🗃️ Expected Repository Areas

- `src/` for application code
- `tests/` or `test/` for verification
- `docs/` for deeper reference material
- `scripts/` for local and operational automation
- `config/` for shared configuration
- `templates/` for copyable starter configuration sets
- `infra/` for deployment and environment assets
- `.github/` for collaboration and automation files

Projects may adjust the exact shape to match their ecosystem, but they should keep the same intent and readability.

## 🤝 Collaboration Model

The GitHub workflow is designed to make work visible and reviewable.

### 🐞 Issues

Issues should clearly define:

- background
- goal
- scope
- constraints
- definition of done

### 🔀 Pull Requests

Pull requests should answer:

- what changed
- why it changed
- how it was validated
- what tradeoffs exist
- what remains open

### 🔍 Reviews

Reviews should focus on:

- correctness
- maintainability
- design clarity
- testing impact

Formatting concerns should be resolved by tooling whenever possible.

## 🔁 CI/CD Model

The baseline CI in this repository is intentionally minimal and documentation-aware.

### 1️⃣ Repository Hygiene

This layer validates the boilerplate itself:

- required documentation exists
- Markdown is formatted consistently
- workflow YAML is valid

### 2️⃣ Stack Validation

Downstream projects should add runtime-specific checks such as:

- linting
- type checking
- unit tests
- build verification
- migration safety checks

### 3️⃣ Delivery Validation

When a repository ships artifacts or deploys services, the pipeline should verify:

- packaging
- deployment readiness
- environment assumptions
- health checks

### 4️⃣ Production Delivery

Production deployment should be explicit, gated, and reversible where possible.

## 🔐 Security Model

Security is a baseline operating concern, not a later-stage add-on.

The repository assumes:

- secrets are never committed
- credentials are environment-scoped
- access follows least privilege
- dependency risk is monitored
- logs avoid sensitive values

Security reporting should be clearly documented in `SECURITY.md`, while [`docs/policy/SECURITY_POLICY.md`](./docs/policy/SECURITY_POLICY.md) explains the operating baseline in more detail.

## 📜 License Model

The repository should always declare a license clearly.

For this boilerplate, the default starting point is `MIT` because it is easy to understand and easy to reuse.  
If a downstream project needs stronger patent language or different distribution constraints, it should replace the default deliberately and document the reason.

## ☁️ Cloud Policy

Cloud choice should be driven by workload fit, operating model, and organizational reality.

### 🟦 GCP

Use `GCP` when the project benefits from:

- data and analytics services
- Cloud Run or GKE-centric delivery
- a relatively clean managed-service path
- AI and data platform integration

### 🟧 AWS

Use `AWS` when the project needs:

- the broadest service catalog
- complex infrastructure flexibility
- mature enterprise operating patterns
- multi-account governance at scale

### 🟦 Azure

Use `Azure` when the project is tightly aligned with:

- Microsoft enterprise tooling
- hybrid environments
- Entra ID and corporate identity standards
- Windows or .NET-centered organizational ecosystems

The default rule is to choose the provider that best matches the team's actual operating constraints, not the one that looks best on paper.

## 🧱 Stack Coverage

The boilerplate is intentionally stack-neutral at the top level.

Reference examples currently cover:

- JavaScript
- TypeScript
- Node.js
- Python
- Go
- Java
- Spring
- MySQL
- PostgreSQL
- gcloud

The blueprint does not prescribe a single application architecture for every stack.  
Instead, it ensures that whichever stack is adopted stays readable and conventionally organized.

## 🤖 AI Agent Operating Model

AI agents should treat this repository as a structured operating environment.

Recommended agent sequence:

1. read `BLUEPRINT.md`
2. read `AGENTS.md`
3. inspect the relevant policy docs
4. inspect stack examples or templates
5. make the smallest change that preserves the repository contract

When a rule is unclear, agents should prefer the document that is:

- more specific
- more operational
- closer to the affected surface

## Maturity Levels

The boilerplate supports three practical maturity levels.

### 🌱 Bootstrap

The repository contains the documentation skeleton, GitHub templates, and baseline CI.

### 🌿 Standard

The repository adds stack-specific linting, tests, and deployment rules.

### 🌳 Scaled

The repository adds ownership, release discipline, stronger security controls, and environment-specific delivery pipelines.

## 🔄 Release And Sync Model

The boilerplate is distributed as a versioned convention package.

- The source of truth is this repository.
- Releases use semantic-style tags in the form `vMAJOR.MINOR.PATCH`.
- Downstream repositories should pin to a release tag before importing updates.
- Convention-owned files are synchronized with `scripts/sync-boilerplate.sh` (Linux/macOS) or `scripts/sync-boilerplate.ps1` (Windows).
- [`docs/standards/RELEASE_AND_SYNC.md`](./docs/standards/RELEASE_AND_SYNC.md) defines the operational release and sync rules in detail.

## 🚫 Non-Goals

This blueprint does not attempt to define:

- one universal application architecture
- one universal database schema
- one universal deployment topology
- one universal language stack

Those choices belong in project-specific design documents.

## 📐 Change Rule

Any new convention added to this boilerplate must satisfy one question:

Can this rule still make sense when copied into a different repository?

If the answer is no, the rule belongs in a downstream project document instead of the shared boilerplate.

## 🎬 Summary

This repository exists to make future projects easier to read, easier to review, and easier to operate.  
The blueprint is the architectural anchor that keeps the rest of the documentation aligned.
