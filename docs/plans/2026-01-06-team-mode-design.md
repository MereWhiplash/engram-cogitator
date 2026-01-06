# Team Mode Design

**Date:** 2026-01-06
**Status:** Draft
**Author:** Brainstormed with Claude

## Overview

Extend Engram Cogitator to support centralised team usage (10-20 engineers on shared codebases) while preserving the existing solo-developer mode.

### Goals

- Cross-pollination: "Has anyone on the team solved this before?"
- Repo-scoped queries: "What decisions have we made in this codebase?"
- Keep solo mode unchanged and working
- Self-hosted on Kubernetes

### Non-Goals (for now)

- Hybrid mode (local + team memories merged)
- Public SaaS deployment
- Authentication/authorization (internal tool, trusted network)

---

## Architecture

### Two Deployment Modes

```
SOLO MODE (unchanged):
[Claude Code] → stdio → [MCP Server] → [SQLite]
                              ↓
                         [Local Ollama]

TEAM MODE (new):
[Claude Code] → stdio → [MCP Shim] → HTTP → [Central API] → [Postgres/MongoDB]
                              ↓                    ↓
                        (extracts git info)   [Ollama]
```

**Key principle:** Same MCP tool interface (`ec_add`, `ec_search`, etc.) in both modes. Claude doesn't know or care which mode it's in.

### Team Mode Components

| Component | What it does | Runs where |
|-----------|--------------|------------|
| MCP Shim | Extracts git author/repo, forwards to API | Each dev machine (stdio) |
| Central API | Business logic, embedding, storage | K8s pod |
| Postgres OR MongoDB | Persistent storage, vector search | K8s / managed service |
| Ollama | Embedding generation | K8s pod |

### Mode Selection

The shim checks for `EC_API_URL` environment variable:
- Set → team mode (forward to central API)
- Not set → solo mode (use local SQLite + Ollama)

---

## Data Model

### Current Schema (Solo)

```sql
memories (id, type, area, content, rationale, is_valid, superseded_by, created_at)
```

### Team Schema

**New fields:**
- `author_name` - from git config user.name
- `author_email` - from git config user.email
- `repo` - normalized git remote URL (e.g., `yourorg/repo-name`)

#### Postgres + pgvector

```sql
CREATE EXTENSION vector;

CREATE TABLE memories (
  id            SERIAL PRIMARY KEY,
  type          TEXT NOT NULL CHECK(type IN ('decision', 'learning', 'pattern')),
  area          TEXT NOT NULL,
  content       TEXT NOT NULL,
  rationale     TEXT,
  is_valid      BOOLEAN NOT NULL DEFAULT TRUE,
  superseded_by INTEGER REFERENCES memories(id),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  author_name   TEXT NOT NULL,
  author_email  TEXT NOT NULL,
  repo          TEXT NOT NULL
);

CREATE TABLE memory_embeddings (
  memory_id     INTEGER PRIMARY KEY REFERENCES memories(id),
  embedding     vector(768)
);

CREATE INDEX idx_memories_repo ON memories(repo);
CREATE INDEX idx_memories_author ON memories(author_email);
CREATE INDEX idx_memories_area_repo ON memories(area, repo);
```

#### MongoDB + Atlas Vector Search

```json
{
  "_id": ObjectId,
  "type": "decision",
  "area": "auth",
  "content": "Use JWT with short-lived tokens...",
  "rationale": "Stateless, scales horizontally",
  "is_valid": true,
  "superseded_by": null,
  "created_at": ISODate,
  "author": {
    "name": "Alice Smith",
    "email": "alice@company.com"
  },
  "repo": "yourorg/backend",
  "embedding": [0.123, -0.456, ...]
}
```

Atlas Vector Search index on `embedding` field.

### Search Behavior

- `ec_search` defaults to **all repos** (cross-pollination)
- Optional `repo` filter to scope to current codebase
- Results include author attribution

---

## Local MCP Shim

### Responsibilities

1. Receive MCP tool calls from Claude Code (stdio)
2. Extract git author + repo from local environment
3. Forward to central API with context attached
4. Return response to Claude

### Git Extraction (runs once at startup)

```go
// Author from git config
name, _  := exec.Command("git", "config", "user.name").Output()
email, _ := exec.Command("git", "config", "user.email").Output()

// Repo from remote URL, normalized
remote, _ := exec.Command("git", "config", "--get", "remote.origin.url").Output()
// "git@github.com:yourorg/repo.git" → "yourorg/repo"
// "https://github.com/yourorg/repo.git" → "yourorg/repo"
repo := normalizeRemoteURL(remote)
```

### Request Forwarding

Every API request includes context headers:
```
X-EC-Author-Name: Alice Smith
X-EC-Author-Email: alice@company.com
X-EC-Repo: yourorg/backend
```

### Installation

```bash
# Team mode
EC_API_URL=https://engram.internal.company.com claude mcp add engram-cogitator -- ec-shim
```

---

## Central API

### Endpoints

```
POST /v1/memories              # ec_add
GET  /v1/memories              # ec_list
POST /v1/memories/search       # ec_search
PUT  /v1/memories/:id/invalidate   # ec_invalidate
GET  /health                   # K8s probes
```

### Request Flow: ec_add

```
Shim → POST /v1/memories
       Headers: X-EC-Author-Name, X-EC-Author-Email, X-EC-Repo
       Body: { "type": "decision", "area": "auth", "content": "...", "rationale": "..." }

API  → 1. Call Ollama to embed content
       2. Insert into storage (Postgres or MongoDB)
       3. Return created memory with ID
```

### Request Flow: ec_search

