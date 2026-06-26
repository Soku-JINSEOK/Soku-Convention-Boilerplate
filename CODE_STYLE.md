# Code Style

## Purpose

This document defines the operational style expectations for repositories built on `Soku-Convention-Boilerplate`.

The goal is not to create stylistic rigidity for its own sake.  
The goal is to reduce ambiguity, improve readability, and make automated enforcement practical across different projects and stacks.

## Baseline

The default baseline is `Google Style Guide`.

When a language-specific formatter or linter is adopted, it should either:

- align directly with Google-style conventions, or
- document any intentional differences explicitly

## Style Priorities

When making style decisions, use this order of priority:

1. readability
2. consistency
3. explicitness
4. maintainability
5. brevity

Shorter code is not better if it becomes harder to read or review.

## Naming

Names should be descriptive enough to explain intent without requiring extra interpretation.

Prefer:

- clear domain language
- predictable file and directory names
- stable naming patterns across similar modules

Avoid:

- unnecessary abbreviations
- vague utility-style names
- inconsistent naming for equivalent concepts

## File and Module Design

Files should have a clear responsibility.  
If a file starts serving multiple unrelated concerns, split it before the structure becomes hard to reason about.

Modules should be organized to help contributors quickly infer:

- purpose
- boundaries
- dependencies
- expected extension points

## Functions and Methods

Functions should do one coherent job and expose clear inputs and outputs.

Prefer:

- explicit parameters
- predictable return behavior
- guard clauses when they reduce nesting
- extracted helpers when they improve comprehension

Avoid:

- hidden side effects
- mixed levels of abstraction in the same function
- overly compact logic that obscures intent

## Comments

Comments should explain:

- intent
- constraints
- tradeoffs
- non-obvious behavior

Comments should not narrate obvious syntax or restate the code line by line.

## Formatting

Formatting must be delegated to tooling whenever possible.  
Do not use manual formatting preferences as a source of code review friction.

Typical enforcement areas include:

- indentation
- line length
- import ordering
- spacing
- quote rules
- trailing commas where applicable

## Linting

Lint rules should reinforce correctness and maintainability, not just aesthetics.  
If a lint rule creates repeated low-value noise, the rule should be reevaluated rather than ignored silently.

## Documentation in Code

Public APIs, shared abstractions, and non-obvious modules should include enough context for future contributors to understand:

- what the code is for
- what assumptions it depends on
- how it should be extended safely

## Testing Relationship

Code style and testing are related.  
Readable code is easier to test, and well-structured tests reinforce readable design.

Prefer tests that are:

- easy to scan
- behavior-oriented
- explicit about setup and expectations

## Exception Policy

Not every repository needs identical tooling, but every deviation from the baseline should be intentional and documented.

If a project diverges from the default style, document:

- what changed
- why it changed
- where the new rule applies

## Summary

Code style in this boilerplate exists to support reliable collaboration.  
The standard should help humans read faster, help reviewers focus on substance, and help AI agents operate with less ambiguity.
