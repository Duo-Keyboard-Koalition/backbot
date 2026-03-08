# TSA2A Authentication & Authorization Design

**Secure Agent Communication on Tailscale**

*Design Document - 2026-03-08*

---

## 1. Overview

TSA2A uses a **layered security model** that leverages Tailscale's built-in security while adding application-level authorization.

```
┌─────────────────────────────────────────────────────┐
│  Application Layer: Task Authorization              │
│  - What tasks can this agent request?               │
│  - What data can this agent access?                 │
├─────────────────────────────────────────────────────┤
│  TSA2A Layer: Connection Authorization              │
│  - Is this agent allowed to connect?                │
│  - What capabilities can they use?                  │
├─────────────────────────────────────────────────────┤
│  Tailscale Layer: Network Authentication            │
│  - WireGuard encryption                            │
│  - Certificate-based identity                       │
└─────────────────────────────────────────────────────┘
```

---

## 2. Identity Model

### 2.1 Agent Identity Structure

Every TSA2A agent has a **composite identity** derived from Tailscale:

```yaml
agent_identity:
  agent_id: "agent_<hostname>_<instance>"
  
  # From Tailscale
  tailscale:
    user: "user@example.com"
    user_id: "ts_user_abc123"
    node: "hostname"
    node_id: "ts_node_xyz789"
    ips:
      - "100.64.0.1"
      - "fd7a:115c:a1e0::1"
    tags: ["agent", "worker"]
    
  # Agent-specific
  metadata:
    agent_type: "darci"
    version: "1.0.0"
    capabilities: ["task_execution", "file_ops"]
    created: "2026-03-08T10:00:00Z"
```

### 2.2 Identity Verification

Agents verify each other's identity using the **Tailscale Local API**:

```go
// Get peer identity from Tailscale
func getPeerIdentity(conn net.Conn) (*TailscaleIdentity, error) {
    // Get local and remote addresses
    localAddr := conn.LocalAddr().(*net.TCPAddr)
    remoteAddr := conn.RemoteAddr().(*net.TCPAddr)
    
    // Query Tailscale Local API
    resp, err := http.Get("http://localhost:8080/api/v2/status")
    if err != nil {
        return nil, err
    }
    
    // Find peer in status
    for _, peer := range resp.Peers {
        if peer.TailscaleIPs.Contains(remoteAddr.IP) {
            return &TailscaleIdentity{
                User: peer.User.LoginName,
                Node: peer.NodeName,
                Tags: peer.Tags,
            }, nil
        }
    }
    
    return nil, ErrPeerNotFound
}
```

### 2.3 Identity Propagation

Once verified, identity is included in all messages:

```json
{
  "header": {
    "sender": {
      "agent_id": "agent_worker_1",
      "tailscale_user": "admin@example.com",
      "tailscale_node": "darci-worker-1",
      "tailscale_ips": ["100.64.0.1"],
      "verified": true
    },
    ...
  }
}
```

---

## 3. Authentication Flow

### 3.1 Connection-Level Authentication

```
┌─────────┐                              ┌─────────┐
│  Agent A │                              │  Agent B │
│(Client)  │                              │(Server)  │
└────┬────┘                              └────┬────┘
     │                                        │
     │  1. TCP Connect (over Tailscale)       │
     │───────────────────────────────────────►│
     │                                        │
     │  2. Get peer identity from Tailscale   │
     │     (Local API call)                   │
     │                                        │
     │  3. Verify identity against policy     │
     │     - Is user allowed?                 │
     │     - Is node allowed?                 │
     │     - Are required tags present?       │
     │                                        │
     │  4. Send Auth Challenge                │
     │◄───────────────────────────────────────│
     │                                        │
     │  5. Respond with identity proof        │
     │     (Tailscale cert + nonce signature) │
     │───────────────────────────────────────►│
     │                                        │
     │  6. Verify response                    │
     │                                        │
     │  7. Auth Result (OK/Failed)            │
     │◄───────────────────────────────────────│
     │                                        │
     │  [If OK: Proceed to Handshake]         │
     │                                        │
```

