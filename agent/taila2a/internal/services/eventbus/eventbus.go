// Package eventbus provides a Kafka-inspired event bus for Taila2a.
// It enables topic-based message routing between agents with consumer group support.
package eventbus

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/codejedi-ai/kaggle-for-tensors/tailscale-app/internal/config"
	"github.com/codejedi-ai/kaggle-for-tensors/tailscale-app/internal/services/eventbus/wal"
)

// Common errors
var (
	ErrTopicExists         = errors.New("topic already exists")
	ErrTopicNotFound       = errors.New("topic not found")
	ErrInvalidPartitions   = errors.New("invalid number of partitions")
	ErrPartitionNotFound   = errors.New("partition not found")
	ErrConsumerGroupExists = errors.New("consumer group already exists")
	ErrNotLeader           = errors.New("not the group leader")
	ErrShutdown            = errors.New("event bus is shutting down")
)

// Message represents a single message in the event bus
type Message struct {
	Offset    int64             `json:"offset"`
	Key       string            `json:"key,omitempty"`
	Value     []byte            `json:"value"`
	Timestamp time.Time         `json:"timestamp"`
	Headers   map[string]string `json:"headers,omitempty"`
	Topic     string            `json:"topic"`
	Partition int               `json:"partition"`
}

// TopicInfo provides information about a topic
type TopicInfo struct {
	Name       string    `json:"name"`
	Partitions int       `json:"partitions"`
	CreatedAt  time.Time `json:"created_at"`
}

// EventBusConfig configures the event bus
type EventBusConfig struct {
	DefaultPartitions int           // Default number of partitions for new topics
	WALEnabled        bool          // Enable write-ahead log
	WALDir            string        // Directory for WAL storage
	SyncInterval      time.Duration // WAL sync interval
}

// DefaultConfig returns a default configuration
func DefaultConfig() EventBusConfig {
	return EventBusConfig{
		DefaultPartitions: config.DefaultDefaultPartitions,
		WALEnabled:        config.DefaultWALEnabled,
		WALDir:            config.DefaultWALDir,
		SyncInterval:      config.DefaultSyncInterval,
	}
}

// NewFromConfig creates a new event bus from config
func NewFromConfig(cfg config.EventBusConfig) (*EventBus, error) {
	ebConfig := EventBusConfig{
		DefaultPartitions: cfg.DefaultPartitions,
		WALEnabled:        cfg.WAL.Enabled,
		WALDir:            cfg.WAL.Dir,
		SyncInterval:      cfg.WAL.SyncInterval,
	}

	if ebConfig.DefaultPartitions == 0 {
		ebConfig.DefaultPartitions = config.DefaultDefaultPartitions
	}
	if ebConfig.WALDir == "" {
		ebConfig.WALDir = config.DefaultWALDir
	}
	if ebConfig.SyncInterval == 0 {
		ebConfig.SyncInterval = config.DefaultSyncInterval
	}

	return New(ebConfig)
}

// EventBus is the central message broker
type EventBus struct {
	mu           sync.RWMutex
	topics       map[string]*Topic
	consumerGroups map[string]*ConsumerGroup
	config       EventBusConfig
	wal          wal.WAL
	shutdown     chan struct{}
	shutdownOnce sync.Once
	closed       bool
}

// New creates a new event bus
func New(config EventBusConfig) (*EventBus, error) {
	eb := &EventBus{
		topics:         make(map[string]*Topic),
		consumerGroups: make(map[string]*ConsumerGroup),
		config:         config,
		shutdown:       make(chan struct{}),
	}

	if config.WALEnabled {
		w, err := wal.New(config.WALDir)
		if err != nil {
			return nil, fmt.Errorf("failed to create WAL: %w", err)
		}
		eb.wal = w

		// Recover from WAL
		if err := eb.recover(); err != nil {
			w.Close()
			return nil, fmt.Errorf("failed to recover from WAL: %w", err)
		}

		// Start background sync
		go eb.syncLoop()
	}

	return eb, nil
}

// recover rebuilds state from WAL
func (eb *EventBus) recover() error {
	if eb.wal == nil {
		return nil
	}
	// WAL recovery will be implemented in task_003
	return nil
}

// syncLoop periodically syncs WAL to disk
func (eb *EventBus) syncLoop() {
	ticker := time.NewTicker(eb.config.SyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if eb.wal != nil {
				eb.wal.Sync()
			}
		case <-eb.shutdown:
			return
		}
	}
}

// CreateTopic creates a new topic with the specified number of partitions
func (eb *EventBus) CreateTopic(name string, partitions int) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.closed {
		return ErrShutdown
	}

	if partitions <= 0 {
		return ErrInvalidPartitions
	}

	if _, exists := eb.topics[name]; exists {
		return ErrTopicExists
	}

	topic := newTopic(name, partitions)
	eb.topics[name] = topic

	return nil
}

// DeleteTopic deletes a topic
func (eb *EventBus) DeleteTopic(name string) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.closed {
		return ErrShutdown
	}

	topic, exists := eb.topics[name]
	if !exists {
		return ErrTopicNotFound
	}

	// Close all partitions
	topic.Close()

	delete(eb.topics, name)
	return nil
}

// ListTopics returns information about all topics
func (eb *EventBus) ListTopics() []TopicInfo {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	infos := make([]TopicInfo, 0, len(eb.topics))
	for name, topic := range eb.topics {
		infos = append(infos, TopicInfo{
			Name:       name,
			Partitions: topic.PartitionCount(),
			CreatedAt:  topic.CreatedAt(),
		})
	}

	return infos
}

