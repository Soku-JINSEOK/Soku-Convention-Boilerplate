# Cloud Run CI/CD and bootstrap guide

This deployment path is manual by design. Local defaults and ordinary CI perform
only syntax, formatting, validation, and mock regression checks. They never apply
Terraform, push images, call GCP APIs, or deploy Cloud Run.

Cloud Build validation is a separate opt-in path. It validates GCP-specific
configuration but never publishes an image or deploys Cloud Run. Production
delivery authority remains exclusively in the manual GitHub Actions workflow.

## Required repository variables

The bootstrap command registers exactly these GitHub Repository Variables:

| Variable | Value |
| --- | --- |
| `GCP_PROJECT_ID` | Supplied project ID |
| `GCP_REGION` | Region, default `asia-northeast1` |
| `GCP_SERVICE_NAME` | Service, default `soku-convention-boilerplate` |
| `GCP_ARTIFACT_REPOSITORY` | Repository, default `cloud-run` |
| `GCP_WIF_PROVIDER` | Full Terraform WIF provider resource name |
| `GCP_WIF_SERVICE_ACCOUNT` | Terraform deployer service-account email |
| `GCP_CI_WIF_PROVIDER` | Main Validation image-builder WIF provider |
| `GCP_CI_WIF_SERVICE_ACCOUNT` | Artifact Registry-only CI builder identity |

OIDC/WIF requires no long-lived service-account JSON secret. Its trust condition
requires the immutable GitHub repository and owner IDs, the `main` ref, and the
exact workflow on `main`. The deployer provider accepts only
`.github/workflows/deploy-gcp.yml`; the CI provider accepts only
`.github/workflows/validation.yml`.

## From project ID to first infrastructure

Authenticate locally with an identity allowed to enable APIs and create GCS,
Artifact Registry, IAM, WIF, and Cloud Run resources. Put the project ID in the
CLI environment and preview first:

```bash
export GCP_PROJECT_ID="<GCP_PROJECT_ID>"
scripts/gcp-bootstrap.sh
```

The preview validates defaults and prints commands without invoking `gcloud`,
`docker`, `terraform`, or `gh`. `GCP_REGION`, `GCP_SERVICE_NAME`, and
`GCP_ARTIFACT_REPOSITORY` may also override their documented defaults. A command
line `--project-id` takes precedence over `GCP_PROJECT_ID`.

Apply only after reviewing and explicitly repeating the exact project ID:

```bash
scripts/gcp-bootstrap.sh \
  --apply \
  --confirm-project-id "$GCP_PROJECT_ID"
```

The apply sequence is:

1. Create `gs://<GCP_PROJECT_ID>-tfstate` if it does not exist, then enforce
   uniform bucket-level access, public access prevention, object versioning,
   and removal of legacy project Viewer read bindings.
2. Resolve immutable GitHub repository and owner IDs and initialize the partial
   GCS backend with prefix `cloud-run`.
3. Apply only the explicit foundation Terraform targets with
   `deploy_runtime=false`; no image is needed and an existing runtime is not
   destroyed on a repeated bootstrap.
4. Build and push the bootstrap image, then resolve its immutable digest.
5. Apply runtime Terraform with `deploy_runtime=true` and the digest URI.
6. Upsert the eight repository variables with `gh variable set`.

Bucket lookup/creation, protection updates, legacy Viewer cleanup, and variable
writes are safe to repeat. Terraform uses the same remote state for both stages;
state and project-specific tfvars are never committed.

To add only the CI image-builder identity to an existing foundation, use
`--ci-builder-only` with the same apply confirmation. That path creates the
provider, service account, repository-scoped Artifact Registry writer binding,
and the two `GCP_CI_*` variables, then exits. It does not build or push an
image, change Cloud Run, access secrets, or publish a package.

## Optional Cloud Build validation

The GCP project must already have its first-generation Cloud Build GitHub App
connection for this repository. Preview the additional API, dedicated service
account, Logs Writer binding, and two triggers:

```bash
scripts/gcp-bootstrap.sh \
  --enable-cloud-build-validation
```

Apply only after reviewing the preview and repeating the project ID:

```bash
scripts/gcp-bootstrap.sh \
  --enable-cloud-build-validation \
  --apply \
  --confirm-project-id "$GCP_PROJECT_ID"
```

Trigger creation fails if the existing first-generation connection is
unavailable. The bootstrap does not create or migrate to a second-generation
connection. Keep the enable flag on subsequent bootstrap applies; the Terraform
variable defaults to `false`, which is also the rollback setting.

The enable flag uses a validation-only apply path. It exits after the targeted
API, identity, IAM, and trigger resources succeed, before Docker authentication,
image build or push, runtime Terraform, and GitHub repository-variable writes.
This makes the integration apply safe to compare against an unchanged Cloud Run
revision.

The validation-only path stores state under the isolated
`cloud-build-validation` GCS prefix. Existing foundation, runtime, and deployment
plans keep using `cloud-run`, so their default `false` value cannot remove the
validation triggers.

The PR trigger, `soku-convention-boilerplate-pr`, validates pull requests whose
target is `main`. External contributors require a repository writer to comment
`/gcbrun`. The `soku-convention-boilerplate-main` trigger validates pushes to
`main`. Both use a dedicated service account with only
`roles/logging.logWriter`, Cloud Logging-only output, and
`cloudbuild/validation.yaml`.

