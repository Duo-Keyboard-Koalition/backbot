package eventbus

import (
	"fmt"
	"hash/fnv"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Consumer group states
const (
	ConsumerGroupStateEmpty       = "Empty"
	ConsumerGroupStatePreparing   = "Preparing"
	ConsumerGroupStateStable      = "Stable"
	ConsumerGroupStateDead        = "Dead"
)

// Default consumer group settings
const (
	DefaultSessionTimeoutMs   = 30000
	DefaultHeartbeatIntervalMs = 3000
	DefaultConsumerBufferSize = 100
)

// JoinGroupResponse is returned when a member joins a group
type JoinGroupResponse struct {
	MemberID     string
	GenerationID int
	LeaderID     string
	Members      []MemberInfo
	Assignments  map[string][]PartitionAssignment
}

// MemberInfo holds information about a group member
type MemberInfo struct {
	ID           string
	Topics       []string
	Capabilities []string
	LastHeartbeat time.Time
}

// PartitionAssignment represents a partition assignment for a consumer
type PartitionAssignment struct {
	Topic     string
	Partition int
}

// ConsumerGroup manages a group of consumers that share message consumption
type ConsumerGroup struct {
	mu                sync.RWMutex
	id                string
	topics            []string
	members           map[string]*ConsumerGroupMember
	generationID      int
	leaderID          string
	state             string
	offsets           map[string]map[int]int64 // topic -> partition -> offset
	eventBus          *EventBus
	sessionTimeout    time.Duration
	heartbeatInterval time.Duration
	rebalanceCh       chan struct{}
	shutdownCh        chan struct{}
	consumers         map[string]*Consumer
}

// ConsumerGroupMember represents a member of a consumer group
type ConsumerGroupMember struct {
	ID            string
	Topics        []string
	Capabilities  []string
	LastHeartbeat time.Time
	GenerationID  int
}

// newConsumerGroup creates a new consumer group
func newConsumerGroup(groupID string, topics []string, eb *EventBus) (*ConsumerGroup, error) {
	if groupID == "" {
		return nil, fmt.Errorf("group ID cannot be empty")
	}

	if len(topics) == 0 {
		return nil, fmt.Errorf("must subscribe to at least one topic")
	}

	cg := &ConsumerGroup{
		id:                groupID,
		topics:            topics,
		members:           make(map[string]*ConsumerGroupMember),
		generationID:      0,
		state:             ConsumerGroupStateEmpty,
		offsets:           make(map[string]map[int]int64),
		eventBus:          eb,
		sessionTimeout:    time.Duration(DefaultSessionTimeoutMs) * time.Millisecond,
		heartbeatInterval: time.Duration(DefaultHeartbeatIntervalMs) * time.Millisecond,
		rebalanceCh:       make(chan struct{}, 1),
		shutdownCh:        make(chan struct{}),
		consumers:         make(map[string]*Consumer),
	}

	// Initialize offset maps for topics
	for _, topic := range topics {
		cg.offsets[topic] = make(map[int]int64)
	}

	// Start background goroutines
	go cg.runHeartbeatChecker()
	go cg.handleRebalances()

	return cg, nil
}

// Join adds a member to the consumer group
func (cg *ConsumerGroup) Join(memberID string, topics []string) (*JoinGroupResponse, error) {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	// Generate member ID if not provided
	if memberID == "" {
		memberID = uuid.New().String()
	}

	// Check if member already exists
	if _, exists := cg.members[memberID]; exists {
		// Member rejoining, update topics
		cg.members[memberID].Topics = topics
		cg.members[memberID].LastHeartbeat = time.Now().UTC()
	} else {
		// New member
		cg.members[memberID] = &ConsumerGroupMember{
			ID:            memberID,
			Topics:        topics,
			Capabilities:  []string{},
			LastHeartbeat: time.Now().UTC(),
			GenerationID:  cg.generationID,
		}
	}

	// Create consumer for this member
	consumer := newConsumer(memberID, DefaultConsumerBufferSize)
	cg.consumers[memberID] = consumer

	// First member becomes leader
	if cg.leaderID == "" {
		cg.leaderID = memberID
	}

	// Transition to Preparing state
	if cg.state == ConsumerGroupStateEmpty {
		cg.state = ConsumerGroupStatePreparing
	}

	// Trigger rebalance
	cg.triggerRebalance()

	// Perform rebalance synchronously for join
	assignments := cg.performRebalance()

	return &JoinGroupResponse{
		MemberID:     memberID,
		GenerationID: cg.generationID,
		LeaderID:     cg.leaderID,
		Members:      cg.getMemberInfos(),
		Assignments:  assignments,
	}, nil
}

// Leave removes a member from the consumer group
func (cg *ConsumerGroup) Leave(memberID string) error {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	if _, exists := cg.members[memberID]; !exists {
		return nil // Member doesn't exist, nothing to do
	}

	// Close consumer
	if consumer, ok := cg.consumers[memberID]; ok {
		consumer.Close()
		delete(cg.consumers, memberID)
	}

	delete(cg.members, memberID)

	// If leader left, elect new leader
	if cg.leaderID == memberID {
		cg.electNewLeader()
	}

	// Trigger rebalance
	if len(cg.members) == 0 {
		cg.state = ConsumerGroupStateEmpty
	} else {
		cg.triggerRebalance()
	}

	return nil
}

// Heartbeat updates a member's last heartbeat time
func (cg *ConsumerGroup) Heartbeat(memberID string, generationID int) error {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	member, exists := cg.members[memberID]
	if !exists {
		return fmt.Errorf("member %s not found", memberID)
	}

	if member.GenerationID != generationID {
		return fmt.Errorf("stale generation ID: expected %d, got %d", cg.generationID, generationID)
	}

	member.LastHeartbeat = time.Now().UTC()
	return nil
}

// CommitOffset commits an offset for a partition
func (cg *ConsumerGroup) CommitOffset(topic string, partition int, offset int64) error {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	if _, exists := cg.offsets[topic]; !exists {
		cg.offsets[topic] = make(map[int]int64)
	}

	cg.offsets[topic][partition] = offset
	return nil
}

// GetCommittedOffset gets the committed offset for a partition
func (cg *ConsumerGroup) GetCommittedOffset(topic string, partition int) (int64, error) {
	cg.mu.RLock()
	defer cg.mu.RUnlock()

	if topicOffsets, exists := cg.offsets[topic]; exists {
		if offset, ok := topicOffsets[partition]; ok {
			return offset, nil
		}
	}

	return -1, nil // No committed offset, start from beginning
}

// AddConsumer adds a consumer to the group (for simple subscribe without join protocol)
func (cg *ConsumerGroup) AddConsumer() *Consumer {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	id := uuid.New().String()
	consumer := newConsumer(id, DefaultConsumerBufferSize)
	cg.consumers[id] = consumer
	return consumer
}

// RemoveConsumer removes a consumer from the group
func (cg *ConsumerGroup) RemoveConsumer(consumerID string) {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	if consumer, ok := cg.consumers[consumerID]; ok {
		consumer.Close()
		delete(cg.consumers, consumerID)
	}
}

// electNewLeader elects a new leader from remaining members
func (cg *ConsumerGroup) electNewLeader() {
	for memberID := range cg.members {
		cg.leaderID = memberID
		return
	}
	cg.leaderID = ""
}

// triggerRebalance triggers a rebalance
func (cg *ConsumerGroup) triggerRebalance() {
	select {
	case cg.rebalanceCh <- struct{}{}:
	default:
		// Rebalance already pending
	}
}

// handleRebalances processes rebalance requests
func (cg *ConsumerGroup) handleRebalances() {
	for {
		select {
		case <-cg.rebalanceCh:
			cg.mu.Lock()
			cg.performRebalance()
			cg.mu.Unlock()
		case <-cg.shutdownCh:
			return
		}
	}
}

// performRebalance performs partition assignment
func (cg *ConsumerGroup) performRebalance() map[string][]PartitionAssignment {
	// Increment generation ID
	cg.generationID++

	// Get all partitions for subscribed topics
	partitions := cg.getPartitions()

	// Get members sorted by ID for deterministic assignment
	members := cg.getSortedMembers()

	if len(members) == 0 || len(partitions) == 0 {
		return make(map[string][]PartitionAssignment)
	}

	// Use round-robin assignment
	assignments := cg.roundRobinAssign(partitions, members)

	cg.state = ConsumerGroupStateStable
	return assignments
}

// getPartitions returns all partitions for subscribed topics
func (cg *ConsumerGroup) getPartitions() []PartitionAssignment {
	var partitions []PartitionAssignment

	for _, topicName := range cg.topics {
		if topic, exists := cg.eventBus.topics[topicName]; exists {
			for _, partition := range topic.Partitions() {
				partitions = append(partitions, PartitionAssignment{
					Topic:     topicName,
					Partition: partition.ID(),
				})
			}
		}
	}

	return partitions
}

// getSortedMembers returns members sorted by ID for deterministic assignment
func (cg *ConsumerGroup) getSortedMembers() []*ConsumerGroupMember {
	members := make([]*ConsumerGroupMember, 0, len(cg.members))
	for _, m := range cg.members {
		members = append(members, m)
	}

	// Simple bubble sort for determinism
	for i := 0; i < len(members)-1; i++ {
		for j := 0; j < len(members)-i-1; j++ {
			if members[j].ID > members[j+1].ID {
				members[j], members[j+1] = members[j+1], members[j]
			}
		}
	}

	return members
}

// roundRobinAssign assigns partitions using round-robin
func (cg *ConsumerGroup) roundRobinAssign(
	partitions []PartitionAssignment,
	members []*ConsumerGroupMember,
) map[string][]PartitionAssignment {
	assignments := make(map[string][]PartitionAssignment)

	for i, partition := range partitions {
		memberIdx := i % len(members)
		memberID := members[memberIdx].ID
		assignments[memberID] = append(assignments[memberID], partition)
	}

	return assignments
}

// getMemberInfos returns information about all members
func (cg *ConsumerGroup) getMemberInfos() []MemberInfo {
	infos := make([]MemberInfo, 0, len(cg.members))
	for _, m := range cg.members {
		infos = append(infos, MemberInfo{
			ID:            m.ID,
			Topics:        m.Topics,
			Capabilities:  m.Capabilities,
			LastHeartbeat: m.LastHeartbeat,
		})
	}
	return infos
}

// runHeartbeatChecker checks for dead members
func (cg *ConsumerGroup) runHeartbeatChecker() {
	ticker := time.NewTicker(cg.heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cg.checkHeartbeats()
		case <-cg.shutdownCh:
			return
		}
	}
}

