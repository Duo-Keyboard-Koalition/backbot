# TSA2A Documentation Index

**Tailscale Agent-to-Agent Protocol**

---

## 📚 Documentation Overview

Welcome to TSA2A - the secure, zero-config protocol for AI agent communication over Tailscale.

> "No HTTPS, just TSA2A"

---

## 📖 Getting Started

| Document | Description | For |
|----------|-------------|-----|
| [Quickstart](TSA2A-QUICKSTART.md) | Get agents talking in 5 minutes | Everyone |
| [Protocol Spec](TSA2A-PROTOCOL.md) | Full protocol specification | Implementers |
| [Architecture](TSA2A-ARCHITECTURE.md) | System design and brainstorming | Architects |
| [Authentication](TSA2A-AUTH.md) | Security and authorization model | Security engineers |

---

## 🚀 Quick Links

### New to TSA2A?
1. Start with the [Quickstart Guide](TSA2A-QUICKSTART.md)
2. Read the [Architecture overview](TSA2A-ARCHITECTURE.md)
3. Try the [examples](../examples/)

### Implementing TSA2A?
1. Read the [Protocol Specification](TSA2A-PROTOCOL.md)
2. Review the [Authentication design](TSA2A-AUTH.md)
3. Check the [Go implementation](../go/) or [Python implementation](../python/)

