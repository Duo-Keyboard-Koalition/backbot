#!/bin/bash
set -e

# Walkie-Talkie Bridge Installation Script
# Installs the bridge service on Linux systems
# Usage: curl -fsSL https://raw.githubusercontent.com/codejedi-ai/tailtailkie/main/tailscale-app/scripts/install.sh | sudo bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
REPO_URL="https://github.com/codejedi-ai/tailtailkie.git"
INSTALL_DIR="/opt/walkie-talkie"
LEGACY_REPO_DIR="$INSTALL_DIR/tailtailkie"
SYSTEMD_SERVICE="walkie-talkie-bridge.service"
CONFIG_DIR="$HOME/.tailtalkie"
BIN_DIR="/usr/local/bin"

echo -e "${GREEN}=== Walkie-Talkie Bridge Installer ===${NC}"
echo

# Ensure common binary directories are in PATH for non-interactive sudo shells.
export PATH="$PATH:/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

# Returns 0 if a command exists in PATH or common absolute locations.
has_cmd() {
    local cmd="$1"
    command -v "$cmd" >/dev/null 2>&1 \
    || [ -x "/usr/local/go/bin/$cmd" ] \
        || [ -x "/usr/local/bin/$cmd" ] \
        || [ -x "/usr/bin/$cmd" ] \
        || [ -x "/bin/$cmd" ]
}

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}Error: This script must be run as root (use sudo)${NC}"
    exit 1
fi

# Detect system architecture
ARCH=$(uname -m)
case $ARCH in
    x86_64)
        GOARCH="amd64"
        ;;
    aarch64)
        GOARCH="arm64"
        ;;
    armv7l)
        GOARCH="arm"
        ;;
    *)
        echo -e "${RED}Error: Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

echo -e "${YELLOW}Detected architecture: $ARCH ($GOARCH)${NC}"

# Check if git is installed
if ! has_cmd git; then
    echo -e "${YELLOW}Git not found. Installing git...${NC}"
    
    # Detect package manager
    if has_cmd apt-get; then
        apt-get update -qq
        apt-get install -y -qq git
    elif has_cmd yum; then
        yum install -y -q git
    elif has_cmd dnf; then
        dnf install -y -q git
    elif has_cmd apk; then
        apk add --no-cache git
    elif has_cmd zypper; then
        zypper install -y git
    else
        echo -e "${RED}Error: Could not detect package manager. Please install git manually.${NC}"
        exit 1
    fi

    if ! has_cmd git; then
        echo -e "${RED}Error: Git installation did not make 'git' available on PATH.${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✓ Git installed successfully${NC}"
fi

# Check if Go is installed
if ! has_cmd go; then
    echo -e "${YELLOW}Go not found. Installing Go...${NC}"

    # Download and install Go
    GO_VERSION="1.25.0"
    GO_TARBALL="go${GO_VERSION}.linux-${GOARCH}.tar.gz"
    GO_URL="https://go.dev/dl/${GO_TARBALL}"

    cd /tmp
    
    # Try to download with timeout and progress (try wget first, then curl)
    echo "Downloading Go $GO_VERSION..."
    DOWNLOAD_OK=false
    
    if has_cmd wget; then
        if wget --timeout=60 --tries=2 -q --show-progress "$GO_URL" -O "$GO_TARBALL"; then
            DOWNLOAD_OK=true
        fi
    elif has_cmd curl; then
        if curl --connect-timeout 60 --retry 2 -sS -L "$GO_URL" -o "$GO_TARBALL"; then
            DOWNLOAD_OK=true
        fi
    fi
    
    if [ "$DOWNLOAD_OK" = true ] && [ -f "$GO_TARBALL" ]; then
        echo "Extracting Go..."
        tar -C /usr/local -xzf "$GO_TARBALL"
        rm -f "$GO_TARBALL"

        # Add Go to PATH
        if ! grep -q "export PATH=\$PATH:/usr/local/go/bin" /etc/profile; then
            echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
        fi

        export PATH=$PATH:/usr/local/go/bin
        echo -e "${GREEN}✓ Go installed successfully${NC}"
    else
        echo -e "${RED}Error: Failed to download Go.${NC}"
        echo "Please install Go manually:"
        echo "  1. Download: $GO_URL"
        echo "  2. Extract: sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-${GOARCH}.tar.gz"
        echo "  3. Add to PATH: export PATH=\$PATH:/usr/local/go/bin"
        exit 1
    fi
fi

# Verify Go installation
export PATH=$PATH:/usr/local/go/bin
if ! has_cmd go; then
    echo -e "${RED}Error: Go is still not available after installation.${NC}"
    exit 1
fi
GO_VERSION=$(go version | awk '{print $3}')
echo -e "${GREEN}✓ Found $GO_VERSION${NC}"

# Create installation directory
echo -e "${YELLOW}Creating installation directory...${NC}"
mkdir -p "$INSTALL_DIR"

# Check if this is an update or fresh install
IS_UPDATE=false
if [ -f "$BIN_DIR/walkie-talkie-bridge" ]; then
    IS_UPDATE=true
    CURRENT_VERSION=$("$BIN_DIR/walkie-talkie-bridge" version 2>/dev/null || echo "unknown")
    echo -e "${YELLOW}Existing installation detected (version: $CURRENT_VERSION)${NC}"
fi

# Clone or update repository
# Support both legacy layout (/opt/walkie-talkie/tailtailkie)
# and normalized layout (/opt/walkie-talkie as git root).
REPO_DIR=""
VALID_INSTALL=false

# Check for valid installation (must have .git AND tailscale-app directory)
if [ -d "$INSTALL_DIR/.git" ] && [ -d "$INSTALL_DIR/tailscale-app" ]; then
    REPO_DIR="$INSTALL_DIR"
    VALID_INSTALL=true
