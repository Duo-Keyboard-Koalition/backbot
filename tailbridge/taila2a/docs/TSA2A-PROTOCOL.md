# TSA2A Protocol Specification

**Tailscale Agent-to-Agent Protocol**

Version: 0.1.0-draft  
Status: Brainstorming  
Date: 2026-03-08

---

## 1. Overview

TSA2A (pronounced "tsa-tsa") is a lightweight, secure agent communication protocol built on top of Tailscale. It provides automatic discovery, authentication, and handshaking between AI agents operating within a Tailscale network.

### 1.1 Design Philosophy

> "No HTTPS, just TSA2A"

TSA2A leverages Tailscale's built-in encryption and identity management to eliminate the complexity of TLS/PKI while providing stronger security guarantees than traditional HTTP-based protocols.

### 1.2 Key Features

- **Auto-Discovery**: Agents automatically find each other on the Tailscale network
- **Zero-Config Authentication**: Uses Tailscale identities for mutual authentication
- **Automatic Handshaking**: Protocol negotiation and capability exchange
- **Message-Oriented**: Async-first communication with delivery guarantees
- **Capability-Based**: Agents advertise and discover capabilities dynamically

---

## 2. Architecture

### 2.1 Network Topology

```
┌─────────────────────────────────────────────────────────────────┐
│                        Tailscale Network                         │
│                                                                  │
│  ┌──────────┐         ┌──────────┐         ┌──────────┐        │
│  │  Agent A │◄───────►│  Agent B │◄───────►│  Agent C │        │
│  │ :4848    │  TSA2A  │ :4848    │  TSA2A  │ :4848    │        │
│  └──────────┘         └──────────┘         └──────────┘        │
│       ▲                    ▲                    ▲               │
│       │                    │                    │               │
│       └────────────────────┼────────────────────┘               │
│                            │                                    │
│                     ┌──────┴──────┐                             │
│                     │  Coordinator │                            │
│                     │  (optional)  │                            │
│                     └──────────────┘                            │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 Protocol Stack

```
┌─────────────────────────────────────┐
│      Application Layer (Agents)      │
├─────────────────────────────────────┤
│         TSA2A Protocol Layer         │
│  ┌─────────┬─────────┬───────────┐  │
│  │Discovery│ Handshake│  Message  │  │
│  └─────────┴─────────┴───────────┘  │
├─────────────────────────────────────┤
│      Tailscale (WireGuard)          │
├─────────────────────────────────────┤
│         Transport (UDP/TCP)          │
└─────────────────────────────────────┘
```

---

## 3. Identity & Authentication

### 3.1 Agent Identity

Each agent has a unique identity derived from Tailscale:

```json
{
  "agent_id": "agent_<tailscale_node_id>",
  "tailscale_user": "user@example.com",
  "tailscale_node": "hostname",
  "tailscale_ips": ["100.64.0.1", "fd7a:115c:a1e0::1"],
  "capabilities": ["task_execution", "file_ops", "web_search"],
  "metadata": {
    "agent_type": "darci",
    "version": "1.0.0",
    "tags": ["python", "general-purpose"]
  }
}
```

### 3.2 Authentication Flow

TSA2A uses Tailscale's built-in authentication:

1. **Connection**: Tailscale provides encrypted tunnel automatically
2. **Identity Verification**: Verify peer's Tailscale certificate
3. **Authorization**: Check if peer is in allowed users/nodes list
4. **Session**: Establish authenticated session for message exchange

```
┌─────────┐                              ┌─────────┐
│  Agent A │                              │  Agent B │
└────┬────┘                              └────┬────┘
     │                                        │
     │  1. Tailscale connection (encrypted)   │
     │───────────────────────────────────────►│
     │                                        │
     │  2. Request identity                   │
     │───────────────────────────────────────►│
     │                                        │
     │  3. Return Tailscale cert + metadata   │
     │◄───────────────────────────────────────│
     │                                        │
     │  4. Verify cert with Tailscale         │
     │     Check authorization                │
     │                                        │
     │  5. Auth OK / Auth Failed              │
     │◄───────────────────────────────────────│
     │                                        │
