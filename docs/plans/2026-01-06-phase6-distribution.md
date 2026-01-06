# Phase 6: Distribution - Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Set up release automation for binaries, container images, and Helm chart. Update documentation for team mode.

**Architecture:** GitHub Actions for CI/CD, GitHub Container Registry for images, GitHub Releases for binaries, GitHub Pages for Helm repo.

**Tech Stack:** GitHub Actions, GoReleaser, Docker, Helm

**Prerequisites:** Phase 5 complete (Helm chart exists)

---

## Task 1: Create GoReleaser Config

**Files:**
- Create: `.goreleaser.yaml`

**Step 1: Write GoReleaser config**

```yaml
# .goreleaser.yaml
version: 2

project_name: engram-cogitator

before:
  hooks:
    - go mod tidy

builds:
  # Solo mode MCP server (CGO required for SQLite)
  - id: server
    main: ./cmd/server
    binary: ec-server
    env:
      - CGO_ENABLED=1
    goos:
      - linux
      - darwin
    goarch:
      - amd64
      - arm64
    ignore:
      - goos: linux
        goarch: arm64
    ldflags:
      - -s -w -X main.version={{.Version}}

  # API server (no CGO)
  - id: api
    main: ./cmd/api
    binary: ec-api
    env:
      - CGO_ENABLED=0
    goos:
      - linux
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w -X main.version={{.Version}}

  # Shim (no CGO)
  - id: shim
    main: ./cmd/shim
    binary: ec-shim
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w -X main.version={{.Version}}

archives:
  - id: server
    builds:
      - server
    name_template: "ec-server_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    format: tar.gz
    format_overrides:
      - goos: windows
        format: zip

  - id: shim
    builds:
      - shim
    name_template: "ec-shim_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    format: tar.gz
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: 'checksums.txt'

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^chore:'

release:
  github:
    owner: MereWhiplash
    name: engram-cogitator
  prerelease: auto
  draft: false
```

**Step 2: Commit**

```bash
git add .goreleaser.yaml
git commit -m "chore: add GoReleaser configuration"
```

---

## Task 2: Create GitHub Actions for CI

**Files:**
- Create: `.github/workflows/ci.yaml`

**Step 1: Write CI workflow**

```yaml
# .github/workflows/ci.yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Install dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y gcc

      - name: Run tests
        run: CGO_ENABLED=1 go test -v -race ./...

      - name: Build binaries
        run: |
          CGO_ENABLED=1 go build -o ec-server ./cmd/server
          CGO_ENABLED=0 go build -o ec-api ./cmd/api
          CGO_ENABLED=0 go build -o ec-shim ./cmd/shim

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest

  helm-lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Helm
        uses: azure/setup-helm@v4

      - name: Lint Helm chart
        run: helm lint charts/engram-cogitator
```

**Step 2: Commit**

```bash
mkdir -p .github/workflows
git add .github/workflows/ci.yaml
git commit -m "ci: add CI workflow for tests and linting"
```

---

## Task 3: Create GitHub Actions for Release

**Files:**
- Create: `.github/workflows/release.yaml`

**Step 1: Write release workflow**

```yaml
# .github/workflows/release.yaml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write
  packages: write

jobs:
  release-binaries:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Install dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y gcc

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v5
        with:
          distribution: goreleaser
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

  release-images:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Login to GitHub Container Registry
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract version
        id: version
        run: echo "VERSION=${GITHUB_REF#refs/tags/v}" >> $GITHUB_OUTPUT

      - name: Build and push API image
        uses: docker/build-push-action@v5
        with:
          context: .
          file: Dockerfile.api
          push: true
          platforms: linux/amd64,linux/arm64
          tags: |
            ghcr.io/merewhiplash/engram-cogitator:${{ steps.version.outputs.VERSION }}
            ghcr.io/merewhiplash/engram-cogitator:latest

      - name: Build and push shim image
        uses: docker/build-push-action@v5
        with:
          context: .
          file: Dockerfile.shim
          push: true
          platforms: linux/amd64,linux/arm64
          tags: |
            ghcr.io/merewhiplash/engram-cogitator-shim:${{ steps.version.outputs.VERSION }}
            ghcr.io/merewhiplash/engram-cogitator-shim:latest

  release-helm:
    runs-on: ubuntu-latest
    needs: release-images
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Configure Git
        run: |
          git config user.name "$GITHUB_ACTOR"
          git config user.email "$GITHUB_ACTOR@users.noreply.github.com"

      - name: Set up Helm
        uses: azure/setup-helm@v4

      - name: Extract version
        id: version
        run: echo "VERSION=${GITHUB_REF#refs/tags/v}" >> $GITHUB_OUTPUT

      - name: Update chart version
        run: |
          sed -i "s/^version:.*/version: ${{ steps.version.outputs.VERSION }}/" charts/engram-cogitator/Chart.yaml
          sed -i "s/^appVersion:.*/appVersion: \"${{ steps.version.outputs.VERSION }}\"/" charts/engram-cogitator/Chart.yaml

      - name: Package Helm chart
        run: |
          helm package charts/engram-cogitator -d .helm-packages

      - name: Checkout gh-pages
        uses: actions/checkout@v4
        with:
          ref: gh-pages
          path: gh-pages

      - name: Update Helm repo
        run: |
          cp .helm-packages/*.tgz gh-pages/
          cd gh-pages
          helm repo index . --url https://merewhiplash.github.io/engram-cogitator
          git add .
          git commit -m "Release helm chart ${{ steps.version.outputs.VERSION }}"
          git push
```

