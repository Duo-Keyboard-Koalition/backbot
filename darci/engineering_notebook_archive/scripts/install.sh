#!/bin/bash
# darci Install Script for Linux/macOS
# Usage: ./install.sh [GEMINI_API_KEY]
# Example: ./install.sh AIzaSyCSvcSZsC8Bg1k343y9l3as3vlOrhsXRSw

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DARCI_DIR="$(dirname "$SCRIPT_DIR")"

echo "🐈 darci - Installation (Linux/macOS)"
echo "========================================="

# Check for API key argument
API_KEY="${1:-}"

# ========== Prerequisites Check ==========
echo ""
echo "[0/5] Checking prerequisites..."
echo ""

# Helper function to check command
check_cmd() {
    command -v "$1" &> /dev/null
}

# Check Python (3.11+)
echo "Checking Python..."
if ! check_cmd python3; then
    echo "Python3 is not installed."
    if check_cmd apt-get; then
        echo "Installing Python3 via apt..."
        sudo apt-get update
        sudo apt-get install -y python3 python3-pip python3-venv
    elif check_cmd brew; then
        echo "Installing Python3 via Homebrew..."
        brew install python@3.11
    elif check_cmd dnf; then
        echo "Installing Python3 via dnf..."
        sudo dnf install -y python3 python3-pip
    elif check_cmd pacman; then
        echo "Installing Python3 via pacman..."
        sudo pacman -S --noconfirm python python-pip
    else
        echo "ERROR: Could not install Python automatically."
        echo "Please install Python 3.11+ from https://www.python.org/downloads/"
        exit 1
    fi
fi

PYTHON_VERSION=$(python3 --version 2>&1 | grep -oP '\d+\.\d+' | head -1)
PYTHON_MAJOR=$(echo "$PYTHON_VERSION" | cut -d. -f1)
PYTHON_MINOR=$(echo "$PYTHON_VERSION" | cut -d. -f2)

if [ "$PYTHON_MAJOR" -lt 3 ] || ([ "$PYTHON_MAJOR" -eq 3 ] && [ "$PYTHON_MINOR" -lt 11 ]); then
    echo "WARNING: Python $PYTHON_VERSION found, but 3.11+ is recommended."
    echo "Continuing with installation..."
fi

# Check uv
echo "Checking uv..."
if ! check_cmd uv; then
    echo "uv is not installed. Installing..."
    curl -LsSf https://astral.sh/uv/install.sh | sh
    if [ $? -ne 0 ]; then
        echo "Failed to install uv."
        echo "Please install uv from https://docs.astral.sh/uv/getting-started/installation/"
        exit 1
    fi
    export PATH="$HOME/.local/bin:$PATH"
fi

# Check Go
echo "Checking Go..."
if ! check_cmd go; then
    echo "Go is not installed."
    if check_cmd apt-get; then
        echo "Installing Go via apt..."
        sudo apt-get update
        sudo apt-get install -y golang-go
    elif check_cmd brew; then
        echo "Installing Go via Homebrew..."
        brew install go
    elif check_cmd dnf; then
        echo "Installing Go via dnf..."
        sudo dnf install -y golang
    elif check_cmd pacman; then
        echo "Installing Go via pacman..."
        sudo pacman -S --noconfirm go
    else
        echo "ERROR: Could not install Go automatically."
        echo "Please install Go from https://go.dev/dl/"
        exit 1
    fi
fi

GO_VERSION=$(go version 2>&1 | grep -oP '\d+\.\d+' | head -1)
GO_MAJOR=$(echo "$GO_VERSION" | cut -d. -f1)

if [ "$GO_MAJOR" -lt 1 ]; then
    echo "ERROR: Go 1.22+ is required, found $GO_VERSION"
    echo "Please install Go from https://go.dev/dl/"
    exit 1
fi

echo "All prerequisites are installed!"

# ========== Python Bot Installation ==========
echo ""
echo "[1/5] Installing Python Bot..."
echo ""

cd "$DARCI_DIR/darci-python"

# Sync dependencies
echo "Installing Python dependencies..."
uv sync
if [ $? -ne 0 ]; then
    echo "Failed to install Python dependencies."
    exit 1
fi

# ========== Go Bot Installation ==========
echo ""
echo "[2/5] Building Go Bot..."
echo ""

cd "$DARCI_DIR/darci-go"

# Build Go bot
echo "Building darci-go..."
go build -o darci-go ./cmd/darci-go
if [ $? -ne 0 ]; then
    echo "Failed to build Go bot."
    exit 1
fi

echo "Go bot built successfully!"

# ========== Configuration ==========
echo ""
echo "[3/5] Configuring..."
echo ""

# Create config directory
CONFIG_DIR="$HOME/.config/darci"
mkdir -p "$CONFIG_DIR"

CONFIG_FILE="$CONFIG_DIR/config.json"

