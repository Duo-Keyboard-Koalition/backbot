package mock

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// AgentState represents the current state of an agent
type AgentState string

const (
	AgentStateInitializing AgentState = "initializing"
	AgentStateStarting     AgentState = "starting"
	AgentStateRunning      AgentState = "running"
	AgentStateStopping     AgentState = "stopping"
	AgentStateStopped      AgentState = "stopped"
	AgentStateError        AgentState = "error"
)

// DiscoveryEvent represents a network discovery event
type DiscoveryEvent struct {
	Type      string    `json:"type"` // "agent_joined", "agent_left", "agent_updated"
	AgentName string    `json:"agent_name"`
	Timestamp time.Time `json:"timestamp"`
	Details   string    `json:"details,omitempty"`
}

// NetworkStats provides statistics about the mock network
type NetworkStats struct {
	TotalAgents       int `json:"total_agents"`
	RunningAgents     int `json:"running_agents"`
	StoppedAgents     int `json:"stopped_agents"`
	TotalMessages     int `json:"total_messages"`
	TotalTopics       int `json:"total_topics"`
	ConsumerGroups    int `json:"consumer_groups"`
	DiscoveryEvents   int `json:"discovery_events"`
}

// IPAllocator allocates Tailscale-like IPs for mock agents
type IPAllocator struct {
	mu          sync.Mutex
	nextOctet   byte
	baseIP      string
	baseIPv6    string
}

// NewIPAllocator creates a new IP allocator
func NewIPAllocator() *IPAllocator {
	return &IPAllocator{
		nextOctet:   1,
		baseIP:      "100.64.0",
		baseIPv6:    "fd7a:115c:a1e0::",
	}
}

// NextIP returns the next available IP pair (IPv4 + IPv6)
func (ipa *IPAllocator) NextIP() (string, string) {
	ipa.mu.Lock()
	defer ipa.mu.Unlock()
	
	ip := fmt.Sprintf("%s.%d", ipa.baseIP, ipa.nextOctet)
	ipv6 := fmt.Sprintf("%s%d", ipa.baseIPv6, ipa.nextOctet)
	ipa.nextOctet++
	
	return ip, ipv6
}

// Message represents an A2A message envelope
type Message struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Source    string          `json:"source"`
	Dest      string          `json:"dest"`
	Topic     string          `json:"topic"`
	Timestamp time.Time       `json:"timestamp"`
	Body      MessageBody     `json:"body"`
	CorrelationID string      `json:"correlation_id,omitempty"`
	ReplyTo   string          `json:"reply_to,omitempty"`
	TTL       time.Duration   `json:"ttl_ms,omitempty"`
}

