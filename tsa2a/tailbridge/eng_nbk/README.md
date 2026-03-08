# Engineering Notebook (eng_nbk)

## Tailbridge Multi-Agent System

**Two complementary systems for secure peer-to-peer communication over Tailscale:**

1. **Taila2a** - Agent-to-Agent communication with phone book discovery
2. **TailFS** - Secure file transfer between agents

---

## Quick Navigation

| Document | Description |
|----------|-------------|
| [AGENT_TASKS.md](AGENT_TASKS.md) | **START HERE** - Agent task delegation |
| [next_steps/](next_steps/README.md) | Detailed task specifications |
| [A2A_PROTOCOL.md](A2A_PROTOCOL.md) | Protocol specification |
| [workflows/](workflows/) | CI/CD pipelines |

---

## Agent Task Delegation

**4 Agents are working on this project:**

| Agent | Focus Area | Key Tasks |
|-------|-----------|-----------|
| **Agent 1** | Event Bus & Messaging | Event bus, consumer groups, WAL |
| **Agent 2** | File Transfer Protocol | Chunking, resume, compression, encryption |
| **Agent 3** | Integration & Discovery | Phone book sync, CLI, metrics, ACLs |
| **Agent 4** | Testing & Quality | Unit tests, integration tests, E2E |

See [AGENT_TASKS.md](AGENT_TASKS.md) for detailed assignments.

---

## System Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                      Tailbridge Ecosystem                                │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌─────────────────────┐         ┌─────────────────────┐               │
│  │      Taila2a        │         │       TailFS        │               │
│  │  (Agent-to-Agent)   │         │   (File Transfer)   │               │
│  │                     │         │                     │               │
│  │  • Phone Book       │         │  • Chunked Transfer │               │
│  │  • Agent Discovery  │◀───────▶│  • Progress Track   │               │
│  │  • A2A Messaging    │  Shared │  • Compression      │               │
│  │  • Command Exec     │ Tailscale│  • Encryption      │               │
│  │  • Buffer Trigger   │  Tailnet│  • Transfer History │               │
│  └─────────────────────┘         └─────────────────────┘               │
│                                                                          │
│                    ┌─────────────────────────┐                          │
│                    │    Tailscale Tailnet    │                          │
│                    │    (WireGuard Encrypted)│                          │
│                    └─────────────────────────┘                          │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Repository Structure

```
tailbridge/
├── tailbridge-service/          # Taila2a implementation
│   ├── cmd/taila2a/             # Main entry point
│   ├── internal/
│   │   ├── models/              # Phone book, agents, config
│   │   ├── controllers/         # HTTP handlers
│   │   └── views/               # Response formatting
│   └── eng_nbk/                 # Engineering docs
│
├── tail-agent-file-send/        # TailFS implementation
│   ├── cmd/tailfs/              # Main entry point
│   ├── internal/
│   │   ├── models/              # File transfer models
│   │   ├── services/            # Transfer service
│   │   └── controllers/         # HTTP handlers
│   └── docs/                    # Documentation
│
├── .github/workflows/           # CI/CD pipelines
│   ├── taila2a-ci.yml
│   └── tailfs-ci.yml
│
└── eng_nbk/                     # This directory
    ├── README.md
    └── A2A_PROTOCOL.md
```

---

## Taila2a - Agent-to-Agent Communication

### Purpose
Secure messaging, command execution, and agent discovery on your Tailscale tailnet.

### Key Components

#### Phone Book (Agent Discovery)
```go
// Automatic discovery of all agents on tailnet
phonebook := models.NewPhoneBook(srv)
phonebook.Start(nil)

// Search for agents
agents := phonebook.SearchAgents("alpha")
fileAgents := phonebook.GetAgentsByCapability(CapabilityFileSend)
```

#### Agent Capabilities
| Capability | Description |
|------------|-------------|
| `file_send` | Can send files |
| `file_receive` | Can receive files |
| `chat` | Can send/receive messages |
| `command` | Can execute commands |
| `stream` | Can stream data |

#### Buffer-Triggered Agents
- Agents auto-start when buffer receives messages (0→1)
- Agents auto-stop when buffer empties
- TUI notifications for state changes

### API Endpoints

```bash
# Get full phone book
GET http://localhost:8080/phonebook

# Get online agents
GET http://localhost:8080/agents/online

# Search agents
GET http://localhost:8080/agents/search?q=alpha

# Filter by capability
GET http://localhost:8080/agents/capability?cap=file_send

# Get agent details
GET http://localhost:8080/agents/detail?name=taila2a-alpha

# Trigger status
GET http://localhost:8080/trigger/status
```

### Quick Start
```bash
cd tailbridge-service
go run ./cmd/taila2a init
go run ./cmd/taila2a run
```

---

## TailFS - Secure File Transfer

### Purpose
Peer-to-peer file transfer between computers on your tailnet.

### Key Components