### Security focused?
1. Read [Authentication & Authorization](TSA2A-AUTH.md)
2. Review the [Security considerations](TSA2A-PROTOCOL.md#8-security-considerations)
3. Check the [Threat model](TSA2A-ARCHITECTURE.md#53-threat-model)

---

## 📋 Document Summary

### [Quickstart Guide](TSA2A-QUICKSTART.md)
- Installation instructions (Go & Python)
- Basic agent setup
- Configuration examples
- Discovery and testing
- Common patterns
- Troubleshooting

### [Protocol Specification](TSA2A-PROTOCOL.md)
- Overview and design philosophy
- Architecture and network topology
- Identity and authentication
- Discovery protocols (mDNS, Tailscale DNS, Coordinator, Gossip)
- Handshake protocol (SYN, SYN-ACK, ACK)
- Message format and types
- Error handling
- Security considerations
- Implementation guidelines

### [Architecture Brainstorm](TSA2A-ARCHITECTURE.md)
- Vision and problem statement
- Core concepts (stack, identity, capabilities)
- Discovery mechanisms (detailed)
- Authentication flow
- Authorization policies
- Handshake protocol details
- Message system
- Implementation considerations
- Use cases
- Open questions

### [Authentication Design](TSA2A-AUTH.md)
- Layered security model
- Identity structure and verification
- Authentication flow
- Authorization policies
- Session management
- Message security (signatures, replay protection)
- Rate limiting
- Audit logging
- Security best practices

---

## 🔑 Key Concepts

### What makes TSA2A different?

| Traditional A2A | TSA2A |
|-----------------|-------|
| HTTPS + TLS | Tailscale WireGuard |
| Certificate management | Tailscale identities |
| Separate auth system | Built-in authorization |
| Complex service discovery | Auto-discovery via mDNS/Tailscale |
| HTTP overhead | Lightweight JSON messages |

### Core Features

1. **Auto-Discovery**: Agents find each other automatically
2. **Zero-Config Security**: Tailscale handles encryption and identity
3. **Capability-Based**: Agents advertise what they can do
4. **Message-Oriented**: Async-first with delivery guarantees
5. **Flexible Authz**: Policy-based authorization

---

## 🏗️ Architecture at a Glance

```
┌─────────────────────────────────────────────────┐
│              Tailscale Network                   │
│                                                  │
│  ┌──────────┐         ┌──────────┐              │
│  │ Agent A  │◄───────►│ Agent B  │              │
│  │ :4848    │  TSA2A  │ :4848    │              │
│  └──────────┘         └──────────┘              │
│       ▲                    ▲                    │
│       │                    │                    │
│       └──── Tailscale ─────┘                    │
│           (encryption + identity)                │
└─────────────────────────────────────────────────┘
```

### Protocol Stack

```
┌─────────────────────────────┐
│  Application (Agent Logic)  │
├─────────────────────────────┤
│  TSA2A Protocol Layer       │
│  - Discovery                │
│  - Handshake                │
│  - Auth                     │
│  - Messages                 │
├─────────────────────────────┤
│  Tailscale (WireGuard)      │
├─────────────────────────────┤
│  Transport (UDP/TCP)        │
└─────────────────────────────┘
```

---

## 📦 Message Format Example

```json
{
  "header": {
    "message_id": "msg_abc123",
    "type": "task_request",
    "version": "1.0",
    "timestamp": "2026-03-08T10:30:00Z",
    "sender": {"agent_id": "agent_coordinator"},
    "recipient": {"agent_id": "agent_worker_1"},
    "correlation_id": "task_xyz",
    "priority": "normal",
    "flags": {"requires_ack": true}
  },
  "payload": {
    "task_type": "code_review",
    "input": {
      "repository": "github.com/example/repo",
      "pull_request": 42
    }
  }
}
```

---

## 🔐 Security Model

```
┌─────────────────────────────────────────┐
│  Layer 1: Tailscale                     │
│  - WireGuard encryption                 │
│  - Certificate auth                     │
├─────────────────────────────────────────┤
│  Layer 2: TSA2A                         │
│  - Authorization policies               │
│  - Capability-based access              │
├─────────────────────────────────────────┤
│  Layer 3: Agent                         │
│  - Task permissions                     │
│  - Data policies                        │
└─────────────────────────────────────────┘
```

---

## 🛠️ Implementations

| Language | Status | Package |
|----------|--------|---------|
| Go | 🟢 In Progress | `github.com/tailbridge/taila2a/go` |
| Python | 🟡 Planned | `tsa2a` (PyPI) |
| Rust | ⚪ Future | - |
| TypeScript | ⚪ Future | - |

---

## 📝 Use Cases

### 1. DarCI Agent Coordination
Coordinate multiple DarCI agents across your Tailscale network for distributed task execution.

### 2. Multi-User Agent Networks
Enable agents from different users to securely collaborate with proper authorization.

### 3. Edge-to-Cloud Communication
Connect edge devices to cloud agents seamlessly over Tailscale.

### 4. Home Lab Automation
Run agents on your home servers that can discover and coordinate with each other.

---

## 🔧 Configuration

### Minimal Config

```yaml
agent_id: "agent_worker_1"
port: 4848
capabilities: ["task_execution"]
```

### Full Config

```yaml
agent_id: "agent_worker_1"
port: 4848

capabilities:
  - name: task_execution
    version: "1.0"
  - name: file_ops
    version: "1.0"

discovery:
  mdns: true
  tailscale_dns: true
  coordinator: false
  gossip: true

auth:
  allowed_users: ["admin@example.com"]
  allowed_tags: ["agent"]
  rate_limit: "60/minute"

session:
  heartbeat_interval: 30s
  idle_timeout: 300s
```

---

## 🤝 Contributing

1. Read the [Protocol Spec](TSA2A-PROTOCOL.md)
2. Pick an [open question](TSA2A-ARCHITECTURE.md#10-open-questions) to answer
3. Implement a feature in Go or Python
4. Submit a PR

---

## 📞 Community

- **GitHub**: https://github.com/tailbridge/taila2a
- **Discord**: https://discord.gg/tailscale
- **Issues**: https://github.com/tailbridge/taila2a/issues

---

## 📄 License

MIT License - see [LICENSE](../LICENSE)

---

## 📅 Changelog

| Date | Version | Changes |
|------|---------|---------|
| 2026-03-08 | 0.1.0 | Initial documentation draft |

---

## 🔮 Roadmap

### Phase 1: Core Protocol (Q2 2026)
- [ ] Go implementation complete
- [ ] Basic discovery (mDNS)
- [ ] Handshake protocol
- [ ] Message routing

### Phase 2: Security (Q3 2026)
- [ ] Authorization policies
- [ ] Rate limiting
- [ ] Audit logging
- [ ] Session management

### Phase 3: Production (Q4 2026)
- [ ] Python implementation
- [ ] Coordinator service
- [ ] Monitoring/observability
- [ ] Performance optimization

### Phase 4: Ecosystem (2027)
- [ ] More language implementations
- [ ] Example applications
- [ ] Community extensions
- [ ] Production deployments

---

**Last Updated**: 2026-03-08
