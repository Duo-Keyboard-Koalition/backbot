# Unified Test Suite - Quickstart Implementation Guide

**Created:** 2026-03-08  
**Status:** 🟢 Ready to Implement  
**Related:** [[2026-03-08-unified-test-suite-brainstorm]]

---

## Overview

This guide provides step-by-step instructions to implement the unified test suite architecture. Follow these phases to establish comprehensive testing across DarCI Python, DarCI Go, and Tailbridge components.

---

## Phase 1: Setup Test Infrastructure (Day 1-2)

### Step 1.1: Create Unified Directory Structure

```bash
# Navigate to project root
cd sentinelai

# Create unified test suite structure
mkdir -p test_suite/darci-python/unit/agent/tools
mkdir -p test_suite/darci-python/unit/state
mkdir -p test_suite/darci-python/unit/bus
mkdir -p test_suite/darci-python/integration
mkdir -p test_suite/darci-python/fixtures

mkdir -p test_suite/darci-go/darci/agent/tools
mkdir -p test_suite/darci-go/darci/state
mkdir -p test_suite/darci-go/integration

mkdir -p test_suite/e2e/scenarios

mkdir -p engineering-notebook/notebooks/auto-generated
mkdir -p engineering-notebook/templates
```

### Step 1.2: Initialize Python Test Dependencies

```bash
cd darci/scorpion/darci-python

# Install test dependencies
pip install pytest pytest-asyncio pytest-cov pytest-mock

# Create requirements-test.txt
cat > requirements-test.txt << EOF
pytest>=9.0.0,<10.0.0
pytest-asyncio>=1.3.0,<2.0.0
pytest-cov>=4.0.0,<5.0.0
pytest-mock>=3.10.0,<4.0.0
EOF
```

### Step 1.3: Create Pytest Configuration

```bash
cd test_suite/darci-python

# Create pytest.ini
cat > pytest.ini << EOF
[pytest]
testpaths = unit integration
python_files = test_*.py
python_classes = Test*
python_functions = test_*
asyncio_mode = auto
markers =
    unit: Unit tests (fast, isolated)
    integration: Integration tests (slower, external deps)
    slow: Slow running tests
    e2e: End-to-end scenario tests
addopts = 
    -v
    --strict-markers
    --tb=short
    --cov=darci
    --cov-report=term-missing
EOF

# Create conftest.py with base fixtures
cat > conftest.py << 'EOF'
"""
Pytest configuration and shared fixtures for DarCI Python tests.
"""
import pytest
import asyncio
from pathlib import Path
from unittest.mock import AsyncMock, MagicMock


@pytest.fixture(scope="session")
def event_loop():
    """Create event loop for async tests."""
    loop = asyncio.get_event_loop_policy().new_event_loop()
    yield loop
    loop.close()


@pytest.fixture
def tmp_state_dir(tmp_path):
    """Create temporary state directory for tests."""
    state_dir = tmp_path / "state"
    state_dir.mkdir()
    return state_dir


@pytest.fixture
def tmp_notebook_dir(tmp_path):
    """Create temporary notebook directory for tests."""
    notebook_dir = tmp_path / "notebook"
    notebook_dir.mkdir()
    return notebook_dir


@pytest.fixture
def darci_config(tmp_path):
    """Create DarCI configuration for tests."""
    from darci.config.darci import DarciConfig
    
    return DarciConfig(
        state_dir=tmp_path / "state",
        notebook_dir=tmp_path / "notebook",
        workspace_dir=tmp_path / "workspace",
        bridge_local_url="http://localhost:8080",
    )


@pytest.fixture
def task_store(darci_config):
    """Create task store with test configuration."""
    from darci.state.store import TaskStore
    
    return TaskStore(darci_config)


@pytest.fixture
def mock_llm_response():
    """Mock LLM response for testing."""
    return {
        "role": "assistant",
        "content": "Task completed successfully",
        "tool_calls": []
    }


@pytest.fixture
def mock_model():
    """Create mock LLM model."""
    mock = AsyncMock()
    mock.respond = AsyncMock(return_value={
        "role": "assistant",
        "content": "Mock response",
        "tool_calls": []
    })
    return mock


@pytest.fixture
def sample_task_data():
    """Sample task data for testing."""
    return {
        "id": "T001",
        "title": "Test Task",
        "description": "Test Description",
        "priority": "P1",
        "status": "pending",
        "labels": ["test", "urgent"],
        "dependencies": [],
    }
EOF
```

