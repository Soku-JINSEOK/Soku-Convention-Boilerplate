# Issue #117 — code-validation preservation and current-main Hosted Full forward-port

> **Document purpose:** Task decision record. Records the scope, decisions and verification for Issue #117: preserving required code validation and adding current-source Hosted Full.
>
> **Key point:** Read approval, implementation status and verification separately; this report is dated task evidence, not a fresh claim of current completion.

## Purpose

This integration combines the narrow code-validation correction with the
additive current-main Hosted Full caller. It preserves the existing required
contexts and does not mutate branch protection, rulesets, delivery, or deploy
behavior.

## Code-validation correction

Code run 34682376195 on PR #229 failed while later metadata-only runs skipped
all code groups and reported success with the same required check names.
Concurrency isolation did not prevent that overwrite. Code events retain
`CI Quick Gate` and `Validation Gate`; metadata events use
`CI Quick Metadata Only` and `Validation Metadata Only`. Names do not depend
on job success or cancellation.

The integrated regression tests evaluate the actual workflow expressions for
code and metadata events, including a base edit, and execute the aggregate
shell. Failed, cancelled, or unexpectedly skipped code groups fail closed.
No extra trigger, permission, or heavy validation execution is added.

## Current-main Hosted Full contract

The current security workflow accepts two distinct inputs:

- `base-sha`: the trusted policy and comparison source;
- `head-sha`: the exact repository commit to scan.

Repository and template reusable workflows receive `head-sha` and pass it to
every checkout. Hosted Full passes both values to Security and aggregates
repository, template, and Security results into Hosted Full Gate. Manual and
scheduled Hosted Full runs default both values to `github.sha`; reusable
callers provide exact values when validating a separate source.

## Integrated scope

- Preserve the #230 metadata/code gate separation and fail-closed aggregate
  tests.
- Forward the #233 exact-head input through `ci.yml` and
  `templates-ci.yml`.
- Add `.github/workflows/full-validation.yml` with repository, template, and
  security aggregation.
- Keep release, deploy, IAM, ruleset, and Cloud Build resources unchanged.
- Retain both historical source reports in this single review record.

## Verification evidence

The source candidates separately recorded:

- #230: regression tests for metadata-only naming, cancellation isolation,
  exact head/base propagation, and aggregate shell failures.
- #233: 48 hosted-contract tests passed; PyYAML parsed the reusable workflows;
  39 protected files and 11 update targets were checked; exact-head checkout
  propagation and `git diff --check) passed.

The new integration commit requires fresh hosted validation against its exact
head. Earlier results are not reused as evidence for that commit.

## Follow-up validation — 2026-09-15

The six Node regression suites now convert repository-relative file URLs with
`fileURLToPath`, preserving Unicode workspace paths such as the local Korean
directory. The combined `node --test scripts/*.test.mjs .github/*.test.mjs`
run passed 210/210. The full local profile with infrastructure and database
checks explicitly skipped stopped at the release-tag regression because the
local GPG runtime was unavailable. Checks after that failure were not verified
by that run.

The 2026-09-16 readback of head
`e34feca6a21186a70390cf9c360b2a246aa7c0bd` found successful PR Validation,
policy, and metadata runs. All 76 returned check runs were successful or
intentionally skipped. This does not establish completion of the separate
Hosted Full workflow or its Hosted Full Gate. The earlier Hosted Full
completion claim is withdrawn. The PR remains Draft pending the applicable
hosted, owner, signing, and merge gates.

On 2026-09-16, the isolated release-tag regression passed using the installed
GnuPG 2.5.21 and libgcrypt 1.12.2 with a temporary launcher that supplies the
library path to Git's GPG subprocess. The test used temporary repositories and
temporary keys; it did not sign or publish a real release.

The resumed full run passed repository hygiene, Soku unit/race/lifecycle and
five-target reproducible packaging checks, all runtime templates, and the
MySQL/PostgreSQL schema checks. Its secret scan then stopped on a historical
synthetic key in the strict-config rejection test. `.gitleaksignore` records
only that exact commit/file/rule/line fingerprint. A negative control confirmed
that the same synthetic key at another location is still detected.

Both Terraform roots passed formatting, backend-disabled initialization, and
validation using the profile's pinned Terraform 1.15.3 container. Local Go
1.26.5 produced four reachable standard-library vulnerability findings; the
security profile passed with fixed toolchain Go 1.26.6, including both Go
modules, npm and Python dependency audits, and the OSV scan.

The final uninterrupted `scripts/verify.sh --profile full` run subsequently
passed with exit code 0 on 2026-09-16, with Go 1.26.6, Node 22.23.2, the
temporary GPG launcher, and Docker Desktop 28.0.4. Neither `--skip-infra` nor
`--skip-db` was used. The profile's explicitly hosted-only checks remain
unexecuted and are not counted as local passes. This validates the working
tree containing the narrow historical-fixture exclusion, not a new hosted
commit or release.

The cached actionlint source was built into an isolated temporary binary and
reported no findings for `.github/workflows/*.yml`.

## Remaining gates

Manual Validation now also calls the exact-head Hosted Full workflow. This
provides a registered dispatch entry point while `full-validation.yml` is
still absent from the default branch. Automatic PR events do not invoke the
additional caller. The workflow regression, policy, and supply-chain suites
pass 43 tests after this change.

This source change does not run Hosted Full, alter GitHub rulesets, create or
change Cloud Build triggers, publish a release, or deploy. Issue #116
comparison evidence, owner-approved hosted acceptance, and any ruleset
transition remain separate gates. The predecessor PR #158 remains preserved
because its source-SHA implementation targets an older workflow contract.

## 2026-09-27 successor integration

The new successor branch starts at remote `main`
`0f9ac36abe4255c02cc27487cc68dc34bd8aaeba` and carries the Hosted Full
contract together with the runner coverage and dependency corrections described
in the Issue #227 report. It also includes the metadata event concurrency fix
from #234. Its focused workflow, PR policy and supply-chain tests passed 43/43;
actionlint v1.7.12 and the uninterrupted full local profile passed. A new
signed commit, exact-head PR checks and manual Hosted Full are required before
merge. The prior #236 check history does not
substitute for this candidate's checks. The CI ruleset transition remains
dependent on separate natural GCP-shadow evidence and an owner decision.

## Required-context correction after hosted observation

On PR #237, code-event Validation and manually dispatched Hosted Full passed at
`78791d1675b8a7fe277fdb0b1351dac9cbbfd678`. Later label, assignment,
description, and Ready events started metadata-only Validation runs. GitHub then
showed `Validation Gate` as expected but absent and blocked the squash merge.
Rerunning the original code workflow did not restore the required context.

The correction subscribes Validation only to `opened`, `synchronize`,
`reopened`, and `edited` PR events. Every subscribed PR run now executes the
code-bearing Quick, repository, template, and Security paths with stable
required gate names. `edited` deliberately reruns all groups because a base
branch edit is among those events. The separate PR policy workflow continues
to check labels, assignment, title, and body changes. The regression tests
assert this event split and reject failed, cancelled, or skipped aggregate
results. Fresh hosted checks are required on the correction commit.
The correction working tree also passed the uninterrupted full local profile,
including DB, Terraform, and security scopes, with exit code 0; actionlint
v1.7.12 and 41 focused regression tests passed.
