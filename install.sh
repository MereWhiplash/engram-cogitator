#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Team mode installation
install_team_mode() {
    echo -e "${GREEN}╔═══════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║  Engram Cogitator - Team Mode Install     ║${NC}"
    echo -e "${GREEN}╚═══════════════════════════════════════════╝${NC}"
    echo ""

    if [ -z "$EC_API_URL" ]; then
        echo -e "${RED}Error: EC_API_URL environment variable required for team mode${NC}"
        echo "Usage: EC_API_URL=https://engram.company.com ./install.sh --team"
        exit 1
    fi

    # Check for Claude Code CLI
    if ! command -v claude &> /dev/null; then
        echo -e "${RED}Error: Claude Code CLI is not installed.${NC}"
        echo "Please install Claude Code first: https://docs.anthropic.com/en/docs/claude-code"
        exit 1
    fi

    # Determine OS and architecture
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    case $ARCH in
        x86_64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
    esac

    # Download shim binary
    echo -e "${YELLOW}Downloading shim binary...${NC}"
    SHIM_URL="https://github.com/MereWhiplash/engram-cogitator/releases/latest/download/ec-shim-${OS}-${ARCH}"
    SHIM_PATH="${HOME}/.local/bin/ec-shim"
    mkdir -p "$(dirname "$SHIM_PATH")"

    if curl -sSL "$SHIM_URL" -o "$SHIM_PATH" 2>/dev/null; then
        chmod +x "$SHIM_PATH"
    else
        echo -e "${RED}Error: Failed to download shim binary.${NC}"
        echo "You may need to build from source: go build ./cmd/shim"
        exit 1
    fi

    # Check if ~/.local/bin is in PATH
    if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
        echo -e "${YELLOW}Warning: ${HOME}/.local/bin is not in your PATH.${NC}"
        echo "Add it to your shell profile: export PATH=\"\$HOME/.local/bin:\$PATH\""
    fi

    # Remove existing config if present
    claude mcp remove engram-cogitator 2>/dev/null || true

    # Configure MCP with shim
    echo -e "${YELLOW}Configuring MCP server (team mode)...${NC}"
    claude mcp add engram-cogitator \
        --scope user \
        -- "$SHIM_PATH" --api-url "$EC_API_URL"

    echo ""
    echo -e "${GREEN}╔═══════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║     Team Mode Installation Complete!      ║${NC}"
    echo -e "${GREEN}╚═══════════════════════════════════════════╝${NC}"
    echo ""
    echo "Shim installed to: $SHIM_PATH"
    echo "API URL: $EC_API_URL"
    echo ""
    echo -e "${YELLOW}Restart Claude Code to activate.${NC}"
    echo ""
}

# Check for team mode flag
if [ "$1" = "--team" ]; then
    install_team_mode
    exit 0
fi

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

# Check for Claude Code CLI
if ! command -v claude &> /dev/null; then
    echo -e "${RED}Error: Claude Code CLI is not installed.${NC}"
    echo "Please install Claude Code first: https://docs.anthropic.com/en/docs/claude-code"
    exit 1
fi

# Create .claude directory if it doesn't exist
if [ ! -d ".claude" ]; then
    echo -e "${YELLOW}Creating .claude directory...${NC}"
    mkdir -p .claude
    chmod 777 .claude
fi

# Add memory.db to .gitignore if not already there
if [ -f ".gitignore" ]; then
    if ! grep -q "\.claude/memory\.db" .gitignore; then
        echo -e "${YELLOW}Adding memory.db to .gitignore...${NC}"
        echo "" >> .gitignore
        echo "# Engram Cogitator local memory" >> .gitignore
        echo ".claude/memory.db" >> .gitignore
    fi
else
    echo -e "${YELLOW}Creating .gitignore with memory.db...${NC}"
    echo "# Engram Cogitator local memory" > .gitignore
    echo ".claude/memory.db" >> .gitignore
fi

# Pull images
echo -e "${YELLOW}Pulling Ollama image...${NC}"
docker pull ${OLLAMA_IMAGE}

echo -e "${YELLOW}Pulling Engram Cogitator image...${NC}"
docker pull ${EC_IMAGE} 2>/dev/null || {
    echo -e "${YELLOW}Image not found in registry, will build locally...${NC}"
    EC_IMAGE="engram-cogitator:local"
}

# Configure MCP server
echo -e "${YELLOW}Configuring MCP server...${NC}"

# Remove existing config if present
claude mcp remove engram-cogitator 2>/dev/null || true

# Add MCP server using CLI
claude mcp add --transport stdio engram-cogitator \
  --scope project \
  -- docker run -i --rm \
  --network engram-network \
  -v "$(pwd)/.claude:/data" \
  "${EC_IMAGE}" \
  --db-path /data/memory.db \
  --ollama-url http://engram-ollama:11434

echo -e "${GREEN}MCP configuration added to .mcp.json${NC}"

# Base URL for raw files
EC_RAW_URL="https://raw.githubusercontent.com/MereWhiplash/engram-cogitator/main"

# Install EC skill
echo -e "${YELLOW}Installing EC skill...${NC}"
mkdir -p .claude/skills/ec-remember
curl -sSL "${EC_RAW_URL}/claude/skills/ec-remember/SKILL.md" \
    -o .claude/skills/ec-remember/SKILL.md

# Install session-start hook
echo -e "${YELLOW}Installing session-start hook...${NC}"
mkdir -p .claude/hooks
curl -sSL "${EC_RAW_URL}/claude/hooks/ec-session-start.sh" \
    -o .claude/hooks/ec-session-start.sh
chmod +x .claude/hooks/ec-session-start.sh

# Configure hooks in settings.json
echo -e "${YELLOW}Configuring hooks...${NC}"
if [ -f ".claude/settings.json" ]; then
    echo -e "${YELLOW}Note: .claude/settings.json exists. You may need to manually merge EC hooks.${NC}"
    echo -e "${YELLOW}See: ${EC_RAW_URL}/claude/settings.json${NC}"
else
    curl -sSL "${EC_RAW_URL}/claude/settings.json" \
        -o .claude/settings.json
fi

# Add EC section to CLAUDE.md
if [ -f "CLAUDE.md" ]; then
    if ! grep -q "Engram Cogitator" CLAUDE.md; then
        echo -e "${YELLOW}Adding EC section to CLAUDE.md...${NC}"
        curl -sSL "${EC_RAW_URL}/claude/CLAUDE.md.snippet" >> CLAUDE.md
    fi
else
    echo -e "${YELLOW}Creating CLAUDE.md with EC section...${NC}"
    curl -sSL "${EC_RAW_URL}/claude/CLAUDE.md.snippet" > CLAUDE.md
fi

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
echo "Engram Cogitator is now configured for this project."
echo ""
echo "What's installed:"
echo "  - MCP server config in .mcp.json"
echo "  - ec:remember skill in .claude/skills/"
echo "  - Session-start hook in .claude/hooks/"
echo "  - EC section added to CLAUDE.md"
echo ""
echo "MCP tools available:"
echo "  - ec_add       : Store a memory"
echo "  - ec_search    : Find relevant memories"
echo "  - ec_list      : List recent memories"
echo "  - ec_invalidate: Soft-delete a memory"
echo ""
echo -e "${YELLOW}Restart Claude Code to activate.${NC}"
echo "To start Ollama: docker start engram-ollama"
echo ""
