#!/bin/bash
# Engram Cogitator - Session Start Hook
# Retrieves recent memories and injects them into Claude's context

# Try ghcr.io first, fall back to local
EC_IMAGE="ghcr.io/merewhiplash/engram-cogitator:latest"
if ! docker image inspect "$EC_IMAGE" &>/dev/null; then
    EC_IMAGE="engram-cogitator:local"
fi

# Check if EC network exists (silent fail if not set up)
if ! docker network inspect engram-network &>/dev/null; then
    exit 0
fi

# Check if Ollama is running
if ! docker ps --format '{{.Names}}' | grep -q '^engram-ollama$'; then
    exit 0
fi

# Use project directory (set by Claude Code)
PROJECT_DIR="${CLAUDE_PROJECT_DIR:-.}"

# Check if memory.db exists
if [ ! -f "$PROJECT_DIR/.engram/memory.db" ]; then
    exit 0
fi

# Derive container name for the hook
PROJECT_SHORT="$(basename "$PROJECT_DIR" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9._-]/-/g')"
HOOK_CONTAINER_NAME="ec-hook-${PROJECT_SHORT}-$$"

# Get recent memories via CLI mode
# Hook containers use --rm (they exit quickly so --rm works fine)
MEMORIES=$(docker run -i --rm \
    --name "$HOOK_CONTAINER_NAME" \
    --label "io.engram-cogitator.role=session-hook" \
    --label "io.engram-cogitator.project=${PROJECT_DIR}" \
    --label "io.engram-cogitator.project-short=${PROJECT_SHORT}" \
    --network engram-network \
    -v "$PROJECT_DIR/.engram:/data" \
    "$EC_IMAGE" \
    --db-path /data/memory.db \
    --list --limit 5 2>/dev/null)

# If we got memories, output them for context injection
if [ -n "$MEMORIES" ]; then
    echo "=== Previous Session Memories ==="
    echo "$MEMORIES"
    echo "================================="
fi

exit 0
