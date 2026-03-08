# SentinelAI Build Guide

**Last updated:** 2026-03-07

---

## Prerequisites

- Python 3.11+
- Go 1.21+
- Node 18+ (frontend)
- Tailscale account + auth key (tailbridge)
- Gemini API key (Google AI Studio)

---

## Component 1: Sentinel Backend

Risk monitoring WebSocket server — the Approver in DARCI.

```bash
# From project root
cp .env.example .env
# Edit .env: set GEMINI_API_KEY=...

pip install -r requirements.txt
uvicorn backend.main:app --port 8000 --reload
```

**WebSocket endpoint:** `ws://localhost:8000/ws/run`

**Protocol:**
- Send on connect: `{"goal": "your task", "api_key": "...", "max_steps": 50}`
- Receive stream:
  - `{"type": "step", "risk_score": 0.42, "failure_types": ["loop"]}`
  - `{"type": "intervention", "intervention": {"intervention_type": "REPROMPT"}}`
  - `{"type": "complete", "state": {...}}`

---

## Component 2: Scorpion / DarCI (Meta-Manager)

Scorpion is the agent framework. DarCI = Scorpion + DARCI tools + DARCI system prompt.

```bash
cd scorpion/darci-python
pip install -e .

# First-time setup
scorpion onboard
# Edit ~/.scorpion/config.json:
#   providers.gemini.apiKey = "your key here"

# CLI agent mode
scorpion agent

# Persistent multi-channel gateway
scorpion gateway
```

**With DarCI tools (once darci/ is implemented):**
```bash
python -m darci
```

---

## Component 3: Taila2a Bridge

P2P messaging layer over Tailscale. Required for agent discovery and coordination.

```bash
cd tailbridge/taila2a
go build ./cmd/taila2a/

# First-time config
mkdir -p ~/.taila2a
cat > ~/.taila2a/config.json << 'EOF'
{
  "name": "darci-bridge",
  "auth_key": "tskey-auth-YOUR_KEY_HERE",
  "local_agent_url": "http://127.0.0.1:18790",
  "inbound_port": 8001,
  "peer_inbound_port": 8001,
  "local_listen": "127.0.0.1:8080"
}
EOF

./taila2a
```

**Local API:** `http://localhost:8080`
- `GET /agents` — phone book: all peers on tailnet
- `POST /send` — send to peer: `{"dest_node": "name", "payload": {...}}`
- `GET /status` — bridge health

---

## Component 4: TailFS (File Transfer)

Optional. Used for context file sync between agents.

```bash
cd tailbridge/tailfs
go build ./cmd/tailfs/
./tailfs
# API: http://localhost:8081
```

---

## Component 5: Frontend

```bash
cd frontend
npm install
npm start
# Dev server: http://localhost:3000
```

---

## Worker Agents

### openclaw (toy soldier — Python)
*Integration: must expose HTTP endpoint for taila2a `local_agent_url`*
*Inbound darci_directive handler: TBD*

### nanobot (Python/Go)
*Integration: TBD*

### zero claw / sclaw (Rust)
*Integration: must expose HTTP endpoint on tailnet*
*Inbound darci_directive handler: TBD*

---

## Connectivity Tests

```bash
# Sentinel backend alive?
curl http://localhost:8000/

# Taila2a phone book populated?
curl http://localhost:8080/agents

# Scorpion responds?
scorpion agent -m "hello"

# Send a test message via taila2a (replace <node-name>)
curl -X POST http://localhost:8080/send \
  -H "Content-Type: application/json" \
  -d '{"dest_node":"<node-name>","payload":{"type":"ping"}}'
```

---

## Troubleshooting

| Symptom | Cause | Fix |
|---|---|---|
| Sentinel 500 on connect | Missing GEMINI_API_KEY | Check .env file |
| taila2a /agents returns empty | Tailscale not connected | Run `tailscale status` |
| Scorpion "no provider" error | Missing apiKey in config | Edit ~/.scorpion/config.json |
| taila2a "connection refused" | Bridge not running | Start ./taila2a |
| Frontend blank | API not running | Start backend first |

---

## Run Everything (Dev)

```bash
# Terminal 1: Sentinel
uvicorn backend.main:app --port 8000 --reload

# Terminal 2: Taila2a bridge
./tailbridge/taila2a/taila2a

# Terminal 3: Scorpion/DarCI
scorpion gateway

# Terminal 4: Frontend (optional)
cd frontend && npm start
```
