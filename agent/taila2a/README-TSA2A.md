# tsA2A Protocol - Complete Implementation Summary

**Date:** 2026-03-08  
**Status:** ✅ Complete  
**Version:** 1.0.0

---

## Problem Solved

**Original Issue:**
> "The bot A2A stuff is picking up everything on the network... scanning all IPs"

The discovery service was:
- Port-scanning every machine on the Tailscale network
- Identifying all devices, not just DarCI agents
- Creating unnecessary network traffic
- Posing security concerns

---

## Solution: tsA2A Handshake Protocol

**Key Principle:**
> **NO auto-discovery scanning.** Agents are identified via **challenge-response handshake** + **explicit registration**.

### Two-Layer Verification

```
Layer 1: Handshake (Prove you're an agent)
  ↓
Layer 2: Registration (Tell me about yourself)
  ↓
Layer 3: Approval (Admin verifies)
```

---

## Protocol Details

### Handshake Flow

```
┌─────────────┐                          ┌─────────────┐
│   Bridge    │                          │   Agent     │
│             │                          │             │
│ 1. Generate │                          │             │
│    challenge │                          │             │
│             │                          │             │
│ 2. POST     │                          │             │
│    /aip/handshake                      │             │
│    {challenge, timestamp, nonce}       │             │
│ ──────────────────────────────────────▶│             │
│                                        │             │
│                          3. Compute HMAC-SHA256      │
│                             signature = HMAC(        │
│                               challenge:timestamp:   │
│                               nonce,                 │
│                               AGENT_SECRET         │
│                             )                        │
│                                        │             │
│ 4. Response                            │             │
│    {signature, agent_id, agent_type}  │             │
│ ◀───────────────────────────────────── │             │
│                                        │             │
│ 5. Verify signature with stored secret │             │
│    ✓ Valid → Agent identified!        │             │
│    ✗ Invalid → Ignore/log             │             │
│                                        │             │
```

### Signature Computation

```python
# Python (Agent)
def compute_signature(challenge, timestamp, nonce, secret):
    message = f"{challenge}:{timestamp}:{nonce}"
    signature = hmac.new(
        secret.encode(),
        message.encode(),
        hashlib.sha256
    ).hexdigest()
    return signature
```

```go
// Go (Bridge)
func verifySignature(challenge, timestamp, nonce, signature, secret string) bool {
    message := fmt.Sprintf("%s:%s:%s", challenge, timestamp, nonce)
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write([]byte(message))
    expectedSignature := hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
```

---

## Files Created

### Documentation (4 files)

| File | Description |
|------|-------------|
| `docs/aip-agent-identification-protocol.md` | Full AIP specification |
| `docs/tsa2a-handshake-protocol.md` | Handshake protocol specification |
| `AIP-QUICKSTART.md` | User quickstart guide |
| `AIP-IMPLEMENTATION.md` | Implementation summary |

### Go Bridge (8 files)

| File | Description | Lines |
|------|-------------|-------|
| `bridge/registry.go` | Agent registry store | ~300 |
| `bridge/aip_handlers.go` | AIP HTTP handlers | ~250 |
| `bridge/handshake.go` | Challenge-response service | ~350 |
| `bridge/secrets.go` | Secret management | ~150 |
| `bridge/aip_command.go` | CLI commands | ~500 |
| `bridge/discovery.go` | Passive discovery (updated) | +80 |
| `bridge/app.go` | Integration (updated) | +50 |
| `bridge/main.go` | CLI (updated) | +30 |

**Total:** ~1,710 lines of Go code

### Python Agent (2 files)

| File | Description | Lines |
|------|-------------|-------|
| `scripts/aip_client.py` | Async AIP client | ~300 |
| `scripts/register-agent.sh` | Registration script | ~100 |

**Total:** ~400 lines of Python/Bash

---

## CLI Commands

### AIP Management

```bash
# List agents
taila2a aip list          # All registered agents
taila2a aip pending       # Pending approvals
taila2a aip info <id>     # Agent details

# Approve/Reject
taila2a aip approve <id>  # Approve registration
taila2a aip reject <id>   # Reject registration
taila2a aip remove <id>   # Remove agent
```

### Secrets Management

```bash
# Generate secret for new agent
taila2a secrets generate darci-python-001
# Output: tskey-secret-abc123...

# Show existing secret
taila2a secrets show darci-python-001

# List all agents with secrets
taila2a secrets list

# Remove secret
taila2a secrets remove darci-python-001
```

---

## HTTP API Reference

### Agent Endpoints (Local)

