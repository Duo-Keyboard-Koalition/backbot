# Engineering Notebook — DarCI Architecture Session

**Date:** 2026-03-07
**Session:** 001
**Engineer:** Claude Code (Sonnet 4.6) + Anthony
**Topic:** Full codebase analysis + DarCI integration architecture design

---

## Objective

Analyze existing codebase (backend/, scorpion/, tailbridge/) to determine how to build a DarCI-based AI agent project manager. Map every component, trace every protocol, design the integration architecture, and offload all context into persistent logs.

---

## Components Analyzed

### 1. backend/ — Sentinel Risk Engine

FastAPI + WebSocket at `/ws/run`. Streams agent step events with risk scores and interventions. This becomes the **Approver** in DARCI.

**Key files:**
- `main.py` — WebSocket handler + orchestration
- `agent.py` — Gemini-powered agent loop with 4 tools (web_search, write_to_file, read_file, calculate)
- `sentinel.py` — Risk scoring with weighted composite formula
- `intervention.py` — 4 intervention strategies
- `state.py` — Data models (Step, Intervention, ExecutionState, FailureType, InterventionType)

**Risk scoring formula:**
```
risk = 0.35×LOOP + 0.30×GOAL_DRIFT + 0.20×LOW_CONFIDENCE + 0.15×INCOHERENT_TOOL
```
Intervention fires at: `risk_score >= 0.5` OR `len(failure_types) >= 2`

**Event stream protocol (WebSocket):**
```json
// Client sends:
{"goal": "task description", "api_key": "...", "max_steps": 50}

// Server streams:
{"type": "start", "goal": "..."}
{"type": "step", "step": {...}, "risk_score": 0.42, "failure_types": ["loop"]}
{"type": "intervention", "intervention": {"intervention_type": "REPROMPT", "message": "..."}}
{"type": "complete", "state": {...}}
{"type": "timeout"}
```

---

### 2. scorpion/scorpion-python/ — Agent Framework (= DarCI Runtime)

Ultra-lightweight (3,935 LOC) Python AI agent framework. **Scorpion IS DarCI** when configured with DARCI tools and system prompt.

**Architecture:**
```
AdkAgentLoop (perceive → think → act)
├── MessageBus (InboundMessage / OutboundMessage pub/sub)
├── ToolRegistry (JSON-schema based, registered at startup)
├── MemorySystem (short-term buffer / working / long-term / semantic)
├── SkillsLoader (workspace/skills/<name>/SKILL.md frontmatter)
├── SubagentManager (asyncio.Task lifecycle, bus.publish_inbound reports)
└── ChannelManager (9 platform integrations)
```

**Critical Tool base class pattern:**
```python
class MyTool(Tool):
    @property
    def name(self) -> str: return "my_tool"
    @property
    def description(self) -> str: return "what it does"
    @property
    def parameters(self) -> dict:
        return {
            "type": "object",
            "properties": {"arg": {"type": "string", "description": "..."}},
            "required": ["arg"]
        }
    async def execute(self, arg: str, **kwargs) -> str:
        return "result"
```

**Config structure (~/.scorpion/config.json):**
```json
{
  "providers": {"gemini": {"apiKey": "...", "model": "gemini-2.5-flash"}},
  "agents": {"defaults": {"temperature": 0.7, "max_tokens": 8192, "memory_window": 50}},
  "channels": {"telegram": {"enabled": false}, "cli": {"enabled": true}},
  "tools": {"restrictToWorkspace": false}
}
```

**Supported channels:** Telegram, Discord, Slack, WhatsApp, CLI, Email, Feishu, DingTalk, Matrix (+ custom)

---

### 3. tailbridge/ — P2P Communication Layer

Go-based bridge over Tailscale. Two subsystems connecting all agents.

