#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage: scripts/cd-plan.sh \
  --environment <dev|staging|prod> \
  --project-id <gcp-project-id> \
  --region <gcp-region> \
  --service-name <cloud-run-service> \
  --artifact-repository <artifact-repository>

Optional:
  --image-repository <path>   Override image name prefix (default: service-name)
  --container-path <path>     Docker build context (default: templates/gcloud)
  --output-dir <path>         Directory for generated plan files
  --skip-infra                Skip Terraform plan generation
  --skip-local-checks         Skip scripts/ci-local.sh invocation
  --push-image                Push built image to Artifact Registry
  --skip-image-push           Do not push built image
  --rollback-only             Generate rollback metadata without build or checks
  --help                      Show this help and exit
USAGE
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

ENVIRONMENT="dev"
PROJECT_ID=""
REGION=""
SERVICE_NAME=""
ARTIFACT_REPOSITORY=""
IMAGE_REPOSITORY=""
CONTAINER_PATH="$REPO_ROOT/templates/gcloud"
SKIP_INFRA=false
SKIP_LOCAL_CHECKS=false
PUSH_IMAGE=false
OUTPUT_DIR="$REPO_ROOT/.cd"
ROLLBACK_ONLY=false

if [[ "${GITHUB_ACTIONS:-}" == "true" ]]; then
  PUSH_IMAGE=true
fi

while ((${#})); do
  case "$1" in
    --environment)
      if [[ "${2-}" == "" ]]; then
        echo "Missing value for --environment" >&2
        usage
        exit 2
      fi
      ENVIRONMENT="$2"
      shift 2
      ;;
    --project-id)
      if [[ "${2-}" == "" ]]; then
        echo "Missing value for --project-id" >&2
        usage
        exit 2
      fi
      PROJECT_ID="$2"
      shift 2
      ;;
    --region)
      if [[ "${2-}" == "" ]]; then
        echo "Missing value for --region" >&2
        usage
        exit 2
      fi
      REGION="$2"
      shift 2
      ;;
    --service-name)
      if [[ "${2-}" == "" ]]; then
        echo "Missing value for --service-name" >&2
        usage
        exit 2
      fi
      SERVICE_NAME="$2"
      shift 2
      ;;
    --artifact-repository)
      if [[ "${2-}" == "" ]]; then
        echo "Missing value for --artifact-repository" >&2
        usage
        exit 2
      fi
      ARTIFACT_REPOSITORY="$2"
      shift 2
      ;;
    --image-repository)
      if [[ "${2-}" == "" ]]; then
        echo "Missing value for --image-repository" >&2
        usage
        exit 2
      fi
      IMAGE_REPOSITORY="$2"
      shift 2
      ;;
    --container-path)
      if [[ "${2-}" == "" ]]; then
        echo "Missing value for --container-path" >&2
        usage
        exit 2
      fi
      CONTAINER_PATH="$2"
      shift 2
      ;;
    --output-dir)
      if [[ "${2-}" == "" ]]; then
        echo "Missing value for --output-dir" >&2
        usage
        exit 2
      fi
      OUTPUT_DIR="$2"
      shift 2
      ;;
    --skip-infra)
      SKIP_INFRA=true
      shift
      ;;
    --skip-local-checks)
      SKIP_LOCAL_CHECKS=true
      shift
      ;;
    --push-image)
      PUSH_IMAGE=true
      shift
      ;;
    --skip-image-push)
      PUSH_IMAGE=false
      shift
      ;;
    --rollback-only)
      ROLLBACK_ONLY=true
      PUSH_IMAGE=false
      SKIP_INFRA=true
      SKIP_LOCAL_CHECKS=true
      shift
      ;;
    --help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage
      exit 2
      ;;
  esac
done

if [[ "$PUSH_IMAGE" != true && "$SKIP_INFRA" != true ]]; then
  echo "Skipping Terraform plan because no pushed image digest can be resolved"
  SKIP_INFRA=true
fi

