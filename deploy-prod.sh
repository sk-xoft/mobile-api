#!/usr/bin/env bash
set -Eeuo pipefail

APP_NAME="${APP_NAME:-smart-seaman-mobile-api}"
IMAGE_NAME="${IMAGE_NAME:-smart-seaman-mobile-api:latest}"
HOST_PORT="${HOST_PORT:-30000}"
CONTAINER_PORT="${CONTAINER_PORT:-8080}"

ENV_FILE="${ENV_FILE:-/home/ssmuser/apps/config/mobile-api/prod/.env}"
FIREBASE_FILE="${FIREBASE_FILE:-/home/ssmuser/apps/config/mobile-api/prod/smart-seaman-firebase.json}"
LOG_DIR="${LOG_DIR:-/home/ssmuser/apps/logs/mobile-api}"
FCM_CREDENTIAL_FILE="${FCM_CREDENTIAL_FILE:-/app/firebase.json}"

MVN_CMD="${MVN_CMD:-mvn}"
MVN_ARGS="${MVN_ARGS:-clean package}"

log() {
  printf '[%s] %s\n' "$(date '+%Y-%m-%d %H:%M:%S')" "$*"
}

step() {
  log "[$1/6] $2"
}

done_step() {
  log "[$1/6] DONE: $2"
}

fail() {
  local exit_code=$?
  local line_no=$1

  log "ERROR: Deployment failed at line ${line_no} with exit code ${exit_code}"
  exit "$exit_code"
}

require_file() {
  local file_path="$1"
  local label="$2"

  if [[ ! -f "$file_path" ]]; then
    log "ERROR: ${label} not found: ${file_path}"
    exit 1
  fi
}

trap 'fail $LINENO' ERR

log "Starting deployment"
log "Container name: ${APP_NAME}"
log "Image name: ${IMAGE_NAME}"
log "Port mapping: ${HOST_PORT}:${CONTAINER_PORT}"
log "Env file: ${ENV_FILE}"
log "Firebase file: ${FIREBASE_FILE}"
log "Log directory: ${LOG_DIR}"

step 1 "Validating required files and directories"
require_file "$ENV_FILE" "env file"
require_file "$FIREBASE_FILE" "Firebase credential file"
mkdir -p "$LOG_DIR"
done_step 1 "Required files and log directory are ready"

step 2 "Stopping and removing existing container"
if docker ps -a --format '{{.Names}}' | grep -Fxq "$APP_NAME"; then
  log "Stopping container: ${APP_NAME}"
  docker stop "$APP_NAME" >/dev/null 2>&1 || true

  log "Removing container: ${APP_NAME}"
  docker rm "$APP_NAME" >/dev/null
  done_step 2 "Existing container removed: ${APP_NAME}"
else
  log "Container not found, skip stop/remove: ${APP_NAME}"
  done_step 2 "No existing container to remove"
fi

step 3 "Removing existing Docker image"
if docker image inspect "$IMAGE_NAME" >/dev/null 2>&1; then
  log "Removing image: ${IMAGE_NAME}"
  docker image rm "$IMAGE_NAME" >/dev/null
  done_step 3 "Existing image removed: ${IMAGE_NAME}"
else
  log "Image not found, skip remove: ${IMAGE_NAME}"
  done_step 3 "No existing image to remove"
fi

step 4 "Packaging application with Maven: ${MVN_CMD} ${MVN_ARGS}"
# shellcheck disable=SC2086
"$MVN_CMD" $MVN_ARGS
done_step 4 "Maven package completed"

step 5 "Building Docker image: ${IMAGE_NAME}"
docker build -t "$IMAGE_NAME" .
done_step 5 "Docker image built: ${IMAGE_NAME}"

step 6 "Starting container: ${APP_NAME}"
docker run \
  --name "$APP_NAME" \
  --detach \
  --restart unless-stopped \
  --env-file "$ENV_FILE" \
  --volume "${FIREBASE_FILE}:/app/firebase.json:ro" \
  --volume "${LOG_DIR}:/logs" \
  --env "FCM_CREDENTIAL_FILE=${FCM_CREDENTIAL_FILE}" \
  --publish "${HOST_PORT}:${CONTAINER_PORT}/tcp" \
  "$IMAGE_NAME"
done_step 6 "Container started: ${APP_NAME}"

docker ps --filter "name=${APP_NAME}" --format 'table {{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}'
log "Deployment completed"
