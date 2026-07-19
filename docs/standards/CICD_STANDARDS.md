# 🔁 CI/CD Standards

## 🎯 Purpose

This document defines the baseline expectations for continuous integration and continuous delivery in repositories built on `Soku-Convention-Boilerplate`.

CI/CD should support confidence, consistency, and safe iteration.  
It should not exist only as deployment automation, but as a quality enforcement layer.

## 🥅 Core Goals

CI/CD should help teams:

- catch regressions early
- enforce repository standards automatically
- keep delivery repeatable
- reduce manual release risk
- make validation visible

## 🔨 Continuous Integration Expectations

At minimum, CI should validate:

- formatting
- linting
- tests
- build or compile health

If relevant to the stack, CI may also validate:

- type checks
- security scanning
- dependency health
- migration safety
- package integrity

## 🚚 Continuous Delivery Expectations

CD should be designed so that deployment behavior is:

- predictable
- observable
- auditable
- reversible where possible

Deployment workflows should document:

- target environment
- trigger conditions
- required approvals
- rollback expectations

## 📐 Pipeline Design Principles

Pipelines should be:

- small enough to understand
- explicit in purpose
- separated by responsibility
- stable under repeated execution

Avoid building opaque pipelines that only one person can maintain.

For this boilerplate, one top-level Validation workflow must run on every pull
request and every push to `main`. It calls the complete repository and
runtime-template workflows, enforces the pull-request title and every in-scope
commit title, and aggregates their results into one stable `Validation Gate`.
Required branch protection targets that aggregate gate. Component workflows are
reusable implementation details and must not also trigger independently, which
would duplicate execution.

Cancel older in-progress validation for the same ref. Keep default workflow
permissions at `contents: read`; only the release delivery job may request
`contents: write`, after the shared validation and signed release-record checks
succeed. Pin every external action to a verified full commit SHA and retain a
nearby version comment so maintainers can audit upgrades.

## 🌍 Environment Strategy

Projects should define environment expectations clearly, such as:

- local
- development
- staging
- production

Differences between environments should be intentional and documented.

## 🔑 Secrets and Credentials

Do not hardcode secrets into the repository.  
Use the platform's secret management features and keep credential flow explicit in deployment documentation.

## 🚨 Failure Policy

Pipelines should fail loudly and informatively.  
A failing step should make it clear:

- what failed
- why it likely failed
- what area is affected

The aggregate gate must fail when any required component fails, is cancelled,
or does not run unexpectedly. A deliberate skip may be accepted only when the
event makes the check inapplicable, such as contribution-title validation on a
direct post-merge `main` push or a release preflight.

## ✅ Minimum Recommended CI Stages

1. checkout
2. dependency installation
3. formatter and linter validation
4. unit or integration tests
5. build or packaging validation

## ✅ Minimum Recommended CD Stages

1. artifact preparation
2. deployment approval if required
3. deployment execution
4. health verification
5. rollback or remediation path

## 📝 Documentation Rule

If a repository uses CI/CD, its README or `docs/` folder should explain:

- how validation runs
- what must pass before merge
- how deployments are triggered
- who owns deployment decisions

Document the local equivalents, hosted-only checks, audit evidence, and failure
handling in [VERIFICATION_GUIDE.md](../../VERIFICATION_GUIDE.md). For public
repositories, standard hosted runners are free, but larger runners remain
billable. Audit artifacts, caches, Packages, Git LFS, Codespaces, Marketplace
apps, and external services independently of runner minutes; do not treat a
successful or free compute run as proof that storage or account-wide cost is
zero.

## 🎬 Summary

CI/CD should turn repository standards into repeatable system behavior.  
The best pipeline is one that contributors can trust, understand, and maintain without hidden ceremony.
