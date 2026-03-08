# AIP Quickstart Guide

**Agent Identification Protocol (AIP)** - Secure agent registration without network scanning

---

## Overview

AIP replaces automatic network scanning with explicit agent registration. This provides:

- ✅ **Security**: No unauthorized agents can join
- ✅ **Privacy**: No network-wide port scanning
- ✅ **Control**: Admin approves all registrations
- ✅ **Audit**: Complete registration history

---

## Quick Start

### 1. Start the Bridge

```bash
cd tailbridge/taila2a/bridge
taila2a run
```

### 2. Register an Agent (from agent side)

Agents register via HTTP POST to the local bridge:

```bash
curl -X POST http://127.0.0.1:8080/aip/register \
  -H "Content-Type: application/json" \
  -d '{
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
    }
  }'
```

Response:
```json
{
  "status": "pending",
  "message": "Registration received, awaiting approval",
  "registration_id": "darci-python-001"
}
```

### 3. Approve the Registration

List pending registrations:

```bash
taila2a aip pending
```

Output:
```
Pending Registrations (1):

  Agent ID: darci-python-001
  Type: darci-python
  Version: 1.0.0
  Registered: 2026-03-08T10:00:00Z
  Hostname: workstation-alpha
  Endpoints: http://127.0.0.1:9090/api
  Capabilities: [task-execution notebook file-ops]

  To approve: taila2a aip approve darci-python-001
  To reject:  taila2a aip reject darci-python-001
```

Approve:

```bash
taila2a aip approve darci-python-001
```

### 4. Agent Sends Heartbeat

Once approved, agents send periodic heartbeats:

```bash
curl -X POST http://127.0.0.1:8080/aip/heartbeat \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "darci-python-001",
    "timestamp": "2026-03-08T10:05:00Z",
    "status": "healthy",
    "metrics": {
      "cpu_usage": 0.15,
      "memory_mb": 256,
      "active_tasks": 3
    }
  }'
```

---

## CLI Commands

### List Registered Agents

```bash
# All agents
taila2a aip list

# Pending only
taila2a aip pending
```

### Manage Registrations

```bash
# Approve an agent
taila2a aip approve darci-python-001

# Reject an agent
taila2a aip reject darci-python-001

# Remove an agent
taila2a aip remove darci-python-001

# Get agent details
taila2a aip info darci-python-001
```

### View Registry

```bash
# Show raw registry file
taila2a aip registry
```

---

## HTTP Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/aip/register` | POST | Register a new agent |
| `/aip/heartbeat` | POST | Send agent heartbeat |
| `/aip/agents` | GET | List all registered agents |
| `/aip/approve/{agent_id}` | POST | Approve pending registration |
| `/aip/reject/{agent_id}` | POST | Reject pending registration |
| `/aip/pair` | POST | Bridge-to-bridge pairing |

---

## Agent Types

Supported agent types:

- `darci-python` - DarCI Python agent
- `darci-go` - DarCI Go agent
- `sentinel` - Sentinel monitoring agent
- `custom` - Custom agent implementation

---

## Capabilities

Common capabilities:

- `task-execution` - Can execute tasks
- `notebook` - Engineering notebook access
- `file-ops` - File system operations
- `shell` - Shell command execution
- `sentinel` - Security monitoring
- `web-search` - Web search capability
- `code-execution` - Code execution sandbox

---

## Configuration

### Bridge Config (`~/.tailtalkie/config.json`)

```json
{
  "bridge_name": "bridge-alpha",
  "state_dir": "/home/user/.tailtalkie/state",
  "auth_key": "tskey-auth-xxx",
  "local_agent_url": "http://127.0.0.1:9090/api",
  "peer_inbound_port": 8001,
  "inbound_port": 8001,
  "local_listen": "127.0.0.1:8080"
}
```

### Registry File (`~/.tailtalkie/state/registry.json`)

Stores all registered agents and peer bridges.

---

## Troubleshooting

### Agent registration stuck in "pending"

1. Check bridge is running: `taila2a run`
2. List pending: `taila2a aip pending`
3. Approve manually: `taila2a aip approve <agent_id>`

### Heartbeat rejected

- Agent must be approved first
- Check agent_id matches registration

### Bridge not listening on port 8080

- Check `local_listen` in config
- Ensure port not in use by another service

---

## Migration from Auto-Discovery

If you were using the old auto-discovery:

1. **Stop the bridge**
2. **Backup registry**: `cp ~/.tailtalkie/state/registry.json ~/.tailtalkie/state/registry.json.bak`
3. **Start bridge**: AIP is now enabled by default
4. **Re-register agents**: Each agent must register via `/aip/register`
5. **Approve agents**: Use `taila2a aip approve`

Old peer bridges still work in passive mode (no port scanning).

---

## Security Best Practices

1. **Review registrations**: Always check hostname and capabilities before approving
2. **Use tags**: Tag agents by environment (dev, staging, prod)
3. **Monitor heartbeats**: Set up alerts for missed heartbeats
4. **Audit logs**: Review registry changes periodically
5. **Limit capabilities**: Only grant necessary capabilities per agent

---

## Related Documentation

- [[Agent Identification Protocol](./docs/aip-agent-identification-protocol.md)] - Full protocol specification
- [[Agent Communication](./docs/agent-communication.md)] - How agents communicate
- [[Tailscale ACL](./docs/tailscale-acl.example.json)] - Network access control
