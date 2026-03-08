# tsA2A Handshake Protocol

**Version:** 1.0.0  
**Created:** 2026-03-08  
**Status:** 🟡 Draft

---

## Overview

The **tsA2A (Tailscale Agent-to-Agent) Handshake Protocol** provides secure agent identification through a **challenge-response mechanism** instead of network scanning.

### Key Principle

> **NO auto-discovery scanning.** Instead, bridges send a **challenge** to the AIP endpoint, and only agents that respond with the correct **handshake signature** are identified as legitimate agents.

---

## Protocol Flow

```
┌─────────────┐                              ┌─────────────┐
│   Bridge    │                              │   Agent     │
│  (tsA2A)    │                              │  (DarCI)    │
└──────┬──────┘                              └──────┬──────┘
       │                                            │
       │  1. Discover potential agent (IP:PORT)     │
       │  ──────────────────────────────────────▶   │
       │                                            │
       │  2. POST /aip/handshake {                  │
       │       challenge: "abc123...",              │
       │       timestamp: "...",                    │
       │       nonce: "xyz789"                      │
       │     }                                      │
       │  ──────────────────────────────────────▶   │
       │                                            │
       │  3. Agent computes:                        │
       │     signature = HMAC-SHA256(               │
       │       challenge + timestamp + nonce,       │
       │       AGENT_SECRET                         │
       │     )                                      │
       │                                            │
       │  4. Response: {                            │
       │       signature: "computed_hmac",          │
       │       agent_id: "darci-python-001",        │
       │       agent_type: "darci-python"           │
       │     }                                      │
       │  ◀─────────────────────────────────────    │
       │                                            │
       │  5. Bridge verifies signature              │
       │     - If valid → mark as legitimate agent  │
       │     - If invalid → ignore/log              │
       │                                            │
       │  6. POST /aip/register (full registration) │
       │  ──────────────────────────────────────▶   │
       │                                            │
```

---

## Handshake Challenge Format

### Request

```http
POST /aip/handshake
Host: <agent-ip>:<port>
Content-Type: application/json

{
  "challenge": "random-string-128-bits",
  "timestamp": "2026-03-08T10:00:00Z",
  "nonce": "unique-identifier",
  "bridge_id": "bridge-alpha",
  "protocol_version": "1.0"
}
```

### Response (Valid Agent)

```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "signature": "hmac-sha256-result",
  "agent_id": "darci-python-001",
  "agent_type": "darci-python",
  "agent_version": "1.0.0",
  "capabilities": ["task-execution", "notebook"],
  "handshake_version": "1.0"
}
```

### Response (Not an Agent / Invalid)

```http
HTTP/1.1 404 Not Found
Content-Type: application/json

{
  "error": "not_an_agent"
}
```

Or no response at all (timeout).

---

## Signature Computation

### Algorithm

```python
def compute_handshake_signature(challenge, timestamp, nonce, secret):
    """
    Compute HMAC-SHA256 signature for handshake.
    
    Args:
        challenge: Random challenge string from bridge
        timestamp: ISO 8601 timestamp
        nonce: Unique nonce
        secret: Shared secret (AGENT_SECRET)
    
    Returns:
        Hex-encoded HMAC-SHA256 signature
    """
    import hmac
    import hashlib
    
    # Message is concatenation of challenge, timestamp, nonce
    message = f"{challenge}:{timestamp}:{nonce}".encode('utf-8')
    secret_bytes = secret.encode('utf-8')
    
    # Compute HMAC-SHA256
    signature = hmac.new(secret_bytes, message, hashlib.sha256)
    
    return signature.hexdigest()
```

### Verification

Bridge verifies by:
1. Looking up the shared secret for the claimed `agent_id`
2. Recomputing the signature
3. Comparing with constant-time comparison

```go
func verifyHandshakeSignature(challenge, timestamp, nonce, signature, secret string) bool {
    message := fmt.Sprintf("%s:%s:%s", challenge, timestamp, nonce)
    
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write([]byte(message))
    expectedSignature := hex.EncodeToString(mac.Sum(nil))
    
    // Constant-time comparison
    return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
```

---

## Shared Secret Distribution

### Option 1: Pre-Shared Secrets (Manual)

Configure secrets in bridge config:

```json
{
  "bridge_name": "bridge-alpha",
  "agent_secrets": {
    "darci-python-001": "super-secret-key-123",
    "darci-go-001": "another-secret-456"
  }
}
```

### Option 2: Dynamic Secret Exchange

1. Agent generates secret on first run
2. Admin copies secret to bridge config
3. Bridge and agent share secret

### Option 3: PKI-Based (Future)

