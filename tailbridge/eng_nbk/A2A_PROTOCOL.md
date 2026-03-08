# Secure Agent-to-Agent (A2A) Protocol

## Overview

A Kafka-inspired, secure agent-to-agent communication protocol built on Tailscale's zero-trust network. This protocol enables distributed multi-agent systems with guaranteed message delivery, topic-based routing, and end-to-end encryption.

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    A2A Protocol Stack                                    │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  Application Layer                                                       │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Agent Logic / Business Rules                                    │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                      │                                                   │
│  A2A Protocol Layer  ◀── This specification                             │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Message Envelope  │  Topic Routing  │  Consumer Groups         │    │
│  │  Correlation IDs   │  Reply Channels │  Offset Tracking         │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                      │                                                   │
│  Security Layer                                                        │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  mTLS (Tailscale)  │  Ed25519 Signatures  │  ACL Enforcement   │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                      │                                                   │
│  Transport Layer                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  WireGuard (via Tailscale)  │  HTTP/2  │  gRPC (optional)       │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Message Envelope

### Schema

```json
{
  "header": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "type": "request",
    "version": "1.0",
    "source_agent": {
      "id": "agent-alpha",
      "node_id": "tailscale-node-key-abc123",
      "tailnet": "example.com"
    },
    "dest_agent": {
      "id": "agent-beta",
      "node_id": "tailscale-node-key-def456",
      "tailnet": "example.com"
    },
    "topic": "agent.requests",
    "partition": 0,
    "timestamp": "2026-03-07T12:00:00.000Z",
    "correlation_id": "corr-uuid-123",
    "reply_to": "agent.responses",
    "ttl_ms": 30000
  },
  "body": {
    "action": "execute_task",
    "content_type": "application/json",
    "payload": {
      "task_id": "task-001",
      "parameters": {}
    },
    "metadata": {
      "priority": "high",
      "retry_count": 0
    }
  },
  "security": {
    "signature": "ed25519-base64-signature",
    "public_key": "ed25519-public-key",
    "timestamp": "2026-03-07T12:00:00.000Z"
  }
}
```

### Header Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | Unique message identifier |
| `type` | string | ✅ | `request`, `response`, or `event` |
| `version` | string | ✅ | Protocol version |
| `source_agent` | object | ✅ | Sending agent identity |
| `dest_agent` | object | ❌ | Target agent (omit for broadcasts) |
| `topic` | string | ✅ | Target topic for routing |
| `partition` | int | ❌ | Topic partition (default: 0) |
| `timestamp` | datetime | ✅ | Message creation time (UTC) |
| `correlation_id` | UUID | ❌ | Links request/response pairs |
| `reply_to` | string | ❌ | Topic for responses |
| `ttl_ms` | int | ❌ | Time-to-live in milliseconds |

### Body Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `action` | string | ✅ | Action/command to execute |
| `content_type` | string | ✅ | MIME type of payload |
| `payload` | any | ❌ | Message content |
| `metadata` | object | ❌ | Additional context |

### Security Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `signature` | string | ✅ | Ed25519 signature of header+body |
| `public_key` | string | ✅ | Sender's public key for verification |
| `timestamp` | datetime | ✅ | Signature timestamp (replay protection) |

---

## Topic-Based Routing

### Topic Naming

```
<category>.<subcategory>.<action>

Examples:
  agent.requests          # Requests for agent processing
  agent.responses         # Responses from agents
  agent.events            # Agent lifecycle events
  system.health           # Health monitoring
  system.discovery        # Agent discovery
```

### Topic Structure

