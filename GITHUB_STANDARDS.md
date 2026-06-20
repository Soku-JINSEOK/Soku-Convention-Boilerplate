# GitHub Standards

## Purpose

This document defines repository collaboration standards for GitHub-based workflows.

It exists to ensure that issues, pull requests, reviews, and automation remain consistent across projects built on this boilerplate.

## Principles

GitHub workflows should optimize for:

- clear communication
- low review friction
- explicit decision history
- predictable collaboration patterns

## Issue Standards

Issues should capture a problem, request, or decision point clearly enough that another contributor can act on them without extra guesswork.

Every issue should make the following as clear as possible:

- background
- goal
- scope
- constraints
- definition of done

### Recommended Issue Types

- bug report
- feature request
- refactor proposal
- documentation update
- chore or maintenance task

## Pull Request Standards

Pull requests should be structured to help reviewers understand the change quickly.

Every PR should answer:

- What changed?
- Why was it needed?
- How was it validated?
- What risks or tradeoffs exist?
- What follow-up work remains?

## Review Standards

Reviews should focus on:

- behavioral correctness
- architectural fit
- maintainability
- test quality
- clarity of intent

Review noise should be minimized.  
Formatting concerns should be delegated to tooling whenever possible.

## Branching Guidance

Repositories may choose their own branching strategy, but the strategy should be documented and consistently applied.

At minimum, teams should define:

- default branch policy
- feature branch naming pattern
- hotfix handling
- release tagging approach

## Release And Sync

Releases should be explicit and repeatable.

- Use semantic-style tags in the form `vMAJOR.MINOR.PATCH`.
- Pin downstream repositories to a specific boilerplate tag.
- Record the imported tag in the downstream README or setup notes.
- Use `scripts/sync-boilerplate.ps1` to copy convention-owned files into a downstream repository.
- Keep downstream application code separate from boilerplate updates.

See [RELEASE_AND_SYNC.md](./RELEASE_AND_SYNC.md) for the full operating contract.

## Labels and Metadata

GitHub labels should help organize work rather than create clutter.

Prefer a small, stable label system covering:

- type
- priority
- status
- area or domain

## Templates

Repositories should provide templates where they reduce ambiguity.

Recommended templates:

- issue templates
- pull request template
- bug report template
- feature request template

## Automation Expectations

GitHub should be used as an operational surface, not just a code host.  
That means repository automation should support:

- CI validation
- consistent review flow
- visibility into project health
- structured collaboration

## Summary

Well-managed GitHub workflows reduce coordination cost.  
The goal is to make repository collaboration explicit, reviewable, and repeatable across teams and projects.
