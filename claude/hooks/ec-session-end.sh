#!/bin/bash
# Engram Cogitator - Session End Hook
# Stops and removes MCP server containers for the current project.
# This catches the case where Docker Desktop on macOS doesn't propagate
# stdin EOF, leaving the wrapper and container hanging after session exit.

PROJECT_DIR="${CLAUDE_PROJECT_DIR:-.}"
LABEL_PREFIX="io.engram-cogitator"

# Find running MCP containers for this project
CONTAINERS=$(docker ps \
    --filter "label=${LABEL_PREFIX}.role=mcp-server" \
    --filter "label=${LABEL_PREFIX}.project=${PROJECT_DIR}" \
    --format '{{.ID}}' 2>/dev/null) || true

if [ -n "$CONTAINERS" ]; then
    for cid in $CONTAINERS; do
        # Graceful stop (allows SQLite WAL flush), then force-remove
        docker stop --time 5 "$cid" &>/dev/null || true
        docker rm -f "$cid" &>/dev/null || true
    done
fi

exit 0
