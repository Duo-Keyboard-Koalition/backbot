# Task 001: Implement Event Bus Core

## Priority
**P0** - Critical

## Status
⏳ Pending

## Objective
Implement a Kafka-inspired event bus core for Taila2a that enables topic-based message routing between agents.

## Background
The event bus is the central messaging backbone of Taila2a. It provides:
- Topic-based routing (like Kafka topics)
- Message publishing and subscribing
- In-memory message storage with optional persistence
- Consumer group coordination

## Requirements

### Core Functionality
- [ ] Create topic management (create, delete, list topics)
- [ ] Implement message publishing to topics
- [ ] Implement message subscription from topics
- [ ] Support multiple partitions per topic
- [ ] Message ordering within partitions

### Data Structures
- [ ] Topic structure with partitions
- [ ] Message structure with offset tracking
- [ ] Consumer group state management

### API
- [ ] `Publish(topic, message) -> offset`
- [ ] `Subscribe(topic, consumerGroup) -> message channel`
- [ ] `CreateTopic(name, partitions)`
- [ ] `DeleteTopic(name)`
- [ ] `ListTopics() -> []TopicInfo`

## Technical Specification

### Topic Structure
```go
type Topic struct {
    Name       string
    Partitions []*Partition
    CreatedAt  time.Time
}

type Partition struct {
    ID       int
    Messages []*Message
    Head     int64  // Next offset to assign
    Tail     int64  // Oldest message offset
}

type Message struct {
    Offset    int64
    Key       string
    Value     []byte
    Timestamp time.Time
    Headers   map[string]string
}
```

### Event Bus Interface
```go
type EventBus interface {
    // Topic management
    CreateTopic(name string, partitions int) error
    DeleteTopic(name string) error
    ListTopics() []TopicInfo
    
    // Publishing
    Publish(topic string, message []byte) (int64, error)
    PublishWithKey(topic string, key string, message []byte) (int64, error)
    
    // Subscribing
    Subscribe(topic string, groupID string) (<-chan *Message, error)
    Unsubscribe(topic string, groupID string) error
    
    // Consumer groups
    CommitOffset(topic string, groupID string, offset int64) error
    GetCommittedOffset(topic string, groupID string) (int64, error)
}
```

## Acceptance Criteria

### Functional
- [ ] Can create and delete topics
- [ ] Can publish messages to topics
- [ ] Can subscribe to topics and receive messages
- [ ] Messages are ordered within partitions
- [ ] Consumer groups track offsets correctly

### Non-Functional
- [ ] Publish latency < 10ms (p99)
- [ ] Support 1000+ messages/second
- [ ] Thread-safe for concurrent access
- [ ] No memory leaks under sustained load

## Testing Requirements

### Unit Tests
- [ ] Test topic creation/deletion
- [ ] Test message publishing
- [ ] Test message ordering
- [ ] Test concurrent access
- [ ] Test consumer group offset tracking

### Integration Tests
- [ ] Test end-to-end publish/subscribe
- [ ] Test multiple consumers
- [ ] Test partition rebalancing

## Implementation Steps

1. **Phase 1: Core Data Structures** (2 days)
   - Define Topic, Partition, Message structs
   - Implement partition management
   - Add offset tracking

2. **Phase 2: Publish/Subscribe** (3 days)
   - Implement Publish API
   - Implement Subscribe API
   - Add message channels

3. **Phase 3: Consumer Groups** (2 days)
   - Implement consumer group state
   - Add offset commit/retrieve
   - Handle rebalancing

4. **Phase 4: Testing & Polish** (3 days)
   - Write unit tests
   - Write integration tests
   - Performance optimization
   - Documentation

## Files to Create/Modify

### Create
- `taila2a/internal/services/eventbus/eventbus.go` - Core event bus
- `taila2a/internal/services/eventbus/topic.go` - Topic management
- `taila2a/internal/services/eventbus/partition.go` - Partition logic
- `taila2a/internal/services/eventbus/consumer_group.go` - Consumer groups
- `taila2a/internal/services/eventbus/eventbus_test.go` - Tests

### Modify
- `taila2a/cmd/taila2a/main.go` - Integrate event bus
- `taila2a/internal/controllers/controller.go` - Add event bus endpoints

## References
- [A2A Protocol Specification](../eng_nbk/A2A_PROTOCOL.md)
- [Kafka Architecture](https://kafka.apache.org/documentation/)
- [Engineering Notebook](../eng_nbk/README.md)

## Assignment
**Agent:** [Unassigned]  
**Assigned:** [Date]  
**Due:** [Date]  

## Progress Log
- [Date]: [Update]
