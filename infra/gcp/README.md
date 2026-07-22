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

The GCS backend is partial configuration. Initialize it with the project-derived
bucket rather than committing backend values or state:

```bash
terraform -chdir=infra/gcp init \
  -backend-config="bucket=${GCP_PROJECT_ID}-tfstate" \
  -backend-config="prefix=cloud-run"
```

Set `GCP_PROJECT_ID=<id>` and run `scripts/gcp-bootstrap.sh` to preview the full
sequence. `--project-id` is also supported and takes precedence over the
environment. Actual creation additionally requires
`--apply --confirm-project-id <id>`.

During apply, the bootstrap resolves immutable GitHub repository and owner IDs.
The WIF provider accepts only the configured repository IDs, `refs/heads/main`,
and `.github/workflows/deploy-gcp.yml` from `main`. The state bucket is hardened
with uniform access, enforced public-access prevention, object versioning, and
no legacy project Viewer object access.

The outputs `wif_provider_name` and `deployer_service_account_email` map directly
to the GitHub repository variables `GCP_WIF_PROVIDER` and
`GCP_WIF_SERVICE_ACCOUNT`.
