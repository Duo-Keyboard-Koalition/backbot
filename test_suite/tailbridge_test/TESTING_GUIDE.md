# Tailbridge Test Platform - Complete Guide

## Overview

The Tailbridge Test Platform provides comprehensive testing for A2A (Agent-to-Agent) communication and file transfer functionality across isolated environments with Tailscale connectivity.

## Quick Start

### 1. Run Mock Tests (Fastest - No Dependencies)

```bash
cd test_platform

# Windows
go test ./mock/... -v

# Linux/Mac
go test ./mock/... -v
```

### 2. Run Integration Tests (Requires Docker + Tailscale Auth Keys)

```bash
# Create .env file with Tailscale auth keys
cd test_platform/docker
cat > .env << EOF
TS_AUTH_KEY_1=tskey-auth-xxxxx1
TS_AUTH_KEY_2=tskey-auth-xxxxx2
TS_AUTH_KEY_3=tskey-auth-xxxxx3
EOF

# Run tests
cd ..
.\scripts\run-all-tests.ps1 -integration  # Windows
./scripts/run-all-tests.sh -integration   # Linux/Mac
```

### 3. Run All Tests

```bash
.\scripts\run-all-tests.ps1 -all  # Windows
./scripts/run-all-tests.sh -all   # Linux/Mac
```

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Test Platform                             │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────────┐         ┌──────────────────┐         │
│  │   Mock Tests     │         │ Integration      │         │
│  │   (In-Memory)    │         │ Tests (Docker)   │         │
│  │                  │         │                  │         │
│  │  • Fast          │         │  • Real Tailscale│         │
│  │  • No deps       │         │  • End-to-end    │         │
│  │  • Deterministic │         │  • Network tests │         │
│  └──────────────────┘         └──────────────────┘         │
│                                                              │
│  ┌──────────────────────────────────────────────────┐       │
│  │           Test Orchestrator                       │       │
│  │  • Coordinates test execution                     │       │
│  │  • Manages Docker containers                      │       │
│  │  • Generates reports                              │       │
│  └──────────────────────────────────────────────────┘       │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Directory Structure

```
test_platform/
├── mock/                          # Mock testing framework
│   ├── mock_network.go            # Simulated Tailscale network
│   └── testify/                   # Test suites
│       ├── a2a_test.go            # A2A communication tests
│       └── filetransfer_test.go   # File transfer tests
│
├── docker/                        # Docker-based testing
│   ├── docker-compose.test.yml    # Multi-agent setup
│   ├── Dockerfile.agent           # Test agent image
│   └── .env.example               # Environment template
│
├── integration/                   # Integration tests
│   ├── a2a/
│   │   └── a2a_integration_test.go
│   └── filetransfer/
│       └── filetransfer_integration_test.go
│
├── cmd/
│   ├── test-agent/                # Dummy agent for Docker
│   │   └── main.go
│   └── test-orchestrator/         # Test automation
│       └── main.go
│
├── scripts/
│   ├── run-all-tests.ps1          # Windows test runner
│   └── run-all-tests.sh           # Unix test runner
│
├── go.mod                         # Go module definition
└── README.md                      # This file
```

---

## Mock Tests

### What They Test

Mock tests simulate the entire Tailscale network in memory without requiring:
- Real Tailscale connection
- Docker containers
- Network configuration

### Test Coverage

#### A2A Communication Tests
- ✅ Agent discovery (search, filter by capability)
- ✅ Basic messaging (send, receive)
- ✅ Bidirectional messaging
- ✅ Topic-based routing
- ✅ Consumer groups
- ✅ Message correlation (request/response)
- ✅ Network latency simulation
- ✅ Broadcast messaging
- ✅ Phone book functionality
- ✅ Agent lifecycle (start/stop)
- ✅ Concurrent messaging
- ✅ Error handling

#### File Transfer Tests
- ✅ Small file transfer (< 1MB)
- ✅ Large file transfer (100MB+)
- ✅ Progress tracking
- ✅ Concurrent transfers
- ✅ Transfer notifications
- ✅ File integrity verification
- ✅ Offline agent handling
- ✅ Multiple receivers
- ✅ Network latency simulation
- ✅ Edge cases (tiny files)