```
Shim → POST /v1/memories/search
       Headers: X-EC-Author-*, X-EC-Repo
       Body: { "query": "how did we handle rate limiting", "limit": 5, "repo": null }

API  → 1. Embed query via Ollama
       2. Vector similarity search (all repos since repo=null)
       3. Return memories with author attribution
```

---

## Storage Interface

```go
type Storage interface {
    Add(ctx context.Context, memory Memory, embedding []float32) (*Memory, error)
    Search(ctx context.Context, embedding []float32, opts SearchOpts) ([]Memory, error)
    List(ctx context.Context, opts ListOpts) ([]Memory, error)
    Invalidate(ctx context.Context, id int64, supersededBy *int64) error
    Migrate(ctx context.Context) error
    Close() error
}
```

### Implementations

| Implementation | Use Case |
|----------------|----------|
| `sqlite.go` | Solo mode (current, extracted) |
| `postgres.go` | Team mode + pgvector |
| `mongodb.go` | Team mode + Atlas Vector Search |

---

## Kubernetes Deployment

### Components

| Component | Replicas | Resources | Notes |
|-----------|----------|-----------|-------|
| engram-api | 2-3 | 256Mi-512Mi RAM | Stateless, HPA enabled |
| ollama | 1 | 2Gi RAM | Single replica fine for 10-20 users |
| postgres | 1 | 1Gi RAM + PVC | Only if using Postgres |

### Helm Chart Structure

```
charts/engram-cogitator/
  Chart.yaml
  values.yaml
  templates/
    _helpers.tpl

    # API
    api-deployment.yaml
    api-service.yaml
    api-hpa.yaml
    api-pdb.yaml

    # Ollama
    ollama-deployment.yaml
    ollama-service.yaml
    ollama-pvc.yaml

    # Postgres (optional)
    postgres-statefulset.yaml
    postgres-service.yaml
    postgres-pvc.yaml

    # Security
    networkpolicy.yaml
    secrets.yaml

    # Ingress
    ingress.yaml

    # Migrations
    configmap-migrations.yaml
```

### Network Policies

Only allow:
- Ingress → API (8080)
- API → Postgres (5432)
- API → Ollama (11434)
- API → MongoDB (27017, egress to Atlas or internal)

Deny all other traffic.

### Probes

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 2
  periodSeconds: 5
```

### Init Container (Migrations)

```yaml
initContainers:
  - name: migrate
    image: engram-cogitator:{{ .Values.image.tag }}
    command: ["./ec-api", "migrate"]
    env:
      - name: DATABASE_URL
        valueFrom:
          secretKeyRef: ...
```

### Values Configuration

```yaml
storage:
  driver: postgres  # or mongodb

  postgres:
    internal: true  # spin up StatefulSet
    # OR
    internal: false
    host: external-postgres.example.com
    port: 5432
    database: engram

  mongodb:
    # User provides connection URI
    uri: mongodb+srv://user:pass@cluster.mongodb.net
    database: engram

ollama:
  model: nomic-embed-text
  resources:
    requests:
      memory: 2Gi
      cpu: 1

api:
  replicas: 2
  resources:
    requests:
      memory: 256Mi
      cpu: 250m
```

### Future Option: MongoDB Kubernetes Operator

For teams with the MongoDB Kubernetes Operator installed, could optionally create a `MongoDBCommunity` resource:

```yaml
storage:
  mongodb:
    operator:
      enabled: true
      members: 3
      version: "7.0.0"
```

Not in initial scope - start with URI-only, add operator support later if needed.

---

## Migration Path

### Phase 1: Refactor Internals

- Extract `Storage` interface (SQLite implementation)
- Extract `Embedder` interface (Ollama implementation)
- Move business logic to `Service` layer
- No new features, just cleaner separation

### Phase 2: Add Storage Backends

- Add Postgres implementation with pgvector
- Add MongoDB implementation with Atlas Vector Search
- Add migration tooling (golang-migrate for Postgres)
- Config flag: `--storage-driver sqlite|postgres|mongodb`
- Tests pass with all backends

### Phase 3: Build Central API

- New `cmd/api` entry point
- HTTP handlers calling `Service` layer
- Extract git context from headers
- `/health` endpoint
- Separate Dockerfile

### Phase 4: Build Shim

- New `cmd/shim` entry point
- Git author/repo extraction
- HTTP client to central API
- Same MCP interface as current server

### Phase 5: Helm Chart

- Chart with all components
- Network policies, probes, init container
- Values for Postgres vs MongoDB
- Internal ingress

### Phase 6: Distribution

- Shim binary releases (homebrew, curl install)
- Helm repo for chart
- Updated docs with team setup

---

## Code Structure (Target)

```
cmd/
  server/main.go      # solo mode entry (current)
  shim/main.go        # team mode shim (new)
  api/main.go         # central API (new)

internal/
  storage/
    storage.go        # interface
    sqlite.go         # solo mode
    postgres.go       # team mode
    mongodb.go        # team mode
  embed/
    embed.go          # interface + Ollama implementation
  service/
    service.go        # business logic
  api/
    handlers.go       # HTTP handlers
    middleware.go     # extract git context from headers
  mcp/
    tools.go          # MCP tool definitions (shared)

charts/
  engram-cogitator/
    ...

migrations/
  postgres/
    001_initial.up.sql
    001_initial.down.sql
```

---

## Summary

| Decision | Choice |
|----------|--------|
| Use case | Cross-pollination + repo-scoped |
| Deployment | Self-hosted, Kubernetes, Helm |
| Identity | Git author info (no auth) |
| Repo ID | Git remote URL, normalized |
| Embeddings | Central Ollama |
| Storage | Postgres OR MongoDB (user's choice) |
| Solo mode | Unchanged, still works |
| Hybrid mode | Future enhancement |
| MongoDB Operator | Future enhancement |
