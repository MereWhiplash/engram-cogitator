# Engram Cogitator

Go MCP server providing persistent semantic memory for Claude Code sessions.

## Quick Context

Read `HANDOVER.md` for full context on what's built and what's next.

## Rules

- User runs dev/build - never run servers
- No summary docs without asking
- Use `tree` command to see structure
- Don't commit without being told

## Stack

- Go 1.22+
- Official `modelcontextprotocol/go-sdk`
- sqlite-vec for vector search
- Ollama for embeddings

## Commands

```bash
# Build locally (requires CGO)
CGO_ENABLED=1 go build ./cmd/server

# Docker build
docker build -t engram-cogitator:local .

# Run tests (when added)
go test ./...
```
