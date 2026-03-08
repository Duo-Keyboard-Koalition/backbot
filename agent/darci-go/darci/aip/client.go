package aip

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// HandshakeRequest is sent by bridge to verify agent identity
type HandshakeRequest struct {
	Challenge       string `json:"challenge"`
	Timestamp       string `json:"timestamp"`
	Nonce           string `json:"nonce"`
	BridgeID        string `json:"bridge_id"`
	ProtocolVersion string `json:"protocol_version"`
}

// HandshakeResponse is the agent's response with signature
type HandshakeResponse struct {
	Signature    string   `json:"signature"`
	AgentID      string   `json:"agent_id"`
	AgentType    string   `json:"agent_type"`
	AgentVersion string   `json:"agent_version"`
	Capabilities []string `json:"capabilities"`
}

// RegisterRequest is sent to register agent with bridge
type RegisterRequest struct {
	AgentID      string         `json:"agent_id"`
	AgentType    string         `json:"agent_type"`
	AgentVersion string         `json:"agent_version"`
	Capabilities []string       `json:"capabilities"`
	Endpoints    AgentEndpoints `json:"endpoints"`
	Metadata     AgentMetadata  `json:"metadata"`
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

// RegisterResponse is the registration response
type RegisterResponse struct {
	Status         string `json:"status"`
	Message        string `json:"message"`
	RegistrationID string `json:"registration_id,omitempty"`
	NextSteps      string `json:"next_steps,omitempty"`
}

// HeartbeatRequest is sent periodically to maintain active status
type HeartbeatRequest struct {
	AgentID   string                 `json:"agent_id"`
	Timestamp string                 `json:"timestamp"`
	Status    string                 `json:"status"`
	Metrics   map[string]interface{} `json:"metrics,omitempty"`
}

// HeartbeatResponse is the heartbeat response
type HeartbeatResponse struct {
	Status   string `json:"status"`
	AgentID  string `json:"agent_id"`
	Approved bool   `json:"approved"`
}

// AIPClient handles agent registration and communication with bridge
type AIPClient struct {
	bridgeURL      string
	agentID        string
	agentType      string
	agentVersion   string
	secret         string
	capabilities   []string
	endpointURL    string
	healthURL      string
	hostname       string
	httpClient     *http.Client
	mu             sync.RWMutex
	approved       bool
	lastHeartbeat  time.Time
	heartbeatStop  chan struct{}
}

// AIPClientConfig configures the AIP client
type AIPClientConfig struct {
	BridgeURL    string
	AgentID      string
	AgentType    string
	AgentVersion string
	Secret       string
	Capabilities []string
	EndpointURL  string
	HealthURL    string
	Hostname     string
}

// NewAIPClient creates a new AIP client
func NewAIPClient(cfg AIPClientConfig) *AIPClient {
	if cfg.BridgeURL == "" {
		cfg.BridgeURL = "http://127.0.0.1:8080"
	}
	if cfg.AgentType == "" {
		cfg.AgentType = "darci-go"
	}
	if cfg.AgentVersion == "" {
		cfg.AgentVersion = "1.0.0"
	}
	if cfg.Hostname == "" {
		cfg.Hostname = "unknown"
	}

	return &AIPClient{
		bridgeURL:    cfg.BridgeURL,
		agentID:      cfg.AgentID,
		agentType:    cfg.AgentType,
		agentVersion: cfg.AgentVersion,
		secret:       cfg.Secret,
		capabilities: cfg.Capabilities,
		endpointURL:  cfg.EndpointURL,
		healthURL:    cfg.HealthURL,
		hostname:     cfg.Hostname,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		heartbeatStop: make(chan struct{}),
	}
}

// computeSignature computes HMAC-SHA256 signature for handshake
func computeSignature(challenge, timestamp, nonce, secret string) string {
	message := fmt.Sprintf("%s:%s:%s", challenge, timestamp, nonce)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// HandleHandshake processes incoming handshake challenge from bridge
func (c *AIPClient) HandleHandshake(req HandshakeRequest) (*HandshakeResponse, error) {
	c.mu.RLock()
	secret := c.secret
	agentID := c.agentID
	agentType := c.agentType
	agentVersion := c.agentVersion
	capabilities := c.capabilities
	c.mu.RUnlock()

	// Validate request
	if req.Challenge == "" || req.Timestamp == "" || req.Nonce == "" {
		return nil, fmt.Errorf("invalid handshake request")
	}

	// Check challenge expiry
	challengeTime, err := time.Parse(time.RFC3339, req.Timestamp)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp format")
	}
	if time.Since(challengeTime) > 30*time.Second {
		return nil, fmt.Errorf("challenge expired")
	}

	// Compute signature
	signature := computeSignature(req.Challenge, req.Timestamp, req.Nonce, secret)

	return &HandshakeResponse{
		Signature:    signature,
		AgentID:      agentID,
		AgentType:    agentType,
		AgentVersion: agentVersion,
		Capabilities: capabilities,
	}, nil
}

// Register registers the agent with the bridge
func (c *AIPClient) Register(ctx context.Context) (*RegisterResponse, error) {
	c.mu.RLock()
	req := RegisterRequest{
		AgentID:      c.agentID,
		AgentType:    c.agentType,
		AgentVersion: c.agentVersion,
		Capabilities: c.capabilities,
		Endpoints: AgentEndpoints{
			Primary: c.endpointURL,
			Health:  c.healthURL,
		},
		Metadata: AgentMetadata{
			Hostname: c.hostname,
			OS:       "linux",
			Tags:     []string{"darci", "go"},
		},
	}
	c.mu.RUnlock()

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/aip/register", c.bridgeURL)
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("registration request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registration failed with status %d: %s", resp.StatusCode, string(body))
	}

	var registerResp RegisterResponse
	if err := json.Unmarshal(body, &registerResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &registerResp, nil
}

// SendHeartbeat sends a heartbeat to the bridge
func (c *AIPClient) SendHeartbeat(ctx context.Context) (*HeartbeatResponse, error) {
	c.mu.RLock()
	req := HeartbeatRequest{
		AgentID:   c.agentID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Status:    "healthy",
		Metrics: map[string]interface{}{
			"cpu_usage":     0.0,
			"memory_mb":     0,
			"active_tasks":  0,
			"uptime_seconds": 0,
		},
	}
	c.mu.RUnlock()

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/aip/heartbeat", c.bridgeURL)
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("heartbeat request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("heartbeat failed with status %d: %s", resp.StatusCode, string(body))
	}

	var heartbeatResp HeartbeatResponse
	if err := json.Unmarshal(body, &heartbeatResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	c.mu.Lock()
	c.approved = heartbeatResp.Approved
	c.lastHeartbeat = time.Now()
	c.mu.Unlock()

	return &heartbeatResp, nil
}

// StartHeartbeatLoop starts periodic heartbeat in background
func (c *AIPClient) StartHeartbeatLoop(ctx context.Context, interval time.Duration) {
	if interval == 0 {
		interval = 30 * time.Second
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-c.heartbeatStop:
				return
			case <-ticker.C:
				if _, err := c.SendHeartbeat(ctx); err != nil {
					fmt.Printf("[aip] heartbeat error: %v\n", err)
				}
			}
		}
	}()
}

// StopHeartbeatLoop stops the heartbeat loop
func (c *AIPClient) StopHeartbeatLoop() {
	close(c.heartbeatStop)
}

// IsApproved returns true if agent is approved by bridge
func (c *AIPClient) IsApproved() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.approved
}

// GetLastHeartbeat returns the last heartbeat time
func (c *AIPClient) GetLastHeartbeat() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastHeartbeat
}

// GetAgentID returns the agent ID
func (c *AIPClient) GetAgentID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.agentID
}
