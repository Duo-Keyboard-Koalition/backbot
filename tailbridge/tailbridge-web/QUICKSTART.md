# Quick Start Guide

Get Walkie-Talkie for Bots running in minutes.

## Prerequisites

- Go 1.25+
- Tailscale account with auth keys
- Ports 8001, 8080 available

## Step 1: Clone and Setup

```bash
# Navigate to project directory
cd Kaggle-For-Tensors/tailscale-app
```

## Step 2: Start Bridge on Host A

```bash
TS_AUTHKEY=tskey-auth-bridge-a \
BRIDGE_NAME=bridge-alpha \
TSNET_STATE_DIR=./state/bridge-alpha \
PEER_BRIDGE_INBOUND_PORT=8001 \
LOCAL_AGENT_URL=http://127.0.0.1:9090/api \
go run ./bridge
```

## Step 3: Start Bridge on Host B

```bash
TS_AUTHKEY=tskey-auth-bridge-b \
BRIDGE_NAME=bridge-beta \
TSNET_STATE_DIR=./state/bridge-beta \
PEER_BRIDGE_INBOUND_PORT=8001 \
LOCAL_AGENT_URL=http://127.0.0.1:9090/api \
go run ./bridge
```

## Step 4: Send Test Message

```bash
curl -sS http://127.0.0.1:8080/send \
  -H 'content-type: application/json' \
  -d '{
    "source_node": "bridge-alpha",
    "dest_node": "bridge-beta",
    "payload": {"message": "hello from alpha"}
  }'
```

## Common Commands

```bash
# Start bridge
go run ./bridge

# Check bridge status
curl http://127.0.0.1:8080/health
```

## Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `TS_AUTHKEY` | Tailscale auth key | `tskey-auth-xxx` |
| `BRIDGE_NAME` | Unique bridge ID | `bridge-alpha` |
| `TSNET_STATE_DIR` | Node identity storage | `./state/bridge-alpha` |
| `BRIDGE_INBOUND_PORT` | Inbound peer port | `8001` |
| `PEER_BRIDGE_INBOUND_PORT` | Outbound peer port | `8001` |
| `BRIDGE_LOCAL_LISTEN` | Local agent endpoint | `127.0.0.1:8080` |
| `LOCAL_AGENT_URL` | Agent URL | `http://127.0.0.1:9090/api` |

## Troubleshooting

### Bridge won't start
1. Check Tailscale auth key is valid
2. Check ports are available: `lsof -i :8001 -i :8080`
3. Check Go version: `go version` (need 1.25+)

### Can't reach peer bridge
1. Verify both bridges are online on Tailnet
2. Check `BRIDGE_NAME` is unique per host
3. Verify Tailscale ACLs allow peer communication

### Node identity lost
1. Ensure `TSNET_STATE_DIR` persists between restarts
2. Don't share state directories between bridges

## Next Steps

- Read `tailscale-app/docs/agent-communication.md` for message flow details
- Configure Tailscale ACLs: `docs/tailscale-acl.example.json`
- Review engineering notebook: `engineering-notebook/README.md`
