# Tailbridge Test Platform

Comprehensive testing platform for Tailbridge A2A communication and file transfer functionality.

## Overview

This test platform provides:

1. **Mock Testing Framework** - Fast unit tests without Tailscale connection
2. **Docker-Based Integration Tests** - Real Tailscale testing with isolated agents
3. **Test Orchestrator** - Automated test execution and reporting

## Quick Start

### Run Mock Tests (Fast, No Dependencies)

```bash
cd test_platform
go test ./mock/... -v
```

### Run Docker Integration Tests

```bash
# Start test environment
docker-compose -f docker/docker-compose.test.yml up -d

# Run integration tests
go test ./integration/... -v -tags=integration

# Stop test environment
docker-compose -f docker/docker-compose.test.yml down
```

### Run All Tests

```bash
# Use the test runner script
.\scripts\run-all-tests.ps1  # Windows
./scripts/run-all-tests.sh   # Linux/Mac
```

## Directory Structure

```
test_platform/
├── mock/                      # Mock testing framework
│   ├── mock_network.go        # Simulated Tailscale network
│   ├── mock_agent.go          # Dummy agent implementation
│   └── testify/               # Test suites
│       ├── a2a_test.go        # A2A communication tests
│       └── filetransfer_test.go # File transfer tests
│
├── docker/                    # Docker-based testing
│   ├── docker-compose.test.yml
│   ├── Dockerfile.agent
│   └── agents/
│       ├── agent1/            # Agent 1 config
│       ├── agent2/            # Agent 2 config
│       └── agent3/            # Agent 3 config
│
├── integration/               # Integration test suites
│   ├── a2a/
│   │   └── a2a_integration_test.go
│   └── filetransfer/
│       └── filetransfer_integration_test.go
│
├── cmd/
│   └── test-orchestrator/     # Test automation
│       └── main.go
│
└── scripts/
    ├── run-all-tests.ps1      # Windows test runner
    └── run-all-tests.sh       # Unix test runner
```

## Test Categories

### Unit Tests (Mock)
- ✅ A2A message routing
- ✅ Phone book discovery
- ✅ Consumer groups
- ✅ File chunking logic
- ✅ Progress tracking
- ✅ Buffer-triggered agents

### Integration Tests (Docker)
- ✅ Real Tailscale connectivity
- ✅ End-to-end A2A communication
- ✅ End-to-end file transfer
- ✅ Multi-agent scenarios
- ✅ Network failure recovery

### Performance Tests
- ⏳ Large file transfers (100MB+)
- ⏳ Concurrent transfers
- ⏳ Message throughput

## Configuration

### Mock Tests

No configuration needed - runs entirely in memory.

### Docker Tests

Create `.env` file in `test_platform/docker/`:

```env
TS_AUTH_KEY_1=tskey-auth-key-1
TS_AUTH_KEY_2=tskey-auth-key-2
TS_AUTH_KEY_3=tskey-auth-key-3
TAILNET_NAME=example.com
```

## Writing New Tests

### Mock Test Example

```go
func TestA2AMessaging(t *testing.T) {
    // Create mock network
    net := mock.NewMockNetwork()
    
    // Add agents
    agent1 := net.AddAgent("agent-alpha")
    agent2 := net.AddAgent("agent-beta")
    
    // Send message
    msg := protocol.NewMessage("request", "agent.requests", "Hello!")
    err := agent1.Send(agent2.Name, msg)
    
    // Verify
    assert.NoError(t, err)
    assert.Equal(t, "Hello!", agent2.LastMessage().Body.Payload)
}
```

### Integration Test Example

```go
//go:build integration

func TestRealA2ACommunication(t *testing.T) {
    // Connect to real agents
    agent1 := connectToAgent("http://agent1:8080")
    agent2 := connectToAgent("http://agent2:8080")
    
    // Send message via Tailscale
    err := agent1.SendA2A(agent2.TailscaleIP, "Hello via Tailscale!")
    assert.NoError(t, err)
    
    // Verify received
    msg := agent2.WaitForMessage(5 * time.Second)
    assert.Equal(t, "Hello via Tailscale!", msg.Payload)
}
```

## Troubleshooting

### Docker Tests Fail

1. Check Tailscale auth keys are valid
2. Ensure Docker daemon is running
3. Verify network connectivity: `docker network inspect test_platform_default`

### Mock Tests Fail

1. Update dependencies: `go mod tidy`
2. Clear cache: `go clean -testcache`

## CI/CD Integration

Add to GitHub Actions:

```yaml
test:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    
    - name: Run Mock Tests
      run: cd test_platform && go test ./mock/... -v
    
    - name: Run Integration Tests
      env:
        TS_AUTH_KEY_1: ${{ secrets.TS_AUTH_KEY_1 }}
        TS_AUTH_KEY_2: ${{ secrets.TS_AUTH_KEY_2 }}
        TS_AUTH_KEY_3: ${{ secrets.TS_AUTH_KEY_3 }}
      run: cd test_platform && go test ./integration/... -v -tags=integration
```

## License

Same as the parent project.
