# Task 003: Message Persistence (WAL)

## Priority
**P1** - High

## Status
⏳ Pending

## Objective
Implement Write-Ahead Log (WAL) for message persistence in Taila2a event bus, enabling message durability and replay capability.

## Background
Without persistence, messages are lost on restart. A WAL provides:
- Durability: Messages survive restarts
- Replay: Consumers can reprocess historical messages
- Recovery: Event bus can rebuild state from log

## Requirements

### Core Functionality
- [ ] Append-only log file per partition
- [ ] Segment-based log rotation
- [ ] Index files for fast offset lookup
- [ ] Log compaction for cleanup
- [ ] Crash recovery

### WAL Operations
- [ ] Append message to log
- [ ] Read message at offset
- [ ] Read range of messages
- [ ] Truncate log (compaction)
- [ ] Sync log to disk

### API
- [ ] `Append(topic, partition, message) -> offset`
- [ ] `Read(topic, partition, offset) -> message`
- [ ] `ReadRange(topic, partition, start, end) -> []message`
- [ ] `Truncate(topic, partition, offset) -> error`
- [ ] `Recover() -> error`

## Technical Specification

### Log Structure
```
data/
├── agent.requests/
│   ├── 0/
│   │   ├── 00000000000000000000.log
│   │   ├── 00000000000000000000.index
│   │   └── 00000000000000000000.timeindex
│   │   ├── 00000000000000001000.log
│   │   └── ...
│   └── 1/
│       └── ...
└── agent.responses/
    └── ...
```

### Segment Format
```
┌─────────────────────────────────────────┐
│            Segment File                 │
├─────────────────────────────────────────┤
│ [Message][Message][Message]...          │
│                                         │
│ Each Message:                           │
│ ┌─────────────────────────────────┐    │
│ │ Offset (8 bytes)                │    │
│ │ Size (4 bytes)                  │    │
│ │ CRC (4 bytes)                   │    │
│ │ Key Length (4 bytes)            │    │
│ │ Key (variable)                  │    │
│ │ Value (variable)                │    │
│ │ Timestamp (8 bytes)             │    │
│ └─────────────────────────────────┘    │
└─────────────────────────────────────────┘
```

### Index Format
```
┌─────────────────────────────────────────┐
│            Index File                   │
├─────────────────────────────────────────┤
│ [RelativeOffset][Position]...           │
│                                         │
│ Each Entry:                             │
│ ┌─────────────────────────────────┐    │
│ │ Relative Offset (4 bytes)       │    │
│ │ Position in Log (4 bytes)       │    │
│ └─────────────────────────────────┘    │
└─────────────────────────────────────────┘
```

### WAL Interface
```go
type WAL interface {
    // Append writes a message and returns its offset
    Append(topic string, partition int, data []byte) (int64, error)
    
    // Read reads a message at the given offset
    Read(topic string, partition int, offset int64) ([]byte, error)
    
    // ReadRange reads messages in offset range
    ReadRange(topic string, partition int, start, end int64) ([][]byte, error)
    
    // Truncate removes messages before offset
    Truncate(topic string, partition int, offset int64) error
    
    // Sync flushes pending writes to disk
    Sync() error
    
    // Close closes the WAL
    Close() error
}
```

## Acceptance Criteria

### Functional
- [ ] Messages are persisted to disk
- [ ] Messages can be read by offset
- [ ] Log segments rotate at configured size
- [ ] Index enables fast offset lookup
- [ ] Recovery rebuilds state from log

### Non-Functional
- [ ] Append latency < 5ms (p99) with fsync
- [ ] Read latency < 1ms (p99)
- [ ] Support 10GB+ logs
- [ ] Crash recovery < 30 seconds

## Testing Requirements

### Unit Tests
- [ ] Test append operations
- [ ] Test read operations
- [ ] Test index building
- [ ] Test segment rotation
- [ ] Test truncation

### Integration Tests
- [ ] Test crash recovery
- [ ] Test replay from beginning
- [ ] Test concurrent append/read
- [ ] Test large message handling

### Chaos Tests
- [ ] Test recovery after power failure
- [ ] Test recovery after disk full
- [ ] Test corrupted segment handling

## Implementation Steps

1. **Phase 1: Basic WAL** (3 days)
   - Implement append-only log
   - Add message format
   - Implement basic read

2. **Phase 2: Segmentation** (2 days)
   - Implement segment rotation
   - Add segment management
   - Handle segment cleanup

3. **Phase 3: Indexing** (2 days)
   - Implement index files
   - Add time index
   - Optimize lookups

4. **Phase 4: Recovery** (2 days)
   - Implement crash recovery
   - Add log validation
   - Handle corruption

5. **Phase 5: Compaction** (2 days)
   - Implement log compaction
   - Add retention policies
   - Background cleanup

6. **Phase 6: Testing** (2 days)
   - Unit tests
   - Integration tests
   - Chaos tests

## Files to Create/Modify

### Create
- `taila2a/internal/services/eventbus/wal/wal.go` - WAL interface
- `taila2a/internal/services/eventbus/wal/segment.go` - Segment management
- `taila2a/internal/services/eventbus/wal/index.go` - Index files
- `taila2a/internal/services/eventbus/wal/message.go` - Message format
- `taila2a/internal/services/eventbus/wal/recovery.go` - Recovery logic
- `taila2a/internal/services/eventbus/wal/wal_test.go` - Tests

### Modify
- `taila2a/internal/services/eventbus/eventbus.go` - Integrate WAL
- `taila2a/internal/services/eventbus/partition.go` - Use WAL for storage

## References
- [Kafka Log Structure](https://kafka.apache.org/documentation/#storage)
- [Write-Ahead Log Pattern](https://en.wikipedia.org/wiki/Write-ahead_logging)
- [Task 001: Event Bus Core](task_001_taila2a_eventbus.md)

## Assignment
**Agent:** [Unassigned]  
**Assigned:** [Date]  
**Due:** [Date]  

## Progress Log
- [Date]: [Update]
