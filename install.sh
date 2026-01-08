#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Check for team mode flag - delegate to install-team.sh
if [ "$1" = "--team" ]; then
    # If install-team.sh exists locally, use it
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    if [ -f "$SCRIPT_DIR/install-team.sh" ]; then
        exec "$SCRIPT_DIR/install-team.sh"
    else
        # Otherwise, download and run from GitHub
        echo -e "${YELLOW}Downloading team mode installer...${NC}"
        exec bash <(curl -sSL https://raw.githubusercontent.com/MereWhiplash/engram-cogitator/main/install-team.sh)
    fi
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

# Create .engram directory if it doesn't exist
if [ ! -d ".engram" ]; then
    echo -e "${YELLOW}Creating .engram directory...${NC}"
    mkdir -p .engram
    chmod 777 .engram
fi

# Add memory.db to .gitignore if not already there
if [ -f ".gitignore" ]; then
    if ! grep -q "\.engram/memory\.db" .gitignore; then
        echo -e "${YELLOW}Adding memory.db to .gitignore...${NC}"
        echo "" >> .gitignore
        echo "# Engram Cogitator local memory" >> .gitignore
        echo ".engram/memory.db" >> .gitignore
    fi
else
    echo -e "${YELLOW}Creating .gitignore with memory.db...${NC}"
    echo "# Engram Cogitator local memory" > .gitignore
    echo ".engram/memory.db" >> .gitignore
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
echo ""
echo -e "${CYAN}=== MCP Configuration ===${NC}"
echo ""

# Build the docker command
DOCKER_CMD="docker run -i --rm --network engram-network -v \"\$(pwd)/.engram:/data\" ${EC_IMAGE} --db-path /data/memory.db --ollama-url http://engram-ollama:11434"

# Check for Claude Code CLI
if command -v claude &> /dev/null; then
    echo "Claude Code CLI detected."
    read -p "Configure Claude Code automatically? [Y/n] " -n 1 -r
    echo ""

    if [[ ! $REPLY =~ ^[Nn]$ ]]; then
        claude mcp remove engram-cogitator 2>/dev/null || true
        claude mcp add --transport stdio engram-cogitator \
          --scope project \
          -- docker run -i --rm \
          --network engram-network \
          -v "$(pwd)/.engram:/data" \
          "${EC_IMAGE}" \
          --db-path /data/memory.db \
          --ollama-url http://engram-ollama:11434
        echo -e "${GREEN}Claude Code configured! (.mcp.json created)${NC}"
    fi
    echo ""
fi

# Always output generic MCP config
echo -e "${CYAN}For other MCP clients (Cursor, Cline, Windsurf, etc.):${NC}"
echo ""
echo "Add this to your MCP configuration file:"
echo ""
cat << EOF
{
  "mcpServers": {
    "engram-cogitator": {
      "command": "docker",
      "args": [
        "run", "-i", "--rm",
        "--network", "engram-network",
        "-v", "$(pwd)/.engram:/data",
        "${EC_IMAGE}",
        "--db-path", "/data/memory.db",
        "--ollama-url", "http://engram-ollama:11434"
      ]
    }
  }
}
EOF

echo ""
echo -e "${YELLOW}Common config file locations:${NC}"
echo "  Cursor:   ~/.cursor/mcp.json"
echo "  Cline:    VS Code settings > Extensions > Cline > MCP Servers"
echo "  Windsurf: ~/.codeium/windsurf/mcp_config.json"
echo ""

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

# Download generic instructions for other AI assistants
echo -e "${YELLOW}Downloading AI assistant instructions...${NC}"
curl -sSL "${EC_RAW_URL}/INSTRUCTIONS.md" -o .engram/INSTRUCTIONS.md

echo ""
echo -e "${CYAN}=== AI Assistant Instructions ===${NC}"
echo ""
echo "For AI assistants other than Claude Code, add the contents of"
echo ".engram/INSTRUCTIONS.md to your project's instruction file:"
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
echo "Engram Cogitator is now configured for this project."
echo ""
echo "What's installed:"
echo "  - Docker containers (Ollama + EC)"
echo "  - ec:remember skill in .claude/skills/"
echo "  - Session-start hook in .claude/hooks/"
echo "  - EC section added to CLAUDE.md"
echo ""
echo "MCP tools available:"
echo "  - ec_add        : Store a memory"
echo "  - ec_search     : Find relevant memories"
echo "  - ec_list       : List recent memories"
echo "  - ec_invalidate : Soft-delete a memory"
echo ""
echo -e "${YELLOW}Restart your AI coding assistant to activate.${NC}"
echo "To start Ollama: docker start engram-ollama"
echo ""
