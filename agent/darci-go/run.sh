#!/bin/bash
# DarCI Go Agent - Run Script
# Usage: ./run.sh [agent_id]

set -e

# Configuration
AGENT_ID="${1:-darci-go-$(hostname)}"
BRIDGE_URL="${DARCI_BRIDGE_URL:-http://127.0.0.1:8080}"
LISTEN_ADDR="${DARCI_LISTEN_ADDR:-:9090}"

echo "=== DarCI Go Agent ==="
echo ""
echo "Agent ID:    $AGENT_ID"
echo "Bridge URL:  $BRIDGE_URL"
echo "Listen Addr: $LISTEN_ADDR"
echo ""

# Check if secret is set
if [ -z "$DARCI_AGENT_SECRET" ]; then
    echo "⚠️  DARCI_AGENT_SECRET not set!"
    echo ""
    echo "To generate a secret on the bridge:"
    echo "  taila2a secrets generate $AGENT_ID"
    echo ""
    echo "Then set the environment variable:"
    echo "  export DARCI_AGENT_SECRET=<secret>"
    echo ""
    exit 1
fi

# Build if binary doesn't exist
if [ ! -f "./darci" ]; then
    echo "Building darci..."
    go build -o darci ./cmd/darci
fi

echo "Starting DarCI Go agent..."
echo ""

# Run agent
./darci
