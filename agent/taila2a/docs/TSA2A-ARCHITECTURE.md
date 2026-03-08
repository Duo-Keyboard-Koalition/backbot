# TSA2A Architecture Brainstorm

**Tailscale Agent-to-Agent Communication Protocol**

*Brainstorming Document - 2026-03-08*

---

## 1. Vision

> "What if agents could just... talk to each other? No HTTP, no gRPC, no complex setup. Just secure, automatic, peer-to-peer communication over Tailscale."

TSA2A is a play on A2A (Agent-to-Agent) protocols, but built for the modern era where:
- **Tailscale handles the network** (encryption, NAT traversal, identity)
- **Agents handle the semantics** (tasks, capabilities, coordination)
- **No HTTPS overhead** - just pure, encrypted, authenticated agent communication

---

## 2. Problem Statement

### 2.1 Current State

Today's agent communication looks like:

```
┌─────────┐    HTTP/HTTPS    ┌─────────┐
│ Agent A │◄────────────────►│ Agent B │
└─────────┘                  └─────────┘
     │                            │
     ▼                            ▼
┌─────────┐                  ┌─────────┐
│  TLS    │                  │  TLS    │
│  Cert   │                  │  Cert   │
│  Auth   │                  │  Auth   │
└─────────┘                  └─────────┘
     │                            │
     ▼                            ▼
┌─────────┐                  ┌─────────┐
│  Load   │                  │  Load   │
│Balancer │                  │Balancer │
└─────────┘                  └─────────┘
```

**Problems:**
- Complex PKI and certificate management
- NAT traversal headaches
- Service discovery is separate from auth
- HTTP overhead for simple messages
- No built-in identity federation

### 2.2 TSA2A Vision

```
┌─────────────────────────────────────────────────┐
│              Tailscale Network                   │
│                                                  │
│  ┌─────────┐         ┌─────────┐                │
│  │ Agent A │◄───────►│ Agent B │                │
│  │         │  TSA2A  │         │                │
│  └─────────┘         └─────────┘                │
│       ▲                    ▲                    │
│       │                    │                    │
│       └──── Tailscale ─────┘                    │
│           (encryption + identity)                │
└─────────────────────────────────────────────────┘
```

**Benefits:**
- Zero-config encryption (WireGuard)
- Built-in identity (Tailscale users/nodes)
- Automatic NAT traversal
- Simple service discovery
- No certificate management

---

## 3. Core Concepts

### 3.1 The TSA2A Stack

```
┌─────────────────────────────────────────┐
│  Application: Agent Logic (DarCI, etc.) │
├─────────────────────────────────────────┤
│  TSA2A: Message Layer                    │
│  - Discovery                            │
│  - Handshake                            │
│  - Authentication                       │
│  - Message Routing                      │
├─────────────────────────────────────────┤
│  Tailscale: Network Layer                │
│  - WireGuard Encryption                 │
│  - Identity Management                  │
│  - NAT Traversal                        │
│  - DNS Resolution                       │
├─────────────────────────────────────────┤
│  Transport: UDP/TCP                      │
└─────────────────────────────────────────┘
```

### 3.2 Agent Identity

Every agent has a **hierarchical identity**:

```
agent:<tailscale-node>:<user>:<instance>

Example:
  agent:darci-worker-1:admin@example.com:primary
  agent:darci-coordinator:system@example.com:main
```

This identity is **derived from Tailscale** - no separate identity system needed.

### 3.3 Capabilities

Agents advertise what they can do:

```yaml
capabilities:
  - name: task_execution
    version: "1.0"
    input_schema: {...}
    output_schema: {...}
    max_concurrent: 10
    
  - name: file_operations
    version: "1.0"
    operations: [read, write, delete, search]
    allowed_paths: [/workspace/*]
    
  - name: web_search
    version: "1.0"
    providers: [google, duckduckgo]
    rate_limit: 100/hour
```

---

## 4. Discovery Mechanisms

### 4.1 The Discovery Problem

"How does Agent A find Agent B?"

TSA2A supports **multiple discovery methods** that work together:

```
                    ┌─────────────┐
                    │   Agent A   │
                    │  (looking   │
                    │   for B)    │
                    └──────┬──────┘
                           │
         ┌─────────────────┼─────────────────┐
         │                 │                 │
         ▼                 ▼                 ▼
   ┌──────────┐     ┌──────────┐     ┌──────────┐
   │  mDNS    │     │Tailscale │     │Coordinator│
   │ (local)  │     │   DNS    │     │(optional) │
   └──────────┘     └──────────┘     └──────────┘
```

