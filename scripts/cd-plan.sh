#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage: scripts/cd-plan.sh \
  --environment <dev|staging|prod> \
  --project-id <gcp-project-id> \
  --region <gcp-region> \
  --service-name <cloud-run-service> \
  --artifact-repository <artifact-repository> \
  --image-uri <repository@sha256:digest> \
  --source-commit <40-hex-sha>

Options:
  --output-dir <path>         Directory for generated plan files
  --rollback-only             Generate rollback metadata without an image
  --help                      Show this help and exit

This command only converts a previously verified digest into deployment plan
metadata. Image creation, publication, dependency checks, and infrastructure
changes belong to validation and bootstrap workflows.
USAGE
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
ENVIRONMENT="dev"
PROJECT_ID=""
REGION=""
SERVICE_NAME=""
ARTIFACT_REPOSITORY=""
IMAGE_URI=""
SOURCE_COMMIT=""
OUTPUT_DIR="$REPO_ROOT/.cd"
ROLLBACK_ONLY=false

while ((${#})); do
  case "$1" in
    --environment) ENVIRONMENT="${2-}"; shift 2 ;;
    --project-id) PROJECT_ID="${2-}"; shift 2 ;;
    --region) REGION="${2-}"; shift 2 ;;
    --service-name) SERVICE_NAME="${2-}"; shift 2 ;;
    --artifact-repository) ARTIFACT_REPOSITORY="${2-}"; shift 2 ;;
    --image-uri) IMAGE_URI="${2-}"; shift 2 ;;
    --source-commit) SOURCE_COMMIT="${2-}"; shift 2 ;;
    --output-dir) OUTPUT_DIR="${2-}"; shift 2 ;;
    --rollback-only) ROLLBACK_ONLY=true; shift ;;
    --help) usage; exit 0 ;;
    *) echo "Unknown argument: $1" >&2; usage; exit 2 ;;
  esac
done

for value in ENVIRONMENT PROJECT_ID REGION SERVICE_NAME ARTIFACT_REPOSITORY; do
  if [[ -z "${!value}" ]]; then
    echo "Missing required value: $value" >&2
    exit 2
  fi
done
if [[ ! "$ENVIRONMENT" =~ ^(dev|staging|prod)$ ]]; then
  echo "Unsupported environment: $ENVIRONMENT" >&2
  exit 2
fi

if [[ "$ROLLBACK_ONLY" == true ]]; then
  if [[ -z "$SOURCE_COMMIT" ]]; then
    SOURCE_COMMIT="${GITHUB_SHA:-$(git -C "$REPO_ROOT" rev-parse HEAD)}"
  fi
  IMAGE_URI=""
else
  if [[ -z "$IMAGE_URI" || -z "$SOURCE_COMMIT" ]]; then
    echo "--image-uri and --source-commit are required for deployment plans" >&2
    exit 2
  fi
  expected_prefix="${REGION}-docker.pkg.dev/${PROJECT_ID}/${ARTIFACT_REPOSITORY}/${SERVICE_NAME}@sha256:"
  if [[ "$IMAGE_URI" != "$expected_prefix"* ||
    ! "$IMAGE_URI" =~ @sha256:[0-9a-f]{64}$ ]]; then
    echo "--image-uri must be an immutable digest in the configured repository" >&2
    exit 2
  fi
fi
if [[ ! "$SOURCE_COMMIT" =~ ^[0-9a-f]{40}$ ]]; then
  echo "--source-commit must be a full lowercase commit SHA" >&2
  exit 2
fi

COMMIT_SHORT="${SOURCE_COMMIT:0:12}"
PLAN_DIR="$OUTPUT_DIR/${ENVIRONMENT}/${COMMIT_SHORT}"
PLAN_FILE="$PLAN_DIR/cd-plan.env"
PLAN_JSON="$PLAN_DIR/cd-plan.json"
TIMESTAMP="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
IMAGE_DIGEST="${IMAGE_URI##*@}"
mkdir -p "$PLAN_DIR"

{
  echo "CD_PLAN_ENVIRONMENT=$ENVIRONMENT"
  echo "CD_PLAN_COMMIT_SHA=$SOURCE_COMMIT"
  echo "CD_PLAN_COMMIT_SHORT=$COMMIT_SHORT"
  echo "CD_PLAN_PROJECT_ID=$PROJECT_ID"
  echo "CD_PLAN_REGION=$REGION"
  echo "CD_PLAN_SERVICE_NAME=$SERVICE_NAME"
  echo "CD_PLAN_ARTIFACT_REPOSITORY=$ARTIFACT_REPOSITORY"
  echo "CD_PLAN_IMAGE_URI=$IMAGE_URI"
  echo "CD_PLAN_IMAGE_DIGEST=$IMAGE_DIGEST"
  echo "CD_PLAN_ROLLBACK_ONLY=$ROLLBACK_ONLY"
  echo "CD_PLAN_GENERATED_AT=$TIMESTAMP"
} > "$PLAN_FILE"

# shellcheck disable=SC2016 # JavaScript template literals expand in Node.
node -e '
  const {writeFileSync} = require("node:fs");
  const [
    path, environment, commit, projectId, region, serviceName,
    artifactRepository, imageUri, imageDigest, rollbackOnly, generatedAt,
  ] = process.argv.slice(1);
  writeFileSync(path, `${JSON.stringify({
    environment,
    commit,
    project_id: projectId,
    region,
    service_name: serviceName,
    artifact_repository: artifactRepository,
    image_digest_uri: imageUri,
    image_digest: imageDigest,
    rollback_only: rollbackOnly === "true",
    generated_at: generatedAt,
  }, null, 2)}\n`);
' "$PLAN_JSON" "$ENVIRONMENT" "$SOURCE_COMMIT" "$PROJECT_ID" "$REGION" \
  "$SERVICE_NAME" "$ARTIFACT_REPOSITORY" "$IMAGE_URI" "$IMAGE_DIGEST" \
  "$ROLLBACK_ONLY" "$TIMESTAMP"

echo "Plan written: $PLAN_FILE"
echo "Plan JSON written: $PLAN_JSON"
if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
  {
    echo "plan_file=$ENVIRONMENT/$COMMIT_SHORT/cd-plan.env"
    echo "plan_json=$PLAN_JSON"
    echo "image_uri=$IMAGE_URI"
    echo "image_digest=$IMAGE_DIGEST"
    echo "commit_sha=$SOURCE_COMMIT"
  } >> "$GITHUB_OUTPUT"
fi
