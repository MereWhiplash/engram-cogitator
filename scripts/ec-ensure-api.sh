#!/bin/bash
# Ensure the shared Engram Cogitator API container is running, then exec the shim.
# Singleton: fixed name, restart-policy, decoupled from any session PID.
set -uo pipefail

ENGRAM_DIR="$HOME/.engram"
API_NAME="engram-cogitator-api"
NETWORK="engram-network"
PORT="${EC_API_PORT:-8080}"
SHIM_BIN="${EC_SHIM_BIN:-$ENGRAM_DIR/ec-shim}"

EC_IMAGE="ghcr.io/merewhiplash/engram-cogitator:latest"
docker image inspect "$EC_IMAGE" &>/dev/null || EC_IMAGE="engram-cogitator:local"

# Storage config (defaults to sqlite at ~/.engram/memory.db)
EC_DB_PATH="$ENGRAM_DIR/memory.db"
# shellcheck source=/dev/null
[ -f "$ENGRAM_DIR/config" ] && . "$ENGRAM_DIR/config"
# Mount the DB's own directory so a custom EC_DB_PATH outside ~/.engram is honored.
EC_DB_DIR="$(dirname "$EC_DB_PATH")"

# NOTE: the image's default ENTRYPOINT is already /ec-api (Dockerfile), so no --entrypoint needed.
DOCKER_RUN=(docker run -d
  --name "$API_NAME"
  --restart unless-stopped
  --network "$NETWORK"
  -p "127.0.0.1:${PORT}:8080"
  -v "$EC_DB_DIR:/data"
  "$EC_IMAGE"
  --addr ":8080"
  --storage-driver sqlite
  --db-path "/data/$(basename "$EC_DB_PATH")"
  --ollama-url http://engram-ollama:11434)

if [ -n "${EC_DRY_RUN:-}" ]; then printf '%s ' "${DOCKER_RUN[@]}"; echo; exit 0; fi

# Ensure docker + network
command -v docker >/dev/null || { echo "docker not found; use ec-run.sh offline fallback" >&2; exit 1; }
docker network inspect "$NETWORK" &>/dev/null || docker network create "$NETWORK" &>/dev/null || true

# Ensure singleton api up
state="$(docker inspect -f '{{.State.Running}}' "$API_NAME" 2>/dev/null || echo missing)"
case "$state" in
  true) : ;;
  false) docker start "$API_NAME" &>/dev/null ;;
  *) "${DOCKER_RUN[@]}" &>/dev/null ;;
esac

# Wait for health (bounded)
for _ in $(seq 1 20); do
  if curl -fsS "http://127.0.0.1:${PORT}/health" &>/dev/null; then break; fi
  sleep 0.5
done

exec "$SHIM_BIN" --api-url "http://127.0.0.1:${PORT}"
