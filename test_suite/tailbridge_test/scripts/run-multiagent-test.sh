#!/bin/bash
# Multi-Agent Tailscale Test Runner
# This script runs the multi-agent integration test and displays IP addresses and chat logs

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
export TS_AUTH_KEY="${TS_AUTH_KEY:-tskey-auth-k7Q1t39ZWj11CNTRL-FbRR2tKLRcPn5L246vsAcP7LP2YCUxWD}"
export GEMINI_API_KEY="${GEMINI_API_KEY:-AIzaSyC_9M8im8z5F0ING2W3Hu2aQiunJhhWUXI}"
export AGENT1_URL="${AGENT1_URL:-http://localhost:8081}"
export AGENT2_URL="${AGENT2_URL:-http://localhost:8082}"
export AGENT3_URL="${AGENT3_URL:-http://localhost:8083}"

echo -e "${CYAN}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║     MULTI-AGENT TAILSCALE COMMUNICATION TEST RUNNER      ║${NC}"
echo -e "${CYAN}╚═══════════════════════════════════════════════════════════╝${NC}"
echo ""

# Check if Docker is running
if ! docker ps >/dev/null 2>&1; then
    echo -e "${RED}✗ Docker is not running. Please start Docker Desktop.${NC}"
    exit 1
fi
echo -e "${GREEN}✓${NC} Docker is running"

# Check if .env file exists
if [ ! -f "docker/.env" ]; then
    echo -e "${YELLOW}!${NC} Creating docker/.env file with auth keys..."
    cat > docker/.env << EOF
TS_AUTH_KEY_1=${TS_AUTH_KEY}
TS_AUTH_KEY_2=${TS_AUTH_KEY}
TS_AUTH_KEY_3=${TS_AUTH_KEY}
TAILNET_NAME=test.ts.net
EOF
    echo -e "${GREEN}✓${NC} Created docker/.env"
fi

echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Configuration:${NC}"
echo -e "  Tailscale Auth Key: ${TS_AUTH_KEY:0:30}..."
echo -e "  Agent 1 URL: ${AGENT1_URL}"
echo -e "  Agent 2 URL: ${AGENT2_URL}"
echo -e "  Agent 3 URL: ${AGENT3_URL}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo ""

# Start the test environment
echo -e "${YELLOW}📦${NC} Starting Docker test environment..."
docker-compose -f docker/docker-compose.test.yml up -d

echo ""
echo -e "${YELLOW}⏳${NC} Waiting for agents to connect to Tailscale (this may take 2-3 minutes)..."
echo ""

# Wait for agents
MAX_ATTEMPTS=30
ATTEMPT=0

while [ $ATTEMPT -lt $MAX_ATTEMPTS ]; do
    ATTEMPT=$((ATTEMPT + 1))
    echo -ne "${CYAN}  Attempt $ATTEMPT/$MAX_ATTEMPTS...${NC}\\r"
    
    ALL_HEALTHY=true
    
    for i in 1 2 3; do
        PORT=$((8080 + i))
        if ! curl -s --max-time 2 "http://localhost:$PORT/health" >/dev/null 2>&1; then
            ALL_HEALTHY=false
            break
        fi
    done
    
    if [ "$ALL_HEALTHY" = true ]; then
        echo -e "${GREEN}  ✓ All agents are healthy!${NC}          "
        break
    fi
    
    sleep 10
done

if [ "$ALL_HEALTHY" = false ]; then
    echo -e "${RED}✗ Timeout waiting for agents to become healthy${NC}"
    echo -e "${YELLOW}!${NC} Showing agent logs:"
    docker-compose -f docker/docker-compose.test.yml logs --tail=20
    exit 1
fi

echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}AGENT IP ADDRESSES:${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"

# Get agent information
for i in 1 2 3; do
    PORT=$((8080 + i))
    echo ""
    echo -e "${CYAN}Agent $i:${NC}"
    
    # Get status from agent
    STATUS=$(curl -s --max-time 5 "http://localhost:$PORT/status" 2>/dev/null || echo "{}")
    
    NAME=$(echo $STATUS | jq -r '.name // "agent'$i'"' 2>/dev/null || echo "agent$i")
    IP=$(echo $STATUS | jq -r '.ip // "unknown"' 2>/dev/null || echo "unknown")
    TS_IP=$(echo $STATUS | jq -r '.tailscale_ip // "connecting..."' 2>/dev/null || echo "connecting...")
    ONLINE=$(echo $STATUS | jq -r '.online // false' 2>/dev/null || echo "false")
    
    echo -e "  Hostname:      ${NAME}"
    echo -e "  Local IP:      ${IP}"
    echo -e "  Tailscale IP:  ${TS_IP}"
    echo -e "  Status:        ${ONLINE}"
done

echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}RUNNING MULTI-AGENT COMMUNICATION TEST:${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo ""

# Run the integration test
cd "$(dirname "$0")"
go test ./integration/multiagent/... -v -tags=integration -timeout=10m 2>&1 | tee test_logs/multiagent-test-$(date +%Y%m%d-%H%M%S).log

TEST_EXIT_CODE=${PIPESTATUS[0]}

echo ""
if [ $TEST_EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}╔═══════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║              TEST PASSED SUCCESSFULLY                     ║${NC}"
    echo -e "${GREEN}╚═══════════════════════════════════════════════════════════╝${NC}"
else
    echo -e "${RED}╔═══════════════════════════════════════════════════════════╗${NC}"
    echo -e "${RED}║                    TEST FAILED                            ║${NC}"
    echo -e "${RED}╚═══════════════════════════════════════════════════════════╝${NC}"
fi

echo ""
echo -e "${YELLOW}!${NC} Test logs saved to: test_logs/multiagent-test-*.log"
echo ""

# Optionally cleanup
read -p "Stop Docker containers? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}🧹${NC} Stopping Docker containers..."
    docker-compose -f docker/docker-compose.test.yml down
    echo -e "${GREEN}✓${NC} Containers stopped"
fi

exit $TEST_EXIT_CODE