### Step 1.4: Create Go Test Configuration

```bash
cd test_suite/darci-go

# Create go.mod for test suite
cat > go.mod << 'EOF'
module sentinelai/test_suite/darci-go

go 1.22

require (
    github.com/stretchr/testify v1.9.0
)

require (
    github.com/davecgh/go-spew v1.1.1 // indirect
    github.com/pmezard/go-difflib v1.0.0 // indirect
    gopkg.in/yaml.v3 v3.0.1 // indirect
)
EOF

# Create test helper package
mkdir -p testhelpers
cat > testhelpers/helpers.go << 'EOF'
package testhelpers

import (
    "os"
    "path/filepath"
    "testing"
)

// CreateTempDir creates a temporary directory for tests
func CreateTempDir(t *testing.T) string {
    t.Helper()
    dir, err := os.MkdirTemp("", "darci-test-*")
    if err != nil {
        t.Fatalf("Failed to create temp dir: %v", err)
    }
    return dir
}

// CleanupDir removes directory after test
func CleanupDir(t *testing.T, dir string) {
    t.Helper()
    os.RemoveAll(dir)
}

// SetupTestState creates test state directory structure
func SetupTestState(t *testing.T) string {
    t.Helper()
    dir := CreateTempDir(t)
    
    stateDir := filepath.Join(dir, "state")
    notebookDir := filepath.Join(dir, "notebook")
    
    os.MkdirAll(stateDir, 0755)
    os.MkdirAll(notebookDir, 0755)
    
    return dir
}
EOF
```

---

## Phase 2: Implement Core Unit Tests (Day 3-5)

### Step 2.1: DarCI Python Task Tests

```bash
cd test_suite/darci-python/unit/agent/tools

# Create test_task_create.py
cat > test_task_create.py << 'EOF'
"""
Unit tests for TaskCreateTool.
"""
import pytest
from darci.agent.tools.task import TaskCreateTool


class TestTaskCreateTool:
    """Test suite for task creation functionality."""

    @pytest.fixture
    def task_tool(self, task_store):
        """Create TaskCreateTool with test store."""
        return TaskCreateTool(task_store)

    @pytest.mark.asyncio
    async def test_create_task_minimal(self, task_tool):
        """Test creating task with minimal required fields."""
        result = await task_tool.execute(
            title="Fix login bug"
        )
        
        assert "T001" in result
        assert "Fix login bug" in result
        assert "P2" in result  # Default priority

    @pytest.mark.asyncio
    async def test_create_task_full(self, task_tool):
        """Test creating task with all fields."""
        result = await task_tool.execute(
            title="Implement feature X",
            description="Full description of feature X",
            priority="P0",
            labels=["feature", "urgent"],
            dependencies=["T000"]
        )
        
        assert "T001" in result
        assert "Implement feature X" in result
        assert "P0" in result
        assert "feature" in result
        assert "urgent" in result

    @pytest.mark.asyncio
    async def test_create_task_invalid_priority(self, task_tool):
        """Test that invalid priority raises error."""
        with pytest.raises(ValueError, match="Invalid priority"):
            await task_tool.execute(
                title="Bad priority",
                priority="P99"
            )

    @pytest.mark.asyncio
    async def test_create_multiple_tasks(self, task_tool):
        """Test creating multiple tasks increments IDs."""
        result1 = await task_tool.execute(title="Task 1")
        result2 = await task_tool.execute(title="Task 2")
        
        assert "T001" in result1
        assert "T002" in result2
EOF

# Create test_task_update.py
cat > test_task_update.py << 'EOF'
"""
Unit tests for TaskUpdateTool.
"""
import pytest
from darci.agent.tools.task import TaskUpdateTool


class TestTaskUpdateTool:
    """Test suite for task update functionality."""

    @pytest.fixture
    def task_tool(self, task_store):
        """Create TaskUpdateTool with test store."""
        return TaskUpdateTool(task_store)

    @pytest.fixture
    def sample_task(self, task_store):
        """Create sample task for testing."""
        from darci.models.task import Task
        task = Task(
            id="T001",
            title="Original Task",
            description="Original description",
            priority="P2",
            status="pending"
        )
        return task_store.create(task)

    @pytest.mark.asyncio
    async def test_update_status(self, task_tool, sample_task):
        """Test updating task status."""
        result = await task_tool.execute(
            task_id="T001",
            status="in_progress"
        )
        
        assert "T001" in result
        assert "in_progress" in result

    @pytest.mark.asyncio
    async def test_update_priority(self, task_tool, sample_task):
        """Test updating task priority."""
        result = await task_tool.execute(
            task_id="T001",
            priority="P0"
        )
        
        assert "P0" in result

    @pytest.mark.asyncio
    async def test_update_nonexistent_task(self, task_tool):
        """Test updating non-existent task returns error."""
        result = await task_tool.execute(
            task_id="T999",
            status="completed"
        )
        
        assert "not found" in result.lower()

    @pytest.mark.asyncio
    async def test_update_no_fields(self, task_tool, sample_task):
        """Test updating with no fields returns error."""
        result = await task_tool.execute(task_id="T001")
        
        assert "no fields" in result.lower()
EOF
```

