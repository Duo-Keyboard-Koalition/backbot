# DarCI Agent Test Status - Brainstorming Context

**Created:** 2026-03-08  
**Last Updated:** 2026-03-08  
**Status:** 🟡 In Progress - Brainstorming

---

## Overview

This document serves as the centralized context for brainstorming, tracking, and documenting test status across all DarCI agent implementations (Python and Go).

---

## Agent Implementations

### 1. DarCI Python (`scorpion/darci-python/`)

#### Core Components

| Component | File Path | Test Status | Notes |
|-----------|-----------|-------------|-------|
| Agent Loop | `darci/agent/loop.py` | 🔴 Needs Tests | Core agent execution loop |
| Agent Tools | `darci/agent/tools/` | 🟡 Partial | Multiple tool implementations |
| - Base Tool | `darci/agent/tools/base.py` | 🔴 | Base class for all tools |
| - Cron | `darci/agent/tools/cron.py` | 🔴 | Scheduled task execution |
| - Filesystem | `darci/agent/tools/filesystem.py` | 🔴 | File operations |
| - Manage | `darci/agent/tools/manage.py` | 🔴 | Agent management |
| - MCP | `darci/agent/tools/mcp.py` | 🔴 | Model Context Protocol |
| - Message | `darci/agent/tools/message.py` | 🔴 | Messaging system |
| - Registry | `darci/agent/tools/registry.py` | 🔴 | Tool registration |
| - Shell | `darci/agent/tools/shell.py` | 🔴 | Shell command execution |
| - Spawn | `darci/agent/tools/spawn.py` | 🔴 | Sub-agent spawning |
| ADK | `darci/adk/` | 🔴 | Agent Development Kit |
| Bus/Events | `darci/bus/` | 🔴 | Event bus system |
| Channels | `darci/channels/` | 🔴 | Communication channels |
| Config | `darci/config/` | 🔴 | Configuration management |
| Cron Service | `darci/cron/` | 🔴 | Cron scheduling |
| Heartbeat | `darci/heartbeat/` | 🔴 | Health monitoring |
| Models | `darci/models/` | 🔴 | Data models |
| Providers | `darci/providers/` | 🔴 | LLM providers |
| State Store | `darci/state/store.py` | 🔴 | Persistent state |
| Skills | `darci/skills/` | 🔴 | Agent skills |

#### Test Infrastructure Needed

- [ ] Unit tests for each tool class
- [ ] Integration tests for agent loop
- [ ] Mock LLM provider for testing
- [ ] Test fixtures for common scenarios
- [ ] pytest configuration
- [ ] Coverage reporting setup

---

### 2. DarCI Go (`scorpion/darci-go/`)

#### Core Components

| Component | File Path | Test Status | Notes |
|-----------|-----------|-------------|-------|
| Agent Loop | `darci/agent/loop.go` | 🔴 Needs Tests | Core agent execution |
| Agent Tools | `darci/agent/tools/` | 🟡 Partial | Tool implementations |
| - Base | `darci/agent/tools/base.go` | 🔴 | Base tool interface |
| - Notebook | `darci/agent/tools/notebook.go` | 🔴 | Engineering notebook |
| - Register | `darci/agent/tools/register.go` | 🔴 | Tool registration |
| - Sentinel | `darci/agent/tools/sentinel.go` | 🔴 | Sentinel integration |
| - TailA2A | `darci/agent/tools/taila2a.go` | 🔴 | A2A protocol |
| - Task | `darci/agent/tools/task.go` | 🔴 | Task management |
| Config | `darci/config/darci.go` | 🔴 | Configuration |
| State | `darci/state/` | 🔴 | State management |
| Bridge | `bridge/` | 🔴 | Communication bridge |

#### Test Infrastructure Needed

- [ ] Go test files for each package (`*_test.go`)
- [ ] Mock interfaces for external dependencies
- [ ] Table-driven tests for tools
- [ ] Integration test suite
- [ ] Coverage reporting (`go test -cover`)
- [ ] Benchmark tests for performance

