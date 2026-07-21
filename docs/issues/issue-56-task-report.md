# Issue 56 Task Report

## Goal and Background

[Issue #56](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/56)
requires dependency, secret, license, and release-tag integrity checks to become
executable release gates. The corrective release work for Issue #41 provides a
reviewable boundary for adding these checks without changing lifecycle catalog,
manifest, provider, ownership, delivery, or public-release contracts.

## Proposed Approach

Add one reusable security workflow and make it a required dependency of the
aggregate `Validation Gate`. Run the same workflow directly on its schedule and
manual dispatch path, while pull requests, pushes, and release preflight reach
it through the existing reusable validation workflow. Keep every action pinned
to an immutable commit and every scanner version explicit.

Cover full-history secret scanning, JavaScript and Python dependency audits,
both Go modules, OSV source scanning, and declared license/NOTICE evidence. Add
monthly Dependabot coverage for the dependency ecosystems represented in the
repository. Preserve the narrow historical synthetic-secret allowlist and the
existing signed-tag release verification. GitHub release-tag immutability is
verified as repository state rather than implemented in repository files.

## Planned Implementation

1. Add `.github/workflows/security.yml` with secret, dependency, license, Go
   vulnerability, and OSV jobs.
2. Require the reusable security workflow from `.github/workflows/validation.yml`
   and include its result in `Validation Gate`.
3. Add `.github/dependabot.yml` for GitHub Actions, the `soku` Go module, and the
   JavaScript and Python templates.
4. Correct `.gitleaks.toml` to use the supported allowlist table and restrict the
   synthetic fixture exception to its exact test-line shape.
5. Register the new workflow, Dependabot configuration, and this report in the
   repository-hygiene required-file list.
6. Record local and hosted validation plus the active immutable release-tag
   ruleset before closing the issue.

## Acceptance Criteria

- The aggregate `Validation Gate` requires the reusable security workflow.
- Pull-request, push, scheduled, manual preflight, and tag-validation paths run
  the applicable shared security checks without enabling delivery.
- Full-history Gitleaks, npm audit, pip-audit, govulncheck for both Go modules,
  OSV, and license/NOTICE checks pass with recorded evidence.
- Dependabot covers every supported dependency ecosystem in repository scope.
- GitHub state confirms that `v*` and `soku/v*` tags may be created but cannot be
  deleted or non-fast-forward updated.
- No credentials, cloud resources, deployment, release tag, catalog contract,
  provider contract, or existing public release is changed by this issue.

## Approval

- **Status:** `Approved`
- **Approved by:** `Soku-JINSEOK` (corrective implementation plan supplied for
  execution on 2026-07-21)
- **Approval boundary:** Repository changes, signed commits, branch push, and a
  Draft PR may proceed. Release publication and unrelated GitHub metadata
  normalization remain separate approval boundaries.

## Implementation Status

Implementation is prepared with the Issue #41 corrective release changes. The
implementation PR uses `Related to #41` and `Closes #56`; Issue #41 remains open
until the corrective release pair and public lifecycle evidence are complete.

## Verification

- Passed before report creation: full-history Gitleaks, OSV Scanner v2.4.0,
  JavaScript `npm audit`, both Go module vulnerability checks, renderer and
  lifecycle regressions, Markdown/YAML/action validation, and whitespace checks.
- Pending on the implementation PR: fresh local verification of the final diff
  and hosted aggregate `Validation Gate` evidence.
- GitHub state observed before the PR: immutable release-tag ruleset `19336418`
  is active; final evidence will revalidate it without mutating existing tags.

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex
