# Engineering Notebook (eng_nbk)

## Multi-Agent Orchestration Documentation

This directory contains the engineering documentation for the Tailbridge multi-agent orchestration system.

---

## Documents

| Document | Description |
|----------|-------------|
| [A2A_PROTOCOL.md](A2A_PROTOCOL.md) | Secure Agent-to-Agent protocol specification |
| [eng_nbk.md](../eng_nbk.md) | Main engineering notebook (root) |

---

## Quick Links

- **Main Notebook**: `../eng_nbk.md` - Multi-agent orchestration roadmap
- **MVC Structure**: `../internal/` - Models, Controllers, Views
- **Entry Point**: `../cmd/agents/` - Agents service

---

## Architecture Summary

```
tailbridge-service/
├── cmd/
│   └── agents/              # Main entry point
├── internal/
│   ├── models/              # Data structures + business logic
│   ├── controllers/         # Request handling
│   ├── views/               # Presentation layer (TUI, API, Webhook)
│   ├── services/            # Core services (eventbus, buffer, trigger)
│   └── protocol/            # A2A protocol definitions
├── bridge/                  # Legacy bridge (deprecated)
└── protocol/                # Protocol definitions (legacy)
```

---

*Last updated: March 7, 2026*