### 3.2 Auth Challenge Message

```json
{
  "type": "auth_challenge",
  "version": "1.0",
  "nonce": "random_256_bit_challenge",
  "timestamp": "2026-03-08T10:30:00Z",
  "requirements": {
    "require_tailscale": true,
    "require_tags": ["agent"],
    "allowed_users": ["admin@example.com", "agents@example.com"]
  }
}
```

### 3.3 Auth Response Message

```json
{
  "type": "auth_response",
  "version": "1.0",
  "nonce": "random_256_bit_challenge",  // Echo from challenge
  "identity": {
    "agent_id": "agent_worker_1",
    "tailscale_user": "admin@example.com",
    "tailscale_node": "darci-worker-1",
    "tailscale_tags": ["agent", "worker"],
    "tailscale_ips": ["100.64.0.1"]
  },
  "proof": {
    "type": "tailscale_cert",
    "certificate": "base64_encoded_cert",
    "signature": "base64_signature_of_nonce"
  },
  "timestamp": "2026-03-08T10:30:01Z"
}
```

### 3.4 Auth Result Message

```json
{
  "type": "auth_result",
  "version": "1.0",
  "success": true,
  "agent_id": "agent_server_1",
  "session_id": "session_abc123",
  "authorized_capabilities": ["task_execution", "file_ops"],
  "restrictions": {
    "rate_limit": "100/minute",
    "max_concurrent_tasks": 10
  },
  "timestamp": "2026-03-08T10:30:02Z"
}
```

---

## 4. Authorization Policies

### 4.1 Policy Structure

```yaml
# tsaa-policy.yaml
version: "1.0"

# Default action when no rule matches
default_action: deny

# Identity-based rules
identity_rules:
  - name: allow_admin_users
    description: "Allow all agents from admin users"
    match:
      users:
        - "admin@example.com"
        - "root@example.com"
    action: allow
    capabilities: ["*"]
    
  - name: allow_tagged_agents
    description: "Allow agents with specific tags"
    match:
      tags:
        - "agent"
        - "worker"
    action: allow
    capabilities: ["task_execution", "file_ops"]
    
  - name: allow_specific_nodes
    description: "Allow specific node names"
    match:
      nodes:
        - "darci-server"
        - "darci-coordinator"
        - "darci-worker-*"  # Wildcard support
    action: allow
    capabilities: ["*"]
    
  - name: deny_blocked_users
    description: "Explicitly deny blocked users"
    match:
      users:
        - "blocked@example.com"
        - "attacker@example.com"
    action: deny

# Capability-based authorization
capability_rules:
  task_execution:
    allowed_users: ["admin@example.com", "agents@example.com"]
    max_concurrent_per_user: 10
    max_tasks_per_hour: 100
    
  file_operations:
    allowed_users: ["admin@example.com"]
    allowed_paths:
      - "/workspace/*"
      - "/tmp/*"
    denied_paths:
      - "/etc/*"
      - "/root/*"
    denied_operations: ["delete"]
    
  web_search:
    allowed_users: ["*"]  # Anyone can use
    rate_limit: 100/hour
    allowed_providers: ["google", "duckduckgo"]

# Rate limiting
rate_limits:
  default:
    requests_per_minute: 60
    burst: 10
    
  by_user:
    "admin@example.com":
      requests_per_minute: 600
      burst: 100
      
  by_capability:
    task_execution:
      requests_per_minute: 30
    web_search:
      requests_per_minute: 100

# Audit logging
audit:
  log_connections: true
  log_auth_failures: true
  log_capability_denials: true
  log_tasks: true
  retention_days: 90
```

### 4.2 Policy Evaluation

