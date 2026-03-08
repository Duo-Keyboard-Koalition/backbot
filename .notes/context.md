# SentinelAI — Engineering Context Log

---

## Entry · 2026-03-07 · Session 001
**Author:** Claude Code (Sonnet 4.6)
**Summary:** Initial deep-dive codebase analysis. Mapped all four components and designed DarCI integration architecture.

### What Was Analyzed

Performed full read of backend/, scorpion/, tailbridge/, and darci-brainstorm/. Traced every major file and protocol.

### Key Discoveries

**SentinelAI Backend (backend/)**
- FastAPI + WebSocket at /ws/run
- Agent loop: Gemini model, tools: web_search, write_to_file, read_file, calculate
- Risk scoring engine (sentinel.py) with 4 failure types:
  - LOOP (35%) — repeated action patterns
  - GOAL_DRIFT (30%) — semantic deviation from original goal
  - LOW_CONFIDENCE (20%) — hedging/uncertainty language
  - INCOHERENT_TOOL (15%) — invalid/error-prone tool usage
- Intervention types: REPROMPT, ROLLBACK, DECOMPOSE, HALT
- Threshold: risk_score >= 0.5 triggers intervention
- Model: gemini-2.0-flash (env: GEMINI_FLASH_MODEL)

**Scorpion Agent Framework (scorpion/scorpion-python/)**
- Core class: AdkAgentLoop (perceive → think → act)
- Event-driven: MessageBus pub/sub with InboundMessage/OutboundMessage
- Tools: JSON-schema based, inherit from Tool base class
  - async execute(**kwargs) -> str signature required
  - name, description, parameters properties required
- Memory: short-term (50 msg buffer), working (session), long-term (archived)
- Skills: loaded from workspace/skills/<name>/SKILL.md (frontmatter-driven)
- Sub-agents: SubagentManager spawns asyncio tasks, reports via bus.publish_inbound(channel="system")
- Channels: Telegram, Discord, Slack, CLI, Email, WhatsApp, Feishu, DingTalk, Matrix
- Config: ~/.scorpion/config.json (providers.gemini.apiKey, agents.defaults, channels.*)
- Entry points: `scorpion agent` (CLI), `scorpion gateway` (persistent multi-channel)
- LLM: Google Gemini primary (gemini-2.5-flash), extensible to OpenAI, Anthropic, local

**Tailbridge (tailbridge/)**
- Two subsystems: taila2a (messaging) + tailfs (file transfer)
- Taila2a:
  - Language: Go 1.25+
  - Transport: Tailscale/WireGuard (tsnet embedded — no tailscaled daemon needed)
  - Local API (port 8080): POST /send, GET /agents
  - Tailnet API (port 8001): POST /inbound, GET /agents, GET /status
  - Envelope: {source_node, dest_node, payload: json.RawMessage, timestamp}
  - Extended A2A protocol: topics, partitions, consumer groups, Ed25519 signing (planned)
  - Phone book: polls Tailscale local client every 30s + TCP port scans for services
  - Message buffer: persistent JSON store, retry with exponential backoff (max 3)
  - Agent trigger state machine: idle → active when buffer depth 0 → 1+
- TailFS:
  - Chunked file transfer (1MB chunks default)
  - Progress tracking (bytes/s, ETA, status: pending/sending/receiving/complete/failed)
  - Resume support on failure
  - Optional compression + encryption per transfer

**Agent Ecosystem**
- Scorpion = meta-manager (DarCI runs on/as Scorpion — it IS the PM agent)
- openclaw = "toy soldier" worker agent (Python)
- nanobot = lightweight worker agent (Python/Go)
- zero claw / sclaw = performance-critical worker agent (Rust)
- All communicate via tailbridge (taila2a) over Tailscale tailnet

**DarCI Brainstorm Docs (scorpion/darci-brainstorm/)**
- DARCI = Driver, Approver, Responsible, Consulted, Informed
- Extensive design docs already exist but are NOT yet implemented
- Task schema: {id, title, priority P0-P3, status, dependencies, labels, artifacts}
- State dir: ~/.scorpion/darci/{tasks.json, context.json, notebooks/, metrics/}
- 117+ tools brainstormed in TOOL_SPECIFICATION.md
- 11-phase implementation checklist defined, all unchecked