#### `POST /aip/handshake`

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
  "agent_id": "darci-python-001",
  "agent_type": "darci-python",
  "capabilities": ["task-execution", "notebook"]
}
```

#### `POST /aip/register`

Register agent with bridge.

**Request:**
```json
{
  "agent_id": "darci-python-001",
  "agent_type": "darci-python",
  "endpoints": {"primary": "http://127.0.0.1:9090/api"},
  "capabilities": ["task-execution"],
  "metadata": {"hostname": "workstation"}
}
```

**Response:**
```json
{
  "status": "pending",
  "message": "Awaiting admin approval"
}
```

#### `POST /aip/heartbeat`

Send periodic heartbeat.

**Request:**
```json
{
  "agent_id": "darci-python-001",
  "timestamp": "2026-03-08T10:00:00Z",
  "status": "healthy",
  "metrics": {"cpu": 0.15, "memory_mb": 256}
}
```

### Bridge Endpoints (Tailnet)

#### `GET /aip/agents`

List registered agents.

**Response:**
```json
[
  {
    "agent_id": "darci-python-001",
    "status": "approved",
    "last_heartbeat": "2026-03-08T10:00:00Z"
  }
]
```

#### `POST /aip/approve/{agent_id}`

Approve pending registration.

#### `POST /aip/handshake-probe`

Probe IP for agents.

**Request:**
```json
{
  "target_ip": "100.64.1.23",
  "target_port": 9090
}
```

---

## Configuration

### File Structure

```
~/.tailtalkie/
├── config.json              # Bridge configuration
└── state/
    ├── registry.json        # Agent registry
    ├── agent_secrets.json   # Shared secrets
    └── buffer/              # Message buffer
```

### Config Example

```json
{
  "bridge_name": "bridge-alpha",
  "state_dir": "/home/user/.tailtalkie/state",
  "auth_key": "tskey-auth-xxx",
  "local_agent_url": "http://127.0.0.1:9090/api",
  "inbound_port": 8001,
  "local_listen": "127.0.0.1:8080"
}
```

### Secrets Example

```json
{
  "secrets": {
    "darci-python-001": "tskey-secret-abc123...",
    "darci-go-001": "tskey-secret-def456..."
  }
}
```

---

## Security Features

| Feature | Implementation |
|---------|----------------|
| **No scanning** | Passive discovery only |
| **Authentication** | HMAC-SHA256 signatures |
| **Replay prevention** | Nonce tracking |
| **Time validation** | 30s challenge expiry |
| **Secret management** | Per-agent secrets, file mode 0600 |
| **Admin approval** | Manual registration approval |
| **Audit trail** | Registry tracks all events |

---

## Testing Checklist

- [ ] Build bridge: `go build -o taila2a ./bridge`
- [ ] Generate secret: `taila2a secrets generate test-agent`
- [ ] Start bridge: `taila2a run`
- [ ] Send handshake challenge
- [ ] Verify signature response
- [ ] Register agent via curl
- [ ] List pending: `taila2a aip pending`
- [ ] Approve: `taila2a aip approve test-agent`
- [ ] Send heartbeat
- [ ] Verify in registry: `taila2a aip registry`

---

## Integration Guide

### For DarCI Python

```python
from aip_client import AIPClient

client = AIPClient(
    agent_id="darci-python-001",
    agent_secret="tskey-secret-xxx"
)

# Register on startup
await client.register()

# Start heartbeat (30s interval)
await client.start_heartbeat(30)
```

### For DarCI Go

```go
import "darci/channels/aip"

client := aip.NewClient(
    "darci-go-001",
    "tskey-secret-xxx"
)

// Register
client.Register()

// Start heartbeat
go client.HeartbeatLoop(30 * time.Second)
```

---

## Migration Path

### From Auto-Discovery

1. **Stop bridge**
2. **Update**: New version with AIP + handshake
3. **Generate secrets**: `taila2a secrets generate <agent>`
4. **Configure agents**: Add secrets to agent config
5. **Restart bridge**: AIP enabled automatically
6. **Agents register**: Automatic on startup
7. **Admin approves**: `taila2a aip approve <agent>`

### Backward Compatibility

- Old discovery → Passive mode (no scanning)
- Existing peers → Still work
- New agents → Must use AIP + handshake

---

## Metrics & Monitoring

### Stats Endpoint

```bash
curl http://127.0.0.1:8001/aip/agents | jq '. | length'
# Number of registered agents
```

### Logs to Monitor

```
[aip] agent registry initialized
[handshake] service initialized with N agent secrets
[aip] Registration status: pending
[aip] Agent approved!
[handshake] Agent verified: darci-python-001
```

---

## Future Enhancements

1. **PKI-based auth**: Replace shared secrets with asymmetric crypto
2. **Automatic secret rotation**: Periodic secret updates
3. **Multi-bridge support**: Agents registered across bridges
4. **Capability-based routing**: Route by agent capabilities
5. **Health dashboard**: Web UI for agent monitoring

---

## Related Documents

- [[Handshake Protocol Spec](./docs/tsa2a-handshake-protocol.md)]
- [[AIP Specification](./docs/aip-agent-identification-protocol.md)]
- [[Quickstart Guide](./AIP-QUICKSTART.md)]
- [[Agent Communication](./docs/agent-communication.md)]

---

## Contacts

- Implementation: Engineering Team
- Documentation: engineering-notebook/notebooks/2026-03-08-darci-agent-test-status.md
