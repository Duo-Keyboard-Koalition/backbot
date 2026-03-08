# Agent Identification Protocol (AIP)

**Version:** 1.0.0  
**Created:** 2026-03-08  
**Status:** 🟡 Draft

---

## Problem Statement

The current discovery mechanism performs indiscriminate network scanning:
- Scans **all peers** on the tailnet
- Port-scans **every IP address** discovered
- Picks up **all machines** on the network, not just DarCI agents
- Creates unnecessary network traffic and security concerns

## Solution: Explicit Agent Registration

Instead of passive network scanning, agents explicitly **register** with bridges they want to communicate with. This creates a **whitelist-based** discovery model.

---

## Architecture

### Components

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   DarCI Agent   │────▶│   Local Bridge   │◀────│  Registry Store │
│   (Python/Go)   │     │  (taila2a)       │     │  (JSON/SQLite)  │
└─────────────────┘     └──────────────────┘     └─────────────────┘
                               │
                               │ AIP Protocol
                               ▼
                        ┌──────────────────┐
                        │  Peer Bridge     │
                        │  (Whitelisted)   │
                        └──────────────────┘
```

### Key Principles

1. **Opt-in Discovery**: Agents must explicitly register
2. **No Network Scanning**: No port scanning of tailnet peers
3. **Identity Verification**: Verify agent identity via signed tokens
4. **Persistent Registry**: Store approved agents in local database
5. **Manual Approval**: Admin must approve new agent registrations

---

## Registration Protocol

### 1. Agent Registration Request

When an agent starts, it registers with its local bridge:

```http
POST /aip/register
Host: 127.0.0.1:8080
Content-Type: application/json

{
  "agent_id": "darci-python-001",
  "agent_type": "darci-python",
  "agent_version": "1.0.0",
  "capabilities": ["task-execution", "notebook", "file-ops"],
  "endpoints": {
    "primary": "http://127.0.0.1:9090/api",
    "health": "http://127.0.0.1:9090/health"
  },
  "metadata": {
    "hostname": "workstation-alpha",
    "os": "linux",
    "tags": ["development", "primary"]
  },
  "auth_token": "<signed-token>"
}
```

### 2. Bridge Response

```json
{
  "status": "pending",
  "message": "Registration received, awaiting approval",
  "registration_id": "reg-abc123",
  "next_steps": "Contact bridge administrator to approve this registration"
}
```

### 3. Admin Approval

Administrator approves via CLI:

```bash
taila2a aip approve reg-abc123
```

Or via config file edit (manual mode).

### 4. Approved Response

Once approved, agent receives:

```json
{
  "status": "approved",
  "agent_id": "darci-python-001",
  "bridge_identity": "bridge-alpha",
  "peer_bridges": [
    {
      "name": "bridge-beta",
      "address": "bridge-beta:8001",
      "agent_id": "darci-go-001"
    }
  ],
  "heartbeat_interval": 30
}
```

---

## Registry Store

### File Format (`~/.tailtalkie/registry.json`)

```json
{
  "version": "1.0",
  "this_bridge": "bridge-alpha",
  "registered_agents": [
    {
      "agent_id": "darci-python-001",
      "agent_type": "darci-python",
      "status": "approved",
      "registered_at": "2026-03-08T10:00:00Z",
      "approved_at": "2026-03-08T10:05:00Z",
      "last_heartbeat": "2026-03-08T12:00:00Z",
      "endpoints": {
        "primary": "http://127.0.0.1:9090/api"
      },
      "capabilities": ["task-execution", "notebook"],
      "metadata": {
        "hostname": "workstation-alpha",
        "tags": ["development"]
      }
    }
  ],
  "peer_bridges": [
    {
      "name": "bridge-beta",
      "tailnet_address": "bridge-beta:8001",
      "agent_id": "darci-go-001",
      "status": "active",
      "last_seen": "2026-03-08T12:00:00Z"
    }
  ]
}
```

---

## Heartbeat Protocol

Registered agents send periodic heartbeats:

```http
POST /aip/heartbeat
Host: 127.0.0.1:8080
Content-Type: application/json

{
  "agent_id": "darci-python-001",
  "timestamp": "2026-03-08T12:00:00Z",
  "status": "healthy",
  "metrics": {
    "cpu_usage": 0.15,
    "memory_mb": 256,
    "active_tasks": 3
  }
}
```

Bridge marks agent as offline after 3 missed heartbeats.

---

## Peer Bridge Discovery

Bridges exchange peer information via **explicit pairing**:

### Pairing Request

```http
POST /aip/pair
Host: bridge-beta:8001
Content-Type: application/json

