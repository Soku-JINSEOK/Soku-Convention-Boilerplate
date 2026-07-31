# GCP infrastructure

This Terraform stack deliberately separates bootstrap from runtime creation while
keeping both stages in one remote GCS state.

- Foundation (`deploy_runtime=false`) enables APIs and creates Artifact Registry,
  service accounts, IAM, and GitHub Workload Identity Federation. It needs no
  container image. The deployer may act as only the dedicated runtime service
  account and may write only to the configured Artifact Registry repository.
  It can mint an ID token only for itself and invoke only this Cloud Run service,
  allowing authenticated private post-deploy health checks without project-wide
  Token Creator or Invoker access.
- Runtime (`deploy_runtime=true`) creates Cloud Run and requires an immutable
  `repository@sha256:<digest>` value in `image_uri`.
- Cloud Build validation (`enable_cloud_build_validation=true`) enables the
  Cloud Build API and creates a dedicated service account with only
  `roles/logging.logWriter`. It also creates two `asia-northeast1`,
  first-generation GitHub App triggers. Neither trigger reuses the GitHub
  Actions deployer.

The GCS backend is partial configuration. Initialize it with the project-derived
bucket rather than committing backend values or state:

```bash
terraform -chdir=infra/gcp init \
  -backend-config="bucket=${GCP_PROJECT_ID}-tfstate" \
  -backend-config="prefix=cloud-run"
```

The bootstrap applies `state-lifecycle.json` to the versioned state bucket.
Noncurrent objects are deleted only when they are at least 30 days old and have
at least 10 newer versions. Current state objects are never lifecycle deletion
candidates.

Set `GCP_PROJECT_ID=<id>` and run `scripts/gcp-bootstrap.sh` to preview the full
sequence. `--project-id` is also supported and takes precedence over the
environment. Actual creation additionally requires
`--apply --confirm-project-id <id>`.

For the low-cost sandbox profile, also pass:

```bash
export TF_VAR_billing_account_id='<private-billing-account-id>'
scripts/gcp-bootstrap.sh \
  --project-id '<gcp-project-id>' \
  --max-instances 1 \
  --enable-budget-alerts \
  --monthly-budget-amount 1500 \
  --apply \
  --confirm-project-id '<gcp-project-id>'
```

The billing account value is a sensitive Terraform input and must not be put in
committed variable files, command output, Issues, or build artifacts. The
default budget currency is JPY; set the private Terraform input
`TF_VAR_budget_currency_code` to the billing account currency when it differs.
Budget alerts fire at 50%, 80%, and 100% of current spend. A budget is an alert,
not a spending cap, and this stack never disables billing automatically.

Artifact Registry cleanup starts in dry-run mode. It reports untagged images
older than 7 days and ordinary `sha-`/`commit-` images older than 30 days while
retaining the five most recent versions and all `release-`/`protected-` tags.
Review at least seven days of cleanup logs before passing
`--activate-artifact-cleanup`; that flag enables deletion on apply. Keep the
flag on later applies once activation has been approved.

Add `--enable-cloud-build-validation` to the preview and apply commands to
manage the opt-in validation resources. Retain the flag on later applies;
omitting it restores the default `false` value and plans removal of the
validation service account, IAM binding, and triggers.

An apply with this flag is validation-only: after the targeted API, identity,
IAM, and trigger apply succeeds, the bootstrap exits before Docker
authentication, image build or push, runtime Terraform, and repository-variable
writes. This keeps an existing Cloud Run service unchanged.

Validation resources use the same state bucket with the isolated
`cloud-build-validation` prefix. Cloud Run foundation and runtime resources
remain under `cloud-run`, so later deployment plans cannot remove validation
triggers through the variable's default `false` value.

Validation log routing is a separate lifecycle boundary. The three logging
resources live in [`cloud-build-logging`](./cloud-build-logging/README.md) and
use the matching `cloud-build-logging` backend prefix. This root does not
manage, import, or delete foundation, trigger, IAM, WIF, billing, or
`shared-artifacts` state.

During apply, the bootstrap resolves immutable GitHub repository and owner IDs.
The WIF provider accepts only the configured repository IDs, `refs/heads/main`,
and `.github/workflows/deploy-gcp.yml` from `main`. The state bucket is hardened
with uniform access, enforced public-access prevention, object versioning, and
no legacy project Viewer object access.

The outputs `wif_provider_name` and `deployer_service_account_email` map directly
to the GitHub repository variables `GCP_WIF_PROVIDER` and
`GCP_WIF_SERVICE_ACCOUNT`.

