# Agent Task Delegation

**Multi-Agent Development Plan for Tailbridge**

---

## Overview

This document delegates tasks to 4 parallel agents working on the Tailbridge project. Each agent has a clear scope, deliverables, and acceptance criteria.

---

## Agent Assignments

### Agent 1: Event Bus & Messaging

**Focus:** Taila2a event bus core implementation

**Tasks:**
1. [task_001](next_steps/task_001_taila2a_eventbus.md) - Event Bus Core (P0) ✅
2. [task_002](next_steps/task_002_taila2a_consumer_groups.md) - Consumer Groups (P0) ✅
3. [task_003](next_steps/task_003_taila2a_message_persistence.md) - Message Persistence (P1) ✅

**Deliverables:**
- [x] Event bus with topic-based routing
- [x] Consumer group protocol
- [x] WAL persistence layer
- [x] Unit tests for event bus

**Files Created:**
```
taila2a/internal/services/eventbus/
├── eventbus.go           # Core event bus
├── topic.go              # Topic management
├── partition.go          # Partition logic
├── consumer_group.go     # Consumer groups
├── eventbus_test.go      # Tests
└── wal/                  # Write-ahead log
    ├── wal.go
    ├── segment.go
    ├── index.go
    └── wal_test.go
```

**Acceptance Criteria:**
- [x] Publish/subscribe working
- [x] Consumer groups coordinate correctly
- [x] Messages persist across restarts (WAL implemented, recovery optimization pending)
- [x] >80% test coverage (19 tests passing)

**Status:** ✅ COMPLETED - 2026-03-07

**Implementation Summary:**
- Event bus core with topic management, partitioning, and message routing
- Consumer groups with join/leave protocol, heartbeat, and offset tracking
- Round-robin and range partition assignment strategies
- WAL with segment-based storage, indexing, and CRC verification
- All unit tests passing (go test ./internal/services/eventbus/...)

---

### Agent 2: File Transfer Protocol

**Focus:** TailFS chunked transfer implementation

**Tasks:**
1. [task_011](next_steps/task_011_tailfs_chunk_protocol.md) - Chunk Protocol (P0)
2. [task_012](next_steps/task_012_tailfs_resume_support.md) - Resume Support (P0)
3. [task_013](next_steps/task_013_tailfs_compression.md) - Compression (P1)
4. [task_014](next_steps/task_014_tailfs_encryption.md) - Encryption (P1)

**Deliverables:**
- [ ] Chunked file transfer protocol
- [ ] Resume capability
- [ ] Compression support
- [ ] Encryption layer
- [ ] Unit tests for transfers

**Files to Create:**
```
tailfs/internal/services/
├── chunk_protocol.go     # Chunk transfer
├── chunker.go            # File chunking
├── reassembler.go        # File reassembly
├── resume.go             # Resume logic
├── state_store.go        # State persistence
├── compression.go        # Compression
├── encryption.go         # Encryption
└── *_test.go             # Tests
```

**Acceptance Criteria:**
- Files transfer correctly in chunks
- Resume works after interruption
- Compression reduces size
- Encryption secures data
- >80% test coverage

---

### Agent 3: Integration & Discovery

**Focus:** Phone book sync and system integration

**Tasks:**
1. [task_021](next_steps/task_021_integration_phonebook_sync.md) - Phone Book Sync (P0)
2. [task_022](next_steps/task_022_integration_unified_cli.md) - Unified CLI (P1)
3. [task_023](next_steps/task_023_integration_prometheus.md) - Prometheus Metrics (P1)
4. [task_005](next_steps/task_005_taila2a_acl_integration.md) - ACL Integration (P1)

**Deliverables:**
- [ ] Shared phone book package
- [ ] Unified CLI tool
- [ ] Prometheus metrics endpoint
- [ ] Tailscale ACL enforcement

**Files to Create:**
```
tailbridge-common/          # New shared module
├── go.mod
├── phonebook/
│   ├── phonebook.go
│   ├── agent.go
│   ├── capability.go
│   └── events.go
└── config/
    └── config.go

taila2a/cmd/taila2a/tui/   # TUI
├── main.go
├── model.go
├── views.go
└── updates.go
```