### 4.2 Method 1: mDNS (Local Network)

**Best for:** Same-machine or same-LAN discovery

```
Agent broadcasts on _tsa2a._tcp.local:
  - Agent ID
  - Capabilities
  - Port (4848)
  - Metadata

Listeners respond with their own announcements.
```

**Pros:**
- Zero configuration
- Works offline
- Fast

**Cons:**
- Limited to local network
- No authentication (use Tailscale verification)

### 4.3 Method 2: Tailscale DNS

**Best for:** Network-wide discovery

```
Agents register DNS entries:
  <agent-id>.<user>.ts.net → Tailscale IP

Query: dig agent-worker-1.admin.ts.net
Response: 100.64.0.1
```

**Pros:**
- Works across Tailscale network
- Integrated with Tailscale auth
- No extra infrastructure

**Cons:**
- Requires Tailscale DNS
- Manual registration (or use API)

### 4.4 Method 3: Coordinator Service

**Best for:** Large deployments, dynamic environments

```
┌─────────┐    POST /register    ┌─────────────┐
│  Agent  │─────────────────────►│ Coordinator │
│         │◄─────────────────────│   Service   │
└─────────┘   GET /agents        └─────────────┘
```

**Coordinator API:**
```http
POST /api/v1/agents
Content-Type: application/json

{
  "agent_id": "agent_worker_1",
  "tailscale_ip": "100.64.0.1",
  "port": 4848,
  "capabilities": ["task_execution", "file_ops"],
  "metadata": {"load": 0.5, "status": "available"}
}

Response: 201 Created
Location: /api/v1/agents/agent_worker_1
```

```http
GET /api/v1/agents?capability=task_execution

Response: 200 OK
{
  "agents": [
    {"agent_id": "agent_worker_1", "tailscale_ip": "100.64.0.1", ...},
    {"agent_id": "agent_worker_2", "tailscale_ip": "100.64.0.2", ...}
  ]
}
```

**Pros:**
- Centralized view of all agents
- Can implement load balancing
- Rich queries (by capability, load, etc.)

**Cons:**
- Single point of failure (mitigate with replication)
- Extra infrastructure

### 4.5 Method 4: Gossip Protocol

**Best for:** Distributed, resilient discovery

```
Agent A knows about B and C
Agent D connects to A
A tells D about B and C
D now knows about A, B, C
```

**Gossip Message:**
```json
{
  "type": "peer_gossip",
  "known_peers": [
    {"agent_id": "agent_b", "tailscale_ip": "100.64.0.2", "seen": "..."},
    {"agent_id": "agent_c", "tailscale_ip": "100.64.0.3", "seen": "..."}
  ],
  "hop_count": 2,
  "max_hops": 5
}
```

**Pros:**
- No single point of failure
- Scales well
- Self-healing

**Cons:**
- Eventual consistency
- Can spread stale info

---

## 5. Authentication & Authorization

### 5.1 The Security Model

```
┌─────────────────────────────────────────────────────┐
│  Layer 1: Tailscale (Network Security)              │
│  - WireGuard encryption                            │
│  - Certificate-based authentication                 │
│  - User/node identity                               │
├─────────────────────────────────────────────────────┤
│  Layer 2: TSA2A (Application Security)              │
│  - Authorization policies                           │
│  - Capability-based access control                  │
│  - Message signatures                               │
├─────────────────────────────────────────────────────┤
│  Layer 3: Agent (Task Security)                     │
│  - Task-level permissions                           │
│  - Data handling policies                           │
│  - Audit logging                                    │
└─────────────────────────────────────────────────────┘
```

### 5.2 Authentication Flow

```
┌─────────┐                              ┌─────────┐
│  Agent A │                              │  Agent B │
└────┬────┘                              └────┬────┘
     │                                        │
     │  1. TCP connection over Tailscale      │
     │───────────────────────────────────────►│
     │                                        │
     │  2. Get peer credentials from Tailscale│
     │     (local API: tailscale status)      │
     │                                        │
     │  3. Verify:                            │
     │     - Is peer in allowed users?        │
     │     - Is peer in allowed nodes?        │
     │     - Does peer have required tags?    │
     │                                        │
     │  4. Auth OK / Auth Failed              │
     │◄───────────────────────────────────────│
     │                                        │
```

