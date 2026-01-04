# Engram Cogitator (EC)

A local MCP server that gives Claude persistent, semantic memory across sessions.

## What it does

EC stores decisions, learnings, and patterns from your coding sessions in a local SQLite database with vector embeddings. When you start a new session, relevant context is automatically surfaced.

**Tools provided:**
- `ec_add` - Store a new memory (decision, learning, or pattern)
- `ec_search` - Semantic search across memories
- `ec_list` - List recent memories
- `ec_invalidate` - Soft-delete outdated memories

## Quick Install

```bash
curl -fsSL https://raw.githubusercontent.com/MereWhiplash/engram-cogitator/main/install.sh | bash
```

This will:
1. Create `.claude/` directory with `memory.db`
2. Add `memory.db` to `.gitignore`
3. Start Ollama container with embedding model
4. Configure MCP in `.claude/mcp.json`

## Manual Setup

### Prerequisites
- Docker
- Claude Code CLI

### Steps

1. Clone this repo:
   ```bash
   git clone https://github.com/MereWhiplash/engram-cogitator.git
   cd engram-cogitator
   ```

2. Build the image:
   ```bash
   docker build -t engram-cogitator:local .
   ```

3. Start Ollama:
   ```bash
   docker run -d --name engram-ollama --network engram-network -v ollama_data:/root/.ollama ollama/ollama
   docker exec engram-ollama ollama pull nomic-embed-text
   ```

4. Add to your project's `.claude/mcp.json`:
   ```json
   {
     "mcpServers": {
       "engram-cogitator": {
         "command": "docker",
         "args": [
           "run", "-i", "--rm",
           "--network", "engram-network",
           "-v", "${PWD}/.claude:/data",
           "engram-cogitator:local",
           "--db-path", "/data/memory.db",
           "--ollama-url", "http://engram-ollama:11434"
         ]
       }
     }
   }
   ```

## Usage Examples

### Storing a decision
```
Claude, remember this: We decided to use server actions for all mutations
because it keeps auth logic centralized. Store this as a decision in the
"architecture" area.
```

### Storing a learning
```
I just learned that the permissions system requires both org membership
AND project access checks. Add this as a learning in "permissions".
```

### Searching memories
```
What do we know about authentication patterns in this project?
```

## Configuration

Environment variables:
- `EC_DATA_PATH` - Path to data directory (default: `./.claude`)
- `EC_EMBEDDING_MODEL` - Ollama model for embeddings (default: `nomic-embed-text`)

## Architecture

```
┌─────────────────────────────────────┐
│  Claude Code                        │
│  (uses MCP tools)                   │
└──────────────┬──────────────────────┘
               │ MCP (stdio)
               ▼
┌─────────────────────────────────────┐
│  Engram Cogitator                   │
│  (Go MCP Server)                    │
│  - Tool handlers                    │
│  - sqlite-vec queries              │
└──────────────┬──────────────────────┘
               │
      ┌────────┴────────┐
      ▼                 ▼
┌───────────┐    ┌───────────────┐
│ SQLite DB │    │    Ollama     │
│ (vec0)    │    │ (embeddings)  │
└───────────┘    └───────────────────┘
```

## License

MIT