// checkHeartbeats checks for members that have timed out
func (cg *ConsumerGroup) checkHeartbeats() {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	now := time.Now().UTC()
	deadMembers := []string{}

	for memberID, member := range cg.members {
		if now.Sub(member.LastHeartbeat) > cg.sessionTimeout {
			deadMembers = append(deadMembers, memberID)
		}
	}

	for _, memberID := range deadMembers {
		// Close consumer
		if consumer, ok := cg.consumers[memberID]; ok {
			consumer.Close()
			delete(cg.consumers, memberID)
		}
		delete(cg.members, memberID)
	}

	if len(deadMembers) > 0 {
		if len(cg.members) == 0 {
			cg.state = ConsumerGroupStateEmpty
			cg.leaderID = ""
		} else {
			cg.electNewLeader()
			cg.triggerRebalance()
		}
	}
}

// State returns the current state of the consumer group
func (cg *ConsumerGroup) State() string {
	cg.mu.RLock()
	defer cg.mu.RUnlock()
	return cg.state
}

// GenerationID returns the current generation ID
func (cg *ConsumerGroup) GenerationID() int {
	cg.mu.RLock()
	defer cg.mu.RUnlock()
	return cg.generationID
}

// MemberCount returns the number of members in the group
func (cg *ConsumerGroup) MemberCount() int {
	cg.mu.RLock()
	defer cg.mu.RUnlock()
	return len(cg.members)
}

