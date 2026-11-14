#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

usage() {
    cat <<'EOF'
Usage: ./run-docker-dev.sh [up|down|logs|restart|status]

Default: up (build and start in background with Air hot reload).
ENV_FILE: configuration file (default: ../config/mobile-api/non-prod/.env).
HOST_PORT: local API port (default: 8080).
COMPOSE_PROJECT_NAME: isolated Compose project (default: smart-seaman-mobile-local).
EOF
}

ACTION="${1:-up}"
case "$ACTION" in
    -h|--help) usage; exit 0 ;;
    up|down|logs|restart|status) ;;
    *) usage >&2; exit 1 ;;
esac
if [ "$#" -gt 1 ]; then usage >&2; exit 1; fi

command -v docker >/dev/null 2>&1 || { echo "Docker is required." >&2; exit 1; }
if docker compose version >/dev/null 2>&1; then
    COMPOSE=(docker compose)
elif command -v docker-compose >/dev/null 2>&1; then
    COMPOSE=(docker-compose)
else
    echo "Docker Compose is required." >&2
    exit 1
fi
docker info >/dev/null 2>&1 || { echo "Start Docker Desktop or OrbStack first." >&2; exit 1; }

DEV_ENV_FILE="${ENV_FILE:-../config/mobile-api/non-prod/.env}"
if [ ! -r "$DEV_ENV_FILE" ] || [ ! -f "$DEV_ENV_FILE" ]; then
    echo "Environment file is not readable: $DEV_ENV_FILE" >&2
    echo "Prepare Go configuration using .env.example, or set ENV_FILE=.env." >&2
    exit 1
fi
DEV_ENV_FILE="$(cd -- "$(dirname -- "$DEV_ENV_FILE")" && pwd)/$(basename -- "$DEV_ENV_FILE")"
export DEV_ENV_FILE
export HOST_PORT="${HOST_PORT:-8080}"
export COMPOSE_PROJECT_NAME="${COMPOSE_PROJECT_NAME:-smart-seaman-mobile-local}"
# The selected file is injected by Compose; do not load a host path in the container.
unset ENV_FILE

case "$HOST_PORT" in
    ''|*[!0-9]*) echo "HOST_PORT must be a number from 1 to 65535." >&2; exit 1 ;;
esac
if [ "${#HOST_PORT}" -gt 5 ] || [ "$HOST_PORT" -lt 1 ] || [ "$HOST_PORT" -gt 65535 ]; then
    echo "HOST_PORT must be a number from 1 to 65535." >&2
    exit 1
fi

COMPOSE+=(-f "$SCRIPT_DIR/docker-compose.dev.yml")
case "$ACTION" in
    up)
        "${COMPOSE[@]}" up --build -d
        echo "Local dev started: http://localhost:$HOST_PORT (Air hot reload)."
        echo "Logs: ./run-docker-dev.sh logs"
        ;;
    down) "${COMPOSE[@]}" down ;;
    logs) "${COMPOSE[@]}" logs -f --tail=100 api ;;
    restart) "${COMPOSE[@]}" up --build -d --force-recreate api ;;
    status) "${COMPOSE[@]}" ps ;;
esac
