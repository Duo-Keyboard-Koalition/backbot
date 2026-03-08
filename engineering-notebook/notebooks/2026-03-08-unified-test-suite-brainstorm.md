# Unified Test Suite & Engineering Notebook - Integration Brainstorm

**Created:** 2026-03-08  
**Author:** DarCI Engineering  
**Status:** 🟡 Brainstorming / Planning  
**Related:** [[2026-03-07-agent-bridge-network]], [[2026-03-08-darci-agent-test-status]]

---

## Executive Summary

This document proposes a unified test suite and engineering notebook integration strategy that:

1. **Consolidates** scattered test infrastructure across Tailbridge, DarCI Python, and DarCI Go
2. **Integrates** engineering notebook entries with test results for traceability
3. **Automates** test execution with notebook entry generation
4. **Standardizes** testing patterns across Python and Go implementations

---

## Current State Analysis

### Test Suite Landscape

| Component | Location | Language | Test Framework | Coverage | Status |
|-----------|----------|----------|----------------|----------|--------|
| Tailbridge A2A | `test_suite/tailbridge_test/` | Go | testify | ~70% | ✅ Active |
| DarCI Python | `darci/darci/darci-python/` | Python | None yet | 0% | 🔴 Missing |
| DarCI Go | `darci/darci/darci-go/` | Go | None yet | 0% | 🔴 Missing |

### Engineering Notebook Landscape

| Notebook | Location | Format | Linked to Tests | Auto-Generated |
|----------|----------|--------|-----------------|----------------|
| Agent Bridge Network | `engineering-notebook/notebooks/2026-03-07-agent-bridge-network.md` | Markdown | ❌ No | ❌ No |
| DarCI Test Status | `engineering-notebook/notebooks/2026-03-08-darci-agent-test-status.md` | Markdown | 🟡 Partial | ❌ No |

### Pain Points Identified

1. **Fragmented Test Infrastructure**
   - Tailbridge has comprehensive tests, DarCI has none
   - No unified test runner across components
   - Duplicate effort in test setup/teardown

2. **No Traceability**
   - Test results not linked to engineering decisions
   - No automatic notebook entry generation
   - Manual tracking of test coverage vs features

3. **Inconsistent Patterns**
   - Python and Go use different test frameworks
   - No shared test fixtures or mocks
   - Different reporting formats

4. **Missing Automation**
   - No CI/CD integration for DarCI tests
   - Manual test execution for new features
   - No automated regression detection

---

## Proposed Unified Architecture

### High-Level Design