- Agent has private key
- Bridge has agent's public key
- Signature uses asymmetric crypto

---

## Challenge Generation

Challenges must be:
- **Random**: Cryptographically secure random
- **Unique**: Never repeat
- **Time-limited**: Expire after N seconds

```go
func generateChallenge() (string, error) {
    bytes := make([]byte, 16) // 128 bits
    _, err := rand.Read(bytes)
    if err != nil {
        return "", err
    }
    return hex.EncodeToString(bytes), nil
}

func isChallengeExpired(timestamp string, maxAge time.Duration) bool {
    challengeTime, err := time.Parse(time.RFC3339, timestamp)
    if err != nil {
        return true
    }
    return time.Since(challengeTime) > maxAge
}
```

---

## Nonce Tracking

Prevent replay attacks by tracking used nonces:

```go
type NonceStore struct {
    usedNonces map[string]time.Time
    mu         sync.RWMutex
    expiry     time.Duration
}

func (n *NonceStore) IsNonceUsed(nonce string) bool {
    n.mu.RLock()
    defer n.mu.RUnlock()
    _, exists := n.usedNonces[nonce]
    return exists
}

func (n *NonceStore) MarkNonceUsed(nonce string) {
    n.mu.Lock()
    defer n.mu.Unlock()
    n.usedNonces[nonce] = time.Now()
}

// Cleanup old nonces periodically
func (n *NonceStore) Cleanup() {
    n.mu.Lock()
    defer n.mu.Unlock()
    cutoff := time.Now().Add(-n.expiry)
    for nonce, timestamp := range n.usedNonces {
        if timestamp.Before(cutoff) {
            delete(n.usedNonces, nonce)
        }
    }
}
```

---

## Agent Identification States

```
┌──────────────┐
│  UNKNOWN     │ ◀── Initial state
└──────┬───────┘
       │
       │ Send handshake challenge
       ▼
┌──────────────┐
│  PENDING     │ ◀── Awaiting response
└──────┬───────┘
       │
       ├─────────────┬──────────────┐
       │             │              │
       ▼             ▼              ▼
┌──────────┐  ┌──────────┐  ┌──────────┐
│ VERIFIED │  │ REJECTED │  │ TIMEOUT  │
└──────────┘  └──────────┘  └──────────┘
```

### State Transitions

| From | Trigger | To | Action |
|------|---------|-----|--------|
| UNKNOWN | Challenge sent | PENDING | Track challenge |
| PENDING | Valid signature | VERIFIED | Add to registry |
| PENDING | Invalid signature | REJECTED | Log, blacklist |
| PENDING | No response (5s) | TIMEOUT | Remove from tracking |
| PENDING | Expired challenge | TIMEOUT | Remove from tracking |

---

## Implementation

### Bridge Side (Go)

```go
type HandshakeService struct {
    registry    *AgentRegistry
    nonceStore  *NonceStore
    secrets     map[string]string
    challenges  map[string]*PendingChallenge
    mu          sync.RWMutex
}

type PendingChallenge struct {
    challenge string
    nonce     string
    timestamp time.Time
    targetIP  string
}

func (h *HandshakeService) SendHandshake(ip string, port int) (*AgentInfo, error) {
    // Generate challenge
    challenge, _ := generateChallenge()
    nonce, _ := generateNonce()
    timestamp := time.Now().UTC().Format(time.RFC3339)
    
    // Store pending challenge
    h.storeChallenge(ip, challenge, nonce, timestamp)
    
    // Send handshake request
    url := fmt.Sprintf("http://%s:%d/aip/handshake", ip, port)
    payload := HandshakeRequest{
        Challenge: challenge,
        Timestamp: timestamp,
        Nonce:     nonce,
        BridgeID:  h.bridgeID,
    }
    
    resp, err := h.httpClient.Post(url, "application/json", json.Marshal(payload))
    if err != nil {
        return nil, err
    }
    
    // Parse response
    var handshakeResp HandshakeResponse
    json.NewDecoder(resp.Body).Decode(&handshakeResp)
    
    // Verify signature
    secret, exists := h.secrets[handshakeResp.AgentID]
    if !exists {
        return nil, fmt.Errorf("unknown agent")
    }
    
    valid := verifyHandshakeSignature(
        challenge, timestamp, nonce,
        handshakeResp.Signature, secret,
    )
    
    if !valid {
        return nil, fmt.Errorf("invalid signature")
    }
    
    // Agent verified!
    return &AgentInfo{
        AgentID:   handshakeResp.AgentID,
        AgentType: handshakeResp.AgentType,
        IP:        ip,
        Verified:  true,
    }, nil
}
```

### Agent Side (Python)