require_value() {
  local value="$1"
  local name="$2"

  if [[ -z "$value" ]]; then
    echo "Missing required value: $name" >&2
    usage
    exit 2
  fi
}

require_cmd() {
  local cmd="$1"
  local step="$2"
  command -v "$cmd" >/dev/null 2>&1 || {
    echo "Error: $step requires '$cmd'" >&2
    exit 3
  }
}

run_cmd() {
  local name="$1"
  shift

  echo "::group::cd-plan:$name"
  if "$@"; then
    echo "::endgroup::"
    return 0
  fi

  echo "::endgroup::"
  echo "cd-plan step failed: $name" >&2
  exit 10
}

resolve_image_digest_uri() {
  local local_image_name="$1"
  local project_id="$2"
  local expected_prefix="${local_image_name%:*}@sha256:"
  local digest_uri=""

  digest_uri="$(gcloud artifacts docker images describe "$local_image_name" --project="$project_id" --format='value(image_summary.fully_qualified_digest)' 2>/dev/null || true)"
  digest_uri="${digest_uri//$'\r'/}"
  digest_uri="${digest_uri//$'\n'/}"
  local digest_value="${digest_uri#"$expected_prefix"}"
  if [[ "$digest_uri" == "$expected_prefix"* && "$digest_value" =~ ^[0-9a-fA-F]{64}$ ]]; then
    echo "$digest_uri"
    return 0
  fi

  return 1
}

require_value "$ENVIRONMENT" "--environment"
require_value "$PROJECT_ID" "--project-id"
require_value "$REGION" "--region"
require_value "$SERVICE_NAME" "--service-name"
require_value "$ARTIFACT_REPOSITORY" "--artifact-repository"

require_cmd git "cd-plan"
if [[ "$ROLLBACK_ONLY" != true ]]; then
  require_cmd docker "cd-plan"
fi

if [[ "$SKIP_LOCAL_CHECKS" != true && "$ROLLBACK_ONLY" != true ]]; then
  require_cmd python3 "cd-plan"
fi

if [[ "$PUSH_IMAGE" == true ]]; then
  require_cmd gcloud "cd-plan"
fi

COMMIT_SHA_FULL="${GITHUB_SHA:-$(git -C "$REPO_ROOT" rev-parse HEAD)}"
COMMIT_SHORT="${COMMIT_SHA_FULL:0:12}"

IMAGE_REPOSITORY="${IMAGE_REPOSITORY:-$SERVICE_NAME}"
FULL_IMAGE_NAME="${REGION}-docker.pkg.dev/${PROJECT_ID}/${ARTIFACT_REPOSITORY}/${IMAGE_REPOSITORY}:${COMMIT_SHORT}"
IMAGE_DIGEST_URI=""

mkdir -p "$OUTPUT_DIR"
PLAN_DIR="$OUTPUT_DIR/${ENVIRONMENT}/${COMMIT_SHORT}"
mkdir -p "$PLAN_DIR"

PLAN_FILE="$PLAN_DIR/cd-plan.env"
PLAN_JSON="$PLAN_DIR/cd-plan.json"
TIMESTAMP="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
INFRA_DIR="$REPO_ROOT/infra/gcp"
INFRA_PLAN_FILE=""

if [[ "$SKIP_LOCAL_CHECKS" != true && "$ROLLBACK_ONLY" != true ]]; then
  bash "$SCRIPT_DIR/ci-local.sh" --workspace "$REPO_ROOT" --skip-infra
fi

IMAGE_ID=""
IMAGE_DIGEST=""
if [[ "$ROLLBACK_ONLY" != true ]]; then
  run_cmd docker-build docker build --platform linux/amd64 -t "$FULL_IMAGE_NAME" "$CONTAINER_PATH"
  IMAGE_ID="$(docker inspect --format '{{.Id}}' "$FULL_IMAGE_NAME")"
fi

