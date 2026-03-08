# tsA2A Architecture Overview

**Complete system architecture for agent identification without network scanning**

---

## System Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         TAILSCALE TAILNET                                │
│                                                                          │
│  ┌────────────────────────────────────────────────────────────┐         │
│  │  Host A: workstation-alpha                                  │         │
│  │                                                              │         │
│  │  ┌────────────────┐         ┌────────────────────┐         │         │
│  │  │  DarCI Agent   │         │   taila2a Bridge   │         │         │
│  │  │  (Python/Go)   │◀───────▶│   (tsnet node)     │         │         │
│  │  │                │  HTTP   │                    │         │         │
│  │  │  - Tasks       │  :9090  │  - /inbound :8001  │         │         │
│  │  │  - Notebook    │         │  - /aip/*   :8001  │         │         │
│  │  │  - File ops    │         │  - /send    :8080  │         │         │
│  │  └────────────────┘         │  - /handshake      │         │         │
│  │                              └─────────┬──────────┘         │         │
│  │                                        │                      │         │
│  │                              Tailnet IP: 100.64.1.23         │         │
│  │                              Hostname: bridge-alpha          │         │
│  └────────────────────────────────────────┼─────────────────────┘         │
│                                           │                                │
│                        ┌──────────────────┼──────────────────┐            │
│                        │                  │                  │            │
│                        ▼                  ▼                  ▼            │
│              ┌─────────────────┐ ┌─────────────────┐ ┌──────────────┐    │
│              │  bridge-beta    │ │  bridge-gamma   │ │  other-host  │    │
│              │  (DarCI agent)  │ │  (DarCI agent)  │ │  (ignored)   │    │
│              │  100.64.1.45    │ │  100.64.1.67    │ │  100.64.1.89 │    │
│              └─────────────────┘ └─────────────────┘ └──────────────┘    │
│                        ▲                  ▲                  │            │
│                        │                  │                  │            │
│                        └──────────────────┴──────────────────┘            │
│                                           │                                │
│                              Only registered bridges                       │
│                              communicate via tsA2A                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Protocol Stack

```
┌─────────────────────────────────────────────────────┐
│              Application Layer                       │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐ │
│  │   DarCI     │  │   Sentinel  │  │   Custom    │ │
│  │   Agent     │  │   Agent     │  │   Agent     │ │
│  └─────────────┘  └─────────────┘  └─────────────┘ │
└─────────────────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────┐
│         tsA2A Protocol Layer                         │
│  ┌──────────────────┐  ┌──────────────────┐        │
│  │   Handshake      │  │   Registration   │        │
│  │   (Challenge)    │  │   (AIP)          │        │
│  └──────────────────┘  └──────────────────┘        │
│  ┌──────────────────┐  ┌──────────────────┐        │
│  │   Heartbeat      │  │   Pairing        │        │
│  │   (Keep-alive)   │  │   (Bridge-Bridge)│        │
│  └──────────────────┘  └──────────────────┘        │
└─────────────────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────┐
│         Transport Layer (Tailscale)                  │
│  ┌────────────────────────────────────────────────┐ │
│  │  WireGuard encrypted tunnel over tailnet       │ │
│  └────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────┘
```

---

## Message Flow: Agent A → Agent B

```
┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐
│ Agent A  │    │Bridge A  │    │Bridge B  │    │ Agent B  │
│          │    │          │    │          │    │          │
│  1. Send │    │          │    │          │    │          │
│    task  │    │          │    │          │    │          │
│ ────────▶│    │          │    │          │    │          │
│          │    │          │    │          │    │          │
│          │ 2. Wrap    │    │          │    │          │
│          │    envelope│    │          │    │          │
│          │──────────▶│    │          │    │          │
│          │    │          │          │    │          │
│          │    │ 3. POST  │          │    │          │
│          │    │ /inbound │          │    │          │
│          │    │ (Tailnet)│          │    │          │
│          │    │─────────────────▶│    │          │
│          │    │          │    │ 4. Forward│    │          │
│          │    │          │    │  payload │    │          │
│          │    │          │    │─────────▶│    │          │
│          │    │          │    │          │ 5. Process   │
│          │    │          │    │          │    task      │
│          │    │          │    │          │              │
│          │    │          │    │ 6. Response│            │
│          │    │          │    │◀─────────│    │          │
│          │    │ 7. POST  │    │          │    │          │
│          │    │ response │    │          │    │          │
│          │    │◀─────────│    │          │    │          │
│          │ 8. Unwrap  │    │          │    │          │
│          │    response│    │          │    │          │
│ 9. Result│◀──────────│    │          │    │          │
│◀─────────│    │          │    │          │    │          │
└──────────┘    └──────────┘    └──────────┘    └──────────┘
```

---

## Handshake Flow (Agent Identification)

```
┌──────────┐                              ┌──────────┐
│ Bridge   │                              │ Agent    │
│          │                              │          │
│ Discover │                              │          │
│ IP via   │                              │          │
│ Tailscale│                              │          │
│ status   │                              │          │
│    │                                     │          │
│    │ 1. Generate challenge              │          │
│    │    nonce, timestamp                │          │
│    │                                     │          │
│    │ 2. POST /aip/handshake             │          │
│    │    {challenge, nonce, ts}          │          │
│    │──────────────────────────────────▶ │          │
│    │                                     │          │
│    │                          3. Compute HMAC       │
│    │                             signature =        │
│    │                             HMAC(challenge:    │
│    │                             nonce:ts, secret)  │
│    │                                     │          │
│    │ 4. Response                        │          │
│    │    {signature, agent_id, type}     │          │
│    │◀───────────────────────────────────│          │
│    │                                     │          │
│ 5. Verify signature                     │          │
│    - Lookup secret by agent_id          │          │
│    - Recompute HMAC                     │          │
│    - Compare (constant-time)            │          │
│    │                                     │          │
│    ✓ Valid → Agent verified!            │          │
│    ✗ Invalid → Ignore/log               │          │
│                                         │          │
└──────────┘                              └──────────┘
```

---

## Registration Flow

```
┌──────────┐    ┌──────────┐    ┌──────────┐
│ Agent    │    │ Bridge   │    │ Admin    │
│          │    │          │    │ (Human)  │
│          │    │          │    │          │
│ 1. POST  │    │          │    │          │
│    /aip/ │    │          │    │          │
│    register      │          │    │          │
│ ────────▶│    │          │    │          │
│          │    │          │    │          │
│          │ 2. Store as     │    │          │
│          │    "pending"    │    │          │
│          │    registry.json│    │          │
│          │    │          │    │          │
│          │    │ 3. Notify  │    │          │
│          │    │    admin   │    │          │
│          │    │───────────▶│    │          │
│          │    │          │    │          │
│          │    │ 4. Review │    │          │
│          │    │    pending │    │          │
│          │    │◀──────────│    │          │
│          │    │  taila2a  │    │          │
│          │    │  aip      │    │          │
│          │    │  pending  │    │          │
│          │    │          │    │          │
│          │    │ 5. Approve│    │          │
│          │    │    taila2a│    │          │
│          │    │  aip      │    │          │
│          │    │  approve  │    │          │
│          │    │◀──────────│    │          │
│          │    │          │    │          │
│          │ 6. Update     │    │          │
│          │    status →   │    │          │
│          │    "approved" │    │          │
│          │    │          │    │          │
│ 7. Start │    │          │    │          │
│    heart-│    │          │    │          │
│    beat  │    │          │    │          │
│ ────────▶│    │          │    │          │
│          │    │          │    │          │
└──────────┘    └──────────┘    └──────────┘
```

---

## Data Flow Diagram

```
┌──────────────────────────────────────────────────────────────┐
│                      File System                              │
│                                                               │
│  ~/.tailtalkie/                                              │
│  ├── config.json          # Bridge config                   │
│  └── state/                                                   │
│      ├── registry.json    # Agent registry                  │
│      │   └── registered_agents: []                          │
│      │   └── peer_bridges: []                               │
│      ├── agent_secrets.json  # HMAC secrets                 │
│      │   └── secrets: {agent_id: secret}                    │
│      └── buffer/            # Message buffer                │
│          └── pending_messages/                               │
└──────────────────────────────────────────────────────────────┘
                          ▲
                          │
┌─────────────────────────┼──────────────────────────────────┐
│                    Runtime Memory                           │
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │ AgentRegistry│  │HandshakeSvc  │  │ BufferSvc    │     │
│  │  - agents    │  │  - challenges│  │  - messages  │     │
│  │  - status    │  │  - nonces    │  │  - retries   │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│                                                             │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌──────────────────────────────────────────────────────────────┐
│                   HTTP Endpoints                              │
│                                                               │
│  :8001 (Tailnet)                    :8080 (Local)            │
│  ├── /inbound                       ├── /send                │
│  ├── /agents                        ├── /aip/register        │
│  ├── /aip/handshake                 ├── /aip/heartbeat       │
│  └── /aip/handshake-probe           └── /buffer/*            │
│                                                               │
└──────────────────────────────────────────────────────────────┘
```

---

## Security Boundaries

```
┌─────────────────────────────────────────────────────────┐
│                  Trust Boundary                          │
│                                                          │
│  ┌────────────────┐         ┌────────────────┐          │
│  │  Untrusted     │         │  Trusted       │          │
│  │  Network       │         │  Tailnet       │          │
│  │                │         │                │          │
│  │  - Internet    │         │  - bridge-α    │          │
│  │  - Random IPs  │         │  - bridge-β    │          │
│  │                │         │  - bridge-γ    │          │
│  └────────────────┘         └───────┬────────┘          │
│                                     │                    │
│                              ┌──────▼────────┐          │
│                              │  Verification │          │
│                              │  Layer        │          │
│                              │               │          │
│                              │  - Handshake  │          │
│                              │  - HMAC-SHA256│          │
│                              │  - Nonce check│          │
│                              │  - Expiry     │          │
│                              └───────┬────────┘          │
│                                      │                   │
│                              ┌──────▼────────┐          │
│                              │  Approved     │          │
│                              │  Agents Only  │          │
│                              └───────────────┘          │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

---

## Component Interaction Matrix

| Component | Registry | Handshake | AIP Handlers | Secrets | Discovery |
|-----------|----------|-----------|--------------|---------|-----------|
| Registry  | -        | Read      | Write        | Read    | Read      |
| Handshake | Write    | -         | Callback    | Read    | -         |
| AIP Handlers | Write | -         | -            | -       | -         |
| Secrets   | -        | Read      | Read         | -       | -         |
| Discovery | -        | Trigger   | -            | -       | -         |

---

## Deployment Topology

```
┌────────────────────────────────────────────────────────────┐
│                    Production Deployment                    │
│                                                             │
│  Site A                        Site B                       │
│  ┌──────────────┐             ┌──────────────┐             │
│  │  bridge-α    │◀──Tailnet──▶│  bridge-β    │             │
│  │  :8001       │  Encrypted  │  :8001       │             │
│  │  :8080       │  Tunnel     │  :8080       │             │
│  └──────┬───────┘             └──────┬───────┘             │
│         │                            │                      │
│  ┌──────▼───────┐             ┌──────▼───────┐             │
│  │ darci-py-001 │             │ darci-go-001 │             │
│  │ :9090        │             │ :9090        │             │
│  └──────────────┘             └──────────────┘             │
│                                                             │
│  Site C                        Site D                       │
│  ┌──────────────┐             ┌──────────────┐             │
│  │  bridge-γ    │◀──Tailnet──▶│  bridge-δ    │             │
│  │  :8001       │  Encrypted  │  :8001       │             │
│  └──────┬───────┘  Tunnel     └──────┬───────┘             │
│         │                            │                      │
│  ┌──────▼───────┐             ┌──────▼───────┐             │
│  │ sentinel-001 │             │ custom-001   │             │
│  └──────────────┘             └──────────────┘             │
│                                                             │
└────────────────────────────────────────────────────────────┘
```

---

## State Machine: Agent Lifecycle

```
     ┌─────────┐
     │  START  │
     └────┬────┘
          │
          │ Agent starts
          ▼
     ┌─────────┐
     │ UNKNOWN │◀────────────────────────┐
     └────┬────┘                         │
          │                              │
          │ Send handshake               │ Timeout/Retry
          ▼                              │
     ┌─────────┐     Invalid sig    ┌────┴─────┐
     │PENDING  │───────────────────▶│ REJECTED │
     │  (await │                    └──────────┘
     │  verify)│
     └────┬────┘
          │
          │ Valid signature
          ▼
     ┌─────────┐
     │VERIFIED │
     └────┬────┘
          │
          │ Register + Approve
          ▼
     ┌─────────┐     No heartbeat    ┌─────────┐
     │APPROVED │────────────────────▶│ OFFLINE │
     │(active) │                     └─────────┘
     └────┬────┘
          │
          │ Remove/Deregister
          ▼
     ┌─────────┐
     │REMOVED  │
     └─────────┘
```

---

This architecture ensures:
- ✅ **No network scanning** - Only Tailscale status queries
- ✅ **Secure identification** - HMAC-SHA256 challenge-response
- ✅ **Explicit approval** - Admin must approve all agents
- ✅ **Audit trail** - All events logged in registry
- ✅ **Scalable** - Works across multiple sites/bridges
