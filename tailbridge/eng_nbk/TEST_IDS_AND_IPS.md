# Test Platform - Agent IP Assignment & Test Logs

**Date:** March 7, 2026  
**Test Suite:** NetworkDiscovery  
**Status:** Complete ✅

---

## Overview

This document contains test logs with unique test IDs showing agent IP assignment and lifecycle management across isolated container-like test environments.

---

## Test ID Structure

Each test run generates a unique test ID:

```
test-{suite_name}-{timestamp_nanoseconds}

Example: test-NetworkDiscovery-1772941430619493300
```

### Log File Location

```
test_platform/
└── test_logs/
    └── {suite_name}/
        ├── {test_id}.json           # Individual test log
        └── test_report.json         # Suite summary
```

---

## Agent IP Assignment

### Tailscale CGNAT Range

Each agent in the mock network receives a unique IP pair:

| Agent Name | IPv4 | IPv6 | Node ID | Ports |
|------------|------|------|---------|-------|
| discover-agent-1 | 100.64.0.1 | fd7a:115c:a1e0::1 | f135425f-037b-46 | 8080, 8001 |
| discover-agent-2 | 100.64.0.2 | fd7a:115c:a1e0::2 | 9ed72286-bc6c-47 | 8080, 8001 |
| discover-agent-3 | 100.64.0.3 | fd7a:115c:a1e0::3 | 12afd0b8-284f-4d | 8080, 8001 |
| network-agent-1 | 100.64.0.1 | fd7a:115c:a1e0::1 | varies | 8080, 8001 |
| network-agent-2 | 100.64.0.2 | fd7a:115c:a1e0::2 | varies | 8080, 8001 |
| network-agent-3 | 100.64.0.3 | fd7a:115c:a1e0::3 | varies | 8080, 8001 |
| network-agent-4 | 100.64.0.4 | fd7a:115c:a1e0::4 | varies | 8080, 8001 |
| network-agent-5 | 100.64.0.5 | fd7a:115c:a1e0::5 | varies | 8080, 8001 |

### IP Allocation Strategy

```go
// IPAllocator allocates Tailscale-like IPs for mock agents
type IPAllocator struct {
    nextOctet   byte        // Starts at 1, increments per agent
    baseIP      string      // "100.64.0" (CGNAT)
    baseIPv6    string      // "fd7a:115c:a1e0::" (Tailscale ULA)
}

// Each AddAgent() call:
agent.TailscaleIP = fmt.Sprintf("100.64.0.%d", nextOctet)
agent.TailscaleIPv6 = fmt.Sprintf("fd7a:115c:a1e0::%d", nextOctet)
nextOctet++
```

### Network Isolation

Each test gets its own isolated network:

```
Test 1: Network ID: 87a9133c
  ├─ agent-1 (100.64.0.1)
  ├─ agent-2 (100.64.0.2)
  └─ agent-3 (100.64.0.3)

Test 2: Network ID: a4b2c91d  (fresh IP allocation)
  ├─ agent-1 (100.64.0.1)  ← IPs reset per test
  ├─ agent-2 (100.64.0.2)
  └─ agent-3 (100.64.0.3)
```

---

## Test Log: test-NetworkDiscovery-1772941430619493300

### Test Metadata

```json
{
  "test_id": "test-NetworkDiscovery-1772941430619493300",
  "test_name": "NetworkDiscovery",
  "start_time": "2026-03-07T22:43:50.6194933-05:00",
  "end_time": "2026-03-07T22:43:51.7573993-05:00",
  "duration": "1.1379192s"
}
```

### Agent Spin-Up Sequence

#### Agent 1: discover-agent-1

```json
{
  "timestamp": "2026-03-07T22:43:50.6200461-05:00",
  "level": "INFO",
  "category": "AGENT_LIFECYCLE",
  "message": "Agent discover-agent-1 spinning up",
  "details": {
    "name": "discover-agent-1",
    "state": "running",
    "tailscale_ip": "100.64.0.1",
    "tailscale_ipv6": "fd7a:115c:a1e0::1",
    "capabilities": ["chat"],
    "inbound_port": 8001,
    "http_port": 8080,
    "node_id": "f135425f-037b-46"
  }
}
```

#### Agent 2: discover-agent-2

