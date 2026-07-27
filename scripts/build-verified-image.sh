#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage: scripts/build-verified-image.sh \
  --project-id <id> --region <region> --artifact-repository <name> \
  --image-repository <name> --source-sha <40-hex> \
  --source-ref <ref> --repository <owner/name> --workflow-run-id <id> \
  --manifest <path>
USAGE
}

PROJECT_ID=""
REGION=""
ARTIFACT_REPOSITORY=""
IMAGE_REPOSITORY=""
SOURCE_SHA=""
SOURCE_REF=""
REPOSITORY=""
WORKFLOW_RUN_ID=""
MANIFEST=""

while ((${#})); do
  case "$1" in
    --project-id) PROJECT_ID="${2-}"; shift 2 ;;
    --region) REGION="${2-}"; shift 2 ;;
    --artifact-repository) ARTIFACT_REPOSITORY="${2-}"; shift 2 ;;
    --image-repository) IMAGE_REPOSITORY="${2-}"; shift 2 ;;
    --source-sha) SOURCE_SHA="${2-}"; shift 2 ;;
    --source-ref) SOURCE_REF="${2-}"; shift 2 ;;
    --repository) REPOSITORY="${2-}"; shift 2 ;;
    --workflow-run-id) WORKFLOW_RUN_ID="${2-}"; shift 2 ;;
    --manifest) MANIFEST="${2-}"; shift 2 ;;
    --help) usage; exit 0 ;;
    *) echo "Unknown argument: $1" >&2; usage; exit 2 ;;
  esac
done

for value in PROJECT_ID REGION ARTIFACT_REPOSITORY IMAGE_REPOSITORY SOURCE_SHA SOURCE_REF REPOSITORY WORKFLOW_RUN_ID MANIFEST; do
  if [[ -z "${!value}" ]]; then
    echo "Missing required value: $value" >&2
    exit 2
  fi
done
if [[ ! "$SOURCE_SHA" =~ ^[0-9a-f]{40}$ ]]; then
  echo "--source-sha must be a full lowercase commit SHA" >&2
  exit 2
fi
if [[ ! "$REPOSITORY" =~ ^[^/]+/[^/]+$ || ! "$WORKFLOW_RUN_ID" =~ ^[0-9]+$ ]]; then
  echo "Invalid repository or workflow run ID" >&2
  exit 2
fi

for command in docker gcloud curl node; do
  command -v "$command" >/dev/null 2>&1 || {
    echo "Required command not found: $command" >&2
    exit 3
  }
done

IMAGE_TAG_URI="${REGION}-docker.pkg.dev/${PROJECT_ID}/${ARTIFACT_REPOSITORY}/${IMAGE_REPOSITORY}:${SOURCE_SHA}"
CONTAINER_ID=""
cleanup() {
  if [[ -n "$CONTAINER_ID" ]]; then
    docker rm -f "$CONTAINER_ID" >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

docker build --platform linux/amd64 -t "$IMAGE_TAG_URI" templates/gcloud
CONTAINER_ID="$(docker run --detach --publish 127.0.0.1::8080 "$IMAGE_TAG_URI")"
HOST_PORT="$(docker port "$CONTAINER_ID" 8080/tcp | sed -n 's/.*://p' | head -1)"
if [[ ! "$HOST_PORT" =~ ^[0-9]+$ ]]; then
  echo "Could not resolve local smoke-test port" >&2
  exit 4
fi
for attempt in {1..15}; do
  if curl -fsS --max-time 2 "http://127.0.0.1:${HOST_PORT}/health" >/dev/null; then
    break
  fi
  if [[ "$attempt" == 15 ]]; then
    echo "Container /health smoke test failed" >&2
    exit 4
  fi
  sleep 1
done
docker rm -f "$CONTAINER_ID" >/dev/null
CONTAINER_ID=""

gcloud auth configure-docker "${REGION}-docker.pkg.dev" --quiet
docker push "$IMAGE_TAG_URI"
IMAGE_DIGEST_URI="$(
  gcloud artifacts docker images describe "$IMAGE_TAG_URI" \
    --project="$PROJECT_ID" \
    --format='value(image_summary.fully_qualified_digest)'
)"
EXPECTED_PREFIX="${IMAGE_TAG_URI%:*}@sha256:"
if [[ "$IMAGE_DIGEST_URI" != "$EXPECTED_PREFIX"* ]]; then
  echo "Registry returned a digest outside the expected repository" >&2
  exit 5
fi
IMAGE_DIGEST="${IMAGE_DIGEST_URI##*@}"
if [[ ! "$IMAGE_DIGEST" =~ ^sha256:[0-9a-f]{64}$ ]]; then
  echo "Registry did not return a valid sha256 digest" >&2
  exit 5
fi

mkdir -p "$(dirname "$MANIFEST")"
# shellcheck disable=SC2016 # JavaScript template literals expand in Node.
node -e '
  const {writeFileSync} = require("node:fs");
  const [
    path, repository, sourceSha, sourceRef, workflowRunId, imageUri, digest,
  ] = process.argv.slice(1);
  writeFileSync(path, `${JSON.stringify({
    schemaVersion: 1,
    repository,
    sourceSha,
    sourceRef,
    workflowRunId,
    imageUri,
    digest,
    createdAt: new Date().toISOString(),
  }, null, 2)}\n`);
' "$MANIFEST" "$REPOSITORY" "$SOURCE_SHA" "$SOURCE_REF" \
  "$WORKFLOW_RUN_ID" "$IMAGE_DIGEST_URI" "$IMAGE_DIGEST"

echo "Verified image manifest written: $MANIFEST"
