#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

ENGRAM_DIR="$HOME/.engram"
EC_RAW_URL="https://raw.githubusercontent.com/MereWhiplash/engram-cogitator/main"
CUSTOM_DB_PATH=""

# Show help
show_help() {
    echo "Engram Cogitator Installer"
    echo ""
    echo "Usage: install.sh [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  (no args)        Install or update (auto-detected)"
    echo "  --db-path PATH   Database connection (auto-detects type from value)"
    echo "                   Useful for team mode with shared storage"
    echo "  --init           Initialize current project only (deprecated, use marketplace)"
    echo "  --team           Team mode installation (Kubernetes)"
    echo "  --help           Show this help message"
    echo ""
    echo "Examples:"
    echo "  ./install.sh                                                    # Default (SQLite at ~/.engram/memory.db)"
    echo "  ./install.sh --db-path /shared/team/memory.db                   # SQLite at custom path"
    echo "  ./install.sh --db-path postgres://user:pass@host:5432/engram    # PostgreSQL"
    echo "  ./install.sh --db-path mongodb://host:27017/engram              # MongoDB"
    echo ""
    echo "After installation, install the cogitation plugin:"
    echo "  /plugin marketplace add MereWhiplash/engram-cogitator"
    echo "  /plugin install cogitation@engram-cogitator"
    echo ""
}

# Initialize project (deprecated - now uses marketplace)
init_project() {
    echo -e "${GREEN}╔═══════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║     Engram Cogitator - Project Init       ║${NC}"
    echo -e "${GREEN}╚═══════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${YELLOW}Note: --init is deprecated. Use the plugin marketplace instead:${NC}"
    echo ""
    echo "  1. Add the marketplace:"
    echo "     /plugin marketplace add MereWhiplash/engram-cogitator"
    echo ""
    echo "  2. Install the cogitation plugin:"
    echo "     /plugin install cogitation@engram-cogitator"
    echo ""
    echo "  3. Initialize your project:"
    echo "     /cogitation:init"
    echo ""
    echo "The cogitation plugin includes all EC skills plus opinionated"
    echo "development workflows (TDD, debugging, planning, etc.)."
    echo ""

    # Still add CLAUDE.md snippet for EC tool documentation
    if [ -f "CLAUDE.md" ]; then
        if ! grep -q "Engram Cogitator" CLAUDE.md; then
            echo -e "${YELLOW}Adding EC section to CLAUDE.md...${NC}"
            curl -sSL "${EC_RAW_URL}/claude/CLAUDE.md.snippet" >> CLAUDE.md
            echo -e "${GREEN}Added EC documentation to CLAUDE.md${NC}"
        fi
    else
        echo -e "${YELLOW}Creating CLAUDE.md with EC section...${NC}"
        curl -sSL "${EC_RAW_URL}/claude/CLAUDE.md.snippet" > CLAUDE.md
        echo -e "${GREEN}Created CLAUDE.md with EC documentation${NC}"
    fi

    echo ""
    exit 0
}

# Parse flags
while [ $# -gt 0 ]; do
    case "$1" in
        --help|-h)
            show_help
            exit 0
            ;;
        --init)
            init_project
            ;;
        --team)
            # Delegate to install-team.sh
            SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
            if [ -f "$SCRIPT_DIR/install-team.sh" ]; then
                exec "$SCRIPT_DIR/install-team.sh"
            else
                echo -e "${YELLOW}Downloading team mode installer...${NC}"
                exec bash <(curl -sSL https://raw.githubusercontent.com/MereWhiplash/engram-cogitator/main/install-team.sh)
            fi
            ;;
        --db-path)
            shift
            if [ -z "$1" ] || [[ "$1" == --* ]]; then
                echo -e "${RED}Error: --db-path requires a path argument${NC}"
                exit 1
            fi
            CUSTOM_DB_PATH="$1"
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            show_help
            exit 1
            ;;
    esac
    shift
done

EC_VERSION="latest"
EC_IMAGE="ghcr.io/merewhiplash/engram-cogitator:${EC_VERSION}"
OLLAMA_IMAGE="ollama/ollama:latest"
EMBEDDING_MODEL="nomic-embed-text"

