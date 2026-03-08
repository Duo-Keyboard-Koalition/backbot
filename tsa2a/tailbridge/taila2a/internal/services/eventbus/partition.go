package eventbus

import (
	"container/ring"
	"errors"
	"sync"
)

// Consumer represents a message consumer
type Consumer struct {
	ID      string
	channel chan *Message
	mu      sync.Mutex
	closed  bool
}

// newConsumer creates a new consumer
func newConsumer(id string, bufferSize int) *Consumer {
	return &Consumer{
		ID:      id,
		channel: make(chan *Message, bufferSize),
	}
}

// Send sends a message to the consumer
func (c *Consumer) Send(msg *Message) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return false
	}

	select {
	case c.channel <- msg:
		return true
	default:
		// Channel full, drop message (in production, might want to block or error)
		return false
	}
}

// Close closes the consumer channel
func (c *Consumer) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.closed {
		close(c.channel)
		c.closed = true
	}
}

// Partition represents a single partition within a topic
type Partition struct {
	mu          sync.RWMutex
	id          int
	messages    []*Message
	head        int64 // Next offset to assign
	tail        int64 // Oldest message offset
	subscribers []*Consumer
	closed      bool
}

// newPartition creates a new partition
func newPartition(id int) *Partition {
	return &Partition{
		id:          id,
		messages:    make([]*Message, 0, 1000),
		head:        0,
		tail:        0,
		subscribers: make([]*Consumer, 0),
	}
}

// ID returns the partition ID
func (p *Partition) ID() int {
	return p.id
}

// Append appends a message to the partition
func (p *Partition) Append(msg *Message) (int64, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return -1, ErrShutdown
	}

	// Assign offset
	msg.Offset = p.head
	msg.Partition = p.id
	p.head++

	p.messages = append(p.messages, msg)

	// Notify subscribers
	p.notifySubscribers(msg)

	return msg.Offset, nil
}

// Read reads a message at the given offset
func (p *Partition) Read(offset int64) (*Message, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	idx := offset - p.tail
	if idx < 0 || idx >= int64(len(p.messages)) {
		return nil, ErrOffsetOutOfRange
	}

	return p.messages[idx], nil
}

// ReadRange reads a range of messages
func (p *Partition) ReadRange(start, end int64) ([]*Message, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if start < p.tail {
		start = p.tail
	}
	if end >= p.head {
		end = p.head - 1
	}

	if start > end {
		return []*Message{}, nil
	}

	result := make([]*Message, 0, end-start+1)
	for i := start; i <= end; i++ {
		idx := i - p.tail
		if idx >= 0 && idx < int64(len(p.messages)) {
			result = append(result, p.messages[i-p.tail])
		}
	}

	return result, nil
}

// ReadFrom reads messages starting from an offset up to a maximum count
func (p *Partition) ReadFrom(offset int64, maxCount int) ([]*Message, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if offset < p.tail {
		offset = p.tail
	}
	if offset >= p.head {
		return []*Message{}, nil
	}

	maxEnd := offset + int64(maxCount)
	end := p.head
	if maxEnd < end {
		end = maxEnd
	}

	result := make([]*Message, 0, end-offset)
	for i := offset; i < end; i++ {
		idx := i - p.tail
		if idx >= 0 && idx < int64(len(p.messages)) {
			result = append(result, p.messages[idx])
		}
	}

	return result, nil
}

// subscribe adds a consumer to the partition
func (p *Partition) subscribe(consumer *Consumer) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.subscribers = append(p.subscribers, consumer)
}

// unsubscribe removes a consumer from the partition
func (p *Partition) unsubscribe(consumerID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i, sub := range p.subscribers {
		if sub.ID == consumerID {
			p.subscribers = append(p.subscribers[:i], p.subscribers[i+1:]...)
			return
		}
	}
}

// notifySubscribers sends a message to all subscribers
func (p *Partition) notifySubscribers(msg *Message) {
	for _, sub := range p.subscribers {
		sub.Send(msg)
	}
}

// Head returns the next offset to be assigned
func (p *Partition) Head() int64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.head
}

