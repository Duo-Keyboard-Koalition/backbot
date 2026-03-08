package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/codejedi-ai/kaggle-for-tensors/tailscale-app/internal/models"
)

// Taila2aController handles the core business logic and coordination
type Taila2aController struct {
	name            string
	localAgentURL   string
	peerInboundPort int
	discoverySvc    *models.DiscoveryService
	tailnetClient   *http.Client
	localClient     *http.Client
	notifier        *models.ConsoleTUINotifier
}

// NewTaila2aController creates a new Taila2a controller
func NewTaila2aController(
	name string,
	localAgentURL string,
	peerInboundPort int,
	discoverySvc *models.DiscoveryService,
	tailnetClient *http.Client,
	localClient *http.Client,
	notifier *models.ConsoleTUINotifier,
) *Taila2aController {
	return &Taila2aController{
		name:            name,
		localAgentURL:   localAgentURL,
		peerInboundPort: peerInboundPort,
		discoverySvc:    discoverySvc,
		tailnetClient:   tailnetClient,
		localClient:     localClient,
		notifier:        notifier,
	}
}

// HandleInbound processes incoming messages from the tailnet
func (c *Taila2aController) HandleInbound(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 10<<20) // 10MB max
	defer r.Body.Close()

	var env models.Envelope
	if err := json.NewDecoder(r.Body).Decode(&env); err != nil {
		http.Error(w, "invalid envelope json", http.StatusBadRequest)
		return
	}

	if err := c.validateInboundEnvelope(&env); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Forward payload to local agent
	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, c.localAgentURL, bytes.NewReader(env.Payload))
	if err != nil {
		http.Error(w, "failed to create local agent request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.localClient.Do(req)
	if err != nil {
		http.Error(w, "local agent unreachable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	copyHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("taila2a copy response body error: %v", err)
	}

	log.Printf("[agnes:%s] delivered inbound payload from %s (%d)", c.name, env.SourceNode, resp.StatusCode)
	if c.notifier != nil {
		msgType := payloadType(env.Payload)
		c.notifier.LogMessage(fmt.Sprintf("← %s: %s", env.SourceNode, msgType))
	}
}

// HandleSend processes outgoing messages to the tailnet
func (c *Taila2aController) HandleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 10<<20) // 10MB max
	defer r.Body.Close()

	var env models.Envelope
	if err := json.NewDecoder(r.Body).Decode(&env); err != nil {
		http.Error(w, "invalid envelope json", http.StatusBadRequest)
		return
	}

	if err := c.normalizeOutboundEnvelope(&env); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	payload, err := json.Marshal(env)
	if err != nil {
		http.Error(w, "failed to encode envelope", http.StatusInternalServerError)
		return
	}

	targetURL := fmt.Sprintf("http://%s:%d/inbound", env.DestNode, c.peerInboundPort)
	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, targetURL, bytes.NewReader(payload))
	if err != nil {
		http.Error(w, "failed to create destination request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.tailnetClient.Do(req)
	if err != nil {
		http.Error(w, "destination agnes unreachable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	copyHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("agnes outbound copy response error: %v", err)
	}
	log.Printf("[agnes:%s] routed outbound to %s (%d)", c.name, env.DestNode, resp.StatusCode)
	if c.notifier != nil {
		msgType := payloadType(env.Payload)
		c.notifier.LogMessage(fmt.Sprintf("→ %s: %s", env.DestNode, msgType))
	}
}

// HandleAgents returns discovered agents
func (c *Taila2aController) HandleAgents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if c.discoverySvc == nil {
		http.Error(w, "discovery service not initialized", http.StatusInternalServerError)
		return
	}

	agents := c.discoverySvc.GetOnlineAgents()
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"agents": agents,
		"count":  len(agents),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[agents] failed to encode response: %v", err)
	}
}

// ExecuteCommand executes a command on a target agent
func (c *Taila2aController) ExecuteCommand(ctx context.Context, cmd models.Command) (models.Response, error) {
	env := models.Envelope{
		SourceNode: c.name,
		DestNode:   cmd.Target,
		Payload:    nil,
		Timestamp:  time.Now(),
	}

	// Marshal command to payload
	payload, err := json.Marshal(map[string]interface{}{
		"type":    "command",
		"command": cmd,
	})
	if err != nil {
		return models.Response{Success: false, Error: err.Error()}, err
	}
	env.Payload = payload

	// Send command via tailnet
	targetURL := fmt.Sprintf("http://%s:%d/inbound", cmd.Target, c.peerInboundPort)
	envBytes, err := json.Marshal(env)
	if err != nil {
		return models.Response{Success: false, Error: err.Error()}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(envBytes))
	if err != nil {
		return models.Response{Success: false, Error: err.Error()}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.tailnetClient.Do(req)
	if err != nil {
		return models.Response{Success: false, Error: err.Error()}, err
	}
	defer resp.Body.Close()

	var response models.Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return models.Response{Success: false, Error: err.Error()}, err
	}

	return response, nil
}

func (c *Taila2aController) validateInboundEnvelope(env *models.Envelope) error {
	if env.SourceNode == "" {
		return fmt.Errorf("source_node is required")
	}
	if env.DestNode == "" {
		return fmt.Errorf("dest_node is required")
	}
	if len(env.Payload) == 0 {
		return fmt.Errorf("payload is required")
	}
	return nil
}

func (c *Taila2aController) normalizeOutboundEnvelope(env *models.Envelope) error {
	if env.SourceNode == "" {
		env.SourceNode = c.name
	}
	if env.DestNode == "" {
		return fmt.Errorf("dest_node is required")
	}
	if len(env.Payload) == 0 {
		return fmt.Errorf("payload is required")
	}
	if env.Timestamp.IsZero() {
		env.Timestamp = time.Now()
	}
	return nil
}

// payloadType extracts the "type" field from a JSON payload for display.
// Returns "unknown" if the field is absent or unparseable.
func payloadType(payload []byte) string {
	var p struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(payload, &p); err == nil && p.Type != "" {
		return p.Type
	}
	return "message"
}

func copyHeaders(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