// MessageBody contains the message payload
type MessageBody struct {
	Action      string          `json:"action"`
	ContentType string          `json:"content_type"`
	Payload     json.RawMessage `json:"payload"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// FileTransferRequest represents a file transfer request
type FileTransferRequest struct {
	ID            string `json:"id"`
	FilePath      string `json:"file_path"`
	DestAgentName string `json:"dest_agent_name"`
	FileName      string `json:"file_name"`
	FileSize      int64  `json:"file_size"`
	Compress      bool   `json:"compress"`
	Encrypt       bool   `json:"encrypt"`
}

// FileTransferProgress tracks transfer progress
type FileTransferProgress struct {
	TransferID    string  `json:"transfer_id"`
	Status        string  `json:"status"`
	BytesSent     int64   `json:"bytes_sent"`
	BytesTotal    int64   `json:"bytes_total"`
	PercentComplete float64 `json:"percent_complete"`
	BytesPerSecond int64  `json:"bytes_per_second"`
	ETASeconds    int64   `json:"eta_seconds"`
}

// MockAgent represents a simulated agent in the test network
type MockAgent struct {
	Name            string
	NodeID          string
	Tailnet         string
	Capabilities    []string
	Messages        []Message
	Files           map[string][]byte
	Inbox           chan Message
	State           AgentState
	AutoAck         bool
	ProcessFunc     func(Message) Message
	mu              sync.RWMutex
	publicKey       ed25519.PublicKey
	privateKey      ed25519.PrivateKey
	StartTime       time.Time
	LastActivity    time.Time
	LastSeen        time.Time
	MessageCount    int
	Network         *MockNetwork
	TailscaleIP     string
	TailscaleIPv6   string
	InboundPort     int
	HTTPPort        int
}

// MockNetwork simulates a Tailscale tailnet for testing
type MockNetwork struct {
	agents         map[string]*MockAgent
	topics         map[string][]Message
	consumerGroups map[string]*ConsumerGroup
	discoveryEvents []DiscoveryEvent
	mu             sync.RWMutex
	networkLatency time.Duration
	packetLoss     float64
	startTime      time.Time
	messageCount   int
	ipAllocator    *IPAllocator
	networkID      string
}

// ConsumerGroup simulates Kafka-style consumer groups
type ConsumerGroup struct {
	ID         string
	Topics     []string
	Members    []*MockAgent
	Offsets    map[string]int
	Generation int
	mu         sync.RWMutex
}

// NewMockNetwork creates a new mock Tailscale network
func NewMockNetwork() *MockNetwork {
	return &MockNetwork{
		agents:         make(map[string]*MockAgent),
		topics:         make(map[string][]Message),
		consumerGroups: make(map[string]*ConsumerGroup),
		discoveryEvents: make([]DiscoveryEvent, 0),
		networkLatency: 10 * time.Millisecond,
		packetLoss:     0.0,
		startTime:      time.Now(),
		ipAllocator:    NewIPAllocator(),
		networkID:      uuid.New().String()[:8],
	}
}

// NewMockAgent creates a new mock agent (doesn't add to network)
func NewMockAgent(name, tailnet string, capabilities []string) *MockAgent {
	publicKey, privateKey, _ := ed25519.GenerateKey(nil)
	
	return &MockAgent{
		Name:         name,
		NodeID:       uuid.New().String()[:16],
		Tailnet:      tailnet,
		Capabilities: capabilities,
		Messages:     make([]Message, 0),
		Files:        make(map[string][]byte),
		Inbox:        make(chan Message, 100),
		State:        AgentStateInitializing,
		AutoAck:      true,
		publicKey:    publicKey,
		privateKey:   privateKey,
		StartTime:    time.Now(),
		LastSeen:     time.Now(),
		InboundPort:  8001,
		HTTPPort:     8080,
	}
}

// AddAgent adds an agent to the mock network and starts it
func (mn *MockNetwork) AddAgent(agent *MockAgent) error {
	mn.mu.Lock()
	defer mn.mu.Unlock()
	
	if _, exists := mn.agents[agent.Name]; exists {
		return fmt.Errorf("agent %s already exists", agent.Name)
	}
	
	agent.Network = mn
	agent.State = AgentStateStarting
	mn.agents[agent.Name] = agent
	
	// Assign Tailscale IP addresses
	agent.TailscaleIP, agent.TailscaleIPv6 = mn.ipAllocator.NextIP()
	
	// Record discovery event
	mn.discoveryEvents = append(mn.discoveryEvents, DiscoveryEvent{
		Type:      "agent_joined",
		AgentName: agent.Name,
		Timestamp: time.Now(),
		Details:   fmt.Sprintf("Capabilities: %v, IP: %s", agent.Capabilities, agent.TailscaleIP),
	})
	
	// Start the agent
	agent.State = AgentStateRunning
	agent.StartTime = time.Now()
	agent.LastSeen = time.Now()
	
	return nil
}

// RemoveAgent removes an agent from the network
func (mn *MockNetwork) RemoveAgent(name string) error {
	mn.mu.Lock()
	defer mn.mu.Unlock()
	
	agent, exists := mn.agents[name]
	if !exists {
		return fmt.Errorf("agent %s not found", name)
	}
	
	agent.State = AgentStateStopping
	
	// Only close if not already closed
	select {
	case <-agent.Inbox:
		// Channel already closed
	default:
		close(agent.Inbox)
	}
	
	delete(mn.agents, name)
	
	// Record discovery event
	mn.discoveryEvents = append(mn.discoveryEvents, DiscoveryEvent{
		Type:      "agent_left",
		AgentName: name,
		Timestamp: time.Now(),
		Details:   fmt.Sprintf("Uptime: %v", time.Since(agent.StartTime)),
	})
	
	return nil
}

// GetAgent retrieves an agent by name
func (mn *MockNetwork) GetAgent(name string) (*MockAgent, bool) {
	mn.mu.RLock()
	defer mn.mu.RUnlock()
	agent, exists := mn.agents[name]
	return agent, exists
}

// GetAllAgents returns all agents in the network
func (mn *MockNetwork) GetAllAgents() []*MockAgent {
	mn.mu.RLock()
	defer mn.mu.RUnlock()
	
	agents := make([]*MockAgent, 0, len(mn.agents))
	for _, agent := range mn.agents {
		agents = append(agents, agent)
	}
	return agents
}

// GetRunningAgents returns only running agents
func (mn *MockNetwork) GetRunningAgents() []*MockAgent {
	mn.mu.RLock()
	defer mn.mu.RUnlock()
	
	agents := make([]*MockAgent, 0)
	for _, agent := range mn.agents {
		if agent.State == AgentStateRunning {
			agents = append(agents, agent)
		}
	}
	return agents
}

// GetAgentCount returns the total number of agents
func (mn *MockNetwork) GetAgentCount() int {
	mn.mu.RLock()
	defer mn.mu.RUnlock()
	return len(mn.agents)
}

// GetRunningAgentCount returns the number of running agents
func (mn *MockNetwork) GetRunningAgentCount() int {
	mn.mu.RLock()
	defer mn.mu.RUnlock()
	count := 0
	for _, agent := range mn.agents {
		if agent.State == AgentStateRunning {
			count++
		}
	}
	return count
}

// SearchAgents searches for agents by name pattern
func (mn *MockNetwork) SearchAgents(pattern string) []*MockAgent {
	mn.mu.RLock()
	defer mn.mu.RUnlock()
	
	var results []*MockAgent
	for _, agent := range mn.agents {
		if contains(agent.Name, pattern) {
			results = append(results, agent)
		}
	}
	return results
}

// GetAgentsByCapability returns agents with a specific capability
func (mn *MockNetwork) GetAgentsByCapability(capability string) []*MockAgent {
	mn.mu.RLock()
	defer mn.mu.RUnlock()
	
	var results []*MockAgent
	for _, agent := range mn.agents {
		for _, cap := range agent.Capabilities {
			if cap == capability {
				results = append(results, agent)
				break
			}
		}
	}
	return results
}

// WaitForAgent waits for an agent to appear in the network
func (mn *MockNetwork) WaitForAgent(name string, timeout time.Duration) (*MockAgent, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if agent, exists := mn.GetAgent(name); exists {
			return agent, nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return nil, fmt.Errorf("timeout waiting for agent %s", name)
}

// WaitForAllAgents waits for all expected agents to be running
func (mn *MockNetwork) WaitForAllAgents(expectedCount int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if mn.GetRunningAgentCount() >= expectedCount {
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for %d agents, got %d", expectedCount, mn.GetRunningAgentCount())
}

// GetDiscoveryHistory returns all discovery events
func (mn *MockNetwork) GetDiscoveryHistory() []DiscoveryEvent {
	mn.mu.RLock()
	defer mn.mu.RUnlock()
	
	events := make([]DiscoveryEvent, len(mn.discoveryEvents))
	copy(events, mn.discoveryEvents)
	return events
}

// GetRecentDiscoveryEvents returns the last N discovery events
func (mn *MockNetwork) GetRecentDiscoveryEvents(count int) []DiscoveryEvent {
	mn.mu.RLock()
	defer mn.mu.RUnlock()
	
	if len(mn.discoveryEvents) <= count {
		events := make([]DiscoveryEvent, len(mn.discoveryEvents))
		copy(events, mn.discoveryEvents)
		return events
	}
	
	events := make([]DiscoveryEvent, count)
	copy(events, mn.discoveryEvents[len(mn.discoveryEvents)-count:])
	return events
}

// GetNetworkStats returns current network statistics
func (mn *MockNetwork) GetNetworkStats() NetworkStats {
	mn.mu.RLock()
	defer mn.mu.RUnlock()
	
	running := 0
	stopped := 0
	for _, agent := range mn.agents {
		if agent.State == AgentStateRunning {
			running++
		} else {
			stopped++
		}
	}
	
	return NetworkStats{
		TotalAgents:     len(mn.agents),
		RunningAgents:   running,
		StoppedAgents:   stopped,
		TotalMessages:   mn.messageCount,
		TotalTopics:     len(mn.topics),
		ConsumerGroups:  len(mn.consumerGroups),
		DiscoveryEvents: len(mn.discoveryEvents),
	}
}

// GetNetworkID returns the network ID
func (mn *MockNetwork) GetNetworkID() string {
	return mn.networkID
}

// ClearNetwork resets the network (for test cleanup)
func (mn *MockNetwork) ClearNetwork() {
	mn.mu.Lock()
	defer mn.mu.Unlock()
	
	// Stop all agents
	for name, agent := range mn.agents {
		agent.State = AgentStateStopping
		close(agent.Inbox)
		delete(mn.agents, name)
	}
	
	// Clear all data
	mn.topics = make(map[string][]Message)
	mn.consumerGroups = make(map[string]*ConsumerGroup)
	mn.discoveryEvents = make([]DiscoveryEvent, 0)
	mn.messageCount = 0
}

// Send sends a message from one agent to another
func (mn *MockNetwork) Send(fromAgent, toAgent string, msg Message) error {
	mn.mu.RLock()
	sender, exists := mn.agents[fromAgent]
	if !exists {
		mn.mu.RUnlock()
		return fmt.Errorf("sender agent %s not found", fromAgent)
	}
	mn.mu.RUnlock()
	
	// Simulate network latency
	if mn.networkLatency > 0 {
		time.Sleep(mn.networkLatency)
	}
	
	// Simulate packet loss
	if mn.packetLoss > 0 && float64(time.Now().UnixNano()%1000)/1000 < mn.packetLoss {
		return fmt.Errorf("packet lost in network")
	}
	
	// Set message metadata
	msg.ID = uuid.New().String()
	msg.Timestamp = time.Now()
	msg.Source = fromAgent
	sender.LastActivity = time.Now()
	sender.MessageCount++
	
	mn.mu.Lock()
	mn.messageCount++
	mn.mu.Unlock()
	
	mn.mu.RLock()
	receiver, exists := mn.agents[toAgent]
	mn.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("receiver agent %s not found", toAgent)
	}
	
	if receiver.State != AgentStateRunning {
		return fmt.Errorf("receiver agent %s is not running (state: %s)", toAgent, receiver.State)
	}
	
	// Deliver message
	select {
	case receiver.Inbox <- msg:
		receiver.mu.Lock()
		receiver.Messages = append(receiver.Messages, msg)
		receiver.LastActivity = time.Now()
		receiver.LastSeen = time.Now()
		receiver.mu.Unlock()
		
		// Auto-process if enabled
		if receiver.AutoAck && receiver.ProcessFunc != nil {
			response := receiver.ProcessFunc(msg)
			if response.ID != "" {
				mn.Send(toAgent, fromAgent, response)
			}
		}
		
		return nil
	default:
		return fmt.Errorf("receiver inbox full")
	}
}

// Publish publishes a message to a topic
func (mn *MockNetwork) Publish(topic string, msg Message) error {
	mn.mu.Lock()
	defer mn.mu.Unlock()
	
	msg.ID = uuid.New().String()
	msg.Timestamp = time.Now()
	msg.Topic = topic
	
	mn.topics[topic] = append(mn.topics[topic], msg)
	
	// Notify consumer group members
	if cg, exists := mn.consumerGroups[topic]; exists {
		for _, member := range cg.Members {
			select {
			case member.Inbox <- msg:
				member.mu.Lock()
				member.Messages = append(member.Messages, msg)
				member.mu.Unlock()
			default:
				// Inbox full, skip
			}
		}
	}
	
	return nil
}

// Subscribe creates a consumer group for a topic
func (mn *MockNetwork) Subscribe(groupID string, topics []string, agent *MockAgent) error {
	mn.mu.Lock()
	defer mn.mu.Unlock()
	
	cg, exists := mn.consumerGroups[groupID]
	if !exists {
		cg = &ConsumerGroup{
			ID:      groupID,
			Topics:  topics,
			Members: make([]*MockAgent, 0),
			Offsets: make(map[string]int),
		}
		mn.consumerGroups[groupID] = cg
	}
	
	// Check if agent already member
	for _, member := range cg.Members {
		if member.Name == agent.Name {
			return nil
		}
	}
	
	cg.Members = append(cg.Members, agent)
	cg.Generation++
	
	return nil
}

// GetTopicMessages retrieves messages from a topic
func (mn *MockNetwork) GetTopicMessages(topic string, limit int) []Message {
	mn.mu.RLock()
	defer mn.mu.RUnlock()
	
	messages, exists := mn.topics[topic]
	if !exists {
		return []Message{}
	}
	
	if len(messages) <= limit {
		return messages
	}
	
	return messages[len(messages)-limit:]
}

// CreateConsumerGroup creates a new consumer group
func (mn *MockNetwork) CreateConsumerGroup(groupID string, topics []string) *ConsumerGroup {
	mn.mu.Lock()
	defer mn.mu.Unlock()
	
	cg := &ConsumerGroup{
		ID:      groupID,
		Topics:  topics,
		Members: make([]*MockAgent, 0),
		Offsets: make(map[string]int),
	}
	
	mn.consumerGroups[groupID] = cg
	return cg
}

// SetNetworkLatency sets simulated network latency
func (mn *MockNetwork) SetNetworkLatency(latency time.Duration) {
	mn.mu.Lock()
	defer mn.mu.Unlock()
	mn.networkLatency = latency
}

// SetPacketLoss sets simulated packet loss percentage (0.0-1.0)
func (mn *MockNetwork) SetPacketLoss(loss float64) {
	mn.mu.Lock()
	defer mn.mu.Unlock()
	mn.packetLoss = loss
}

// SendFile simulates sending a file between agents
func (mn *MockNetwork) SendFile(fromAgent, toAgent string, req FileTransferRequest) (string, error) {
	transferID := uuid.New().String()
	
	mn.mu.RLock()
	sender, senderExists := mn.agents[fromAgent]
	receiver, receiverExists := mn.agents[toAgent]
	mn.mu.RUnlock()
	
	if !senderExists {
		return "", fmt.Errorf("sender agent %s not found", fromAgent)
	}

	if !receiverExists {
		return "", fmt.Errorf("receiver agent %s not found", toAgent)
	}

	if sender.State != AgentStateRunning {
		return "", fmt.Errorf("sender agent %s is not running (state: %s)", fromAgent, sender.State)
	}

	if receiver.State != AgentStateRunning {
		return "", fmt.Errorf("receiver agent %s is not running (state: %s)", toAgent, receiver.State)
	}
	
	// Simulate file content
	fileContent := make([]byte, req.FileSize)
	for i := range fileContent {
		fileContent[i] = byte(i % 256)
	}
	
	// Store file in receiver
	receiver.mu.Lock()
	receiver.Files[req.FileName] = fileContent
	receiver.mu.Unlock()
	
	// Send notification message
	notification := Message{
		ID:        uuid.New().String(),
		Type:      "file_transfer",
		Source:    fromAgent,
		Dest:      toAgent,
		Topic:     "file.transfers",
		Timestamp: time.Now(),
		Body: MessageBody{
			Action: "file_received",
			Payload: []byte(fmt.Sprintf(`{"transfer_id":"%s","file_name":"%s","size":%d}`, 
				transferID, req.FileName, req.FileSize)),
		},
	}
	
	mn.Send(fromAgent, toAgent, notification)
	
	return transferID, nil
}

// GetFile retrieves a file from an agent
func (mn *MockNetwork) GetFile(agentName, fileName string) ([]byte, error) {
	mn.mu.RLock()
	defer mn.mu.RUnlock()
	
	agent, exists := mn.agents[agentName]
	if !exists {
		return nil, fmt.Errorf("agent %s not found", agentName)
	}
	
	agent.mu.RLock()
	defer agent.mu.RUnlock()
	
	file, exists := agent.Files[fileName]
	if !exists {
		return nil, fmt.Errorf("file %s not found", fileName)
	}
	
	return file, nil
}

// GetTransferProgress simulates getting transfer progress
func (mn *MockNetwork) GetTransferProgress(transferID string) (*FileTransferProgress, error) {
	// For mock, just return completed
	return &FileTransferProgress{
		TransferID:    transferID,
		Status:        "completed",
		BytesSent:     1000,
		BytesTotal:    1000,
		PercentComplete: 100.0,
		BytesPerSecond: 10000,
		ETASeconds:    0,
	}, nil
}

// StopAgent stops an agent
func (mn *MockNetwork) StopAgent(name string) error {
	mn.mu.RLock()
	defer mn.mu.RUnlock()
	
	agent, exists := mn.agents[name]
	if !exists {
		return fmt.Errorf("agent %s not found", name)
	}
	
	agent.State = AgentStateStopping
	
	// Only close if not already closed
	select {
	case <-agent.Inbox:
		// Channel already closed
	default:
		close(agent.Inbox)
	}
	
	// Record discovery event
	mn.discoveryEvents = append(mn.discoveryEvents, DiscoveryEvent{
		Type:      "agent_stopped",
		AgentName: name,
		Timestamp: time.Now(),
		Details:   fmt.Sprintf("Uptime: %v", time.Since(agent.StartTime)),
	})
	
	return nil
}

// StartAgent starts an agent
func (mn *MockNetwork) StartAgent(name string) error {
	mn.mu.RLock()
	defer mn.mu.RUnlock()
	
	agent, exists := mn.agents[name]
	if !exists {
		return fmt.Errorf("agent %s not found", name)
	}
	
	agent.State = AgentStateRunning
	agent.Inbox = make(chan Message, 100)
	agent.LastSeen = time.Now()
	
	// Record discovery event
	mn.discoveryEvents = append(mn.discoveryEvents, DiscoveryEvent{
		Type:      "agent_started",
		AgentName: name,
		Timestamp: time.Now(),
		Details:   "Agent restarted",
	})
	
	return nil
}

// GetPhoneBook returns the phone book of all agents
func (mn *MockNetwork) GetPhoneBook() []AgentInfo {
	mn.mu.RLock()
	defer mn.mu.RUnlock()
	
	phonebook := make([]AgentInfo, 0, len(mn.agents))
	for _, agent := range mn.agents {
		phonebook = append(phonebook, AgentInfo{
			Name:         agent.Name,
			NodeID:       agent.NodeID,
			Tailnet:      agent.Tailnet,
			Capabilities: agent.Capabilities,
			Online:       agent.State == AgentStateRunning,
			State:        string(agent.State),
			Uptime:       time.Since(agent.StartTime).String(),
			LastSeen:     agent.LastSeen,
			MessageCount: agent.MessageCount,
		})
	}
	
	return phonebook
}

// AgentInfo represents agent information in the phone book
type AgentInfo struct {
	Name           string    `json:"name"`
	NodeID         string    `json:"node_id"`
	Tailnet        string    `json:"tailnet"`
	Capabilities   []string  `json:"capabilities"`
	Online         bool      `json:"online"`
	State          string    `json:"state"`
	Uptime         string    `json:"uptime"`
	LastSeen       time.Time `json:"last_seen"`
	MessageCount   int       `json:"message_count"`
	TailscaleIP    string    `json:"tailscale_ip"`
	TailscaleIPv6  string    `json:"tailscale_ipv6"`
	InboundPort    int       `json:"inbound_port"`
	HTTPPort       int       `json:"http_port"`
}

// Helper function to check substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		(s == substr || len(substr) == 0 || 
			findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
