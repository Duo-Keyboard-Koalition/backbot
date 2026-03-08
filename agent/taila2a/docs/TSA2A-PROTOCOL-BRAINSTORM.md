# TSA2A Protocol Analysis & Brainstorm

**Document Type:** Technical Analysis & Design Brainstorm  
**Status:** Draft  
**Date:** 2026-03-08  
**Author:** DarCI Engineering  

---

## Executive Summary

This document analyzes the current TSA2A (Tailscale Agent-to-Agent) protocol implementation and brainstorms improvements for agent discovery, handshaking, and communication. We evaluate whether the current implementation meets the criteria for a robust agent-to-agent communication protocol.

---

## 1. Current State Analysis

### 1.1 What TSA2A Is Today

**TSA2A** is a peer-to-peer communication protocol built on Tailscale's `tsnet` that enables AI agents to:
- Discover each other on a Tailscale network
- Authenticate using shared secrets (HMAC-SHA256)
- Exchange messages via JSON envelopes
- Route messages directly bridge-to-bridge (no central gateway)

### 1.2 Current Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Tailscale Network                         │
│                                                                  │
│  ┌──────────────┐                    ┌──────────────┐          │
│  │   Host A     │                    │   Host B     │          │
│  │              │                    │              │          │
│  │  ┌────────┐  │   TSA2A Protocol   │  ┌────────┐  │          │
│  │  │ Agent  │  │  ◄──────────────►  │  │ Agent  │  │          │
│  │  │   :9090│  │                    │  │   :9090│  │          │
│  │  └───┬────┘  │                    │  └────▲───┘  │          │
│  │      │ HTTP  │                    │  HTTP │      │          │
│  │  ┌───▼────┐  │                    │  ┌────┴───┐  │          │
│  │  │ Bridge │  │                    │  │ Bridge │  │          │
│  │  │ :8080  │  │                    │  │ :8080  │  │          │
│  │  │ :8001  │  │                    │  │ :8001  │  │          │
│  │  └───┬────┘  │                    │  └────▲───┘  │          │
│  │      │ tsnet │                    │ tsnet │      │          │
│  └──────┴────────┴────────────────────┴────────┴──────┘          │
│         100.64.x.x              WireGuard              100.64.y.y│
└─────────────────────────────────────────────────────────────────┘
```

### 1.3 Current Protocol Components

| Component | Status | Implementation |
|-----------|--------|----------------|
| **Discovery** | 🟡 Partial | Port scanning (legacy) + passive mode (AIP) |
| **Handshake** | 🟡 Partial | HMAC challenge-response |
| **Authentication** | 🟡 Partial | Shared secrets (pre-provisioned) |
| **Message Format** | ✅ Complete | JSON envelope |
| **Routing** | ✅ Complete | Direct bridge-to-bridge |
| **Capability Discovery** | ❌ Missing | Not implemented |
| **Session Management** | ❌ Missing | No session state |

---

## 2. Protocol Criteria Evaluation

### 2.1 Required Criteria for Agent-to-Agent Protocol

| Criterion | Current Status | Notes |
|-----------|----------------|-------|
| **1. Agent Discovery** | 🟡 Partial | Can discover peers, but no capability advertisement |
| **2. Identity Management** | 🔴 Limited | Bridge-level identity, not agent-level |
| **3. Authentication** | 🟡 Partial | HMAC with pre-shared secrets (manual provisioning) |
| **4. Authorization** | 🔴 Missing | No fine-grained access control |
| **5. Handshake Protocol** | 🟡 Partial | Challenge-response exists but not standardized |
| **6. Capability Negotiation** | ❌ Missing | No capability exchange |
| **7. Message Delivery** | ✅ Complete | Direct HTTP with retry buffer |
| **8. Error Handling** | 🟡 Partial | Basic error responses |
| **9. Security** | 🟡 Partial | Relies on Tailscale + HMAC |
| **10. Scalability** | 🟡 Partial | No connection pooling or load balancing |

### 2.2 Gap Analysis

**Critical Gaps:**
1. ❌ No standardized agent identity (only bridge identity)
2. ❌ No capability advertisement or discovery
3. ❌ Manual secret provisioning (doesn't scale)
4. ❌ No session management or connection reuse
5. ❌ No authorization policies beyond "has secret"
6. ❌ No protocol versioning or negotiation

**Nice-to-Have Gaps:**
1. ⚠️ No message compression
2. ⚠️ No binary payload support
3. ⚠️ No streaming for long-running tasks
4. ⚠️ No pub/sub for broadcast scenarios

---

## 3. Brainstorm: Enhanced TSA2A Protocol

### 3.1 Proposed Protocol Stack

```
┌─────────────────────────────────────────────────────────┐
│                  Application Layer                       │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐     │
│  │   DarCI     │  │  Sentinel   │  │  Tailbridge │     │
│  │   Agent     │  │   Monitor   │  │   Gateway   │     │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘     │
├─────────────────────────────────────────────────────────┤
│                   TSA2A Protocol Layer                   │
│  ┌──────────┬──────────┬──────────┬──────────┬────────┐│
│  │Identity  │Discovery │ Handshake│ Message  │Session ││
│  │  Mgmt    │  Service │ Service  │  Router  │ Mgmt   ││
│  └──────────┴──────────┴──────────┴──────────┴────────┘│
├─────────────────────────────────────────────────────────┤
│                   Tailscale (WireGuard)                  │
│         Encryption • Identity • Access Control           │
└─────────────────────────────────────────────────────────┘
```

### 3.2 Enhanced Discovery Protocol

#### 3.2.1 Multi-Mode Discovery

```yaml
discovery:
  # Mode 1: Tailscale DNS (automatic)
  tailscale_dns:
    enabled: true
    domain: "ts.net"
    query: "*.agents.ts.net"
    
  # Mode 2: mDNS (local network)
  mdns:
    enabled: true
    service: "_tsa2a._tcp.local"
    ttl: 120
    
  # Mode 3: Coordinator (optional central registry)
  coordinator:
    enabled: false
    url: "http://coordinator:8888"
    heartbeat_interval: 30s
    
  # Mode 4: Gossip (distributed peer sharing)
  gossip:
    enabled: true
    interval: 60s
    max_peers_to_share: 10