### Architecture Decisions (ADRs)

**ADR-001: Scorpion IS DarCI**
DarCI is not a separate process. Scorpion is configured with DARCI-specific tools,
system prompt, and skills. No new agent loop is needed — AdkAgentLoop handles everything.

**ADR-002: DARCI Role Mapping**
- Driver = Scorpion/DarCI (sets goals, owns tasks.json, dispatches work)
- Approver = Sentinel backend (risk_score >= 0.5 or HALT intervention = veto)
- Responsible = openclaw / nanobot / zero claw / sclaw (worker agents on tailnet)
- Consulted = Short-lived Scorpion subagents (spawned for analysis)
- Informed = Engineering notebook + metrics store

**ADR-003: Sentinel Monitoring as asyncio.Task**
Each agent's sentinel WS stream is a persistent asyncio.Task, not a Scorpion subagent.
Subagents use the full LLM stack — overkill for a streaming event listener.
Reports state changes back via bus.publish_inbound(channel="system").

**ADR-004: Transport Always via Taila2a HTTP API**
DarCI talks only to localhost:8080 (local taila2a bridge). Never to peer IPs directly.
taila2a handles NAT traversal, message buffering, retry, and auth.

### Key Integration Points

1. GET http://localhost:8080/agents — taila2a phone book (agent discovery)
2. POST http://localhost:8080/send — taila2a outbound to peer
3. ws://localhost:8000/ws/run — sentinel WebSocket stream
4. scorpion.agent.tools.base.Tool — base class all DarCI tools must subclass
5. scorpion.agent.loop.AdkAgentLoop — agent core
6. scorpion.agent.subagent.SubagentManager — async task lifecycle pattern to follow

### Taila2a Message Topics

Embedded in payload.type field of the envelope:
- darci_directive — task assignment or goal correction to a worker agent
- darci_status_request — DarCI requesting current goal/context from a worker
- darci_status_response — worker replying with its ExecutionState snapshot

### Open Questions

- What HTTP port does each worker agent expose for taila2a local_agent_url?
- Is zero claw / sclaw already on the Tailscale tailnet?
- Which Gemini model should DarCI use? (brainstorm docs say gemini-2.5-flash)

### Next Steps (Prioritized)

- P0: darci/tools/task_store.py — task management foundation
- P0: darci/tools/taila2a_client.py — agent discovery
- P0: darci/agent.py + darci/system_prompt.py — wire Scorpion as DarCI
- P1: darci/tools/sentinel_client.py — WS risk monitor
- P1: darci/tools/notebook_writer.py — auto notebook generation
- P2: Full coordination loop (discover → assign → monitor → intervene)
- P3: Worker agent adapters for darci_directive inbound messages

---

## Quick Reference

### Component Entry Points

| Component | Start Command | Port/Endpoint |
|---|---|---|
| Sentinel backend | `uvicorn backend.main:app --port 8000` | ws://localhost:8000/ws/run |
| Scorpion/DarCI | `scorpion gateway` | http://localhost:18790 |
| Taila2a bridge | `./tailbridge/taila2a/taila2a` | 8080 local, 8001 tailnet |
| TailFS | `./tailbridge/tailfs/tailfs` | http://localhost:8081 |
| Frontend | `cd frontend && npm start` | http://localhost:3000 |

### File Naming Conventions

- Engineering notebooks: scorpion/engineering_notebook/YYYY-MM-DD_{topic}.md
- DarCI state: ~/.scorpion/darci/{tasks.json, context.json}
- Taila2a config: ~/.taila2a/config.json
- Scorpion config: ~/.scorpion/config.json

### Environment Variables

| Variable | Component | Notes |
|---|---|---|
| GEMINI_API_KEY | backend, scorpion | Primary LLM key |
| GEMINI_FLASH_MODEL | backend | e.g. gemini-2.0-flash |
| GEMINI_PRO_MODEL | backend | e.g. gemini-3-1-pro |
