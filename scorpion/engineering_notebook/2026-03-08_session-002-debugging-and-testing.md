# Engineering Notebook — Debugging Session & Platform Bringup

**Date:** 2026-03-08
**Session:** 002
**Engineer:** Claude Code (Sonnet 4.6) + Anthony
**Topic:** Full codebase analysis, bug triage, and platform bringup — DarCI session history, Taila2a TUI rendering, frontend testing

---

## Objective

Bring up all platform components for end-to-end testing: DarCI agent (Python/Scorpion), Taila2a bridge + TUI (Go/Bubbletea), and Sentinel frontend (React). Identify and fix bugs blocking first-run operation.

---

## Codebase Analysis Summary

Full codebase audit completed. Final component map:

| Component | Language | Port | Location | Status |
|---|---|---|---|---|
| Sentinel backend | Python/FastAPI | 8000 (WS) | `backend/` | ~95% complete |
| Frontend dashboard | React 19 | 3000 | `frontend/` | ~90% complete |
| DarCI orchestrator | Python/Scorpion | — | `orchestrator/` | ~85% complete |
| Scorpion framework | Python | 18790 (Docker) | `scorpion/darci-python/` | ~90% complete |
| Taila2a bridge | Go | 8080 / 8001 | `tailbridge/taila2a/` | ~85% complete |
| TailFS | Go | — | `tailbridge/tailfs/` | ~80% complete |
| Webguide dashboard | React/TS | 3000 | `tailbridge/webguide/` | ~70% complete |
| DarCI Go | Go | — | `scorpion/darci-go/` | ~5% (stub) |

Key gaps: worker agents (openclaw/nanobot/sclaw) have no `darci_directive` inbound handlers; no end-to-end tests; DarCI Go is empty.

---

## Bugs Found and Fixed

### BUG-001 — Taila2a TUI: Tab header renders twice (visual corruption)

**Symptom:** When switching to `[2] Messages` tab, the tab bar appeared twice stacked vertically. After the first fix attempt, stats AND tabs appeared doubled across ticks.

**Root cause:** `tabStyle` had a bottom border (`NormalBorder(), false, false, true, false`). Each `Render()` returned a 2-line string (text + `────` border). Directly concatenating two 2-line strings in a `strings.Builder` stacked them vertically. After fixing with `lipgloss.JoinHorizontal`, the tab bar was correctly 2 lines tall — but this made the total view height `m.height + 1`, one line taller than the terminal. Bubbletea's frame erasing could not erase the previous frame completely, leaving the old frame visible underneath the new one.