```
┌─────────────────────────────────────────────────────────────────┐
│                    Unified Test Platform                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │              Test Orchestrator (Cross-Language)          │   │
│  │  • Unified CLI: `darci test [options]`                   │   │
│  │  • Multi-language support (Python + Go)                  │   │
│  │  • Parallel execution                                    │   │
│  │  • Coverage aggregation                                  │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                  │
│  ┌──────────────────┐         ┌──────────────────┐             │
│  │   DarCI Python   │         │   DarCI Go       │             │
│  │   Tests          │         │   Tests          │             │
│  │                  │         │                  │             │
│  │  • pytest        │         │  • go test       │             │
│  │  • asyncio       │         │  • testify       │             │
│  │  • fixtures      │         │  • table-driven  │             │
│  └──────────────────┘         └──────────────────┘             │
│                                                                  │
│  ┌──────────────────┐         ┌──────────────────┐             │
│  │   Tailbridge     │         │   Integration    │             │
│  │   Tests          │         │   Tests          │             │
│  │                  │         │                  │             │
│  │  • A2A protocol  │         │  • E2E scenarios │             │
│  │  • File transfer │         │  • Cross-agent   │             │
│  └──────────────────┘         └──────────────────┘             │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │           Engineering Notebook Generator                  │  │
│  │  • Auto-generates notebook entries from test results     │  │
│  │  • Links test coverage to feature documentation          │  │
│  │  • Tracks regression history                             │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Unified Directory Structure

### Proposed Layout

```
sentinelai/
├── test_suite/                          # Root test platform
│   ├── README.md                        # Unified test guide
│   ├── run-tests.ps1                    # Windows test runner
│   ├── run-tests.sh                     # Unix test runner
│   │
│   ├── tailbridge/                      # Existing Tailbridge tests
│   │   ├── mock/
│   │   ├── integration/
│   │   └── ...
│   │
│   ├── darci-python/                    # DarCI Python tests
│   │   ├── unit/
│   │   │   ├── agent/
│   │   │   ├── tools/
│   │   │   ├── state/
│   │   │   └── ...
│   │   ├── integration/
│   │   ├── fixtures/
│   │   └── conftest.py
│   │
│   ├── darci-go/                        # DarCI Go tests
│   │   ├── darci/
│   │   │   ├── agent/
│   │   │   │   ├── loop_test.go
│   │   │   │   └── tools/
│   │   │   ├── state/
│   │   │   └── ...
│   │   └── integration/
│   │
│   └── e2e/                             # Cross-component E2E tests
│       ├── test_agent_collaboration.py
│       ├── test_sentinel_monitoring.go
│       └── scenarios/
│
├── engineering-notebook/
│   ├── README.md
│   ├── index.json                       # Machine-readable index
│   ├── notebooks/
│   │   ├── 2026-03-07-agent-bridge-network.md
│   │   ├── 2026-03-08-darci-agent-test-status.md
│   │   └── auto-generated/
│   │       └── test-run-2026-03-08-143022.md
│   └── templates/
│       ├── feature-entry.md
│       └── test-run-entry.md
│
└── .github/
    └── workflows/
        └── test-and-doc.yml             # CI/CD pipeline
```

---

## Test Strategy by Component

### 1. DarCI Python Tests

#### Unit Tests (Priority: High)

```python
# test_suite/darci-python/unit/tools/test_task.py
import pytest
from unittest.mock import AsyncMock, patch
from darci.agent.tools.task import TaskCreateTool, TaskUpdateTool
from darci.state.store import TaskStore

class TestTaskCreateTool:
    @pytest.fixture
    def task_store(self, tmp_path):
        """Create temporary task store"""
        store = TaskStore(state_dir=tmp_path)
        return store

    @pytest.fixture
    def task_tool(self, task_store):
        return TaskCreateTool(task_store)

    async def test_create_task_success(self, task_tool):
        result = await task_tool.execute(
            title="Test task",
            description="Test description",
            priority="P1"
        )
        assert "T001" in result
        assert "Test task" in result

    async def test_create_task_with_labels(self, task_tool):
        result = await task_tool.execute(
            title="Labeled task",
            labels=["bug", "urgent"]
        )
        assert "bug" in result
        assert "urgent" in result

    async def test_create_task_invalid_priority(self, task_tool):
        with pytest.raises(ValueError, match="Invalid priority"):
            await task_tool.execute(
                title="Bad priority",
                priority="P99"
            )
```

#### Integration Tests (Priority: Medium)

```python
# test_suite/darci-python/integration/test_agent_workflow.py
import pytest
from darci.agent.loop import AdkAgentLoop
from darci.providers import MockProvider

class TestAgentWorkflow:
    @pytest.fixture
    def agent_loop(self):
        provider = MockProvider()
        return AdkAgentLoop(provider=provider)

    async def test_task_creation_and_assignment(self, agent_loop):
        # Simulate user command
        response = await agent_loop.run("Create a task to fix the login bug")
        assert "T001" in response

        # Assign to agent
        response = await agent_loop.run("Assign T001 to agent-alpha")
        assert "responsible" in response.lower()
```

#### Fixtures

```python
# test_suite/darci-python/conftest.py
import pytest
from pathlib import Path
from darci.config import DarciConfig
from darci.state.store import TaskStore

