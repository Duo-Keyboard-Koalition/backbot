# Task 031: Unit Tests - Taila2a

## Priority
**P0** - Critical

## Status
⏳ Pending

## Objective
Write comprehensive unit tests for Taila2a core components to ensure code quality and prevent regressions.

## Background
Unit tests are essential for maintaining code quality and enabling confident refactoring. This task covers unit testing of all Taila2a core components.

## Requirements

### Coverage Goals
- [ ] Overall coverage > 80%
- [ ] Critical paths > 90%
- [ ] All public APIs tested
- [ ] Edge cases covered

### Components to Test

#### Models (internal/models/)
- [ ] phonebook.go - Agent discovery
- [ ] models.go - Data structures
- [ ] config.go - Configuration
- [ ] agent_trigger.go - Trigger service
- [ ] tui_notifier.go - TUI notifications

#### Controllers (internal/controllers/)
- [ ] controller.go - Main controller
- [ ] phonebook_controller.go - Phone book API
- [ ] trigger_controller.go - Trigger API

#### Services (internal/services/)
- [ ] eventbus.go - Event bus (when implemented)
- [ ] buffer.go - Message buffering

### Test Infrastructure
- [ ] Test helpers and mocks
- [ ] Test data generators
- [ ] Table-driven tests where appropriate
- [ ] CI integration

## Technical Specification

### Test Structure
```
internal/
├── models/
│   ├── phonebook.go
│   └── phonebook_test.go
├── controllers/
│   ├── controller.go
│   └── controller_test.go
└── services/
    ├── eventbus.go
    └── eventbus_test.go
```

### Mock Examples
```go
// Mock PhoneBook for testing
type MockPhoneBook struct {
    Agents []models.AgentInfo
    Err    error
}

func (m *MockPhoneBook) GetAgents() []models.AgentInfo {
    return m.Agents
}

func (m *MockPhoneBook) GetOnlineAgents() []models.AgentInfo {
    var online []models.AgentInfo
    for _, a := range m.Agents {
        if a.Status == models.StatusOnline {
            online = append(online, a)
        }
    }
    return online
}

// Mock HTTP ResponseWriter
type MockResponseWriter struct {
    StatusCode int
    Body       bytes.Buffer
    HeaderMap  http.Header
}
```

### Test Patterns
```go
// Table-driven tests
func TestPhoneBookSearch(t *testing.T) {
    tests := []struct {
        name     string
        agents   []models.AgentInfo
        query    string
        expected int
    }{
        {
            name:     "empty phonebook",
            agents:   []models.AgentInfo{},
            query:    "test",
            expected: 0,
        },
        {
            name: "single match",
            agents: []models.AgentInfo{
                {Name: "test-agent", Status: models.StatusOnline},
            },
            query:    "test",
            expected: 1,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            pb := &models.PhoneBook{Agents: tt.agents}
            results := pb.SearchAgents(tt.query)
            if len(results) != tt.expected {
                t.Errorf("expected %d, got %d", tt.expected, len(results))
            }
        })
    }
}
```

## Acceptance Criteria

### Coverage
- [ ] models/ > 85%
- [ ] controllers/ > 80%
- [ ] services/ > 75%
- [ ] Overall > 80%

### Quality
- [ ] All tests pass consistently
- [ ] No flaky tests
- [ ] Tests run in < 30 seconds
- [ ] Clear test names
- [ ] Good error messages

## Testing Requirements

### Unit Tests
- [ ] Happy path tests
- [ ] Error handling tests
- [ ] Edge case tests
- [ ] Boundary condition tests

### Integration Tests (within unit test scope)
- [ ] Controller + model integration
- [ ] Service + model integration

## Implementation Steps

1. **Phase 1: Test Infrastructure** (1 day)
   - Set up test helpers
   - Create mocks
   - Define test data

2. **Phase 2: Models Tests** (2 days)
   - Test phonebook.go
   - Test models.go
   - Test config.go
   - Test agent_trigger.go

3. **Phase 3: Controllers Tests** (2 days)
   - Test controller.go
   - Test phonebook_controller.go
   - Test trigger_controller.go

4. **Phase 4: Services Tests** (2 days)
   - Test eventbus.go (when ready)
   - Test buffer.go
   - Test trigger service

5. **Phase 5: Coverage Improvement** (1 day)
   - Identify gaps
   - Add missing tests
   - Improve assertions

## Files to Create

### Test Files
- `taila2a/internal/models/phonebook_test.go`
- `taila2a/internal/models/models_test.go`
- `taila2a/internal/models/config_test.go`
- `taila2a/internal/models/agent_trigger_test.go`
- `taila2a/internal/controllers/controller_test.go`
- `taila2a/internal/controllers/phonebook_controller_test.go`
- `taila2a/internal/controllers/trigger_controller_test.go`
- `taila2a/internal/services/buffer_test.go`

### Test Helpers
- `taila2a/internal/testutil/mocks.go`
- `taila2a/internal/testutil/testdata.go`

## References
- [Go Testing Package](https://pkg.go.dev/testing)
- [Table-Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Testify](https://github.com/stretchr/testify) (optional)

## Assignment
**Agent:** [Unassigned]  
**Assigned:** [Date]  
**Due:** [Date]  

## Progress Log
- [Date]: [Update]