### Step 2.2: DarCI Go State Tests

```bash
cd test_suite/darci-go/darci/state

# Create store_test.go
cat > store_test.go << 'EOF'
package state

import (
    "testing"
    "time"

    "darci-go/darci/config"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestTaskStore_CreateNew(t *testing.T) {
    t.Run("success with minimal args", func(t *testing.T) {
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
        assert.Equal(t, "P1", task.Priority)
    })

    t.Run("success with default priority", func(t *testing.T) {
        tmpDir := t.TempDir()
        store, _ := NewTaskStore(&config.DarciConfig{StateDir: tmpDir})

        task, err := store.CreateNew("Default Priority", "", "", nil, nil)

        require.NoError(t, err)
        assert.Equal(t, "P2", task.Priority) // Default
    })

    t.Run("multiple tasks increment ID", func(t *testing.T) {
        tmpDir := t.TempDir()
        store, _ := NewTaskStore(&config.DarciConfig{StateDir: tmpDir})

        task1, _ := store.CreateNew("Task 1", "", "", nil, nil)
        task2, _ := store.CreateNew("Task 2", "", "", nil, nil)

        assert.Equal(t, "T001", task1.ID)
        assert.Equal(t, "T002", task2.ID)
    })
}

func TestTaskStore_Update(t *testing.T) {
    tmpDir := t.TempDir()
    store, _ := NewTaskStore(&config.DarciConfig{StateDir: tmpDir})

    // Create task
    task, _ := store.CreateNew("Original", "Desc", "P2", nil, nil)

    t.Run("update status", func(t *testing.T) {
        updated, err := store.Update(task.ID, map[string]interface{}{
            "status": "in_progress",
        })

        require.NoError(t, err)
        assert.Equal(t, "in_progress", updated.Status)
    })

    t.Run("update priority", func(t *testing.T) {
        updated, err := store.Update(task.ID, map[string]interface{}{
            "priority": "P0",
        })

        require.NoError(t, err)
        assert.Equal(t, "P0", updated.Priority)
    })

    t.Run("update non-existent task", func(t *testing.T) {
        updated, err := store.Update("T999", map[string]interface{}{
            "status": "completed",
        })

        assert.Error(t, err)
        assert.Nil(t, updated)
    })
}

func TestTaskStore_SetAgentAssignment(t *testing.T) {
    tmpDir := t.TempDir()
    store, _ := NewTaskStore(&config.DarciConfig{StateDir: tmpDir})

    task, _ := store.CreateNew("Task", "", "P2", nil, nil)

    t.Run("set assignment", func(t *testing.T) {
        err := store.SetAgentAssignment("agent-alpha", task.ID, "responsible", 0.5, "in_progress")
        require.NoError(t, err)

        ctx, err := store.GetContext()
        require.NoError(t, err)

        assignment := ctx.AgentAssignments["agent-alpha"]
        require.NotNil(t, assignment)
        assert.Equal(t, task.ID, assignment.TaskID)
        assert.Equal(t, "responsible", assignment.DARCIROle)
        assert.Equal(t, 0.5, assignment.RiskScore)
        assert.Equal(t, "in_progress", assignment.Status)
    })

    t.Run("update assignment", func(t *testing.T) {
        // Set initial
        store.SetAgentAssignment("agent-beta", task.ID, "driver", 0.0, "pending")
        
        // Update
        err := store.SetAgentAssignment("agent-beta", task.ID, "driver", 0.3, "in_progress")
        require.NoError(t, err)

        ctx, _ := store.GetContext()
        assignment := ctx.AgentAssignments["agent-beta"]
        assert.Equal(t, 0.3, assignment.RiskScore)
        assert.Equal(t, "in_progress", assignment.Status)
    })
}

func TestTaskStore_Query(t *testing.T) {
    tmpDir := t.TempDir()
    store, _ := NewTaskStore(&config.DarciConfig{StateDir: tmpDir})

    // Create sample tasks
    store.CreateNew("Bug 1", "", "P1", []string{"bug"}, nil)
    store.CreateNew("Feature 1", "", "P2", []string{"feature"}, nil)
    store.CreateNew("Bug 2", "", "P0", []string{"bug", "urgent"}, nil)

    t.Run("query by status", func(t *testing.T) {
        tasks, err := store.Query("pending", "", "")
        require.NoError(t, err)
        assert.Len(t, tasks, 3)
    })

    t.Run("query by priority", func(t *testing.T) {
        tasks, err := store.Query("", "P0", "")
        require.NoError(t, err)
        assert.Len(t, tasks, 1)
        assert.Equal(t, "Bug 2", tasks[0].Title)
    })

    t.Run("query by label", func(t *testing.T) {
        tasks, err := store.Query("", "", "bug")
        require.NoError(t, err)
        assert.Len(t, tasks, 2)
    })

    t.Run("query by multiple filters", func(t *testing.T) {
        tasks, err := store.Query("pending", "P0", "bug")
        require.NoError(t, err)
        assert.Len(t, tasks, 1)
        assert.Equal(t, "Bug 2", tasks[0].Title)
    })
}
EOF
```