# Solo default is the shared-api model: one singleton ec-api container + a native
# ec-shim per session. EC_LAUNCHER flips to ec-ensure-api.sh once a native shim is
# acquired; otherwise we fall back to the per-session ec-run.sh wrapper.
API_NAME="engram-cogitator-api"
EC_LAUNCHER="ec-run.sh"

# Detect host OS/arch for native shim selection (the shim runs on the host, not in docker).
detect_platform() {
    SHIM_OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    SHIM_ARCH=$(uname -m)
    case "$SHIM_ARCH" in
        x86_64) SHIM_ARCH="amd64" ;;
        aarch64|arm64) SHIM_ARCH="arm64" ;;
    esac
}

# Resolve the latest release tag (ported from install-team.sh). Honors GITHUB_TOKEN.
fetch_shim_version() {
    local auth_header=""
    if [ -n "${GITHUB_TOKEN:-}" ]; then
        auth_header="-H \"Authorization: token $GITHUB_TOKEN\""
    fi
    local response
    response=$(eval curl -s "$auth_header" https://api.github.com/repos/MereWhiplash/engram-cogitator/releases/latest)
    echo "$response" | grep tag_name | cut -d '"' -f 4
}

# Assert the acquired shim is a NATIVE host binary (image /ec-shim is a Linux ELF and
# will not execute on macOS/Windows). Mach-O on darwin, ELF on linux.
assert_native_shim() {
    local info
    info=$(file "$1" 2>/dev/null || echo "")
    case "$SHIM_OS" in
        darwin) echo "$info" | grep -qi "mach-o" ;;
        linux)  echo "$info" | grep -qi "elf" ;;
        *)      [ -s "$1" ] ;;
    esac
}

# Acquire a native host ec-shim into ~/.engram/ec-shim.
# Order: prebuilt release -> local go build (no CGO) -> docker cp (linux only).
acquire_shim() {
    local shim="$ENGRAM_DIR/ec-shim"
    detect_platform

    # 1) Prebuilt release (primary)
    local ver
    ver="${SHIM_VERSION:-$(fetch_shim_version)}"
    if [ -n "$ver" ]; then
        local url="https://github.com/MereWhiplash/engram-cogitator/releases/download/${ver}/ec-shim_${ver#v}_${SHIM_OS}_${SHIM_ARCH}.tar.gz"
        local tmp
        tmp=$(mktemp -d)
        if curl -sSL "$url" 2>/dev/null | tar -xz -C "$tmp" 2>/dev/null && [ -f "$tmp/ec-shim" ]; then
            mv "$tmp/ec-shim" "$shim" && chmod +x "$shim"
            rm -rf "$tmp"
            if assert_native_shim "$shim"; then
                echo -e "${GREEN}Installed ec-shim from release ${ver}${NC}"
                return 0
            fi
            echo -e "${YELLOW}Release shim is not native for ${SHIM_OS}/${SHIM_ARCH}; trying local build...${NC}"
        else
            rm -rf "$tmp"
        fi
    fi

    # 2) Local build (fallback; the shim has no CGO deps)
    if command -v go >/dev/null && [ -d "$SCRIPT_DIR/cmd/shim" ]; then
        echo -e "${YELLOW}Building ec-shim from source...${NC}"
        if (cd "$SCRIPT_DIR" && CGO_ENABLED=0 go build -o "$shim" ./cmd/shim) && chmod +x "$shim" && assert_native_shim "$shim"; then
            echo -e "${GREEN}Built native ec-shim from source${NC}"
            return 0
        fi
    fi

    # 3) docker cp from image (Linux hosts only — image /ec-shim is a Linux ELF)
    if [ "$SHIM_OS" = "linux" ]; then
        echo -e "${YELLOW}Extracting ec-shim from image...${NC}"
        local cid
        cid=$(docker create "$EC_IMAGE" 2>/dev/null) || true
        if [ -n "$cid" ]; then
            docker cp "$cid:/ec-shim" "$shim" &>/dev/null && chmod +x "$shim"
            docker rm "$cid" &>/dev/null || true
            assert_native_shim "$shim" && { echo -e "${GREEN}Extracted ec-shim from image${NC}"; return 0; }
        fi
    fi

    return 1
}

