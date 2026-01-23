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

# Show help
show_help() {
    echo "Engram Cogitator Installer"
    echo ""
    echo "Usage: install.sh [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  (no args)   Full installation (MCP server + cogitation plugin)"
    echo "  --init      Initialize current project only (deprecated, use marketplace)"
    echo "  --team      Team mode installation (Kubernetes)"
    echo "  --help      Show this help message"
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

# Check for flags
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
esac

EC_VERSION="latest"
EC_IMAGE="ghcr.io/merewhiplash/engram-cogitator:${EC_VERSION}"
OLLAMA_IMAGE="ollama/ollama:latest"
EMBEDDING_MODEL="nomic-embed-text"

echo -e "${GREEN}╔═══════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║     Engram Cogitator - Installation       ║${NC}"
echo -e "${GREEN}╚═══════════════════════════════════════════╝${NC}"
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

# Pull images
echo -e "${YELLOW}Pulling Ollama image...${NC}"
docker pull ${OLLAMA_IMAGE}

echo -e "${YELLOW}Pulling Engram Cogitator image...${NC}"
docker pull ${EC_IMAGE} 2>/dev/null || {
    echo -e "${YELLOW}Image not found in registry, will build locally...${NC}"
    EC_IMAGE="engram-cogitator:local"
}

# Configure MCP server
echo ""
echo -e "${CYAN}=== MCP Configuration ===${NC}"
echo ""
echo -e "${YELLOW}Global mode:${NC} All memories stored in ~/.engram/memory.db"
echo -e "${YELLOW}Project identity:${NC} Auto-detected from git remote or directory path"
echo ""

# Build the docker command - note: uses global storage now
DOCKER_CMD="docker run -i --rm --network engram-network -v \"$ENGRAM_DIR:/data\" ${EC_IMAGE} --db-path /data/memory.db --ollama-url http://engram-ollama:11434"

# Check for Claude Code CLI
if command -v claude &> /dev/null; then
    echo "Claude Code CLI detected."
    read -p "Configure Claude Code automatically? [Y/n] " -n 1 -r
    echo ""

    if [[ ! $REPLY =~ ^[Nn]$ ]]; then
        claude mcp remove engram-cogitator 2>/dev/null || true
        # Use /bin/sh -c to get working directory at runtime for project detection
        # --entrypoint overrides default /ec-api to use MCP server instead
        claude mcp add --transport stdio engram-cogitator \
          --scope user \
          -- /bin/sh -c "docker run -i --rm --entrypoint /usr/local/bin/engram-cogitator --network engram-network -v \$HOME/.engram:/data ${EC_IMAGE} --db-path /data/memory.db --repo \"\$(pwd)\" --ollama-url http://engram-ollama:11434"
        echo -e "${GREEN}Claude Code configured globally!${NC}"
    fi
    echo ""
fi

# Always output generic MCP config
echo -e "${CYAN}For Cursor / VS Code (supports \${workspaceFolder}):${NC}"
echo ""
echo "Add to ~/.cursor/mcp.json or VS Code MCP settings:"
echo ""
cat << 'EOF'
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
EOF

echo ""
echo -e "${YELLOW}Config file locations:${NC}"
echo "  Cursor (global):   ~/.cursor/mcp.json"
echo "  Cursor (project):  .cursor/mcp.json"
echo "  VS Code/Copilot:   VS Code settings > MCP Servers"
echo ""

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

# Create Docker network if it doesn't exist
if ! docker network inspect engram-network &> /dev/null; then
    echo -e "${YELLOW}Creating Docker network...${NC}"
    docker network create engram-network
fi

# Start Ollama container if not running
if ! docker ps --format '{{.Names}}' | grep -q '^engram-ollama$'; then
    echo -e "${YELLOW}Starting Ollama container...${NC}"
    docker run -d \
        --name engram-ollama \
        --network engram-network \
        -v ollama_data:/root/.ollama \
        ${OLLAMA_IMAGE}

    echo -e "${YELLOW}Waiting for Ollama to start...${NC}"
    sleep 5
fi

# Pull embedding model
echo -e "${YELLOW}Pulling embedding model (${EMBEDDING_MODEL})...${NC}"
docker exec engram-ollama ollama pull ${EMBEDDING_MODEL}

echo ""
echo -e "${GREEN}╔═══════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║     Installation Complete!                ║${NC}"
echo -e "${GREEN}╚═══════════════════════════════════════════╝${NC}"
echo ""
echo "Engram Cogitator MCP server is now configured."
echo ""
echo "What's installed:"
echo "  - Docker containers (Ollama + EC)"
echo "  - Global storage at ~/.engram/memory.db"
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
echo ""
echo "Commands:"
echo "  docker start engram-ollama     Start Ollama if stopped"
echo "  ./uninstall.sh                 Remove EC from this project"
echo ""