#### File Transfer Service
```go
// Initialize transfer service
config := models.DefaultFileTransferConfig()
svc, _ := services.NewFileTransferService(config)

// Send file
req := &FileTransferRequest{
    ID: "transfer-123",
    FilePath: "/path/to/file.pdf",
    DestAgentName: "tailfs-beta",
}
transferID, _ := svc.SendFile(ctx, req)

// Get progress
progress, _ := svc.GetProgress(transferID)
```

#### Transfer Features
- **Chunked Transfer** - Large files split into 1MB chunks
- **Progress Tracking** - Real-time bytes sent, ETA, speed
- **Compression** - Optional gzip compression
- **Encryption** - End-to-end via WireGuard
- **Resume Support** - Failed transfers can resume

### API Endpoints

```bash
# Send file
POST http://localhost:8081/send
{
  "file": "/path/to/file.pdf",
  "destination": "tailfs-beta",
  "compress": true
}

# Get progress
GET http://localhost:8081/progress?transfer_id=uuid

# Get history
GET http://localhost:8081/history

# List agents
GET http://localhost:8081/agents
```

### Quick Start
```bash
cd tail-agent-file-send
go run ./cmd/tailfs init
go run ./cmd/tailfs run
go run ./cmd/tailfs send file.pdf tailfs-beta
```

---

## Combined Workflows

### File Transfer with Discovery

```
1. Query Taila2a phone book for file_receive agents
   GET /agents/capability?cap=file_receive

2. Select destination from results

3. Initiate TailFS transfer
   POST /send {file, destination}

4. Monitor progress
   GET /progress?transfer_id=xxx

5. Receive completion notification
```

### Agent Communication Flow

```
┌──────────┐     ┌──────────┐     ┌──────────┐
│  Sender  │────▶│ Phone    │────▶│ Receiver │
│  Agent   │     │ Book     │     │  Agent   │
└──────────┘     └──────────┘     └──────────┘
     │                 │                 │
     │  1. Discover    │                 │
     │────────────────▶│                 │
     │                 │  2. Lookup      │
     │                 │────────────────▶│
     │                 │                 │
     │  3. Connect Direct (Tailscale)    │
     │──────────────────────────────────▶│
     │                 │                 │
     │  4. Transfer Data                 │
     │◀──────────────────────────────────│
     │                 │                 │
```

---

## CI/CD Workflows

### Taila2a Pipeline
```yaml
push/PR → Build → Test → Lint → Vet → Security → Release
```

### TailFS Pipeline
```yaml
push/PR → Build → Test → Lint → Vet → Security → Integration → Release
```

---

## Design Principles

### 1. Zero Trust Security
- All agents authenticated via Tailscale
- WireGuard encryption for all traffic
- Optional additional encryption layer

### 2. Discovery First
- Automatic agent discovery
- Capability detection via port scanning
- Real-time status updates

### 3. Chunked Transfers
- Large files split into manageable chunks
- Progress tracking per chunk
- Resume from failure point

### 4. MVC Architecture
- Models: Data structures + business logic
- Controllers: HTTP request handling
- Views: Response formatting

### 5. Event-Driven
- Buffer-triggered agent activation
- Progress notifications
- Completion callbacks

---

## Configuration

### Taila2a (`~/.taila2a/config.json`)
```json
{
  "name": "taila2a-alpha",
  "auth_key": "tskey-auth-xxxxx",
  "local_agent_url": "http://127.0.0.1:9090/api",
  "inbound_port": 8001,
  "local_listen": "127.0.0.1:8080"
}
```

### TailFS (`~/.tailfs/config.json`)
```json
{
  "node_name": "tailfs-alpha",
  "auth_key": "tskey-auth-xxxxx",
  "local_listen": "127.0.0.1:8081",
  "download_dir": "~/Downloads/tailfs"
}
```

---

## Development Roadmap

### Phase 1: Foundation ✅
- [x] MVC structure for both systems
- [x] Phone book agent discovery
- [x] File transfer service
- [x] CI/CD pipelines

### Phase 2: Integration
- [ ] Taila2a + TailFS unified CLI
- [ ] Cross-system notifications
- [ ] Shared agent state

### Phase 3: Advanced Features
- [ ] Transfer scheduling
- [ ] Bandwidth throttling
- [ ] Multi-recipient transfers
- [ ] Web UI dashboard

### Phase 4: Production
- [ ] Prometheus metrics
- [ ] Structured logging
- [ ] Kubernetes deployment
- [ ] Helm charts

---

## Documentation

| Document | Location |
|----------|----------|
| Main README | [README.md](../README.md) |
| Taila2a README | [tailbridge-service/cmd/taila2a/README.md](../tailbridge-service/cmd/taila2a/README.md) |
| TailFS README | [tail-agent-file-send/README.md](../tail-agent-file-send/README.md) |
| A2A Protocol | [A2A_PROTOCOL.md](A2A_PROTOCOL.md) |

---

*Last updated: March 7, 2026*
*Status: Phase 1 Complete - Foundation Ready*