# Start the singleton shared api (after ollama is up). Mirrors the ollama block's
# idempotency: running -> noop, stopped -> start, missing -> docker run.
start_shared_api() {
    local db_path="$ENGRAM_DIR/memory.db"
    # shellcheck source=/dev/null
    [ -f "$ENGRAM_DIR/config" ] && . "$ENGRAM_DIR/config"
    [ -n "${EC_DB_PATH:-}" ] && db_path="$EC_DB_PATH"

    if docker ps --format '{{.Names}}' | grep -q "^${API_NAME}$"; then
        echo -e "${GREEN}Shared api already running.${NC}"
        return 0
    fi
    if docker ps -a --format '{{.Names}}' | grep -q "^${API_NAME}$"; then
        echo -e "${YELLOW}Starting existing shared api...${NC}"
        docker start "$API_NAME" &>/dev/null
        return 0
    fi
    echo -e "${YELLOW}Starting shared api (${API_NAME})...${NC}"
    docker run -d \
        --name "$API_NAME" \
        --restart unless-stopped \
        --network engram-network \
        -p "127.0.0.1:8080:8080" \
        -v "$ENGRAM_DIR:/data" \
        "$EC_IMAGE" \
        --addr ":8080" \
        --storage-driver sqlite \
        --db-path "/data/$(basename "$db_path")" \
        --ollama-url http://engram-ollama:11434 &>/dev/null
}

# Detect if this is an update (wrapper + engram dir already exist)
IS_UPDATE=false
if [ -f "$ENGRAM_DIR/ec-run.sh" ] && [ -d "$ENGRAM_DIR" ]; then
    IS_UPDATE=true
fi

if [ "$IS_UPDATE" = true ]; then
    echo -e "${GREEN}╔═══════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║     Engram Cogitator - Update             ║${NC}"
    echo -e "${GREEN}╚═══════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${CYAN}Existing installation detected. Updating...${NC}"
else
    echo -e "${GREEN}╔═══════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║     Engram Cogitator - Installation       ║${NC}"
    echo -e "${GREEN}╚═══════════════════════════════════════════╝${NC}"
fi
echo ""

# Check for Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Error: Docker is not installed.${NC}"
    echo "Please install Docker first: https://docs.docker.com/get-docker/"
    exit 1
fi

# Check if Docker daemon is running
if ! docker info &> /dev/null; then
    echo -e "${RED}Error: Docker daemon is not running.${NC}"
    echo "Please start Docker and try again."
    exit 1
fi

# Create global .engram directory if it doesn't exist
if [ ! -d "$ENGRAM_DIR" ]; then
    echo -e "${YELLOW}Creating ~/.engram directory...${NC}"
    mkdir -p "$ENGRAM_DIR"
fi

# Save custom DB config if provided (auto-detect driver from value)
if [ -n "$CUSTOM_DB_PATH" ]; then
    case "$CUSTOM_DB_PATH" in
        postgres://*|postgresql://*)
            echo -e "${YELLOW}Detected PostgreSQL connection${NC}"
            echo -e "${YELLOW}  DSN: ${CUSTOM_DB_PATH}${NC}"
            cat > "$ENGRAM_DIR/config" <<CONF
EC_STORAGE_DRIVER=postgres
EC_POSTGRES_DSN=${CUSTOM_DB_PATH}
CONF
            ;;
        mongodb://*|mongodb+srv://*)
            echo -e "${YELLOW}Detected MongoDB connection${NC}"
            echo -e "${YELLOW}  URI: ${CUSTOM_DB_PATH}${NC}"
            cat > "$ENGRAM_DIR/config" <<CONF
