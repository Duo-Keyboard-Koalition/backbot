# DarCI Go Agent

Go implementation of the DarCI agent with tsA2A protocol support.

## Features

- ✅ **AIP Protocol** - Agent Identification Protocol for secure registration
- ✅ **Handshake Support** - HMAC-SHA256 challenge-response verification
- ✅ **Heartbeat** - Automatic heartbeat to maintain active status
- ✅ **Auto-reconnect** - Exponential backoff retry on connection failure

## Quick Start

### 1. Generate Secret on Bridge

On your bridge machine:

```bash
taila2a secrets generate darci-go-001
# Output: tskey-secret-abc123...
```

### 2. Configure Agent

Set environment variables:

```bash
export DARCI_AGENT_ID="darci-go-001"
export DARCI_AGENT_SECRET="tskey-secret-abc123..."
export DARCI_BRIDGE_URL="http://127.0.0.1:8080"
export DARCI_LISTEN_ADDR=":9090"
```

### 3. Build

```bash
cd agent/darci-go
go build -o darci ./cmd/darci
```

### 4. Run

```bash
./darci
```

### 5. Approve Agent

On the bridge machine:

```bash
taila2a aip pending
taila2a aip approve darci-go-001
```

## Architecture

```
┌─────────────────┐     ┌──────────────────┐
│  DarCI Go Agent │     │  taila2a Bridge  │
│                 │     │                  │
│  :9090          │     │  :8080 (local)   │
│  /aip/handshake │◀───▶│  /aip/register   │
│  /health        │     │  /aip/heartbeat  │
│                 │     │                  │
│  AIP Client     │     │  Handshake Svc   │
│  - Register     │     │  - Challenge     │
│  - Heartbeat    │     │  - Verify        │
└─────────────────┘     └──────────────────┘
```

## Protocol Flow

```
1. Agent starts, listens on :9090
2. Bridge sends handshake challenge → /aip/handshake
3. Agent computes HMAC-SHA256 signature
4. Bridge verifies signature → Agent identified
5. Agent registers → /aip/register
6. Admin approves → taila2a aip approve
7. Agent sends heartbeat every 30s → /aip/heartbeat
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DARCI_AGENT_ID` | `darci-go-001` | Unique agent identifier |
| `DARCI_AGENT_SECRET` | *(required)* | Shared HMAC secret |
| `DARCI_BRIDGE_URL` | `http://127.0.0.1:8080` | Bridge URL |
| `DARCI_LISTEN_ADDR` | `:9090` | Agent HTTP listen address |
| `DARCI_CAPABILITIES` | `task-execution,notebook,file-ops,shell` | Agent capabilities |

### Example .env File

```bash
DARCI_AGENT_ID=darci-go-001
DARCI_AGENT_SECRET=tskey-secret-abc123xyz
DARCI_BRIDGE_URL=http://127.0.0.1:8080
DARCI_LISTEN_ADDR=:9090
DARCI_CAPABILITIES=task-execution,notebook,shell
```

## API Endpoints

### `POST /aip/handshake`

Handle challenge from bridge.

**Request:**
```json
{
  "challenge": "abc123...",
  "timestamp": "2026-03-08T10:00:00Z",
  "nonce": "xyz789",
  "bridge_id": "bridge-alpha"
}
```

**Response:**
```json
{
  "signature": "hmac-sha256-result",
  "agent_id": "darci-go-001",
  "agent_type": "darci-go",
  "agent_version": "1.0.0",
  "capabilities": ["task-execution", "notebook", "file-ops", "shell"]
}
```

### `GET /health`

Health check endpoint.

**Response:**
```json
{
  "status": "healthy",
  "agent_id": "darci-go-001",
  "approved": true,
  "last_heartbeat": "2026-03-08T10:00:00Z",
  "timestamp": "2026-03-08T10:01:00Z"
}
```

## Development

### Build

```bash
go build -o darci ./cmd/darci
```

### Test

```bash
go test ./...
```

### Run with Debug Logging

```bash
export DEBUG=1
./darci
```

## Integration with DarCI Python

The Go agent can coexist with DarCI Python:

```bash
# Run both agents on same machine
./darci-go &  # Listens on :9090
python -m darci --port 9091  # Listens on :9091
```

Each agent registers separately with the bridge.

## Troubleshooting

### Registration Pending

```bash
# Check pending registrations
taila2a aip pending

# Approve if pending
taila2a aip approve darci-go-001
```

### Handshake Failed

- Verify `DARCI_AGENT_SECRET` matches bridge secret
- Check bridge is running: `taila2a run`
- Verify network connectivity: `curl http://bridge:8080/aip/agents`

### Heartbeat Not Sending

- Check agent logs for errors
- Verify bridge approved the agent
- Check firewall allows outbound to bridge port

## Related Documentation

- [tsA2A Handshake Protocol](../../tsa2a/tailbridge/taila2a/docs/tsa2a-handshake-protocol.md)
- [AIP Specification](../../tsa2a/tailbridge/taila2a/docs/aip-agent-identification-protocol.md)
- [AIP Quickstart](../../tsa2a/tailbridge/taila2a/AIP-QUICKSTART.md)