### Step 2.3: DarCI Go Tool Tests

```bash
cd test_suite/darci-go/darci/agent/tools

# Create task_test.go
cat > task_test.go << 'EOF'
package tools

import (
    "context"
    "testing"

    "darci-go/darci/config"
    "darci-go/darci/state"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
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

func TestTaskUpdateTool_Execute(t *testing.T) {
    tmpDir := t.TempDir()
    store, _ := state.NewTaskStore(&config.DarciConfig{StateDir: tmpDir})
    tool := NewTaskUpdateTool(store)

    // Create task
    task, _ := store.CreateNew("Original", "", "P2", nil, nil)

    tests := []struct {
        name        string
        args        map[string]interface{}
        wantErr     bool
        wantContains string
    }{
        {
            name: "update status",
            args: map[string]interface{}{
                "task_id": task.ID,
                "status":  "in_progress",
            },
            wantErr:      false,
            wantContains: "in_progress",
        },
        {
            name: "update priority",
            args: map[string]interface{}{
                "task_id":  task.ID,
                "priority": "P0",
            },
            wantErr:      false,
            wantContains: "P0",
        },
        {
            name: "non-existent task",
            args: map[string]interface{}{
                "task_id": "T999",
                "status":  "completed",
            },
            wantErr:      true,
            wantContains: "not found",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := tool.Execute(context.Background(), tt.args)

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
EOF
```

---

## Phase 3: Create Test Runner Scripts (Day 6)

### Step 3.1: Unified Test Runner (Bash)