```
┌─────────────────────────────────────────┐
│  Incoming Connection from user@example  │
│  Node: darci-worker-1                   │
│  Tags: [agent, worker]                  │
└─────────────────┬───────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────┐
│  1. Evaluate identity_rules (in order)  │
│     - allow_admin_users? NO             │
│     - allow_tagged_agents? YES → allow  │
└─────────────────┬───────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────┐
│  2. Apply capability_rules              │
│     - task_execution: allowed           │
│     - file_operations: allowed          │
│     - web_search: allowed               │
└─────────────────┬───────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────┐
│  3. Apply rate_limits                   │
│     - 60 requests/minute                │
│     - 10 burst                          │
└─────────────────┬───────────────────────┘
                  │
                  ▼
         ┌────────────────┐
         │  CONNECTION    │
         │  AUTHORIZED    │
         └────────────────┘
```

### 4.3 Policy API

```http
GET /api/v1/policy

Response: 200 OK
{
  "version": "1.0",
  "default_action": "deny",
  "identity_rules": [...],
  "capability_rules": {...},
  "rate_limits": {...}
}
```

```http
POST /api/v1/policy/evaluate
Content-Type: application/json

{
  "identity": {
    "user": "user@example.com",
    "node": "darci-worker-1",
    "tags": ["agent"]
  },
  "requested_capability": "task_execution"
}

Response: 200 OK
{
  "allowed": true,
  "restrictions": {
    "max_concurrent": 10,
    "rate_limit": "30/minute"
  }
}
```

---

## 5. Session Management

### 5.1 Session Lifecycle

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│  CLOSED  │────►│  AUTH    │────►│  ACTIVE  │────►│  CLOSED  │
└──────────┘     └──────────┘     └──────────┘     └──────────┘
                    │                  │
                    │                  │
                    ▼                  ▼
               ┌──────────┐     ┌──────────┐
               │  FAILED  │     │  EXPIRED │
               └──────────┘     └──────────┘
```

### 5.2 Session State

```go
type Session struct {
    ID            string
    AgentID       string
    PeerIdentity  *TailscaleIdentity
    State         SessionState
    Created       time.Time
    LastActivity  time.Time
    Expires       time.Time
    
    // Authorized capabilities
    Capabilities []string
    
    // Rate limiting
    RateLimiter   *RateLimiter
    
    // Statistics
    MessagesSent  int64
    MessagesRecv  int64
}

type SessionState int

const (
    SessionAuth SessionState = iota
    SessionActive
    SessionExpired
    SessionClosed
)
```

### 5.3 Session Expiration

Sessions expire when:
- **Idle timeout**: No activity for 300 seconds
- **Absolute timeout**: Session older than 8 hours
- **Identity expired**: Tailscale cert expired
- **Manual revoke**: Admin revokes session

```json
{
  "type": "session_expired",
  "reason": "idle_timeout",
  "session_id": "session_abc123",
  "last_activity": "2026-03-08T10:30:00Z",
  "expired_at": "2026-03-08T10:35:00Z"
}
```

---

## 6. Message Security

### 6.1 Message Signatures

For non-repudiation, messages can be signed:

```json
{
  "header": {
    "message_id": "msg_001",
    "type": "task_request",
    "sender": {"agent_id": "agent_a"},
    "recipient": {"agent_id": "agent_b"},
    "timestamp": "2026-03-08T10:30:00Z",
    "nonce": "random_nonce"
  },
  "payload": {...},
  "signature": {
    "algorithm": "ed25519",
    "key_id": "key_abc123",
    "value": "base64_encoded_signature"
  }
}
```

### 6.2 Signature Verification

```go
func verifyMessage(msg Message, publicKey ed25519.PublicKey) bool {
    // Get signed content (header + payload)
    signedContent, err := json.Marshal(struct {
        Header  json.RawMessage
        Payload json.RawMessage
    }{
        Header:  msg.Header,
        Payload: msg.Payload,
    })
    if err != nil {
        return false
    }
    
    // Decode signature
    sig, err := base64.StdEncoding.DecodeString(msg.Signature.Value)
    if err != nil {
        return false
    }
    
    // Verify
    return ed25519.Verify(publicKey, signedContent, sig)
}
```

### 6.3 Replay Protection

Messages include nonce and timestamp to prevent replay:

```go
type ReplayProtector struct {
    seenNonces map[string]time.Time
    mu         sync.Mutex
    maxAge     time.Duration
}