// Close shuts down the consumer group
func (cg *ConsumerGroup) Close() {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	close(cg.shutdownCh)

	// Close all consumers
	for _, consumer := range cg.consumers {
		consumer.Close()
	}
}

// PartitionAssigner defines the interface for partition assignment strategies
type PartitionAssigner interface {
	Assign(
		partitions map[string][]int,
		members []MemberInfo,
	) map[string]map[string][]int
}

// RoundRobinAssigner assigns partitions using round-robin
type RoundRobinAssigner struct{}

// Assign performs round-robin partition assignment
func (r *RoundRobinAssigner) Assign(
	partitions map[string][]int,
	members []MemberInfo,
) map[string]map[string][]int {
	assignments := make(map[string]map[string][]int)

	if len(members) == 0 {
		return assignments
	}

	// Flatten partitions into a single list
	type topicPartition struct {
		topic     string
		partition int
	}

	var allPartitions []topicPartition
	for topic, parts := range partitions {
		for _, p := range parts {
			allPartitions = append(allPartitions, topicPartition{topic, p})
		}
	}

	// Sort members by ID for determinism
	sortedMembers := make([]MemberInfo, len(members))
	copy(sortedMembers, members)
	for i := 0; i < len(sortedMembers)-1; i++ {
		for j := 0; j < len(sortedMembers)-i-1; j++ {
			if sortedMembers[j].ID > sortedMembers[j+1].ID {
				sortedMembers[j], sortedMembers[j+1] = sortedMembers[j+1], sortedMembers[j]
			}
		}
	}

	// Round-robin assignment
	for i, tp := range allPartitions {
		memberIdx := i % len(sortedMembers)
		memberID := sortedMembers[memberIdx].ID

		if assignments[memberID] == nil {
			assignments[memberID] = make(map[string][]int)
		}

		assignments[memberID][tp.topic] = append(assignments[memberID][tp.topic], tp.partition)
	}

	return assignments
}

