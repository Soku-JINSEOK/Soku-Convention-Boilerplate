# ✅ Verification Guide

## Purpose

This guide is the operational checklist for validating this repository, its
runtime templates, published artifacts, GitHub governance, historical
collaboration records, and cost exposure. Run the smallest relevant subset
during development and the complete checklist for a release or repository-wide
audit.

Record the command, date, environment, result, and any retained evidence in the
linked task report. A skipped or unavailable check is not a pass: record the
specific limitation and the follow-up needed to close it.

## Supported Release Baseline

- Current published boilerplate convention package: `v1.0.5`
- Current published CLI: `soku/v0.1.4`
- Recommended full-verification baseline: boilerplate `v1.0.5` with
  `soku/v0.1.4`.
- Superseded CLIs: `soku/v0.1.0` and `soku/v0.1.1`; use `soku/v0.1.2`, which
  preserves manifest-v1 and Provider API v1 while making the fetched provider
  revision authoritative and fully supporting optional legacy provider `ref`.
- Immutable `v1.0.0` limitations: generated JavaScript fails `init --verify`,
  and its dependency snapshots predate the current `tmp` and Jackson fixes.
- Published `v1.0.2` and `soku/v0.1.3` contain the source-authoritative
  renderer, but their public four-stack smoke found a cross-stack Prettier
  boundary defect, and `soku/v0.1.3` can record a false baseline after an
  unchanged mergeable-file transition. The `v1.0.3`/`soku/v0.1.4` companion
  pair was published to correct those defects.
- Published `v1.0.3` and `soku/v0.1.4` correct those defects, but required
  public four-stack smoke found that Python Ruff traverses a generated
  JavaScript `node_modules` tree. Published `v1.0.4` corrected that boundary,
  and its public migration smoke found that JavaScript formatting traversed
  lifecycle-owned `.soku/` state. Published `v1.0.5` excludes `.soku/` from
  JavaScript and TypeScript formatting and is the current baseline.

Published tags and releases are immutable. Verification must never move,
delete, or reuse them, and must not publish a new release as a side effect.

## Local Repository Checks

Run from the repository root unless a command changes directory explicitly.

```bash
npx --yes markdownlint-cli2@0.22.1 --config .markdownlint.jsonc \
  "**/*.md" "#**/node_modules/**"
npx --yes yaml-lint@1.7.0 .github/*.yml .github/**/*.yml \
  templates/**/*.yml templates/**/*.yaml
actionlint
git diff --check
```

Validate every JSON document against its declared schema where applicable, and
run a link checker across tracked Markdown. Links to authenticated settings may
return access errors to anonymous tools; verify those interactively and record
the distinction rather than suppressing all failures.

```bash
bash -n scripts/*.sh soku/scripts/*.sh
shellcheck scripts/*.sh soku/scripts/*.sh
pwsh -NoProfile -Command \
  '$errors = $null; [System.Management.Automation.Language.Parser]::ParseFile(
    "scripts/sync-boilerplate.ps1", [ref]$null, [ref]$errors) > $null;
    if ($errors) { $errors | Out-String | Write-Error; exit 1 }'
pwsh -NoProfile -Command \
  'Invoke-ScriptAnalyzer scripts/sync-boilerplate.ps1 -EnableExit'
scripts/verify-sync-parity.sh
scripts/verify-release-tag_test.sh
```

## Local CI/CD Parity and Cloud Run Deploy Commands

These are the canonical local commands for this repository's CD contract:

```bash
scripts/ci-local.sh --skip-infra
```

```bash
scripts/cd-plan.sh \
  --environment dev \
  --project-id <GCP_PROJECT_ID> \
  --region <GCP_REGION> \
  --service-name <CLOUD_RUN_SERVICE> \
  --artifact-repository <ARTIFACT_REPOSITORY> \
  --image-repository <IMAGE_REPOSITORY> \
  --skip-infra \
  --skip-image-push
```

Image-pushing plans require a registry digest and record the immutable digest URI
as `CD_PLAN_IMAGE_URI`; the tag URI is retained as `CD_PLAN_IMAGE_TAG_URI` for audit.
Use `--rollback-only` to create rollback metadata without Docker, local checks, or
Terraform.