@pytest.fixture
def darci_config(tmp_path):
    """Create test configuration"""
    return DarciConfig(
        state_dir=tmp_path / "state",
        notebook_dir=tmp_path / "notebook",
        workspace_dir=tmp_path / "workspace"
    )

@pytest.fixture
def task_store(darci_config):
    """Create task store with test config"""
    return TaskStore(darci_config)

@pytest.fixture
def mock_llm_response():
    """Mock LLM response for testing"""
    return {
        "role": "assistant",
        "content": "Task created successfully",
        "tool_calls": []
    }
```

---

### 2. DarCI Go Tests

#### Unit Tests (Priority: High)

```go
// darci/state/store_test.go
package state

import (
    "os"
    "path/filepath"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestTaskStore_CreateNew(t *testing.T) {
    // Setup
    tmpDir := t.TempDir()
    cfg := &config.DarciConfig{StateDir: tmpDir}
    store, err := NewTaskStore(cfg)
    require.NoError(t, err)

    // Execute
    task, err := store.CreateNew("Test Task", "Description", "P1", nil, nil)

    // Verify
    require.NoError(t, err)
    assert.Equal(t, "T001", task.ID)
    assert.Equal(t, "Test Task", task.Title)
    assert.Equal(t, "pending", task.Status)
}

func TestTaskStore_Update(t *testing.T) {
    tmpDir := t.TempDir()
    store, _ := NewTaskStore(&config.DarciConfig{StateDir: tmpDir})

    // Create task
    task, _ := store.CreateNew("Original", "", "P2", nil, nil)

    // Update
    updated, err := store.Update(task.ID, map[string]interface{}{
        "status": "in_progress",
        "priority": "P0",
    })

    require.NoError(t, err)
    assert.Equal(t, "in_progress", updated.Status)
    assert.Equal(t, "P0", updated.Priority)
}

func TestTaskStore_SetAgentAssignment(t *testing.T) {
    tmpDir := t.TempDir()
    store, _ := NewTaskStore(&config.DarciConfig{StateDir: tmpDir})

    task, _ := store.CreateNew("Task", "", "P2", nil, nil)

    err := store.SetAgentAssignment("agent-alpha", task.ID, "responsible", 0.5, "in_progress")
    require.NoError(t, err)

    ctx, _ := store.GetContext()
    assignment := ctx.AgentAssignments["agent-alpha"]
    assert.Equal(t, task.ID, assignment.TaskID)
    assert.Equal(t, "responsible", assignment.DARCIROle)
    assert.Equal(t, 0.5, assignment.RiskScore)
}
```

#### Table-Driven Tests

```go
// darci/agent/tools/task_test.go
package tools

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestTaskCreateTool_Execute(t *testing.T) {
    tests := []struct {
        name        string
        args        map[string]interface{}
        wantErr     bool
        wantContains string
    }{
        {
            name: "success with minimal args",
            args: map[string]interface{}{
                "title": "Test task",
            },
            wantErr:      false,
            wantContains: "T001",
        },
        {
            name: "success with all args",
            args: map[string]interface{}{
                "title":       "Full task",
                "description": "Description",
                "priority":    "P1",
                "labels":      []interface{}{"bug", "urgent"},
            },
            wantErr:      false,
            wantContains: "Full task",
        },
        {
            name: "missing title",
            args: map[string]interface{}{},
            wantErr:      true,
            wantContains: "title is required",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            tmpDir := t.TempDir()
            store, _ := state.NewTaskStore(&config.DarciConfig{StateDir: tmpDir})
            tool := NewTaskCreateTool(store)

            // Execute
            result, err := tool.Execute(context.Background(), tt.args)

            // Verify
            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.wantContains)
            } else {
                assert.NoError(t, err)
                assert.Contains(t, result, tt.wantContains)
            }
        })
    }
}
```

---

### 3. Cross-Component Integration Tests

#### E2E Scenario: Agent Collaboration

```python
# test_suite/e2e/test_agent_collaboration.py
import pytest
import asyncio
from darci.agent.loop import AdkAgentLoop
from tailbridge import A2AProtocol