// RangeAssigner assigns partitions using range assignment
type RangeAssigner struct{}

// Assign performs range partition assignment
func (r *RangeAssigner) Assign(
	partitions map[string][]int,
	members []MemberInfo,
) map[string]map[string][]int {
	assignments := make(map[string]map[string][]int)

	if len(members) == 0 {
		return assignments
	}

	// Sort members by ID for determinism
	sortedMembers := make([]MemberInfo, len(members))
	copy(sortedMembers, members)
	for i := 0; i < len(sortedMembers)-1; i++ {
		for j := 0; j < len(sortedMembers)-i-1; j++ {
			if sortedMembers[j].ID > sortedMembers[j+1].ID {
				sortedMembers[j], sortedMembers[j+1] = sortedMembers[j+1], sortedMembers[j]
			}
		}
	}

	// Assign partitions per topic
	for topic, parts := range partitions {
		if len(parts) == 0 {
			continue
		}

		// Sort partitions
		sortedParts := make([]int, len(parts))
		copy(sortedParts, parts)
		for i := 0; i < len(sortedParts)-1; i++ {
			for j := 0; j < len(sortedParts)-i-1; j++ {
				if sortedParts[j] > sortedParts[j+1] {
					sortedParts[j], sortedParts[j+1] = sortedParts[j+1], sortedParts[j]
				}
			}
		}

		// Calculate range per member
		numParts := len(sortedParts)
		numMembers := len(sortedMembers)
		partitionsPerMember := numParts / numMembers
		extraPartitions := numParts % numMembers

		partitionIdx := 0
		for i, member := range sortedMembers {
			if assignments[member.ID] == nil {
				assignments[member.ID] = make(map[string][]int)
			}

			// Calculate range for this member
			memberPartitions := partitionsPerMember
			if i < extraPartitions {
				memberPartitions++
			}

			for j := 0; j < memberPartitions && partitionIdx < numParts; j++ {
				assignments[member.ID][topic] = append(assignments[member.ID][topic], sortedParts[partitionIdx])
				partitionIdx++
			}
		}
	}

	return assignments
}

// hashPartitioner computes partition based on key hash
func hashPartitioner(key string, numPartitions int) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32()) % numPartitions
}
