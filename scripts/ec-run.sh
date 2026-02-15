#!/bin/bash
# Engram Cogitator - MCP Server Wrapper
# Manages container lifecycle: naming, labeling, signal trapping, stale pruning.
# Replaces raw `docker run -i --rm` as the MCP stdio entry point.

set -uo pipefail
# No set -e: this is a long-running wrapper and we don't want transient
# docker errors (daemon hiccups, network blips) to kill the MCP session.

EC_IMAGE="ghcr.io/merewhiplash/engram-cogitator:latest"
if ! docker image inspect "$EC_IMAGE" &>/dev/null; then
    EC_IMAGE="engram-cogitator:local"
fi

ENGRAM_DIR="$HOME/.engram"
LABEL_PREFIX="io.engram-cogitator"

# Load storage config if present
EC_STORAGE_DRIVER=""
EC_DB_PATH=""
EC_POSTGRES_DSN=""
EC_MONGODB_URI=""
if [ -f "$ENGRAM_DIR/config" ]; then
    # shellcheck source=/dev/null
    . "$ENGRAM_DIR/config"
fi

# Derive project name from cwd, sanitized for Docker container names
PROJECT_DIR="$(pwd)"
PROJECT_SHORT="$(basename "$PROJECT_DIR" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9._-]/-/g')"
WRAPPER_PID=$$
CONTAINER_NAME="ec-mcp-${PROJECT_SHORT}-${WRAPPER_PID}"
HOST_NAME="$(hostname -s 2>/dev/null || hostname)"
STARTED="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

# --- Stale container pruning (runs in background, adds no latency) ---
prune_stale_containers() {
    # Remove any exited/dead/created EC containers immediately
    local stale
    stale=$(docker ps -a \
        --filter "label=${LABEL_PREFIX}.role=mcp-server" \
        --filter "status=exited" \
        --filter "status=dead" \
        --filter "status=created" \
        --format '{{.ID}}' 2>/dev/null) || true

    if [ -n "$stale" ]; then
        docker rm -f $stale &>/dev/null || true
    fi

    # For running containers: check if wrapper-pid is still alive on this host
    local running
    running=$(docker ps \
        --filter "label=${LABEL_PREFIX}.role=mcp-server" \
        --filter "label=${LABEL_PREFIX}.host=${HOST_NAME}" \
        --format '{{.ID}} {{.Label "io.engram-cogitator.wrapper-pid"}}' 2>/dev/null) || true

    if [ -n "$running" ]; then
        while IFS=' ' read -r cid pid; do
            # Skip our own container (it hasn't started yet, but be safe)
            if [ "$pid" = "$WRAPPER_PID" ]; then
                continue
            fi
            # Check if the wrapper process is still alive
            if [ -n "$pid" ] && ! kill -0 "$pid" 2>/dev/null; then
                docker rm -f "$cid" &>/dev/null || true
            fi
        done <<< "$running"
    fi
}

# Run pruning in background so it doesn't delay MCP startup
prune_stale_containers &
PRUNE_PID=$!

# --- Signal handling: ensure container cleanup on exit ---
cleanup() {
    # Graceful stop (allows SQLite WAL flush), then force-remove
    docker stop --time 5 "$CONTAINER_NAME" &>/dev/null || true
    docker rm -f "$CONTAINER_NAME" &>/dev/null || true
    # Reap background prune if still running
    kill "$PRUNE_PID" 2>/dev/null || true
    wait "$PRUNE_PID" 2>/dev/null || true
}
trap cleanup EXIT INT TERM HUP

# --- Start the MCP server container ---
# Do NOT use --rm — we own the lifecycle via the trap.
# Do NOT use exec — the trap must fire after docker exits.

# Guard against PID reuse: if a stale container with our name exists
# (background prune may not have reached it yet), remove it synchronously.
docker rm -f "$CONTAINER_NAME" &>/dev/null || true

# Build storage flags and volume mount based on configured driver
VOLUME_ARGS=""
STORAGE_ARGS=""

case "${EC_STORAGE_DRIVER:-sqlite}" in
    postgres)
        # No volume mount needed — connects over network
        STORAGE_ARGS="--storage-driver postgres --postgres-dsn ${EC_POSTGRES_DSN}"
        ;;
    mongodb)
        # No volume mount needed — connects over network
        STORAGE_ARGS="--storage-driver mongodb --mongodb-uri ${EC_MONGODB_URI}"
        ;;
    *)
        # SQLite: mount the DB directory into the container
        if [ -n "$EC_DB_PATH" ]; then
            DB_HOST_DIR="$(dirname "$EC_DB_PATH")"
            DB_FILENAME="$(basename "$EC_DB_PATH")"
            VOLUME_ARGS="-v $DB_HOST_DIR:/data"
            STORAGE_ARGS="--db-path /data/$DB_FILENAME"
        else
            VOLUME_ARGS="-v $ENGRAM_DIR:/data"
            STORAGE_ARGS="--db-path /data/memory.db"
        fi
        ;;
esac

docker run -i \
    --name "$CONTAINER_NAME" \
    --label "${LABEL_PREFIX}.role=mcp-server" \
    --label "${LABEL_PREFIX}.project=${PROJECT_DIR}" \
    --label "${LABEL_PREFIX}.project-short=${PROJECT_SHORT}" \
    --label "${LABEL_PREFIX}.wrapper-pid=${WRAPPER_PID}" \
    --label "${LABEL_PREFIX}.host=${HOST_NAME}" \
    --label "${LABEL_PREFIX}.started=${STARTED}" \
    --entrypoint /usr/local/bin/engram-cogitator \
    --network engram-network \
    ${VOLUME_ARGS} \
    "$EC_IMAGE" \
    ${STORAGE_ARGS} \
    --repo "$PROJECT_DIR" \
    --ollama-url http://engram-ollama:11434
DOCKER_EXIT=$?

# The exit trap handles cleanup regardless of how we get here.
exit $DOCKER_EXIT