EC_STORAGE_DRIVER=mongodb
EC_MONGODB_URI=${CUSTOM_DB_PATH}
CONF
            ;;
        *)
            echo -e "${YELLOW}Using SQLite database: ${CUSTOM_DB_PATH}${NC}"
            CUSTOM_DB_DIR="$(dirname "$CUSTOM_DB_PATH")"
            if [ ! -d "$CUSTOM_DB_DIR" ]; then
                echo -e "${YELLOW}Creating database directory: ${CUSTOM_DB_DIR}${NC}"
                mkdir -p "$CUSTOM_DB_DIR"
            fi
            cat > "$ENGRAM_DIR/config" <<CONF
EC_STORAGE_DRIVER=sqlite
EC_DB_PATH=${CUSTOM_DB_PATH}
CONF
            ;;
    esac
    echo -e "${GREEN}Saved to ~/.engram/config${NC}"
elif [ -f "$ENGRAM_DIR/config" ]; then
    # Preserve existing config on updates
    echo -e "${CYAN}Using existing config from ~/.engram/config${NC}"
fi

# Deploy container lifecycle wrapper script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ -f "$SCRIPT_DIR/scripts/ec-run.sh" ]; then
    echo -e "${YELLOW}Updating container lifecycle wrapper...${NC}"
    cp "$SCRIPT_DIR/scripts/ec-run.sh" "$ENGRAM_DIR/ec-run.sh"
    chmod +x "$ENGRAM_DIR/ec-run.sh"
    echo -e "${GREEN}Installed ~/.engram/ec-run.sh${NC}"
else
    echo -e "${YELLOW}Downloading container lifecycle wrapper...${NC}"
    curl -sSL "${EC_RAW_URL}/scripts/ec-run.sh" -o "$ENGRAM_DIR/ec-run.sh"
    chmod +x "$ENGRAM_DIR/ec-run.sh"
    echo -e "${GREEN}Installed ~/.engram/ec-run.sh${NC}"
fi

# Deploy the shared-api singleton launcher (solo default)
if [ -f "$SCRIPT_DIR/scripts/ec-ensure-api.sh" ]; then
    cp "$SCRIPT_DIR/scripts/ec-ensure-api.sh" "$ENGRAM_DIR/ec-ensure-api.sh"
else
    curl -sSL "${EC_RAW_URL}/scripts/ec-ensure-api.sh" -o "$ENGRAM_DIR/ec-ensure-api.sh"
fi
chmod +x "$ENGRAM_DIR/ec-ensure-api.sh"
echo -e "${GREEN}Installed ~/.engram/ec-ensure-api.sh${NC}"

# Pull EC image (always — this is the main thing an update does)
echo -e "${YELLOW}Pulling Engram Cogitator image...${NC}"
docker pull ${EC_IMAGE} 2>/dev/null || {
    echo -e "${YELLOW}Image not found in registry, will build locally...${NC}"
    EC_IMAGE="engram-cogitator:local"
}

# The shared-api model solves the local SQLite multi-writer problem. Postgres/MongoDB
# solo users already point at external shared storage (no local contention) and the
# launcher is sqlite-only — keep them on ec-run.sh to avoid a silent driver switch.
CONFIGURED_DRIVER="sqlite"
if [ -f "$ENGRAM_DIR/config" ]; then
    CONFIGURED_DRIVER="$(grep -E '^EC_STORAGE_DRIVER=' "$ENGRAM_DIR/config" | cut -d '=' -f2)"
    CONFIGURED_DRIVER="${CONFIGURED_DRIVER:-sqlite}"
fi

# Acquire a native ec-shim (sqlite only). On success, register the shared-api
# launcher; otherwise fall back to the per-session ec-run.sh wrapper.
if [ "$CONFIGURED_DRIVER" != "sqlite" ]; then
    echo -e "${CYAN}Storage driver is ${CONFIGURED_DRIVER}; keeping per-session server (ec-run.sh).${NC}"
    EC_LAUNCHER="ec-run.sh"
else
    echo -e "${YELLOW}Acquiring native ec-shim...${NC}"
    if acquire_shim; then
        EC_LAUNCHER="ec-ensure-api.sh"
    else
        echo -e "${YELLOW}Could not acquire a native ec-shim — falling back to per-session server (ec-run.sh).${NC}"
        EC_LAUNCHER="ec-run.sh"
    fi
