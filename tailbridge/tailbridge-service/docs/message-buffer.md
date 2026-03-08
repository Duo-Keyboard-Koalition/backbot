# Message Buffer (Persistent Queue with Retry)

The message buffer provides reliable message delivery with automatic retries and persistent storage.

## Overview

When a message cannot be delivered immediately (destination bridge unreachable, network issues, etc.), it is automatically queued in a persistent buffer. The buffer service will retry delivery with exponential backoff until successful or max retries is exceeded.

## Features

- **Persistent Storage**: Messages survive bridge restarts
- **Automatic Retry**: Exponential backoff for failed deliveries
- **Background Processing**: Non-blocking delivery attempts
- **Monitoring Endpoints**: Track buffer health and message status
- **Manual Retry**: Retry failed messages on demand

## Configuration

The buffer is configured automatically when the bridge starts:

```go
bufferConfig := &buffer.BufferServiceConfig{
    DataDir:         filepath.Join(cfg.StateDir, "buffer"),
    RetryConfig:     buffer.DefaultRetryConfig(),
    ProcessInterval: 5 * time.Second,
    HTTPTimeout:     20 * time.Second,
    PeerInboundPort: cfg.PeerInboundPort,
}
```

### Retry Configuration

Default retry settings:

| Setting | Default | Description |
|---------|---------|-------------|
| `MaxRetries` | 5 | Maximum delivery attempts |
| `InitialDelay` | 1s | Delay before first retry |
| `MaxDelay` | 5m | Maximum delay between retries |
| `Multiplier` | 2.0 | Exponential backoff multiplier |

Retry schedule: 1s → 2s → 4s → 8s → 16s → 32s → ... (capped at 5m)

## Message Lifecycle

```
┌─────────────┐
│   Pending   │ ──→ Delivery Attempt
└─────────────┘
       │
       ├── Success ──→ Delivered
       │
       └── Failure
            │
            ├── Retries < Max ──→ Retrying (scheduled)
            │
            └── Retries >= Max ──→ Failed
```

### Status Values

| Status | Description |
|--------|-------------|
| `pending` | Message waiting to be delivered |
| `retrying` | Delivery failed, scheduled for retry |
| `delivered` | Successfully delivered |
| `failed` | Permanently failed (max retries exceeded) |

## API Endpoints

All buffer endpoints are available on the bridge's tailnet interface.

### Get Buffer Statistics

```bash
curl http://<bridge-host>:8001/buffer/stats
```

Response:
```json
{
  "total_messages": 10,
  "pending_count": 3,
  "retrying_count": 2,
  "failed_count": 1,
  "delivered_count": 4,
  "oldest_message": "2026-03-07T12:00:00Z",
  "newest_message": "2026-03-07T12:30:00Z"
}
```

### List Messages

```bash
# List pending and failed messages (default)
curl http://<bridge-host>:8001/buffer/messages

# List all messages
curl http://<bridge-host>:8001/buffer/messages?status=all

# List only pending messages
curl http://<bridge-host>:8001/buffer/messages?status=pending

# List only failed messages
curl http://<bridge-host>:8001/buffer/messages?status=failed
```

Response:
```json
{
  "messages": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "created_at": "2026-03-07T12:00:00Z",
      "updated_at": "2026-03-07T12:05:00Z",
      "next_retry_at": "2026-03-07T12:10:00Z",
      "status": "retrying",
      "retry_count": 2,
      "max_retries": 5,
      "last_error": "destination bridge unreachable: dial tcp: i/o timeout",
      "envelope": {
        "source_node": "bridge-alpha",
        "dest_node": "bridge-beta",
        "payload": {"type": "task", "message": "hello"}
      }
    }
  ],
  "count": 1
}
```

### Retry Failed Message

```bash
curl -X POST http://<bridge-host>:8001/buffer/retry \
  -H 'content-type: application/json' \
  -d '{"message_id": "550e8400-e29b-41d4-a716-446655440000"}'
```