```

#### 3.2.2 Discovery Message Format

```json
{
  "type": "agent_announcement",
  "version": "2.0",
  "timestamp": "2026-03-08T12:00:00Z",
  "agent": {
    "agent_id": "darci-python-001",
    "agent_type": "darci",
    "agent_version": "0.1.4",
    "bridge_id": "bridge-alpha",
    "tailscale_ips": ["100.64.0.1", "fd7a:115c:a1e0::1"],
    "hostname": "darci-server.local",
    "port": 4848,
    "capabilities": [
      {
        "name": "task_execution",
        "version": "1.0",
        "description": "Execute AI tasks with Gemini provider",
        "input_schema": {"type": "object", "properties": {...}},
        "output_schema": {"type": "object", "properties": {...}},
        "max_concurrent": 5,
        "current_load": 0.3
      },
      {
        "name": "file_operations",
        "version": "1.0",
        "operations": ["read", "write", "delete", "list"],
        "restrictions": {"allowed_paths": ["/workspace/*"]}
      },
      {
        "name": "web_search",
        "version": "1.0",
        "providers": ["google", "tavily"],
        "rate_limit": "100/hour"
      }
    ],
    "metadata": {
      "tags": ["python", "general-purpose", "gemini"],
      "region": "us-west",
      "status": "available",
      "uptime_seconds": 86400
    }
  },
  "ttl": 300
}
```

### 3.3 Enhanced Authentication & Authorization

#### 3.3.1 Three-Tier Auth Model

```
┌─────────────────────────────────────────────────────────┐
│  Tier 1: Tailscale Authentication (Network Layer)       │
│  - WireGuard encryption                                 │
│  - Node identity verification                           │
│  - ACL enforcement                                      │
└─────────────────────────────────────────────────────────┘
                          ▼
┌─────────────────────────────────────────────────────────┐
│  Tier 2: TSA2A Authentication (Protocol Layer)          │
│  - HMAC-SHA256 challenge-response                       │
│  - Agent identity verification                          │
│  - Session key establishment                            │
└─────────────────────────────────────────────────────────┘
                          ▼