### 5.3 Authorization Policies

```yaml
# tsaa-policy.yaml
policies:
  default: deny
  
  rules:
    # Allow all agents from admin@example.com
    - name: allow_admin
      match:
        user: "admin@example.com"
      action: allow
      
    # Allow agents with 'worker' tag
    - name: allow_workers
      match:
        tags: ["worker"]
      action: allow
      
    # Allow specific nodes
    - name: allow_specific
      match:
        nodes: ["darci-server", "darci-coordinator"]
      action: allow
      
    # Deny blocked users
    - name: deny_blocked
      match:
        user: "blocked@example.com"
      action: deny
      
    # Rate limit unknown users
    - name: rate_limit_others
      match:
        user: "*"
      action: rate_limit
      rate: 10/minute
```

### 5.4 Capability-Based Security

Not just "can you connect" but "what can you do":

```yaml
capability_policies:
  task_execution:
    allowed_users: ["admin@example.com", "agents@example.com"]
    max_concurrent_per_user: 10
    
  file_operations:
    allowed_users: ["admin@example.com"]
    allowed_paths: ["/workspace/*"]
    denied_operations: [delete]
    
  web_search:
    allowed_users: ["*"]
    rate_limit: 100/hour
```

---

## 6. Handshake Protocol

### 6.1 The Handshake Dance

```
Agent A                              Agent B
   │                                    │
   │  SYN                               │
   │  "Hello, I'm A"                    │
   │  ───────────────────────────────►  │
   │                                    │
   │  SYN-ACK                           │
   │  "Hello A, I'm B"                  │
   │  "Here's what I can do"            │
   │  ◄───────────────────────────────  │
   │                                    │
   │  ACK                               │
   │  "Great, let's work together"      │
   │  ───────────────────────────────►  │
   │                                    │
   │  [Session Active]                  │
   │                                    │
```

### 6.2 Why Handshake?

1. **Capability negotiation**: "I can do X" / "I need X"
2. **Protocol version**: Ensure compatibility
3. **Session config**: Heartbeat, timeouts, limits
4. **Mutual verification**: Both sides agree to communicate

### 6.3 Handshake State Machine

```
                    ┌──────────┐
                    │  CLOSED  │
                    └────┬─────┘
                         │ Send SYN
                         ▼
                    ┌──────────┐
              ┌─────│  SYN_SENT│
              │     └────┬─────┘
         SYN  │          │ Receive SYN-ACK
              │          ▼
         ┌────┴─────┐  ┌──────────┐
         │ SYN_RCVD │◄─│  WAIT_ACK│
         └────┬─────┘  └────┬─────┘
              │             │ Send ACK
              │             ▼
              │        ┌──────────┐
              └───────►│  ACTIVE  │
                       └──────────┘
```

---

## 7. Message System

### 7.1 Message Envelope

Every message has:
- **Header**: Routing, metadata, security
- **Payload**: Actual content
- **Signature**: Optional non-repudiation

```json
{
  "header": {
    "message_id": "msg_abc123",
    "type": "task_request",
    "version": "1.0",
    "timestamp": "2026-03-08T10:30:00Z",
    "sender": {"agent_id": "agent_a", "hostname": "host1"},
    "recipient": {"agent_id": "agent_b", "hostname": "host2"},
    "correlation_id": "corr_xyz",
    "ttl": 3600,
    "priority": "normal",
    "flags": {"requires_ack": true}
  },
  "payload": {...},
  "signature": "base64..."
}
```

### 7.2 Message Types

**Control Messages:**
- `handshake_syn`, `handshake_syn_ack`, `handshake_ack`
- `heartbeat`, `heartbeat_ack`
- `disconnect`

**Discovery Messages:**
- `discovery_announce`
- `discovery_query`
- `discovery_response`

**Task Messages:**
- `task_request`
- `task_response`
- `task_cancel`
- `task_status`

**Error Messages:**
- `error`

### 7.3 Delivery Guarantees

TSA2A supports different delivery modes:

| Mode | Guarantee | Use Case |
|------|-----------|----------|
| `at_most_once` | No retries, may lose | Heartbeats, status |
| `at_least_once` | Retries until ACK | Task requests |
| `exactly_once` | Deduplication + ACK | Critical operations |

---

## 8. Implementation Considerations

### 8.1 Language Support

TSA2A should be implementable in:
- **Go**: Native Tailscale integration
- **Python**: Easy agent development
- **Rust**: High-performance implementations
- **TypeScript**: Web-based agents

