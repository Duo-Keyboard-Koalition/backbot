# Task 012: Transfer Resume Support

## Priority
**P0** - Critical

## Status
⏳ Pending

## Objective
Implement transfer resume capability for TailFS, allowing failed transfers to continue from the point of failure rather than restarting from the beginning.

## Background
Large file transfers can fail due to network issues, system crashes, or user interruption. Resume support saves transfer state and enables continuation, saving bandwidth and time.

## Requirements

### State Persistence
- [ ] Save transfer state to disk
- [ ] Track received chunks
- [ ] Store partial file data
- [ ] Maintain transfer metadata

### Resume Protocol
- [ ] Resume request with last received chunk
- [ ] Sender validates resume request
- [ ] Resume from correct offset
- [ ] Verify resumed transfer integrity

### Recovery
- [ ] Detect incomplete transfers on startup
- [ ] Offer resume option
- [ ] Clean up abandoned transfers
- [ ] Handle concurrent resume attempts

### API
- [ ] `GetTransferState(transferID) -> TransferState`
- [ ] `SaveTransferState(state)`
- [ ] `ResumeTransfer(transferID) -> error`
- [ ] `ListResumableTransfers() -> []TransferState`
- [ ] `CleanupAbandonedTransfers() -> int`

## Technical Specification

### Transfer State Format
```go
type TransferState struct {
    // Identity
    TransferID   string    `json:"transfer_id"`
    FileName     string    `json:"file_name"`
    FileSize     int64     `json:"file_size"`
    FileHash     string    `json:"file_hash"`
    
    // Progress
    TotalChunks  int       `json:"total_chunks"`
    ReceivedChunks []int   `json:"received_chunks"`
    BytesReceived int64    `json:"bytes_received"`
    
    // State
    Status       string    `json:"status"`
    StartedAt    time.Time `json:"started_at"`
    UpdatedAt    time.Time `json:"updated_at"`
    
    // Paths
    SourcePath   string    `json:"source_path,omitempty"`
    DestPath     string    `json:"dest_path,omitempty"`
    
    // Checkpoint
    LastChunkIndex int     `json:"last_chunk_index"`
    LastChunkHash  string  `json:"last_chunk_hash"`
}

func (ts *TransferState) PercentComplete() float64 {
    return float64(ts.BytesReceived) * 100.0 / float64(ts.FileSize)
}

func (ts *TransferState) MissingChunks() []int {
    // Calculate missing chunk indices
}
```

### Resume Message Flow
```
RECEIVER (wants to resume):
  ┌────────────────────────────────────────┐
  │ 1. Load saved transfer state           │
  │ 2. Send ResumeRequest with:            │
  │    - TransferID                        │
  │    - LastReceivedChunk                 │
  │    - ReceivedChunkBitmap               │
  └────────────────────────────────────────┘
                    │
                    ▼
SENDER (validates resume):
  ┌────────────────────────────────────────┐
  │ 3. Validate transfer exists            │
  │ 4. Verify sender still has file        │
  │ 5. Send ResumeAccept or ResumeReject   │
  └────────────────────────────────────────┘
                    │
                    ▼
RESUME TRANSFER:
  ┌────────────────────────────────────────┐
  │ 6. Sender transmits missing chunks     │
  │ 7. Receiver validates and appends      │
  │ 8. Complete transfer normally          │
  └────────────────────────────────────────┘
```

### Resume Messages
```go
type ResumeRequest struct {
    TransferID       string `json:"transfer_id"`
    LastChunkIndex   int    `json:"last_chunk_index"`
    ReceivedBitmap   []byte `json:"received_bitmap"` // Bitmask of received chunks
    LastChunkHash    string `json:"last_chunk_hash"`
}

type ResumeAccept struct {
    TransferID     string `json:"transfer_id"`
    NextChunkIndex int    `json:"next_chunk_index"`
    MissingChunks  []int  `json:"missing_chunks"`
}

type ResumeReject struct {
    TransferID string `json:"transfer_id"`
    Reason     string `json:"reason"`
    // Possible reasons:
    // - "transfer_not_found"
    // - "file_no_longer_available"
    // - "invalid_checkpoint"
    // - "transfer_too_old"
}
```

## Acceptance Criteria

### Functional
- [ ] Transfer state saved periodically
- [ ] Resume request handled correctly
- [ ] Missing chunks retransmitted
- [ ] Resumed transfer verified
- [ ] Abandoned transfers cleaned up

### Non-Functional
- [ ] State save overhead < 1% of transfer time
- [ ] Resume decision < 100ms
- [ ] Support resuming after days/weeks
- [ ] Handle corrupted state files gracefully

## Testing Requirements

### Unit Tests
- [ ] Test state serialization
- [ ] Test bitmap operations
- [ ] Test missing chunk calculation
- [ ] Test state recovery

### Integration Tests
- [ ] Test resume after interruption
- [ ] Test resume with corrupted chunks
- [ ] Test concurrent resume attempts
- [ ] Test abandoned transfer cleanup

### Chaos Tests
- [ ] Test resume after crash
- [ ] Test resume with network partition
- [ ] Test resume with modified source file

## Implementation Steps

1. **Phase 1: State Persistence** (2 days)
   - Define transfer state structure
   - Implement save/load functions
   - Add periodic checkpointing

2. **Phase 2: Resume Protocol** (2 days)
   - Implement resume messages
   - Add resume request handling
   - Implement validation logic

3. **Phase 3: Recovery** (2 days)
   - Detect incomplete transfers
   - Implement resume UI/CLI
   - Add cleanup logic

4. **Phase 4: Testing** (2 days)
   - Interruption tests
   - Crash recovery tests
   - Documentation

## Files to Create/Modify

### Create
- `tail-agent-file-send/internal/services/resume.go` - Resume logic
- `tail-agent-file-send/internal/services/state_store.go` - State persistence
- `tail-agent-file-send/internal/services/resume_test.go` - Tests

### Modify
- `tail-agent-file-send/internal/services/transfer_service.go` - Integrate resume
- `tail-agent-file-send/internal/models/file_transfer.go` - Add resume types
- `tail-agent-file-send/cmd/tailfs/main.go` - Add resume command

## References
- [Task 011: Chunk Protocol](task_011_tailfs_chunk_protocol.md)
- [TailFS README](../tail-agent-file-send/README.md)

## Assignment
**Agent:** [Unassigned]  
**Assigned:** [Date]  
**Due:** [Date]  

## Progress Log
- [Date]: [Update]
