#!/bin/bash
set -e

# Walkie-Talkie Bridge Uninstall Script
# Usage: curl -fsSL https://raw.githubusercontent.com/codejedi-ai/tailtailkie/main/tailscale-app/scripts/uninstall.sh | sudo bash

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

INSTALL_DIR="/opt/walkie-talkie"
SYSTEMD_SERVICE="walkie-talkie-bridge.service"
CONFIG_DIR="$HOME/.tailtalkie"
BIN_DIR="/usr/local/bin"

echo -e "${RED}=== Walkie-Talkie Bridge Uninstaller ===${NC}"
echo

if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}Error: This script must be run as root (use sudo)${NC}"
    exit 1
fi

echo -e "${YELLOW}Stopping service...${NC}"
systemctl stop $SYSTEMD_SERVICE 2>/dev/null || true
systemctl disable $SYSTEMD_SERVICE 2>/dev/null || true

echo -e "${YELLOW}Removing systemd service...${NC}"
rm -f /etc/systemd/system/$SYSTEMD_SERVICE
systemctl daemon-reload

echo -e "${YELLOW}Removing binary...${NC}"
rm -f "$BIN_DIR/walkie-talkie-bridge"

echo -e "${YELLOW}Removing installation directory...${NC}"
rm -rf "$INSTALL_DIR"

echo
echo -e "${YELLOW}Configuration preserved at: $CONFIG_DIR${NC}"
echo "To remove configuration as well, run:"
echo -e "  ${RED}rm -rf $CONFIG_DIR${NC}"
echo

echo -e "${GREEN}✓ Uninstall complete${NC}"