// GetTopicInfo returns information about a specific topic
func (eb *EventBus) GetTopicInfo(name string) (TopicInfo, error) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	topic, exists := eb.topics[name]
	if !exists {
		return TopicInfo{}, ErrTopicNotFound
	}

	return TopicInfo{
		Name:       name,
		Partitions: topic.PartitionCount(),
		CreatedAt:  topic.CreatedAt(),
	}, nil
}

// Publish publishes a message to a topic (round-robin across partitions)
func (eb *EventBus) Publish(topic string, message []byte) (int64, error) {
	return eb.PublishWithKey(topic, "", message)
}

// PublishWithKey publishes a message to a topic with a key for partition routing
func (eb *EventBus) PublishWithKey(topic string, key string, message []byte) (int64, error) {
	eb.mu.RLock()
	t, exists := eb.topics[topic]
	eb.mu.RUnlock()

	if !exists {
		return -1, ErrTopicNotFound
	}

	if eb.closed {
		return -1, ErrShutdown
	}

	// Select partition based on key
	partition := t.SelectPartition(key)

	// Create message
	msg := &Message{
		Key:       key,
		Value:     message,
		Timestamp: time.Now().UTC(),
		Headers:   make(map[string]string),
		Topic:     topic,
		Partition: partition.ID(),
	}

	// Append to partition
	offset, err := partition.Append(msg)
	if err != nil {
		return -1, err
	}

	// Write to WAL if enabled
	if eb.wal != nil {
		if _, err := eb.wal.Append(topic, partition.ID(), msg.Value); err != nil {
			// WAL write failure shouldn't block, but should be logged
			// In production, this might trigger a rollback
		}
	}

	return offset, nil
}

// Subscribe creates a subscription to a topic for a consumer group
func (eb *EventBus) Subscribe(topic string, groupID string) (<-chan *Message, error) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.closed {
		return nil, ErrShutdown
	}

	t, exists := eb.topics[topic]
	if !exists {
		return nil, ErrTopicNotFound
	}

	// Get or create consumer group
	group, err := eb.getOrCreateConsumerGroup(groupID, topic)
	if err != nil {
		return nil, err
	}

	// Create consumer and get channel
	consumer := group.AddConsumer()

	// Subscribe to all assigned partitions
	for _, partition := range t.partitions {
		partition.subscribe(consumer)
	}

	return consumer.channel, nil
}

// Unsubscribe removes a consumer from a topic
func (eb *EventBus) Unsubscribe(topic string, groupID string, consumerID string) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	group, exists := eb.consumerGroups[groupID]
	if !exists {
		return nil // Group doesn't exist, nothing to do
	}

	group.RemoveConsumer(consumerID)
	return nil
}

// CommitOffset commits an offset for a consumer group
func (eb *EventBus) CommitOffset(topic string, groupID string, partition int, offset int64) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	group, exists := eb.consumerGroups[groupID]
	if !exists {
		return fmt.Errorf("consumer group %s not found", groupID)
	}

	return group.CommitOffset(topic, partition, offset)
}

// GetCommittedOffset gets the committed offset for a consumer group
func (eb *EventBus) GetCommittedOffset(topic string, groupID string, partition int) (int64, error) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	group, exists := eb.consumerGroups[groupID]
	if !exists {
		return -1, fmt.Errorf("consumer group %s not found", groupID)
	}

	return group.GetCommittedOffset(topic, partition)
}

// JoinGroup adds a consumer to a consumer group
func (eb *EventBus) JoinGroup(groupID string, topics []string, memberID string) (*JoinGroupResponse, error) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.closed {
		return nil, ErrShutdown
	}

	// Get or create consumer group
	var group *ConsumerGroup
	var err error

	group, exists := eb.consumerGroups[groupID]
	if !exists {
		group, err = newConsumerGroup(groupID, topics, eb)
		if err != nil {
			return nil, err
		}
		eb.consumerGroups[groupID] = group
	}

	// Add member
	return group.Join(memberID, topics)
}

// LeaveGroup removes a consumer from a consumer group
func (eb *EventBus) LeaveGroup(groupID string, memberID string) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	group, exists := eb.consumerGroups[groupID]
	if !exists {
		return nil
	}

	return group.Leave(memberID)
}

// Heartbeat sends a heartbeat for a consumer group member
func (eb *EventBus) Heartbeat(groupID string, memberID string, generationID int) error {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	group, exists := eb.consumerGroups[groupID]
	if !exists {
		return fmt.Errorf("consumer group %s not found", groupID)
	}

	return group.Heartbeat(memberID, generationID)
}

// getOrCreateConsumerGroup gets or creates a consumer group
func (eb *EventBus) getOrCreateConsumerGroup(groupID, topic string) (*ConsumerGroup, error) {
	group, exists := eb.consumerGroups[groupID]
	if !exists {
		var err error
		group, err = newConsumerGroup(groupID, []string{topic}, eb)
		if err != nil {
			return nil, err
		}
		eb.consumerGroups[groupID] = group
	}
	return group, nil
}

// Close shuts down the event bus
func (eb *EventBus) Close() error {
	eb.shutdownOnce.Do(func() {
		eb.mu.Lock()
		defer eb.mu.Unlock()

		eb.closed = true
		close(eb.shutdown)

		// Close all topics
		for _, topic := range eb.topics {
			topic.Close()
		}

		// Close WAL
		if eb.wal != nil {
			eb.wal.Close()
		}
	})

	return nil
}

// Context returns a context that is cancelled when the event bus shuts down
func (eb *EventBus) Context() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-eb.shutdown
		cancel()
	}()
	return ctx
}
