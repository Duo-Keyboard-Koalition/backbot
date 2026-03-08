package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AgentStatus represents the registration status of an agent
type AgentStatus string

const (
	StatusPending   AgentStatus = "pending"
	StatusApproved  AgentStatus = "approved"
	StatusRejected  AgentStatus = "rejected"
	StatusOffline   AgentStatus = "offline"
	StatusRemoved   AgentStatus = "removed"
)

// RegisteredAgent represents an agent registered with the bridge
type RegisteredAgent struct {
	AgentID       string            `json:"agent_id"`
	AgentType     string            `json:"agent_type"`
	AgentVersion  string            `json:"agent_version"`
	Status        AgentStatus       `json:"status"`
	RegisteredAt  time.Time         `json:"registered_at"`
	ApprovedAt    *time.Time        `json:"approved_at,omitempty"`
	LastHeartbeat time.Time         `json:"last_heartbeat"`
	Endpoints     AgentEndpoints    `json:"endpoints"`
	Capabilities  []string          `json:"capabilities"`
	Metadata      AgentMetadata     `json:"metadata"`
}

// AgentEndpoints contains agent endpoint URLs
type AgentEndpoints struct {
	Primary string `json:"primary"`
	Health  string `json:"health,omitempty"`
}

// AgentMetadata contains optional agent metadata
type AgentMetadata struct {
	Hostname string   `json:"hostname"`
	OS       string   `json:"os,omitempty"`
	Tags     []string `json:"tags"`
}

// PeerBridge represents a paired peer bridge
type PeerBridge struct {
	Name           string    `json:"name"`
	TailnetAddress string    `json:"tailnet_address"`
	AgentID        string    `json:"agent_id"`
	Status         string    `json:"status"`
	LastSeen       time.Time `json:"last_seen"`
	PairedAt       time.Time `json:"paired_at"`
}

// AgentRegistry manages registered agents and peer bridges
type AgentRegistry struct {
	filepath      string
	bridgeName    string
	agents        map[string]*RegisteredAgent
	peerBridges   map[string]*PeerBridge
	mu            sync.RWMutex
	dirty         bool
}

// RegistryConfig contains registry configuration
type RegistryConfig struct {
	Version       string          `json:"version"`
	ThisBridge    string          `json:"this_bridge"`
	RegisteredAgents []RegisteredAgent `json:"registered_agents"`
	PeerBridges   []PeerBridge    `json:"peer_bridges"`
}

// NewAgentRegistry creates a new agent registry
func NewAgentRegistry(stateDir, bridgeName string) (*AgentRegistry, error) {
	registryPath := filepath.Join(stateDir, "registry.json")

	registry := &AgentRegistry{
		filepath:    registryPath,
		bridgeName:  bridgeName,
		agents:      make(map[string]*RegisteredAgent),
		peerBridges: make(map[string]*PeerBridge),
	}

	// Load existing registry if it exists
	if _, err := os.Stat(registryPath); err == nil {
		if err := registry.load(); err != nil {
			return nil, fmt.Errorf("failed to load registry: %w", err)
		}
	}

	return registry, nil
}

// load reads the registry from disk
func (r *AgentRegistry) load() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := os.ReadFile(r.filepath)
	if err != nil {
		return err
	}

	var config RegistryConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse registry: %w", err)
	}

	// Index agents by ID
	for _, agent := range config.RegisteredAgents {
		r.agents[agent.AgentID] = &agent
	}

	// Index peer bridges by name
	for _, peer := range config.PeerBridges {
		r.peerBridges[peer.Name] = &peer
	}

	return nil
}