```

### 3.3 Authorization Policies

```yaml
# tsaa-auth.yaml
authorization:
  allow_users:
    - "admin@example.com"
    - "agents@example.com"
  
  allow_nodes:
    - "darci-server"
    - "darci-worker-*"
  
  require_tags:
    - "agent"
  
  deny_users:
    - "blocked@example.com"
```

---

## 4. Discovery Protocol

### 4.1 Discovery Methods

TSA2A supports multiple discovery mechanisms:

#### 4.1.1 Multicast DNS (mDNS) - Local Network

```
Service: _tsa2a._tcp.local
Port: 4848
TXT Records:
  - agent_id=<id>
  - capabilities=<cap1,cap2>
  - version=1.0.0
```

#### 4.1.2 Tailscale DNS - Network-wide

Agents register in Tailscale DNS:

```
<agent-id>.ts.net → 100.64.x.x
```

#### 4.1.3 Coordinator Service - Optional

Central coordinator for agent registration:

```http
POST /api/v1/register
Content-Type: application/json

{
  "agent_id": "agent_abc123",
  "tailscale_ip": "100.64.0.1",
  "port": 4848,
  "capabilities": ["task_execution"],
  "heartbeat_interval": 30
}
```

#### 4.1.4 Peer Gossip - Distributed

Agents share knowledge of other agents:

```json
{
  "type": "peer_announcement",
  "known_peers": [
    {
      "agent_id": "agent_xyz",
      "tailscale_ip": "100.64.0.2",
      "last_seen": "2026-03-08T10:30:00Z"
    }
  ]
}
```

### 4.2 Discovery Message Format

```json
{
  "type": "discovery",
  "version": "1.0",
  "timestamp": "2026-03-08T10:30:00Z",
  "agent": {
    "id": "agent_abc123",
    "hostname": "darci-worker-1",
    "tailscale_ips": ["100.64.0.1"],
    "port": 4848,
    "capabilities": [
      {
        "name": "task_execution",
        "version": "1.0",
        "input_schema": {...},
        "output_schema": {...}
      }
    ],
    "metadata": {
      "agent_type": "darci",
      "version": "1.0.0",
      "tags": ["python", "general-purpose"],
      "load": 0.5,
      "status": "available"
    }
  },
  "ttl": 300
}
```

---

## 5. Handshake Protocol

### 5.1 Handshake Sequence

```
Agent A                              Agent B
   │                                    │
   │  SYN: Hello, I'm agent_A           │
   │  ───────────────────────────────►  │
   │                                    │
   │  SYN-ACK: Hello agent_A,           │
   │           I'm agent_B              │
   │           Here are my capabilities │
   │  ◄───────────────────────────────  │
   │                                    │
   │  ACK: Great, let's talk            │
   │  ───────────────────────────────►  │
   │                                    │
   │  [Session Established]             │
   │                                    │
