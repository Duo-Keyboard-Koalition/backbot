# Taila2a - Secure A2A Protocol over Tailscale

**Taila2a** (Tailscale Agent-to-Agent) is a Kafka-inspired, secure agent-to-agent communication protocol built on Tailscale's zero-trust network.

---

## Quick Start

### 1. Initialize Configuration

```bash
cd tailbridge-service
go run ./taila2a init
```

### 2. Start Taila2a

```bash
go run ./taila2a
# or
go run ./taila2a run
```

### 3. Test Buffer Trigger

```bash
# Check trigger status
curl http://localhost:8080/trigger/status

# Add message to buffer (triggers agent)
curl -X POST http://localhost:8080/buffer/add

# Remove message from buffer
curl -X POST http://localhost:8080/buffer/remove
```

---

## Architecture

### MVC Pattern

```
tailbridge-service/
├── cmd/
│   └── taila2a/           # Main entry point
├── internal/
│   ├── models/            # Model layer
│   ├── controllers/       # Controller layer
│   └── views/             # View layer
└── eng_nbk/               # Engineering documentation
```

### Buffer-Triggered Agent System

When buffer transitions from 0→1 messages:
1. Agent automatically starts
2. TUI shows notification
3. Agent processes messages until buffer empty
4. Agent stops automatically

---

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/send` | POST | Send message to peer |
| `/agents` | GET | List discovered agents |
| `/status` | GET | Get taila2a status |
| `/trigger/status` | GET | Get trigger service status |
| `/trigger/manual` | POST | Manually trigger agent |
| `/trigger/stop` | POST | Stop running agent |
| `/trigger/notifications` | GET | Get TUI notifications |
| `/buffer/add` | POST | Simulate buffer increase (test) |
| `/buffer/remove` | POST | Simulate buffer decrease (test) |

---

## Configuration

Config file: `~/.taila2a/config.json`

```json
{
  "name": "taila2a-alpha",
  "auth_key": "tskey-auth-xxxxx",
  "local_agent_url": "http://127.0.0.1:9090/api",
  "inbound_port": 8001,
  "local_listen": "127.0.0.1:8080"
}
```

---

## A2A Protocol

Taila2a implements a Kafka-inspired A2A (Agent-to-Agent) protocol:

### Message Envelope

```json
{
  "header": {
    "id": "uuid",
    "type": "request|response|event",
    "source_agent": {"id": "taila2a-alpha", "node_id": "..."},
    "dest_agent": {"id": "taila2a-beta", "node_id": "..."},
    "topic": "agent.requests",
    "timestamp": "2026-03-07T12:00:00Z",
    "correlation_id": "uuid",
    "reply_to": "agent.responses"
  },
  "body": {
    "action": "execute_task",
    "payload": {...}
  },
  "security": {
    "signature": "ed25519",
    "public_key": "...",
    "timestamp": "2026-03-07T12:00:00Z"
  }
}
```

### Security Features

1. **Tailscale Identity** - Each node has unique identity
2. **mTLS** - Mutual TLS for all A2A communication
3. **Message Signing** - Ed25519 signatures
4. **ACL Enforcement** - Topic access control
5. **Zero Trust** - No implicit trust between nodes

---

## Engineering Notebook

See [eng_nbk/](eng_nbk/) for detailed documentation:

- [eng_nbk.md](../eng_nbk.md) - Multi-agent orchestration roadmap
- [eng_nbk/A2A_PROTOCOL.md](eng_nbk/A2A_PROTOCOL.md) - Full A2A protocol specification

---

## Build

```bash
cd tailbridge-service
go build -o taila2a ./cmd/taila2a
```

---

## License

Same as the parent project.