```bash
scripts/cd-deploy.sh \
  --plan-file .cd/dev/<short-sha>/cd-plan.env \
  --health-path /health \
  --health-attempts 18 \
  --health-delay 10 \
  --confirm
```

Rollback command (manual):

```bash
scripts/cd-deploy.sh \
  --plan-file .cd/dev/<short-sha>/cd-plan.env \
  --rollback-only \
  --rollback-revision <revision-id> \
  --confirm
```

Deployment evidence is emitted as JSON in the path provided by the `evidence_file`
field in `$GITHUB_OUTPUT`. It records tag and digest URIs, pre-deploy and new
revisions, rollback target, run attempt, and final status.

## `soku` Checks

```bash
cd soku
go mod verify
go test ./...
go test -race ./...
go vet ./...
test -z "$(gofmt -l .)"
test -z "$(goimports -l .)"
golangci-lint run ./...
scripts/run_lifecycle_gate.sh "$(mktemp -d)/soku-lifecycle"
cd ..
soku/scripts/package_test.sh
```

The lifecycle gate covers the core and provider contract. The package check
must reproduce the five supported target archives, checksums, file modes, and
packaged binary smoke result without leaving tracked output.

## Runtime-Template Checks

Use the locked dependency files and commands documented in
[`STACK_CONFIGS.md`](./docs/guides/STACK_CONFIGS.md).

```bash
cd templates/javascript-typescript-node
npm ci
npm run lint
npm run typecheck
npm test
npm run build
npm run format:check

cd ../python
python -m venv .venv
.venv/bin/pip install -r requirements-lock.txt -e '.[dev]'
.venv/bin/ruff check .
.venv/bin/mypy .
.venv/bin/black --check .
.venv/bin/pytest

cd ../go
make fmt-check
make lint
make test
make build

cd ../java-spring
mvn -B verify
```

Run the MySQL and PostgreSQL schemas against the major versions declared in
`.github/workflows/templates-ci.yml`. Build `templates/gcloud/Dockerfile`, and
parse the AWS and Azure YAML. Hosted validation is authoritative when local
Docker, database servers, PowerShell, Windows, or macOS are unavailable.

## Security, Dependency, Secret, and License Checks

Use current scanners from their official distributions and record their
versions. At minimum:

- `govulncheck ./...` in `soku/` and the Go template;
- `npm audit` for the locked JavaScript template;
- `pip-audit -r requirements-lock.txt` for the Python template;
- OSV scanning across all supported lockfiles and modules;
- a full-history secret scan with verified findings reviewed manually (the
  repository config allowlists only the exact synthetic credential-rejection
  fixture used by `TestArchiveSecurityValidation`);
- license inventory comparison against `LICENSE`, dependency metadata, and
  `soku/THIRD_PARTY_NOTICES.md`.

Do not paste a discovered credential into an issue, pull request, log, or task
report. Revoke and rotate it through the appropriate provider, then record only
the sanitized incident and remediation status.

## Published Artifact Checks

For `v1.0.0`, `v1.0.1`, `v1.0.2`, and the corrective baselines
`v1.0.3`/`soku/v0.1.4` and `v1.0.4`/`soku/v0.1.4`:

1. Resolve each public tag and verify its signed annotated tag record and release
   metadata without changing either object.
2. Download every CLI asset for the selected immutable CLI release to a
   temporary directory.
3. Verify `checksums.txt`, archive names, archive contents, executable modes, and
   embedded version/commit/build metadata.
4. Smoke-test the native archive on the available platform.
5. Run `go install
   github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku@v0.1.2` in an
   isolated Go cache and use it for a read-only `v1.0.0` lifecycle smoke test.