**Step 2: Commit**

```bash
git add .github/workflows/release.yaml
git commit -m "ci: add release workflow for binaries, images, and helm"
```

---

## Task 4: Create Team Mode Install Script

**Files:**
- Create: `install-team.sh`

**Step 1: Write install script**

```bash
#!/bin/bash
# install-team.sh - Install Engram Cogitator shim for team mode
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}Installing Engram Cogitator (Team Mode)${NC}"
echo ""

# Check for EC_API_URL
if [ -z "$EC_API_URL" ]; then
    echo -e "${RED}Error: EC_API_URL environment variable is required${NC}"
    echo ""
    echo "Usage:"
    echo "  EC_API_URL=https://engram.yourcompany.com ./install-team.sh"
    exit 1
fi

echo -e "API URL: ${YELLOW}$EC_API_URL${NC}"
echo ""

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo -e "${RED}Unsupported architecture: $ARCH${NC}"; exit 1 ;;
esac

echo "Detected: $OS/$ARCH"

# Download shim
VERSION=$(curl -s https://api.github.com/repos/MereWhiplash/engram-cogitator/releases/latest | grep tag_name | cut -d '"' -f 4)
DOWNLOAD_URL="https://github.com/MereWhiplash/engram-cogitator/releases/download/${VERSION}/ec-shim_${VERSION#v}_${OS}_${ARCH}.tar.gz"

echo "Downloading ec-shim ${VERSION}..."
TEMP_DIR=$(mktemp -d)
curl -sSL "$DOWNLOAD_URL" | tar -xz -C "$TEMP_DIR"

# Install binary
INSTALL_DIR="${HOME}/.local/bin"
mkdir -p "$INSTALL_DIR"
mv "$TEMP_DIR/ec-shim" "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/ec-shim"
rm -rf "$TEMP_DIR"

echo -e "${GREEN}Installed ec-shim to $INSTALL_DIR/ec-shim${NC}"

# Check if in PATH
if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
    echo -e "${YELLOW}Warning: $INSTALL_DIR is not in your PATH${NC}"
    echo "Add this to your shell profile:"
    echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
fi

# Configure Claude Code MCP
echo ""
echo "Configuring Claude Code..."

# Check for claude CLI
if ! command -v claude &> /dev/null; then
    echo -e "${RED}Error: claude CLI not found. Please install Claude Code first.${NC}"
    exit 1
fi

# Remove existing if present
claude mcp remove engram-cogitator 2>/dev/null || true

# Add new config
claude mcp add engram-cogitator \
    --scope user \
    -- "$INSTALL_DIR/ec-shim" --api-url "$EC_API_URL"

echo ""
echo -e "${GREEN}Installation complete!${NC}"
echo ""
echo "Restart Claude Code to activate Engram Cogitator."
echo ""
echo "The following MCP tools are now available:"
echo "  - ec_add      : Add memories (decisions, learnings, patterns)"
echo "  - ec_search   : Search team memories semantically"
echo "  - ec_list     : List recent memories"
echo "  - ec_invalidate : Mark memories as outdated"
```

**Step 2: Make executable**

Run: `chmod +x install-team.sh`

**Step 3: Commit**

```bash
git add install-team.sh
git commit -m "feat: add team mode installation script"
```

---

## Task 5: Update README with Team Mode

**Files:**
- Modify: `README.md`

**Step 1: Add team mode section to README**

Add after the existing solo mode section:

```markdown
---

## Team Mode (Beta)

Deploy Engram Cogitator for your engineering team. Memories are shared and searchable across all team members and repositories.

### Quick Start (Team)

**1. Deploy to Kubernetes:**

```bash
# Add Helm repo
helm repo add engram https://merewhiplash.github.io/engram-cogitator
helm repo update

