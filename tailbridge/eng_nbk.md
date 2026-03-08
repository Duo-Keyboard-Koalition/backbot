# Engineering Notebook (eng_nbk)

## Multi-Agent Orchestration Roadmap

**Vision:** Kafka-inspired secure A2A (Agent-to-Agent) communication over Tailscale

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [MVC Structure](#mvc-structure)
3. [Multi-Agent Roadmap](#multi-agent-roadmap)
4. [A2A Protocol](#a2a-protocol)
5. [Buffer-Triggered Agents](#buffer-triggered-agents)
6. [Design Principles](#design-principles)

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    Tailbridge Multi-Agent System                         │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐              │
│  │   Request    │───▶│   Ingress    │───▶│   Event      │              │
│  │   Gateway    │    │   Controller │    │   Bus        │              │
│  │              │    │              │    │  (Kafka-like)│              │
│  └──────────────┘    └──────────────┘    └──────────────┘              │
│                                              │                          │
│                    ┌─────────────────────────┼─────────────────────────┐│
│                    │                         │                         ││
│                    ▼                         ▼                         ▼│
│           ┌──────────────┐          ┌──────────────┐          ┌──────────────┐
│           │   Agent      │          │   Agent      │          │   Agent      │
│           │   Topic A    │          │   Topic B    │          │   Topic C    │
│           │              │          │              │          │              │
│           │ ┌──────────┐ │          │ ┌──────────┐ │          │ ┌──────────┐ │
│           │ │ Buffer   │ │          │ │ Buffer   │ │          │ │ Buffer   │ │
│           │ │ Trigger  │ │          │ │ Trigger  │ │          │ │ Trigger  │ │
│           │ └──────────┘ │          │ └──────────┘ │          │ └──────────┘ │
│           └──────────────┘          └──────────────┘          └──────────────┘
│                    │                         │                         │
│                    └─────────────────────────┼─────────────────────────┘
│                                              │
│                                              ▼
│                                    ┌──────────────┐
│                                    │   Tailscale  │
│                                    │   Tailnet    │
│                                    │   (Secure)   │
│                                    └──────────────┘
│
└─────────────────────────────────────────────────────────────────────────┘
```

---

## MVC Structure

### Directory Layout

```
tailbridge-service/
├── cmd/
│   └── agents/              # Main entry point
│       ├── main.go          # Application bootstrap
│       └── init.go          # Configuration setup
├── internal/
│   ├── models/              # Model Layer
│   │   ├── models.go        # Core data structures
│   │   ├── config.go        # Configuration management
│   │   ├── discovery.go     # Tailscale discovery
│   │   ├── agent_trigger.go # Buffer-triggered agent service
│   │   ├── tui_notifier.go  # TUI notifications
│   │   └── buffer_adapter.go# Buffer service adapter
│   ├── controllers/         # Controller Layer
│   │   ├── controller.go          # Main agent controller
│   │   └── trigger_controller.go  # Trigger lifecycle
│   ├── views/               # View Layer
│   │   ├── views.go         # JSON API responses
│   │   └── trigger_views.go # Trigger status views
│   ├── services/            # Core Services (TODO)
│   │   ├── eventbus/        # Kafka-like event bus
│   │   ├── buffer/          # Message buffering
│   │   └── trigger/         # Agent triggering
│   └── protocol/            # Protocol Definitions
│       ├── a2a.go           # A2A protocol
│       └── security.go      # Security primitives
├── bridge/                  # Legacy (deprecated)
└── protocol/                # Protocol (legacy)
```

### Layer Responsibilities

| Layer | Package | Responsibility |
|-------|---------|----------------|
| **Model** | `internal/models` | Data structures, business logic, state |
| **Controller** | `internal/controllers` | HTTP handling, coordination |
| **View** | `internal/views` | Response formatting (JSON, TUI, Web) |
| **Service** | `internal/services` | Core business services |
| **Protocol** | `internal/protocol` | Wire formats, security |

---

## Multi-Agent Roadmap

### Phase 1: Foundation (Current) ✅

- [x] MVC structure created
- [x] Buffer-triggered agent system
- [x] TUI notifications
- [x] Tailscale discovery
- [ ] Event bus core implementation
- [ ] TUI with bubbletea

### Phase 2: Event Streaming

- [ ] Topic-based routing
- [ ] Consumer groups
- [ ] Message persistence (WAL)
- [ ] Replay capability

### Phase 3: Secure A2A

- [ ] Tailscale identity integration
- [ ] mTLS for all A2A traffic
- [ ] Ed25519 message signing
- [ ] ACL-based authorization

### Phase 4: Multi-Agent Orchestration

- [ ] Agent pool management
- [ ] Auto-scaling
- [ ] Health monitoring
- [ ] Graceful shutdown

---

## A2A Protocol

### Message Envelope

```json
{
  "header": {
    "id": "uuid",
    "type": "request|response|event",
    "source_agent": {"id": "agent-alpha", "node_id": "..."},
    "dest_agent": {"id": "agent-beta", "node_id": "..."},
    "topic": "agent.requests",
    "timestamp": "2026-03-07T12:00:00Z",
    "correlation_id": "uuid",
    "reply_to": "agent.responses"
  },
  "body": {
    "action": "execute_task",
    "payload": {},
    "metadata": {}
  },
  "security": {
    "signature": "ed25519",
    "public_key": "...",
    "timestamp": "2026-03-07T12:00:00Z"
  }
}
```

### Security Model

1. **Tailscale Identity** - Each agent has unique node identity
2. **mTLS** - Mutual TLS for all A2A communication
3. **Message Signing** - Ed25519 signatures
4. **ACL Enforcement** - Topic access control
5. **Zero Trust** - No implicit trust

See [eng_nbk/A2A_PROTOCOL.md](eng_nbk/A2A_PROTOCOL.md) for full specification.

---

## Buffer-Triggered Agents

### State Machine

```
    ┌─────────┐  0→1   ┌─────────┐  empty  ┌──────────┐
    │  IDLE   │───────▶│ ACTIVE  │────────▶│ STOPPING │
    └─────────┘        └─────────┘         └──────────┘
                                                 │
                                                 │ done
                                                 ▼
                                           ┌─────────┐
                                           │  IDLE   │
                                           └─────────┘
```

### Trigger Conditions

| Transition | Condition | Action |
|------------|-----------|--------|
| IDLE → ACTIVE | Buffer 0→1 | Start agent, notify TUI |
| ACTIVE → STOPPING | Buffer empty | Signal stop |
| STOPPING → IDLE | Agent exits | Cleanup |

### API Endpoints

```bash
GET  /trigger/status          # Get trigger status
POST /trigger/manual          # Manual trigger
POST /trigger/stop            # Stop agent
GET  /trigger/notifications   # Get notifications
POST /buffer/add              # Test: add message
POST /buffer/remove           # Test: remove message
```

---

## Design Principles

### 1. Event-Driven
- All communication via events
- Loose coupling between agents
- Async by default

### 2. Secure by Default
- Zero trust network
- Encrypt everything
- Verify identities

### 3. Kafka-Inspired
- Topic-based routing
- Consumer groups
- Message persistence
- Replay support

### 4. Tailscale-Native
- Leverage tailnet for discovery
- Use WireGuard for transport
- ACLs for authorization

### 5. MVC Architecture
- Clear separation of concerns
- Testable components
- Swappable views (TUI, Web, API)

---

## Quick Start

### Build

```bash
cd tailbridge-service
go build ./cmd/agents
```

### Initialize

```bash
go run ./cmd/agents init
```

### Run

```bash
go run ./cmd/agents
```

### Test Buffer Trigger

```bash
# Check status
curl http://localhost:8080/trigger/status

# Add message (triggers agent)
curl -X POST http://localhost:8080/buffer/add

# Remove message
curl -X POST http://localhost:8080/buffer/remove
```

---

## Configuration

Config file: `~/.agents/config.json`

```json
{
  "name": "agent-alpha",
  "auth_key": "tskey-auth-xxxxx",
  "local_agent_url": "http://127.0.0.1:9090/api",
  "inbound_port": 8001,
  "local_listen": "127.0.0.1:8080"
}
```

---

## Metrics & Monitoring

### Key Metrics

- **Throughput**: Events/second per topic
- **Latency**: P50, P95, P99 message delivery
- **Buffer Depth**: Messages waiting per agent
- **Agent Utilization**: Active vs idle time
- **Trigger Rate**: Agents started/stopped per minute

---

## Technology Stack

| Layer | Technology | Rationale |
|-------|------------|-----------|
| **Transport** | Tailscale/WireGuard | Secure, zero-config |
| **Event Bus** | Custom (Kafka-inspired) | Lightweight, embedded |
| **Persistence** | WAL | Message durability |
| **TUI** | bubbletea (Go) | Modern terminal UI |
| **API** | REST + WebSocket | Standard + real-time |
| **Identity** | Tailscale nodes | Built-in PKI |

---

## Related Documents

- [A2A Protocol Specification](eng_nbk/A2A_PROTOCOL.md)
- [Buffer Trigger Documentation](tailbridge-service/cmd/agents/BUFFER_TRIGGER.md)

---

*Last updated: March 7, 2026*
*Status: Phase 1 Complete - MVC Foundation*
