# Engineering Notebook - 2026-03-07

## Summary
Built out private agent-to-agent bridge capabilities for embedded tsnet usage and documented the production-oriented setup path.

## Features Introduced
- Added optional bridge-to-bridge HMAC signing and verification using `BRIDGE_HMAC_SECRET`.
- Added `POST /a2a/jsonrpc` compatibility endpoint with methods:
  - `agent.card`
  - `message.send`
- Added ACL starter policy example for bridge-only ingress controls.
- Added runnable `.env` profile for agent-specific bridge deployment (`nanobot-bridge`).

## Features Edited
- Updated bridge docs for security hardening and compatibility endpoint usage.
- Updated runtime env examples to include shared-secret auth setting.

## Files Changed
- `tailscale-app/bridge/main.go`
- `tailscale-app/.env.bridge.example`
- `tailscale-app/.env`
- `tailscale-app/README.md`
- `tailscale-app/docs/agent-communication.md`
- `tailscale-app/docs/tailscale-acl.example.json`

## Verification
- Formatted Go source with `gofmt`.
- Build validation passed with `go build ./...` from `tailscale-app`.

## Risks / Follow-ups
- HMAC currently protects integrity/authentication but does not include replay protection.
- Recommended next hardening step: add timestamp + nonce checking and expiration window.
- Add ACL policy into live tailnet policy before production rollout.

## TODO (Deferred)
- Implement bridge-level file sharing over tailnet with streaming upload/download endpoints.
- Add file metadata + SHA256 integrity verification.
- Add chunked upload + resume support for large files.
- Reuse `BRIDGE_HMAC_SECRET` for authenticated file transfer requests.