# Create or update config with API key
if [ -z "$API_KEY" ]; then
    echo "No API key provided. You can add it later to $CONFIG_FILE"
else
    echo "Setting API key..."
    if [ -f "$CONFIG_FILE" ]; then
        sed -i.bak "s/\"apiKey\": \"\"/\"apiKey\": \"$API_KEY\"/g" "$CONFIG_FILE"
        rm -f "$CONFIG_FILE.bak"
    else
        cat > "$CONFIG_FILE" << EOF
{
  "agents": {
    "defaults": {
      "model": "gemini-2.5-flash",
      "provider": "gemini"
    }
  },
  "providers": {
    "gemini": {
      "apiKey": "$API_KEY"
    }
  },
  "gateway": {
    "port": 18790,
    "heartbeat": {
      "enabled": true,
      "intervalS": 1800
    }
  },
  "tools": {
    "web": {
      "search": {
        "apiKey": ""
      }
    },
    "restrictToWorkspace": false
  }
}
EOF
    fi
fi

# Create workspace
echo "Creating workspace..."
WORKSPACE="$HOME/darci-workspace"
mkdir -p "$WORKSPACE"

# Copy bootstrap files
cp -r "$DARCI_DIR/darci-python/docs/"*.md "$WORKSPACE/" 2>/dev/null || true

# ========== Create Launcher Scripts ==========
echo ""
echo "[4/5] Creating launcher scripts..."
echo ""

BIN_DIR="$HOME/.local/bin"
mkdir -p "$BIN_DIR"

# Get absolute paths
DARCI_PYTHON_DIR="$(cd "$DARCI_DIR/darci-python" && pwd)"
DARCI_GO_DIR="$(cd "$DARCI_DIR/darci-go" && pwd)"
SCRIPTS_DIR="$(cd "$SCRIPT_DIR" && pwd)"

# Create Python launcher that activates venv automatically
cat > "$BIN_DIR/darci-python" << LAUNCHER
#!/bin/bash
# darci-python launcher - Linux/macOS
cd "$DARCI_PYTHON_DIR"

# Activate virtual environment
if [ -f ".venv/bin/activate" ]; then
    source .venv/bin/activate
fi

# Run darci agent
python -m darci agent "\$@"
LAUNCHER
chmod +x "$BIN_DIR/darci-python"

# Create Go launcher
cat > "$BIN_DIR/darci-go" << LAUNCHER
#!/bin/bash
# darci-go launcher - Linux/macOS
cd "$DARCI_GO_DIR"
./darci-go "\$@"
LAUNCHER
chmod +x "$BIN_DIR/darci-go"

# Create uninstall launcher
cat > "$BIN_DIR/uninstall-darci" << LAUNCHER
#!/bin/bash
# darci uninstall launcher
cd "$SCRIPTS_DIR"
./uninstall.sh "\$@"
LAUNCHER
chmod +x "$BIN_DIR/uninstall-darci"

# ========== PATH Configuration ==========
echo ""
echo "[5/5] Configuring PATH..."
echo ""

# Check if ~/.local/bin is in PATH
if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
    echo "Adding ~/.local/bin to PATH..."
    
    # Detect shell and update appropriate config
    if [ -n "$ZSH_VERSION" ] || [ -f "$HOME/.zshrc" ]; then
        SHELL_RC="$HOME/.zshrc"
        SHELL_NAME="zsh"
    elif [ -n "$BASH_VERSION" ] || [ -f "$HOME/.bashrc" ]; then
        SHELL_RC="$HOME/.bashrc"
        SHELL_NAME="bash"
    else
        SHELL_RC="$HOME/.profile"
        SHELL_NAME="profile"
    fi
    
    if ! grep -q ".local/bin" "$SHELL_RC" 2>/dev/null; then
        echo "" >> "$SHELL_RC"
        echo '# Added by darci installer' >> "$SHELL_RC"
        echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$SHELL_RC"
        echo "Added PATH configuration to $SHELL_RC for $SHELL_NAME"
    fi
fi

echo ""
echo "========================================="
echo "✓ Installation complete!"
echo "========================================="
echo ""
echo "To run the bots:"
echo "  From anywhere: darci-python or darci-go"
echo "  Manual: cd darci-python && uv run darci agent -m \"Hello!\""
echo "          cd darci-go && ./darci-go"
echo ""
if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
    echo "IMPORTANT: Run the following to update your PATH:"
    echo "  source $SHELL_RC"
    echo "  Or restart your terminal"
    echo ""
fi
echo "To uninstall:"
echo "  Run: uninstall-darci"
echo ""
if [ -z "$API_KEY" ]; then
    echo "Remember to add your Gemini API key to:"
    echo "  $CONFIG_FILE"
    echo "  Get one at: https://aistudio.google.com/apikey"
fi
echo ""