```python
class AIPHandshakeHandler:
    def __init__(self, agent_id: str, agent_secret: str):
        self.agent_id = agent_id
        self.agent_secret = agent_secret
    
    async def handle_handshake(self, request: dict) -> dict:
        """Handle incoming handshake challenge."""
        challenge = request['challenge']
        timestamp = request['timestamp']
        nonce = request['nonce']
        
        # Compute signature
        message = f"{challenge}:{timestamp}:{nonce}"
        signature = hmac.new(
            self.agent_secret.encode(),
            message.encode(),
            hashlib.sha256
        ).hexdigest()
        
        return {
            'signature': signature,
            'agent_id': self.agent_id,
            'agent_type': 'darci-python',
            'agent_version': '1.0.0',
            'capabilities': ['task-execution', 'notebook']
        }
```

---

## HTTP Endpoint Specification

### `/aip/handshake` (Agent)

**Method:** POST

**Request:**
```json
{
  "challenge": "abc123...",
  "timestamp": "2026-03-08T10:00:00Z",
  "nonce": "xyz789",
  "bridge_id": "bridge-alpha"
}
```

**Response (200 OK):**
```json
{
  "signature": "hmac-result",
  "agent_id": "darci-python-001",
  "agent_type": "darci-python"
}
```

**Response (404 Not Found):**
```json
{
  "error": "not_an_agent"
}
```

**Response (400 Bad Request):**
```json
{
  "error": "invalid_challenge",
  "message": "Challenge expired"
}
```

### `/aip/handshake-probe` (Bridge → Network)

**Method:** POST

Used to probe potential agents on network.

```json
{
  "target_ip": "100.64.1.23",
  "target_port": 9090,
  "challenge": "..."
}
```

---

## Security Considerations

### Replay Attack Prevention

1. **Nonces**: Track used nonces, reject duplicates
2. **Timestamps**: Reject challenges older than N seconds
3. **Challenge expiry**: Challenges valid for max 30 seconds

### Rate Limiting

- Max 10 handshake attempts per minute per IP
- Exponential backoff for failed attempts

### Secret Management

- Store secrets encrypted at rest
- Rotate secrets periodically
- Use different secrets per agent

### Network Security

- Only probe IPs on Tailscale tailnet
- Respect Tailscale ACLs
- Log all handshake attempts

---

## Error Codes

| Code | HTTP Status | Meaning |
|------|-------------|---------|
| `HANDSHAKE-001` | 400 | Invalid challenge format |
| `HANDSHAKE-002` | 400 | Challenge expired |
| `HANDSHAKE-003` | 400 | Nonce already used |
| `HANDSHAKE-004` | 401 | Invalid signature |
| `HANDSHAKE-005` | 403 | Agent not in whitelist |
| `HANDSHAKE-006` | 404 | Not an agent endpoint |
| `HANDSHAKE-007` | 429 | Rate limit exceeded |
| `HANDSHAKE-008` | 500 | Internal error |

---

## Integration with AIP

The handshake is **Step 0** in the AIP flow:

```
1. Handshake (verify identity)
   ↓
2. Registration (submit capabilities)
   ↓
3. Approval (admin approves)
   ↓
4. Heartbeat (maintain active status)
```

### Combined Flow

```bash
# 1. Bridge discovers IP via Tailscale status
# 2. Bridge sends handshake challenge
curl -X POST http://100.64.1.23:9090/aip/handshake \
  -d '{"challenge": "...", "timestamp": "...", "nonce": "..."}'

# 3. If valid response received, agent is verified
# 4. Agent should also register proactively
curl -X POST http://bridge:8001/aip/register \
  -d '{"agent_id": "...", ...}'

# 5. Admin approves registration
taila2a aip approve darci-python-001
```

---

## Testing

### Test Cases

1. **Valid handshake**: Correct signature → 200 OK
2. **Invalid signature**: Wrong secret → 401
3. **Expired challenge**: Old timestamp → 400
4. **Duplicate nonce**: Replay attempt → 400
5. **Unknown agent**: No matching secret → 403
6. **Non-agent endpoint**: Random IP → 404 or timeout
7. **Rate limiting**: Too many attempts → 429

### Integration Test

```python
async def test_handshake_flow():
    # Setup
    agent = DarCIAgent(agent_id="test-001", secret="test-secret")
    bridge = Bridge()
    
    # Start handshake
    result = await bridge.send_handshake(agent.ip, agent.port)
    
    # Verify
    assert result.verified == True
    assert result.agent_id == "test-001"
```

---

## Related Documents

- [[AIP Specification](./aip-agent-identification-protocol.md)]
- [[Agent Communication](./agent-communication.md)]
- [[Security Hardening](./security-hardening.md)]