if [[ "$PUSH_IMAGE" == true ]]; then
  run_cmd docker-configure-docker gcloud auth configure-docker "${REGION}-docker.pkg.dev" --quiet
  run_cmd docker-push docker push "$FULL_IMAGE_NAME"
  IMAGE_DIGEST_URI="$(resolve_image_digest_uri "$FULL_IMAGE_NAME" "$PROJECT_ID" || true)"
  if [[ -z "$IMAGE_DIGEST_URI" ]]; then
    echo "Artifact Registry digest lookup unavailable for $FULL_IMAGE_NAME; falling back to docker image metadata"
    IMAGE_DIGEST_URI="$(docker inspect --format '{{join .RepoDigests \"\\n\"}}' "$FULL_IMAGE_NAME" 2>/dev/null | sed -n "\\|^${FULL_IMAGE_NAME%:*}@sha256:[0-9a-fA-F]\\{64\\}$|{p;q;}" || true)"
  fi
  if [[ -z "$IMAGE_DIGEST_URI" ]]; then
    echo "Image push completed but no repository digest was available for $FULL_IMAGE_NAME" >&2
    exit 11
  fi
  IMAGE_DIGEST="${IMAGE_DIGEST_URI##*@}"
fi

if [[ "$SKIP_INFRA" == false && "$ROLLBACK_ONLY" != true && -d "$INFRA_DIR" ]]; then
  require_cmd terraform "cd-plan"

  run_cmd terraform-init terraform -chdir="$INFRA_DIR" init -backend=false -input=false
  run_cmd terraform-fmt terraform -chdir="$INFRA_DIR" fmt -check -recursive
  run_cmd terraform-validate terraform -chdir="$INFRA_DIR" validate

  INFRA_PLAN_FILE="$PLAN_DIR/terraform-${ENVIRONMENT}.plan"
  run_cmd terraform-plan terraform -chdir="$INFRA_DIR" plan \
    -input=false \
    -refresh=false \
    -no-color \
    -out="$INFRA_PLAN_FILE" \
    -var "project_id=$PROJECT_ID" \
    -var "region=$REGION" \
    -var "service_name=$SERVICE_NAME" \
    -var "artifact_repository=$ARTIFACT_REPOSITORY" \
    -var "deploy_runtime=true" \
    -var "image_uri=$IMAGE_DIGEST_URI"
  run_cmd terraform-show terraform -chdir="$INFRA_DIR" show -json "$INFRA_PLAN_FILE" > "$PLAN_DIR/terraform-${ENVIRONMENT}.json"
fi

{
  echo "CD_PLAN_ENVIRONMENT=$ENVIRONMENT"
  echo "CD_PLAN_COMMIT_SHA=$COMMIT_SHA_FULL"
  echo "CD_PLAN_COMMIT_SHORT=$COMMIT_SHORT"
  echo "CD_PLAN_PROJECT_ID=$PROJECT_ID"
  echo "CD_PLAN_REGION=$REGION"
  echo "CD_PLAN_SERVICE_NAME=$SERVICE_NAME"
  echo "CD_PLAN_ARTIFACT_REPOSITORY=$ARTIFACT_REPOSITORY"
  echo "CD_PLAN_IMAGE_REPOSITORY=$IMAGE_REPOSITORY"
  echo "CD_PLAN_IMAGE_TAG=$COMMIT_SHORT"
  echo "CD_PLAN_IMAGE_TAG_URI=$FULL_IMAGE_NAME"
  echo "CD_PLAN_IMAGE_URI=$IMAGE_DIGEST_URI"
  echo "CD_PLAN_IMAGE_ID=$IMAGE_ID"
  echo "CD_PLAN_IMAGE_DIGEST=${IMAGE_DIGEST}"
  echo "CD_PLAN_IMAGE_PUSH=$PUSH_IMAGE"
  echo "CD_PLAN_ROLLBACK_ONLY=$ROLLBACK_ONLY"
  echo "CD_PLAN_CONTAINER_PATH=$CONTAINER_PATH"
  echo "CD_PLAN_GENERATED_AT=$TIMESTAMP"
  if [[ "$SKIP_INFRA" == false ]]; then
    echo "CD_PLAN_INFRA_DIR=$INFRA_DIR"
  fi
  if [[ -n "$INFRA_PLAN_FILE" ]]; then
    echo "CD_PLAN_INFRA_PLAN_FILE=$INFRA_PLAN_FILE"
    echo "CD_PLAN_INFRA_JSON=$PLAN_DIR/terraform-${ENVIRONMENT}.json"
  fi
} > "$PLAN_FILE"