The build runs the existing GCP deployment regression tests and the Cloud Build
policy tests under Node 24, checks Terraform formatting and validity, executes
Terraform mock plans, and builds `templates/gcloud` for `linux/amd64`. It does
not push the resulting image, write to Artifact Registry, access secrets, or run
a deployment command.

Cloud Build does not replace the repository's existing pull-request policy.
Before opening the validation PR, use the repository template, add one `type:*`
and one `area:*` label, assign an owner, and keep a Draft linked with
`Related to #N`. An immediate `PR Metadata Gate` failure that is replaced by a
successful metadata-event run on the same head commit is a metadata violation,
not a GCP incident. Classify it as a CI defect only when the current,
metadata-complete head commit still fails. Cancelled duplicate runs are not
Cloud Build failures.

Treat both checks as informational at first. With a controlled PR and its merged
`main` commit, verify the repository and commit SHA, all build steps, the
dedicated service account, successful status, and GitHub log links. Also compare
Artifact Registry images and Cloud Run revision, image digest, and traffic
before and after, then confirm the authenticated `/health` response remains
`ok`.

After both builds have successful evidence, add only the exact PR check context
shown by GitHub to `main` branch protection. Do not make the main push trigger a
required check. Verify the required flow again with an external pull request and
writer-issued `/gcbrun` approval.

Rollback is an apply with `enable_cloud_build_validation=false`. It removes only
the validation triggers, dedicated identity, and IAM binding. It preserves the
enabled API, GitHub App connection, GitHub Actions deployer, Artifact Registry,
and Cloud Run.

## First dev deployment

Open **Actions → Deploy to GCP (Cloud Run) → Run workflow**. Select
`operation=check` first; it is the default and has no OIDC permission or cloud
commands. A canonical `main` Validation run first builds the `linux/amd64`
image after Quick succeeds, checks `/health`, pushes the full-commit tag,
resolves the registry digest, and uploads `verified-cloud-run-image`. Select
`operation=deploy`, `environment=dev`, and that run's numeric `source_run_id`.
The deploy workflow verifies the source repository, workflow, event, branch,
conclusion, commit, and manifest before accepting its digest. It never builds,
tests, installs dependencies, or pushes an image.

Deployment is operator-bound: only a workflow dispatch from
`refs/heads/main` may receive GCP credentials, and deploy/rollback explicitly
checkout the current protected `main` before authentication. The verified
source SHA identifies the application image and is recorded in the deployment
plan, but historical application source does not supply the credential-bearing
deployment scripts. This keeps operational policy on the current protected
branch while still promoting the exact previously verified digest.

Only deploy and rollback jobs receive `id-token: write` and authenticate to
GCP. Only `dev` is exposed by this workflow. Staging and production stay
unavailable until separate GitHub Environments, approval rules,
environment-scoped variables, and isolated GCP runtime targets are configured
and reviewed.

The deployer has project-level Cloud Run administration because service creation
requires it, and `iam.serviceAccountUser` is limited to the dedicated runtime
service account. The CI builder has only writer access to the configured
Artifact Registry repository and no Cloud Run, Secret Manager, or package
publication role. The deployer's temporary repository writer grant remains
enabled only until a real CI-built digest deployment succeeds; then apply with
`grant_deployer_artifact_writer=false` and reverify deploy and rollback. The
deployer has no project-level Token Creator role. It can
mint an ID token only for itself and has service-scoped Cloud Run Invoker access,
which lets the deployment script authenticate its private `/health` request.
The workflow passes that service-account email explicitly to `cd-deploy.sh`,
which mints an audience-bound ID token through service-account impersonation.
Local callers may omit `--identity-service-account`; the helper then keeps the
active-account token path for backward compatibility. Never enable shell tracing
around this command or persist identity tokens or generated credential paths.

Every deployment and rollback attempt writes a sanitized JSON record under the
non-hidden `deploy-evidence/` directory. The workflow uploads that directory even
when the operation fails and treats a missing evidence file as an operation
failure. Inspect `final_status`, the before/after revisions, rollback target, and
run URL in the downloaded artifact; no token or credential path is recorded.
The deploy helper also assigns 100% traffic to the resolved ready revision and
verifies Cloud Run's reported percentage before calling `/health`. A missing or
non-100% value fails the deployment and restores the exact pre-deploy revision.
Successful evidence records `verified_traffic_percent: 100`.

## Recovery

Run the same workflow with `operation=rollback`. Optionally supply an exact
`rollback_revision`; otherwise the deployment helper selects the previous ready
revision. A failed post-deploy health check automatically sends all traffic back
to the revision recorded immediately before deployment and retains evidence.

Local emergency rollback is also available after generating a rollback plan:

```bash
scripts/cd-plan.sh --environment dev --project-id "$GCP_PROJECT_ID" \
  --region "$GCP_REGION" --service-name "$GCP_SERVICE_NAME" \
  --artifact-repository "$GCP_ARTIFACT_REPOSITORY" --rollback-only

scripts/cd-deploy.sh --plan-file <PLAN_FILE> --rollback-only \
  --rollback-revision <REVISION> --confirm
```
