#!/bin/bash
# install-team.sh - Install Engram Cogitator shim for team mode
# Works with any MCP-compatible client (Claude Code, Cursor, Cline, Windsurf, etc.)
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
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

# Fetch latest version from GitHub API
# Uses GITHUB_TOKEN if available to avoid rate limits (60/hr anonymous, 5000/hr authenticated)
fetch_version() {
    local auth_header=""
    if [ -n "$GITHUB_TOKEN" ]; then
        auth_header="-H \"Authorization: token $GITHUB_TOKEN\""
    fi

    local response
    response=$(eval curl -s $auth_header https://api.github.com/repos/MereWhiplash/engram-cogitator/releases/latest)

    # Check for rate limit error
    if echo "$response" | grep -q "API rate limit exceeded"; then
        echo -e "${RED}Error: GitHub API rate limit exceeded${NC}"
        echo ""
        echo "Solutions:"
        echo "  1. Wait an hour for the rate limit to reset"
        echo "  2. Set GITHUB_TOKEN environment variable for higher limits:"
        echo "     GITHUB_TOKEN=ghp_xxx EC_API_URL=https://... ./install-team.sh"
        echo "  3. Specify version manually: EC_VERSION=v1.0.0 EC_API_URL=https://... ./install-team.sh"
        exit 1
    fi

    echo "$response" | grep tag_name | cut -d '"' -f 4
}

# Allow manual version override
if [ -n "$EC_VERSION" ]; then
    VERSION="$EC_VERSION"
    echo "Using specified version: $VERSION"
else
    echo "Fetching latest version..."
    VERSION=$(fetch_version)
fi

if [ -z "$VERSION" ]; then
    echo -e "${RED}Error: Could not determine latest version${NC}"
    echo ""
    echo "This may be due to GitHub API rate limiting or no releases available."
    echo "Try setting EC_VERSION manually: EC_VERSION=v1.0.0 ./install-team.sh"
    exit 1
fi

DOWNLOAD_URL="https://github.com/MereWhiplash/engram-cogitator/releases/download/${VERSION}/ec-shim_${VERSION#v}_${OS}_${ARCH}.tar.gz"

echo "Downloading ec-shim ${VERSION}..."
TEMP_DIR=$(mktemp -d)
if ! curl -sSL "$DOWNLOAD_URL" | tar -xz -C "$TEMP_DIR"; then
    echo -e "${RED}Error: Failed to download ec-shim${NC}"
    rm -rf "$TEMP_DIR"
    exit 1
fi

# Install binary
INSTALL_DIR="${HOME}/.local/bin"
mkdir -p "$INSTALL_DIR"
mv "$TEMP_DIR/ec-shim" "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/ec-shim"
rm -rf "$TEMP_DIR"

SHIM_PATH="$INSTALL_DIR/ec-shim"
echo -e "${GREEN}Installed ec-shim to $SHIM_PATH${NC}"

# Check if in PATH
if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
    echo -e "${YELLOW}Warning: $INSTALL_DIR is not in your PATH${NC}"
    echo "Add this to your shell profile:"
    echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
fi

echo ""

# ============================================================
# MCP Configuration
# ============================================================

echo -e "${CYAN}=== MCP Configuration ===${NC}"
echo ""

# Check for Claude Code CLI
if command -v claude &> /dev/null; then
    echo "Claude Code CLI detected."
    read -p "Configure Claude Code automatically? [Y/n] " -n 1 -r
    echo ""

    if [[ ! $REPLY =~ ^[Nn]$ ]]; then
        echo "Configuring Claude Code..."
        claude mcp remove engram-cogitator 2>/dev/null || true
        claude mcp add engram-cogitator \
            --scope user \
            -- "$SHIM_PATH" --api-url "$EC_API_URL"
        echo -e "${GREEN}Claude Code configured!${NC}"
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
      "command": "$SHIM_PATH",
      "args": ["--api-url", "$EC_API_URL"]
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

# ============================================================
# AI Assistant Instructions
# ============================================================

echo -e "${CYAN}=== AI Assistant Instructions ===${NC}"
echo ""
echo "Add the following to your AI assistant's instruction file:"
echo ""
echo "  Claude Code:    CLAUDE.md"
echo "  Cursor:         .cursor/rules/engram.mdc"
echo "  GitHub Copilot: .github/copilot-instructions.md"
echo "  Gemini:         GEMINI.md"
echo "  Generic:        AGENTS.md"
echo ""
echo "Instructions content:"
echo "  https://raw.githubusercontent.com/MereWhiplash/engram-cogitator/main/INSTRUCTIONS.md"
echo ""

# ============================================================
# Done
# ============================================================

echo -e "${GREEN}Installation complete!${NC}"
echo ""
echo "Restart your AI coding assistant to activate Engram Cogitator."
echo ""
echo "Available MCP tools:"
echo "  - ec_add        : Add memories (decisions, learnings, patterns)"
echo "  - ec_search     : Search team memories semantically"
echo "  - ec_list       : List recent memories"
echo "  - ec_invalidate : Mark memories as outdated"
