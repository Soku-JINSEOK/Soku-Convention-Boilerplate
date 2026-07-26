# Issue #115 Task Report — Add fast verification and scope detection

## Goal and Background

Issue [#115](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/issues/115)
completes the local phase of the profile-based verification roadmap. Developers
need a fail-closed changed-file detector and a fast pre-commit profile without
weakening the existing full local or hosted validation gates.

## Proposed Approach

Keep the configuration dependency-free and constrained to a documented YAML
subset. Detect changes from the staged index, an explicit commit range, or an
explicit file list. Map every known path to a named verification scope and
select all scopes when a shared or unknown path could invalidate the mapping.

The fast profile always checks changed-file whitespace, repository hygiene, and
changed-line secrets. It then runs only the selected Soku, runtime-template,
database/config, or infrastructure smoke checks. Full-only race, lifecycle,
packaging, OS-matrix, and complete vulnerability checks remain outside the fast
profile.

## Planned Implementation

- Add `verification/profiles.yml` and `verification/scopes.yml` using one
  constrained, versioned YAML schema.
- Add `scripts/detect-verification-scope.mjs` and regression tests for staged,
  commit-range, file/stdin, rename/delete, shared/global, unknown, and empty
  inputs.
- Add scoped command entry points and implement
  `scripts/verify.sh --profile fast`.
- Always run diff whitespace and changed-line secret checks in the fast
  profile.
- Add `scripts/bootstrap-dev.sh`, `scripts/bootstrap-dev.ps1`,
  `scripts/verify.ps1`, `.devcontainer/`, and Git hooks.
- Keep `.soku/verification/local-full.json` optional, ignored, and explicitly
  non-authoritative.
- Document usage, Windows parity, Docker Compose database verification, and
  the measured fast/full timing ratio.

## Acceptance Criteria

- The detector emits `schemaVersion`, `changedFiles`, `scopes`, `reasons`, and
  `allSelected` as JSON.
- No argument and `--staged` inspect staged changes;
  `--base <sha> --head <sha>` and `--files-from <path|->` are supported.
- Rename/delete inputs are recognized, and shared, generator, sync, lint,
  provider/catalog schema, or unknown paths select all scopes.
- A single known template change selects only that template plus always-on
  checks.
- `scripts/verify.sh --profile fast` omits race, complete lifecycle,
  cross-platform, full packaging, and full vulnerability scanning.
- Pre-commit runs `fast`; pre-push runs `full`.
- Windows and dev-container entry points preserve the same profile contract.
- A single-template fast run is no more than 25% of the median same-environment
  full run.
- Existing required checks, branch protection, release, and CD behavior do not
  change.

## Approval

- **Status:** `Approved`
- **Approved by:** `Soku-JINSEOK` through the user-authorized implementation
  plan supplied for this continuation

## Implementation Status

Implemented in pull request
[#141](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/pull/141).
The first hosted pull-request run passed the full Validation Gate, PR Metadata
Gate, CodeQL analysis, and the external Cloud Build check. The evidence update
itself must pass the same required checks before merge.

## Verification

- Scope-detector regression suite: 10/10 passing, including staged, range,
  file/stdin, rename/delete, shared/global, unknown, and empty inputs.
- Verification and diff-secret regression suites: 21/21 combined tests
  passing.
- Bash syntax and ShellCheck 0.11.0: passing for scripts, verification
  commands, and hooks.
- YAML lint: passing for profiles, scopes, and the modified workflow.
- Markdown lint: passing for all modified operational documentation.
- Immutable supply-chain and Dependabot coverage: 9/9 tests passing; current
  repository passes with the pinned dev-container image and 11 update targets.
- Documentation-only fast profile: passing in 0.9 seconds with no runtime
  scope selected.
- Affected Soku-package fast profile: passing for formatting, unit test, vet,
  native build, and smoke without race/lifecycle/packaging.
- JavaScript-template fast profile: three successful samples at 10.75, 10.93,
  and 11.64 seconds; median 10.93 seconds.
- Full profile with DB/Terraform excluded for the same-environment timing
  comparison: three successful samples at 80.61, 82.45, and 83.57 seconds;
  median 82.45 seconds. Fast/full ratio: 13.26%, below the 25% limit.
- Pinned MySQL/PostgreSQL Docker Compose services: healthy; both schemas loaded
  successfully and services were removed by the cleanup trap.
- Dev-container Dockerfile: Docker build check passed with no warnings.
- Full security group: passing after moving pinned `pip-audit` installation
  into an isolated temporary virtual environment.
- Local Terraform and PowerShell execution: unavailable; the pull-request
  workflow provided the authoritative Windows PowerShell parsing and hosted
  infrastructure validation.
- Hosted full validation:
  [Actions run 30200224278](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/30200224278)
  passed the Linux, macOS, and Windows Soku matrix, all runtime templates,
  MySQL/PostgreSQL schemas, security scans, repository hygiene, and the
  aggregate `Validation Gate`.
- Hosted CodeQL:
  [Actions run 30200223609](https://github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/actions/runs/30200223609)
  passed the Actions, Go, Java/Kotlin, JavaScript/TypeScript, and Python
  analyses.
- External hosted validation: the Cloud Build pull-request check reported
  success on PR #141; the public report intentionally omits cloud account and
  build identifiers.

## Public Disclosure Review

- [x] No credentials, tokens, private keys, or credential-bearing URLs
- [x] No private repository, project, or product names
- [x] No cloud project IDs, account numbers, service URLs, image URIs, or
      revision identifiers
- [x] No personal billing, subscription, budget, or payment-status information
- [x] No personal email, phone, address, or local absolute path
- [x] No private Issue, PR, Project, or control-plane identifiers

## AI Assistance

- **Planning/implementation/drafting:** OpenAI Codex (GPT-5.6)
