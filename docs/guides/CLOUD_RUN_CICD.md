# Cloud Run CI/CD and bootstrap guide

This deployment path is manual by design. Local defaults and ordinary CI perform
only syntax, formatting, validation, and mock regression checks. They never apply
Terraform, push images, call GCP APIs, or deploy Cloud Run.

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

OIDC/WIF requires no long-lived service-account JSON secret. Its trust condition
requires the immutable GitHub repository and owner IDs, the `main` ref, and the
exact `.github/workflows/deploy-gcp.yml` workflow on `main`.

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
6. Upsert the six repository variables with `gh variable set`.

Bucket lookup/creation, protection updates, legacy Viewer cleanup, and variable
writes are safe to repeat. Terraform uses the same remote state for both stages;
state and project-specific tfvars are never committed.

## First dev deployment

Open **Actions → Deploy to GCP (Cloud Run) → Run workflow**. Select
`operation=check` first; it is the default and has no OIDC permission or cloud
commands. Then select `operation=deploy` and `environment=dev`. Only deploy and
rollback jobs receive `id-token: write` and authenticate to GCP.

The deployment builds and pushes a commit-tagged image, resolves the immutable
digest, deploys it, checks `/health`, and stores evidence. Only `dev` is exposed
by this workflow. Staging and production stay unavailable until separate GitHub
Environments, approval rules, environment-scoped variables, and isolated GCP
runtime targets are configured and reviewed.

The deployer has project-level Cloud Run administration because service creation
requires it, but Artifact Registry write access is limited to the configured
repository and `iam.serviceAccountUser` is limited to the dedicated runtime
service account. It has no project-level Token Creator role.

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