```json
{
  "timestamp": "2026-03-07T22:43:50.6200461-05:00",
  "level": "INFO",
  "category": "AGENT_LIFECYCLE",
  "message": "Agent discover-agent-2 spinning up",
  "details": {
    "name": "discover-agent-2",
    "state": "running",
    "tailscale_ip": "100.64.0.2",
    "tailscale_ipv6": "fd7a:115c:a1e0::2",
    "capabilities": ["file_send"],
    "inbound_port": 8001,
    "http_port": 8080,
    "node_id": "9ed72286-bc6c-47"
  }
}
```

#### Agent 3: discover-agent-3

```json
{
  "timestamp": "2026-03-07T22:43:50.6200461-05:00",
  "level": "INFO",
  "category": "AGENT_LIFECYCLE",
  "message": "Agent discover-agent-3 spinning up",
  "details": {
    "name": "discover-agent-3",
    "state": "running",
    "tailscale_ip": "100.64.0.3",
    "tailscale_ipv6": "fd7a:115c:a1e0::3",
    "capabilities": ["file_receive"],
    "inbound_port": 8001,
    "http_port": 8080,
    "node_id": "12afd0b8-284f-4d"
  }
}
```

### Agent Tear-Down Sequence

Each agent is properly torn down at test conclusion:

```json
{
  "timestamp": "2026-03-07T22:43:50.6713228-05:00",
  "level": "INFO",
  "category": "AGENT_LIFECYCLE",
  "message": "Agent discover-agent-1 tearing down",
  "details": {
    "messages_sent": 0,
    "name": "discover-agent-1",
    "state": "stopping",
    "uptime": "51.2773ms"
  }
}
```

### Network State Log

```json
{
  "timestamp": "2026-03-07T22:43:50.6713228-05:00",
  "level": "INFO",
  "category": "NETWORK_STATE",
  "message": "Network state snapshot",
  "details": {
    "network_id": "87a9133c",
    "total_agents": 3,
    "running_agents": 3,
    "message_count": 0
  }
}
```

---

## Test Log: test-NetworkDiscovery-1772941422915198400

### Test Metadata

```json
{
  "test_id": "test-NetworkDiscovery-1772941422915198400",
  "duration": "1.1346081s",
  "start_time": "2026-03-07T22:43:42.9151984-05:00",
  "end_time": "2026-03-07T22:43:44.0498065-05:00"
}
```

### Network-Wide Discovery Test

This test spawned 5 agents with different capabilities:

| Agent | IPv4 | IPv6 | Capabilities |
|-------|------|------|--------------|
| network-agent-1 | 100.64.0.1 | fd7a:115c:a1e0::1 | chat, file_send |
| network-agent-2 | 100.64.0.2 | fd7a:115c:a1e0::2 | chat, file_receive |
| network-agent-3 | 100.64.0.3 | fd7a:115c:a1e0::3 | file_send, file_receive |
| network-agent-4 | 100.64.0.4 | fd7a:115c:a1e0::4 | chat, command |
| network-agent-5 | 100.64.0.5 | fd7a:115c:a1e0::5 | stream, chat |

### Log Entries

```json
{
  "timestamp": "2026-03-07T22:43:43.9170815-05:00",
  "level": "INFO",
  "category": "TEST",
  "message": "Spawning 5 agents for network-wide discovery test"
}
```

Agent spawn logs:

```json
{
  "timestamp": "2026-03-07T22:43:43.9170815-05:00",
  "level": "INFO",
  "category": "AGENT",
  "message": "Spawned agent network-agent-1",
  "details": {
    "tailscale_ip": "100.64.0.1",
    "tailscale_ipv6": "fd7a:115c:a1e0::1",
    "capabilities": ["chat", "file_send"]
  }
}
```

Discovery confirmation:

```json
{
  "timestamp": "2026-03-07T22:43:43.9170815-05:00",
  "level": "INFO",
  "category": "AGENT",
  "message": "Agent network-agent-1 discovered",
  "details": {
    "tailscale_ip": "100.64.0.1",
    "tailscale_ipv6": "fd7a:115c:a1e0::1",
    "state": "running"
  }
}
```

---

## Container Test Environment Mapping

### Docker Compose Agent Configuration

When running integration tests with Docker, each container gets:

```yaml
services:
  agent1:
    container_name: tailbridge-agent1
    hostname: agent1
    environment:
      - AGENT_NAME=agent1
      - TS_AUTHKEY=${TS_AUTH_KEY_1}
    ports:
      - "8081:8080"   # HTTP
      - "8011:8001"   # Inbound A2A
    # Tailscale assigns: 100.64.0.1 (via tailnet)

  agent2:
    container_name: tailbridge-agent2
    hostname: agent2
    environment:
      - AGENT_NAME=agent2
      - TS_AUTHKEY=${TS_AUTH_KEY_2}
    ports:
      - "8082:8080"
      - "8012:8001"
    # Tailscale assigns: 100.64.0.2

  agent3:
    container_name: tailbridge-agent3
    hostname: agent3
    environment:
      - AGENT_NAME=agent3
      - TS_AUTHKEY=${TS_AUTH_KEY_3}
    ports:
      - "8083:8080"
      - "8013:8001"
    # Tailscale assigns: 100.64.0.3
```

### Mock vs Real IP Assignment

| Aspect | Mock Tests | Docker Integration |
|--------|-----------|-------------------|
| IP Source | IPAllocator | Tailscale daemon |
| IPv4 Range | 100.64.0.x | 100.64.0.x (CGNAT) |
| IPv6 Range | fd7a:115c:a1e0::x | fd7a:115c:a1e0::x (ULA) |
| Isolation | Per test | Per container |
| Network ID | UUID (8 chars) | Tailnet name |

---

## Test Execution Timeline

```
Test: test-NetworkDiscovery-1772941430619493300
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

22:43:50.619  Test started
22:43:50.619  Network created (ID: 87a9133c)
22:43:50.620  Agent discover-agent-1 spinning up (IP: 100.64.0.1)
22:43:50.620  Agent discover-agent-2 spinning up (IP: 100.64.0.2)
22:43:50.620  Agent discover-agent-3 spinning up (IP: 100.64.0.3)
22:43:50.671  Agent discover-agent-1 tearing down (uptime: 51ms)
22:43:50.671  Agent discover-agent-2 tearing down (uptime: 51ms)
22:43:50.671  Agent discover-agent-3 tearing down (uptime: 51ms)
22:43:50.671  Network state snapshot (3 agents, 0 messages)
22:43:51.757  Test completed (duration: 1.14s)
```

---

## Viewing Test Logs

### PowerShell

```powershell
cd test_platform

# View latest test log
.\scripts\view-logs.ps1 NetworkDiscovery

# View test report
.\scripts\view-logs.ps1 -Report

# List available suites
.\scripts\view-logs.ps1
```

### Manual

```powershell
# Read latest log
Get-ChildItem test_logs\NetworkDiscovery\*.json | 
    Sort-Object LastWriteTime -Descending | 
    Select-Object -First 1 | 
    Get-Content | 
    ConvertFrom-Json
```

---

## Test ID Reference

| Test Suite | Test ID Pattern | Log Location |
|------------|----------------|--------------|
| NetworkDiscovery | test-NetworkDiscovery-{timestamp} | test_logs/NetworkDiscovery/ |
| A2AMessaging | test-A2AMessaging-{timestamp} | test_logs/A2AMessaging/ |
| FileTransfer | test-FileTransfer-{timestamp} | test_logs/FileTransfer/ |

---

## Agent Lifecycle per Test

```
┌─────────────────────────────────────────────────────────────┐
│  Test: test-NetworkDiscovery-{timestamp}                    │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  SetupTest()                                                │
│    ├─ Create MockNetwork (ID: {uuid})                      │
│    ├─ Create Agents                                        │
│    ├─ AddAgent() → Assigns IP (100.64.0.x)                │
│    └─ WaitForAllAgents()                                   │
│                                                              │
│  <Test Execution>                                          │
│    ├─ Agent Discovery Tests                                │
│    ├─ Messaging Tests                                      │
│    └─ State Verification                                   │
│                                                              │
│  TearDownTest()                                            │
│    ├─ LogAgentTearDown() for each agent                   │
│    ├─ RemoveAgent() from network                           │
│    ├─ LogNetworkState()                                    │
│    └─ ClearNetwork()                                       │
│                                                              │
│  TearDownSuite()                                           │
│    └─ Save test log to test_logs/{suite}/{test_id}.json   │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Summary

✅ **Each test spins up isolated agents with unique IPs**  
✅ **IP assignment uses Tailscale CGNAT range (100.64.0.x)**  
✅ **Each test has unique test ID for traceability**  
✅ **Logs capture full agent lifecycle (spin up → test → tear down)**  
✅ **Network state tracked per test**  
✅ **Test logs stored in organized folder structure by test ID**

---

*Test logs generated: March 7, 2026 22:43:50 UTC*  
*Test Platform Version: 1.0.0*