func (r *ReplayProtector) IsReplay(nonce string, timestamp time.Time) bool {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    // Check if too old
    if time.Since(timestamp) > r.maxAge {
        return true
    }
    
    // Check if seen before
    if _, seen := r.seenNonces[nonce]; seen {
        return true
    }
    
    // Record nonce
    r.seenNonces[nonce] = time.Now()
    
    // Cleanup old nonces
    r.cleanup()
    
    return false
}
```

---

## 7. Rate Limiting

### 7.1 Rate Limit Strategies

**Token Bucket:**
```go
type TokenBucket struct {
    tokens     float64
    maxTokens  float64
    refillRate float64  // tokens per second
    lastRefill time.Time
}

func (t *TokenBucket) Allow() bool {
    t.refill()
    if t.tokens >= 1 {
        t.tokens--
        return true
    }
    return false
}
```

**Sliding Window:**
```go
type SlidingWindow struct {
    requests  []time.Time
    window    time.Duration
    maxReqs   int
}

func (s *SlidingWindow) Allow() bool {
    now := time.Now()
    cutoff := now.Add(-s.window)
    
    // Remove old requests
    for len(s.requests) > 0 && s.requests[0].Before(cutoff) {
        s.requests = s.requests[1:]
    }
    
    if len(s.requests) >= s.maxReqs {
        return false
    }
    
    s.requests = append(s.requests, now)
    return true
}
```

### 7.2 Rate Limit Headers

```json
{
  "type": "rate_limit_info",
  "limits": {
    "requests_per_minute": 60,
    "remaining": 45,
    "reset_at": "2026-03-08T10:31:00Z"
  }
}
```

### 7.3 Rate Limit Exceeded

```json
{
  "type": "error",
  "error_code": 7001,
  "error_name": "RATE_LIMIT_EXCEEDED",
  "message": "Rate limit exceeded. Try again in 30 seconds.",
  "retry_after": 30
}
```

---

## 8. Audit Logging

### 8.1 Log Events

```yaml
events:
  - name: connection_accepted
    fields: [session_id, agent_id, user, node]
    
  - name: connection_denied
    fields: [agent_id, user, node, reason]
    
  - name: capability_authorized
    fields: [session_id, capability, restrictions]
    
  - name: capability_denied
    fields: [session_id, capability, reason]
    
  - name: message_sent
    fields: [session_id, message_id, type, size]
    
  - name: message_received
    fields: [session_id, message_id, type, size]
    
  - name: task_executed
    fields: [session_id, task_id, type, duration_ms, result]
    
  - name: auth_failure
    fields: [agent_id, user, reason, ip]
    
  - name: rate_limit_exceeded
    fields: [session_id, limit, current]
    
  - name: session_expired
    fields: [session_id, reason, duration]
```

### 8.2 Log Format

```json
{
  "timestamp": "2026-03-08T10:30:00Z",
  "event": "connection_accepted",
  "session_id": "session_abc123",
  "agent": {
    "id": "agent_worker_1",
    "user": "admin@example.com",
    "node": "darci-worker-1"
  },
  "source_ip": "100.64.0.1",
  "metadata": {
    "user_agent": "DarCI/1.0.0",
    "capabilities": ["task_execution", "file_ops"]
  }
}
```

### 8.3 Log Storage

```yaml
logging:
  format: json
  output:
    - type: file
      path: /var/log/tsa2a/auth.log
      rotation: daily
      retention: 90
      
    - type: syslog
      facility: local0
      
    - type: stdout
      format: pretty
      
  audit:
    enabled: true
    path: /var/log/tsa2a/audit.log
    retention: 365
