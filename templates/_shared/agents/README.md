# 🤝 Domain Agent Adapter Guide

## 🎯 Purpose

The files in this directory (`frontend-agent.md`, `backend-agent.md`, `app-agent.md`, `db-agent.md`, `infra-agent.md`, `docs-agent.md`) describe ownership boundaries for parallel AI agents working on a domain-based repository (see [Multi-Domain Layout in PROJECT_STRUCTURE.md](../../docs/standards/PROJECT_STRUCTURE.md#multi-domain-layout-alternative)).

These documents are written as **plain, tool-agnostic markdown** on purpose. They are not locked to any single AI coding tool's file format. Each one is the source of truth for that domain's agent; how you wire it into a specific tool is a thin adapter step, described below.

## 🔌 Wiring a Domain Charter Into Your Tool

- **Claude Code**: copy the relevant file to `.claude/agents/<name>.md` in the target repo and add YAML frontmatter (`name`, `description`, optionally `tools`) above the existing content. Do not rewrite the body — the frontmatter is the only tool-specific addition.
- **Cursor / Windsurf / other rule-file tools**: paste the file's content into that tool's project rules or context file (e.g. `.cursor/rules/`, `.windsurfrules`), scoped to the domain's directory if the tool supports path-scoped rules.
- **Codex or other prompt-only tools**: include the file's content directly in the system prompt or task instructions when assigning that domain's work to an agent.
- **No supported tool / manual use**: read the charter yourself before starting work on that domain, and follow the same ownership and boundary rules a human contributor.

Whichever tool you use, the file in this directory is what changes when the ownership rule changes. Tool-specific copies are regenerated from it, not edited independently.

## 📐 Shared Rules Across All Domain Agents

- An agent only edits files inside its own domain directory. It does not reach into another domain's folder, even to fix something obviously broken there — it reports the issue instead.
- Any file that multiple domains depend on (API contracts, shared type definitions, DB schemas, shared config) is a **shared contract file**. Shared contract files are agreed and written *sequentially*, before the parallel phase starts. Once the parallel phase begins, agents treat shared contract files as read-only; if a contract needs to change mid-flight, that is escalated rather than edited directly.
- Use exactly one domain charter per repository shape: `app-agent.md` for a single monolithic app, or `frontend-agent.md` + `backend-agent.md` for a split repository — never both sets at once, matching the [Multi-Domain Layout exclusivity rule](../../docs/standards/PROJECT_STRUCTURE.md#multi-domain-layout-alternative).