```
┌─────────────────────────────────────────────────────────────────┐
│  Topic: agent.requests                                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Partition 0    Partition 1    Partition 2                       │
│  ┌─────────┐   ┌─────────┐   ┌─────────┐                       │
│  │ Msg 0   │   │ Msg 1   │   │ Msg 2   │                       │
│  │ Msg 3   │   │ Msg 4   │   │ Msg 5   │                       │
│  │ Msg 6   │   │ Msg 7   │   │ Msg 8   │                       │
│  └─────────┘   └─────────┘   └─────────┘                       │
│      │             │             │                               │
│      ▼             ▼             ▼                               │
│  Agent A1      Agent B1      Agent C1                            │
│  (Consumer)    (Consumer)    (Consumer)                          │
│                                                                  │
│  Consumer Group: agent-pool-alpha                                │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Consumer Groups

### Configuration

```json
{
  "group_id": "agent-pool-alpha",
  "topics": ["agent.requests"],
  "partitions": [0, 1],
  "offset_policy": "latest",
  "auto_commit": true,
  "session_timeout_ms": 30000,
  "heartbeat_interval_ms": 3000
}
```

### Offset Policies

| Policy | Behavior |
|--------|----------|
| `earliest` | Start from oldest unprocessed message |
| `latest` | Start from newest message |
| `committed` | Start from last committed offset |

---

## Security Model

### Zero Trust Principles

1. **Never trust, always verify** - Every message is authenticated
2. **Least privilege** - Agents only access authorized topics
3. **Explicit verification** - All identities cryptographically verified

### Authentication Flow

```
┌──────────────┐                          ┌──────────────┐
│   Agent A    │                          │   Agent B    │
│              │                          │              │
│  1. Sign     │                          │              │
│     message  │                          │              │
│      │       │                          │              │
│      ▼       │                          │              │
│  2. Send     │─────────────────────────▶│  3. Verify   │
│     (mTLS)   │      Tailscale           │     signature│
│              │      WireGuard           │      │       │
│              │                          │      ▼       │
│  5. Verify   │◀─────────────────────────│  4. Process  │
│     response │      Encrypted           │     & Reply  │
│              │                          │              │
└──────────────┘                          └──────────────┘
```

### Message Signing

```go
// Sign message
func SignMessage(msg *Envelope, privateKey ed25519.PrivateKey) {
    // Create canonical representation
    canonical := marshalCanonical(msg.Header, msg.Body)
    
    // Sign with timestamp for replay protection
    timestamp := time.Now().UTC()
    data := append(canonical, timestamp.Bytes()...)
    
    // Generate signature
    signature := ed25519.Sign(privateKey, data)
    
    msg.Security = Security{
        Signature:  base64Encode(signature),
        PublicKey:  base64Encode(privateKey.Public().Bytes()),
        Timestamp:  timestamp,
    }
}

