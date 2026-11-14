#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
APP_NAME="${APP_NAME:-smart-seaman-mobile-api}"
IMAGE_NAME="${IMAGE_NAME:-smart-seaman-mobile-api:non-prod}"
HOST_PORT="${HOST_PORT:-30000}"
CONTAINER_PORT="${CONTAINER_PORT:-8080}"
ENV_FILE="${ENV_FILE:-/home/ssmuser/apps/config/mobile-api/non-prod/.env}"
HEALTH_TIMEOUT="${HEALTH_TIMEOUT:-120}"
LOG_DIR="${LOG_DIR:-}"

log() {
  printf '[%s] %s\n' "$(date '+%Y-%m-%d %H:%M:%S')" "$*"
}

fail() {
  local exit_code=$?
  log "ERROR: Deployment failed at line $1 with exit code ${exit_code}"
  exit "$exit_code"
}

require_env_key() {
  local key="$1"
  # Docker env files are data, not shell scripts. Never source or print them.
  if ! awk -v key="$key" '
    index($0, key "=") == 1 {
      value = substr($0, length(key) + 2)
      sub(/\r$/, "", value)
      found = (value ~ /[^[:space:]]/)
    }
    END { exit !found }
  ' "$ENV_FILE"; then
    log "ERROR: ${key} must have a value in ${ENV_FILE}; use Go keys from .env.example, not legacy Spring keys"
    exit 1
  fi
}

trap 'fail $LINENO' ERR

log "[1/5] Validating non-prod deployment"
command -v docker >/dev/null || { log "ERROR: Docker is required"; exit 1; }
[[ -f "$ENV_FILE" && -r "$ENV_FILE" ]] || { log "ERROR: Env file is missing or unreadable: ${ENV_FILE}"; exit 1; }
# Resolve relative overrides before switching to the repository directory.
ENV_FILE="$(cd -- "$(dirname -- "$ENV_FILE")" && pwd)/$(basename -- "$ENV_FILE")"
# Read only this deployment setting as data; never source the secret env file.
if [[ -z "$LOG_DIR" ]]; then
  LOG_DIR="$(awk '
    index($0, "LOG_DIR=") == 1 {
      value = substr($0, 9)
      sub(/\r$/, "", value)
    }
    END { printf "%s", value }
  ' "$ENV_FILE")"