┌─────────────────────────────────────────────────────────┐
│  Tier 3: Capability Authorization (Application Layer)   │
│  - Capability-level permissions                         │
│  - Task-level authorization                            │
│  - Rate limiting & quotas                               │
└─────────────────────────────────────────────────────────┘
```

#### 3.3.2 Authorization Policy Format

```yaml
# tsaa-auth.yaml
authorization:
  # User-level policies
  allow_users:
    - "admin@example.com"
    - "agents@example.com"
    
  # Node-level policies
  allow_nodes:
    - "darci-server"
    - "darci-worker-*"
    
  # Tag-based policies
  require_tags:
    - "agent"
    - "trusted"
    
  # Capability-level policies
  capability_policies:
    task_execution:
      allow_users: ["admin@example.com"]
      allow_agents: ["darci-worker-*"]
      rate_limit: 100/hour
      
    file_operations:
      allow_users: ["admin@example.com"]
      restrictions:
        allowed_paths: ["/workspace/*"]
        denied_operations: ["delete"]
        
    web_search:
      allow_all: true
      rate_limit: 50/hour
      
  # Deny policies (take precedence)
  deny_users:
    - "blocked@example.com"
```

### 3.4 Enhanced Handshake Protocol

#### 3.4.1 Full Handshake Sequence

```
Agent A (Initiator)                    Agent B (Responder)
        │                                      │
        │  1. DISCOVER (multicast/broadcast)   │
        │─────────────────────────────────────►│
        │                                      │
        │  2. ADVERTISE (capabilities, identity)
        │◄─────────────────────────────────────│
        │                                      │
        │  3. SYN (handshake request, nonce)   │
        │─────────────────────────────────────►│
        │                                      │
        │  4. SYN-ACK (challenge, capabilities)
        │◄─────────────────────────────────────│
        │                                      │
        │  5. ACK (challenge response)         │
        │─────────────────────────────────────►│
        │                                      │
        │  6. ACK-OK (session established)     │
        │◄─────────────────────────────────────│
        │                                      │
        │  [Secure Session Active]             │
        │  [Heartbeat every 30s]               │
        │                                      │
```

#### 3.4.2 Handshake Message Formats

**DISCOVER:**
```json
{
  "type": "discover",
  "version": "2.0",
  "sender": "darci-python-001",
  "search_criteria": {
    "capabilities": ["task_execution"],
    "min_load": 0.0,
    "max_load": 0.8
  }
}
```

**ADVERTISE:**
```json
{
  "type": "advertise",
  "version": "2.0",
  "sender": "darci-worker-001",
  "recipient": "darci-python-001",
  "agent": {
    "agent_id": "darci-worker-001",
    "agent_type": "darci",
    "capabilities": [...],
    "current_load": 0.3,
    "status": "available"
  }
}
```

**SYN:**
```json
{
  "type": "handshake_syn",
  "version": "2.0",
  "sender": "darci-python-001",
  "recipient": "darci-worker-001",
  "nonce": "abc123...",
  "timestamp": "2026-03-08T12:00:00Z",
  "protocol_versions": ["tsa2a/2.0", "tsa2a/1.0"],
  "supported_encodings": ["json", "msgpack"],
  "supported_features": ["streaming", "batch", "priority"]
}
```

**SYN-ACK:**
```json
{
  "type": "handshake_syn_ack",
  "version": "2.0",
  "sender": "darci-worker-001",
  "recipient": "darci-python-001",
  "nonce": "xyz789...",
  "ack_nonce": "abc123...",
  "timestamp": "2026-03-08T12:00:01Z",
  "selected_protocol": "tsa2a/2.0",
  "selected_encoding": "json",
  "challenge": "challenge_string_for_hmac",
  "capabilities": {...},
  "session_config": {
    "heartbeat_interval": 30,
    "max_message_size": 1048576,
    "idle_timeout": 300
  }
}
```

**ACK:**
```json
{
  "type": "handshake_ack",
  "version": "2.0",
  "sender": "darci-python-001",
  "recipient": "darci-worker-001",
  "ack_nonce": "xyz789...",
  "timestamp": "2026-03-08T12:00:02Z",
  "challenge_response": "hmac_signature_here",
  "session_id": "session_abc123"
}
```

### 3.5 Session Management

#### 3.5.1 Session State

```go
type Session struct {
    SessionID       string
    RemoteAgentID   string
    RemoteBridgeID  string
    EstablishedAt   time.Time
    LastActivity    time.Time
    State           SessionState  // HANDSHAKING, ACTIVE, CLOSING, CLOSED
    Capabilities    []Capability
    Config          SessionConfig
    Stats           SessionStats
}

type SessionState int
const (
    SessionHandshaking SessionState = iota
    SessionActive
    SessionClosing
    SessionClosed
)

