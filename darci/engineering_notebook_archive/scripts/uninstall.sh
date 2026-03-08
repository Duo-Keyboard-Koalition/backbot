#!/bin/bash
# darci Uninstall Script for Linux/macOS
# Usage: ./uninstall.sh [-y]
#   -y  Skip confirmation prompt

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DARCI_DIR="$(dirname "$SCRIPT_DIR")"

echo "🐈 darci - Uninstallation (Linux/macOS)"
echo "==========================================="

# Check for -y flag
CONFIRM="n"
if [ "$1" = "-y" ] || [ "$1" = "-Y" ]; then
    CONFIRM="y"
fi

if [ "$CONFIRM" != "y" ]; then
    echo ""
    echo "This will remove:"
    echo "  - Python dependencies (virtual environment)"
    echo "  - Go binary (darci-go)"
    echo "  - Configuration files"
    echo "  - Workspace files"
    echo "  - Desktop shortcuts"
    echo ""
    read -p "Are you sure you want to continue? (y/N): " CONFIRM
    if [[ "$CONFIRM" != "y" && "$CONFIRM" != "Y" ]]; then
        echo "Uninstallation cancelled."
        exit 0
    fi
fi

# ========== Remove Python Bot ==========
echo ""
echo "[1/5] Removing Python Bot..."
cd "$DARCI_DIR/darci-python"
rm -rf .venv darci_env

# ========== Remove Go Bot ==========
echo ""
echo "[2/5] Removing Go Bot..."
cd "$DARCI_DIR/darci-go"
rm -f darci-go darci-go.exe

# ========== Remove Configuration ==========
echo ""
echo "[3/5] Removing configuration..."
rm -rf "$HOME/.config/darci"

# ========== Remove Workspace ==========
echo ""
echo "[4/5] Removing workspace..."
rm -rf "$HOME/darci-workspace"

# ========== Remove Shortcuts ==========
echo ""
echo "[5/5] Removing shortcuts..."
rm -f "$HOME/.local/bin/darci-python"
rm -f "$HOME/.local/bin/darci-go"
rm -f "$HOME/.local/bin/uninstall-darci"

echo ""
echo "==========================================="
echo "✓ Uninstallation complete!"
echo "==========================================="
echo ""
echo "To reinstall, run: ./install.sh [API_KEY]"
echo "Example: ./install.sh AIzaSyCSvcSZsC8Bg1k343y9l3as3vlOrhsXRSw"
echo ""