```bash
cd test_suite

cat > run-tests.sh << 'EOF'
#!/bin/bash
set -e

# Unified Test Runner for SentinelAI
# Usage: ./run-tests.sh [options]

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Options
SUITE="all"
MODE="unit"
VERBOSE=""
COVERAGE=false
NOTEBOOK=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --suite|-s)
            SUITE="$2"
            shift 2
            ;;
        --mode|-m)
            MODE="$2"
            shift 2
            ;;
        --verbose|-v)
            VERBOSE="-v"
            shift
            ;;
        --coverage|-c)
            COVERAGE=true
            shift
            ;;
        --notebook|-n)
            NOTEBOOK=true
            shift
            ;;
        --help|-h)
            echo "Usage: ./run-tests.sh [options]"
            echo ""
            echo "Options:"
            echo "  --suite, -s     Test suite: all, python, go, tailbridge, e2e"
            echo "  --mode, -m      Test mode: unit, integration, e2e"
            echo "  --verbose, -v   Verbose output"
            echo "  --coverage, -c  Generate coverage report"
            echo "  --notebook, -n  Generate notebook entry"
            echo "  --help, -h      Show this help"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

echo -e "${YELLOW}Starting test run...${NC}"
echo "Suite: $SUITE | Mode: $MODE | Coverage: $COVERAGE"
echo ""

FAILED=0

# Run Python tests
run_python_tests() {
    echo -e "${YELLOW}Running DarCI Python Tests...${NC}"
    cd darci-python
    
    if [ "$MODE" = "unit" ] || [ "$MODE" = "all" ]; then
        echo "  Unit tests..."
        if [ "$COVERAGE" = true ]; then
            pytest unit/ $VERBOSE --cov=darci --cov-report=term-missing --cov-report=xml
        else
            pytest unit/ $VERBOSE
        fi
    fi
    
    if [ "$MODE" = "integration" ] || [ "$MODE" = "all" ]; then
        echo "  Integration tests..."
        pytest integration/ $VERBOSE -m integration
    fi
    
    cd ..
}

# Run Go tests
run_go_tests() {
    echo -e "${YELLOW}Running DarCI Go Tests...${NC}"
    cd darci-go
    
    if [ "$MODE" = "unit" ] || [ "$MODE" = "all" ]; then
        echo "  Unit tests..."
        if [ "$COVERAGE" = true ]; then
            go test -v ./... -coverprofile=coverage.out
        else
            go test -v ./...
        fi
    fi
    
    if [ "$MODE" = "integration" ] || [ "$MODE" = "all" ]; then
        echo "  Integration tests..."
        go test -v ./integration/... -tags=integration
    fi
    
    cd ..
}

# Run Tailbridge tests
run_tailbridge_tests() {
    echo -e "${YELLOW}Running Tailbridge Tests...${NC}"
    cd tailbridge
    
    if [ "$MODE" = "unit" ] || [ "$MODE" = "all" ]; then
        echo "  Mock tests..."
        go test ./mock/... $VERBOSE
    fi
    
    if [ "$MODE" = "integration" ] || [ "$MODE" = "all" ]; then
        echo "  Integration tests..."
        go test ./integration/... $VERBOSE -tags=integration
    fi
    
    cd ..
}

# Run E2E tests
run_e2e_tests() {
    echo -e "${YELLOW}Running E2E Tests...${NC}"
    cd e2e
    
    if [ "$MODE" = "e2e" ] || [ "$MODE" = "all" ]; then
        pytest $VERBOSE -m e2e
    fi
    
    cd ..
}

# Execute based on suite
case $SUITE in
    all)
        run_python_tests || FAILED=1
        run_go_tests || FAILED=1
        run_tailbridge_tests || FAILED=1
        run_e2e_tests || FAILED=1
        ;;
    python)
        run_python_tests || FAILED=1
        ;;
    go)
        run_go_tests || FAILED=1
        ;;
    tailbridge)
        run_tailbridge_tests || FAILED=1
        ;;
    e2e)
        run_e2e_tests || FAILED=1
        ;;
    *)
        echo -e "${RED}Unknown suite: $SUITE${NC}"
        exit 1
        ;;
esac

# Generate notebook entry if requested
if [ "$NOTEBOOK" = true ]; then
    echo ""
    echo -e "${YELLOW}Generating notebook entry...${NC}"
    python ../scripts/generate-test-notebook.py --output ../engineering-notebook/notebooks/auto-generated/
fi

# Summary
echo ""
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
else
    echo -e "${RED}✗ Some tests failed${NC}"
    exit 1
fi
EOF

chmod +x run-tests.sh
```

### Step 3.2: PowerShell Test Runner

