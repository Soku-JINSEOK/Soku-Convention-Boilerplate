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
  --enable-cloud-build-validation    Create validation-only PR/main triggers
  --ci-builder-only                  Apply only the main Validation image-builder
                                     identity and set its two repository variables
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
ENABLE_CLOUD_BUILD_VALIDATION=false
CI_BUILDER_ONLY=false
APPLY=false
CONFIRM_PROJECT_ID=""

while (($#)); do
  case "$1" in
    --project-id) PROJECT_ID="${2-}"; shift 2 ;;
    --region) REGION="${2-}"; shift 2 ;;
    --service) SERVICE="${2-}"; shift 2 ;;
    --artifact-repository) ARTIFACT_REPOSITORY="${2-}"; shift 2 ;;
    --github-repository) GITHUB_REPOSITORY="${2-}"; shift 2 ;;
    --enable-cloud-build-validation) ENABLE_CLOUD_BUILD_VALIDATION=true; shift ;;
    --ci-builder-only) CI_BUILDER_ONLY=true; shift ;;
    --apply) APPLY=true; shift ;;
    --confirm-project-id) CONFIRM_PROJECT_ID="${2-}"; shift 2 ;;
    --help) usage; exit 0 ;;
    *) echo "Unknown argument: $1" >&2; usage; exit 2 ;;
  esac
done

if [[ "$ENABLE_CLOUD_BUILD_VALIDATION" == true && "$CI_BUILDER_ONLY" == true ]]; then
  echo "--enable-cloud-build-validation and --ci-builder-only are mutually exclusive" >&2
  exit 2
fi

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
STATE_PREFIX="$([[ "$ENABLE_CLOUD_BUILD_VALIDATION" == true ]] && echo cloud-build-validation || echo cloud-run)"
IMAGE_TAG="${REGION}-docker.pkg.dev/${PROJECT_ID}/${ARTIFACT_REPOSITORY}/${SERVICE}:bootstrap"

print_summary() {
  printf 'Mode: %s\nProject: %s\nRegion: %s\nService: %s\nArtifact repository: %s\nState bucket: gs://%s\nState prefix: %s\nCloud Build validation: %s\nCI builder only: %s\n' \
    "$([[ "$APPLY" == true ]] && echo apply || echo dry-run)" "$PROJECT_ID" "$REGION" "$SERVICE" "$ARTIFACT_REPOSITORY" "$STATE_BUCKET" \
    "$STATE_PREFIX" "$([[ "$ENABLE_CLOUD_BUILD_VALIDATION" == true ]] && echo enabled || echo disabled)" \
    "$([[ "$CI_BUILDER_ONLY" == true ]] && echo enabled || echo disabled)"
}

