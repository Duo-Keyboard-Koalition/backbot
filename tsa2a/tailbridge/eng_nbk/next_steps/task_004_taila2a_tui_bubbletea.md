# Task 004: TUI with Bubbletea

## Priority
**P1** - High

## Status
⏳ Pending

## Objective
Implement a terminal user interface (TUI) for Taila2a using the Bubbletea framework, providing real-time monitoring and control of agents, topics, and transfers.

## Background
A TUI provides operators with real-time visibility into the system without needing a web browser. Bubbletea is a Go framework for building terminal applications.

## Requirements

### Dashboard Views
- [ ] Main dashboard with system overview
- [ ] Agent phone book view
- [ ] Topic monitoring view
- [ ] Consumer group view
- [ ] Transfer progress view (TailFS integration)
- [ ] Log viewer

### Interactive Features
- [ ] Keyboard navigation between views
- [ ] Search/filter agents
- [ ] Sort columns
- [ ] Real-time updates
- [ ] Trigger actions (start/stop agents)

### Real-time Data
- [ ] Agent status updates
- [ ] Message throughput
- [ ] Consumer lag
- [ ] Transfer progress
- [ ] System metrics (CPU, memory)

## Technical Specification

### View Layout
```
┌─────────────────────────────────────────────────────────────┐
│  Taila2a Dashboard                              [q] Quit    │
├─────────────────────────────────────────────────────────────┤
│  Agents: 12 Online  │  Topics: 5  │  Messages: 1,234/s     │
├─────────────────────────────────────────────────────────────┤
│  AGENTS (12 online)                                         │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ Name            Status    Messages    Lag     CPU    │  │
│  │ taila2a-alpha   ● Online  1,234/s     0       12%    │  │
│  │ taila2a-beta    ● Online  987/s       12      8%     │  │
│  │ taila2a-gamma   ○ Offline 0           -       -      │  │
│  └──────────────────────────────────────────────────────┘  │
├─────────────────────────────────────────────────────────────┤
│  TOPICS                                                     │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ Name              Partitions  Consumers   Backlog    │  │
│  │ agent.requests    3           6           145        │  │
│  │ agent.responses   3           4           23         │  │
│  └──────────────────────────────────────────────────────┘  │
├─────────────────────────────────────────────────────────────┤
│  [1]Agents  [2]Topics  [3]Consumers  [4]Transfers  [5]Logs │
└─────────────────────────────────────────────────────────────┘
```

### Bubbletea Model
```go
type model struct {
    // State
    agents      []AgentInfo
    topics      []TopicInfo
    consumers   []ConsumerGroupInfo
    transfers   []TransferInfo
    
    // UI State
    activeTab   int
    searchQuery string
    sortBy      string
    sortOrder   sort.Order
    
    // Real-time
    ticker      *time.Ticker
    lastUpdate  time.Time
    
    // Configuration
    refreshRate time.Duration
}

type tickMsg struct{}
type agentsLoadedMsg struct{ agents []AgentInfo }
type topicsLoadedMsg struct{ topics []TopicInfo }
```

### Keyboard Shortcuts
```
Global:
  q           Quit
  1-5         Switch tabs
  /           Search
  r           Refresh
  ?           Help

Agent View:
  t           Trigger agent
  s           Stop agent
  f           Filter

Topic View:
  p           Purge topic
  c           View consumers
```

## Acceptance Criteria

### Functional
- [ ] All dashboard views implemented
- [ ] Real-time data updates
- [ ] Keyboard navigation works
- [ ] Search/filter functional
- [ ] Actions (trigger/stop) work

### Non-Functional
- [ ] UI updates smoothly (60fps)
- [ ] Memory usage < 50MB
- [ ] CPU usage < 5%
- [ ] Works on 80x24 terminal

## Testing Requirements

### Manual Testing
- [ ] Test all views
- [ ] Test keyboard shortcuts
- [ ] Test search/filter
- [ ] Test with 100+ agents
- [ ] Test long-running session

### Automated Tests
- [ ] Test model updates
- [ ] Test message handling
- [ ] Test view rendering

## Implementation Steps

1. **Phase 1: Setup** (1 day)
   - Add bubbletea dependency
   - Create basic model
   - Implement main view skeleton

2. **Phase 2: Agent View** (2 days)
   - Implement agent table
   - Add search/filter
   - Add sorting
   - Connect to phone book API

3. **Phase 3: Topic View** (2 days)
   - Implement topic table
   - Show message rates
   - Show consumer lag

4. **Phase 4: Consumer View** (1 day)
   - Show consumer groups
   - Show partition assignments
   - Show offsets

5. **Phase 5: Transfer View** (1 day)
   - Integrate with TailFS
   - Show transfer progress
   - Show history

6. **Phase 6: Polish** (2 days)
   - Add help screen
   - Add error handling
   - Performance optimization
   - Documentation

## Files to Create/Modify

### Create
- `taila2a/cmd/taila2a/tui/main.go` - TUI entry point
- `taila2a/cmd/taila2a/tui/model.go` - TUI state model
- `taila2a/cmd/taila2a/tui/views.go` - View implementations
- `taila2a/cmd/taila2a/tui/updates.go` - Update handlers
- `taila2a/cmd/taila2a/tui/api.go` - API client for TUI
- `taila2a/cmd/taila2a/tui/styles.go` - Styling

### Modify
- `taila2a/cmd/taila2a/main.go` - Add TUI command
- `taila2a/go.mod` - Add bubbletea dependency

### Dependencies
```go
require (
    github.com/charmbracelet/bubbletea v0.25.0
    github.com/charmbracelet/bubbles v0.18.0
    github.com/charmbracelet/lipgloss v0.9.1
)
```

## References
- [Bubbletea Documentation](https://github.com/charmbracelet/bubbletea)
- [Bubbles Components](https://github.com/charmbracelet/bubbles)
- [Taila2a Phone Book](../taila2a/internal/models/phonebook.go)

## Assignment
**Agent:** [Unassigned]  
**Assigned:** [Date]  
**Due:** [Date]  

## Progress Log
- [Date]: [Update]