**Taila2a (messaging):**
```
Local side (agent-facing, port 8080):
  POST /send — route message to peer: {dest_node, payload}
  GET /agents — phone book: list of all tailnet peers

Tailnet side (peer-facing, port 8001):
  POST /inbound — receive from peer, forward to local_agent_url
  GET /status — bridge health

Internal services:
  EventBus — Kafka-inspired, topic/partition/consumer-group routing
  MessageBuffer — persistent JSON store, exponential backoff retry (max 3)
  Discovery — polls Tailscale local client every 30s, TCP port scans
  AgentTrigger — idle→active state machine on buffer depth change
```

**Envelope format:**
```json
{
  "source_node": "darci-node",
  "dest_node": "openclaw-node",
  "payload": {"type": "darci_directive", "task_id": "T001"},
  "timestamp": "2026-03-07T14:22:00Z"
}
```

**Extended A2A protocol (in progress):**
- Full envelope: header + body + security (Ed25519 signature)
- Topics: `agent.requests`, `agent.responses`, `darci.context`, `system.health`
- Consumer groups with offset tracking (Kafka semantics)
- Replay attack protection (±5 minute timestamp window)

**TailFS (file transfer):**
- 1MB chunks, resume support, optional compression + encryption
- Progress tracking: bytes/s, ETA, status (pending/sending/receiving/complete/failed)
- Config: ChunkSize=1MB, MaxConcurrentTransfers=3, TransferTimeout=30min

---

### 4. scorpion/darci-brainstorm/ — Design Documentation

Extensive design docs already exist (19 strategy docs). None are implemented yet.

**Key docs:**
- `DARCI_PROJECT_MANAGEMENT.md` — vision, architecture, workflow patterns
- `AGENT_ARCHITECTURE.md` — component diagram, sub-agent roles, memory schema
- `IMPLEMENTATION_CHECKLIST.md` — 11 phases, all unchecked
- `TOOL_SPECIFICATION.md` — 117+ tool definitions

**Existing task schema (reuse verbatim):**
```json
{
  "id": "T001",
  "title": "task name",
  "description": "details",
  "priority": "P0",
  "status": "pending",
  "dependencies": [],
  "labels": [],
  "created_at": "ISO datetime",
  "updated_at": "ISO datetime",
  "artifacts": {}
}
```

**Sub-agent roles defined:**
- Project Manager: task creation, prioritization, dependencies
- Build Engineer: build monitoring, test execution, error fixing
- Code Analyst: feature comparison, code quality
- Scribe: engineering notebooks, status reports, feature matrices

---

## Architecture Decisions (ADRs)

### ADR-001 — Scorpion IS DarCI

**Status:** ACCEPTED

**Decision:** DarCI is not a separate process. Scorpion is configured with DARCI-specific tools, system prompt, and skills. Running `scorpion gateway` with the darci skill loaded = running DarCI. No new agent loop class is created.

**Rationale:** AdkAgentLoop already provides tool execution, multi-layer memory, skill plugins, sub-agent spawning, and multi-channel delivery. Creating a separate agent loop would duplicate this infrastructure for no benefit.

**Consequences:** The `darci/` package contains only tools, models, state store, system prompt, and skill file. It is imported by Scorpion, not the other way around.

---

### ADR-002 — DARCI Role Mapping

**Status:** ACCEPTED

**Decision:**

| DARCI Role | Entity | Mechanism |
|---|---|---|
| Driver | Scorpion/DarCI | Owns tasks.json, sets goals, dispatches via taila2a |
| Approver | Sentinel backend | risk_score >= 0.5 or HALT = veto signal |
| Responsible | openclaw / nanobot / sclaw | Receive darci_directive via taila2a inbound |
| Consulted | Scorpion subagents | Spawned for analysis before task assignment |
| Informed | Engineering notebook | Written on every state transition |

**Rationale:** Sentinel's four intervention types map directly to Approver veto behaviors (REPROMPT=advisory, ROLLBACK=partial veto, HALT=full veto). DARCI gives this existing system a formal project management vocabulary.

---

### ADR-003 — Sentinel Monitoring as Persistent asyncio.Task

**Status:** ACCEPTED

