#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage: scripts/gcp-bootstrap.sh [--project-id <id>] [options]

Project ID:
  --project-id <id>                  Overrides the GCP_PROJECT_ID environment variable

Options:
  --region <region>                  Default: asia-northeast1
  --service <name>                   Default: soku-convention-boilerplate
  --artifact-repository <name>       Default: cloud-run
  --github-repository <owner/repo>   Default: current gh repository (apply only)
  --apply                            Perform cloud and GitHub changes
  --confirm-project-id <id>          Required with --apply; must exactly match
  --help

Without --apply this command only validates and prints the intended commands.
USAGE
}

PROJECT_ID="${GCP_PROJECT_ID:-}"
REGION="${GCP_REGION:-asia-northeast1}"
SERVICE="${GCP_SERVICE_NAME:-soku-convention-boilerplate}"
ARTIFACT_REPOSITORY="${GCP_ARTIFACT_REPOSITORY:-cloud-run}"
GITHUB_REPOSITORY=""
APPLY=false
CONFIRM_PROJECT_ID=""

while (($#)); do
  case "$1" in
    --project-id) PROJECT_ID="${2-}"; shift 2 ;;
    --region) REGION="${2-}"; shift 2 ;;
    --service) SERVICE="${2-}"; shift 2 ;;
    --artifact-repository) ARTIFACT_REPOSITORY="${2-}"; shift 2 ;;
    --github-repository) GITHUB_REPOSITORY="${2-}"; shift 2 ;;
    --apply) APPLY=true; shift ;;
    --confirm-project-id) CONFIRM_PROJECT_ID="${2-}"; shift 2 ;;
    --help) usage; exit 0 ;;
    *) echo "Unknown argument: $1" >&2; usage; exit 2 ;;
  esac
done

if [[ -z "$PROJECT_ID" ]]; then echo "Set GCP_PROJECT_ID or pass --project-id" >&2; exit 2; fi
if [[ ! "$PROJECT_ID" =~ ^[a-z][a-z0-9-]{4,28}[a-z0-9]$ ]]; then echo "Invalid GCP project ID: $PROJECT_ID" >&2; exit 2; fi
if [[ "$APPLY" == true && "$CONFIRM_PROJECT_ID" != "$PROJECT_ID" ]]; then
  echo "--confirm-project-id must exactly match --project-id" >&2
  exit 2
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
INFRA_DIR="$REPO_ROOT/infra/gcp"
STATE_BUCKET="${PROJECT_ID}-tfstate"
IMAGE_TAG="${REGION}-docker.pkg.dev/${PROJECT_ID}/${ARTIFACT_REPOSITORY}/${SERVICE}:bootstrap"

print_summary() {
  printf 'Mode: %s\nProject: %s\nRegion: %s\nService: %s\nArtifact repository: %s\nState bucket: gs://%s\n' \
    "$([[ "$APPLY" == true ]] && echo apply || echo dry-run)" "$PROJECT_ID" "$REGION" "$SERVICE" "$ARTIFACT_REPOSITORY" "$STATE_BUCKET"
}

print_commands() {
  cat <<EOF
gcloud storage buckets describe gs://${STATE_BUCKET} || gcloud storage buckets create gs://${STATE_BUCKET} --project=${PROJECT_ID} --location=${REGION} --uniform-bucket-level-access
terraform -chdir=infra/gcp init -backend-config=bucket=${STATE_BUCKET} -backend-config=prefix=cloud-run
terraform -chdir=infra/gcp apply -var=project_id=${PROJECT_ID} -var=region=${REGION} -var=service_name=${SERVICE} -var=artifact_repository=${ARTIFACT_REPOSITORY} -var=deploy_runtime=false
docker build --platform linux/amd64 -t ${IMAGE_TAG} templates/gcloud
docker push ${IMAGE_TAG}
terraform -chdir=infra/gcp apply ... -var=deploy_runtime=true -var=image_uri=<repository@sha256:digest>
gh variable set GCP_PROJECT_ID/GCP_REGION/GCP_SERVICE_NAME/GCP_ARTIFACT_REPOSITORY/GCP_WIF_PROVIDER/GCP_WIF_SERVICE_ACCOUNT
EOF
}

print_summary
if [[ "$APPLY" != true ]]; then print_commands; exit 0; fi

for command in gcloud terraform docker gh; do command -v "$command" >/dev/null || { echo "Required command not found: $command" >&2; exit 3; }; done
if [[ -z "$GITHUB_REPOSITORY" ]]; then GITHUB_REPOSITORY="$(gh repo view --json nameWithOwner --jq .nameWithOwner)"; fi
if [[ ! "$GITHUB_REPOSITORY" =~ ^[^/]+/[^/]+$ ]]; then echo "Invalid GitHub repository: $GITHUB_REPOSITORY" >&2; exit 2; fi
GITHUB_ORG="${GITHUB_REPOSITORY%%/*}"
GITHUB_REPO="${GITHUB_REPOSITORY##*/}"

if ! gcloud storage buckets describe "gs://${STATE_BUCKET}" --project="$PROJECT_ID" >/dev/null 2>&1; then
  gcloud storage buckets create "gs://${STATE_BUCKET}" --project="$PROJECT_ID" --location="$REGION" --uniform-bucket-level-access
fi
terraform -chdir="$INFRA_DIR" init -reconfigure -input=false -backend-config="bucket=$STATE_BUCKET" -backend-config="prefix=cloud-run"
COMMON_VARS=(-input=false -auto-approve -var="project_id=$PROJECT_ID" -var="region=$REGION" -var="service_name=$SERVICE" -var="artifact_repository=$ARTIFACT_REPOSITORY" -var="github_org=$GITHUB_ORG" -var="github_repo=$GITHUB_REPO")
terraform -chdir="$INFRA_DIR" apply "${COMMON_VARS[@]}" -var="deploy_runtime=false"
gcloud auth configure-docker "${REGION}-docker.pkg.dev" --quiet
docker build --platform linux/amd64 -t "$IMAGE_TAG" "$REPO_ROOT/templates/gcloud"
docker push "$IMAGE_TAG"
IMAGE_URI="$(gcloud artifacts docker images describe "$IMAGE_TAG" --project="$PROJECT_ID" --format='value(image_summary.fully_qualified_digest)')"
if [[ ! "$IMAGE_URI" =~ @sha256:[0-9a-fA-F]{64}$ ]]; then echo "Could not resolve immutable image digest" >&2; exit 4; fi
terraform -chdir="$INFRA_DIR" apply "${COMMON_VARS[@]}" -var="deploy_runtime=true" -var="image_uri=$IMAGE_URI"
WIF_PROVIDER="$(terraform -chdir="$INFRA_DIR" output -raw wif_provider_name)"
WIF_SERVICE_ACCOUNT="$(terraform -chdir="$INFRA_DIR" output -raw deployer_service_account_email)"
for pair in \
  "GCP_PROJECT_ID=$PROJECT_ID" "GCP_REGION=$REGION" "GCP_SERVICE_NAME=$SERVICE" \
  "GCP_ARTIFACT_REPOSITORY=$ARTIFACT_REPOSITORY" "GCP_WIF_PROVIDER=$WIF_PROVIDER" \
  "GCP_WIF_SERVICE_ACCOUNT=$WIF_SERVICE_ACCOUNT"; do
  gh variable set "${pair%%=*}" --body "${pair#*=}" --repo "$GITHUB_REPOSITORY"
done
echo "Bootstrap complete. Run the Deploy to GCP workflow with operation=deploy and environment=dev."