### 8.2 Reference Implementation

```
tailbridge/taila2a/
├── go/
│   ├── cmd/tsa2a/         # CLI tool
│   ├── pkg/tsa2a/         # Go library
│   │   ├── discovery/     # Discovery protocols
│   │   ├── handshake/     # Handshake logic
│   │   ├── auth/          # Authentication
│   │   ├── message/       # Message handling
│   │   └── transport/     # Tailscale transport
│   └── examples/
├── python/
│   ├── tsa2a/             # Python library
│   └── examples/
└── docs/
    ├── spec/              # Protocol spec
    └── guides/            # Implementation guides
```

### 8.3 Port Selection

**Proposed:** 4848

Why?
- TSA2A = 4 letters → 48
- 48 repeated → 4848
- Not a well-known port
- Easy to remember

**Alternatives:**
- 8484 (visual symmetry)
- 4448 (44 = HTTPS-ish, 48 = TSA2A)

---

## 9. Use Cases

### 9.1 DarCI Agent Coordination

```
┌─────────────────┐
│   Coordinator   │
│     Agent       │
└────────┬────────┘
         │ TSA2A
         │
    ┌────┴────┬────────────┬────────────┐
    │         │            │            │
    ▼         ▼            ▼            ▼
┌───────┐ ┌───────┐ ┌───────────┐ ┌───────────┐
│Worker │ │Worker │ │  Worker   │ │  Worker   │
│  Python│ │  Go   │ │Specialized│ │Specialized│
└───────┘ └───────┘ └───────────┘ └───────────┘
```

### 9.2 Multi-User Agent Network

```
User A's Agents              User B's Agents
┌──────────┐                 ┌──────────┐
│  Agent A1│◄────TSA2A─────►│  Agent B1│
└──────────┘   (authorized)  └──────────┘
     │                             │
     │                             │
┌──────────┐                 ┌──────────┐
│  Agent A2│                 │  Agent B2│
└──────────┘                 └──────────┘
```

### 9.3 Edge-to-Cloud

```
┌─────────────────────────────────────────────────┐
│              Tailscale Network                   │
│                                                  │
│  ┌─────────┐                                    │
│  │  Edge   │◄────TSA2A─────►┌─────────┐         │
│  │ Device  │                │  Cloud  │         │
│  │ Agent   │                │  Agent  │         │
│  └─────────┘                └─────────┘         │
│                                                  │
└─────────────────────────────────────────────────┘
```

---

## 10. Open Questions

### 10.1 Authentication

- [ ] Should we support non-Tailscale auth? (API keys, OAuth)
- [ ] How to handle Tailscale identity expiration?
- [ ] Should we implement message-level signatures?

### 10.2 Discovery

- [ ] Which discovery method should be default?
- [ ] How to handle discovery in disconnected environments?
- [ ] Should we support DNS-SD (RFC 6763)?

### 10.3 Message Delivery

- [ ] Is exactly-once delivery necessary or too complex?
- [ ] How long to retry failed messages?
- [ ] Should we support message persistence?

### 10.4 Performance

- [ ] What's the max message size?
- [ ] Should we support streaming for large payloads?
- [ ] Compression: always, never, or negotiable?

### 10.5 Security

- [ ] Rate limiting strategy?
- [ ] How to handle compromised agents?
- [ ] Audit logging requirements?

---

## 11. Next Steps

### Phase 1: Core Protocol
- [ ] Finalize message format
- [ ] Implement handshake in Go
- [ ] Implement basic discovery (mDNS)
- [ ] Test agent-to-agent communication

### Phase 2: Authentication
- [ ] Implement Tailscale-based auth
- [ ] Add authorization policies
- [ ] Test with multiple users

### Phase 3: Production Features
- [ ] Add coordinator service
- [ ] Implement gossip discovery
- [ ] Add monitoring/observability
- [ ] Performance optimization

### Phase 4: Ecosystem
- [ ] Python implementation
- [ ] Documentation
- [ ] Example applications
- [ ] Community feedback

---

## Appendix: Naming

**TSA2A** can stand for:
- Tailscale Agent-to-Agent
- TailScale A2A
- The Secure Agent-to-Agent protocol

**Pronunciation:** "tsa-tsa" (like the fly) or "T-S-A-2-A"

**Logo idea:** Two agents connected by a Tailscale wire, forming a handshake