{
  "bridge_name": "bridge-alpha",
  "auth_token": "<shared-secret>",
  "agents": [
    {
      "agent_id": "darci-python-001",
      "agent_type": "darci-python",
      "capabilities": ["task-execution"]
    }
  ],
  "request_peers": true
}
```

### Pairing Response

```json
{
  "status": "approved",
  "bridge_name": "bridge-beta",
  "agents": [
    {
      "agent_id": "darci-go-001",
      "agent_type": "darci-go",
      "capabilities": ["notebook", "sentinel"]
    }
  ]
}
```

---

## CLI Commands

### Register Agent (Manual Mode)

```bash
# Generate registration token
taila2a aip token generate --agent-id darci-python-001 --expires 24h

# Register agent manually
taila2a aip register \
  --agent-id darci-python-001 \
  --agent-type darci-python \
  --endpoint http://127.0.0.1:9090/api \
  --bridge bridge-alpha
```

### Approve Registration

```bash
# List pending registrations
taila2a aip pending

# Approve specific registration
taila2a aip approve reg-abc123

# Approve all from known host
taila2a aip approve --hostname workstation-alpha
```

### List Registered Agents

```bash
# All agents
taila2a aip list

# Online only
taila2a aip list --online

# By type
taila2a aip list --type darci-python
```

### Remove Agent

```bash
# Deregister agent
taila2a aip remove darci-python-001

# Force remove (even if active)
taila2a aip remove darci-python-001 --force
```

### Pair Bridges

```bash
# Initiate pairing with another bridge
taila2a aip pair \
  --peer bridge-beta \
  --peer-address bridge-beta:8001 \
  --shared-secret mysecret
```

---

## Security Considerations

### Authentication

1. **Registration Tokens**: HMAC-signed tokens for agent registration
2. **Bridge Pairing**: Shared secrets or mutual TLS for bridge-to-bridge
3. **Heartbeat Validation**: Validate agent_id against registry

### Authorization

1. **Capability-based**: Agents can only access peers they're authorized for
2. **Tag-based Routing**: Route messages based on agent tags
3. **Rate Limiting**: Limit registration attempts per IP

### Audit Logging

All AIP events logged:
```json
{
  "timestamp": "2026-03-08T10:00:00Z",
  "event": "agent_registered",
  "agent_id": "darci-python-001",
  "bridge": "bridge-alpha",
  "ip": "127.0.0.1",
  "status": "pending"
}
```

---

## Migration from Auto-Discovery

### Phase 1: Dual Mode (Current → v1.1)

- Keep auto-discovery for existing agents
- Add AIP registration for new agents
- Admin can view both lists

### Phase 2: AIP Default (v1.2)

- AIP enabled by default
- Auto-discovery deprecated but available
- Warning logs for auto-discovered agents

### Phase 3: AIP Only (v2.0)

- Auto-discovery removed
- All agents must register via AIP
- Cleaner, more secure operation

---

## Implementation Checklist

### Bridge (Go)

- [ ] Create `registry/store.go` for agent registry
- [ ] Add `/aip/register` endpoint
- [ ] Add `/aip/heartbeat` endpoint
- [ ] Add `/aip/pair` endpoint for bridge pairing
- [ ] Implement CLI commands (`aip token`, `aip approve`, etc.)
- [ ] Remove/sunset port scanning in `discovery.go`
- [ ] Add heartbeat monitoring goroutine
- [ ] Add audit logging

### Agent (Python)

- [ ] Add AIP registration client in `darci/channels/aip.py`
- [ ] Add heartbeat service in `darci/heartbeat/`
- [ ] Update agent startup to register with bridge
- [ ] Handle registration pending/approved states

### Agent (Go)

- [ ] Add AIP registration in `darci/channels/aip.go`
- [ ] Add heartbeat goroutine
- [ ] Update agent initialization

### Documentation

- [ ] Update `agent-communication.md` with AIP
- [ ] Add AIP troubleshooting guide
- [ ] Document migration steps

---

## Error Codes

| Code | Meaning | Resolution |
|------|---------|------------|
| `AIP-001` | Agent already registered | Use update endpoint |
| `AIP-002` | Invalid auth token | Regenerate token |
| `AIP-003` | Registration pending approval | Contact admin |
| `AIP-004` | Agent not found | Re-register agent |
| `AIP-005` | Bridge pairing failed | Verify shared secret |
| `AIP-006` | Heartbeat timeout | Check agent health |

---

## Related Documents

- [[Agent Communication Flow](./agent-communication.md)]
- [[Tailscale ACL Configuration](./tailscale-acl.example.json)]
- [[Bridge Configuration Guide](../README.md)]
