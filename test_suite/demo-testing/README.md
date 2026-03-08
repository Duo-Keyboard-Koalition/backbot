# Demo Testing Suite

This folder contains pre-show demo tests that demonstrate the DarCI coordination workflow.

## Quick Start

Run the demo immediately (no API keys or Docker required):

```bash
cd test_suite/demo-testing
python3 run_demo.py
```

**Note:** This demo uses **simulated agents** to demonstrate the DarCI coordination pattern. It does NOT require:
- ❌ Real API keys
- ❌ Docker containers
- ❌ Tailscale network
- ❌ Multiple agent processes

This is a **proof-of-concept demo** showing how DarCI would coordinate multiple nanobot agents.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│              Your Current Setup                         │
│                                                         │
│  GEMINI_API_KEY  → Shared across all agents            │
│  TS_AUTH_KEY     → Shared across all agents            │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│              Demo Test Setup                            │
│                                                         │
│  DarCI (simulated) ──→ Nanobot #1 (simulated)          │
│                    └─→ Nanobot #2 (simulated)          │
│                                                         │
│  No real API calls, no Docker, no Tailscale needed     │
└─────────────────────────────────────────────────────────┘
```

## Tests

### run_demo.py (Standalone Demo) ⭐ RECOMMENDED

Quick demonstration of DarCI coordinating 2 nanobot agents.

**What it demonstrates:**
- DarCI discovers available nanobots
- DarCI creates and assigns tasks
- DarCI monitors nanobot progress
- Tasks complete successfully

**Run it:**
```bash
python3 run_demo.py
```

**Output:**
```
🎬 Demo: DarCI Manages 2 Nanobot Agents
🚀 Starting nanobot agents...
  ✓ nanobot-worker-001 registered with DarCI
  ✓ nanobot-worker-002 registered with DarCI
🔍 Phase 1: Agent Discovery
  ✓ Discovered 2 agent(s)
📋 Phase 2: Task Creation
  ✓ 2 tasks created
📤 Phase 3: Task Assignment
  ✓ Tasks assigned to nanobots
⚡ Phase 4: Task Execution
  ✓ Both nanobots completed their tasks
🎉 Demo completed successfully!
```

## Prerequisites

### For Standalone Demo (run_demo.py)

**None!** Just Python 3:
```bash
python3 run_demo.py
```

### For Pytest Tests (test_darci_nanobot_simple.py)

```bash
pip install pytest pytest-asyncio pytest-timeout
pytest test_suite/demo-testing/test_darci_nanobot_simple.py -v
```

## Production Deployment (Future)

When you're ready to run **real agents** with Docker and actual API keys, you would need:

1. **Separate Docker containers** for each agent:
   - `darci-coordinator` container
   - `nanobot-worker-1` container
   - `nanobot-worker-2` container

2. **Tailscale auth keys** (one per agent or use tags):
   ```bash
   TS_AUTH_KEY_DARCI=tskey-auth-xxx
   TS_AUTH_KEY_NANOBOT1=tskey-auth-yyy
   TS_AUTH_KEY_NANOBOT2=tskey-auth-zzz
   ```

3. **Docker Compose** configuration with separate services

The current demo shows the **coordination pattern** without requiring this infrastructure.

## Test Flow

```
┌─────────────────┐
│   DarCI Agent   │ ← Coordinator/Manager
│  (Project Mgr)  │
└────────┬────────┘
         │
    ┌────┴────┐
    │         │
    ▼         ▼
┌─────────┐ ┌─────────┐
│ Nanobot │ │ Nanobot │ ← Workers
│   #1    │ │   #2    │
└─────────┘ └─────────┘
```

1. **Startup**: Launch 1 darci + 2 nanobot agents on tailnet
2. **Discovery**: DarCI discovers available nanobots
3. **Task Assignment**: DarCI assigns tasks to nanobots via taila2a
4. **Monitoring**: DarCI monitors nanobot progress
5. **Completion**: Verify all tasks completed successfully

## Expected Output

```
🎬 Starting demo: DarCI manages 2 nanobots
🚀 Starting DarCI agent...
✓ DarCI agent started: darci-coordinator-001
🚀 Starting Nanobot agent 1...
✓ Nanobot agent 1 started: nanobot-worker-001
🚀 Starting Nanobot agent 2...
✓ Nanobot agent 2 started: nanobot-worker-002
⏳ Waiting for agents to connect...
🔍 DarCI discovering agents...
✓ Discovered 2 agents: [nanobot-worker-001, nanobot-worker-002]
📋 DarCI creating tasks...
✓ Task TSK-001 created (P1)
✓ Task TSK-002 created (P1)
📤 DarCI assigning tasks to nanobots...
✓ Task TSK-001 assigned to nanobot-worker-001
✓ Task TSK-002 assigned to nanobot-worker-002
⏳ Monitoring nanobot progress...
✓ Nanobot #1 completed task TSK-001
✓ Nanobot #2 completed task TSK-002
🎉 Demo completed successfully!
```