fi

# --- Fresh install only: full setup ---
if [ "$IS_UPDATE" = false ]; then

    # Detect AI tooling
    echo -e "${CYAN}=== AI Tooling Detection ===${NC}"
    echo ""

    DETECTED_TOOLS=""
    if [ -d ".claude" ] || command -v claude &> /dev/null; then
        DETECTED_TOOLS="${DETECTED_TOOLS}claude "
        echo -e "  ${GREEN}✓${NC} Claude Code detected"
    fi
    if [ -d ".cursor" ]; then
        DETECTED_TOOLS="${DETECTED_TOOLS}cursor "
        echo -e "  ${GREEN}✓${NC} Cursor detected"
    fi
    if [ -d ".github/copilot" ]; then
        DETECTED_TOOLS="${DETECTED_TOOLS}copilot "
        echo -e "  ${GREEN}✓${NC} GitHub Copilot detected"
    fi

    if [ -z "$DETECTED_TOOLS" ]; then
        echo -e "  ${YELLOW}No AI tooling detected in current directory${NC}"
    fi
    echo ""

    # Pull Ollama image (only on fresh install — user manages Ollama updates separately)
    echo -e "${YELLOW}Pulling Ollama image...${NC}"
    docker pull ${OLLAMA_IMAGE}

    # Configure MCP server
    echo ""
    echo -e "${CYAN}=== MCP Configuration ===${NC}"
    echo ""
    case "$CUSTOM_DB_PATH" in
        postgres://*|postgresql://*)
            echo -e "${YELLOW}Database:${NC} PostgreSQL (${CUSTOM_DB_PATH})"
            ;;
        mongodb://*|mongodb+srv://*)
            echo -e "${YELLOW}Database:${NC} MongoDB (${CUSTOM_DB_PATH})"
            ;;
        "")
            echo -e "${YELLOW}Database:${NC} SQLite at ~/.engram/memory.db (default)"
            ;;
        *)
            echo -e "${YELLOW}Database:${NC} SQLite at ${CUSTOM_DB_PATH}"
            ;;
    esac
    echo -e "${YELLOW}Project identity:${NC} Auto-detected from git remote or directory path"
    echo ""

    # Check for Claude Code CLI
    if command -v claude &> /dev/null; then
        echo "Claude Code CLI detected."
        read -p "Configure Claude Code automatically? [Y/n] " -n 1 -r
        echo ""

        if [[ ! $REPLY =~ ^[Nn]$ ]]; then
            claude mcp remove engram-cogitator --scope user 2>/dev/null || true
            # Register the chosen launcher (shared-api singleton, or per-session fallback)
            claude mcp add --transport stdio engram-cogitator \
              --scope user \
              -- /bin/sh -c "\$HOME/.engram/${EC_LAUNCHER}"
            echo -e "${GREEN}Claude Code configured globally!${NC}"
        fi
        echo ""
    fi

    # Always output generic MCP config
    echo -e "${CYAN}For Cursor / VS Code (supports \${workspaceFolder}):${NC}"
    echo ""
    echo "Add to ~/.cursor/mcp.json or VS Code MCP settings:"
    echo ""
    cat << EOF
{
  "mcpServers": {
    "engram-cogitator": {
      "command": "/bin/sh",
      "args": ["-c", "\$HOME/.engram/${EC_LAUNCHER}"]
    }
  }
}
EOF

    echo ""
    echo -e "${YELLOW}Config file locations:${NC}"
    echo "  Cursor (global):   ~/.cursor/mcp.json"
    echo "  Cursor (project):  .cursor/mcp.json"
    echo "  VS Code/Copilot:   VS Code settings > MCP Servers"
    echo ""

    # Download generic instructions
    echo -e "${YELLOW}Downloading AI assistant instructions...${NC}"
    curl -sSL "${EC_RAW_URL}/INSTRUCTIONS.md" -o "$ENGRAM_DIR/INSTRUCTIONS.md"

    echo ""
    echo -e "${CYAN}=== AI Assistant Instructions ===${NC}"
    echo ""
    echo "For AI assistants other than Claude Code, add the contents of"
    echo "~/.engram/INSTRUCTIONS.md to your project's instruction file:"
    echo ""
    echo "  Cursor:        .cursor/rules/engram.mdc"
    echo "  GitHub Copilot: .github/copilot-instructions.md"
    echo "  Gemini:        GEMINI.md"
    echo "  Generic:       AGENTS.md"
    echo ""

