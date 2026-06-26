# AGENTS

## Purpose

This document provides stable operating guidance for AI agents working in repositories based on `Soku-Convention-Boilerplate`.

It should be read after `BLUEPRINT.md`, which defines the repository-level architecture and authority order.  
This file then acts as the AI-facing behavioral contract.  
If local project instructions exist, agents should follow the more specific rule as long as it does not conflict with higher-priority system constraints.

When editing release behavior, tag policy, or downstream sync logic, read `RELEASE_AND_SYNC.md` before making changes.

## Repository Intent

This repository prioritizes:

- readability over cleverness
- consistency over personal preference
- maintainability over short-term speed
- automation over subjective formatting debate
- documentation clarity over implicit tribal knowledge

## Default Assumptions

Unless local documentation states otherwise, agents should assume the following:

1. `Google Style Guide` is the baseline convention.
2. Formatting and linting should be enforced by tools where possible.
3. Human-facing overview documents default to Korean and English.
4. Japanese may be added selectively when needed.
5. Rules, governance, philosophy, and operational guidance should remain in English.
6. Changes should preserve predictable structure across repositories.

## Agent Behavior Rules

Agents working in this repository should:

- make changes that are narrow, intentional, and easy to review
- preserve existing structure unless restructuring is necessary
- prefer explicit naming and straightforward logic
- avoid introducing one-off patterns that do not generalize
- update relevant documentation when behavior or structure changes
- optimize for future readability, not just immediate task completion

## Editing Policy

When editing code or documentation:

- follow existing repository patterns first
- keep file responsibilities clear
- avoid mixing unrelated edits in one change
- prefer small diffs with obvious intent
- do not rewrite established conventions without a documented reason

## Documentation Policy

Agents should treat documentation as part of the codebase, not as optional polish.

Use multilingual content for:

- overview sections
- onboarding summaries
- high-level repository introductions

Use English-only content for:

- operational rules
- coding standards
- contribution policy
- architecture constraints
- agent instructions

## Review Heuristics

When evaluating or generating changes, agents should prioritize:

- correctness
- readability
- consistency with repository standards
- maintainability
- testability

Style-only commentary should be minimized when tooling can enforce the rule automatically.

## Decision Framework

When multiple implementations are possible, agents should prefer the option that:

1. is easier to understand on first read
2. matches existing repository patterns
3. creates the least policy ambiguity
4. scales better across multiple repositories

## Anti-Patterns

Agents should avoid introducing:

- unnecessary abstraction
- naming shortcuts that reduce clarity
- formatting-only churn without operational value
- repository-specific conventions disguised as global standards
- undocumented deviations from the baseline style

## Expected Outputs

Good agent work in this repository should produce:

- readable code
- stable structure
- low-noise diffs
- clear rationale
- documentation that remains useful to the next contributor

## Summary

The repository is designed so that both humans and AI agents can work with shared expectations.  
Agents should contribute in ways that make the project easier to understand, easier to review, and easier to extend over time.
