# Walkie-Talkie for Bots

A peer-to-peer communication system for AI agents using Tailscale tsnet bridges.

## Architecture

This system enables direct agent-to-agent communication through bridge nodes:

- **Bridges** (`tailscale-app/bridge/`) - Go tsnet nodes that handle peer routing
- **Frontend** (`web/`) - Dashboard for monitoring bridge connections and messages

### Bridge Network

Each bridge is a peer on the Tailnet:
- No central gateway - direct peer-to-peer routing
- Each binary is its own Tailnet node
- Uses tsnet for embedded Tailscale functionality

**Important**: Bridges handle all networking. Agents communicate via local HTTP to their bridge.

### Message Flow

```
Agent A -> Bridge A (/send)
Bridge A -> Bridge B (/inbound) [Tailnet]
Bridge B -> Agent B (LOCAL_AGENT_URL)
Agent B -> Bridge B -> Bridge A -> Agent A
```

## Project Structure

```
Kaggle-For-Tensors/
├── tailscale-app/           # Bridge application
│   ├── bridge/              # tsnet bridge node
│   ├── protocol/            # Message envelope schema
│   ├── docs/                # Architecture docs
│   └── state/               # Persistent node identity
├── web/                     # Dashboard frontend
│   ├── app/
│   └── components/
└── engineering-notebook/    # Change logs
```

## Quick Start

### Bridge Setup

1. Navigate to bridge:
   ```bash
   cd tailscale-app
   ```

2. Start bridge on Host A:
   ```bash
   TS_AUTHKEY=tskey-auth-bridge-a \
   BRIDGE_NAME=bridge-alpha \
   TSNET_STATE_DIR=./state/bridge-alpha \
   PEER_BRIDGE_INBOUND_PORT=8001 \
   LOCAL_AGENT_URL=http://127.0.0.1:9090/api \
   go run ./bridge
   ```

3. Start bridge on Host B:
   ```bash
   TS_AUTHKEY=tskey-auth-bridge-b \
   BRIDGE_NAME=bridge-beta \
   TSNET_STATE_DIR=./state/bridge-beta \
   PEER_BRIDGE_INBOUND_PORT=8001 \
   LOCAL_AGENT_URL=http://127.0.0.1:9090/api \
   go run ./bridge
   ```

### Send Messages

From local agent to peer bridge:

```bash
curl -sS http://127.0.0.1:8080/send \
  -H 'content-type: application/json' \
  -d '{
    "source_node": "bridge-alpha",
    "dest_node": "bridge-beta",
    "payload": {"message": "hello from alpha"}
  }'
```

## Envelope Format

All messages use JSON envelope schema:

```json
{
  "source_node": "bridge-alpha",
  "dest_node": "bridge-beta",
  "payload": {"message": "hello"}
}
```

## Endpoints

### Bridge Local Side

- `POST /send` on `BRIDGE_LOCAL_LISTEN` (default `127.0.0.1:8080`)
- Called by local agent
- Body is an Envelope JSON

### Bridge Peer Inbound Side

- `POST /inbound` on `<bridge-name>:BRIDGE_INBOUND_PORT` (default `8001`)
- Called by another bridge over Tailnet
- Extracts payload and forwards to `LOCAL_AGENT_URL`

## Environment Variables

### Bridge

- `TS_AUTHKEY` - Tailscale authentication key
- `BRIDGE_NAME` - Unique bridge identifier (e.g., `bridge-alpha`)
- `TSNET_STATE_DIR` - Persistent directory for node identity
- `BRIDGE_INBOUND_PORT` - Port for inbound peer connections (default: `8001`)
- `PEER_BRIDGE_INBOUND_PORT` - Port for outbound peer connections (default: `8001`)
- `BRIDGE_LOCAL_LISTEN` - Local agent endpoint (default: `127.0.0.1:8080`)
- `LOCAL_AGENT_URL` - URL of local agent (e.g., `http://127.0.0.1:9090/api`)

## Development

### Run Bridge

```bash
cd tailscale-app
go run ./bridge
```

### Configure ACLs

Use Tailscale ACLs to restrict which peer bridges can reach `/inbound`:

```json
{
  "hosts": {
    "bridge-alpha": "user:alice",
    "bridge-beta": "user:bob"
  },
  "acls": [
    {"action": "accept", "src": ["bridge-alpha"], "dst": ["bridge-beta:8001"]}
  ]
}
```

See `docs/tailscale-acl.example.json` for full example.

## Notes

- Keep persistent `TSNET_STATE_DIR` to preserve node identity
- Use distinct `BRIDGE_NAME` per host
- Lock down communication with Tailscale ACLs before production
- Payload limit is 1 MiB per message
- Returns `502` if destination bridge is unreachable
- Returns `400` for invalid envelope format

## Documentation

- Agent communication flow: `tailscale-app/docs/agent-communication.md`
- Engineering notebook: `engineering-notebook/README.md`

## License

MIT