else
    # --- Update path: re-register MCP against the shared-api launcher ---
    if command -v claude &> /dev/null; then
        echo -e "${YELLOW}Updating Claude Code MCP registration...${NC}"
        claude mcp remove engram-cogitator --scope user 2>/dev/null || true
        claude mcp add --transport stdio engram-cogitator \
          --scope user \
          -- /bin/sh -c "\$HOME/.engram/${EC_LAUNCHER}"
        echo -e "${GREEN}Claude Code MCP updated.${NC}"
    fi

    # Old per-session servers are left to DRAIN (not force-killed): a window where
    # both they and the new shared api hold memory.db open. Old connections lack
    # busy_timeout, so a rare write collision could surface "database is locked" for
    # them; WAL is a persistent DB-header setting, so they still operate correctly.
    echo -e "${YELLOW}Note: existing per-session EC servers will exit as those Claude windows close.${NC}"
    echo -e "${YELLOW}      For a clean cutover, close other Claude windows before upgrading.${NC}"

    # Update instructions file
    echo -e "${YELLOW}Updating AI assistant instructions...${NC}"
    curl -sSL "${EC_RAW_URL}/INSTRUCTIONS.md" -o "$ENGRAM_DIR/INSTRUCTIONS.md"

fi

# --- Shared steps (both install and update) ---

# Add EC documentation to CLAUDE.md (if Claude Code project detected)
if [ -d ".claude" ] || [ -f "CLAUDE.md" ]; then
    if [ -f "CLAUDE.md" ]; then
        if ! grep -q "Engram Cogitator" CLAUDE.md; then
            echo -e "${YELLOW}Adding EC section to CLAUDE.md...${NC}"
            curl -sSL "${EC_RAW_URL}/claude/CLAUDE.md.snippet" >> CLAUDE.md
        fi
    else
        echo -e "${YELLOW}Creating CLAUDE.md with EC section...${NC}"
        curl -sSL "${EC_RAW_URL}/claude/CLAUDE.md.snippet" > CLAUDE.md
    fi
fi

# Create Docker network if it doesn't exist
if ! docker network inspect engram-network &> /dev/null; then
    echo -e "${YELLOW}Creating Docker network...${NC}"
    docker network create engram-network
fi

# Clean up orphaned EC containers (labeled or legacy unlabeled)
echo -e "${YELLOW}Checking for orphaned EC containers...${NC}"
LABELED_STALE=$(docker ps -a \
    --filter "label=io.engram-cogitator.role=mcp-server" \
    --filter "status=exited" --filter "status=dead" --filter "status=created" \
    --format '{{.ID}} {{.Names}}' 2>/dev/null) || true
# Exclude the singleton shared api (engram-cogitator-api) — it is long-lived, not orphaned.
ORPHANED=$(docker ps -a --filter "ancestor=${EC_IMAGE}" --format '{{.ID}} {{.Names}}' 2>/dev/null | grep -v 'ec-mcp-\|ec-hook-\|engram-cogitator-api' || true)
ALL_STALE=$(printf '%s\n%s' "$LABELED_STALE" "$ORPHANED" | sort -u | grep -v '^$' || true)
if [ -n "$ALL_STALE" ]; then
    echo -e "${YELLOW}Found stale EC containers:${NC}"
    echo "$ALL_STALE" | while read -r cid cname; do
        echo -e "  Removing: ${cname} (${cid})"
        docker rm -f "$cid" &>/dev/null || true
    done
    echo -e "${GREEN}Stale containers cleaned up.${NC}"
else
    echo -e "${GREEN}No stale containers found.${NC}"
fi
echo ""