```

### 5.2 SYN Message

```json
{
  "type": "handshake_syn",
  "version": "1.0",
  "agent_id": "agent_abc123",
  "nonce": "random_128_bit_value",
  "timestamp": "2026-03-08T10:30:00Z",
  "capabilities": {
    "protocols": ["tsa2a/1.0"],
    "encodings": ["json", "msgpack"],
    "features": ["streaming", "batch", "priority"]
  },
  "metadata": {
    "agent_type": "darci",
    "version": "1.0.0"
  }
}
```

### 5.3 SYN-ACK Message

```json
{
  "type": "handshake_syn_ack",
  "version": "1.0",
  "agent_id": "agent_xyz789",
  "nonce": "another_random_value",
  "ack_nonce": "random_128_bit_value",  // Echo from SYN
  "timestamp": "2026-03-08T10:30:01Z",
  "capabilities": {
    "protocols": ["tsa2a/1.0"],
    "encodings": ["json", "msgpack"],
    "features": ["streaming", "batch"]
  },
  "available_capabilities": [
    {
      "name": "task_execution",
      "version": "1.0",
      "max_concurrent": 10
    },
    {
      "name": "file_ops",
      "version": "1.0",
      "supported_ops": ["read", "write", "delete"]
    }
  ],
  "metadata": {
    "agent_type": "darci",
    "version": "1.0.0",
    "load": 0.3,
    "status": "available"
  }
}
```

### 5.4 ACK Message

```json
{
  "type": "handshake_ack",
  "version": "1.0",
  "agent_id": "agent_abc123",
  "ack_nonce": "another_random_value",  // Echo from SYN-ACK
  "timestamp": "2026-03-08T10:30:02Z",
  "selected_capabilities": [
    "task_execution",
    "file_ops"
  ],
  "session_config": {
    "heartbeat_interval": 30,
    "max_message_size": 1048576,
    "compression": false
  }
}
```

---

## 6. Message Format

### 6.1 Envelope Structure

```json
{
  "header": {
    "message_id": "msg_abc123",
    "type": "task_request",
    "version": "1.0",
    "timestamp": "2026-03-08T10:30:00Z",
    "sender": {
      "agent_id": "agent_abc123",
      "hostname": "darci-worker-1"
    },
    "recipient": {
      "agent_id": "agent_xyz789",
      "hostname": "darci-worker-2"
    },
    "correlation_id": "corr_def456",
    "ttl": 3600,
    "priority": "normal",
    "flags": {
      "requires_ack": true,
      "encrypted_payload": false,
      "compressed": false
    }
  },
  "payload": {
    ...
  },
  "signature": "base64_encoded_signature"
}
```

### 6.2 Message Types

| Type | Direction | Description |
|------|-----------|-------------|
| `task_request` | A→B | Request task execution |
| `task_response` | B→A | Task execution result |
| `task_cancel` | A→B | Cancel pending task |
| `capability_query` | A→B | Query available capabilities |
| `capability_response` | B→A | Capability information |
| `heartbeat` | A↔B | Keep-alive signal |
| `peer_announcement` | A→B | New peer discovered |
| `error` | A↔B | Error notification |

### 6.3 Task Request

```json
{
  "header": {
    "message_id": "msg_001",
    "type": "task_request",
    "version": "1.0",
    "timestamp": "2026-03-08T10:30:00Z",
    "sender": {"agent_id": "agent_coordinator"},
    "recipient": {"agent_id": "agent_worker_1"},
    "correlation_id": "task_abc123",
    "priority": "high",
    "flags": {"requires_ack": true}
  },
  "payload": {
    "task_type": "code_review",
    "input": {
      "repository": "github.com/example/repo",
      "pull_request": 42,
      "files": ["src/main.go", "src/utils.go"]
    },
    "parameters": {
      "max_tokens": 4096,
      "temperature": 0.7
    },
    "context": {
      "user": "admin@example.com",
      "session_id": "session_xyz"
    }
  }
}
```

### 6.4 Task Response

```json
{
  "header": {
    "message_id": "msg_002",
    "type": "task_response",
    "version": "1.0",
    "timestamp": "2026-03-08T10:35:00Z",
    "sender": {"agent_id": "agent_worker_1"},
    "recipient": {"agent_id": "agent_coordinator"},
    "correlation_id": "task_abc123",
    "flags": {}
  },
  "payload": {
    "status": "completed",
    "output": {
      "review": "Code looks good. Consider adding error handling...",
      "suggestions": [
        {"file": "src/main.go", "line": 42, "suggestion": "..."}
      ]
    },
    "metrics": {
      "duration_ms": 300000,
      "tokens_used": 2048,
      "model": "gemini-2.0-flash"
    }
  }
}
```

---

## 7. Error Handling

### 7.1 Error Codes

| Code | Name | Description |
|------|------|-------------|
| 1001 | AUTH_FAILED | Authentication failed |
| 1002 | AUTH_EXPIRED | Session expired |
| 2001 | DISCOVERY_FAILED | Could not discover peer |
| 2002 | PEER_UNREACHABLE | Peer not reachable |
| 3001 | HANDSHAKE_FAILED | Handshake negotiation failed |
| 3002 | VERSION_MISMATCH | Protocol version mismatch |
| 4001 | CAPABILITY_NOT_FOUND | Requested capability not available |
| 4002 | CAPABILITY_BUSY | Capability at capacity |
| 5001 | TASK_FAILED | Task execution failed |
| 5002 | TASK_TIMEOUT | Task timed out |
| 5003 | TASK_CANCELLED | Task was cancelled |
| 6001 | MESSAGE_INVALID | Malformed message |
| 6002 | MESSAGE_TOO_LARGE | Message exceeds size limit |

### 7.2 Error Message

```json
{
  "header": {
    "message_id": "msg_err_001",
    "type": "error",
    "version": "1.0",
    "timestamp": "2026-03-08T10:30:00Z",
    "sender": {"agent_id": "agent_worker_1"},
    "recipient": {"agent_id": "agent_coordinator"},
    "correlation_id": "task_abc123"
  },
  "payload": {
    "error_code": 5001,
    "error_name": "TASK_FAILED",
    "message": "Task execution failed: API rate limit exceeded",
    "details": {
      "retry_after": 60,
      "retryable": true
    },
    "stack_trace": "..."
  }
}
```

---

## 8. Security Considerations

### 8.1 Trust Model

- **Tailscale provides**: Encryption, identity, access control
- **TSA2A provides**: Application-level authz, capability security
- **Agents provide**: Task-level security, data handling policies

### 8.2 Security Best Practices

1. **Always verify Tailscale certificates**
2. **Use authorization policies** to restrict which users/nodes can connect
3. **Implement rate limiting** per agent
4. **Log all inter-agent communication** for audit
5. **Use message signatures** for non-repudiation
6. **Implement message expiration** (TTL) to prevent replay

### 8.3 Threat Model

| Threat | Mitigation |
|--------|------------|
| Unauthorized access | Tailscale auth + TSA2A authorization |
| Man-in-the-middle | Tailscale WireGuard encryption |
| Replay attacks | Message nonces + TTL |
| DoS | Rate limiting + circuit breakers |
| Compromised agent | Authorization policies + capability restrictions |

---

## 9. Implementation Guidelines

### 9.1 Port Assignment

- **Default port**: 4848 (TSA2A = 4 letters, 4 digits)
- **Alternative**: 8484 (visual symmetry)

### 9.2 Connection Management

```go
type ConnectionManager struct {
    activeConnections map[AgentID]*Connection
    pendingHandshakes map[Nonce]*HandshakeState
    maxConnections    int
    heartbeatInterval time.Duration
}
```

### 9.3 Message Queue

```go
type MessageQueue struct {
    outgoing chan Message
    incoming chan Message
    pending  map[CorrelationID]*PendingMessage
    maxRetries int
    ackTimeout time.Duration
}
```

---

## 10. Future Extensions

### 10.1 Planned Features

- [ ] Streaming responses for long-running tasks
- [ ] Batch message processing
- [ ] Priority queue support
- [ ] Message compression
- [ ] Binary payload support (msgpack, protobuf)
- [ ] Multi-cast messaging
- [ ] Pub/sub capability discovery

### 10.2 Experimental Features

- [ ] Federated identity (beyond Tailscale)
- [ ] Cross-network bridging
- [ ] Message persistence/replay
- [ ] Distributed task scheduling

---

## Appendix A: Quick Reference

### A.1 Default Configuration

```yaml
tsa2a:
  port: 4848
  discovery:
    mdns: true
    tailscale_dns: true
    coordinator: false
    gossip: true
  handshake:
    timeout: 30s
    max_retries: 3
  session:
    heartbeat_interval: 30s
    idle_timeout: 300s
    max_message_size: 1048576  # 1MB
  security:
    require_auth: true
    allow_users: ["*"]
    allow_tags: ["agent"]
```

### A.2 Common Workflows

**Agent Startup:**
1. Initialize Tailscale connection
2. Register in discovery system
3. Listen on port 4848
4. Accept incoming handshakes

**Task Execution:**
1. Discover capable agents
2. Establish handshake
3. Send task request
4. Wait for response
5. Close or keep connection

**Heartbeat:**
1. Send heartbeat every 30s
2. Expect response within 10s
3. Mark peer dead after 3 missed heartbeats

---

## Appendix B: Changelog

| Version | Date | Changes |
|---------|------|---------|
| 0.1.0 | 2026-03-08 | Initial draft specification |
