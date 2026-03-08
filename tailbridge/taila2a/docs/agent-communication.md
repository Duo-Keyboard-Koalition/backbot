# Agent Communication Flow (Peer-to-Peer)

This document explains how local AI agents communicate through peer `Bridge` nodes on Tailnet using `tsnet`, with no central gateway in the message path.

## Core Requirement (Agent Unchanged)

- You do **not** modify the agent codebase.
- Each agent runs on its own host as-is.
- The host installs and runs a `Bridge` next to the agent.
- The bridge handles networking and peer routing.
- Agents communicate through their local bridges.

## Architecture Diagram

```mermaid
flowchart LR
  subgraph HostA[Host A]
    AgentA[Agent A\n(Unmodified box)]
    BridgeA[Bridge A\n(tsnet node)]
    AgentA -->|local HTTP /send| BridgeA
  end

  subgraph HostB[Host B]
    AgentB[Agent B\n(Unmodified box)]
    BridgeB[Bridge B\n(tsnet node)]
    BridgeB -->|local HTTP to agent| AgentB
  end

  BridgeA -->|Tailnet POST /inbound| BridgeB
```

## Peer Registration (Host-Level)

Registration is host-level, not agent-code-level. A practical model is:

1. Host starts bridge with unique `BRIDGE_NAME`.
2. Bridge identity exists on Tailnet via tsnet (`BRIDGE_NAME` as hostname).
3. Other bridges route directly to that name (`dest_node`).

You can optionally maintain a lightweight directory service for discovery, but message forwarding remains direct peer-to-peer.

## Roles

- `Agent`: Local application logic (for example, Python service).
- `Bridge`: Runs on each host next to the agent. Converts local HTTP into Tailnet peer routing.

## Endpoints

### Bridge local side

- `POST /send` on `BRIDGE_LOCAL_LISTEN` (default `127.0.0.1:8080`)
- Called by the local agent.
- Body is an `Envelope` JSON.

### Bridge peer inbound side

- `POST /inbound` on `<bridge-name>:BRIDGE_INBOUND_PORT` (default `8001`)
- Called by another bridge over Tailnet.
- Extracts `payload` and forwards to `LOCAL_AGENT_URL`.

## Envelope Contract

```json
{
  "source_node": "bridge-alpha",
  "dest_node": "bridge-beta",
  "payload": {
    "message": "hello from alpha"
  }
}
```

Fields:

- `source_node`: Sender bridge name (if omitted, bridge fills it with `BRIDGE_NAME`).
- `dest_node`: Destination bridge name (required).
- `payload`: Raw JSON forwarded to destination local agent (required).

## Step-by-Step Message Flow

Example: Agent A sends to Agent B.

1. Agent A sends envelope to `http://127.0.0.1:8080/send` (Bridge A).
2. Bridge A reads `dest_node=bridge-beta`.
3. Bridge A sends envelope directly to `http://bridge-beta:8001/inbound` over Tailnet.
4. Bridge B receives envelope on `/inbound`.
5. Bridge B forwards `payload` to local Agent B at `LOCAL_AGENT_URL`.
6. Agent B responds to Bridge B.
7. Bridge B returns response to Bridge A.
8. Bridge A returns response to Agent A.

## Sequence Diagram

```text
Agent A -> Bridge A (/send)
Bridge A -> Bridge B (/inbound) [Tailnet]
Bridge B -> Agent B (LOCAL_AGENT_URL)
Agent B -> Bridge B -> Bridge A -> Agent A
```

## Example Curl From Source Agent

```bash
curl -sS http://127.0.0.1:8080/send \
  -H 'content-type: application/json' \
  -d '{
    "source_node": "bridge-alpha",
    "dest_node": "bridge-beta",
    "payload": {
      "type": "task",
      "message": "run health check"
    }
  }'
```

## Required Runtime Settings (Bridge)

- `TS_AUTHKEY`
- `BRIDGE_NAME` (example `bridge-alpha`)
- `TSNET_STATE_DIR`
- `BRIDGE_INBOUND_PORT` (default `8001`)
- `PEER_BRIDGE_INBOUND_PORT` (default `8001`, used for outbound to peers)
- `BRIDGE_LOCAL_LISTEN` (default `127.0.0.1:8080`)
- `LOCAL_AGENT_URL` (example `http://127.0.0.1:9090/api`)

## Error Handling Behavior

- Bridge returns `502` if destination bridge is unreachable.
- Destination bridge returns `502` if destination local agent is unreachable.
- Invalid envelope returns `400`.

## Operational Notes

- Keep a persistent `TSNET_STATE_DIR` to preserve node identity.
- Use distinct `BRIDGE_NAME` per host.
- Use Tailscale ACLs/tags to restrict which peer bridges can reach `/inbound`.
- Keep payload small (current body limit is 1 MiB).

## Suggested Next Hardening

- Authenticate envelopes (HMAC signature).
- Validate source identity against Tailnet peer identity.
- Add request IDs and structured logs for end-to-end tracing.
- Add retries/backoff for transient peer delivery failures.