print_commands() {
  if [[ "$CI_BUILDER_ONLY" == true ]]; then
    cat <<EOF
gcloud storage buckets describe gs://${STATE_BUCKET} || gcloud storage buckets create gs://${STATE_BUCKET} --project=${PROJECT_ID} --location=${REGION} --uniform-bucket-level-access
gcloud storage buckets update gs://${STATE_BUCKET} --uniform-bucket-level-access --public-access-prevention --versioning
gh api repos/<owner>/<repo> --jq .id
gh api repos/<owner>/<repo> --jq .owner.id
terraform -chdir=infra/gcp init -backend-config=bucket=${STATE_BUCKET} -backend-config=prefix=${STATE_PREFIX}
terraform -chdir=infra/gcp apply <ci-builder-identity-targets> -var=deploy_runtime=false
gh variable set GCP_CI_WIF_PROVIDER/GCP_CI_WIF_SERVICE_ACCOUNT
# No image build, image push, Cloud Run change, or package publication occurs.
EOF
    return
  fi
  if [[ "$ENABLE_CLOUD_BUILD_VALIDATION" == true ]]; then
    cat <<EOF
gcloud storage buckets describe gs://${STATE_BUCKET} || gcloud storage buckets create gs://${STATE_BUCKET} --project=${PROJECT_ID} --location=${REGION} --uniform-bucket-level-access
gcloud storage buckets update gs://${STATE_BUCKET} --uniform-bucket-level-access --public-access-prevention --versioning
gh api repos/<owner>/<repo> --jq .id
gh api repos/<owner>/<repo> --jq .owner.id
terraform -chdir=infra/gcp init -backend-config=bucket=${STATE_BUCKET} -backend-config=prefix=${STATE_PREFIX}
# API: cloudbuild.googleapis.com
terraform -chdir=infra/gcp apply -target=google_project_service.cloud_build -target=google_service_account.cloud_build_validation -target=google_project_iam_member.cloud_build_validation_log_writer -target=google_cloudbuild_trigger.pull_request -target=google_cloudbuild_trigger.main -var=enable_cloud_build_validation=true
# Trigger creation verifies the existing first-generation GitHub App connection and fails when it is unavailable; no second-generation connection is created.
EOF
    return
  fi
  cat <<EOF
gcloud storage buckets describe gs://${STATE_BUCKET} || gcloud storage buckets create gs://${STATE_BUCKET} --project=${PROJECT_ID} --location=${REGION} --uniform-bucket-level-access
gcloud storage buckets update gs://${STATE_BUCKET} --uniform-bucket-level-access --public-access-prevention --versioning
gcloud storage buckets remove-iam-policy-binding gs://${STATE_BUCKET} --member=projectViewer:${PROJECT_ID} --role=<legacy-reader-role>
gh api repos/<owner>/<repo> --jq .id
gh api repos/<owner>/<repo> --jq .owner.id
terraform -chdir=infra/gcp init -backend-config=bucket=${STATE_BUCKET} -backend-config=prefix=${STATE_PREFIX}
terraform -chdir=infra/gcp apply <foundation-targets> -var=project_id=${PROJECT_ID} -var=region=${REGION} -var=service_name=${SERVICE} -var=artifact_repository=${ARTIFACT_REPOSITORY} -var=github_repository_id=<id> -var=github_repository_owner_id=<id> -var=deploy_runtime=false
docker build --platform linux/amd64 -t ${IMAGE_TAG} templates/gcloud
docker push ${IMAGE_TAG}
terraform -chdir=infra/gcp apply ... -var=deploy_runtime=true -var=image_uri=<repository@sha256:digest>
gh variable set GCP_PROJECT_ID/GCP_REGION/GCP_SERVICE_NAME/GCP_ARTIFACT_REPOSITORY/GCP_WIF_PROVIDER/GCP_WIF_SERVICE_ACCOUNT
EOF
}

print_summary
if [[ "$APPLY" != true ]]; then print_commands; exit 0; fi

for command in gcloud terraform gh jq; do command -v "$command" >/dev/null || { echo "Required command not found: $command" >&2; exit 3; }; done
if [[ "$ENABLE_CLOUD_BUILD_VALIDATION" != true && "$CI_BUILDER_ONLY" != true ]]; then
  command -v docker >/dev/null || { echo "Required command not found: docker" >&2; exit 3; }
fi
if [[ -z "$GITHUB_REPOSITORY" ]]; then GITHUB_REPOSITORY="$(gh repo view --json nameWithOwner --jq .nameWithOwner)"; fi
if [[ ! "$GITHUB_REPOSITORY" =~ ^[^/]+/[^/]+$ ]]; then echo "Invalid GitHub repository: $GITHUB_REPOSITORY" >&2; exit 2; fi
GITHUB_ORG="${GITHUB_REPOSITORY%%/*}"
GITHUB_REPO="${GITHUB_REPOSITORY##*/}"
GITHUB_REPOSITORY_ID="$(gh api "repos/$GITHUB_REPOSITORY" --jq '.id')"
GITHUB_REPOSITORY_OWNER_ID="$(gh api "repos/$GITHUB_REPOSITORY" --jq '.owner.id')"
if [[ ! "$GITHUB_REPOSITORY_ID" =~ ^[0-9]+$ || ! "$GITHUB_REPOSITORY_OWNER_ID" =~ ^[0-9]+$ ]]; then
  echo "Could not resolve immutable GitHub repository and owner IDs" >&2
  exit 4