// Verify message
func VerifyMessage(msg *Envelope) error {
    // Check timestamp freshness (±5 minutes)
    if time.Since(msg.Security.Timestamp) > 5*time.Minute {
        return ErrReplayAttack
    }
    
    // Decode public key
    publicKey, err := base64Decode(msg.Security.PublicKey)
    if err != nil {
        return err
    }
    
    // Reconstruct signed data
    canonical := marshalCanonical(msg.Header, msg.Body)
    data := append(canonical, msg.Security.Timestamp.Bytes()...)
    
    // Decode signature
    signature, err := base64Decode(msg.Security.Signature)
    if err != nil {
        return err
    }
    
    // Verify
    if !ed25519.Verify(publicKey, data, signature) {
        return ErrInvalidSignature
    }
    
    return nil
}
```

### Tailscale Integration

```json
{
  "tailscale": {
    "node_id": "Nabcdefghijklmnopqrstuvwxyz123456",
    "node_name": "agent-alpha",
    "tailnet": "example.com",
    "ips": ["100.64.0.1", "fd7a:115c:a1e0::1"],
    "capabilities": ["tag:agent", "tag:producer"],
    "key_expiry": "2026-12-31T23:59:59Z"
  }
}
```

---

## Consumer Group Protocol

### Join Group

```json
{
  "type": "join_group",
  "group_id": "agent-pool-alpha",
  "member_id": "agent-alpha-uuid",
  "topics": ["agent.requests"],
  "capabilities": ["process_type_a", "process_type_b"]
}
```

### Heartbeat

```json
{
  "type": "heartbeat",
  "group_id": "agent-pool-alpha",
  "member_id": "agent-alpha-uuid",
  "generation_id": 5
}
```

### Offset Commit

```json
{
  "type": "offset_commit",
  "group_id": "agent-pool-alpha",
  "member_id": "agent-alpha-uuid",
  "offsets": [
    {
      "topic": "agent.requests",
      "partition": 0,
      "offset": 12345
    }
  ]
}
```

---

## Error Handling

### Error Codes

| Code | Name | Description |
|------|------|-------------|
| 1001 | `ERR_INVALID_SIGNATURE` | Message signature verification failed |
| 1002 | `ERR_REPLAY_ATTACK` | Message timestamp outside acceptable window |
| 1003 | `ERR_UNKNOWN_TOPIC` | Topic does not exist |
| 1004 | `ERR_UNAUTHORIZED` | Agent not authorized for topic |
| 1005 | `ERR_INVALID_PARTITION` | Partition does not exist |
| 1006 | `ERR_OFFSET_OUT_OF_RANGE` | Requested offset unavailable |
| 1007 | `ERR_GROUP_COORDINATOR` | Consumer group coordinator unavailable |
| 1008 | `ERR_DUPLICATE_MESSAGE` | Message ID already processed |

### Error Response

```json
{
  "header": {
    "type": "error",
    "correlation_id": "corr-uuid-123",
    "timestamp": "2026-03-07T12:00:00.000Z"
  },
  "body": {
    "error_code": 1001,
    "error_message": "Invalid message signature",
    "original_message_id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

---

## Implementation Guide

### Sending a Message

```go
package main

import (
    "context"
    "crypto/ed25519"
    "time"
    
    "github.com/codejedi-ai/kaggle-for-tensors/tailscale-app/internal/protocol"
)

func SendMessage(ctx context.Context, agent *Agent, msg *protocol.Envelope) error {
    // 1. Set header
    msg.Header = protocol.Header{
        ID:        uuid.New().String(),
        Type:      "request",
        Version:   "1.0",
        Source:    agent.Identity,
        Topic:     "agent.requests",
        Timestamp: time.Now().UTC(),
        TTL:       30 * time.Second,
    }
    
    // 2. Sign message
    msg.Sign(agent.PrivateKey)
    
    // 3. Discover destination
    dest, err := agent.Discovery.Lookup(ctx, "agent-beta")
    if err != nil {
        return err
    }
    
    // 4. Send via Tailscale
    url := fmt.Sprintf("https://%s:%d/a2a/inbound", dest.IP, 8001)
    return agent.HTTPClient.Post(ctx, url, msg)
}
```

### Receiving a Message

```go
func HandleInbound(w http.ResponseWriter, r *http.Request) {
    // 1. Parse envelope
    var msg protocol.Envelope
    if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // 2. Verify Tailscale identity
    identity, err := tailscale.VerifyPeer(r.Context(), r.RemoteAddr)
    if err != nil {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }
    
    // 3. Verify message signature
    if err := msg.Verify(); err != nil {
        http.Error(w, err.Error(), http.StatusForbidden)
        return
    }
    
    // 4. Check authorization (ACL)
    if !acl.CanPublish(identity, msg.Header.Topic) {
        http.Error(w, "forbidden", http.StatusForbidden)
        return
    }
    
    // 5. Route to topic
    if err := eventBus.Publish(&msg); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusAccepted)
}
```

---

## Comparison: Kafka vs A2A Protocol

| Feature | Kafka | A2A Protocol |
|---------|-------|--------------|
| **Transport** | TCP | Tailscale/WireGuard |
| **Security** | SSL/TLS + SASL | mTLS + Ed25519 |
| **Discovery** | ZooKeeper/KRaft | Tailscale DNS |
| **Topology** | Centralized brokers | Peer-to-peer |
| **Topics** | Broker-managed | Distributed |
| **Consumer Groups** | Coordinator-based | Distributed consensus |
| **Message Persistence** | Disk (logs) | Memory + optional WAL |
| **Use Case** | High-throughput streaming | Secure multi-agent |

---

## Best Practices

### 1. Message Design

- Keep payloads small (< 1MB)
- Use correlation IDs for request/response tracking
- Set appropriate TTL for time-sensitive messages
- Include version in schema for evolution

### 2. Security

- Rotate keys regularly
- Verify timestamps (replay protection)
- Use Tailscale ACLs for topic authorization
- Log all authentication failures

### 3. Reliability

- Implement idempotent consumers
- Use at-least-once delivery semantics
- Handle duplicate messages gracefully
- Monitor consumer lag

### 4. Performance

- Partition topics for parallelism
- Batch messages when possible
- Use compression for large payloads
- Monitor queue depths

---

*Last updated: March 7, 2026*
*Version: 1.0.0*
