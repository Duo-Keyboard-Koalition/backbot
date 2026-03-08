package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// HandshakeRequest is sent to potential agents
type HandshakeRequest struct {
	Challenge    string `json:"challenge"`
	Timestamp    string `json:"timestamp"`
	Nonce        string `json:"nonce"`
	BridgeID     string `json:"bridge_id"`
	ProtocolVersion string `json:"protocol_version"`
}

// HandshakeResponse is received from verified agents
type HandshakeResponse struct {
	Signature   string   `json:"signature"`
	AgentID     string   `json:"agent_id"`
	AgentType   string   `json:"agent_type"`
	AgentVersion string  `json:"agent_version,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
}

// PendingChallenge tracks an outstanding handshake challenge
type PendingChallenge struct {
	Challenge string
	Nonce     string
	Timestamp time.Time
	TargetIP  string
	Used      bool
}

// HandshakeService manages agent verification via challenge-response
type HandshakeService struct {
	bridgeID       string
	secrets        map[string]string // agent_id -> shared secret
	challenges     map[string]*PendingChallenge // challenge_id -> challenge
	nonceStore     map[string]time.Time // nonce -> timestamp
	mu             sync.RWMutex
	httpClient     *http.Client
	challengeExpiry time.Duration
	nonceExpiry    time.Duration
}

// NewHandshakeService creates a new handshake service
func NewHandshakeService(bridgeID string, secrets map[string]string) *HandshakeService {
	return &HandshakeService{
		bridgeID:        bridgeID,
		secrets:         secrets,
		challenges:      make(map[string]*PendingChallenge),
		nonceStore:      make(map[string]time.Time),
		httpClient:      &http.Client{Timeout: 5 * time.Second},
		challengeExpiry: 30 * time.Second,
		nonceExpiry:     5 * time.Minute,
	}
}

// generateChallenge creates a cryptographically secure random challenge
func (h *HandshakeService) generateChallenge() (string, error) {
	bytes := make([]byte, 16) // 128 bits
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// generateNonce creates a unique nonce
func (h *HandshakeService) generateNonce() (string, error) {
	bytes := make([]byte, 8) // 64 bits
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// computeSignature computes HMAC-SHA256 signature
func computeSignature(challenge, timestamp, nonce, secret string) string {
	message := fmt.Sprintf("%s:%s:%s", challenge, timestamp, nonce)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// verifySignature verifies a handshake signature
func verifySignature(challenge, timestamp, nonce, signature, secret string) bool {
	expectedSignature := computeSignature(challenge, timestamp, nonce, secret)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// storeChallenge stores a pending challenge
func (h *HandshakeService) storeChallenge(challengeID, challenge, nonce, targetIP string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.challenges[challengeID] = &PendingChallenge{
		Challenge: challenge,
		Nonce:     nonce,
		Timestamp: time.Now(),
		TargetIP:  targetIP,
		Used:      false,
	}

	h.nonceStore[nonce] = time.Now()
}

// getChallenge retrieves a pending challenge
func (h *HandshakeService) getChallenge(challengeID string) (*PendingChallenge, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	challenge, exists := h.challenges[challengeID]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Since(challenge.Timestamp) > h.challengeExpiry {
		return nil, false
	}

	return challenge, true
}

// markChallengeUsed marks a challenge as used
func (h *HandshakeService) markChallengeUsed(challengeID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if challenge, exists := h.challenges[challengeID]; exists {
		challenge.Used = true
	}
}

// isNonceUsed checks if a nonce has been used
func (h *HandshakeService) isNonceUsed(nonce string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	_, exists := h.nonceStore[nonce]
	return exists
}

// cleanupExpired removes expired challenges and nonces
func (h *HandshakeService) cleanupExpired() {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now()
	challengeCutoff := now.Add(-h.challengeExpiry)
	nonceCutoff := now.Add(-h.nonceExpiry)

	// Cleanup challenges
	for id, challenge := range h.challenges {
		if challenge.Timestamp.Before(challengeCutoff) || challenge.Used {
			delete(h.challenges, id)
		}
	}

	// Cleanup nonces
	for nonce, timestamp := range h.nonceStore {
		if timestamp.Before(nonceCutoff) {
			delete(h.nonceStore, nonce)
		}
	}
}

// StartCleanup starts periodic cleanup goroutine
func (h *HandshakeService) StartCleanup(interval time.Duration, stopChan <-chan struct{}) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				h.cleanupExpired()
			case <-stopChan:
				return
			}
		}
	}()
}

// SendHandshake sends a handshake challenge to a potential agent
func (h *HandshakeService) SendHandshake(ip string, port int) (*HandshakeResponse, error) {
	// Generate challenge components
	challenge, err := h.generateChallenge()
	if err != nil {
		return nil, fmt.Errorf("failed to generate challenge: %w", err)
	}

	nonce, err := h.generateNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)

	// Create challenge ID for tracking
	challengeID := fmt.Sprintf("%s:%s", ip, challenge)

	// Store the challenge
	h.storeChallenge(challengeID, challenge, nonce, ip)

	// Create handshake request
	req := HandshakeRequest{
		Challenge:       challenge,
		Timestamp:       timestamp,
		Nonce:           nonce,
		BridgeID:        h.bridgeID,
		ProtocolVersion: "1.0",
	}

	// Send request
	url := fmt.Sprintf("http://%s:%d/aip/handshake", ip, port)
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := h.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		// Agent not reachable - not necessarily an error, might not be an agent
		return nil, fmt.Errorf("handshake request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for error responses
	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		if err := json.Unmarshal(body, &errResp); err == nil {
			if errMsg, ok := errResp["error"].(string); ok {
				return nil, fmt.Errorf("handshake error: %s", errMsg)
			}
		}
		return nil, fmt.Errorf("handshake failed with status %d", resp.StatusCode)
	}

	// Parse response
	var handshakeResp HandshakeResponse
	if err := json.Unmarshal(body, &handshakeResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Verify signature
	secret, exists := h.secrets[handshakeResp.AgentID]
	if !exists {
		// Unknown agent - return response but mark as unverified
		handshakeResp.Signature = "" // Clear signature for security
		return &handshakeResp, fmt.Errorf("unknown agent: %s", handshakeResp.AgentID)
	}

	valid := verifySignature(challenge, timestamp, nonce, handshakeResp.Signature, secret)
	if !valid {
		return nil, fmt.Errorf("invalid signature for agent %s", handshakeResp.AgentID)
	}

	// Mark challenge as used
	h.markChallengeUsed(challengeID)

	// Agent verified!
	return &handshakeResp, nil
}

// VerifyAgentResponse verifies a handshake response from an agent
func (h *HandshakeService) VerifyAgentResponse(challengeID string, resp *HandshakeResponse) error {
	// Get the challenge
	challenge, exists := h.getChallenge(challengeID)
	if !exists {
		return fmt.Errorf("challenge not found or expired")
	}

	// Check if already used
	if challenge.Used {
		return fmt.Errorf("challenge already used - possible replay attack")
	}

	// Get secret for agent
	secret, exists := h.secrets[resp.AgentID]
	if !exists {
		return fmt.Errorf("unknown agent: %s", resp.AgentID)
	}

	// Verify signature
	valid := verifySignature(challenge.Challenge, challenge.Timestamp.Format(time.RFC3339), challenge.Nonce, resp.Signature, secret)
	if !valid {
		return fmt.Errorf("invalid signature")
	}

	// Mark as used
	h.markChallengeUsed(challengeID)

	return nil
}

// AddSecret adds a shared secret for an agent
func (h *HandshakeService) AddSecret(agentID, secret string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.secrets[agentID] = secret
}

// RemoveSecret removes a shared secret
func (h *HandshakeService) RemoveSecret(agentID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.secrets, agentID)
}

// GetStats returns handshake service statistics
func (h *HandshakeService) GetStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return map[string]interface{}{
		"pending_challenges": len(h.challenges),
		"stored_nonces":      len(h.nonceStore),
		"registered_secrets": len(h.secrets),
	}
}

// HandshakeHandler creates HTTP handlers for handshake endpoints
type HandshakeHandler struct {
	service *HandshakeService
	agentID string
	secret  string
}

// NewHandshakeHandler creates a new handshake handler
func NewHandshakeHandler(service *HandshakeService, agentID, secret string) *HandshakeHandler {
	return &HandshakeHandler{
		service: service,
		agentID: agentID,
		secret:  secret,
	}
}

// HandleHandshake handles POST /aip/handshake (agent side)
func (h *HandshakeHandler) HandleHandshake(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req HandshakeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "invalid_json",
			"message": err.Error(),
		})
		return
	}

	// Validate required fields
	if req.Challenge == "" || req.Timestamp == "" || req.Nonce == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "invalid_challenge",
			"message": "challenge, timestamp, and nonce are required",
		})
		return
	}

	// Check if challenge is expired
	challengeTime, err := time.Parse(time.RFC3339, req.Timestamp)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "invalid_timestamp",
			"message": "timestamp must be in RFC3339 format",
		})
		return
	}

	if time.Since(challengeTime) > 30*time.Second {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "challenge_expired",
			"message": "challenge has expired (max 30 seconds)",
		})
		return
	}

	// Check if nonce was already used (replay attack prevention)
	if h.service.isNonceUsed(req.Nonce) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "nonce_reused",
			"message": "nonce has already been used",
		})
		return
	}

	// Compute signature
	signature := computeSignature(req.Challenge, req.Timestamp, req.Nonce, h.secret)

	// Store nonce as used
	h.service.mu.Lock()
	h.service.nonceStore[req.Nonce] = time.Now()
	h.service.mu.Unlock()

	// Return response
	resp := HandshakeResponse{
		Signature:    signature,
		AgentID:      h.agentID,
		AgentType:    "darci-python", // Could be dynamic
		AgentVersion: "1.0.0",
		Capabilities: []string{"task-execution", "notebook", "file-ops"},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// HandleHandshakeProbe handles POST /aip/handshake-probe (bridge side)
func (h *HandshakeHandler) HandleHandshakeProbe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req struct {
		TargetIP  string `json:"target_ip"`
		TargetPort int    `json:"target_port"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.TargetIP == "" || req.TargetPort == 0 {
		http.Error(w, "target_ip and target_port required", http.StatusBadRequest)
		return
	}

	// Send handshake
	resp, err := h.service.SendHandshake(req.TargetIP, req.TargetPort)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // Return 200 with error in body
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"agent":   resp,
	})
}
