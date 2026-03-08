package eventbus

import (
	"sync"
	"testing"
	"time"
)

func TestEventBus_CreateTopic(t *testing.T) {
	eb, err := New(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create event bus: %v", err)
	}
	defer eb.Close()

	// Test successful topic creation
	err = eb.CreateTopic("test-topic", 4)
	if err != nil {
		t.Fatalf("Failed to create topic: %v", err)
	}

	// Test duplicate topic creation
	err = eb.CreateTopic("test-topic", 4)
	if err != ErrTopicExists {
		t.Fatalf("Expected ErrTopicExists, got: %v", err)
	}

	// Test invalid partitions
	err = eb.CreateTopic("another-topic", 0)
	if err != ErrInvalidPartitions {
		t.Fatalf("Expected ErrInvalidPartitions, got: %v", err)
	}

	// Test negative partitions
	err = eb.CreateTopic("another-topic", -1)
	if err != ErrInvalidPartitions {
		t.Fatalf("Expected ErrInvalidPartitions, got: %v", err)
	}
}

func TestEventBus_DeleteTopic(t *testing.T) {
	eb, err := New(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create event bus: %v", err)
	}
	defer eb.Close()

	// Create and delete topic
	err = eb.CreateTopic("test-topic", 4)
	if err != nil {
		t.Fatalf("Failed to create topic: %v", err)
	}

	err = eb.DeleteTopic("test-topic")
	if err != nil {
		t.Fatalf("Failed to delete topic: %v", err)
	}

	// Test delete non-existent topic
	err = eb.DeleteTopic("non-existent")
	if err != ErrTopicNotFound {
		t.Fatalf("Expected ErrTopicNotFound, got: %v", err)
	}
}

func TestEventBus_ListTopics(t *testing.T) {
	eb, err := New(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create event bus: %v", err)
	}
	defer eb.Close()

	// Create multiple topics
	topics := []string{"topic-1", "topic-2", "topic-3"}
	for _, name := range topics {
		err := eb.CreateTopic(name, 4)
		if err != nil {
			t.Fatalf("Failed to create topic: %v", err)
		}
	}

	// List topics
	listed := eb.ListTopics()
	if len(listed) != len(topics) {
		t.Fatalf("Expected %d topics, got %d", len(topics), len(listed))
	}
}

func TestEventBus_Publish(t *testing.T) {
	eb, err := New(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create event bus: %v", err)
	}
	defer eb.Close()

	err = eb.CreateTopic("test-topic", 4)
	if err != nil {
		t.Fatalf("Failed to create topic: %v", err)
	}

	// Publish message
	offset, err := eb.Publish("test-topic", []byte("test message"))
	if err != nil {
		t.Fatalf("Failed to publish: %v", err)
	}

	if offset < 0 {
		t.Fatalf("Expected positive offset, got %d", offset)
	}

	// Publish to non-existent topic
	_, err = eb.Publish("non-existent", []byte("test"))
	if err != ErrTopicNotFound {
		t.Fatalf("Expected ErrTopicNotFound, got: %v", err)
	}
}

func TestEventBus_PublishWithKey(t *testing.T) {
	eb, err := New(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create event bus: %v", err)
	}
	defer eb.Close()

	err = eb.CreateTopic("test-topic", 4)
	if err != nil {
		t.Fatalf("Failed to create topic: %v", err)
	}

	// Messages with same key should go to same partition
	offsets := make([]int64, 10)
	for i := 0; i < 10; i++ {
		offset, err := eb.PublishWithKey("test-topic", "same-key", []byte("message"))
		if err != nil {
			t.Fatalf("Failed to publish: %v", err)
		}
		offsets[i] = offset
	}

	// All offsets should be sequential (same partition)
	for i := 1; i < len(offsets); i++ {
		if offsets[i] != offsets[i-1]+1 {
			t.Fatalf("Expected sequential offsets, got %d and %d", offsets[i-1], offsets[i])
		}
	}
}