fi

if ! gcloud storage buckets describe "gs://${STATE_BUCKET}" --project="$PROJECT_ID" >/dev/null 2>&1; then
  gcloud storage buckets create "gs://${STATE_BUCKET}" --project="$PROJECT_ID" --location="$REGION" --uniform-bucket-level-access
fi
gcloud storage buckets update "gs://${STATE_BUCKET}" \
  --uniform-bucket-level-access --public-access-prevention --versioning
for role in roles/storage.legacyBucketReader roles/storage.legacyObjectReader; do
  if gcloud storage buckets get-iam-policy "gs://${STATE_BUCKET}" --format=json | \
    jq -e --arg role "$role" --arg member "projectViewer:$PROJECT_ID" \
      '.bindings[]? | select(.role == $role and ((.members // []) | index($member)))' \
      >/dev/null; then
    gcloud storage buckets remove-iam-policy-binding "gs://${STATE_BUCKET}" \
      --member="projectViewer:$PROJECT_ID" --role="$role"
  fi
done
terraform -chdir="$INFRA_DIR" init -reconfigure -input=false -backend-config="bucket=$STATE_BUCKET" -backend-config="prefix=$STATE_PREFIX"
COMMON_VARS=(
  -input=false
  -auto-approve
  -var="project_id=$PROJECT_ID"
  -var="region=$REGION"
  -var="service_name=$SERVICE"
  -var="artifact_repository=$ARTIFACT_REPOSITORY"
  -var="github_org=$GITHUB_ORG"
  -var="github_repo=$GITHUB_REPO"
  -var="github_repository_id=$GITHUB_REPOSITORY_ID"
  -var="github_repository_owner_id=$GITHUB_REPOSITORY_OWNER_ID"
  -var="enable_cloud_build_validation=$ENABLE_CLOUD_BUILD_VALIDATION"
)
if [[ "$ENABLE_CLOUD_BUILD_VALIDATION" == true ]]; then
  VALIDATION_TARGETS=(
    -target=google_project_service.cloud_build
    -target=google_service_account.cloud_build_validation
    -target=google_project_iam_member.cloud_build_validation_log_writer
    -target=google_cloudbuild_trigger.pull_request
    -target=google_cloudbuild_trigger.main
  )
  echo "Verifying the existing first-generation Cloud Build GitHub App connection while creating validation triggers."
  terraform -chdir="$INFRA_DIR" apply "${COMMON_VARS[@]}" "${VALIDATION_TARGETS[@]}" -var="deploy_runtime=false"
  echo "Cloud Build validation enabled without building, publishing, or deploying an image."
  exit 0
fi

if [[ "$CI_BUILDER_ONLY" == true ]]; then
  CI_BUILDER_TARGETS=(
    -target=google_project_service.required_apis
    -target=google_artifact_registry_repository.repository
    -target=google_service_account.github_actions_ci_builder
    -target=google_artifact_registry_repository_iam_member.ci_builder_repository_writer
    -target=google_iam_workload_identity_pool.github
    -target=google_iam_workload_identity_pool_provider.github_ci
    -target=google_service_account_iam_member.github_ci_builder_wi
  )
  terraform -chdir="$INFRA_DIR" apply "${COMMON_VARS[@]}" "${CI_BUILDER_TARGETS[@]}" -var="deploy_runtime=false"
  CI_WIF_PROVIDER="$(terraform -chdir="$INFRA_DIR" output -raw wif_ci_provider_name)"
  CI_WIF_SERVICE_ACCOUNT="$(terraform -chdir="$INFRA_DIR" output -raw ci_builder_service_account_email)"
  for pair in \
    "GCP_CI_WIF_PROVIDER=$CI_WIF_PROVIDER" \
    "GCP_CI_WIF_SERVICE_ACCOUNT=$CI_WIF_SERVICE_ACCOUNT"; do
    gh variable set "${pair%%=*}" --body "${pair#*=}" --repo "$GITHUB_REPOSITORY"
  done
  echo "CI builder identity applied without building, publishing, or deploying an image."
  exit 0