---

## Test Strategy Brainstorming

### Unit Testing Priorities

1. **Tool Implementations** (Highest Priority)
   - Each tool should have isolated unit tests
   - Mock external dependencies (APIs, filesystem, shell)
   - Test success and error paths

2. **Agent Loop** (High Priority)
   - Test state transitions
   - Test message handling
   - Test tool invocation flow

3. **State Management** (High Priority)
   - Test persistence layer
   - Test state recovery
   - Test concurrent access

4. **Communication Channels** (Medium Priority)
   - Test message serialization
   - Test channel routing
   - Test error handling

5. **Configuration** (Medium Priority)
   - Test config loading
   - Test validation
   - Test defaults

### Integration Testing Priorities

1. End-to-end agent execution
2. Multi-agent communication
3. Tool chaining scenarios
4. State persistence across restarts
5. Channel integrations (Slack, Email, etc.)

### Test Data & Fixtures

- Sample agent configurations
- Mock LLM responses
- Sample task definitions
- Test workspace structures
- Recorded event sequences

---

## Proposed Test Structure

### Python (`scorpion/darci-python/tests/`)

```
tests/
├── __init__.py
├── conftest.py              # Pytest fixtures
├── unit/
│   ├── __init__.py
│   ├── agent/
│   │   ├── test_loop.py
│   │   └── tools/
│   │       ├── test_base.py
│   │       ├── test_shell.py
│   │       └── ...
│   ├── bus/
│   │   └── test_events.py
│   ├── state/
│   │   └── test_store.py
│   └── ...
├── integration/
│   ├── test_agent_workflow.py
│   ├── test_tool_chaining.py
│   └── test_state_persistence.py
├── fixtures/
│   ├── sample_configs/
│   ├── mock_responses/
│   └── test_workspaces/
└── e2e/
    └── test_full_scenarios.py
```

### Go (`scorpion/darci-go/`)

```
darci/
├── agent/
│   ├── loop.go
│   ├── loop_test.go
│   └── tools/
│       ├── base.go
│       ├── base_test.go
│       └── ...
├── bus/
│   └── ...
├── state/
│   └── ...
└── ...
```

---

## CI/CD Integration

### GitHub Actions Workflow

```yaml
name: Tests
on: [push, pull_request]
jobs:
  python-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Set up Python
        uses: actions/setup-python@v5
      - name: Install dependencies
        run: pip install -r requirements.txt
      - name: Run tests
        run: pytest --cov=darci tests/
      - name: Upload coverage
        uses: codecov/codecov-action@v3

  go-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Set up Go
        uses: actions/setup-go@v4
      - name: Run tests
        run: go test -v -cover ./...
```

---

## Metrics & Goals

| Metric | Current | Target | Timeline |
|--------|---------|--------|----------|
| Python Unit Test Coverage | ~0% | 80% | Q2 2026 |
| Go Unit Test Coverage | ~0% | 80% | Q2 2026 |
| Integration Tests | 0 | 20+ | Q2 2026 |
| E2E Scenarios | 0 | 10+ | Q3 2026 |

---

## Open Questions

1. Should we use pytest-asyncio for async agent tests?
2. How do we test LLM-dependent behavior without API costs?
3. Should integration tests run against real services or mocks?
4. What's the strategy for testing skills that interact with external APIs?
5. How do we handle testing of tmux-based skills?

---

## Next Actions

- [ ] Set up pytest infrastructure for Python
- [ ] Create base test fixtures and mocks
- [ ] Implement first batch of tool unit tests (Python)
- [ ] Set up Go test structure
- [ ] Implement Go tool tests
- [ ] Configure CI/CD for automated testing
- [ ] Document testing conventions

---

## Related Documents

- [[Agent Architecture](../../darci/workspace/SOUL.md)]
- [[Tool Documentation](../../scorpion/darci-python/docs/)]
- [[Communication Protocol](../../tailbridge/taila2a/docs/agent-communication.md)]
