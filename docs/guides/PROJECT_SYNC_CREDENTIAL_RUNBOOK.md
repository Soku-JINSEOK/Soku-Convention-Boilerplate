# Project Sync Credential Setup and Rotation Runbook

## Purpose

This runbook defines the least-privilege setup, verification, replacement, and
revocation procedure for the credential stored as the repository Actions secret
`PROJECT_SYNC_TOKEN`. It applies only to the user-owned GitHub Project Sync
component documented in [GITHUB_PROJECT_SYNC.md](./GITHUB_PROJECT_SYNC.md).

The procedure is deliberately fail-closed. It does not authorize credential
generation by repository automation, a broader fallback token, organization
Project support, workflow administration, delivery, release, cloud, or billing
changes.

## Roles and approval boundary

The same person may hold more than one role in a personal repository, but each
decision remains explicit:

| Role | Responsibility |
| --- | --- |
| Credential owner | Creates and revokes the dedicated credential outside this repository. |
| Repository maintainer | Replaces the `PROJECT_SYNC_TOKEN` Actions secret. |
| Project owner | Reviews audit findings and approves any apply run. |
| Operator | Runs the commands and records only the rotation date and redacted outcome. |

Creating the replacement credential, changing an Actions secret, selecting
apply mode, and revoking the old credential are four separate mutations. A
successful audit authorizes none of the later mutations by itself.

## Required permission matrix

The replacement credential must be dedicated to Project Sync and must have
exactly the following effective permissions:

| Resource | Permission | Access | Required operation |
| --- | --- | --- | --- |
| Repository | Metadata | Read | Resolve repository identity and public metadata. |
| Repository | Issues | Read and write | Read and normalize Issue labels and relations. |
| Repository | Pull requests | Read and write | Read and normalize pull request labels and Issue relations. |
| Authenticated user | Projects | Read and write | Read existing Project #2 items and update approved fields. |

Do not grant repository Contents write, Actions administration, workflow edit,
release, package, deployment, organization administration, or billing access.
If the selected credential type cannot express the matrix without broader
access, stop and choose a credential type that can; do not accept excess
permission as a temporary workaround.

Repository workflow permissions remain independently bounded. The Project Sync
workflow checks out trusted code with `contents: read` and receives only the
`PROJECT_SYNC_TOKEN` secret at runtime. The credential value must never be
stored in a file, repository variable, command argument, report, Issue, pull
request, shell transcript, or collaboration message.

## Rotation states

Use these states so the active credential and the next safe action are clear:

| State | Old credential | Replacement credential | Repository secret | Permitted next action |
| --- | --- | --- | --- | --- |
| `prepared` | Active | Created, not installed | Old value | Audit replacement only. |
| `replacement-audited` | Active | Audit passed | Old value | Approve secret replacement. |
| `secret-replaced` | Active as rollback | Installed | Replacement value | Run post-replacement audit. |
| `replacement-verified` | Active as rollback | Audit passed | Replacement value | Optional approved apply. |
| `rotation-complete` | Revoked | Active | Replacement value | Record redacted outcome. |
| `aborted` | Active or restored | Quarantined/revoked | Known-good old value | Investigate without broadening access. |

Never revoke the old credential before the replacement passes both its direct
audit and the post-secret-replacement audit.

## Preflight checklist

Before creating or changing any credential:

1. Confirm the authenticated account is the intended Project and repository
   owner. Stop on a similar or unexpected account name.
2. Confirm the repository and existing user-owned Project number. Do not create
   a Project as part of rotation.
3. Confirm the current workflow still references only
   `secrets.PROJECT_SYNC_TOKEN` and uses the trusted base revision.
4. Confirm a maintainer can restore the old secret value until replacement
   verification completes.
5. Select a UTC rotation date and an operator. Do not record a credential ID,
   fingerprint, value, expiry token, or settings URL.
6. Confirm the audit output path is temporary, owner-readable only, outside the
   repository, and excluded from terminal/session capture.
7. Confirm that apply, if needed, has separate approval. The default rotation
   verification is audit-only.

Any failed item stops the rotation before mutation.

## Phase 1: prepare and directly audit the replacement

Create the replacement outside repository automation with the exact permission
matrix above. Keep the old credential active and leave the repository secret
unchanged.

Supply the replacement to the process through an approved secret-injection
mechanism as `PROJECT_SYNC_TOKEN`. Do not paste it into the command line. Set
task-specific shell variables for the public target and a temporary report path,
then run:

```bash
umask 077
node scripts/github-project-sync.mjs \
  --mode audit \
  --repo "$PROJECT_SYNC_TARGET_REPOSITORY" \
  --project-owner @me \
  --project-number "$PROJECT_SYNC_TARGET_NUMBER" \
  --output "$PROJECT_SYNC_AUDIT_REPORT"
```

The tool writes the report with mode `0600`. It stores body hashes instead of
body text and excludes raw metadata, but the report is still operational data
and must remain temporary.

The direct replacement audit passes only when all of the following are true:

