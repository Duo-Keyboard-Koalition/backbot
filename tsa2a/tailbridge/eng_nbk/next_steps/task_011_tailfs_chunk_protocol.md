# Task 011: Chunk Transfer Protocol

## Priority
**P0** - Critical

## Status
⏳ Pending

## Objective
Implement the chunked file transfer protocol for TailFS, enabling reliable transfer of large files over the Tailscale network.

## Background
Large files must be split into chunks for reliable transfer. This task implements the protocol for chunking, transmitting, and reassembling files.

## Requirements

### Chunking
- [ ] Split files into configurable chunk sizes (default 1MB)
- [ ] Generate chunk metadata (index, hash, size)
- [ ] Support parallel chunk transmission
- [ ] Handle out-of-order chunk arrival

### Transmission
- [ ] Send chunks over Tailscale connection
- [ ] Acknowledge received chunks
- [ ] Retry failed chunks
- [ ] Flow control (window-based)

### Reassembly
- [ ] Track received chunks
- [ ] Detect missing chunks
- [ ] Reassemble file from chunks
- [ ] Verify final file hash

### Protocol Messages
- [ ] TransferInit (sender → receiver)
- [ ] TransferAccept (receiver → sender)
- [ ] ChunkData (sender → receiver)
- [ ] ChunkAck (receiver → sender)
- [ ] TransferComplete (sender → receiver)
- [ ] TransferAbort (either party)

## Technical Specification

### Message Format
```go
// TransferInit sent by sender to start transfer
type TransferInit struct {
    TransferID    string            `json:"transfer_id"`
    FileName      string            `json:"file_name"`
    FileSize      int64             `json:"file_size"`
    FileHash      string            `json:"file_hash"`
    ChunkSize     int64             `json:"chunk_size"`
    TotalChunks   int               `json:"total_chunks"`
    Compression   bool              `json:"compression"`
    Encryption    bool              `json:"encryption"`
    Metadata      map[string]string `json:"metadata"`
}

// TransferAccept sent by receiver to accept
type TransferAccept struct {
    TransferID  string `json:"transfer_id"`
    Accepted    bool   `json:"accepted"`
    DestPath    string `json:"dest_path"`
    Message     string `json:"message,omitempty"`
}

// ChunkData contains file chunk
type ChunkData struct {
    TransferID  string `json:"transfer_id"`
    ChunkIndex  int    `json:"chunk_index"`
    Offset      int64  `json:"offset"`
    Data        []byte `json:"data"`
    ChunkHash   string `json:"chunk_hash"`
    Compressed  bool   `json:"compressed"`
}

// ChunkAck acknowledges chunk receipt
type ChunkAck struct {
    TransferID  string `json:"transfer_id"`
    ChunkIndex  int    `json:"chunk_index"`
    Received    bool   `json:"received"`
    Error       string `json:"error,omitempty"`
}

// TransferComplete signals completion
type TransferComplete struct {
    TransferID  string `json:"transfer_id"`
    Success     bool   `json:"success"`
    FinalHash   string `json:"final_hash"`
    Message     string `json:"message"`
}
```

### State Machine
```
SENDER:
┌─────────┐  Init   ┌──────────┐  Chunks ┌───────────┐
│  Idle   │────────▶│ Sending  │────────▶│ Complete  │
└─────────┘         └──────────┘         └───────────┘
     │                    │                    │
     │ Abort              │ Abort              │ Abort
     ▼                    ▼                    ▼
┌─────────────────────────────────────────────────┐
│                  Aborted                        │
└─────────────────────────────────────────────────┘

RECEIVER:
┌─────────┐  Accept ┌──────────┐  Chunks ┌───────────┐
│  Idle   │────────▶│ Receiving│────────▶│ Complete  │
└─────────┘         └──────────┘         └───────────┘
     │                    │                    │
     │ Reject             │ Abort              │ Abort
     ▼                    ▼                    ▼
┌─────────────────────────────────────────────────┐
│                  Aborted                        │
└─────────────────────────────────────────────────┘
```

### Protocol Interface
```go
type ChunkProtocol interface {
    // Initiate transfer
    InitTransfer(ctx context.Context, dest string, file string) (string, error)
    
    // Accept incoming transfer
    AcceptTransfer(transferID string, destPath string) error
    
    // Send chunks
    SendChunk(transferID string, chunk *ChunkData) error
    
    // Receive chunks
    ReceiveChunk(transferID string) (*ChunkData, error)
    
    // Complete transfer
    CompleteTransfer(transferID string) error
    
    // Abort transfer
    AbortTransfer(transferID string, reason string) error
}
```

## Acceptance Criteria

### Functional
- [ ] Files are chunked correctly
- [ ] Chunks transmitted reliably
- [ ] Missing chunks detected and retried
- [ ] Files reassembled correctly
- [ ] Final hash verified

### Non-Functional
- [ ] Transfer speed > 10MB/s on LAN
- [ ] Memory usage < 100MB regardless of file size
- [ ] Support files up to 100GB
- [ ] Handle network interruptions gracefully

## Testing Requirements

### Unit Tests
- [ ] Test chunking logic
- [ ] Test message encoding/decoding
- [ ] Test hash verification
- [ ] Test reassembly logic

### Integration Tests
- [ ] Test end-to-end transfer
- [ ] Test chunk retry
- [ ] Test transfer abort
- [ ] Test large file transfer

### Performance Tests
- [ ] Benchmark transfer speed
- [ ] Test with network latency
- [ ] Test with packet loss

## Implementation Steps

1. **Phase 1: Message Types** (1 day)
   - Define protocol messages
   - Implement encoding/decoding

2. **Phase 2: Chunking Logic** (2 days)
   - Implement file chunking
   - Add chunk hashing
   - Implement reassembly

3. **Phase 3: Transfer State Machine** (2 days)
   - Implement sender state machine
   - Implement receiver state machine
   - Add state transitions

4. **Phase 4: Transmission** (2 days)
   - Implement chunk sending
   - Implement acknowledgments
   - Add retry logic

5. **Phase 5: Testing** (2 days)
   - Unit tests
   - Integration tests
   - Performance tests

## Files to Create/Modify

### Create
- `tail-agent-file-send/internal/services/chunk_protocol.go` - Protocol implementation
- `tail-agent-file-send/internal/services/chunker.go` - File chunking
- `tail-agent-file-send/internal/services/reassembler.go` - File reassembly
- `tail-agent-file-send/internal/services/transfer_state.go` - State machine
- `tail-agent-file-send/internal/services/chunk_protocol_test.go` - Tests

### Modify
- `tail-agent-file-send/internal/services/transfer_service.go` - Integrate chunk protocol
- `tail-agent-file-send/internal/models/file_transfer.go` - Add chunk message types

## References
- [TailFS README](../tail-agent-file-send/README.md)
- [A2A Protocol](../eng_nbk/A2A_PROTOCOL.md)

## Assignment
**Agent:** [Unassigned]  
**Assigned:** [Date]  
**Due:** [Date]  

## Progress Log
- [Date]: [Update]