class TestAgentCollaboration:
    @pytest.fixture
    def driver_agent(self):
        return AdkAgentLoop(role="driver")

    @pytest.fixture
    def responsible_agent(self):
        return AdkAgentLoop(role="responsible")

    async def test_task_handoff(self, driver_agent, responsible_agent):
        # Driver creates task
        create_result = await driver_agent.run(
            "Create task: Fix authentication bug in login module"
        )
        assert "T001" in create_result

        # Driver assigns to responsible agent
        assign_result = await driver_agent.run(
            "Assign T001 to responsible-agent with goal: Fix the auth bug"
        )
        assert "assigned" in assign_result.lower()

        # Responsible agent receives task
        task_status = await responsible_agent.run("Show my tasks")
        assert "T001" in task_status
        assert "Fix authentication" in task_status
```

---

## Engineering Notebook Integration

### Automated Entry Generation

#### Test Run Template

```markdown
# Test Run Entry - {{timestamp}}

**Date:** {{date}}  
**Test Suite:** {{suite_name}}  
**Execution Mode:** {{mode}} (mock/integration/e2e)  
**Duration:** {{duration}}  

## Summary

{{summary}}

## Test Results

| Category | Total | Passed | Failed | Skipped | Coverage |
|----------|-------|--------|--------|---------|----------|
| Unit     | {{unit_total}} | {{unit_pass}} | {{unit_fail}} | {{unit_skip}} | {{unit_cov}}% |
| Integration | {{int_total}} | {{int_pass}} | {{int_fail}} | {{int_skip}} | {{int_cov}}% |
| E2E      | {{e2e_total}} | {{e2e_pass}} | {{e2e_fail}} | {{e2e_skip}} | - |

## New Tests Added

- [ ] `test_file.py::test_function` - {{description}}
- [ ] `test_file.go::TestFunction` - {{description}}

## Failures & Regressions

{{failure_details}}

## Files Changed

{{changed_files}}

## Follow-ups

- [ ] Fix failing test: {{test_name}}
- [ ] Add missing coverage for: {{component}}
- [ ] Update documentation for: {{feature}}

---

*Generated automatically by `darci test --generate-notebook`*
```

### CLI Integration

```python
# darci/cli/commands.py
import typer
from pathlib import Path

app = typer.Typer()

@app.command()
def test(
    suite: str = typer.Option("all", help="Test suite: all, python, go, tailbridge"),
    mode: str = typer.Option("mock", help="Mode: mock, integration, e2e"),
    coverage: bool = typer.Option(False, help="Generate coverage report"),
    generate_notebook: bool = typer.Option(False, help="Generate notebook entry"),
    verbose: bool = typer.Option(False, help="Verbose output"),
):
    """Run unified test suite"""
    runner = TestRunner(suite, mode, verbose)
    results = runner.run()

    if generate_notebook:
        from engineering_notebook import generate_test_entry
        generate_test_entry(results, coverage)

    if coverage:
        runner.show_coverage()

    if results.failed > 0:
        raise typer.Exit(code=1)

@app.command()
def notebook(
    action: str = typer.Option("create", help="Action: create, list, index"),
    template: str = typer.Option("feature", help="Template: feature, test-run, bug"),
):
    """Manage engineering notebook"""
    if action == "create":
        entry = NotebookEntry(template=template)
        entry.create()
    elif action == "list":
        NotebookEntry.list_all()
    elif action == "index":
        NotebookEntry.rebuild_index()
