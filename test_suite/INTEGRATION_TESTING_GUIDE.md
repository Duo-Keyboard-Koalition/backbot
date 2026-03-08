# SentinelAI Integration Testing Guide

## Overview

This test suite provides **comprehensive integration testing** for SentinelAI using **real APIs only** (no mocks).

### What We Test

| Component | APIs Used | Test File |
|-----------|-----------|-----------|
| **Backend + Gemini** | Google Generative AI | `test_backend_gemini.py` |
| **Tailbridge + Tailscale** | Tailscale Network | `test_tailbridge.py` |
| **DarCI Coordination** | Gemini + Tailscale | `test_darci.py` |
| **End-to-End Workflows** | All APIs | `test_e2e.py` |

---

## Quick Start

### 1. Setup Environment

```bash
# Copy and configure test environment
cp .env.test.example .env.test

# Edit .env.test with your API keys:
# - GEMINI_API_KEY (from https://makersuite.google.com/app/apikey)
# - TS_AUTH_KEY_* (from https://login.tailscale.com/admin/settings/keys)
```

### 2. Install Dependencies

```bash
pip install pytest pytest-asyncio pytest-cov pytest-timeout httpx websockets google-generativeai
```

### 3. Verify Tailscale

```bash
# Ensure Tailscale is running
tailscale status

# If not connected:
tailscale up
```

### 4. Run Tests

```bash
# All tests
./test_suite/run-integration-tests.sh

# Specific category
./test_suite/run-integration-tests.sh -m gemini
./test_suite/run-integration-tests.sh -m tailscale

# With parallel execution
./test_suite/run-integration-tests.sh -n auto

# Skip slow/costly tests
./test_suite/run-integration-tests.sh -s slow -s api_cost
```

---

## Test Categories

### Backend + Gemini Tests (`test_backend_gemini.py`)

Tests the Sentinel backend with real Gemini API calls.

**Markers:** `@pytest.mark.gemini`, `@pytest.mark.api_cost`

**Example Tests:**
- Agent response parsing
- Agent step execution (web_search, calculate, write_to_file)
- Risk scoring (loop detection, goal drift, confidence, coherence)
- Interventions (reprompt, rollback, decompose, halt)

**Run:**
```bash
pytest test_suite/integration/test_backend_gemini.py -v
```

---

### Tailbridge + Tailscale Tests (`test_tailbridge.py`)

Tests real A2A communication and file transfer over Tailscale.

**Markers:** `@pytest.mark.tailscale`, `@pytest.mark.slow`

**Example Tests:**
- Agent health checks
- Phonebook discovery
- A2A messaging (send, receive, bidirectional)
- Message correlation
- Concurrent messaging
- File transfers (small, large, compressed, encrypted)

**Prerequisites:**
- Tailbridge agents running (Go binaries)
- Tailscale authenticated

**Run:**
```bash
pytest test_suite/integration/test_tailbridge.py -v
```

---

### DarCI Coordination Tests (`test_darci.py`)

Tests DarCI multi-agent coordination with Gemini and Tailscale.

**Markers:** `@pytest.mark.gemini`, `@pytest.mark.tailscale`, `@pytest.mark.slow`

**Example Tests:**
- Agent discovery
- Task assignment
- Progress monitoring
- Multi-agent collaboration
- Task reassignment
- Sentinel integration (alerts, interventions)

**Run:**
```bash
pytest test_suite/integration/test_darci.py -v
```

---

### End-to-End Tests (`test_e2e.py`)

Complete workflow tests exercising the entire SentinelAI pipeline.

**Markers:** `@pytest.mark.e2e`, `@pytest.mark.slow`, `@pytest.mark.load`

**Example Tests:**
- Full risk assessment workflow
- Multi-agent collaboration with monitoring
- WebSocket real-time monitoring
- Failure scenarios (rate limits, network partitions)
- Performance tests (concurrent execution, throughput)
- Real-world scenarios (research, code generation, data analysis)

**Run:**
```bash
pytest test_suite/integration/test_e2e.py -v
```

---

## Test Markers Reference

| Marker | Description | Skip If |
|--------|-------------|---------|
| `gemini` | Requires Gemini API key | No GEMINI_API_KEY |
| `tailscale` | Requires Tailscale connection | Tailscale not running |
| `e2e` | End-to-end workflow tests | - |
| `slow` | Takes >30 seconds | - |
| `api_cost` | Consumes API quota | - |
| `load` | Performance/load tests | - |

### Skip Markers During Run

```bash
# Skip slow tests
pytest test_suite/integration/ -m "not slow"

# Skip costly API tests
pytest test_suite/integration/ -m "not api_cost"

# Skip Tailscale tests
pytest test_suite/integration/ -m "not tailscale"
```

---

## Configuration

### Environment Variables (.env.test)

```bash
# Gemini API
GEMINI_API_KEY=AIzaSy...
GEMINI_FLASH_MODEL=gemini-2.0-flash
GEMINI_PRO_MODEL=gemini-2.0-pro

# Tailscale
TS_AUTH_KEY_1=tskey-auth-xxxxx1
TS_AUTH_KEY_2=tskey-auth-xxxxx2
TS_AUTH_KEY_3=tskey-auth-xxxxx3
TAILNET_NAME=your-tailnet.ts.net

# Test Configuration
TEST_TIMEOUT=300
TEST_RETRY_COUNT=3
```