// Tail returns the oldest message offset
func (p *Partition) Tail() int64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.tail
}

// Count returns the number of messages in the partition
func (p *Partition) Count() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.messages)
}

// Close closes the partition
func (p *Partition) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}

	// Close all subscribers
	for _, sub := range p.subscribers {
		sub.Close()
	}

	p.closed = true
}

// IsClosed returns whether the partition is closed
func (p *Partition) IsClosed() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.closed
}

// Truncate removes messages before the given offset
func (p *Partition) Truncate(offset int64) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if offset <= p.tail {
		return nil
	}

	if offset >= p.head {
		p.messages = make([]*Message, 0)
		p.tail = p.head
		return nil
	}

	// Calculate new tail position
	newTailIdx := offset - p.tail
	if newTailIdx < 0 || newTailIdx >= int64(len(p.messages)) {
		return nil
	}

	// Remove old messages
	p.messages = p.messages[newTailIdx:]
	p.tail = offset

	return nil
}

// PartitionStats holds statistics for a partition
type PartitionStats struct {
	ID              int
	MessageCount    int
	Head            int64
	Tail            int64
	SubscriberCount int
}

// Stats returns partition statistics
func (p *Partition) Stats() PartitionStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return PartitionStats{
		ID:              p.id,
		MessageCount:    len(p.messages),
		Head:            p.head,
		Tail:            p.tail,
		SubscriberCount: len(p.subscribers),
	}
}

// RingBufferPartition is an alternative partition implementation using a ring buffer
// for memory-efficient storage with a fixed capacity
type RingBufferPartition struct {
	mu          sync.RWMutex
	id          int
	buffer      *ring.Ring
	capacity    int
	head        int64
	tail        int64
	subscribers []*Consumer
	closed      bool
}

// newRingBufferPartition creates a new ring buffer partition
func newRingBufferPartition(id int, capacity int) *RingBufferPartition {
	return &RingBufferPartition{
		id:          id,
		buffer:      ring.New(capacity),
		capacity:    capacity,
		head:        0,
		tail:        0,
		subscribers: make([]*Consumer, 0),
	}
}

// Append appends a message to the ring buffer partition
func (p *RingBufferPartition) Append(msg *Message) (int64, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return -1, ErrShutdown
	}

	// Assign offset
	msg.Offset = p.head
	msg.Partition = p.id

	// If buffer is full, advance tail
	if p.head >= p.tail+int64(p.capacity) {
		p.tail = p.head - int64(p.capacity) + 1
	}

	// Store message
	p.buffer.Value = msg
	p.buffer = p.buffer.Next()
	p.head++

	// Notify subscribers
	p.notifySubscribers(msg)

	return msg.Offset, nil
}

// Read reads a message at the given offset
func (p *RingBufferPartition) Read(offset int64) (*Message, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if offset < p.tail || offset >= p.head {
		return nil, ErrOffsetOutOfRange
	}

	// Calculate position in ring
	idx := (offset - p.tail) % int64(p.capacity)
	r := p.buffer
	for i := int64(0); i < idx; i++ {
		r = r.Prev()
	}

	msg, ok := r.Value.(*Message)
	if !ok {
		return nil, ErrMessageNotFound
	}

	return msg, nil
}

// subscribe adds a consumer to the ring buffer partition
func (p *RingBufferPartition) subscribe(consumer *Consumer) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.subscribers = append(p.subscribers, consumer)
}

// notifySubscribers sends a message to all subscribers
func (p *RingBufferPartition) notifySubscribers(msg *Message) {
	for _, sub := range p.subscribers {
		sub.Send(msg)
	}
}

// Close closes the ring buffer partition
func (p *RingBufferPartition) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}

	for _, sub := range p.subscribers {
		sub.Close()
	}

	p.closed = true
}

// Common errors for partition operations
var (
	ErrOffsetOutOfRange = &OffsetOutOfRangeError{}
	ErrMessageNotFound  = errors.New("message not found")
)

// OffsetOutOfRangeError indicates an offset is outside the valid range
type OffsetOutOfRangeError struct{}

func (e *OffsetOutOfRangeError) Error() string {
	return "offset out of range"
}