```

### Index File (Machine-Readable)

```json
// engineering-notebook/index.json
{
  "version": "1.0",
  "generated": "2026-03-08T14:30:22Z",
  "entries": [
    {
      "id": "2026-03-07-agent-bridge-network",
      "date": "2026-03-07",
      "title": "Agent Bridge Network",
      "area": "Infrastructure",
      "files_changed": [
        "bridge/main.go",
        ".env.bridge.example",
        "docs/agent-communication.md"
      ],
      "tests_linked": [
        "test_suite/tailbridge/mock/testify/a2a_test.go"
      ],
      "path": "notebooks/2026-03-07-agent-bridge-network.md"
    },
    {
      "id": "test-run-2026-03-08-143022",
      "date": "2026-03-08",
      "title": "Test Run - DarCI Python Unit Tests",
      "area": "Testing",
      "auto_generated": true,
      "test_results": {
        "total": 45,
        "passed": 42,
        "failed": 2,
        "skipped": 1,
        "coverage": 67.3
      },
      "path": "notebooks/auto-generated/test-run-2026-03-08-143022.md"
    }
  ],
  "test_coverage": {
    "darci-python": {
      "unit": 67.3,
      "integration": 45.2,
      "last_run": "2026-03-08T14:30:22Z"
    },
    "darci-go": {
      "unit": 72.1,
      "integration": 50.0,
      "last_run": "2026-03-08T14:30:22Z"
    },
    "tailbridge": {
      "unit": 85.4,
      "integration": 78.9,
      "last_run": "2026-03-08T14:30:22Z"
    }
  }
}
```

---

## CI/CD Integration

### GitHub Actions Workflow

```yaml
# .github/workflows/test-and-doc.yml
name: Tests & Documentation