```

---

## 9. Security Best Practices

### 9.1 Implementation Checklist

- [ ] Always verify Tailscale certificates
- [ ] Implement authorization policies
- [ ] Enable rate limiting
- [ ] Log all authentication events
- [ ] Use message signatures for critical operations
- [ ] Implement session expiration
- [ ] Handle identity revocation
- [ ] Monitor for anomalous behavior

### 9.2 Threat Mitigation

| Threat | Mitigation |
|--------|------------|
| Unauthorized access | Tailscale auth + TSA2A policies |
| Compromised agent | Authorization policies + capability restrictions |
| Replay attacks | Nonces + timestamps + signature |
| DoS | Rate limiting + circuit breakers |
| Privilege escalation | Capability-based access control |
| Data exfiltration | Path restrictions + audit logging |

### 9.3 Incident Response

```yaml
incident_response:
  detection:
    - failed_auth_threshold: 10/minute
    - rate_limit_violations: 100/minute
    - unusual_capability_requests: true
    
  response:
    - auto_block_after_failures: 50
    - alert_admin: true
    - log_details: true
    
  recovery:
    - manual_unblock_required: true
    - review_audit_logs: true
    - update_policies: true
```

---

## 10. Implementation Guide

### 10.1 Go Implementation

```go
package tsa2a

type Authenticator struct {
    policy     *AuthorizationPolicy
    sessions   *SessionManager
    rateLimiter *RateLimiter
}

func (a *Authenticator) Authenticate(conn net.Conn) (*Session, error) {
    // Get peer identity from Tailscale
    identity, err := getTailscaleIdentity(conn)
    if err != nil {
        return nil, err
    }
    
    // Check authorization policy
    if !a.policy.Allows(identity) {
        return nil, ErrUnauthorized
    }
    
    // Create session
    session := a.sessions.Create(identity)
    
    return session, nil
}

func (a *Authenticator) AuthorizeCapability(session *Session, cap string) error {
    if !session.HasCapability(cap) {
        return ErrCapabilityDenied
    }
    
    if !a.rateLimiter.Allow(session.AgentID, cap) {
        return ErrRateLimitExceeded
    }
    
    return nil
}
```

### 10.2 Python Implementation

```python
class TSA2AAuthenticator:
    def __init__(self, policy: AuthorizationPolicy):
        self.policy = policy
        self.sessions = SessionManager()
        self.rate_limiter = RateLimiter()
    
    async def authenticate(self, reader, writer) -> Session:
        # Get peer info from Tailscale
        peer_ip = writer.get_extra_info('peername')[0]
        identity = await self.get_tailscale_identity(peer_ip)
        
        # Check policy
        if not self.policy.allows(identity):
            raise UnauthorizedError()
        
        # Create session
        session = self.sessions.create(identity)
        
        return session
    
    async def authorize_capability(self, session: Session, cap: str):
        if cap not in session.capabilities:
            raise CapabilityDeniedError()
        
        if not self.rate_limiter.allow(session.agent_id, cap):
            raise RateLimitExceededError()
```

---

## Appendix: Configuration Reference

### A.1 Full Configuration Example

```yaml
# tsaa-config.yaml
version: "1.0"

server:
  port: 4848
  host: "0.0.0.0"
  
tailscale:
  local_api: "http://localhost:8080"
  require_tags: ["agent"]
  
authentication:
  require_auth: true
  challenge_timeout: 30s
  max_retries: 3
  
authorization:
  default_action: deny
  
  identity_rules:
    - name: allow_admin
      match:
        users: ["admin@example.com"]
      action: allow
      
    - name: allow_workers
      match:
        tags: ["worker"]
      action: allow
      capabilities: ["task_execution"]
  
  capability_rules:
    task_execution:
      allowed_users: ["admin@example.com", "agents@example.com"]
      max_concurrent: 10
      
    file_operations:
      allowed_users: ["admin@example.com"]
      allowed_paths: ["/workspace/*"]
      
rate_limiting:
  enabled: true
  default:
    requests_per_minute: 60
    burst: 10
    
session:
  idle_timeout: 300s
  absolute_timeout: 8h
  heartbeat_interval: 30s
  
audit:
  enabled: true
  log_connections: true
  log_auth_failures: true
  log_capability_denials: true
  retention_days: 90
```
