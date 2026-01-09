#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

ENGRAM_DIR="$HOME/.engram"

show_help() {
    echo "Engram Cogitator Uninstaller"
    echo ""
    echo "Usage: uninstall.sh [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --project   Remove project-level config only (default)"
    echo "              Removes: .claude/skills/ec-remember, .claude/hooks/ec-session-start.sh"
    echo "              Keeps: Docker containers, global ~/.engram/memory.db"
    echo ""
    echo "  --global    Remove Docker containers and network"
    echo "              Removes: engram-ollama container, engram-network, ollama_data volume"
    echo "              Keeps: ~/.engram/memory.db (your memories)"
    echo ""
    echo "  --all       Remove everything including memories"
    echo "              Removes: All of the above + ~/.engram/ directory"
    echo "              WARNING: This deletes all your stored memories!"
    echo ""
    echo "  --help      Show this help message"
    echo ""
}

remove_project_config() {
    echo -e "${CYAN}=== Removing Project Config ===${NC}"
    echo ""

    # Remove EC skill
    if [ -d ".claude/skills/ec-remember" ]; then
        echo -e "${YELLOW}Removing .claude/skills/ec-remember...${NC}"
        rm -rf .claude/skills/ec-remember
    fi

    # Remove session-start hook
    if [ -f ".claude/hooks/ec-session-start.sh" ]; then
        echo -e "${YELLOW}Removing .claude/hooks/ec-session-start.sh...${NC}"
        rm -f .claude/hooks/ec-session-start.sh
    fi

    # Remove .mcp.json if it only contains engram-cogitator
    if [ -f ".mcp.json" ]; then
        if grep -q "engram-cogitator" .mcp.json; then
            echo -e "${YELLOW}Note: .mcp.json contains engram-cogitator config.${NC}"
            echo -e "${YELLOW}Please manually remove the engram-cogitator entry if desired.${NC}"
        fi
    fi

    # Note about CLAUDE.md
    if [ -f "CLAUDE.md" ] && grep -q "Engram Cogitator" CLAUDE.md; then
        echo -e "${YELLOW}Note: CLAUDE.md contains Engram Cogitator section.${NC}"
        echo -e "${YELLOW}Please manually remove if desired.${NC}"
    fi

    # Remove old per-project .engram directory if it exists
    if [ -d ".engram" ]; then
        echo -e "${YELLOW}Found legacy per-project .engram/ directory.${NC}"
        read -p "Remove .engram/ directory? [y/N] " -n 1 -r
        echo ""
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            rm -rf .engram
            echo -e "${GREEN}Removed .engram/${NC}"
        fi
    fi

    echo -e "${GREEN}Project config removed.${NC}"
}

remove_global_docker() {
    echo -e "${CYAN}=== Removing Docker Resources ===${NC}"
    echo ""

    # Check for Docker
    if ! command -v docker &> /dev/null; then
        echo -e "${YELLOW}Docker not found, skipping container removal.${NC}"
        return
    fi

    # Stop and remove Ollama container
    if docker ps -a --format '{{.Names}}' | grep -q '^engram-ollama$'; then
        echo -e "${YELLOW}Stopping engram-ollama container...${NC}"
        docker stop engram-ollama 2>/dev/null || true
        echo -e "${YELLOW}Removing engram-ollama container...${NC}"
        docker rm engram-ollama 2>/dev/null || true
    fi

    # Remove Docker network
    if docker network inspect engram-network &> /dev/null; then
        echo -e "${YELLOW}Removing engram-network...${NC}"
        docker network rm engram-network 2>/dev/null || true
    fi

    # Remove Ollama data volume
    if docker volume inspect ollama_data &> /dev/null; then
        read -p "Remove ollama_data volume (embedding models)? [y/N] " -n 1 -r
        echo ""
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            docker volume rm ollama_data 2>/dev/null || true
            echo -e "${GREEN}Removed ollama_data volume.${NC}"
        fi
    fi

    # Remove MCP config from Claude Code
    if command -v claude &> /dev/null; then
        echo -e "${YELLOW}Removing MCP config from Claude Code...${NC}"
        claude mcp remove engram-cogitator 2>/dev/null || true
    fi

    echo -e "${GREEN}Docker resources removed.${NC}"
}

remove_all_data() {
    echo -e "${CYAN}=== Removing All Data ===${NC}"
    echo ""

    if [ -d "$ENGRAM_DIR" ]; then
        # Count memories if possible
        if [ -f "$ENGRAM_DIR/memory.db" ]; then
            echo -e "${RED}WARNING: This will delete your memory database!${NC}"
            echo -e "${YELLOW}Database: $ENGRAM_DIR/memory.db${NC}"
        fi

        read -p "Are you sure you want to delete ~/.engram/? [y/N] " -n 1 -r
        echo ""
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            rm -rf "$ENGRAM_DIR"
            echo -e "${GREEN}Removed ~/.engram/${NC}"
        else
            echo -e "${YELLOW}Skipped ~/.engram/ removal.${NC}"
        fi
    else
        echo -e "${YELLOW}~/.engram/ not found.${NC}"
    fi
}

# Parse arguments
MODE="project"

while [[ $# -gt 0 ]]; do
    case $1 in
        --project)
            MODE="project"
            shift
            ;;
        --global)
            MODE="global"
            shift
            ;;
        --all)
            MODE="all"
            shift
            ;;
        --help|-h)
            show_help
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            show_help
            exit 1
            ;;
    esac
done

echo -e "${GREEN}╔═══════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║     Engram Cogitator - Uninstaller        ║${NC}"
echo -e "${GREEN}╚═══════════════════════════════════════════╝${NC}"
echo ""
echo -e "Mode: ${CYAN}$MODE${NC}"
echo ""

case $MODE in
    project)
        remove_project_config
        echo ""
        echo -e "${YELLOW}To also remove Docker resources: uninstall.sh --global${NC}"
        echo -e "${YELLOW}To remove everything: uninstall.sh --all${NC}"
        ;;
    global)
        remove_project_config
        remove_global_docker
        echo ""
        echo -e "${GREEN}Your memories are preserved at ~/.engram/memory.db${NC}"
        echo -e "${YELLOW}To remove everything: uninstall.sh --all${NC}"
        ;;
    all)
        remove_project_config
        remove_global_docker
        remove_all_data
        ;;
esac

echo ""
echo -e "${GREEN}Uninstall complete.${NC}"