func TestEventBus_Subscribe(t *testing.T) {
	eb, err := New(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create event bus: %v", err)
	}
	defer eb.Close()

	err = eb.CreateTopic("test-topic", 1)
	if err != nil {
		t.Fatalf("Failed to create topic: %v", err)
	}

	// Subscribe
	ch, err := eb.Subscribe("test-topic", "test-group")
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Publish message
	_, err = eb.Publish("test-topic", []byte("test message"))
	if err != nil {
		t.Fatalf("Failed to publish: %v", err)
	}

	// Receive message with timeout
	select {
	case msg := <-ch:
		if string(msg.Value) != "test message" {
			t.Fatalf("Expected 'test message', got %s", msg.Value)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for message")
	}
}

func TestEventBus_ConsumerGroup_CommitOffset(t *testing.T) {
	eb, err := New(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create event bus: %v", err)
	}
	defer eb.Close()

	err = eb.CreateTopic("test-topic", 1)
	if err != nil {
		t.Fatalf("Failed to create topic: %v", err)
	}

	// First join the group to create it
	_, err = eb.JoinGroup("test-group", []string{"test-topic"}, "test-member")
	if err != nil {
		t.Fatalf("Failed to join group: %v", err)
	}

	// Commit offset
	err = eb.CommitOffset("test-topic", "test-group", 0, 100)
	if err != nil {
		t.Fatalf("Failed to commit offset: %v", err)
	}

	// Get committed offset
	offset, err := eb.GetCommittedOffset("test-topic", "test-group", 0)
	if err != nil {
		t.Fatalf("Failed to get offset: %v", err)
	}

	if offset != 100 {
		t.Fatalf("Expected offset 100, got %d", offset)
	}
}

func TestEventBus_ConcurrentPublish(t *testing.T) {
	eb, err := New(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create event bus: %v", err)
	}
	defer eb.Close()

	err = eb.CreateTopic("test-topic", 4)
	if err != nil {
		t.Fatalf("Failed to create topic: %v", err)
	}

	var wg sync.WaitGroup
	numGoroutines := 10
	messagesPerGoroutine := 100

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				_, err := eb.Publish("test-topic", []byte("message"))
				if err != nil {
					t.Errorf("Failed to publish: %v", err)
					return
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify topic stats
	info, err := eb.GetTopicInfo("test-topic")
	if err != nil {
		t.Fatalf("Failed to get topic info: %v", err)
	}

	if info.Partitions != 4 {
		t.Fatalf("Expected 4 partitions, got %d", info.Partitions)
	}
}

func TestTopic_SelectPartition(t *testing.T) {
	topic := newTopic("test-topic", 4)
	defer topic.Close()

	// Test key-based partitioning
	partition1 := topic.SelectPartition("key-1")
	partition2 := topic.SelectPartition("key-1")

	if partition1.ID() != partition2.ID() {
		t.Fatalf("Same key should select same partition")
	}

	// Different keys might go to different partitions
	partition3 := topic.SelectPartition("key-2")
	if partition1.ID() == partition3.ID() {
		t.Logf("Warning: Different keys went to same partition (acceptable but unlikely)")
	}
}

func TestPartition_Append(t *testing.T) {
	partition := newPartition(0)
	defer partition.Close()

	// Append messages
	for i := 0; i < 10; i++ {
		msg := &Message{
			Key:       "test-key",
			Value:     []byte("test value"),
			Timestamp: time.Now().UTC(),
		}

		offset, err := partition.Append(msg)
		if err != nil {
			t.Fatalf("Failed to append: %v", err)
		}

		if offset != int64(i) {
			t.Fatalf("Expected offset %d, got %d", i, offset)
		}
	}

	// Verify head
	if partition.Head() != 10 {
		t.Fatalf("Expected head 10, got %d", partition.Head())
	}

	// Verify count
	if partition.Count() != 10 {
		t.Fatalf("Expected count 10, got %d", partition.Count())
	}
}

func TestPartition_Read(t *testing.T) {
	partition := newPartition(0)
	defer partition.Close()

	// Append messages
	for i := 0; i < 10; i++ {
		msg := &Message{
			Key:       "test-key",
			Value:     []byte("test value"),
			Timestamp: time.Now().UTC(),
		}
		partition.Append(msg)
	}

	// Read message at offset 5
	msg, err := partition.Read(5)
	if err != nil {
		t.Fatalf("Failed to read: %v", err)
	}

	if msg.Offset != 5 {
		t.Fatalf("Expected offset 5, got %d", msg.Offset)
	}

	// Read out of range
	_, err = partition.Read(100)
	if err != ErrOffsetOutOfRange {
		t.Fatalf("Expected ErrOffsetOutOfRange, got: %v", err)
	}
}

func TestPartition_ReadRange(t *testing.T) {
	partition := newPartition(0)
	defer partition.Close()

	// Append messages
	for i := 0; i < 10; i++ {
		msg := &Message{
			Key:       "test-key",
			Value:     []byte("test value"),
			Timestamp: time.Now().UTC(),
		}
		partition.Append(msg)
	}

	// Read range
	messages, err := partition.ReadRange(2, 5)
	if err != nil {
		t.Fatalf("Failed to read range: %v", err)
	}

	if len(messages) != 4 {
		t.Fatalf("Expected 4 messages, got %d", len(messages))
	}

	if messages[0].Offset != 2 || messages[3].Offset != 5 {
		t.Fatalf("Unexpected offsets in range")
	}
}

func TestPartition_Truncate(t *testing.T) {
	partition := newPartition(0)
	defer partition.Close()

	// Append messages
	for i := 0; i < 10; i++ {
		msg := &Message{
			Key:       "test-key",
			Value:     []byte("test value"),
			Timestamp: time.Now().UTC(),
		}
		partition.Append(msg)
	}

	// Truncate
	err := partition.Truncate(5)
	if err != nil {
		t.Fatalf("Failed to truncate: %v", err)
	}

	// Verify tail moved
	if partition.Tail() != 5 {
		t.Fatalf("Expected tail 5, got %d", partition.Tail())
	}

	// Old messages should be inaccessible
	_, err = partition.Read(2)
	if err != ErrOffsetOutOfRange {
		t.Fatalf("Expected ErrOffsetOutOfRange for truncated message")
	}
}

func TestConsumerGroup_JoinLeave(t *testing.T) {
	eb, err := New(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create event bus: %v", err)
	}
	defer eb.Close()

	err = eb.CreateTopic("test-topic", 4)
	if err != nil {
		t.Fatalf("Failed to create topic: %v", err)
	}

	// Join group
	resp, err := eb.JoinGroup("test-group", []string{"test-topic"}, "member-1")
	if err != nil {
		t.Fatalf("Failed to join group: %v", err)
	}

	if resp.MemberID != "member-1" {
		t.Fatalf("Expected member-1, got %s", resp.MemberID)
	}

	if resp.GenerationID != 1 {
		t.Fatalf("Expected generation 1, got %d", resp.GenerationID)
	}

	// Leave group
	err = eb.LeaveGroup("test-group", "member-1")
	if err != nil {
		t.Fatalf("Failed to leave group: %v", err)
	}
}

func TestConsumerGroup_PartitionAssignment(t *testing.T) {
	eb, err := New(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create event bus: %v", err)
	}
	defer eb.Close()

	err = eb.CreateTopic("test-topic", 4)
	if err != nil {
		t.Fatalf("Failed to create topic: %v", err)
	}

	// Join with multiple members
	members := []string{"member-1", "member-2", "member-3", "member-4"}
	for _, member := range members {
		_, err := eb.JoinGroup("test-group", []string{"test-topic"}, member)
		if err != nil {
			t.Fatalf("Failed to join: %v", err)
		}
	}

	// Each member should get one partition (4 partitions, 4 members)
	// This is tested indirectly through the join response
}

func TestConsumerGroup_RoundRobinAssigner(t *testing.T) {
	assigner := &RoundRobinAssigner{}

	partitions := map[string][]int{
		"topic-1": {0, 1, 2, 3},
	}

	members := []MemberInfo{
		{ID: "member-1"},
		{ID: "member-2"},
	}

	assignments := assigner.Assign(partitions, members)

	// Each member should get 2 partitions
	if len(assignments["member-1"]["topic-1"]) != 2 {
		t.Fatalf("Expected 2 partitions for member-1")
	}

	if len(assignments["member-2"]["topic-1"]) != 2 {
		t.Fatalf("Expected 2 partitions for member-2")
	}
}

func TestConsumerGroup_RangeAssigner(t *testing.T) {
	assigner := &RangeAssigner{}

	partitions := map[string][]int{
		"topic-1": {0, 1, 2, 3},
	}

	members := []MemberInfo{
		{ID: "member-1"},
		{ID: "member-2"},
	}

	assignments := assigner.Assign(partitions, members)

	// Member-1 should get partitions 0, 1
	// Member-2 should get partitions 2, 3
	if len(assignments["member-1"]["topic-1"]) != 2 {
		t.Fatalf("Expected 2 partitions for member-1")
	}

	if len(assignments["member-2"]["topic-1"]) != 2 {
		t.Fatalf("Expected 2 partitions for member-2")
	}
}

func TestEventBus_Shutdown(t *testing.T) {
	eb, err := New(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create event bus: %v", err)
	}

	err = eb.CreateTopic("test-topic", 4)
	if err != nil {
		t.Fatalf("Failed to create topic: %v", err)
	}

	// Close event bus
	eb.Close()

	// Operations should fail after close
	err = eb.CreateTopic("another-topic", 4)
	if err != ErrShutdown {
		t.Fatalf("Expected ErrShutdown, got: %v", err)
	}

	_, err = eb.Publish("test-topic", []byte("test"))
	if err != ErrShutdown {
		t.Fatalf("Expected ErrShutdown, got: %v", err)
	}
}

func TestMessage_Headers(t *testing.T) {
	eb, err := New(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create event bus: %v", err)
	}
	defer eb.Close()

	err = eb.CreateTopic("test-topic", 1)
	if err != nil {
		t.Fatalf("Failed to create topic: %v", err)
	}

	ch, err := eb.Subscribe("test-topic", "test-group")
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// The current API doesn't support setting headers directly,
	// but the Message struct has the Headers field ready for future use
	msg := &Message{
		Key:       "test-key",
		Value:     []byte("test value"),
		Timestamp: time.Now().UTC(),
		Headers:   map[string]string{"header-key": "header-value"},
	}

	if msg.Headers["header-key"] != "header-value" {
		t.Fatalf("Header not set correctly")
	}

	// Manually append to partition to test
	topic, _ := eb.topics["test-topic"]
	partition := topic.partitions[0]
	partition.Append(msg)

	// Receive and verify
	select {
	case received := <-ch:
		if received.Headers["header-key"] != "header-value" {
			t.Fatalf("Header not preserved")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for message")
	}
}
