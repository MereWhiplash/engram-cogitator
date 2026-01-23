<p align="center">
  <img src="logo.jpg" alt="Engram Cogitator" width="300">
</p>

<h1 align="center">Engram Cogitator</h1>

<p align="center">
  <em>The Machine Spirit remembers what you forget.</em><br>
  <em>(And unlike your da, it won't bring up that one time you dropped prod)</em>
</p>

<p align="center">
  Persistent semantic memory for AI coding assistants.<br>
  By the grace of the Omnissiah, your decisions shall not be lost to the warp.
</p>

<p align="center">
  <strong>Works with:</strong> Claude Code, Cursor, Cline, Windsurf, and any MCP-compatible client
</p>

---

## Quick Start (Solo Mode)

**Step 1: Install the MCP server**
```bash
curl -sSL https://raw.githubusercontent.com/MereWhiplash/engram-cogitator/main/install.sh | bash
```

**Step 2: Install the cogitation plugin** (in Claude Code)
```
/plugin marketplace add MereWhiplash/engram-cogitator
/plugin install cogitation@engram-cogitator
```

**Step 3: Initialize your project**
```
/cogitation:init
```

**Prerequisites:** Docker running. That's it.

---

## What's Included

### MCP Server (ec_* tools)
The memory backend. Provides `ec_add`, `ec_search`, `ec_list`, `ec_invalidate` for storing and retrieving semantic memories.

### Cogitation Plugin (skills)
Opinionated development workflows that leverage EC's persistent memory.

#### The Development Workflow

```
┌─────────────────────────────────────────────────────────────────────┐
│                                                                     │
│   /cogitation:init          (one-time project setup)                │
│         │                                                           │
│         ▼                                                           │
│   /cogitation:brainstorming (explore idea → design doc)             │
│         │                                                           │
│         ▼                                                           │
│   /cogitation:writing-plans (design doc → implementation plan)      │
│         │                                                           │
│         ▼                                                           │
│   /cogitation:executing-plans (plan → working code)                 │
│         │                                                           │
│         ▼                                                           │
│   /cogitation:finish        (verify, store learnings, merge)        │
│                                                                     │
│   /cogitation:audit         (periodic memory cleanup)               │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

#### Main Commands

**`/cogitation:init`** - First-time project setup
- Verifies EC connection is working
- Scaffolds `docs/designs/` and `docs/plans/` directories
- Configures test, build, and lint commands for your project
- Stores project config in EC for other skills to use

**`/cogitation:brainstorming`** - Turn ideas into designs
- Searches EC for prior decisions and patterns in related areas
- Creates a feature branch
- Guides you through requirements via Q&A
- Generates multiple approaches with tradeoffs
- Produces a design doc in `docs/designs/YYYY-MM-DD-<topic>.md`
- Stores key decisions in EC

**`/cogitation:writing-plans`** - Turn designs into plans
- Loads the design doc and relevant EC context
- Breaks work into small, testable batches (TDD-style)
- Each batch: write test → implement → verify
- Produces a plan in `docs/plans/YYYY-MM-DD-<topic>.md`
- Plans are detailed enough for zero-context execution

**`/cogitation:executing-plans`** - Turn plans into code
- Loads plan and EC context before each batch
- Executes batches with review checkpoints
- Runs tests after each batch
- Pauses for your approval before continuing
- Stores learnings discovered during implementation

**`/cogitation:finish`** - Complete the work
- Runs all verification (tests, lint, build)
- Reviews session for memories worth storing
- Presents merge options (squash, merge, rebase)
- Cleans up feature branch after merge

**`/cogitation:audit`** - Memory maintenance
- Lists all stored memories by type
- Finds duplicates and stale entries
- Presents cleanup recommendations
- Invalidates outdated memories

#### Additional Skills

These are invoked automatically by Claude based on context:

| Skill | When It's Used |
|-------|----------------|
| `onboard` | Start of session - loads project context from EC |
| `tdd` | Writing tests and implementation |
| `debugging` | Investigating bugs or test failures |
| `remember` | When you say "remember this" or make key decisions |
| `config` | Updating project test/build/lint commands |
| `verifying` | Before claiming tests pass or work is done |
| `requesting-review` | Preparing a PR for code review |
| `receiving-review` | Responding to code review feedback |
| `parallel-agents` | When multiple independent tasks can run concurrently |

---

## What It Does

The Engram Cogitator serves as an auxiliary memory core for your AI coding sessions. Like a good Irishman's grudge, it never forgets. It stores:

| Type          | What It Captures                                |
| ------------- | ----------------------------------------------- |
| **Decisions** | Architectural choices and the "why" behind them |
| **Learnings** | Hard-won knowledge from debugging sessions      |
| **Patterns**  | Recurring solutions and team conventions        |

All memories are searchable by semantic similarity. Ask "how do we handle auth?" and it finds relevant memories even if they don't contain the word "auth".

---

## Two Modes

|              | Solo Mode                   | Team Mode                        |
| ------------ | --------------------------- | -------------------------------- |
| **For**      | Individual developers       | Engineering teams                |
| **Storage**  | Local SQLite per project    | Shared Postgres/MongoDB          |
| **Memories** | Private to you              | Shared across team               |
| **Setup**    | One command                 | Kubernetes + Helm                |
| **Best for** | Personal projects, learning | Cross-pollinating team knowledge |

---

## Solo Mode

Your memories stay local in a single global database at `~/.engram/memory.db`. Each project is automatically identified by its git remote (or directory path as fallback), so memories are scoped to the current project but searchable across all projects. No data leaves your machine.

### Architecture

```
┌─────────────────────────────────────────────────────┐
│                   Your Machine                       │
│                                                      │
│  ┌──────────────┐      MCP/stdio      ┌──────────┐  │
│  │ Claude Code  │◄───────────────────►│ EC Server│  │
│  │   Cursor     │                     │ (Docker) │  │
│  │   Cline      │                     └────┬─────┘  │
│  │   Windsurf   │                          │        │
│  └──────────────┘                          │        │
│                                            ▼        │
│                               ┌────────────────────┐│
│                               │ SQLite + Ollama    ││
│                               │ ~/.engram/memory.db││
│                               └────────────────────┘│
└─────────────────────────────────────────────────────┘
```

**Project identity is auto-detected:**
1. Git remote origin → normalized to "org/repo"
2. If no git remote, uses absolute directory path

### Installation Options

#### Option 1: Docker (Recommended)

```bash
curl -sSL https://raw.githubusercontent.com/MereWhiplash/engram-cogitator/main/install.sh | bash
```

This installs:

- Ollama container for local embeddings
- EC server container
- MCP configuration for your client

**Note:** Skills are now installed separately via the cogitation plugin (see Quick Start).

#### Option 2: Binary + Local Ollama

If you already have Ollama running locally:

```bash
# Download binary
VERSION=$(curl -s https://api.github.com/repos/MereWhiplash/engram-cogitator/releases/latest | grep tag_name | cut -d '"' -f 4)
curl -sSL "https://github.com/MereWhiplash/engram-cogitator/releases/download/${VERSION}/ec-server_${VERSION#v}_$(uname -s | tr '[:upper:]' '[:lower:]')_amd64.tar.gz" | tar -xz

# Run with local Ollama
./ec-server --db-path .engram/memory.db --ollama-url http://localhost:11434
```

### MCP Client Configuration

The install script auto-configures Claude Code. For other clients, add this to your MCP config:

<details>
<summary><strong>Claude Code</strong></summary>

Auto-configured by install script, or manually:

```bash
claude mcp add --transport stdio engram-cogitator \
  --scope user \
  -- /bin/sh -c 'docker run -i --rm --entrypoint /usr/local/bin/engram-cogitator --network engram-network -v $HOME/.engram:/data ghcr.io/merewhiplash/engram-cogitator:latest --db-path /data/memory.db --repo "$(pwd)" --ollama-url http://engram-ollama:11434'
```

Note: The `/bin/sh -c` wrapper captures the working directory at runtime for project detection. The `--entrypoint` flag selects the MCP server binary.

</details>

<details>
<summary><strong>Cursor</strong> (~/.cursor/mcp.json)</summary>

```json
{
  "mcpServers": {
    "engram-cogitator": {
      "command": "docker",
      "args": [
        "run", "-i", "--rm",
        "--entrypoint", "/usr/local/bin/engram-cogitator",
        "--network", "engram-network",
        "-v", "${HOME}/.engram:/data",
        "ghcr.io/merewhiplash/engram-cogitator:latest",
        "--db-path", "/data/memory.db",
        "--repo", "${workspaceFolder}",
        "--ollama-url", "http://engram-ollama:11434"
      ]
    }
  }
}
```

</details>

<details>
<summary><strong>VS Code / GitHub Copilot</strong></summary>

Add to VS Code settings (MCP configuration):

```json
{
  "mcpServers": {
    "engram-cogitator": {
      "command": "docker",
      "args": [
        "run", "-i", "--rm",
        "--entrypoint", "/usr/local/bin/engram-cogitator",
        "--network", "engram-network",
        "-v", "${HOME}/.engram:/data",
        "ghcr.io/merewhiplash/engram-cogitator:latest",
        "--db-path", "/data/memory.db",
        "--repo", "${workspaceFolder}",
        "--ollama-url", "http://engram-ollama:11434"
      ]
    }
  }
}
```

</details>

### AI Assistant Instructions

For your AI assistant to understand how to use Engram Cogitator, add the contents of [`INSTRUCTIONS.md`](INSTRUCTIONS.md) to your project's instruction file:

| AI Assistant | Instruction File |
|--------------|------------------|
| Claude Code | `CLAUDE.md` |
| Cursor | `.cursor/rules/engram.mdc` |
| GitHub Copilot | `.github/copilot-instructions.md` |
| Gemini | `GEMINI.md` |
| Generic | `AGENTS.md` |

The install script automatically sets up `CLAUDE.md` and downloads `INSTRUCTIONS.md` to `.engram/` for other assistants.

---

## Team Mode

Share memories across your engineering team. When Alice discovers a gotcha in the payment API, Bob finds it when he searches a week later.

### Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  Developer A    │     │  Developer B    │     │  Developer C    │
│  (frontend)     │     │  (backend)      │     │  (mobile)       │
│  ┌───────────┐  │     │  ┌───────────┐  │     │  ┌───────────┐  │
│  │ Claude/   │  │     │  │ Cursor    │  │     │  │ Cline     │  │
│  │ Cursor    │  │     │  │           │  │     │  │           │  │
│  └─────┬─────┘  │     │  └─────┬─────┘  │     │  └─────┬─────┘  │
│        │        │     │        │        │     │        │        │
│  ┌─────▼─────┐  │     │  ┌─────▼─────┐  │     │  ┌─────▼─────┐  │
│  │ ec-shim   │  │     │  │ ec-shim   │  │     │  │ ec-shim   │  │
│  └─────┬─────┘  │     │  └─────┬─────┘  │     │  └─────┬─────┘  │
└────────┼────────┘     └────────┼────────┘     └────────┼────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │ HTTPS
                                 ▼
                    ┌────────────────────────┐
                    │   Kubernetes Cluster   │
                    │  ┌──────────────────┐  │
                    │  │    EC API        │  │
                    │  │   (replicas: 3)  │  │
                    │  └────────┬─────────┘  │
                    │           │            │
                    │  ┌────────▼─────────┐  │
                    │  │ Postgres/MongoDB │  │
                    │  │ + Ollama         │  │
                    │  └──────────────────┘  │
                    └────────────────────────┘
```

### Quick Start (Team)

**1. Deploy to Kubernetes:**

```bash
helm repo add engram https://merewhiplash.github.io/engram-cogitator
helm repo update

helm install engram engram/engram-cogitator \
  --namespace engram \
  --create-namespace \
  --set storage.postgres.password=<your-password> \
  --set ingress.enabled=true \
  --set ingress.hosts[0].host=engram.yourcompany.com
```

**2. Install shim on developer machines:**

```bash
EC_API_URL=https://engram.yourcompany.com \
  curl -sSL https://raw.githubusercontent.com/MereWhiplash/engram-cogitator/main/install-team.sh | bash
```

**3. Restart your AI coding assistant**

### How The Shim Works

The shim is a lightweight MCP proxy that:

1. Extracts your git identity (`git config user.name/email`)
2. Extracts the repo from `git remote origin`
3. Forwards requests to the central API with attribution

Every memory is tagged with who added it and from which repo.

### Cross-Repo Search

By default, searches find memories from **all repositories**:

```json
{
  "content": "Use circuit breakers for external API calls",
  "author_name": "Alice Smith",
  "author_email": "alice@company.com",
  "repo": "myorg/backend-api",
  "type": "pattern"
}
```

### Helm Configuration

Key options (see [values.yaml](charts/engram-cogitator/values.yaml) for all):

```yaml
storage:
  driver: postgres # or mongodb
  postgres:
    internal: true # deploy StatefulSet, or false for external
    password: <secret>

api:
  replicas: 3

autoscaling:
  enabled: true
  maxReplicas: 10

ingress:
  enabled: true
  hosts:
    - host: engram.yourcompany.com
```

---

## MCP Tools Reference

| Tool            | Purpose                 | Example                                         |
| --------------- | ----------------------- | ----------------------------------------------- |
| `ec_add`        | Store a memory          | "Remember that we use UUIDs for all entity IDs" |
| `ec_search`     | Find relevant memories  | "How do we handle authentication?"              |
| `ec_list`       | Show recent memories    | "What did we decide recently?"                  |
| `ec_invalidate` | Mark memory as outdated | "That decision about Redux is no longer valid"  |

### Memory Types

| Type       | When to Use                                       |
| ---------- | ------------------------------------------------- |
| `decision` | Architectural choices, tech selections, tradeoffs |
| `learning` | Debugging insights, gotchas, "TIL" moments        |
| `pattern`  | Recurring solutions, conventions, best practices  |

---

## Usage Examples

The AI assistant uses these tools automatically when relevant. You can also prompt it directly:

**Storing decisions:**

> "Remember that we chose Postgres over MongoDB because we need strong consistency for financial transactions"

**Finding context:**

> "Search memories for how we handle rate limiting"

**Reviewing recent work:**

> "List the last 10 memories from this project"

**Updating outdated info:**

> "Invalidate the memory about using REST - we switched to GraphQL"

### Cogitation Workflows (Claude Code)

The cogitation plugin provides opinionated workflows that integrate with EC. Use `/cogitation:onboard` at the start of a session to load relevant memories:

```
Project Context Loaded

Configuration:
- Test: go test ./...
- Lint: golangci-lint run

Key Decisions (3):
- [auth] Use JWT with 15-minute expiry, refresh tokens in httpOnly cookies
- [api] All endpoints return {data, error, meta} shape

Gotchas to Know (2):
- [postgres] Connection pooling maxes at 20 for this instance size
```

The workflows automatically search and store memories as you work through the development cycle.

---

## Troubleshooting

### "readonly database" error

```bash
chmod 777 .engram
rm .engram/memory.db  # Let it rebuild
```

### "Dimension mismatch" error

Embedding model changed. Reset the database:

```bash
rm .engram/memory.db
```

### MCP server not showing up

```bash
claude mcp list  # Should show engram-cogitator
```

Config lives in `.mcp.json` (project root).

### Docker network issues

```bash
docker network create engram-network
docker start engram-ollama
```

### Still broken?

Restart your AI coding assistant. Works 60% of the time, every time.

---

## Development

```bash
# Run tests
CGO_ENABLED=1 go test ./...

# Build locally
CGO_ENABLED=1 go build -o ec-server ./cmd/server
CGO_ENABLED=0 go build -o ec-api ./cmd/api
CGO_ENABLED=0 go build -o ec-shim ./cmd/shim

# Build Docker image
docker build -t engram-cogitator:local .

# Lint
golangci-lint run ./...
```

### Pre-commit Hooks

```bash
pip install pre-commit
pre-commit install
```

---

## Acknowledgments

A huge thank you to the team at [Obra](https://github.com/obra) whose [Superpowers](https://github.com/obra/superpowers) project was a major influence on the cogitation skills. Their work on structured AI-assisted development workflows helped shape the patterns and philosophies behind this project.

---

<p align="center">
  <em>Praise the Omnissiah. Store your memories. Ship your code.</em><br>
  <em>The Emperor Protects, but version control saves.</em>
</p>

<p align="center">
  Made with mass-produced servitor love in Ireland 🇮🇪<br>
  <sub>Now stop reading READMEs and go build something, ya gobshite</sub>
</p>