**Acceptance Criteria:**
- Both systems share phone book
- CLI works for both taila2a and tailfs
- Metrics exposed at /metrics
- ACLs enforced correctly

---

### Agent 4: Testing & Quality

**Focus:** Comprehensive test coverage and CI/CD

**Tasks:**
1. [task_031](next_steps/task_031_test_unit_taila2a.md) - Unit Tests Taila2a (P0)
2. [task_032](next_steps/task_032_test_unit_tailfs.md) - Unit Tests TailFS (P0)
3. [task_033](next_steps/task_033_test_integration.md) - Integration Tests (P1)
4. [task_034](next_steps/task_034_test_e2e.md) - E2E Tests (P2)

**Deliverables:**
- [ ] Unit tests for Taila2a (>80% coverage)
- [ ] Unit tests for TailFS (>80% coverage)
- [ ] Integration test suite
- [ ] E2E test scenarios
- [ ] CI/CD pipeline validation

**Files to Create:**
```
taila2a/internal/testutil/
├── mocks.go
└── testdata.go

taila2a/**/*_test.go       # All test files

tailfs/internal/testutil/
├── mocks.go
├── testdata.go
└── mock_fs.go

tailfs/**/*_test.go        # All test files

eng_nbk/tests/
├── integration/
│   ├── phonebook_test.go
│   └── transfer_test.go
└── e2e/
    └── scenarios.go
```

**Acceptance Criteria:**
- All unit tests pass
- Integration tests pass
- Coverage reports generated
- CI pipeline green

---

## Sprint Timeline

### Sprint 1 (2 weeks) - Core Functionality

**Week 1:**
- Agent 1: Event bus core (task_001)
- Agent 2: Chunk protocol (task_011)
- Agent 3: Phone book sync (task_021)
- Agent 4: Test infrastructure (setup)

**Week 2:**
- Agent 1: Consumer groups (task_002)
- Agent 2: Resume support (task_012)
- Agent 3: Unified CLI (task_022)
- Agent 4: Unit tests (task_031, task_032)

### Sprint 2 (2 weeks) - Enhanced Features

**Week 3:**
- Agent 1: Message persistence (task_003)
- Agent 2: Compression (task_013)
- Agent 3: Prometheus metrics (task_023)
- Agent 4: Integration tests (task_033)

**Week 4:**
- Agent 1: ACL integration (task_005)
- Agent 2: Encryption (task_014)
- Agent 3: TUI implementation (task_004)
- Agent 4: E2E tests (task_034)

---

## Daily Standup Format

Each agent posts daily updates:

```markdown
## Agent [N] Standup - [Date]

### Completed
- [Task ID]: What was done

### Planned
- [Task ID]: What will be done

### Blockers
- Any blockers or questions
```

---

## Communication

### Sync Points
- **Daily:** Async standup updates
- **Weekly:** Live sync (Monday 10:00 UTC)
- **Sprint Review:** End of each sprint

### Escalation
If blocked for >4 hours:
1. Post in standup with `🚨 BLOCKER` prefix
2. Tag relevant agent for help
3. If unresolved in 2 hours, escalate to human

---

## Quality Standards

### Code Requirements
- All code must pass `go build`
- All code must pass `go vet`
- All code must pass `golangci-lint`
- Test coverage >80%
- Documentation comments required

### Review Process
1. Agent completes task
2. Agent runs all tests locally
3. Agent updates task status
4. Other agents review (peer review)
5. Merge to main

---

## Task Status Tracking

Update task files with progress:

```markdown
## Assignment
**Agent:** Agent 1
**Assigned:** 2026-03-07
**Due:** 2026-03-14
**Status:** In Progress

## Progress Log
- 2026-03-07: Started implementation
- 2026-03-08: Core data structures complete
```

---

## Getting Help

### Documentation
- [Engineering Notebook](README.md)
- [A2A Protocol](A2A_PROTOCOL.md)
- [Task Specifications](next_steps/README.md)

### Code References
- Taila2a: `../taila2a/`
- TailFS: `../tailfs/`

---

*Last updated: March 7, 2026*
*Next Sprint Review: Monday 10:00 UTC*