- the command exits successfully;
- `schemaVersion` is `1` and `mode` is `audit`;
- `privacy.bodiesStored` is `false`;
- `privacy.bodyHashesStored` is `true`;
- `privacy.rawMetadataExcluded` is `true`;
- `summary.failures` is `0`;
- every warning and planned operation is understood; and
- no output contains a credential value, Authorization header, raw Issue or
  pull request body, credential-bearing URL, or private environment value.

A clean audit may contain planned changes. Review them; do not turn audit into
apply merely to test write access. If the permission matrix is insufficient,
stop and diagnose the missing narrow permission. Never retry with a personal or
broadly scoped fallback token.

## Phase 2: replace the repository secret

After the direct audit passes and the secret replacement is explicitly
approved:

1. Keep the old credential active and recoverable.
2. Replace only the repository Actions secret named `PROJECT_SYNC_TOKEN` using
   the GitHub settings surface or an approved secret manager integration.
3. Do not change repository variables, workflow files, permissions, branch
   rules, environments, or delivery settings during this step.
4. Record only that secret replacement succeeded or failed. Never record the
   value, a credential identifier, or a settings URL.

The repository does not contain a command that creates or rotates this secret.
That absence is an intentional approval boundary.

## Phase 3: post-replacement audit

Run the Project Sync workflow manually in its default `audit` mode, or run the
same local audit command with the repository-secret-equivalent credential
injection. Do not select apply.

The post-replacement audit must meet the same pass criteria as the direct audit.
Additionally confirm that:

- the workflow used the trusted base revision;
- the expected repository and existing user-owned Project were selected;
- no duplicate pull request Project item was proposed;
- existing Project items, custom labels, and project-owned settings were
  preserved; and
- the report contains hashes and sanitized before/after values, not raw bodies
  or credential material.

Do not treat successful authentication alone as a passing audit.

## Phase 4: optional separately approved apply

Rotation does not normally require apply. If the reviewed audit contains
legitimate synchronization changes, obtain explicit Project-owner approval and
run apply once against the same intended repository and Project:

```bash
umask 077
node scripts/github-project-sync.mjs \
  --mode apply \
  --repo "$PROJECT_SYNC_TARGET_REPOSITORY" \
  --project-owner @me \
  --project-number "$PROJECT_SYNC_TARGET_NUMBER" \
  --output "$PROJECT_SYNC_APPLY_REPORT"
```

Apply rereads each target immediately before mutation. A changed target is
reported as `stale-read` and skipped; do not bypass that protection or broaden
the target to force success. After apply, run a fresh audit and require zero
unexplained operations and zero failures.

## Phase 5: revoke and close out

Revoke the old credential only after the replacement and post-replacement
audits pass, and after any separately approved apply is followed by a clean
audit. Confirm revocation through the credential owner, then remove temporary
reports according to the operator's approved secure-disposal policy.

The durable rotation record contains only:

```text
Rotation date (UTC): YYYY-MM-DD
Outcome: passed | aborted | rolled-back
Direct replacement audit: passed | failed | not-run
Secret replacement: passed | failed | not-run
Post-replacement audit: passed | failed | not-run
Approved apply: passed | failed | not-required | not-run
Old credential revocation: confirmed | not-revoked
Redaction review: passed | failed
```

Do not add report contents, hashes that identify private metadata, run IDs,
credential identifiers, settings URLs, actor email addresses, or token expiry
details to the durable record.

## Abort and rollback matrix

| Failure point | Required response | Old credential |
| --- | --- | --- |
| Replacement creation or direct audit | Stop; quarantine or revoke the replacement and investigate the narrow permission failure. | Keep active. |
| Secret replacement | Stop; confirm the repository still has a known-good value before any workflow run. | Keep active. |
| Post-replacement audit authentication | Restore the old secret value, audit with the old credential, and stop. | Keep active. |
| Post-replacement audit content mismatch | Restore the old secret if credential behavior is suspect; otherwise leave apply disabled and investigate the target/configuration. | Keep active. |
| Apply reports `stale-read` | Do not retry blindly; audit the changed target and obtain renewed approval. | Keep active until closeout. |
| Apply reports a mutation error | Stop further broad runs, audit current state, and repair only approved targets. | Keep active until closeout. |
| Old credential revocation cannot be confirmed | Keep the rotation open and escalate to the credential owner; do not claim completion. | Treat as active. |

Rollback means restoring the known-good old secret while the old credential is
still valid, then proving that state with an audit. It never means weakening
permissions, disabling validation, force-writing Project fields, or fabricating
a successful report.

## Redaction checklist

Before sharing any outcome, confirm every item:

- [ ] No credential, token, private key, Authorization header, or secret value.
- [ ] No credential-bearing URL, settings URL, or secret-manager path.
- [ ] No raw Issue or pull request body.
- [ ] No private Project item, field, repository, or account identifier.
- [ ] No actor email, machine path, shell history, environment dump, or billing
      information.
- [ ] Only the UTC rotation date and the allowlisted redacted outcome fields are
      durable.

If redaction cannot be verified, do not publish the evidence. Keep the rotation
open and repeat the audit with safe capture settings.

## Periodic review

Review the permission matrix and this runbook before each rotation and after a
GitHub permission-model change. A scheduled audit may prove current behavior,
but it must not rotate credentials automatically. Expiry policy, credential
ownership, and the next review date belong in the operator's external secret
inventory, not in this repository.