elif [ -d "$LEGACY_REPO_DIR/.git" ] && [ -d "$LEGACY_REPO_DIR/tailscale-app" ]; then
    REPO_DIR="$LEGACY_REPO_DIR"
    VALID_INSTALL=true
fi

if [ "$VALID_INSTALL" = true ]; then
    echo -e "${YELLOW}Updating repository...${NC}"
    cd "$REPO_DIR"
    git fetch --quiet
    git reset --hard origin/main --quiet
else
    echo -e "${YELLOW}Cloning repository...${NC}"
    # Remove old/broken installation if it exists and clone fresh.
    rm -rf "$INSTALL_DIR"
    git clone "$REPO_URL" "$INSTALL_DIR" --quiet
    REPO_DIR="$INSTALL_DIR"
fi

# If we updated from legacy layout, keep building from that location.
if [ -z "$REPO_DIR" ]; then
    REPO_DIR="$INSTALL_DIR"
fi

# Build the bridge
echo -e "${YELLOW}Building bridge...${NC}"
cd "$REPO_DIR/tailscale-app"
go build -o "$BIN_DIR/walkie-talkie-bridge" ./bridge

if [ ! -f "$BIN_DIR/walkie-talkie-bridge" ]; then
    echo -e "${RED}Error: Failed to build bridge${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Bridge built successfully${NC}"

# Restart service if it was running (update scenario)
if [ "$IS_UPDATE" = true ]; then
    echo -e "${YELLOW}Restarting service with new version...${NC}"
    systemctl restart $SYSTEMD_SERVICE 2>/dev/null || true
    echo -e "${GREEN}✓ Service updated and restarted${NC}"
fi

# Create config directory
echo -e "${YELLOW}Setting up configuration...${NC}"
mkdir -p "$CONFIG_DIR"
chown -R $(whoami):$(whoami) "$CONFIG_DIR"
chmod 700 "$CONFIG_DIR"

# Create systemd service
echo -e "${YELLOW}Installing systemd service...${NC}"
cat > /etc/systemd/system/$SYSTEMD_SERVICE <<EOF
[Unit]
Description=Walkie-Talkie Bridge Service
Documentation=https://github.com/codejedi-ai/Kaggle-For-Tensors
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=$(whoami)
Group=$(whoami)
ExecStart=$BIN_DIR/walkie-talkie-bridge run
Restart=on-failure
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=walkie-talkie-bridge

# Security hardening
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=read-only
PrivateTmp=true

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd
systemctl daemon-reload

# Enable service (but don't start yet)
systemctl enable $SYSTEMD_SERVICE --quiet

echo -e "${GREEN}✓ Systemd service installed${NC}"

# Check if config exists
if [ ! -f "$CONFIG_DIR/config.json" ]; then
    echo
    echo -e "${YELLOW}=== Initial Configuration Required ===${NC}"
    echo
    echo "1. Configure your bridge:"
    echo -e "${GREEN}  sudo walkie-talkie-bridge init${NC}"
    echo
    echo "   You'll be prompted for:"
    echo "   - Tailscale Auth Key (from https://login.tailscale.com/admin/settings/keys)"
    echo "   - Bridge Name (unique identifier, e.g., bridge-alpha)"
    echo "   - Local Agent URL (your agent's API endpoint)"
    echo
    echo "2. Start the service:"
    echo -e "${GREEN}  sudo systemctl start $SYSTEMD_SERVICE${NC}"
else
    echo -e "${GREEN}✓ Configuration found${NC}"
    echo
    echo "Starting service..."
    systemctl start $SYSTEMD_SERVICE
    echo -e "${GREEN}✓ Service started${NC}"
fi

echo
echo -e "${GREEN}=== Installation Complete ===${NC}"
echo
echo -e "${YELLOW}Quick Start Guide:${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo
echo "1. Initialize (first time only):"
echo -e "   ${GREEN}sudo walkie-talkie-bridge init${NC}"
echo
echo "2. Start the service:"
echo -e "   ${GREEN}sudo systemctl start $SYSTEMD_SERVICE${NC}"
echo
echo "3. Check status:"
echo -e "   ${GREEN}sudo systemctl status $SYSTEMD_SERVICE${NC}"
echo
echo -e "${YELLOW}CLI Commands:${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  walkie-talkie-bridge init    # Interactive configuration"
echo "  walkie-talkie-bridge run     # Start bridge (foreground)"
echo "  walkie-talkie-bridge help    # Show help message"
echo
echo -e "${YELLOW}Service Management:${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  sudo systemctl start $SYSTEMD_SERVICE    # Start service"
echo "  sudo systemctl stop $SYSTEMD_SERVICE     # Stop service"
echo "  sudo systemctl restart $SYSTEMD_SERVICE  # Restart service"
echo "  sudo systemctl status $SYSTEMD_SERVICE   # Check status"
echo "  journalctl -u $SYSTEMD_SERVICE -f        # View logs (follow)"
echo "  journalctl -u $SYSTEMD_SERVICE --since today  # Today's logs"
echo
echo -e "${YELLOW}Configuration:${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Config file: $CONFIG_DIR/config.json"
echo "  State directory: ~/.tailtalkie/state/"
echo
echo -e "${YELLOW}Discover Agents:${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  curl http://localhost:8001/agents  # List all agents on tailnet"
echo
echo -e "${YELLOW}Uninstall:${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  curl -fsSL https://raw.githubusercontent.com/codejedi-ai/tailtailkie/main/tailscale-app/scripts/uninstall.sh | sudo bash"
echo
