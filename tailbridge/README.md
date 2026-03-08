# Tailbridge Project

**Secure peer-to-peer communication over Tailscale**

---

## Repository Structure

```
tailbridge/
├── eng_nbk/           # Engineering notebook & documentation
│   ├── README.md      # Main eng_nbk index
│   ├── AGENT_TASKS.md # Agent task delegation (START HERE)
│   ├── A2A_PROTOCOL.md# Protocol specification
│   ├── next_steps/    # Detailed task specs
│   └── workflows/     # CI/CD pipelines
├── taila2a/           # Agent-to-Agent communication protocol
│   ├── cmd/taila2a/   # Main entry point
│   ├── internal/      # Internal packages
│   └── README.md      # Taila2a docs
├── tailfs/            # Secure file transfer (Tail Agent File Send)
│   ├── cmd/tailfs/    # Main entry point
│   ├── internal/      # Internal packages
│   └── README.md      # TailFS docs
└── webguide/          # Web UI components (TODO)
    └── README.md      # Webguide planning
```

---

## Quick Start

### Taila2a - Agent-to-Agent Communication

```bash
cd taila2a
go run ./cmd/taila2a init    # First-time setup
go run ./cmd/taila2a         # Start service
```

**Features:**
- Phone book agent discovery
- Topic-based messaging
- Buffer-triggered agents
- TUI notifications

### TailFS - Secure File Transfer

```bash
cd tailfs
go run ./cmd/tailfs init     # First-time setup
go run ./cmd/tailfs          # Start service
go run ./cmd/tailfs send file.pdf tailfs-beta  # Send file
```

**Features:**
- Chunked file transfers
- Progress tracking
- Resume support
- End-to-end encryption

---

## Documentation

| Document | Location |
|----------|----------|
| Engineering Notebook | [eng_nbk/README.md](eng_nbk/README.md) |
| A2A Protocol | [eng_nbk/A2A_PROTOCOL.md](eng_nbk/A2A_PROTOCOL.md) |
| Next Steps/Tasks | [eng_nbk/next_steps/README.md](eng_nbk/next_steps/README.md) |
| Agent Task Delegation | [eng_nbk/AGENT_TASKS.md](eng_nbk/AGENT_TASKS.md) |
| Taila2a README | [taila2a/README.md](taila2a/README.md) |
| TailFS README | [tailfs/README.md](tailfs/README.md) |

---

## Development

### Build All

```bash
# Taila2a
cd taila2a && go build ./cmd/taila2a

# TailFS
cd tailfs && go build ./cmd/tailfs
```

### Run Tests

```bash
# Taila2a
cd taila2a && go test ./...

# TailFS
cd tailfs && go test ./...
```

---

## Agent Task Delegation

This project is designed for multi-agent development. See [eng_nbk/AGENT_TASKS.md](eng_nbk/AGENT_TASKS.md) for task assignments.

### Current Sprint: Phase 1 - Core Functionality

| Agent | Task | Status |
|-------|------|--------|
| Agent 1 | Event Bus Core | Pending |
| Agent 2 | Chunk Transfer Protocol | Pending |
| Agent 3 | Phone Book Sync | Pending |
| Agent 4 | Unit Tests | Pending |

---

## Technology Stack

| Component | Technology |
|-----------|------------|
| Transport | Tailscale/WireGuard |
| Language | Go 1.25+ |
| Event Bus | Custom (Kafka-inspired) |
| TUI | Bubbletea |
| Testing | Go testing + testify |

---

## License

Same as the parent project.

---

*Last updated: March 7, 2026*
*Status: Phase 1 Complete - Foundation Ready*
