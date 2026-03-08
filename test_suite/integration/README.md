# SentinelAI Integration Tests

**Real API integration tests - NO MOCKS**

## Prerequisites

### 1. API Keys Required

```bash
# Copy and configure .env.test
cp .env.test.example .env.test
```

**Edit `.env.test` with your real API keys:**

```bash
# Gemini API (get from https://makersuite.google.com/app/apikey)
GEMINI_API_KEY=AIzaSy...
GEMINI_FLASH_MODEL=gemini-2.0-flash
GEMINI_PRO_MODEL=gemini-2.0-pro

# Tailscale (get from https://login.tailscale.com/admin/settings/keys)
TS_AUTH_KEY_1=tskey-auth-xxxxx1
TS_AUTH_KEY_2=tskey-auth-xxxxx2
TS_AUTH_KEY_3=tskey-auth-xxxxx3
TAILNET_NAME=your-tailnet.ts.net

# Test configuration
TEST_TIMEOUT=300
TEST_RETRY_COUNT=3
```

### 2. Install Dependencies

```bash
# Python test dependencies
pip install pytest pytest-asyncio pytest-cov pytest-timeout httpx websockets

# Tailscale must be installed and running
tailscale status  # Verify you're authenticated
```

### 3. Tailscale Setup

```bash
# Install Tailscale (if not already)
# Ubuntu/WSL:
curl -fsSL https://tailscale.com/install.sh | sh

# Authenticate
tailscale up

# Verify connection
tailscale status
```

## Running Tests

```bash
# All integration tests
pytest test_suite/integration/ -v

# Specific test categories
pytest test_suite/integration/test_backend_gemini.py -v
pytest test_suite/integration/test_tailbridge.py -v
pytest test_suite/integration/test_darci.py -v
pytest test_suite/integration/test_e2e.py -v

# With markers
pytest test_suite/integration/ -m gemini -v
pytest test_suite/integration/ -m tailscale -v
pytest test_suite/integration/ -m e2e -v

# With coverage
pytest test_suite/integration/ --cov=backend --cov=darci --cov-report=html

# Parallel execution (faster)
pytest test_suite/integration/ -n auto

# Stop on first failure
pytest test_suite/integration/ -x

# Verbose output with print statements
pytest test_suite/integration/ -v -s
```

## Test Markers

| Marker | Description | Example |
|--------|-------------|---------|
| `@pytest.mark.gemini` | Requires Gemini API | Agent reasoning tests |
| `@pytest.mark.tailscale` | Requires Tailscale connection | A2A communication tests |
| `@pytest.mark.e2e` | Full end-to-end workflow | Complete pipeline tests |
| `@pytest.mark.slow` | Takes >30 seconds | Multi-agent tests |
| `@pytest.mark.api_cost` | Costs API quota | Gemini generation tests |

## Skipping Tests

```bash
# Skip slow tests
pytest test_suite/integration/ -m "not slow"

# Skip tests that cost API quota
pytest test_suite/integration/ -m "not api_cost"

# Skip Tailscale tests (no connection)
pytest test_suite/integration/ -m "not tailscale"
```

## Troubleshooting

### Gemini API Errors

```bash
# Check API key is valid
curl "https://generativelanguage.googleapis.com/v1beta/models?key=YOUR_API_KEY"

# Check quota
# Visit: https://console.cloud.google.com/apis/api/generativelanguage.googleapis.com/quotas
```

### Tailscale Errors

```bash
# Check Tailscale status
tailscale status

# Re-authenticate if needed
tailscale logout
tailscale up

# Check firewall
sudo ufw status
```

### Test Hangs

```bash
# Run with timeout
pytest test_suite/integration/ --timeout=60

# Increase timeout in .env.test
TEST_TIMEOUT=300
```

## Cost Management

Expected costs per full test run (Gemini API):
- **Flash model**: ~$0.01-0.05 per run
- **Pro model**: ~$0.05-0.20 per run
- **Monthly budget**: ~$5-10 for active development

To reduce costs:
1. Use `gemini-2.0-flash` for most tests
2. Run `-m "not api_cost"` for local development
3. Use mocks for unit tests (in `test_suite/tailbridge_test/mock/`)

---

*Last updated: March 8, 2026*