**Decision:** Each agent's sentinel stream is monitored by a persistent `asyncio.Task` (not a Scorpion subagent). Reports back via `bus.publish_inbound(channel="system")`.

**Rationale:** Scorpion subagents are designed for one-shot LLM tasks. Sentinel monitoring is a long-lived WebSocket event stream. Using asyncio.Task directly mirrors the pattern in Scorpion's `SubagentManager` and `tailbridge/taila2a/internal/services/` without the LLM overhead.

---

### ADR-004 — Transport Always via Taila2a Local HTTP API

**Status:** ACCEPTED

**Decision:** DarCI calls only `localhost:8080` (local taila2a bridge). It never talks to peer Tailscale IPs directly.

**Rationale:** taila2a handles NAT traversal, message buffering (persistent JSON store), retry (exponential backoff), and Tailscale authentication. DarCI stays above the transport layer.

---

## Integration Points

| # | API | Used For |
|---|---|---|
| 1 | GET http://localhost:8080/agents | Agent discovery (phone book) |
| 2 | POST http://localhost:8080/send | Send darci_directive / status_request |
| 3 | ws://localhost:8000/ws/run | Monitor agent risk in real time |
| 4 | scorpion.agent.tools.base.Tool | Base class for all DarCI tools |
| 5 | scorpion.agent.loop.AdkAgentLoop | Core agent runtime |
| 6 | bus.publish_inbound(channel="system") | Report asyncio task results to agent |

---

## Proposed darci/ Package Structure

```
darci/
├── __init__.py
├── agent.py              # Wire AdkAgentLoop with DarCI tools + startup coroutine
├── config.py             # DarciConfig dataclass
├── run.py                # python -m darci entry point
├── system_prompt.py      # DarCI soul: DARCI roles, tool API, examples
├── models/
│   ├── task.py           # Task, DarciRoles, SentinelSnapshot dataclasses
│   └── agent_context.py  # AgentContextSnapshot dataclass
├── state/
│   └── store.py          # TaskStore: JSON I/O for ~/.scorpion/darci/
├── tools/
│   ├── task_store.py     # TaskCreateTool, TaskUpdateTool, TaskQueryTool, StatusReportTool
│   ├── taila2a_client.py # DiscoverAgentsTool, SendDarciMessageTool
│   ├── sentinel_client.py # MonitorAgentTool + SentinelMonitorRegistry
│   ├── notebook_writer.py # NotebookCreateTool, NotebookAppendTool
│   └── darci_tools.py    # register_darci_tools(registry)
└── skills/darci/SKILL.md # Workspace skill for SkillsLoader
```

---

## Implementation Roadmap

| Phase | Goal | Key Files |
|---|---|---|
| 1 | Skeleton + task store | models/task.py, state/store.py, tools/task_store.py |
| 2 | Taila2a discovery + send | tools/taila2a_client.py, config.py |
| 3 | Sentinel WS monitoring | tools/sentinel_client.py |
| 4 | Notebook auto-generation | tools/notebook_writer.py, skills/darci/SKILL.md |
| 5 | Full coordination loop | agent.py startup coroutine |
| 6 | Status dashboard | StatusReportTool |

---

## Open Questions

1. What HTTP port/path does openclaw expose? (needed for taila2a local_agent_url)
2. What HTTP port/path does nanobot expose?
3. Is sclaw (zero claw / Rust) already on the Tailscale tailnet?
4. Do worker agents expose a `/context` endpoint, or do we use darci_status_request messages?
5. Which Gemini model for DarCI? (gemini-2.5-flash per brainstorm docs, or gemini-2.0-flash like backend?)

---

## Velocity & Metrics

| Metric | Value |
|---|---|
| Components fully analyzed | 4 |
| Files read | ~40+ |
| ADRs recorded | 4 |
| Tasks scoped (impl) | 7 (P0×3, P1×2, P2×1, P3×1) |
| Open questions | 5 |
| Blockers | 0 |

---

*Next session: Begin Phase 1 implementation — darci/ package skeleton and TaskStore.*
