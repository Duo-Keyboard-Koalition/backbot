package eventbus

import (
	"hash/fnv"
	"sync"
	"time"
)

// Topic represents a message topic with multiple partitions
type Topic struct {
	mu        sync.RWMutex
	name      string
	partitions []*Partition
	createdAt time.Time
	closed    bool
}

// newTopic creates a new topic with the specified number of partitions
func newTopic(name string, partitions int) *Topic {
	t := &Topic{
		name:       name,
		partitions: make([]*Partition, partitions),
		createdAt:  time.Now().UTC(),
	}

	// Create partitions
	for i := 0; i < partitions; i++ {
		t.partitions[i] = newPartition(i)
	}

	return t
}

// Name returns the topic name
func (t *Topic) Name() string {
	return t.name
}

// PartitionCount returns the number of partitions
func (t *Topic) PartitionCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.partitions)
}

// CreatedAt returns the topic creation time
func (t *Topic) CreatedAt() time.Time {
	return t.createdAt
}

// SelectPartition selects a partition for a message based on key
// Uses consistent hashing to ensure messages with the same key go to the same partition
func (t *Topic) SelectPartition(key string) *Partition {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if len(t.partitions) == 1 {
		return t.partitions[0]
	}

	var partitionIdx int
	if key == "" {
		// Round-robin for keyless messages (use time-based selection)
		partitionIdx = int(time.Now().UnixNano() % int64(len(t.partitions)))
	} else {
		// Hash-based partitioning
		h := fnv.New32a()
		h.Write([]byte(key))
		partitionIdx = int(h.Sum32()) % len(t.partitions)
	}

	return t.partitions[partitionIdx]
}

// GetPartition returns a specific partition by ID
func (t *Topic) GetPartition(id int) (*Partition, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if id < 0 || id >= len(t.partitions) {
		return nil, ErrPartitionNotFound
	}

	return t.partitions[id], nil
}

// Partitions returns all partitions
func (t *Topic) Partitions() []*Partition {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make([]*Partition, len(t.partitions))
	copy(result, t.partitions)
	return result
}

// Close closes all partitions
func (t *Topic) Close() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return
	}

	for _, p := range t.partitions {
		p.Close()
	}

	t.closed = true
}

// IsClosed returns whether the topic is closed
func (t *Topic) IsClosed() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.closed
}

// Stats returns statistics about the topic
type TopicStats struct {
	Name           string
	PartitionCount int
	TotalMessages  int64
	PartitionStats []PartitionStats
}

// Stats returns topic statistics
func (t *Topic) Stats() TopicStats {
	t.mu.RLock()
	defer t.mu.RUnlock()

	stats := TopicStats{
		Name:           t.name,
		PartitionCount: len(t.partitions),
		PartitionStats: make([]PartitionStats, len(t.partitions)),
	}

	for i, p := range t.partitions {
		stats.PartitionStats[i] = p.Stats()
		stats.TotalMessages += int64(p.Stats().MessageCount)
	}

	return stats
}
