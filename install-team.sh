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