Response:
```json
{
  "status": "retrying",
  "message_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Clear Delivered Messages

```bash
curl -X POST http://<bridge-host>:8001/buffer/clear
```

Response:
```json
{
  "cleared": 42
}
```

## Message Flow

### Successful Delivery (Fast Path)

1. Agent sends message to `POST /send`
2. Bridge attempts immediate delivery
3. Destination responds successfully
4. Message delivered (no buffering needed)

### Failed Delivery (Buffered Path)

1. Agent sends message to `POST /send`
2. Bridge attempts immediate delivery
3. Destination unreachable → delivery fails
4. Message saved to persistent buffer
5. Background processor schedules retry
6. Retry attempts with exponential backoff
7. Eventually succeeds or moves to failed state

## Storage

Messages are stored as JSON files in `<state_dir>/buffer/`:

```
~/.tailtalkie/state/
└── buffer/
    ├── 550e8400-e29b-41d4-a716-446655440000.json
    ├── 6ba7b810-9dad-11d1-80b4-00c04fd430c8.json
    └── ...
```

Each file contains the full message with metadata:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "created_at": "2026-03-07T12:00:00Z",
  "updated_at": "2026-03-07T12:05:00Z",
  "next_retry_at": "2026-03-07T12:10:00Z",
  "status": "retrying",
  "retry_count": 2,
  "max_retries": 5,
  "last_error": "destination bridge unreachable",
  "envelope": {
    "source_node": "bridge-alpha",
    "dest_node": "bridge-beta",
    "payload": {"type": "task", "message": "hello"}
  }
}
```

## Error Handling

### Common Errors

| Error | Cause | Resolution |
|-------|-------|------------|
| `destination bridge unreachable` | Target bridge offline or network issue | Automatic retry, check bridge status |
| `local agent unreachable` | Destination's local agent down | Destination bridge will retry |
| `invalid envelope json` | Malformed request | Fix request format |
| `message not found` | Invalid message ID for retry | Check message ID |

### Max Retries Exceeded

When a message exceeds max retries:
1. Status changes to `failed`
2. No more automatic retries
3. Message persists for manual inspection
4. Use `/buffer/retry` to manually retry

## Monitoring

### Health Checks

Monitor buffer health:

```bash
# Check for growing backlog
curl -s http://<bridge-host>:8001/buffer/stats | jq '.pending_count'

# Alert if failed messages accumulate
curl -s http://<bridge-host>:8001/buffer/stats | jq '.failed_count'

# Check oldest pending message age
curl -s http://<bridge-host>:8001/buffer/stats | jq '.oldest_message'
```

### Recommended Alerts

- `pending_count > 100`: Delivery backlog growing
- `failed_count > 10`: Multiple permanent failures
- `oldest_message` age > 1 hour: Stuck messages

## Best Practices

1. **Monitor buffer stats regularly** - Catch delivery issues early
2. **Clear delivered messages** - Prevent storage bloat
3. **Review failed messages** - Identify systemic issues
4. **Adjust retry config** - Tune for your network conditions
5. **Use unique bridge names** - Ensure proper routing

## Example: Sending with Buffer Awareness

```bash
# Send a message
RESPONSE=$(curl -sS http://127.0.0.1:8080/send \
  -H 'content-type: application/json' \
  -d '{
    "source_node": "bridge-alpha",
    "dest_node": "bridge-beta",
    "payload": {"message": "hello"}
  }')

# Check if buffered
STATUS=$(echo $RESPONSE | jq -r '.status // "direct"')
if [ "$STATUS" = "buffered" ]; then
  MSG_ID=$(echo $RESPONSE | jq -r '.message_id')
  echo "Message buffered for retry: $MSG_ID"
  
  # Monitor delivery
  while true; do
    MSG=$(curl -s "http://<bridge>:8001/buffer/messages?status=all" | \
          jq ".messages[] | select(.id == \"$MSG_ID\")")
    STATUS=$(echo $MSG | jq -r '.status')
    echo "Status: $STATUS"
    [ "$STATUS" = "delivered" ] && break
    sleep 5
  done
fi
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Bridge Service                          │
│                                                              │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐  │
│  │  /send       │    │  /inbound    │    │  /buffer/*   │  │
│  │  Handler     │    │  Handler     │    │  Handlers    │  │
│  └──────┬───────┘    └──────┬───────┘    └──────┬───────┘  │
│         │                   │                   │          │
│         ▼                   ▼                   ▼          │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Buffer Service                           │  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │  │
│  │  │   Enqueue   │  │  Processor  │  │   Retry     │  │  │
│  │  │             │  │  (5s loop)  │  │   Logic     │  │  │
│  │  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  │  │
│  │         │                │                │          │  │
│  │         ▼                ▼                ▼          │  │
│  │  ┌───────────────────────────────────────────────┐   │  │
│  │  │          Persistent Store (JSON files)        │   │  │
│  │  └───────────────────────────────────────────────┘   │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```
