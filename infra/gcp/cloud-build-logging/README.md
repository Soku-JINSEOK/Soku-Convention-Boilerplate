# Cloud Build logging

This Terraform root owns only the regional validation log bucket, its
project-level sink, and a disabled rollout exclusion. It deliberately has no
billing, IAM, Workload Identity Federation, trigger, or foundation resources.

Use the root's matching backend prefix:

```bash
terraform -chdir=infra/gcp/cloud-build-logging init \
  -backend-config="bucket=${GCP_PROJECT_ID}-tfstate" \
  -backend-config="prefix=cloud-build-logging"
```

Create and inspect a saved plan without applying it:

```bash
terraform -chdir=infra/gcp/cloud-build-logging plan \
  -var="project_id=${GCP_PROJECT_ID}" \
  -out=cloud-build-logging.tfplan
terraform -chdir=infra/gcp/cloud-build-logging show \
  -json cloud-build-logging.tfplan > cloud-build-logging.tfplan.json
python3 scripts/verify-cloud-build-logging-plan.py \
  cloud-build-logging.tfplan.json
```

The accepted plan is exactly three creates and no updates or deletes. Applying
the plan, enabling the exclusion, changing triggers, importing state, or
changing IAM requires a separate approval.
