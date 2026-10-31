#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
APP_NAME="${APP_NAME:-smart-seaman-mobile-api}"
IMAGE_NAME="${IMAGE_NAME:-smart-seaman-mobile-api:non-prod}"
HOST_PORT="${HOST_PORT:-30000}"
CONTAINER_PORT="${CONTAINER_PORT:-8080}"
ENV_FILE="${ENV_FILE:-/home/ssmuser/apps/config/mobile-api/non-prod/.env}"
HEALTH_TIMEOUT="${HEALTH_TIMEOUT:-120}"

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
docker build --file "$SCRIPT_DIR/Dockerfile" --tag "$IMAGE_NAME" "$SCRIPT_DIR"

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
  --env "APP_ADDRESS=:${CONTAINER_PORT}" \
  --log-driver json-file \
  --log-opt max-size=10m \
  --log-opt max-file=3 \
  --health-cmd "wget --no-verbose --tries=1 --spider http://127.0.0.1:${CONTAINER_PORT}/health || exit 1" \
  --health-interval 5s \
  --health-timeout 3s \
  --health-start-period 10s \
  --health-retries 3 \
  --publish "${HOST_PORT}:${CONTAINER_PORT}/tcp" \
  "$IMAGE_NAME"

log "[5/5] Waiting for container health (up to ${HEALTH_TIMEOUT}s)"
deadline=$((SECONDS + HEALTH_TIMEOUT))
while (( SECONDS < deadline )); do
  state="$(docker inspect --format '{{.State.Status}}' "$APP_NAME")"
  health="$(docker inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}missing{{end}}' "$APP_NAME")"
  if [[ "$state" == running && "$health" == healthy ]]; then
    log "Deployment completed: ${APP_NAME} is healthy"
    exit 0
  fi
  if [[ "$state" != running || "$health" == unhealthy || "$health" == missing ]]; then
    log "ERROR: Container state=${state}, health=${health}; inspect locally with: docker logs ${APP_NAME}"
    exit 1
  fi
  sleep 2
done
log "ERROR: Health check timed out; inspect locally with: docker logs ${APP_NAME}"
exit 1
