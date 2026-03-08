# TSA2A Quickstart Guide

**Get agents talking in 5 minutes**

---

## 1. What is TSA2A?

TSA2A (Tailscale Agent-to-Agent) is a protocol for AI agents to communicate securely over Tailscale.

> "No HTTPS, just TSA2A"

**Features:**
- ✅ Auto-discovery of agents on your Tailscale network
- ✅ Automatic encryption (via Tailscale/WireGuard)
- ✅ Built-in authentication (via Tailscale identities)
- ✅ Capability-based authorization
- ✅ Simple JSON message format

---

## 2. Prerequisites

- [Tailscale](https://tailscale.com) installed and running
- Go 1.21+ or Python 3.10+
- Two or more machines/VMs/containers on Tailscale

---

## 3. Quick Start (Go)

### 3.1 Install

```bash
go get github.com/tailbridge/taila2a/go/pkg/tsa2a
```

### 3.2 Run Agent A (Server)

```go
package main

import (
    "log"
    "github.com/tailbridge/taila2a/go/pkg/tsa2a"
)

func main() {
    // Create agent
    agent := tsa2a.NewAgent(tsa2a.Config{
        Port: 4848,
        AgentID: "agent_worker_1",
        Capabilities: []string{"task_execution", "file_ops"},
    })
    
    // Register task handler
    agent.Handle("task_request", func(msg tsa2a.Message) error {
        log.Printf("Received task: %s", msg.Payload)
        return nil
    })
    
    // Start listening
    log.Fatal(agent.ListenAndServe())
}
```

### 3.3 Run Agent B (Client)

```go
package main

import (
    "log"
    "github.com/tailbridge/taila2a/go/pkg/tsa2a"
)

func main() {
    agent := tsa2a.NewAgent(tsa2a.Config{
        AgentID: "agent_coordinator",
    })
    
    // Connect to worker
    session, err := agent.Connect("agent_worker_1")
    if err != nil {
        log.Fatal(err)
    }
    
    // Send task
    err = session.Send(tsa2a.Message{
        Type: "task_request",
        Payload: map[string]interface{}{
            "task": "process_data",
            "input": "file.txt",
        },
    })
    
    log.Fatal(err)
}
```

---

## 4. Quick Start (Python)

### 4.1 Install

```bash
pip install tsa2a
```

### 4.2 Run Agent A (Server)

```python
from tsa2a import Agent, Message

agent = Agent(
    agent_id="agent_worker-1",
    port=4848,
    capabilities=["task_execution", "file_ops"]
)

@agent.handle("task_request")
def handle_task(msg: Message):
    print(f"Received task: {msg.payload}")
    return {"status": "completed"}

agent.listen()
```

### 4.3 Run Agent B (Client)

```python
from tsa2a import Agent, Message

agent = Agent(agent_id="agent-coordinator")

# Connect to worker
session = agent.connect("agent-worker-1")

# Send task
response = session.send(Message(
    type="task_request",
    payload={
        "task": "process_data",
        "input": "file.txt"
    }
))

print(f"Response: {response}")
```

---

## 5. Configuration

### 5.1 Basic Config (YAML)

```yaml
# tsaa.yaml
agent_id: "agent_worker_1"
port: 4848

capabilities:
  - name: task_execution
    version: "1.0"
  - name: file_ops
    version: "1.0"

discovery:
  mdns: true
  tailscale_dns: true

auth:
  allowed_users:
    - "admin@example.com"
  allowed_tags:
    - "agent"
```

### 5.2 Run with Config

```bash
# Go
tsa2a-agent -config tsaa.yaml

# Python
python -m tsa2a.agent --config tsaa.yaml
```

---

## 6. Discovery

### 6.1 Find Agents

```bash
# List all agents on Tailscale network
tsa2a discover

# Output:
# AGENT ID              IP            CAPABILITIES
# agent_worker_1        100.64.0.1    task_execution, file_ops
# agent_worker_2        100.64.0.2    web_search, summarization
```

### 6.2 Query by Capability

```bash
tsa2a discover --capability task_execution
```

---

## 7. Testing

### 7.1 Test Connection

```bash
tsa2a connect agent_worker_1 --test
```

### 7.2 Send Test Message

```bash
tsa2a send agent_worker_1 \
  --type task_request \
  --payload '{"task": "test"}'
```

---

## 8. Common Patterns

### 8.1 Request-Response

```python
# Client
response = session.send(Message(
    type="task_request",
    payload={"task": "process"},
    requires_ack=True
))
print(f"Result: {response.payload}")
```

### 8.2 Fire-and-Forget

```python
session.send(Message(
    type="status_update",
    payload={"status": "busy"}
))
```

### 8.3 Streaming

```python
session.send(Message(
    type="task_request",
    payload={"task": "process_large_file"},
    flags={"streaming": True}
))

for chunk in session.stream():
    print(f"Chunk: {chunk}")
```

---

## 9. Troubleshooting

### 9.1 Can't Connect

```bash
# Check Tailscale status
tailscale status

# Check if agent is listening
netstat -tlnp | grep 4848

# Test connectivity
tailscale ping agent_worker_1
```

### 9.2 Auth Failed

```bash
# Check allowed users in config
cat tsaa.yaml | grep allowed_users

# Verify Tailscale user
tailscale whoami
```

### 9.3 Discovery Not Working

```bash
# Restart mDNS
sudo systemctl restart avahi-daemon

# Check DNS
dig agent_worker_1.ts.net
```

---

## 10. Next Steps

- [Protocol Specification](TSA2A-PROTOCOL.md) - Full protocol details
- [Architecture](TSA2A-ARCHITECTURE.md) - System design
- [Authentication](TSA2A-AUTH.md) - Security model
- [Examples](../examples/) - Code samples

---

## 11. Help

```bash
tsa2a --help
tsa2a connect --help
tsa2a discover --help
```

**Issues:** https://github.com/tailbridge/taila2a/issues

**Discord:** https://discord.gg/tailscale
