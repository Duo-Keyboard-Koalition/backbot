# Shared Agent Logging

All three agents share a common log directory for centralized monitoring and debugging.

## Log Directory Structure

```
test_suite/integration/logs/
├── test_run.log                    # Main test run log (all agents)
├── agent_test_name.log            # Per-test agent logs
└── ...
```

## Configuration

The shared logging is configured in `conftest.py`:

```python
# Shared log directory for all agents
LOG_DIR = Path(__file__).parent / "logs"
LOG_DIR.mkdir(exist_ok=True)
```

## Usage

### In Tests

```python
@pytest.mark.timeout(TEST_TIMEOUT)
async def test_example(self, agent_logger, shared_log_dir):
    """Test with shared logging."""
    # Log from Agent 1's perspective
    agent_logger.info("[AGENT 1] Initializing...")
    
    # Log from Agent 2's perspective
    agent_logger.info("[AGENT 2] Ready...")
    
    # Log from Agent 3's perspective
    agent_logger.info("[AGENT 3] Monitoring...")
    
    # Inter-agent communication
    agent_logger.info("[AGENT 1 -> AGENT 2] Sending message...")
    agent_logger.info("[AGENT 2] Message received")
    
    # All logs go to the same shared directory
    print(f"Logs directory: {shared_log_dir}")
```

### Log Format

```
2026-03-08 05:23:08,816 - agent.test_name - INFO - [AGENT 1] Initializing connection to tailnet...
2026-03-08 05:23:08,817 - agent.test_name - INFO - [AGENT 2] Ready to receive messages...
2026-03-08 05:23:08,817 - agent.test_name - INFO - [AGENT 3] Monitoring agent status...
```

## Fixtures

### `shared_log_dir`
Returns the Path to the shared log directory.

### `agent_logger`
Returns a logger configured for the current test. All agents share this logger.

### `shared_log_file`
Returns the path to the main test run log file.

## Helper Functions

```python
# Get all agent logs as a dictionary
logs = get_all_agent_logs(LOG_DIR)
for agent_name, log_content in logs.items():
    print(f"=== {agent_name} ===")
    print(log_content)

# Clear all agent logs
clear_agent_logs(LOG_DIR)
```

## Benefits

1. **Centralized Monitoring**: All agent activity in one place
2. **Easy Debugging**: See inter-agent communication flow
3. **Test Isolation**: Each test gets its own log file
4. **Shared Context**: All agents share the main test_run.log

## Example Output

```
============================================================
SHARED LOG VERIFICATION
All three agents log to: /mnt/c/Users/darcy/repos/sentinelai/test_suite/integration/logs
This allows centralized monitoring of all agent activity
============================================================
```

## Viewing Logs

```bash
# View main test log
tail -f test_suite/integration/logs/test_run.log

# View all agent logs
cat test_suite/integration/logs/agent_*.log

# View specific agent log
cat test_suite/integration/logs/agent_test_name.log

# List all logs
ls -la test_suite/integration/logs/
```

## Integration with Tailbridge

When running Tailbridge agents, configure each agent to log to the shared directory:

```bash
# Agent 1
TAILBRIDGE_LOG_DIR=test_suite/integration/logs \
TAILBRIDGE_AGENT_NAME=agent1 \
go run ./tailbridge/taila2a/bridge

# Agent 2
TAILBRIDGE_LOG_DIR=test_suite/integration/logs \
TAILBRIDGE_AGENT_NAME=agent2 \
go run ./tailbridge/taila2a/bridge

# Agent 3
TAILBRIDGE_LOG_DIR=test_suite/integration/logs \
TAILBRIDGE_AGENT_NAME=agent3 \
go run ./tailbridge/taila2a/bridge
```

All three agents will write to the same log directory, allowing you to see:
- When each agent starts/stops
- Inter-agent message flow
- A2A communication patterns
- File transfer events
- Network connectivity issues

---

*Last updated: March 8, 2026*