### Running Mock Tests

```bash
# All mock tests
go test ./mock/... -v

# Specific test suite
go test ./mock/testify -run TestA2AMessaging -v
go test ./mock/testify -run TestFileTransferSuite -v

# With coverage
go test ./mock/... -v -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## Integration Tests

### What They Test

Integration tests run real agents in Docker containers connected via Tailscale:
- Real Tailscale connectivity
- End-to-end A2A communication
- End-to-end file transfer
- Network conditions
- Multi-agent scenarios

### Prerequisites

1. **Docker Desktop** or **Docker Engine** + **docker-compose**
2. **Tailscale Account** with ability to generate auth keys
3. **3 Auth Keys** (one per agent)

### Setup

#### Step 1: Generate Tailscale Auth Keys

Go to [Tailscale Admin Panel](https://login.tailscale.com/admin/settings/keys)

Create 3 pre-authentication keys:
- Key 1: Agent 1 (tag: agent1)
- Key 2: Agent 2 (tag: agent2)
- Key 3: Agent 3 (tag: agent3)

#### Step 2: Create Environment File

```bash
cd test_platform/docker
cat > .env << EOF
TS_AUTH_KEY_1=tskey-auth-key-1-xxxxx
TS_AUTH_KEY_2=tskey-auth-key-2-xxxxx
TS_AUTH_KEY_3=tskey-auth-key-3-xxxxx
TAILNET_NAME=your-tailnet.ts.net
EOF
```

#### Step 3: Run Integration Tests

```bash
# Using test runner script
.\scripts\run-all-tests.ps1 -integration  # Windows
./scripts/run-all-tests.sh -integration   # Linux/Mac

# Manual execution
docker-compose -f docker/docker-compose.test.yml up -d

# Wait for agents (check logs)
docker-compose -f docker/docker-compose.test.yml logs -f

# Run tests
go test ./integration/... -v -tags=integration

# Cleanup
docker-compose -f docker/docker-compose.test.yml down
```

### Test Coverage

#### A2A Integration Tests
- ✅ Agent health checks
- ✅ Agent status endpoints
- ✅ Phone book discovery
- ✅ Agent capability filtering
- ✅ Send/receive messages
- ✅ Bidirectional messaging
- ✅ Message correlation
- ✅ Concurrent messaging
- ✅ Different message types
- ✅ Large messages
- ✅ Error handling

#### File Transfer Integration Tests
- ✅ Small file transfer (100 KB)
- ✅ Large file transfer (10 MB)
- ✅ Progress tracking
- ✅ Concurrent transfers
- ✅ Compression
- ✅ Multiple receivers
- ✅ Transfer history
- ✅ File integrity
- ✅ Edge cases

---

## Test Orchestrator

The test orchestrator automates test execution:

```bash
# Run mock tests only
go run ./cmd/test-orchestrator -mode=mock -v

# Run Docker integration tests only
go run ./cmd/test-orchestrator -mode=docker -v

# Run all tests
go run ./cmd/test-orchestrator -mode=all -v -coverage

# Custom agent count
go run ./cmd/test-orchestrator -mode=docker -agents=5
```

### Options

| Flag | Description | Default |
|------|-------------|---------|
| `-mode, -m` | Test mode: mock, docker, all | mock |
| `-v, -verbose` | Verbose output | false |
| `-coverage` | Generate coverage report | false |
| `-timeout` | Test timeout | 10m |
| `-agents` | Number of agents | 3 |
| `-cleanup` | Cleanup on exit | true |

---

## Writing New Tests

### Adding Mock Tests

```go
package testify

import (
    "testing"
    "github.com/codejedi-ai/kaggle-for-tensors/tailscale-app/tailbridge/test_platform/mock"
    "github.com/stretchr/testify/suite"
)

type MyTestSuite struct {
    suite.Suite
    network *mock.MockNetwork
}

func (s *MyTestSuite) SetupTest() {
    s.network = mock.NewMockNetwork()
    // Setup agents
}