fi

FOUNDATION_TARGETS=(
  -target=google_project_service.required_apis
  -target=google_artifact_registry_repository.repository
  -target=google_service_account.cloud_run_runtime
  -target=google_service_account.github_actions_deployer
  -target=google_service_account.github_actions_ci_builder
  -target=google_project_iam_member.deployer_run_admin
  -target=google_service_account_iam_member.deployer_runtime_user
  -target=google_service_account_iam_member.deployer_self_token_creator
  -target=google_artifact_registry_repository_iam_member.deployer_repository_writer
  -target=google_artifact_registry_repository_iam_member.ci_builder_repository_writer
  -target=google_iam_workload_identity_pool.github
  -target=google_iam_workload_identity_pool_provider.github
  -target=google_iam_workload_identity_pool_provider.github_ci
  -target=google_service_account_iam_member.github_deployer_wi
  -target=google_service_account_iam_member.github_ci_builder_wi
)
terraform -chdir="$INFRA_DIR" apply "${COMMON_VARS[@]}" "${FOUNDATION_TARGETS[@]}" -var="deploy_runtime=false"
gcloud auth configure-docker "${REGION}-docker.pkg.dev" --quiet
docker build --platform linux/amd64 -t "$IMAGE_TAG" "$REPO_ROOT/templates/gcloud"
docker push "$IMAGE_TAG"
IMAGE_URI="$(gcloud artifacts docker images describe "$IMAGE_TAG" --project="$PROJECT_ID" --format='value(image_summary.fully_qualified_digest)')"
if [[ ! "$IMAGE_URI" =~ @sha256:[0-9a-fA-F]{64}$ ]]; then echo "Could not resolve immutable image digest" >&2; exit 4; fi
terraform -chdir="$INFRA_DIR" apply "${COMMON_VARS[@]}" -var="deploy_runtime=true" -var="image_uri=$IMAGE_URI"
WIF_PROVIDER="$(terraform -chdir="$INFRA_DIR" output -raw wif_provider_name)"
WIF_SERVICE_ACCOUNT="$(terraform -chdir="$INFRA_DIR" output -raw deployer_service_account_email)"
CI_WIF_PROVIDER="$(terraform -chdir="$INFRA_DIR" output -raw wif_ci_provider_name)"
CI_WIF_SERVICE_ACCOUNT="$(terraform -chdir="$INFRA_DIR" output -raw ci_builder_service_account_email)"
for pair in \
  "GCP_PROJECT_ID=$PROJECT_ID" "GCP_REGION=$REGION" "GCP_SERVICE_NAME=$SERVICE" \
  "GCP_ARTIFACT_REPOSITORY=$ARTIFACT_REPOSITORY" "GCP_WIF_PROVIDER=$WIF_PROVIDER" \
  "GCP_WIF_SERVICE_ACCOUNT=$WIF_SERVICE_ACCOUNT" "GCP_CI_WIF_PROVIDER=$CI_WIF_PROVIDER" \
  "GCP_CI_WIF_SERVICE_ACCOUNT=$CI_WIF_SERVICE_ACCOUNT"; do
  gh variable set "${pair%%=*}" --body "${pair#*=}" --repo "$GITHUB_REPOSITORY"
done
echo "Bootstrap complete. Run the Deploy to GCP workflow with operation=deploy and environment=dev."