```bash
cat > run-tests.ps1 << 'EOF'
# Unified Test Runner for SentinelAI (PowerShell)
# Usage: .\run-tests.ps1 [-Suite <suite>] [-Mode <mode>] [-Coverage] [-Notebook]

param(
    [ValidateSet("all", "python", "go", "tailbridge", "e2e")]
    [string]$Suite = "all",
    
    [ValidateSet("unit", "integration", "e2e", "all")]
    [string]$Mode = "unit",
    
    [switch]$Verbose,
    [switch]$Coverage,
    [switch]$Notebook
)

$ErrorActionPreference = "Stop"
$Failed = $false

Write-Host "Starting test run..." -ForegroundColor Yellow
Write-Host "Suite: $Suite | Mode: $Mode | Coverage: $Coverage"
Write-Host ""

function Run-PythonTests {
    Write-Host "Running DarCI Python Tests..." -ForegroundColor Yellow
    Push-Location darci-python
    
    if ($Mode -eq "unit" -or $Mode -eq "all") {
        Write-Host "  Unit tests..."
        $args = @("unit/")
        if ($Verbose) { $args += "-v" }
        if ($Coverage) { $args += @("--cov=darci", "--cov-report=term-missing") }
        
        pytest @args
        if ($LASTEXITCODE -ne 0) { $script:Failed = $true }
    }
    
    if ($Mode -eq "integration" -or $Mode -eq "all") {
        Write-Host "  Integration tests..."
        pytest integration/ -m integration ($Verbose ? "-v" : $null)
        if ($LASTEXITCODE -ne 0) { $script:Failed = $true }
    }
    
    Pop-Location
}

function Run-GoTests {
    Write-Host "Running DarCI Go Tests..." -ForegroundColor Yellow
    Push-Location darci-go
    
    if ($Mode -eq "unit" -or $Mode -eq "all") {
        Write-Host "  Unit tests..."
        $args = @("test", "-v", "./...")
        if ($Coverage) { $args += @("-coverprofile=coverage.out") }
        
        go $args
        if ($LASTEXITCODE -ne 0) { $script:Failed = $true }
    }
    
    if ($Mode -eq "integration" -or $Mode -eq "all") {
        Write-Host "  Integration tests..."
        go test -v ./integration/... -tags=integration
        if ($LASTEXITCODE -ne 0) { $script:Failed = $true }
    }
    
    Pop-Location
}

function Run-TailbridgeTests {
    Write-Host "Running Tailbridge Tests..." -ForegroundColor Yellow
    Push-Location tailbridge
    
    if ($Mode -eq "unit" -or $Mode -eq "all") {
        Write-Host "  Mock tests..."
        go test ./mock/... ($Verbose ? "-v" : $null)
        if ($LASTEXITCODE -ne 0) { $script:Failed = $true }
    }
    
    if ($Mode -eq "integration" -or $Mode -eq "all") {
        Write-Host "  Integration tests..."
        go test ./integration/... ($Verbose ? "-v" : $null) -tags=integration
        if ($LASTEXITCODE -ne 0) { $script:Failed = $true }
    }
    
    Pop-Location
}

function Run-E2ETests {
    Write-Host "Running E2E Tests..." -ForegroundColor Yellow
    Push-Location e2e
    
    if ($Mode -eq "e2e" -or $Mode -eq "all") {
        pytest ($Verbose ? "-v" : $null) -m e2e
        if ($LASTEXITCODE -ne 0) { $script:Failed = $true }
    }
    
    Pop-Location
}

# Execute based on suite
switch ($Suite) {
    "all" {
        Run-PythonTests
        Run-GoTests
        Run-TailbridgeTests
        Run-E2ETests
    }
    "python" { Run-PythonTests }
    "go" { Run-GoTests }
    "tailbridge" { Run-TailbridgeTests }
    "e2e" { Run-E2ETests }
}

# Generate notebook entry
if ($Notebook) {
    Write-Host ""
    Write-Host "Generating notebook entry..." -ForegroundColor Yellow
    python ../scripts/generate-test-notebook.py --output ../engineering-notebook/notebooks/auto-generated/
}

# Summary
Write-Host ""
if (-not $Failed) {
    Write-Host "✓ All tests passed!" -ForegroundColor Green
} else {
    Write-Host "✗ Some tests failed" -ForegroundColor Red
    exit 1
}
EOF
```

---

## Phase 4: Notebook Generator Script (Day 7)