type SessionConfig struct {
    HeartbeatInterval time.Duration
    IdleTimeout       time.Duration
    MaxMessageSize    int
    Compression       bool
    Encoding          string
}
```

#### 3.5.2 Session Lifecycle

```
┌─────────────┐
│  CREATED    │  (New connection request)
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ HANDSHAKING │  (Exchanging SYN/SYN-ACK/ACK)
└──────┬──────┘
       │
       │ Handshake complete
       ▼
┌─────────────┐
│   ACTIVE    │◄───────┐
└──────┬──────┘        │
       │               │
       │ Heartbeat     │ Message exchange
       ├───────────────┘
       │
       │ Idle timeout / Close request
       ▼
┌─────────────┐
│  CLOSING    │  (Graceful shutdown)
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   CLOSED    │  (Session terminated)
└─────────────┘
```

---

## 4. Implementation Roadmap

### Phase 1: Foundation (Week 1-2)

- [ ] Define TSA2A v2.0 protocol specification
- [ ] Implement agent identity management
- [ ] Add capability advertisement to discovery
- [ ] Create authorization policy engine

### Phase 2: Enhanced Handshake (Week 3-4)

- [ ] Implement full SYN/SYN-ACK/ACK handshake
- [ ] Add session management
- [ ] Implement heartbeat protocol
- [ ] Add connection pooling

### Phase 3: Advanced Features (Week 5-6)

- [ ] Implement message compression
- [ ] Add binary payload support (msgpack)
- [ ] Implement streaming for long tasks
- [ ] Add pub/sub for broadcast

### Phase 4: Production Hardening (Week 7-8)

- [ ] Add comprehensive logging
- [ ] Implement circuit breakers
- [ ] Add rate limiting
- [ ] Performance testing and optimization

---

## 5. Does Current TSA2A Meet Criteria?

### Summary Assessment

| Criteria | Meets? | Notes |
|----------|--------|-------|
| Agent Discovery | 🟡 Partially | Discovers bridges, not agents |
| Identity Management | 🔴 No | Bridge identity only |
| Authentication | 🟡 Partially | HMAC but manual provisioning |
| Authorization | 🔴 No | No fine-grained policies |
| Handshake | 🟡 Partially | Challenge-response exists |
| Capability Negotiation | 🔴 No | Not implemented |
| Message Delivery | ✅ Yes | Works well |
| Error Handling | 🟡 Partially | Basic implementation |
| Security | 🟡 Partially | Relies on Tailscale |
| Scalability | 🟡 Partially | No connection management |

### Overall Verdict

**Current TSA2A is a PROOF OF CONCEPT, not a production-ready protocol.**

**Strengths:**
- ✅ Simple and understandable
- ✅ Leverages Tailscale well
- ✅ Direct peer-to-peer routing works
- ✅ Message buffer for reliability

**Weaknesses:**
- ❌ No agent-level identity
- ❌ Manual secret provisioning doesn't scale
- ❌ No capability discovery
- ❌ No session management
- ❌ Limited error handling

**Recommendation:** Use current TSA2A for development/testing, but implement TSA2A v2.0 (as specified above) for production use.

---

## 6. Next Steps

1. **Review this brainstorm** with the team
2. **Prioritize features** for MVP vs. future releases
3. **Create detailed spec** for TSA2A v2.0
4. **Implement Phase 1** (Foundation)
5. **Test with existing agents** (DarCI, Sentinel, Tailbridge)
6. **Document migration path** from v1.0 to v2.0

---

## Appendix A: Glossary

| Term | Definition |
|------|------------|
| **TSA2A** | Tailscale Agent-to-Agent Protocol |
| **Bridge** | tsnet node that proxies agent traffic |
| **Agent** | AI application (DarCI, Sentinel, etc.) |
| **Capability** | Service an agent provides |
| **Handshake** | Protocol negotiation sequence |
| **Session** | Persistent connection state |

---

## Appendix B: References

- [TSA2A Protocol Spec (draft)](../agent/taila2a/docs/TSA2A-PROTOCOL.md)
- [Agent Communication Flow](../agent/taila2a/docs/agent-communication.md)
- [Message Buffer Design](../agent/taila2a/docs/message-buffer.md)
- [Tailscale tsnet Documentation](https://tailscale.com/kb/1220/tsnet/)

---

*Last Updated: 2026-03-08*  
*Status: Draft for Review*