fi
LOG_DIR="${LOG_DIR:-/home/ssmuser/apps/logs/mobile-api/non-prod}"
if [[ "$LOG_DIR" != /* || "$LOG_DIR" == / || "$LOG_DIR" == *","* || "$LOG_DIR" == *\"* || "$LOG_DIR" == *\'* ]]; then
  log "ERROR: LOG_DIR must be a dedicated absolute directory without quotes or commas"
  exit 1
fi
for port in "$HOST_PORT" "$CONTAINER_PORT"; do
  if [[ ! "$port" =~ ^[0-9]{1,5}$ ]] || (( 10#$port < 1 || 10#$port > 65535 )); then
    log "ERROR: Ports must be integers from 1 to 65535"
    exit 1
  fi
done
if [[ ! "$HEALTH_TIMEOUT" =~ ^[0-9]{1,6}$ ]] || (( 10#$HEALTH_TIMEOUT < 1 )); then
  log "ERROR: HEALTH_TIMEOUT must be a positive integer in seconds (maximum 999999)"
  exit 1
fi
# Normalize leading zeros for Docker and Bash arithmetic.
HOST_PORT=$((10#$HOST_PORT))
CONTAINER_PORT=$((10#$CONTAINER_PORT))
HEALTH_TIMEOUT=$((10#$HEALTH_TIMEOUT))
require_env_key DATABASE_DSN
require_env_key JWT_SECRET
cd -- "$SCRIPT_DIR"
[[ -f Dockerfile && -f go.mod ]] || { log "ERROR: Go Dockerfile/go.mod not found"; exit 1; }
docker info >/dev/null
log "Container: ${APP_NAME}; image: ${IMAGE_NAME}; ports: ${HOST_PORT}:${CONTAINER_PORT}"
log "Env file: ${ENV_FILE}"

log "[2/5] Building Go application with the Docker multi-stage build"
# Build successfully before stopping the running service. Do not delete the old image.
APP_COMMIT="${APP_COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || printf unknown)}"
APP_VERSION="${APP_VERSION:-$APP_COMMIT}"
APP_BUILD_DATE="$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
DEPLOYED_BY="${DEPLOYED_BY:-${SUDO_USER:-$(id -un)}}"
docker build --file "$SCRIPT_DIR/Dockerfile" --tag "$IMAGE_NAME" \
  --build-arg "APP_VERSION=$APP_VERSION" \
  --build-arg "APP_COMMIT=$APP_COMMIT" \
  --build-arg "APP_BUILD_DATE=$APP_BUILD_DATE" "$SCRIPT_DIR"

# Prepare and verify the persistent log directory before stopping the old service.
mkdir -p -- "$LOG_DIR"
LOG_DIR="$(cd -- "$LOG_DIR" && pwd)"
[[ "$LOG_DIR" != *","* ]] || { log "ERROR: LOG_DIR must not contain commas"; exit 1; }
docker run --rm --user 0:0 --entrypoint sh \
  --mount "type=bind,src=${LOG_DIR},dst=/app/logs" "$IMAGE_NAME" \
  -c 'chown 10001:10001 /app/logs && chmod 750 /app/logs'
docker run --rm --entrypoint sh \
  --mount "type=bind,src=${LOG_DIR},dst=/app/logs" "$IMAGE_NAME" \
  -c 'touch /app/logs/.write-check && rm /app/logs/.write-check'
log "Persistent application logs: ${LOG_DIR}/api.log"

log "[3/5] Replacing existing container"
# Listing containers separately avoids treating daemon failures as 'not found'.
containers="$(docker container ls --all --format '{{.Names}}')"
if printf '%s\n' "$containers" | grep -Fxq -- "$APP_NAME"; then
  docker stop -- "$APP_NAME" >/dev/null
  docker rm -- "$APP_NAME" >/dev/null
fi

log "[4/5] Starting non-prod container"
docker run \
  --name "$APP_NAME" \
  --detach \
  --restart unless-stopped \
  --env-file "$ENV_FILE" \
  --env "ENV_FILE=" \
  --env "APP_PROFILE=non-prod" \
  --env "APP_VERSION=$APP_VERSION" \
  --env "APP_COMMIT=$APP_COMMIT" \
  --env "APP_BUILD_DATE=$APP_BUILD_DATE" \
  --env "APP_DEPLOYED_AT=$(date -u '+%Y-%m-%dT%H:%M:%SZ')" \
  --env "APP_DEPLOYED_BY=$DEPLOYED_BY" \
  --env "APP_ADDRESS=:${CONTAINER_PORT}" \
  --mount "type=bind,src=${LOG_DIR},dst=/app/logs" \
  --env "LOG_FILE=/app/logs/api.log" \
  --log-driver json-file \
  --log-opt max-size=10m \
  --log-opt max-file=3 \
  --no-healthcheck \
  --publish "${HOST_PORT}:${CONTAINER_PORT}/tcp" \
  "$IMAGE_NAME"

log "[5/5] Waiting for startup HTTP health (up to ${HEALTH_TIMEOUT}s)"
deadline=$((SECONDS + HEALTH_TIMEOUT))
while (( SECONDS < deadline )); do
  state="$(docker inspect --format '{{.State.Status}}' "$APP_NAME")"
  if [[ "$state" == running ]]; then
    remaining=$((deadline - SECONDS))
    (( remaining > 0 )) || break
    probe_timeout=3
    if (( remaining < probe_timeout )); then
      probe_timeout=$remaining
    fi
    if docker exec "$APP_NAME" wget --quiet --tries=1 --timeout="$probe_timeout" \
      --spider "http://127.0.0.1:${CONTAINER_PORT}/health" >/dev/null 2>&1; then
      log "Deployment completed: ${APP_NAME} startup HTTP check passed; periodic health checks are disabled"
      exit 0
    fi
  elif [[ "$state" != restarting ]]; then
    log "ERROR: Container state=${state}; inspect locally with: docker logs ${APP_NAME}"
    exit 1
  fi
  sleep 2
done
log "ERROR: Startup HTTP health check timed out (state=${state})"
log "Inspect application logs: docker logs --tail 100 ${APP_NAME}"
log "Check HTTP manually: docker exec ${APP_NAME} wget -S -O- http://127.0.0.1:${CONTAINER_PORT}/health"
exit 1