func (s *MyTestSuite) TestMyFeature() {
    // Your test logic
}

func TestMySuite(t *testing.T) {
    suite.Run(t, new(MyTestSuite))
}
```

### Adding Integration Tests

```go
//go:build integration

package mytest

import (
    "testing"
    "github.com/stretchr/testify/suite"
)

type MyIntegrationSuite struct {
    suite.Suite
    agent1URL string
}

func (s *MyIntegrationSuite) SetupSuite() {
    s.agent1URL = getEnv("AGENT1_URL", "http://localhost:8081")
}

func (s *MyIntegrationSuite) TestMyFeature() {
    // Your integration test
}

func TestMyIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    suite.Run(t, new(MyIntegrationSuite))
}
```

---

## Troubleshooting

### Mock Tests Fail

```bash
# Clear Go cache
go clean -testcache

# Update dependencies
go mod tidy

# Run with verbose output
go test ./mock/... -v
```

### Docker Tests Fail

```bash
# Check Docker is running
docker ps

# Check docker-compose
docker-compose --version

# View agent logs
docker-compose -f docker/docker-compose.test.yml logs agent1

# Check network
docker network inspect test_platform_default

# Restart containers
docker-compose -f docker/docker-compose.test.yml down
docker-compose -f docker/docker-compose.test.yml up -d
```

### Tailscale Connection Issues

```bash
# Verify auth keys are valid
# Check Tailscale admin panel

# Check tailnet name matches
echo $TAILNET_NAME

# View Tailscale status in container
docker exec tailbridge-agent1 tailscale status
```

---

## CI/CD Integration

### GitHub Actions

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.25'
    
    - name: Run Mock Tests
      run: |
        cd test_platform
        go test ./mock/... -v -coverprofile=coverage.out
    
    - name: Upload Coverage
      uses: codecov/codecov-action@v3
      with:
        files: ./test_platform/coverage.out
    
    - name: Run Integration Tests
      env:
        TS_AUTH_KEY_1: ${{ secrets.TS_AUTH_KEY_1 }}
        TS_AUTH_KEY_2: ${{ secrets.TS_AUTH_KEY_2 }}
        TS_AUTH_KEY_3: ${{ secrets.TS_AUTH_KEY_3 }}
      run: |
        cd test_platform
        ./scripts/run-all-tests.sh -integration
```

---

## Performance Benchmarks

### Mock Tests

```bash
# Run with benchmark
go test ./mock/... -bench=. -benchmem

# Typical results:
# TestA2AMessaging: ~100ms
# TestFileTransferSuite: ~200ms
# Total mock suite: ~500ms
```

### Integration Tests

```bash
# Run with timing
time go test ./integration/... -v -tags=integration

# Typical results:
# Agent startup: 30-60s
# A2A tests: 2-3 minutes
# File transfer tests: 3-5 minutes
# Total: 5-10 minutes
```

---

## Best Practices

### 1. Write Tests in Both Modes

- **Mock tests** for fast feedback during development
- **Integration tests** for validation before release

### 2. Use Appropriate Timeouts

```go
// Mock tests: milliseconds
time.Sleep(50 * time.Millisecond)

// Integration tests: seconds
time.Sleep(2 * time.Second)
```

### 3. Clean Up Resources

```go
func (s *MySuite) TearDownTest() {
    // Clean up files, connections, etc.
}
```

### 4. Use Test Helpers

```go
func (s *MySuite) waitForAgent(url string) {
    s.Eventually(func() bool {
        resp, _ := http.Get(url + "/health")
        return resp != nil && resp.StatusCode == 200
    }, 2*time.Minute, 5*time.Second)
}
```

### 5. Parallelize When Possible

```go
func TestParallel(t *testing.T) {
    t.Parallel()
    // Test logic
}
```

---

## Contributing

1. Write tests for new features
2. Ensure all existing tests pass
3. Update documentation
4. Submit PR with test results

---

## License

Same as the parent project.

---

*Last updated: March 7, 2026*
*Version: 1.0.0*