Delete only the temporary audit directory after recording results. Release
assets must stay below 2 GiB each; GitHub currently documents no total release
size or bandwidth limit, with up to 1,000 assets per release. See
[About releases](https://docs.github.com/en/repositories/releasing-projects-on-github/about-releases).

## npm Wrapper Checks (`@soku-jinseok/soku`)

For `soku/v0.2.0` and later releases:

1. Install the published wrapper package to a temporary prefix:

   ```bash
   npm install -g @soku-jinseok/soku@0.2.0 --prefix "$RUNNER_TEMP/node-soku"
   ```

2. Run `soku --version` from that installation and confirm it matches
   `soku/v0.2.0` metadata.
3. Validate a lightweight command path (`soku status`) without making
   destructive changes.

If your environment blocks global install, validate offline by checking:

```bash
cd soku/npm
npm test
node scripts/prepare-package.mjs --version 0.2.0 --repo-root ../..
```

Record whether runtime verification passed, including any network or publish
restriction that required a reduced evidence scope.

## Issue and Pull Request Audit

Audit every issue and merged pull request in scope against the standards that
applied when it was created. Do not retroactively rewrite historical bodies or
commits merely to match a newer template.

For each issue, check:

- state and completion reason;
- at least one `type:` label plus applicable priority, status, and area labels;
- goal, scope, constraints, acceptance criteria, and verification evidence;
- issue, approved task report, and implementation pull-request links.

For each pull request, check:

- merged state and labels consistent with completion;
- title and in-scope commit titles;
- required checks, review/conversation state, and merge method;
- commit and merge signature status;
- required body fields and links that existed at creation time.

Correct metadata such as labels when intent is unambiguous. Record historical
exceptions centrally; do not rewrite the artifact or its commit history.

## GitHub Repository Audit

Confirm through the API after the implementation workflow has succeeded:

- repository visibility and default branch;
- only standard hosted runner labels are used;
- workflow default permissions are read-only and only release delivery receives
  `contents: write`;
- every external action is pinned to a verified full commit SHA with a version
  comment;
- action SHA-pinning enforcement is enabled where the current plan supports it;
- the active `main` ruleset requires a pull request with zero approvals for this
  personal repository, the latest `Validation Gate`, signed commits, and resolved
  conversations, while blocking deletion and force-push;
- merge commits remain allowed and no routine bypass actor exists;
- Actions artifact/cache counts, sizes, retention, release assets, Packages,
  Git LFS, Codespaces, Marketplace apps, and external services match the audit
  record.

Rulesets are available for public repositories on GitHub Free. See
[available rules for rulesets](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/available-rules-for-rulesets).

## Cost Audit

The repository audit and personal-account billing audit are separate evidence
sets.

1. Record repository visibility, runner labels and duration, Actions artifacts,
   caches, Packages, LFS, Codespaces, release assets, Marketplace actions, and
   external service use.
2. In the authenticated Billing and licensing pages, record the account plan,
   current-month gross and net metered usage, Actions compute SKUs and storage,
   Packages/LFS/Codespaces usage, paid subscriptions and Marketplace apps,
   budget/alert state, and whether a next charge and payment method exist.
3. Classify each result as repository-attributable metered cost, pre-existing
   personal-account usage/subscription, or future quota risk.

Never record card details, billing addresses, payment identifiers, credentials,
or authentication material. Do not change a budget, alert, subscription, or
payment setting during the audit. Report nonzero cost accurately; do not cancel
it automatically.

Standard GitHub-hosted runners are free for public repositories. Larger runners
are always billed, including for public repositories. Actions storage is shared
with Packages under plan allowances. Reusable-workflow usage is attributed to
the caller. See [GitHub Actions billing](https://docs.github.com/en/billing/concepts/product-billing/github-actions)
and [billing and usage](https://docs.github.com/en/actions/concepts/billing-and-usage).

## Failure Handling and Completion

- Stop delivery on any required validation failure.
- Fix an in-scope non-breaking defect on the implementation branch and rerun the
  smallest failed check plus the complete affected gate.
- Open a follow-up issue for a breaking change, a new-release requirement, an
  unavailable external prerequisite, or a cost decision needing account-owner
  approval.
- Never mark a skipped, flaky, or inaccessible check as passed.
- Before closing the audit issue, verify the final protected `main` gate, final
  evidence pull request, labels and links, public tags/releases, and a clean
  local worktree.
