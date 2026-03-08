# Task 021: Phone Book Sync Between Taila2a and TailFS

## Priority
**P0** - Critical

## Status
⏳ Pending

## Objective
Implement phone book synchronization between Taila2a and TailFS so both systems share agent discovery information and capabilities.

## Background
Taila2a has a phone book for agent discovery. TailFS needs this information to know which agents can send/receive files. This task creates a shared phone book service that both systems can use.

## Requirements

### Shared Phone Book
- [ ] Extract phone book into shared package
- [ ] Both Taila2a and TailFS import shared package
- [ ] Single phone book instance shared between services
- [ ] Consistent agent view across systems

### Capability Registration
- [ ] TailFS registers file_send/file_receive capabilities
- [ ] Taila2a registers chat/command capabilities
- [ ] Capabilities visible in unified phone book
- [ ] Filter agents by combined capabilities

### Event Notifications
- [ ] Agent join events propagated to both systems
- [ ] Agent leave events propagated to both systems
- [ ] Capability changes notified
- [ ] Status changes synchronized

### API
- [ ] `RegisterCapability(agentID, capability)`
- [ ] `UnregisterCapability(agentID, capability)`
- [ ] `GetAgentsWithCapabilities([]capability) -> []AgentInfo`
- [ ] `SubscribePhoneBook() -> chan PhoneBookEvent`

## Technical Specification

### Shared Package Structure
```
tailbridge-common/
├── phonebook/
│   ├── phonebook.go      # Core phone book
│   ├── agent.go          # Agent types
│   ├── capability.go     # Capability types
│   └── events.go         # Event types
└── config/
    ├── config.go         # Shared config
    └── tailscale.go      # Tailscale utils
```

### Unified Capability Model
```go
type Capability string

const (
    // Taila2a capabilities
    CapabilityChat     Capability = "taila2a:chat"
    CapabilityCommand  Capability = "taila2a:command"
    
    // TailFS capabilities
    CapabilityFileSend    Capability = "tailfs:file_send"
    CapabilityFileReceive Capability = "tailfs:file_receive"
)

type AgentInfo struct {
    // Identity
    ID          string
    Name        string
    Hostname    string
    
    // Network
    IPs         []string
    PrimaryIP   string
    
    // Status
    Status      AgentStatus
    LastSeen    time.Time
    
    // Combined capabilities from both systems
    Capabilities []Capability
    
    // System-specific metadata
    Taila2aMetadata *Taila2aAgentMeta
    TailFSMetadata  *TailFSAgentMeta
}
```

### Event System
```go
type PhoneBookEvent interface {
    Type() EventType
    Agent() AgentInfo
    Timestamp() time.Time
}

type AgentJoinEvent struct {
    agent     AgentInfo
    timestamp time.Time
}

type AgentLeaveEvent struct {
    agent     AgentInfo
    timestamp time.Time
}

type CapabilityChangedEvent struct {
    agent        AgentInfo
    added        []Capability
    removed      []Capability
    timestamp    time.Time
}

// Event subscription
type PhoneBookSubscription interface {
    Events() <-chan PhoneBookEvent
    Unsubscribe()
}
```

## Acceptance Criteria

### Functional
- [ ] Single phone book shared by both systems
- [ ] Capabilities from both systems visible
- [ ] Events propagated correctly
- [ ] Filter by combined capabilities works

### Non-Functional
- [ ] No circular dependencies
- [ ] Shared package is lightweight
- [ ] Event delivery < 100ms
- [ ] No race conditions

## Testing Requirements

### Unit Tests
- [ ] Test shared phone book access
- [ ] Test capability registration
- [ ] Test event propagation

### Integration Tests
- [ ] Test Taila2a + TailFS integration
- [ ] Test concurrent access
- [ ] Test event ordering

## Implementation Steps

1. **Phase 1: Extract Shared Package** (2 days)
   - Create tailbridge-common module
   - Move phone book code
   - Update imports

2. **Phase 2: Capability Unification** (1 day)
   - Define unified capability model
   - Map existing capabilities
   - Add registration API

3. **Phase 3: Event System** (2 days)
   - Implement event pub/sub
   - Add event types
   - Integrate with both systems

4. **Phase 4: Integration** (2 days)
   - Update Taila2a to use shared phone book
   - Update TailFS to use shared phone book
   - Test end-to-end

5. **Phase 5: Testing** (1 day)
   - Integration tests
   - Documentation

## Files to Create/Modify

### Create
- `tailbridge-common/go.mod`
- `tailbridge-common/phonebook/phonebook.go`
- `tailbridge-common/phonebook/agent.go`
- `tailbridge-common/phonebook/capability.go`
- `tailbridge-common/phonebook/events.go`

### Modify
- `taila2a/internal/models/phonebook.go` - Use shared package
- `tail-agent-file-send/internal/models/agent_discovery.go` - Use shared package
- Both main.go files to initialize shared phone book

## References
- [Taila2a Phone Book](../taila2a/internal/models/phonebook.go)
- [TailFS README](../tail-agent-file-send/README.md)

## Assignment
**Agent:** [Unassigned]  
**Assigned:** [Date]  
**Due:** [Date]  

## Progress Log
- [Date]: [Update]