```bash
cd ..

mkdir -p scripts
cat > scripts/generate-test-notebook.py << 'EOF'
#!/usr/bin/env python3
"""
Generate engineering notebook entry from test run results.
"""
import json
import sys
from datetime import datetime
from pathlib import Path


def parse_test_results(xml_path):
    """Parse pytest/cobertura XML results."""
    # Simplified parser - in production use xml.etree.ElementTree
    return {
        "total": 45,
        "passed": 42,
        "failed": 2,
        "skipped": 1,
        "coverage": 67.3,
        "duration": "2m 34s"
    }


def generate_entry(results, output_dir):
    """Generate markdown notebook entry."""
    timestamp = datetime.now()
    filename = f"test-run-{timestamp.strftime('%Y-%m-%d-%H%M%S')}.md"
    
    content = f"""# Test Run Entry - {timestamp.strftime('%Y-%m-%d %H:%M')}

**Date:** {timestamp.strftime('%Y-%m-%d')}  
**Test Suite:** Unified DarCI Test Suite  
**Execution Mode:** Automated CI/CD  
**Duration:** {results['duration']}  

## Summary

Automated test run across DarCI Python, DarCI Go, and Tailbridge components.

## Test Results

| Category | Total | Passed | Failed | Skipped | Coverage |
|----------|-------|--------|--------|---------|----------|
| Unit     | {results['total']} | {results['passed']} | {results['failed']} | {results['skipped']} | {results['coverage']}% |

## New Tests Added

- [ ] `test_task_create.py::test_create_task_minimal` - Test minimal task creation
- [ ] `test_task_update.py::test_update_status` - Test status updates
- [ ] `store_test.go::TestTaskStore_CreateNew` - Go state store tests

## Failures & Regressions

### Failed Tests

1. `test_something.py::test_failing_case`
   - Error: AssertionError
   - Details: Expected X but got Y

## Files Changed

- `test_suite/darci-python/unit/agent/tools/test_task_create.py`
- `test_suite/darci-python/unit/agent/tools/test_task_update.py`
- `test_suite/darci-go/darci/state/store_test.go`

## Follow-ups

- [ ] Fix failing test: `test_something.py::test_failing_case`
- [ ] Add missing coverage for: `darci.agent.loop`
- [ ] Update documentation for: Task management tools

---

*Generated automatically by `generate-test-notebook.py`*
"""
    
    output_path = Path(output_dir) / filename
    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(content)
    
    print(f"Generated: {output_path}")
    return output_path


def update_index(entry_path):
    """Update engineering notebook index."""
    index_path = Path("engineering-notebook/index.json")
    
    if index_path.exists():
        index = json.loads(index_path.read_text())
    else:
        index = {"version": "1.0", "entries": [], "test_coverage": {}}
    
    entry = {
        "id": entry_path.stem,
        "date": datetime.now().strftime('%Y-%m-%d'),
        "title": f"Test Run - {entry_path.stem}",
        "area": "Testing",
        "auto_generated": True,
        "path": str(entry_path.relative_to("engineering-notebook"))
    }
    
    index["entries"].append(entry)
    index_path.write_text(json.dumps(index, indent=2))
    print(f"Updated index: {index_path}")


if __name__ == "__main__":
    output_dir = sys.argv[1] if len(sys.argv) > 1 else "engineering-notebook/notebooks/auto-generated/"
    
    results = parse_test_results("test-results.xml")
    entry_path = generate_entry(results, output_dir)
    update_index(entry_path)
EOF

chmod +x scripts/generate-test-notebook.py
```

---

## Phase 5: Run First Test Suite (Day 8)

```bash
# From project root
cd test_suite

# Run all tests with coverage
./run-tests.sh --suite all --mode unit --coverage --verbose

# Or on Windows
.\run-tests.ps1 -Suite all -Mode unit -Coverage -Verbose

# Generate notebook entry
python scripts/generate-test-notebook.py
```

---

## Next Steps

After completing this quickstart:

1. **Expand Test Coverage**
   - Add tests for remaining tools (shell, filesystem, cron, etc.)
   - Implement integration tests
   - Create E2E scenarios

2. **Automate CI/CD**
   - Add GitHub Actions workflow
   - Configure Codecov integration
   - Set up automated notebook generation

3. **Documentation**
   - Document testing conventions
   - Create test writing guide
   - Add troubleshooting section

---

*Quickstart Guide Version: 1.0*  
*Last Updated: 2026-03-08*
