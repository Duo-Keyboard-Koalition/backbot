# Task 002: Consumer Group Protocol

## Priority
**P0** - Critical

## Status
⏳ Pending

## Objective
Implement consumer group protocol for Taila2a event bus, enabling multiple consumers to coordinate message consumption from shared topics.

## Background
Consumer groups allow multiple agents to share the workload of consuming messages from topics. Each message is delivered to only one consumer within a group, enabling parallel processing while maintaining order within partitions.

## Requirements

### Core Functionality
- [ ] Consumer group creation and management
- [ ] Partition assignment to consumers
- [ ] Offset tracking per consumer group
- [ ] Consumer heartbeat mechanism
- [ ] Automatic rebalancing on join/leave

### Protocol Messages
- [ ] JoinGroup request/response
- [ ] Heartbeat request/response
- [ ] SyncGroup for partition assignment
- [ ] LeaveGroup for graceful exit
- [ ] OffsetCommit for progress tracking

### API
- [ ] `JoinGroup(groupID, topics) -> memberID`
- [ ] `LeaveGroup(groupID, memberID)`
- [ ] `Heartbeat(groupID, memberID) -> error`
- [ ] `CommitOffset(groupID, topic, partition, offset)`
- [ ] `GetPartitionAssignments(groupID) -> map[memberID][]partition`

## Technical Specification

### Consumer Group State Machine
```
┌─────────┐     Join      ┌──────────┐    Sync     ┌──────────┐
│  Empty  │──────────────▶│ Preparing│────────────▶│ Stable   │
└─────────┘               └──────────┘             └──────────┘
     ▲                          │                      │
     │                          │ Leave/Timeout        │ Leave/Timeout
     └──────────────────────────┴──────────────────────┘
```

### Message Types
```go
type JoinGroupRequest struct {
    GroupID     string
    MemberID    string
    Topics      []string
    Capabilities []string
}

type JoinGroupResponse struct {
    MemberID    string
    GenerationID int
    LeaderID    string
    Members     []MemberInfo
}

type HeartbeatRequest struct {
    GroupID      string
    MemberID     string
    GenerationID int
}

type SyncGroupRequest struct {
    GroupID      string
    MemberID     string
    GenerationID int
    Assignments  map[string][]PartitionAssignment
}

type OffsetCommitRequest struct {
    GroupID   string
    MemberID  string
    Topic     string
    Partition int
    Offset    int64
}
```

### Partition Assignment Strategy
```go
type PartitionAssigner interface {
    Assign(
        partitions map[string][]int,
        members []MemberInfo,
    ) map[string]map[string][]int
}

// Implement RangeAssigner and RoundRobinAssigner
```

## Acceptance Criteria

### Functional
- [ ] Consumers can join groups
- [ ] Partitions are assigned fairly
- [ ] Offsets are tracked correctly
- [ ] Rebalancing occurs on member changes
- [ ] Heartbeats detect failed consumers

### Non-Functional
- [ ] Rebalance completes in < 5 seconds
- [ ] Support 100+ consumers per group
- [ ] Offset commits are durable
- [ ] No duplicate message delivery during rebalance

## Testing Requirements

### Unit Tests
- [ ] Test join/leave group
- [ ] Test partition assignment algorithms
- [ ] Test offset commit/retrieve
- [ ] Test heartbeat timeout

### Integration Tests
- [ ] Test multi-consumer group
- [ ] Test rebalancing scenarios
- [ ] Test offset management end-to-end

## Implementation Steps

1. **Phase 1: Group Management** (2 days)
   - Create consumer group state machine
   - Implement join/leave protocol
   - Add member tracking

2. **Phase 2: Partition Assignment** (3 days)
   - Implement partition assigner interface
   - Add RangeAssigner strategy
   - Add RoundRobinAssigner strategy
   - Implement SyncGroup protocol

3. **Phase 3: Offset Management** (2 days)
   - Implement offset storage
   - Add commit/retrieve APIs
   - Handle offset resets

4. **Phase 4: Heartbeat & Failure Detection** (2 days)
   - Implement heartbeat protocol
   - Add timeout detection
   - Handle automatic rebalancing

5. **Phase 5: Testing** (2 days)
   - Unit tests
   - Integration tests
   - Documentation

## Files to Create/Modify

### Create
- `taila2a/internal/services/eventbus/consumer_group.go` - Group management
- `taila2a/internal/services/eventbus/partition_assigner.go` - Assignment strategies
- `taila2a/internal/services/eventbus/heartbeat.go` - Heartbeat mechanism
- `taila2a/internal/services/eventbus/offset_manager.go` - Offset tracking
- `taila2a/internal/services/eventbus/consumer_group_test.go` - Tests

### Modify
- `taila2a/internal/services/eventbus/eventbus.go` - Integrate consumer groups

## References
- [Kafka Consumer Groups](https://kafka.apache.org/documentation/#intro_consumers)
- [Task 001: Event Bus Core](task_001_taila2a_eventbus.md)
- [A2A Protocol](../eng_nbk/A2A_PROTOCOL.md)

## Assignment
**Agent:** [Unassigned]  
**Assigned:** [Date]  
**Due:** [Date]  

## Progress Log
- [Date]: [Update]
