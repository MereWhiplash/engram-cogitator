# Engram Cogitator

This project uses Engram Cogitator for persistent semantic memory across sessions.

## MCP Tools Available

| Tool            | Purpose                                         |
| --------------- | ----------------------------------------------- |
| `ec_add`        | Store a memory (decision, learning, or pattern) |
| `ec_search`     | Find relevant memories by semantic similarity   |
| `ec_list`       | List recent memories                            |
| `ec_invalidate` | Mark a memory as outdated                       |

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

## Example Usage

```
# Store a decision
ec_add(type="decision", area="auth", content="Use JWT with 15min expiry", rationale="Balance security with UX")

# Search for context
ec_search(query="how do we handle authentication")

# List recent memories
ec_list(limit=10)
```