### Pytest Configuration (pytest.ini)

```ini
[pytest]
asyncio_mode = auto
timeout = 300
markers =
    gemini: Requires Gemini API
    tailscale: Requires Tailscale connection
    e2e: End-to-end workflow tests
    slow: Slow-running tests (>30s)
    api_cost: Tests consuming API quota
    load: Performance/load tests
```

---

## Troubleshooting

### Gemini API Errors

**Problem:** `ResourceExhausted: Rate limit exceeded`

**Solutions:**
1. Add delays between tests: `time.sleep(2)`
2. Use `-m "not api_cost"` to skip costly tests
3. Request higher quota from Google AI Studio

**Problem:** `Invalid API key`

**Solutions:**
1. Verify key in `.env.test`
2. Check key at https://makersuite.google.com/app/apikey
3. Ensure `GEMINI_API_KEY` is exported

---

### Tailscale Errors

**Problem:** `Connection refused` or `Agent unavailable`

**Solutions:**
```bash
# Check Tailscale status
tailscale status

# Reconnect if needed
tailscale logout
tailscale up

# Verify agents are running
curl http://localhost:8081/health
curl http://localhost:8082/health
```

**Problem:** `tailscale: command not found`

**Solutions:**
```bash
# Install Tailscale (Ubuntu/WSL)
curl -fsSL https://tailscale.com/install.sh | sh

# Install Tailscale (Windows)
# Download from https://tailscale.com/download
```

---

### Test Hangs

**Problem:** Tests timeout or hang indefinitely

**Solutions:**
```bash
# Run with timeout
pytest test_suite/integration/ --timeout=60

# Increase timeout in .env.test
TEST_TIMEOUT=300

# Run specific test to isolate
pytest test_suite/integration/test_backend_gemini.py::test_agent_web_search_step -v
```

---

### Common Issues

| Issue | Solution |
|-------|----------|
| `ModuleNotFoundError: No module named 'httpx'` | `pip install httpx websockets pytest-asyncio` |
| `.env.test not found` | `cp .env.test.example .env.test` |
| `Backend server failed to start` | Check port 8000 is available: `lsof -i :8000` |
| `All tests skipped` | Check environment variables are set |

---

## Cost Management

### Expected API Costs (Gemini)

| Test Category | Cost per Run | Notes |
|---------------|--------------|-------|
| Backend + Gemini | ~$0.01-0.03 | Flash model |
| DarCI | ~$0.02-0.05 | Multi-agent coordination |
| E2E | ~$0.05-0.10 | Full workflows |
| **Total** | **~$0.10-0.20** | Per full test run |

### Reduce Costs

```bash
# Skip costly tests during development
pytest test_suite/integration/ -m "not api_cost"

# Run only unit tests (no API calls)
pytest test_suite/tailbridge_test/mock/ -v

# Use cheaper model
export GEMINI_FLASH_MODEL=gemini-2.0-flash
```

---

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Integration Tests

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  integration-tests:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Python
      uses: actions/setup-python@v4
      with:
        python-version: '3.11'
    
    - name: Install dependencies
      run: |
        pip install -r requirements.txt
        pip install pytest pytest-asyncio pytest-cov pytest-timeout httpx websockets
    
    - name: Install Tailscale
      run: |
        curl -fsSL https://tailscale.com/install.sh | sh
    
    - name: Run integration tests
      env:
        GEMINI_API_KEY: ${{ secrets.GEMINI_API_KEY }}
        TS_AUTH_KEY_1: ${{ secrets.TS_AUTH_KEY_1 }}
        TS_AUTH_KEY_2: ${{ secrets.TS_AUTH_KEY_2 }}
      run: |
        ./test_suite/run-integration-tests.sh -m gemini --coverage
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
```

---

## Writing New Tests

### Test Template

```python
"""
Test module description
"""

import pytest
import os

GEMINI_API_KEY = os.getenv("GEMINI_API_KEY")

@pytest.mark.gemini
@pytest.mark.timeout(60)
async def test_your_feature(gemini_flash_model):
    """Test description."""
    # Arrange
    goal = "Your test goal"
    
    # Act
    result = await your_function(gemini_flash_model, goal)
    
    # Assert
    assert result is not None
    print(f"✓ Test passed: {result}")
```

### Best Practices

1. **Use fixtures:** Leverage `gemini_flash_model`, `tailscale_agents` fixtures
2. **Add timeouts:** Always use `@pytest.mark.timeout()`
3. **Handle skips:** Use `pytest.skip()` for missing dependencies
4. **Print progress:** Use `print()` with ✓/⚠/✗ symbols
5. **Clean up:** Use fixtures with `yield` for automatic cleanup

---

## Test Statistics

### Typical Test Run

```
Test Suite              Tests   Time     Cost
-------------------------------------------------
Backend + Gemini         25     3-5 min  $0.02
Tailbridge + Tailscale   20     5-8 min  $0.00
DarCI                    15     4-6 min  $0.03
E2E                      15     8-12 min $0.05
-------------------------------------------------
Total                    75     20-31 min $0.10
```

---

## Support

- **Documentation:** See `test_suite/integration/README.md`
- **Issues:** File on GitHub
- **API Keys:**
  - Gemini: https://makersuite.google.com/app/apikey
  - Tailscale: https://login.tailscale.com/admin/settings/keys

---

*Last updated: March 8, 2026*
*Version: 1.0.0*