// save writes the registry to disk
func (r *AgentRegistry) save() error {
	if !r.dirty {
		return nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	config := RegistryConfig{
		Version:    "1.0",
		ThisBridge: r.bridgeName,
		RegisteredAgents: make([]RegisteredAgent, 0, len(r.agents)),
		PeerBridges: make([]PeerBridge, 0, len(r.peerBridges)),
	}

	for _, agent := range r.agents {
		config.RegisteredAgents = append(config.RegisteredAgents, *agent)
	}

	for _, peer := range r.peerBridges {
		config.PeerBridges = append(config.PeerBridges, *peer)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}

	if err := os.WriteFile(r.filepath, data, 0600); err != nil {
		return fmt.Errorf("failed to write registry: %w", err)
	}

	r.dirty = false
	return nil
}

// RegisterAgent adds a new agent registration
func (r *AgentRegistry) RegisterAgent(agent RegisteredAgent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if already registered
	if existing, exists := r.agents[agent.AgentID]; exists {
		if existing.Status == StatusApproved {
			return fmt.Errorf("agent %s is already registered and approved", agent.AgentID)
		}
	}

	agent.RegisteredAt = time.Now()
	agent.Status = StatusPending
	agent.LastHeartbeat = time.Now()

	r.agents[agent.AgentID] = &agent
	r.dirty = true

	return r.save()
}

// ApproveAgent approves a pending registration
func (r *AgentRegistry) ApproveAgent(agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	agent, exists := r.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	if agent.Status == StatusApproved {
		return fmt.Errorf("agent %s is already approved", agentID)
	}

	now := time.Now()
	agent.Status = StatusApproved
	agent.ApprovedAt = &now
	r.dirty = true

	return r.save()
}

// RejectAgent rejects a pending registration
func (r *AgentRegistry) RejectAgent(agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	agent, exists := r.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	agent.Status = StatusRejected
	agent.RejectedAt = time.Now()
	r.dirty = true

	return r.save()
}

// RemoveAgent removes an agent from the registry
func (r *AgentRegistry) RemoveAgent(agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.agents[agentID]; !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	delete(r.agents, agentID)
	r.dirty = true

	return r.save()
}

// GetAgent retrieves an agent by ID
func (r *AgentRegistry) GetAgent(agentID string) (*RegisteredAgent, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agent, exists := r.agents[agentID]
	if !exists {
		return nil, false
	}

	// Return a copy
	agentCopy := *agent
	return &agentCopy, true
}

// GetApprovedAgents returns all approved agents
func (r *AgentRegistry) GetApprovedAgents() []*RegisteredAgent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*RegisteredAgent, 0)
	for _, agent := range r.agents {
		if agent.Status == StatusApproved {
			agentCopy := *agent
			result = append(result, &agentCopy)
		}
	}

	return result
}

// GetPendingAgents returns all pending registrations
func (r *AgentRegistry) GetPendingAgents() []*RegisteredAgent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*RegisteredAgent, 0)
	for _, agent := range r.agents {
		if agent.Status == StatusPending {
			agentCopy := *agent
			result = append(result, &agentCopy)
		}
	}

	return result
}

// UpdateHeartbeat updates the last heartbeat time for an agent
func (r *AgentRegistry) UpdateHeartbeat(agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	agent, exists := r.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	agent.LastHeartbeat = time.Now()
	r.dirty = true

	return r.save()
}

// GetOfflineAgents returns agents that haven't sent heartbeat within duration
func (r *AgentRegistry) GetOfflineAgents(timeout time.Duration) []*RegisteredAgent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cutoff := time.Now().Add(-timeout)
	result := make([]*RegisteredAgent, 0)

	for _, agent := range r.agents {
		if agent.Status == StatusApproved && agent.LastHeartbeat.Before(cutoff) {
			agentCopy := *agent
			agentCopy.Status = StatusOffline
			result = append(result, &agentCopy)
		}
	}

	return result
}

// RegisterPeerBridge adds a peer bridge to the registry
func (r *AgentRegistry) RegisterPeerBridge(peer PeerBridge) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	peer.PairedAt = time.Now()
	peer.LastSeen = time.Now()
	peer.Status = "active"

	r.peerBridges[peer.Name] = &peer
	r.dirty = true

	return r.save()
}

// GetPeerBridges returns all registered peer bridges
func (r *AgentRegistry) GetPeerBridges() []*PeerBridge {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*PeerBridge, 0, len(r.peerBridges))
	for _, peer := range r.peerBridges {
		peerCopy := *peer
		result = append(result, &peerCopy)
	}

	return result
}

// GetPeerBridge returns a specific peer bridge by name
func (r *AgentRegistry) GetPeerBridge(name string) (*PeerBridge, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	peer, exists := r.peerBridges[name]
	if !exists {
		return nil, false
	}

	peerCopy := *peer
	return peerCopy, true
}

// RemovePeerBridge removes a peer bridge from the registry
func (r *AgentRegistry) RemovePeerBridge(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.peerBridges[name]; !exists {
		return fmt.Errorf("peer bridge %s not found", name)
	}

	delete(r.peerBridges, name)
	r.dirty = true

	return r.save()
}

// GetAllAgentsJSON returns the full registry as JSON
func (r *AgentRegistry) GetAllAgentsJSON() (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agents := make([]RegisteredAgent, 0, len(r.agents))
	for _, agent := range r.agents {
		agents = append(agents, *agent)
	}

	data, err := json.MarshalIndent(agents, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}
