package main

import (
	"encoding/json"
	"net/http"
	"time"
)

// AIP Registration Request/Response types

// RegisterRequest is the agent registration request body
type RegisterRequest struct {
	AgentID      string         `json:"agent_id"`
	AgentType    string         `json:"agent_type"`
	AgentVersion string         `json:"agent_version"`
	Capabilities []string       `json:"capabilities"`
	Endpoints    AgentEndpoints `json:"endpoints"`
	Metadata     AgentMetadata  `json:"metadata"`
	AuthToken    string         `json:"auth_token"`
}

// RegisterResponse is the registration response
type RegisterResponse struct {
	Status         string `json:"status"`
	Message        string `json:"message"`
	RegistrationID string `json:"registration_id,omitempty"`
	NextSteps      string `json:"next_steps,omitempty"`
}

// HeartbeatRequest is the agent heartbeat request body
type HeartbeatRequest struct {
	AgentID   string                 `json:"agent_id"`
	Timestamp time.Time              `json:"timestamp"`
	Status    string                 `json:"status"`
	Metrics   map[string]interface{} `json:"metrics,omitempty"`
}

// HeartbeatResponse is the heartbeat response
type HeartbeatResponse struct {
	Status   string `json:"status"`
	AgentID  string `json:"agent_id"`
	Approved bool   `json:"approved"`
}

// PairRequest is the bridge pairing request
type PairRequest struct {
	BridgeName     string   `json:"bridge_name"`
	AuthToken      string   `json:"auth_token"`
	Agents         []AgentSummary `json:"agents"`
	RequestPeers   bool     `json:"request_peers"`
}

// AgentSummary is a brief agent description for pairing
type AgentSummary struct {
	AgentID      string   `json:"agent_id"`
	AgentType    string   `json:"agent_type"`
	Capabilities []string `json:"capabilities"`
}

// PairResponse is the bridge pairing response
type PairResponse struct {
	Status      string         `json:"status"`
	BridgeName  string         `json:"bridge_name"`
	Agents      []AgentSummary `json:"agents"`
	Message     string         `json:"message,omitempty"`
}

// AIPHandlers holds HTTP handlers for Agent Identification Protocol
type AIPHandlers struct {
	registry *AgentRegistry
}

// NewAIPHandlers creates new AIP HTTP handlers
func NewAIPHandlers(registry *AgentRegistry) *AIPHandlers {
	return &AIPHandlers{
		registry: registry,
	}
}

// HandleRegister handles POST /aip/register
func (h *AIPHandlers) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.AgentID == "" {
		http.Error(w, "agent_id is required", http.StatusBadRequest)
		return
	}
	if req.AgentType == "" {
		http.Error(w, "agent_type is required", http.StatusBadRequest)
		return
	}
	if req.Endpoints.Primary == "" {
		http.Error(w, "endpoints.primary is required", http.StatusBadRequest)
		return
	}

	// Create registered agent
	agent := RegisteredAgent{
		AgentID:      req.AgentID,
		AgentType:    req.AgentType,
		AgentVersion: req.AgentVersion,
		Capabilities: req.Capabilities,
		Endpoints:    req.Endpoints,
		Metadata:     req.Metadata,
	}

	// Register the agent
	if err := h.registry.RegisterAgent(agent); err != nil {
		if err.Error() == "agent already registered and approved" {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return pending response
	resp := RegisterResponse{
		Status:         "pending",
		Message:        "Registration received, awaiting approval",
		RegistrationID: req.AgentID,
		NextSteps:      "Contact bridge administrator to approve this registration",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// HandleHeartbeat handles POST /aip/heartbeat
func (h *AIPHandlers) HandleHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req HeartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.AgentID == "" {
		http.Error(w, "agent_id is required", http.StatusBadRequest)
		return
	}

	// Update heartbeat
	if err := h.registry.UpdateHeartbeat(req.AgentID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Check if agent is approved
	agent, exists := h.registry.GetAgent(req.AgentID)
	approved := exists && agent.Status == StatusApproved

	resp := HeartbeatResponse{
		Status:   "ok",
		AgentID:  req.AgentID,
		Approved: approved,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandlePair handles POST /aip/pair (bridge-to-bridge pairing)
func (h *AIPHandlers) HandlePair(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PairRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.BridgeName == "" {
		http.Error(w, "bridge_name is required", http.StatusBadRequest)
		return
	}

	// TODO: Validate auth token against configured shared secrets
	// For now, accept the pairing (manual approval mode)

	// Register as peer bridge
	peer := PeerBridge{
		Name:           req.BridgeName,
		TailnetAddress: r.RemoteAddr,
		AgentID:        "", // Could be filled from first agent
		Status:         "active",
	}

	if err := h.registry.RegisterPeerBridge(peer); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get our approved agents
	ourAgents := h.registry.GetApprovedAgents()
	agentSummaries := make([]AgentSummary, 0, len(ourAgents))
	for _, agent := range ourAgents {
		agentSummaries = append(agentSummaries, AgentSummary{
			AgentID:      agent.AgentID,
			AgentType:    agent.AgentType,
			Capabilities: agent.Capabilities,
		})
	}

	resp := PairResponse{
		Status:     "approved",
		BridgeName: h.registry.bridgeName,
		Agents:     agentSummaries,
		Message:    "Pairing successful",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleListAgents handles GET /aip/agents
func (h *AIPHandlers) HandleListAgents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check query parameters
	statusFilter := r.URL.Query().Get("status")
	typeFilter := r.URL.Query().Get("type")

	var agents []*RegisteredAgent

	switch statusFilter {
	case "pending":
		agents = h.registry.GetPendingAgents()
	case "approved":
		agents = h.registry.GetApprovedAgents()
	default:
		// Return all agents
		h.registry.mu.RLock()
		agents = make([]*RegisteredAgent, 0, len(h.registry.agents))
		for _, agent := range h.registry.agents {
			agentCopy := *agent
			agents = append(agents, &agentCopy)
		}
		h.registry.mu.RUnlock()
	}

	// Apply type filter if specified
	if typeFilter != "" {
		filtered := make([]*RegisteredAgent, 0)
		for _, agent := range agents {
			if agent.AgentType == typeFilter {
				filtered = append(filtered, agent)
			}
		}
		agents = filtered
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agents)
}

// HandleApproveAgent handles POST /aip/approve/{agent_id}
func (h *AIPHandlers) HandleApproveAgent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract agent_id from URL path
	// Path format: /aip/approve/{agent_id}
	path := r.URL.Path
	agentID := path[len("/aip/approve/"):]
	if agentID == "" {
		http.Error(w, "agent_id is required in path", http.StatusBadRequest)
		return
	}

	if err := h.registry.ApproveAgent(agentID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	resp := map[string]string{
		"status":  "approved",
		"agent_id": agentID,
		"message": "Agent approved successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleRejectAgent handles POST /aip/reject/{agent_id}
func (h *AIPHandlers) HandleRejectAgent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Path
	agentID := path[len("/aip/reject/"):]
	if agentID == "" {
		http.Error(w, "agent_id is required in path", http.StatusBadRequest)
		return
	}

	if err := h.registry.RejectAgent(agentID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	resp := map[string]string{
		"status":  "rejected",
		"agent_id": agentID,
		"message": "Agent registration rejected",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