## Validation-only Cloud Build triggers

The opt-in creates these triggers:

| Trigger | Event | Branch | Execution control |
| --- | --- | --- | --- |
| `soku-convention-boilerplate-pr` | Pull request | `^main$` | A writer must comment `/gcbrun` |
| `soku-convention-boilerplate-main` | Push | `^main$` | Not applicable |

Both triggers use `cloudbuild/validation.yaml`, run as the dedicated validation
identity, store logs with `CLOUD_LOGGING_ONLY`, and return build status and log
links to GitHub. Both run in `asia-northeast1`; their existing names remain
unchanged so the GitHub Check contexts remain stable. The build runs Node 24 GCP regression tests, Terraform
format/init/validate/tests, and an amd64 build of `templates/gcloud`. It has no
artifact output, image push, secret access, or deployment step.
The build uses the default worker machine and has a 15-minute overall timeout.

Do not put `/gcbrun` in the pull-request body. Keep the PR in Draft while its
head is changing, mark it Ready only after the head is stable, and then have a
repository writer add one `/gcbrun` comment. A new head commit invalidates that
evidence: cancel an obsolete running build and approve the latest head again.
The PR trigger intentionally has no path filter while its check is a required
context. Add the same GCP-only filter as the main trigger only after the
required-check migration tracked in issue #117 is approved.

The main trigger remains automatic but runs only when a push changes
`infra/gcp/**`, `cloudbuild/**`, `templates/gcloud/**`,
`scripts/gcp-bootstrap.sh`, or the two GCP regression test files. Its squash
merge SHA is distinct verification and is not considered a duplicate of the PR
head build.

The `github` block in each Terraform trigger is the first-generation GitHub App
interface. Terraform trigger creation is also the connection check: apply fails
if the repository is not already connected to the selected GCP project. This
stack does not create a second-generation Cloud Build connection or repository.

### Regional trigger migration

Cloud Build trigger names are unique across the project, not per location.
Consequently, Terraform cannot create a Tokyo trigger with a canonical name
while the global trigger with that name still exists. Do not apply the location
replacement directly and do not use a temporary alias.

Before the approved cutover window:

1. Export the complete JSON, ID, service account, substitutions, branch and path
   filters, and comment control of both global triggers as private rollback
   evidence.
2. Pull a separate backup of the `cloud-build-validation` remote state and
   verify that it contains the two canonical addresses
   `google_cloudbuild_trigger.pull_request[0]` and
   `google_cloudbuild_trigger.main[0]`.
3. Change the global PR trigger to `COMMENTS_ENABLED` so new pushes do not
   create PR builds.

During the separately approved cutover, keep each mutation explicit:

1. Remove only the two canonical trigger addresses from Terraform state. This
   must not destroy the live global triggers.
2. Delete the two global trigger resources, then immediately create the Tokyo
   triggers with the same canonical names and saved configuration.
3. Run the PR and main triggers at an exact commit SHA and verify their source,
   validation-only service account, unchanged GitHub Check names, complete logs,
   and `locations/asia-northeast1` resource names.
4. Import each verified Tokyo trigger ID into its original canonical Terraform
   address, then require a clean full plan with
   `enable_cloud_build_validation=true`.

The create step must follow the global delete because both locations cannot own
the same trigger name concurrently. Then verify one controlled Ready PR with
one writer-issued `/gcbrun` and the next matching `main` merge. The acceptance
state is no active global trigger and two active `asia-northeast1` triggers. Do
not delete completed build history, existing Artifact Registry images, or old
logs; let the existing 30-day log retention expire naturally.

Cloud Build log routing is intentionally not owned by this per-repository
stack. The shared `ci-cd-control-plane` Terraform root owns the Tokyo log bucket,
sink, and `_Default` exclusion for all three repositories. Create the bucket
and sink and verify duplicate delivery before enabling the exclusion. Never
modify the immutable `_Required` audit-log sink. If regional validation fails,
disable the `_Default` exclusion first. Delete the failed Tokyo pair, restore
the global pair from the saved JSON, import the restored global IDs at the two
canonical state addresses, and check a plan against the previous global code.
Restore the remote-state backup only if the targeted state recovery cannot be
completed safely.

To roll back only the Cloud Build integration, plan and apply with
`enable_cloud_build_validation=false`. This removes its triggers and dedicated
IAM resources while preserving the enabled API, GitHub App connection, and
GitHub Actions deployment path.