# Install with Postgres
helm install engram engram/engram-cogitator \
  --namespace engram \
  --create-namespace \
  --set storage.postgres.password=<your-password>

# Or with MongoDB Atlas
helm install engram engram/engram-cogitator \
  --namespace engram \
  --create-namespace \
  --set storage.driver=mongodb \
  --set storage.mongodb.uri="mongodb+srv://user:pass@cluster.mongodb.net"
```

**2. Install shim on developer machines:**

```bash
EC_API_URL=https://engram.yourcompany.com \
  curl -sSL https://raw.githubusercontent.com/MereWhiplash/engram-cogitator/main/install-team.sh | bash
```

**3. Restart Claude Code**

### How It Works

```
Developer A (repo-frontend)     Developer B (repo-backend)
         |                              |
    [ec-shim]                      [ec-shim]
         |                              |
         +--------- HTTPS -------------+
                     |
              [Central API]
                     |
         [Postgres/MongoDB + Ollama]
```

Each developer's shim:
- Extracts git author from local config
- Extracts repo from `git remote origin`
- Forwards all requests to central API with context

### Search Behavior

By default, `ec_search` searches **all repositories** (cross-pollination). Results show who added each memory and which repo it came from:

```json
{
  "content": "Use circuit breakers for external API calls",
  "author_name": "Alice Smith",
  "author_email": "alice@company.com",
  "repo": "myorg/backend-api",
  "type": "pattern"
}
```

### Helm Values

Key configuration options:

```yaml
# Storage backend
storage:
  driver: postgres  # or mongodb

  postgres:
    internal: true    # Deploy StatefulSet
    # OR
    internal: false   # External Postgres
    host: postgres.example.com
    password: <secret>

  mongodb:
    uri: mongodb+srv://...
    database: engram

# Scaling
api:
  replicas: 3
autoscaling:
  enabled: true
  maxReplicas: 10

# Ingress
ingress:
  enabled: true
  hosts:
    - host: engram.yourcompany.com
```

See [values.yaml](charts/engram-cogitator/values.yaml) for all options.

### Architecture

| Component | Purpose |
|-----------|---------|
| API Server | HTTP API handling memory operations |
| Ollama | Local embedding generation (nomic-embed-text) |
| Postgres/MongoDB | Persistent storage with vector search |
| Shim | MCP bridge on developer machines |

### Data Model (Team)

Memories include attribution:
- `author_name` / `author_email` - Who added it
- `repo` - Which repository (org/repo format)
- All existing fields (type, area, content, rationale)

---
```

**Step 2: Commit**

```bash
git add README.md
git commit -m "docs: add team mode documentation to README"
```

---

## Task 6: Create gh-pages Branch for Helm Repo

**Step 1: Create orphan branch**

```bash
git checkout --orphan gh-pages
git rm -rf .
echo "# Engram Cogitator Helm Repository" > README.md
touch index.yaml
git add README.md index.yaml
git commit -m "Initialize Helm repository"
git push -u origin gh-pages
git checkout main
```

**Step 2: Enable GitHub Pages**

Go to repo Settings > Pages > Source: Deploy from branch > gh-pages

**Step 3: Done**

The release workflow will automatically update the Helm repo on each release.

---

## Task 7: Create First Release

**Step 1: Tag the release**

```bash
git tag -a v1.0.0 -m "Release v1.0.0 - Team mode support"
git push origin v1.0.0
```

**Step 2: Verify release**

- Check GitHub Actions for successful build
- Verify GitHub Releases page has binaries
- Verify container images at ghcr.io
- Verify Helm repo at https://merewhiplash.github.io/engram-cogitator

---

## Summary

After Phase 6, you'll have:

**Distribution channels:**
- GitHub Releases: `ec-server`, `ec-shim` binaries
- Container Registry: `ghcr.io/merewhiplash/engram-cogitator`
- Helm Repository: `https://merewhiplash.github.io/engram-cogitator`

**Install commands:**

```bash
# Solo mode (unchanged)
curl -sSL https://raw.githubusercontent.com/MereWhiplash/engram-cogitator/main/install.sh | bash

# Team mode
EC_API_URL=https://engram.company.com \
  curl -sSL https://raw.githubusercontent.com/MereWhiplash/engram-cogitator/main/install-team.sh | bash

# Helm
helm repo add engram https://merewhiplash.github.io/engram-cogitator
helm install engram engram/engram-cogitator
```

**CI/CD:**
- Tests run on every PR
- Releases triggered by version tags
- Multi-arch container images (amd64, arm64)
- Automatic Helm chart publishing

**Project complete!**
