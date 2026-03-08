package aip

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// HTTPHandler handles HTTP endpoints for AIP protocol
type HTTPHandler struct {
	client *AIPClient
	mux    *http.ServeMux
	server *http.Server
}

// NewHTTPHandler creates a new HTTP handler for AIP endpoints
func NewHTTPHandler(client *AIPClient, listenAddr string) *HTTPHandler {
	mux := http.NewServeMux()
	h := &HTTPHandler{
		client: client,
		mux:    mux,
	}

	// Register handlers
	mux.HandleFunc("/aip/handshake", h.handleHandshake)
	mux.HandleFunc("/health", h.handleHealth)

	h.server = &http.Server{
		Addr:         listenAddr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return h
}

// handleHandshake handles POST /aip/handshake
func (h *HTTPHandler) handleHandshake(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse handshake request
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

	// Handle handshake
	resp, err := h.client.HandleHandshake(req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "handshake_failed",
			"message": err.Error(),
		})
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[aip] failed to encode response: %v", err)
	}

	log.Printf("[aip] handshake completed for agent %s", resp.AgentID)
}

// handleHealth handles GET /health
func (h *HTTPHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := map[string]interface{}{
		"status":        "healthy",
		"agent_id":      h.client.GetAgentID(),
		"approved":      h.client.IsApproved(),
		"last_heartbeat": h.client.GetLastHeartbeat(),
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// Start starts the HTTP server
func (h *HTTPHandler) Start() error {
	go func() {
		log.Printf("[aip] starting HTTP server on %s", h.server.Addr)
		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[aip] HTTP server error: %v", err)
		}
	}()
	return nil
}

// Stop stops the HTTP server
func (h *HTTPHandler) Stop() error {
	if h.server != nil {
		return h.server.Close()
	}
	return nil
}

// GetListenAddr returns the listen address
func (h *HTTPHandler) GetListenAddr() string {
	return h.server.Addr
}

// AgentServer wraps the AIP client and HTTP handler for easy usage
type AgentServer struct {
	client  *AIPClient
	handler *HTTPHandler
}

// AgentServerConfig configures the agent server
type AgentServerConfig struct {
	BridgeURL      string
	AgentID        string
	Secret         string
	ListenAddr     string
	EndpointURL    string
	HealthURL      string
	Capabilities   []string
	HeartbeatInterval time.Duration
}

// NewAgentServer creates a new agent server
func NewAgentServer(cfg AgentServerConfig) (*AgentServer, error) {
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = ":9090"
	}
	if cfg.EndpointURL == "" {
		cfg.EndpointURL = fmt.Sprintf("http://127.0.0.1%s/api", cfg.ListenAddr)
	}
	if cfg.HealthURL == "" {
		cfg.HealthURL = fmt.Sprintf("http://127.0.0.1%s/health", cfg.ListenAddr)
	}
	if cfg.BridgeURL == "" {
		cfg.BridgeURL = "http://127.0.0.1:8080"
	}
	if cfg.HeartbeatInterval == 0 {
		cfg.HeartbeatInterval = 30 * time.Second
	}

	// Get hostname
	hostname := "darci-go-agent"

	// Create AIP client
	client := NewAIPClient(AIPClientConfig{
		BridgeURL:    cfg.BridgeURL,
		AgentID:      cfg.AgentID,
		AgentType:    "darci-go",
		AgentVersion: "1.0.0",
		Secret:       cfg.Secret,
		Capabilities: cfg.Capabilities,
		EndpointURL:  cfg.EndpointURL,
		HealthURL:    cfg.HealthURL,
		Hostname:     hostname,
	})

	// Create HTTP handler
	handler := NewHTTPHandler(client, cfg.ListenAddr)

	return &AgentServer{
		client:  client,
		handler: handler,
	}, nil
}

// Start starts the agent server (HTTP + heartbeat)
func (s *AgentServer) Start() error {
	// Start HTTP server
	if err := s.handler.Start(); err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return nil
}

// RegisterAndStartHeartbeat registers with bridge and starts heartbeat
func (s *AgentServer) RegisterAndStartHeartbeat(ctx context.Context) error {
	// Register with bridge
	resp, err := s.client.Register(ctx)
	if err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}

	log.Printf("[aip] registration status: %s - %s", resp.Status, resp.Message)

	// Start heartbeat loop
	s.client.StartHeartbeatLoop(ctx, 30*time.Second)

	return nil
}

// Stop stops the agent server
func (s *AgentServer) Stop() error {
	s.client.StopHeartbeatLoop()
	return s.handler.Stop()
}

// GetClient returns the AIP client
func (s *AgentServer) GetClient() *AIPClient {
	return s.client
}
