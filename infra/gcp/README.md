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
  `roles/logging.logWriter`. It also creates two global, first-generation
  GitHub App triggers. Neither trigger reuses the GitHub Actions deployer.
- CI image promotion uses a second WIF provider restricted to the immutable
  repository/owner IDs, canonical `main`, and
  `.github/workflows/validation.yml`. Its service account has only
  repository-scoped Artifact Registry writer access.

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

During apply, the bootstrap resolves immutable GitHub repository and owner IDs.
The WIF provider accepts only the configured repository IDs, `refs/heads/main`,
and `.github/workflows/deploy-gcp.yml` from `main`. The state bucket is hardened
with uniform access, enforced public-access prevention, object versioning, and
no legacy project Viewer object access.

The outputs `wif_provider_name` and `deployer_service_account_email` map directly
to the GitHub repository variables `GCP_WIF_PROVIDER` and
`GCP_WIF_SERVICE_ACCOUNT`. `wif_ci_provider_name` and
`ci_builder_service_account_email` map to `GCP_CI_WIF_PROVIDER` and
`GCP_CI_WIF_SERVICE_ACCOUNT`.

Use `scripts/gcp-bootstrap.sh --ci-builder-only` to apply only the CI identity
and its two repository variables to an existing foundation. After a real
CI-built digest deploy and rollback path both succeed, set
`grant_deployer_artifact_writer=false` to remove the deployer's transitional
writer binding.

## Validation-only Cloud Build triggers

The opt-in creates these triggers:

| Trigger | Event | Branch | External contributor control |
| --- | --- | --- | --- |
| `soku-convention-boilerplate-pr` | Pull request | `^main$` | A writer must comment `/gcbrun` |
| `soku-convention-boilerplate-main` | Push | `^main$` | Not applicable |

Both triggers use `cloudbuild/validation.yaml`, run as the dedicated validation
identity, store logs with `CLOUD_LOGGING_ONLY`, and return build status and log
links to GitHub. The build runs Node 24 GCP regression tests, Terraform
format/init/validate/tests, and an amd64 build of `templates/gcloud`. It has no
artifact output, image push, secret access, or deployment step.
The build uses the default worker machine and has a 15-minute overall timeout.

The `github` block in each Terraform trigger is the first-generation GitHub App
interface. Terraform trigger creation is also the connection check: apply fails
if the repository is not already connected to the selected GCP project. This
stack does not create a second-generation Cloud Build connection or repository.

Keep the PR trigger informational until a controlled PR build and the
post-merge main build both succeed. Confirm their repository, commit SHA, steps,
service account, and log links, then verify that Artifact Registry and Cloud Run
did not change. Only after that evidence exists should the exact PR check context
observed on GitHub be added to `main` branch protection. Do not require the main
push trigger.

To roll back only the Cloud Build integration, plan and apply with
`enable_cloud_build_validation=false`. This removes its triggers and dedicated
IAM resources while preserving the enabled API, GitHub App connection, and
GitHub Actions deployment path.
