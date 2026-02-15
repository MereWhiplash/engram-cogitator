# Engram Cogitator

This project uses Engram Cogitator for persistent semantic memory across sessions.

## MCP Tools Available

| Tool            | Purpose                                                          |
| --------------- | ---------------------------------------------------------------- |
| `ec_add`        | Store a memory (decision, learning, or pattern)                  |
| `ec_search`     | Find relevant memories by semantic similarity (returns scores)   |
| `ec_list`       | List recent memories                                             |
| `ec_invalidate` | Mark a memory as outdated                                        |

## When to Use

**Store memories when you:**

- Make architectural decisions (why X over Y)
- Discover gotchas or debugging insights
- Establish patterns or conventions

**Search memories when:**

- Starting work on a feature (check for prior decisions)
- Unsure about project conventions
- Debugging (check for known issues)

## Memory Types

| Type       | Use For                                           |
| ---------- | ------------------------------------------------- |
| `decision` | Architectural choices, tech selections, tradeoffs |
| `learning` | Debugging insights, gotchas, "TIL" moments        |
| `pattern`  | Recurring solutions, conventions, best practices  |

## Search Results

Search results include a `similarity_score` (0.0 to 1.0) indicating how closely each memory matches the query. Results are ranked by a combination of semantic similarity and recency, so recent relevant memories surface higher.

## Example Usage

```
# Store a decision
ec_add(type="decision", area="auth", content="Use JWT with 15min expiry", rationale="Balance security with UX")

# Search for context (results include similarity_score)
ec_search(query="how do we handle authentication")

# List recent memories
ec_list(limit=10)
```

## Storage Configuration

By default, EC uses SQLite at `~/.engram/memory.db`. The installer supports custom storage backends via `--db-path`:

```bash
./install.sh --db-path /shared/team/memory.db                  # SQLite at custom path
./install.sh --db-path postgres://user:pass@host:5432/engram   # PostgreSQL
./install.sh --db-path mongodb://host:27017/engram             # MongoDB
```

The storage driver is auto-detected from the connection string. Config is persisted to `~/.engram/config`.
