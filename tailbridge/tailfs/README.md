# TailFS - Secure File Transfer over Tailscale

**TailFS** (Tail Agent File Send) is a secure, peer-to-peer file transfer system built on Tailscale's zero-trust network. Send files directly between computers on your tailnet without any cloud intermediaries.

---

## Quick Start

### 1. Initialize Configuration

```bash
cd tail-agent-file-send
go run ./cmd/tailfs init
```

### 2. Start TailFS Service

```bash
go run ./cmd/tailfs run
```

### 3. Send a File

```bash
# CLI
go run ./cmd/tailfs send document.pdf tailfs-beta

# API
curl -X POST http://localhost:8081/send \
  -H 'Content-Type: application/json' \
  -d '{
    "file": "document.pdf",
    "destination": "tailfs-beta"
  }'
```

---

## Features

### Security

- **End-to-End Encryption** - Files encrypted in transit via WireGuard
- **Zero Trust** - No cloud servers, direct peer-to-peer
- **Tailscale Identity** - Each node authenticated via Tailscale
- **Optional File Encryption** - Additional encryption layer for sensitive files

### Performance

- **Chunked Transfer** - Large files split into manageable chunks
- **Compression** - Optional compression for faster transfers
- **Resume Support** - Failed transfers can resume from checkpoint
- **Parallel Transfers** - Multiple files simultaneously

### Discovery

- **Agent Phone Book** - Automatic discovery of tailfs agents
- **Capability Detection** - Know which agents can send/receive
- **Status Monitoring** - Real-time transfer progress

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         TailFS System                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────┐                           ┌──────────────┐   │
│  │   Sender     │                           │   Receiver   │   │
│  │   Agent      │                           │   Agent      │   │
│  │              │                           │              │   │
│  │ ┌──────────┐ │      Tailscale Tailnet    │ ┌──────────┐ │   │
│  │ │  File    │ │     (WireGuard Encrypted) │ │  File    │ │   │
│  │ │  Reader  │ │  ◀─────────────────────▶  │ │  Writer  │ │   │
│  │ └──────────┘ │                           │ └──────────┘ │   │
│  │      │       │                           │      │       │   │
│  │ ┌──────────┐ │                           │ ┌──────────┐ │   │
│  │  Chunk    │ │                           │  Chunk    │ │   │
│  │  Sender   │ │                           │  Receiver │ │   │
│  │ └──────────┘ │                           │ └──────────┘ │   │
│  └──────────────┘                           └──────────────┘   │
│         │                                         │            │
│         └──────────────────┬──────────────────────┘            │
│                            │                                   │
│                            ▼                                   │
│                  ┌──────────────────┐                         │
│                  │  Phone Book      │                         │
│                  │  (Discovery)     │                         │
│                  └──────────────────┘                         │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/send` | POST | Initiate file transfer |
| `/receive` | POST | Accept incoming file |
| `/progress` | GET | Get transfer progress |
| `/history` | GET | Get transfer history |
| `/agents` | GET | List available agents |

### Send File

```bash
POST http://localhost:8081/send
Content-Type: application/json

{
  "file": "/path/to/file.pdf",
  "destination": "tailfs-beta",
  "compress": true,
  "encrypt": true
}
```

Response:
```json
{
  "transfer_id": "uuid-123",
  "status": "pending",
  "file_size": 1048576
}
```

### Get Progress

```bash
GET http://localhost:8081/progress?transfer_id=uuid-123
```

Response:
```json
{
  "transfer_id": "uuid-123",
  "status": "sending",
  "bytes_sent": 524288,
  "bytes_total": 1048576,
  "percent_complete": "50.0%",
  "bytes_per_second": 1048576,
  "eta_seconds": 1
}
```

---

## CLI Commands

```bash
# Initialize
tailfs init

# Start service
tailfs run

# Send file
tailfs send <file> <destination-agent>

# List agents
tailfs list

# Check status
tailfs status

# Show version
tailfs version

# Help
tailfs help
```

---

## Configuration

Config file: `~/.tailfs/config.json`

```json
{
  "node_name": "tailfs-alpha",
  "auth_key": "tskey-auth-xxxxx",
  "state_dir": "~/.tailfs/state",
  "local_listen": "127.0.0.1:8081",
  "download_dir": "~/Downloads/tailfs"
}
```

---

## Integration with Taila2a

TailFS works seamlessly with Taila2a for agent discovery:

```bash
# Use taila2a phone book to find agents
curl http://localhost:8080/agents?capability=file_receive

# Send file to discovered agent
tailfs send document.pdf tailfs-beta
```

---

## Workflows

### Basic File Transfer

```
1. Sender discovers receiver via phone book
2. Sender initiates transfer request
3. Receiver accepts (manual or auto-accept)
4. File chunked and encrypted
5. Chunks sent via Tailscale
6. Receiver verifies and reassembles
7. Transfer complete notification
```

### Bulk Transfer

```
1. Select multiple files
2. Create transfer batch
3. Parallel chunked transfers
4. Progress aggregation
5. Batch completion report
```

### Scheduled Transfer

```
1. Queue transfer for later
2. Wait for agent online
3. Auto-initiate when available
4. Progress notification
```

---

## Build

```bash
cd tail-agent-file-send
go build -o tailfs ./cmd/tailfs
```

---

## License

Same as the parent project.