cat <<JSON > "$PLAN_JSON"
{
  "environment": "$ENVIRONMENT",
  "commit": "$COMMIT_SHA_FULL",
  "project_id": "$PROJECT_ID",
  "region": "$REGION",
  "service_name": "$SERVICE_NAME",
  "artifact_repository": "$ARTIFACT_REPOSITORY",
  "image_tag_uri": "$FULL_IMAGE_NAME",
  "image_digest_uri": "$IMAGE_DIGEST_URI",
  "image_tag": "$COMMIT_SHORT",
  "image_id": "$IMAGE_ID",
  "image_digest": "$IMAGE_DIGEST",
  "image_push": $PUSH_IMAGE,
  "rollback_only": $ROLLBACK_ONLY,
  "container_path": "$CONTAINER_PATH",
  "generated_at": "$TIMESTAMP",
  "skip_infra": $SKIP_INFRA
}
JSON

echo "Plan written: $PLAN_FILE"
echo "Plan JSON written: $PLAN_JSON"

echo "::group::cd-plan summary"
printf '%s\n' "project=$PROJECT_ID"
printf '%s\n' "region=$REGION"
printf '%s\n' "service=$SERVICE_NAME"
printf '%s\n' "environment=$ENVIRONMENT"
printf '%s\n' "image=$FULL_IMAGE_NAME"
printf '%s\n' "image_push=$PUSH_IMAGE"
printf '%s\n' "skip_infra=$SKIP_INFRA"
if [[ -n "$IMAGE_DIGEST" ]]; then
  printf '%s\n' "digest=$IMAGE_DIGEST"
fi
echo "::endgroup::"

if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
  {
    echo "plan_file=$ENVIRONMENT/$COMMIT_SHORT/cd-plan.env"
    echo "plan_json=$PLAN_JSON"
    echo "environment=$ENVIRONMENT"
    echo "project_id=$PROJECT_ID"
    echo "region=$REGION"
    echo "service_name=$SERVICE_NAME"
    echo "artifact_repository=$ARTIFACT_REPOSITORY"
    echo "image_uri=$IMAGE_DIGEST_URI"
    echo "image_tag_uri=$FULL_IMAGE_NAME"
    echo "image_digest=$IMAGE_DIGEST"
    echo "image_tag=$COMMIT_SHORT"
    echo "image_push=$PUSH_IMAGE"
    echo "commit_sha=$COMMIT_SHA_FULL"
    echo "commit_short=$COMMIT_SHORT"
    if [[ -n "$INFRA_PLAN_FILE" ]]; then
      echo "infra_plan_file=$INFRA_PLAN_FILE"
      echo "infra_plan_json=$PLAN_DIR/terraform-${ENVIRONMENT}.json"
    fi
  } >> "$GITHUB_OUTPUT"
fi

if [[ -n "${GITHUB_STEP_SUMMARY:-}" ]]; then
  {
    echo "### CD plan"
    echo "- environment: $ENVIRONMENT"
    echo "- project: $PROJECT_ID"
    echo "- region: $REGION"
    echo "- service: $SERVICE_NAME"
    echo "- image: $FULL_IMAGE_NAME"
    echo "- image_push: $PUSH_IMAGE"
    if [[ -n "$IMAGE_DIGEST" ]]; then
      echo "- digest: $IMAGE_DIGEST"
    fi
  } >> "$GITHUB_STEP_SUMMARY"
fi

echo "cd-plan complete"