**Fix:**
- Removed bottom border from `tabStyle` and `activeTabStyle`. Active tab now uses `Bold + Underline + color` (single-line indicator).
- Added explicit full-width `─` separator lines above and below the tab row instead.
- Added `headerLines = 5` and `footerLines = 2` constants and recalculated content height as `m.height - 7` (previously hardcoded guess of `m.height - 10`).
- Table body height: `contentHeight - 1` (subtracting 1 for the table's own header row). Viewport height: `contentHeight`.

**Files changed:** `tailbridge/taila2a/cmd/taila2a/tui/views.go`, `updates.go`

---

### BUG-002 — Taila2a TUI: Wrong API port

**Symptom:** TUI connected to `localhost:8001` (tailnet peer inbound) instead of `localhost:8080` (local outbound server). All API calls failed with `connection refused`.

**Root cause:** `runTUI()` in `main.go` used `cfg.PeerInboundPort` (8001) as the port for `tui.Run()`. The TUI API client talks to the local HTTP server which listens on `cfg.LocalListen` ("127.0.0.1:8080").

**Fix:** Parse port from `cfg.LocalListen` string (split host:port) with fallback to 8080.

**Files changed:** `tailbridge/taila2a/cmd/taila2a/main.go`

---

### BUG-003 — DarCI: Gemini 400 INVALID_ARGUMENT on every call after first

**Symptom:** First invocation of `python3 -m orchestrator` worked, but every subsequent call returned:
```
400 INVALID_ARGUMENT: Please ensure that function call turn comes immediately after a user turn or after a function response turn.
```

**Root cause (3 compounding bugs):**

1. **`build_system_prompt()` was never called.** `build_messages()` constructed the message list from history + new user message only. The system prompt (SOUL.md, identity, skills) was never injected. DarCI had no identity and no DARCI tool context in every LLM call.

2. **User messages were never saved to session history.** `_process_message` called `_save_turn(session, all_msgs, 1 + len(history))`. The user message is at index `len(history)` in `all_msgs`. Skip = `1 + len(history)` jumped past it, so only assistant/tool messages were persisted. The `_save_turn` also had `continue` for user messages with the `[Runtime Context...]` prefix, which would have skipped them even if reached.

3. **Session history started with a model turn.** Because user messages were never saved, the persisted session was `[assistant(tool_call), tool, assistant, ...]`. On the next call, `get_history()` returned this list starting with an assistant message. After `build_messages()` prepended it, Gemini received `[model(FunctionCall), ...]` as the first content turn — which is invalid. Gemini requires the first content to be a user turn, and FunctionCall turns must follow user turns.

**Fix (3 corresponding changes):**

1. **`context.py` `build_messages()`:** Prepend `{"role": "system", "content": self.build_system_prompt()}` at index 0 every call. System prompt rebuilt fresh each turn so identity/skills stay current.

2. **`loop.py` `_save_turn()`:**
   - Added `if role == "system": continue` — system messages are never stored (rebuilt fresh).
   - Changed user message handling: instead of `continue` on `[Runtime Context...]` prefix, strip the prefix with `content.partition("\n\n")` and save the clean user text. Empty results are still skipped.

3. **`loop.py` background-task path:** Changed `self._save_turn(session, all_msgs, len(history))` → `self._save_turn(session, all_msgs, 1 + len(history))` to account for the system message now at index 0.

   The `_process_message` path already used `1 + len(history)` which remains correct — with the system message at `all_msgs[0]`, this now also correctly includes the user message at `all_msgs[1 + len(history) - 1]`... actually: `all_msgs[len(history) + 1]` = new user msg, and `all_msgs[skip:]` where `skip = 1 + len(history)` = `[user_msg, asst_msgs...]`. ✓

4. **Deleted corrupted session:** `~/.scorpion/darci/workspace/sessions/cli_direct.jsonl` contained only assistant/tool turns from prior broken runs. Removed so next run starts with a clean history.

**Files changed:** `scorpion/darci-python/darci/agent/context.py`, `scorpion/darci-python/darci/agent/loop.py`

---

### BUG-004 — Taila2a TUI Messages tab: type mismatch, always empty

**Symptom:** Messages tab showed "Waiting for messages..." regardless of bridge activity.

**Root cause:** The bridge's `/trigger/notifications` endpoint returns `{"notifications": ["plain string", ...]}` (a JSON array of strings). The TUI's `FetchNotifications()` tried to decode into `[]Notification{Timestamp, Level, Message, Data}`. JSON decode fails silently (strings can't unmarshal into structs), `errMsg` was set, and the viewport was never updated.

**Fix:** Changed `FetchNotifications()` to decode each element as `json.RawMessage` then try structured first (`Notification` struct), fall back to plain string. Plain strings are promoted to `Notification{Timestamp: now, Level: "INFO", Message: s}`.

**Files changed:** `tailbridge/taila2a/cmd/taila2a/tui/api.go`

---

## Architecture Clarifications (from session 001 open questions)

### System prompt injection path (newly understood)

`AdkAgentLoop._process_message()` calls `context.build_messages()` which now:
1. Prepends `{"role": "system", "content": build_system_prompt()}` — includes identity, SOUL.md, AGENTS.md, workspace skills with `always: true` frontmatter
2. Appends prior session history (user + assistant + tool turns)
3. Appends new user message with `[Runtime Context...]` prefix (timestamp, channel, chat_id)

Session history (`~/.scorpion/darci/workspace/sessions/cli_direct.jsonl`) now correctly stores alternating `[user, assistant, tool, ..., user, assistant, ...]` turns.

### Scorpion config path

Config loads from `~/.scorpion-python/config.json` (NOT `~/.scorpion/config.json`). If absent, defaults to:
- Model: `gemini-2.5-flash`
- Temperature: 0.1
- api_key: None (falls back to `GEMINI_API_KEY` env var from `.env`)

The `.env` file at project root contains `GEMINI_API_KEY=...` and is loaded via `load_dotenv()` in `orchestrator/run.py`.

### Taila2a local vs tailnet ports

| Port | Interface | Purpose |
|---|---|---|
| 8080 | localhost only | Local outbound server: `/send`, `/agents`, `/trigger/*`, `/buffer/*` |
| 8001 | Tailscale tailnet | Peer inbound: `/inbound`, `/status` |

DarCI talks only to 8080. The TUI talks only to 8080. External peers talk only to 8001.

### DARCI species naming

- **DarCI** = Driver of AI Agent Coordination + Darcy's AI Agents (the name is intentionally a double meaning)
- **openclaw** = creative specialist (image/video/music generation)
- **zeroclaw (sclaw)** = research analyst (web research, data analysis)
- **nanobot** = production engineer (deployments, monitoring, incident response)

---

## How to Run Each Component

### DarCI Orchestrator (Python/Scorpion)
```bash
cd /path/to/sentinelai
python3 -m orchestrator
# Prompt: darci>
```
Prerequisites: `GEMINI_API_KEY` in `.env`, `darci` package importable (venv or direct).

### Taila2a Bridge + TUI (Go)
```bash
# Terminal 1 — bridge
cd tailbridge/taila2a
./taila2a run

# Terminal 2 — TUI
./taila2a tui
# [1] Agents tab, [2] Messages tab, [q] quit

# Test Messages tab (Terminal 3)
curl -X POST http://localhost:8080/buffer/add   # add 3+ to trigger agent
curl -X POST http://localhost:8080/trigger/manual
curl http://localhost:8080/trigger/notifications
```

### Sentinel Backend + Frontend (Python + React)
```bash
# Terminal 1 — backend
cd backend
GEMINI_API_KEY=... uvicorn main:app --port 8000

# Terminal 2 — frontend
cd frontend
npm start    # opens http://localhost:3000
# Enter goal + Gemini API key in UI
```

### Full Stack
| Terminal | Command |
|---|---|
| 1 | `cd backend && uvicorn main:app --port 8000` |
| 2 | `cd tailbridge/taila2a && ./taila2a run` |
| 3 | `cd sentinelai && python3 -m orchestrator` |
| 4 | `cd frontend && npm start` |
| 5 | `cd tailbridge/taila2a && ./taila2a tui` |

---

## Key Integration Points (updated)

| # | Endpoint | Direction | Used By |
|---|---|---|---|
| 1 | `GET localhost:8080/agents` | DarCI → bridge | `discover_agents` tool |
| 2 | `POST localhost:8080/send` | DarCI → bridge | `send_darci_message` tool |
| 3 | `ws://localhost:8000/ws/run` | DarCI → Sentinel | `monitor_agent` tool (asyncio.Task) |
| 4 | `ws://localhost:8000/ws/run` | Frontend → Sentinel | Live step/risk/intervention stream |
| 5 | `POST localhost:8001/inbound` | Peer → bridge | Receive `darci_directive` from DarCI |
| 6 | `GET localhost:8080/trigger/notifications` | TUI → bridge | Messages tab polling (2s interval) |
| 7 | `POST localhost:8080/buffer/add` | Test → bridge | Simulate inbound message (testing only) |

---

## Commits This Session

| Hash | Description |
|---|---|
| `43ed152` | fix: correct DarCI session history and taila2a TUI rendering |

Covers: system prompt injection, user message persistence, background-task skip offset, TUI tab border height overflow, TUI wrong API port, TUI notifications type mismatch.

---

## Open Questions (carried forward)

1. What HTTP port/path does openclaw expose for `local_agent_url`?
2. What HTTP port/path does nanobot expose?
3. Is zeroclaw (sclaw) on the Tailscale tailnet yet?
4. Do worker agents expose a `/context` endpoint, or do we use `darci_status_request` messages for status polling?
5. The `/trigger/notifications` endpoint returns plain strings; should the bridge be updated to return structured `{timestamp, level, message}` objects instead?
6. The DarCI `404 Not Found` on `/agents` during the first run (session 001) was caused by taila2a not running — this is expected. The fix is to always start `./taila2a run` before `python3 -m orchestrator`.

---

## Pending Work

- [ ] Worker agent adapters: openclaw/nanobot/sclaw need `darci_directive` inbound handlers
- [ ] End-to-end test: DarCI assigns task → taila2a routes → worker executes → Sentinel monitors → DarCI receives completion
- [ ] DarCI Go stub: currently ~5% complete, no implementation
- [ ] Taila2a `/trigger/notifications` response: upgrade from `[]string` to `[]Notification` struct on the bridge side
- [ ] Frontend: no agent list view or file transfer UI yet
- [ ] Extended Taila2a protocol: topics, partitions, Ed25519 signing planned but unimplemented

---

*Next session: End-to-end integration test — DarCI → Taila2a → worker agent round trip.*
