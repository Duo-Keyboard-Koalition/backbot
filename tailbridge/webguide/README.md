# Webguide - Web UI for Tailbridge

**Status:** Planned (TODO)

---

## Purpose

Webguide provides a web-based user interface for monitoring and managing the Tailbridge system (Taila2a + TailFS).

---

## Planned Features

### Dashboard
- System overview with key metrics
- Agent phone book with status
- Active transfers monitoring
- Topic/consumer group visualization

### Management
- Agent control (start/stop/restart)
- Topic management
- Transfer initiation
- Configuration management

### Monitoring
- Real-time metrics
- Transfer progress
- Agent health
- Alerting

---

## Technology Stack (Planned)

| Component | Technology |
|-----------|------------|
| Frontend | React + TypeScript |
| UI Library | Material-UI / Chakra UI |
| State | Zustand / Redux |
| API | REST + WebSocket |
| Charts | Recharts / Chart.js |

---

## Directory Structure (Planned)

```
webguide/
├── src/
│   ├── components/
│   ├── pages/
│   ├── hooks/
│   ├── api/
│   └── types/
├── public/
├── package.json
└── README.md
```

---

## API Integration

### Taila2a API
- `GET /phonebook` - Agent list
- `GET /agents/online` - Online agents
- `GET /trigger/status` - Trigger status

### TailFS API
- `GET /transfers` - Active transfers
- `GET /history` - Transfer history
- `POST /send` - Initiate transfer

---

## Status

**Phase:** Planning

**Next Steps:**
1. Design UI mockups
2. Set up React project
3. Implement API client
4. Build dashboard components

---

*Last updated: March 7, 2026*
