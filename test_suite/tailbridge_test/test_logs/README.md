# Test Logs

This directory contains structured test execution logs with unique test IDs.

## Directory Structure

```
test_logs/
├── NetworkDiscovery/
│   ├── test-NetworkDiscovery-<timestamp>.json    # Individual test log
│   ├── test-NetworkDiscovery-<timestamp>.json
│   ├── ...
│   └── test_report.json                          # Suite summary report
├── A2AMessaging/
│   ├── test-A2AMessaging-<timestamp>.json
│   └── test_report.json
└── FileTransfer/
    ├── test-FileTransfer-<timestamp>.json
    └── test_report.json
```

## Log File Format

Each test log is a JSON file containing:

```json
{
  "test_id": "test-NetworkDiscovery-1234567890",
  "test_name": "NetworkDiscovery",
  "start_time": "2026-03-07T12:00:00Z",
  "end_time": "2026-03-07T12:00:02Z",
  "duration": "2.39s",
  "log_entries": [
    {
      "timestamp": "2026-03-07T12:00:00.000Z",
      "level": "INFO",
      "category": "NETWORK",
      "message": "Created new mock network",
      "details": {
        "network_id": "abc12345"
      }
    },
    {
      "timestamp": "2026-03-07T12:00:00.001Z",
      "level": "INFO",
      "category": "AGENT_LIFECYCLE",
      "message": "Agent agent-1 spinning up",
      "details": {
        "name": "agent-1",
        "state": "running",
        "tailscale_ip": "100.64.0.1",
        "tailscale_ipv6": "fd7a:115c:a1e0::1",
        "capabilities": ["chat", "file_send"],
        "inbound_port": 8001,
        "http_port": 8080,
        "node_id": "abc1234567890"
      }
    }
  ]
}
```

## Agent IP Assignment

Each agent receives unique Tailscale-like IP addresses:

| Agent | IPv4 | IPv6 | Ports |
|-------|------|------|-------|
| agent-1 | 100.64.0.1 | fd7a:115c:a1e0::1 | 8080, 8001 |
| agent-2 | 100.64.0.2 | fd7a:115c:a1e0::2 | 8080, 8001 |
| agent-3 | 100.64.0.3 | fd7a:115c:a1e0::3 | 8080, 8001 |
| ... | ... | ... | ... |

IP ranges:
- **IPv4**: 100.64.0.1 - 100.64.0.254 (CGNAT range, same as Tailscale)
- **IPv6**: fd7a:115c:a1e0::1 - ::254 (Tailscale ULA range)

## Log Categories

| Category | Description |
|----------|-------------|
| `TEST_LIFECYCLE` | Test start/end events |
| `NETWORK` | Network creation/state |
| `AGENT_LIFECYCLE` | Agent spin up/tear down |
| `AGENT` | Agent-specific events |
| `DISCOVERY` | Discovery events (join/leave) |
| `MESSAGE` | Message send/receive |
| `FILE_TRANSFER` | File transfer events |
| `ERROR` | Error conditions |

## Viewing Logs

### PowerShell

```powershell
# View latest test log
Get-ChildItem test_logs\NetworkDiscovery\*.json | Sort-Object LastWriteTime -Descending | Select-Object -First 1 | Get-Content | ConvertFrom-Json

# View test report
Get-Content test_logs\NetworkDiscovery\test_report.json | ConvertFrom-Json

# View all test IDs
Get-ChildItem test_logs\**\*.json -Exclude test_report.json | ForEach-Object { 
    $json = Get-Content $_.FullName | ConvertFrom-Json
    [PSCustomObject]@{
        Test = $json.test_name
        TestID = $json.test_id
        Duration = $json.duration
        Timestamp = $json.start_time
    }
}
```

### Bash (Linux/Mac)

```bash
# View latest test log (requires jq)
ls -t test_logs/NetworkDiscovery/*.json | head -1 | xargs cat | jq

# View test report
cat test_logs/NetworkDiscovery/test_report.json | jq

# View all test IDs
for f in test_logs/*/*.json; do
    [[ "$f" == *"test_report"* ]] && continue
    test=$(cat "$f" | jq -r '.test_name')
    id=$(cat "$f" | jq -r '.test_id')
    duration=$(cat "$f" | jq -r '.duration')
    echo "$test | $id | $duration"
done
```

### Go

```go
package main

import (
    "encoding/json"
    "os"
    "fmt"
)

type TestLog struct {
    TestID   string `json:"test_id"`
    TestName string `json:"test_name"`
    Duration string `json:"duration"`
}

func main() {
    data, _ := os.ReadFile("test_logs/NetworkDiscovery/test-xxx.json")
    var log TestLog
    json.Unmarshal(data, &log)
    fmt.Printf("Test: %s, ID: %s, Duration: %s\n", log.TestName, log.TestID, log.Duration)
}
```

## Test Report Format

The `test_report.json` contains suite-level summary:

```json
{
  "suite_name": "NetworkDiscovery",
  "generated_at": "2026-03-07T12:05:00Z",
  "total_tests": 12,
  "passed": 12,
  "failed": 0,
  "total_duration": "1.13s",
  "tests": [
    {
      "test_id": "test-NetworkDiscovery-1234567890",
      "duration": "0.05s",
      "passed": true
    },
    ...
  ]
}
```

## Searching Logs

### Find all tests with specific agent

```powershell
# PowerShell
Get-ChildItem test_logs\**\*.json | ForEach-Object {
    $json = Get-Content $_.FullName | ConvertFrom-Json
    if ($json.log_entries.details.name -contains "agent-1") {
        $_.FullName
    }
}
```

### Find tests by IP address

```powershell
# PowerShell
Get-ChildItem test_logs\**\*.json | ForEach-Object {
    $json = Get-Content $_.FullName | ConvertFrom-Json
    if ($json.log_entries.details.tailscale_ip -contains "100.64.0.1") {
        $_.FullName
    }
}
```

### Find tests by duration

```bash
# Bash with jq
for f in test_logs/*/*.json; do
    [[ "$f" == *"test_report"* ]] && continue
    duration=$(cat "$f" | jq -r '.duration' | sed 's/s$//')
    if (( $(echo "$duration > 0.5" | bc -l) )); then
        echo "$f: $(cat "$f" | jq -r '.duration')"
    fi
done
```

## Log Retention

Logs are appended on each test run. To clean old logs:

```bash
# Delete logs older than 7 days
find test_logs -name "*.json" -mtime +7 -delete

# Delete all logs
rm -rf test_logs/*
```

## Integration with CI/CD

Add to your CI pipeline to preserve test logs:

```yaml
# GitHub Actions example
- name: Run Tests
  run: go test ./mock/... -v

- name: Upload Test Logs
  uses: actions/upload-artifact@v3
  with:
    name: test-logs
    path: test_platform/test_logs/
    retention-days: 30
```

---

*Last updated: March 7, 2026*
