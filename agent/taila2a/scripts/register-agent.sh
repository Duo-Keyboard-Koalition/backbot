#!/bin/bash
# AIP Agent Registration Script
# Usage: ./register-agent.sh <agent_id> <agent_type> <endpoint_url>

set -e

# Configuration
BRIDGE_URL="${BRIDGE_URL:-http://127.0.0.1:8080}"
AGENT_ID="${1:-darci-python-$(hostname)}"
AGENT_TYPE="${2:-darci-python}"
AGENT_VERSION="${3:-1.0.0}"
ENDPOINT_URL="${4:-http://127.0.0.1:9090/api}"
HEALTH_URL="${5:-http://127.0.0.1:9090/health}"

# Capabilities (customize based on agent type)
CAPABILITIES='["task-execution", "notebook", "file-ops"]'

# Metadata
HOSTNAME=$(hostname)
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
TAGS='["development"]'

echo "=== AIP Agent Registration ==="
echo ""
echo "Agent ID:    $AGENT_ID"
echo "Type:        $AGENT_TYPE"
echo "Version:     $AGENT_VERSION"
echo "Endpoint:    $ENDPOINT_URL"
echo "Bridge URL:  $BRIDGE_URL"
echo "Hostname:    $HOSTNAME"
echo "OS:          $OS"
echo ""

# Create registration payload
PAYLOAD=$(cat <<EOF
{
  "agent_id": "$AGENT_ID",
  "agent_type": "$AGENT_TYPE",
  "agent_version": "$AGENT_VERSION",
  "capabilities": $CAPABILITIES,
  "endpoints": {
    "primary": "$ENDPOINT_URL",
    "health": "$HEALTH_URL"
  },
  "metadata": {
    "hostname": "$HOSTNAME",
    "os": "$OS",
    "tags": $TAGS
  }
}
EOF
)

echo "Sending registration request..."
echo ""

# Send registration request
RESPONSE=$(curl -s -X POST "$BRIDGE_URL/aip/register" \
  -H "Content-Type: application/json" \
  -d "$PAYLOAD")

echo "Response:"
echo "$RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$RESPONSE"

# Parse status
STATUS=$(echo "$RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin).get('status', 'unknown'))" 2>/dev/null || echo "unknown")

echo ""
echo "=== Next Steps ==="
echo ""

if [ "$STATUS" = "pending" ]; then
    echo "✓ Registration submitted successfully"
    echo ""
    echo "Your registration is pending admin approval."
    echo ""
    echo "On the bridge machine, run:"
    echo "  taila2a aip pending"
    echo "  taila2a aip approve $AGENT_ID"
    echo ""
    echo "After approval, start sending heartbeats:"
    echo "  curl -X POST $BRIDGE_URL/aip/heartbeat \\"
    echo "    -H 'Content-Type: application/json' \\"
    echo "    -d '{\"agent_id\": \"$AGENT_ID\", \"timestamp\": \"'$(date -Iseconds)'\", \"status\": \"healthy\"}'"
    exit 0
elif [ "$STATUS" = "approved" ]; then
    echo "✓ Agent already approved!"
    echo ""
    echo "Start sending heartbeats to maintain active status."
    exit 0
else
    echo "⚠ Registration status: $STATUS"
    echo ""
    echo "Check with the bridge administrator if this status persists."
    exit 1
fi