# Start Ollama container if not running
if ! docker ps --format '{{.Names}}' | grep -q '^engram-ollama$'; then
    echo -e "${YELLOW}Starting Ollama container...${NC}"
    # Check if stopped container exists
    if docker ps -a --format '{{.Names}}' | grep -q '^engram-ollama$'; then
        docker start engram-ollama
    else
        docker run -d \
            --name engram-ollama \
            --network engram-network \
            -v ollama_data:/root/.ollama \
            ${OLLAMA_IMAGE}
    fi

    echo -e "${YELLOW}Waiting for Ollama to start...${NC}"
    sleep 5
fi

# Pull embedding model
echo -e "${YELLOW}Pulling embedding model (${EMBEDDING_MODEL})...${NC}"
docker exec engram-ollama ollama pull ${EMBEDDING_MODEL}

# Start the singleton shared api (only when registered against the shared-api launcher).
# Per-session sessions otherwise lazily start it via ec-ensure-api.sh.
if [ "$EC_LAUNCHER" = "ec-ensure-api.sh" ]; then
    start_shared_api
fi

echo ""
if [ "$IS_UPDATE" = true ]; then
    echo -e "${GREEN}╔═══════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║     Update Complete!                      ║${NC}"
    echo -e "${GREEN}╚═══════════════════════════════════════════╝${NC}"
    echo ""
    echo "Updated:"
    echo "  - EC Docker image (latest)"
    echo "  - Shared-api launcher (~/.engram/ec-ensure-api.sh) + native ec-shim"
    echo "  - Per-session fallback wrapper (~/.engram/ec-run.sh)"
    echo "  - MCP server registration → ${EC_LAUNCHER}"
    if [ -d ".claude" ] || [ -f "CLAUDE.md" ]; then
    echo "  - EC section in CLAUDE.md"
    fi
    echo ""
    echo -e "${YELLOW}Restart your AI coding assistant to use the new version.${NC}"
else
    echo -e "${GREEN}╔═══════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║     Installation Complete!                ║${NC}"
    echo -e "${GREEN}╚═══════════════════════════════════════════╝${NC}"
    echo ""
    echo "Engram Cogitator MCP server is now configured."
    echo ""
    echo "What's installed:"
    echo "  - Docker containers (Ollama + shared engram-cogitator-api)"
    echo "  - Native ec-shim per session (~/.engram/ec-shim)"
    case "$CUSTOM_DB_PATH" in
        postgres://*|postgresql://*)
    echo "  - PostgreSQL database"
            ;;
        mongodb://*|mongodb+srv://*)
    echo "  - MongoDB database"
            ;;
        "")
    echo "  - Global storage at ~/.engram/memory.db"
            ;;
        *)
    echo "  - SQLite database at ${CUSTOM_DB_PATH}"
            ;;
    esac
    if [ -d ".claude" ] || [ -f "CLAUDE.md" ]; then
    echo "  - EC section in CLAUDE.md"
    fi
    echo ""
    echo "MCP tools available:"
    echo "  - ec_add        : Store a memory"
    echo "  - ec_search     : Find relevant memories"
    echo "  - ec_list       : List recent memories"
    echo "  - ec_invalidate : Soft-delete a memory"
    echo ""
    echo -e "${CYAN}=== Install Cogitation Plugin (Recommended) ===${NC}"
    echo ""
    echo "The cogitation plugin provides opinionated development workflows"
    echo "that leverage EC's persistent memory (TDD, debugging, planning, etc.)"
    echo ""
    echo "In Claude Code, run:"
    echo -e "  ${YELLOW}/plugin marketplace add MereWhiplash/engram-cogitator${NC}"
    echo -e "  ${YELLOW}/plugin install cogitation@engram-cogitator${NC}"
    echo ""
    echo "Then initialize your project:"
    echo -e "  ${YELLOW}/cogitation:init${NC}"
    echo ""
    echo -e "${YELLOW}Restart your AI coding assistant to activate the MCP server.${NC}"
fi
echo ""
echo "Commands:"
echo "  docker start engram-ollama     Start Ollama if stopped"
echo "  ./install.sh                   Update to latest version"
echo "  ./uninstall.sh                 Remove EC from this project"
echo ""
