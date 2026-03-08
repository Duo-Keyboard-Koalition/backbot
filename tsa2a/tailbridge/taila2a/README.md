# tsnet Peer-to-Peer Bridges

This folder contains a peer-to-peer implementation that embeds Tailscale directly into Go applications using `tsnet`.

- No `tailscaled` service required.
- No host Tailscale app required.
- Each binary is its own Tailnet node.
- Configuration stored in `~/.tailtalkie/config.json`
- Auto-discovers agents and gateways on your tailnet

## Quick Install (Linux)

One-line installation:

```bash
curl -fsSL https://raw.githubusercontent.com/codejedi-ai/tailtailkie/main/tailscale-app/scripts/install.sh | sudo bash
```

This will:
- Install Go (if not present)
- Build and install the bridge
- Create systemd service
- Set up configuration directory

After installation:
```bash
# Configure your bridge
sudo walkie-talkie-bridge init

# Start the service
sudo systemctl start walkie-talkie-bridge

# View status
sudo systemctl status walkie-talkie-bridge
```

## Prerequisites

Before running bridges, complete the setup in **PREREQUISITES.md**:

- Go 1.25+
- Tailscale account with auth keys
- Local agent with HTTP endpoint
- Ports 8001, 8080 available

See [PREREQUISITES.md](PREREQUISITES.md) for detailed setup instructions.

## Components

- `bridge`: per-host peer bridge node
- `protocol/envelope.go`: JSON envelope schema

## Envelope format

```json
{
  "source_node": "bridge-alpha",
  "dest_node": "bridge-beta",
  "payload": {"type": "task", "message": "hello"}
}
```

## Quick Start

### 1) Initialize Configuration

Run the interactive setup:

```bash
cd tailscale-app
go run ./bridge init
```

This will:
- Create `~/.tailtalkie/` directory
- Prompt for your Tailscale auth key
- Prompt for bridge name and agent URL
- Save configuration to `~/.tailtalkie/config.json`

### 2) Start the Bridge

```bash
go run ./bridge
```

Or explicitly:

```bash
go run ./bridge run
```

### 3) Send a Message

```bash
curl -sS http://127.0.0.1:8080/send \
  -H 'content-type: application/json' \
  -d '{
    "source_node": "bridge-alpha",
    "dest_node": "bridge-beta",
    "payload": {"message": "hello from alpha"}
  }'
```

## Configuration

### Config File Location

`~/.tailtalkie/config.json`

### Config Schema

```json
{
  "bridge_name": "bridge-alpha",
  "state_dir": "/home/user/.tailtalkie/state",
  "auth_key": "tskey-auth-xxxxx",
  "local_agent_url": "http://127.0.0.1:9090/api",
  "peer_inbound_port": 8001,
  "inbound_port": 8001,
  "local_listen": "127.0.0.1:8080"
}
```

### Fields

| Field | Required | Description |
|-------|----------|-------------|
| `bridge_name` | ✅ | Unique identifier for this bridge |
| `auth_key` | ✅ | Tailscale authentication key |
| `state_dir` | ❌ | Directory for Tailscale state (auto-generated) |
| `local_agent_url` | ❌ | Your local agent's HTTP endpoint |
| `inbound_port` | ❌ | Port for inbound Tailnet connections |
| `peer_inbound_port` | ❌ | Port for outbound peer connections |
| `local_listen` | ❌ | Local endpoint for agent communication |

### Manual Configuration

You can also create the config file manually:

```bash
mkdir -p ~/.tailtalkie
cat > ~/.tailtalkie/config.json <<EOF
{
  "bridge_name": "bridge-alpha",
  "auth_key": "tskey-auth-your-key-here",
  "local_agent_url": "http://127.0.0.1:9090/api",
  "inbound_port": 8001,
  "local_listen": "127.0.0.1:8080"
}
EOF
```

## Commands

| Command | Description |
|---------|-------------|
| `bridge init` | Interactive configuration setup |
| `bridge run` | Start the bridge (default) |
| `bridge help` | Show help message |

## Agent Discovery

The bridge automatically discovers agents and gateways on your tailnet.

### Discover Agents

```bash
# Query the agents endpoint on any bridge
curl http://<bridge-host>:8001/agents
```

Response:
```json
{
  "agents": [
    {
      "name": "bridge-alpha",
      "hostname": "bridge-alpha.tailnet.ts.net",
      "ip": "100.64.0.1",
      "online": true,
      "last_seen": "2026-03-07T12:00:00Z",
      "gateways": [
        {"port": 8001, "protocol": "tcp", "service": "bridge-inbound"},
        {"port": 8080, "protocol": "tcp", "service": "bridge-local"},
        {"port": 9090, "protocol": "tcp", "service": "agent-api"}
      ]
    }
  ],
  "count": 1
}
```

### Discovery Features

- **Auto-detection**: Scans tailnet every 30 seconds
- **Gateway scanning**: Detects open ports on each agent
- **Health monitoring**: Tracks online/offline status
- **Stale cleanup**: Removes agents offline > 5 minutes

## Notes

- The bridge local endpoint is `POST /send` on `local_listen` (default `127.0.0.1:8080`).
- The bridge tailnet endpoint is `POST /inbound` on `inbound_port` (default `8001`).
- Outbound routing is direct bridge-to-bridge using `dest_node` and `peer_inbound_port`.
- State directory persists node identity between restarts.
- Lock down communication with Tailscale ACLs and tags before production use.

## Message Buffer

The bridge includes a persistent message buffer for reliable delivery:

- **Automatic retry**: Failed deliveries are retried with exponential backoff
- **Persistent storage**: Messages survive bridge restarts
- **Background processing**: Non-blocking delivery attempts
- **Monitoring**: Track buffer health via `/buffer/stats` endpoint

See [docs/message-buffer.md](docs/message-buffer.md) for details.

### Buffer Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/buffer/stats` | GET | Buffer statistics |
| `/buffer/messages` | GET | List buffered messages |
| `/buffer/retry` | POST | Retry a failed message |
| `/buffer/clear` | POST | Clear delivered messages |

## Multiple Bridges

To run multiple bridges on the same machine (for testing):

1. Run `bridge init` with different `bridge_name` values
2. Edit `~/.tailtalkie/config.json` to change ports if needed
3. Start each bridge with a different config directory

## Docs

- Setup guide: [PREREQUISITES.md](PREREQUISITES.md)
- Agent communication: [docs/agent-communication.md](docs/agent-communication.md)