on:
  push:
    branches: [main, feature/*]
  pull_request:
    branches: [main]

jobs:
  test-python:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Python
        uses: actions/setup-python@v5
        with:
          python-version: '3.11'

      - name: Install dependencies
        run: |
          cd darci/darci/darci-python
          pip install -e ".[dev]"

      - name: Run unit tests
        run: |
          cd test_suite/darci-python
          pytest unit/ -v --cov=darci --cov-report=xml

      - name: Run integration tests
        run: |
          cd test_suite/darci-python
          pytest integration/ -v -m "not slow"

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./test_suite/darci-python/coverage.xml
          flags: darci-python

  test-go:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.22'

      - name: Run tests
        run: |
          cd darci/darci/darci-go
          go test -v -coverprofile=coverage.out ./...

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./darci/darci/darci-go/coverage.out
          flags: darci-go

  test-tailbridge:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v4

      - name: Run mock tests
        run: |
          cd test_suite/tailbridge
          go test ./mock/... -v

      - name: Run integration tests
        run: |
          cd test_suite/tailbridge
          go test ./integration/... -v -tags=integration

  generate-notebook:
    needs: [test-python, test-go, test-tailbridge]
    runs-on: ubuntu-latest
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4

      - name: Generate notebook entry
        run: |
          python scripts/generate-test-notebook.py \
            --output engineering-notebook/notebooks/auto-generated/ \
            --test-results test-results.json

      - name: Commit notebook entry
        run: |
          git config user.name "GitHub Actions"
          git config user.email "actions@github.com"
          git add engineering-notebook/notebooks/auto-generated/
          git commit -m "docs: auto-generate test run notebook entry" || echo "No changes"
          git push
```

---

## Implementation Roadmap

### Phase 1: Foundation (Week 1-2)

- [ ] Create unified `test_suite/` directory structure
- [ ] Set up DarCI Python test infrastructure (pytest, fixtures)
- [ ] Set up DarCI Go test infrastructure (go test, testify)
- [ ] Create base test fixtures and mocks for both languages
- [ ] Document testing conventions

### Phase 2: Core Tests (Week 3-4)

- [ ] Implement DarCI Python unit tests:
  - [ ] `test_task_create.py`
  - [ ] `test_task_update.py`
  - [ ] `test_state_store.py`
  - [ ] `test_agent_loop.py`
- [ ] Implement DarCI Go unit tests:
  - [ ] `state/store_test.go`
  - [ ] `agent/tools/task_test.go`
  - [ ] `agent/loop_test.go`
- [ ] Achieve 60%+ unit test coverage

### Phase 3: Integration (Week 5-6)

- [ ] Create integration test framework
- [ ] Implement cross-agent communication tests
- [ ] Implement Sentinel monitoring tests
- [ ] Add E2E scenario tests
- [ ] Set up Docker-based integration testing

### Phase 4: Automation (Week 7-8)

- [ ] Build unified test CLI (`darci test`)
- [ ] Implement notebook entry generator
- [ ] Create machine-readable index (`index.json`)
- [ ] Set up CI/CD pipeline
- [ ] Configure coverage reporting and thresholds

### Phase 5: Polish (Week 9-10)

- [ ] Add performance benchmarks
- [ ] Implement flaky test detection
- [ ] Create test dashboards
- [ ] Document troubleshooting guide
- [ ] Conduct testing workshop for team

---

## Success Metrics

| Metric | Baseline | Target (3 months) | Target (6 months) |
|--------|----------|-------------------|-------------------|
| DarCI Python Coverage | 0% | 60% | 80% |
| DarCI Go Coverage | 0% | 60% | 80% |
| Tailbridge Coverage | 70% | 80% | 90% |
| Unit Tests | 0 | 150+ | 300+ |
| Integration Tests | 0 | 30+ | 60+ |
| E2E Scenarios | 0 | 10+ | 25+ |
| Auto-Generated Notebook Entries | 0 | 100% of test runs | 100% + linked |
| CI/CD Test Duration | N/A | <10 min | <5 min |

---

## Open Questions & Decisions

### Architecture

1. **Test Data Management**
   - Should test fixtures be shared across Python/Go?
   - Use JSON/YAML for language-agnostic test data?
   - **Decision:** Use JSON for shared fixtures, language-specific for unit tests

2. **Mock Strategy**
   - How much to mock vs. use real services?
   - Should we create a unified mock LLM provider?
   - **Decision:** Mock LLM for unit tests, real for integration

3. **Notebook Granularity**
   - One entry per test run or per feature?
   - Should failures create separate entries?
   - **Decision:** Per test run for automation, per feature for manual entries

### Tooling

4. **Test Runner**
   - Build custom CLI or use existing (tox, make)?
   - **Decision:** Custom `darci test` CLI for unified experience

5. **Coverage Reporting**
   - Use Codecov, Coveralls, or self-hosted?
   - **Decision:** Codecov (free for open source, good GitHub integration)

6. **Notebook Format**
   - Pure Markdown or add frontmatter/YAML?
   - **Decision:** Markdown with YAML frontmatter for metadata

---

## Risks & Mitigations

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Test suite becomes slow to run | High | Medium | Parallel execution, test categorization (fast/slow) |
| Mock tests diverge from real behavior | High | Medium | Regular integration test validation |
| Notebook entries become noise | Medium | High | Configurable generation, summary mode |
| Team resistance to new patterns | Medium | Medium | Documentation, workshops, gradual rollout |
| CI/CD pipeline complexity | Medium | Low | Start simple, iterate, use templates |

---

## Appendix A: Quick Reference Commands

```bash
# Run all tests
darci test --all

# Run Python tests only
darci test --suite python --mode unit

# Run Go tests with coverage
darci test --suite go --coverage

# Run integration tests
darci test --mode integration --verbose

# Generate notebook entry
darci test --generate-notebook

# View coverage report
darci test --coverage --report html

# List notebook entries
darci notebook list

# Create new feature entry
darci notebook create --template feature
```

---

## Appendix B: Related Documents

- [Tailbridge Test Platform README](../../test_suite/tailbridge_test/README.md)
- [Tailbridge Testing Guide](../../test_suite/tailbridge_test/TESTING_GUIDE.md)
- [DarCI Test Status Brainstorm](./2026-03-08-darci-agent-test-status.md)
- [Agent Bridge Network](./2026-03-07-agent-bridge-network.md)
- [SOUL.md](../../darci/workspace/SOUL.md)

---

*This is a living document. Update as the unified test suite evolves.*

**Next Review:** 2026-03-15  
**Owner:** DarCI Engineering
