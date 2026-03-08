package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"tailscale.com/tsnet"
)

// AgentConfig holds the agent configuration
type AgentConfig struct {
	Name         string   `json:"name"`
	Capabilities []string `json:"capabilities"`
	Tailnet      string   `json:"tailnet"`
	Port         int      `json:"port"`
	InboundPort  int      `json:"inbound_port"`
}

// Message represents an A2A message
type Message struct {
	ID            string      `json:"id"`
	Type          string      `json:"type"`
	Source        string      `json:"source"`
	Dest          string      `json:"dest"`
	Topic         string      `json:"topic"`
	Timestamp     time.Time   `json:"timestamp"`
	Body          MessageBody `json:"body"`
	CorrelationID string      `json:"correlation_id,omitempty"`
	ReplyTo       string      `json:"reply_to,omitempty"`
}

// MessageBody contains the message payload
type MessageBody struct {
	Action      string          `json:"action"`
	ContentType string          `json:"content_type"`
	Payload     json.RawMessage `json:"payload"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// PhoneBookEntry represents an agent in the phone book
type PhoneBookEntry struct {
	Name         string   `json:"name"`
	NodeID       string   `json:"node_id"`
	Tailnet      string   `json:"tailnet"`
	Capabilities []string `json:"capabilities"`
	Online       bool     `json:"online"`
	IP           string   `json:"ip,omitempty"`
}

// TestAgent is a dummy agent for testing
type TestAgent struct {
	config      *AgentConfig
	tsServer    *tsnet.Server
	messages    []Message
	mu          sync.RWMutex
	startTime   time.Time
	phoneBook   []PhoneBookEntry
}

func main() {
	// Parse command line flags
	initOnly := flag.Bool("init", false, "Initialize configuration only")
	flag.Parse()

	// Load configuration from environment
	config := loadConfigFromEnv()

	if *initOnly {
		fmt.Println("Test Agent Configuration:")
		fmt.Printf("  Name: %s\n", config.Name)
		fmt.Printf("  Capabilities: %s\n", strings.Join(config.Capabilities, ", "))
		fmt.Printf("  Port: %d\n", config.Port)
		fmt.Printf("  Inbound Port: %d\n", config.InboundPort)
		fmt.Println("\nConfiguration loaded from environment variables.")
		fmt.Println("Set the following environment variables:")
		fmt.Println("  AGENT_NAME - Agent name")
		fmt.Println("  AGENT_CAPABILITIES - Comma-separated capabilities")
		fmt.Println("  TS_AUTHKEY - Tailscale auth key")
		fmt.Println("  TS_HOSTNAME - Tailscale hostname")
		return
	}

	// Create and start agent
	agent := &TestAgent{
		config:    config,
		startTime: time.Now(),
		messages:  make([]Message, 0),
		phoneBook: make([]PhoneBookEntry, 0),
	}

	if err := agent.Start(); err != nil {
		log.Fatalf("Failed to start agent: %v", err)
	}

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	agent.Stop()
}

func loadConfigFromEnv() *AgentConfig {
	name := os.Getenv("AGENT_NAME")
	if name == "" {
		name = "test-agent-" + uuid.New().String()[:8]
	}

	capabilitiesStr := os.Getenv("AGENT_CAPABILITIES")
	capabilities := []string{"chat", "file_send", "file_receive"}
	if capabilitiesStr != "" {
		capabilities = strings.Split(capabilitiesStr, ",")
		for i := range capabilities {
			capabilities[i] = strings.TrimSpace(capabilities[i])
		}
	}

	port := 8080
	if p := os.Getenv("AGENT_PORT"); p != "" {
		fmt.Sscanf(p, "%d", &port)
	}

	inboundPort := 8001
	if p := os.Getenv("AGENT_INBOUND_PORT"); p != "" {
		fmt.Sscanf(p, "%d", &inboundPort)
	}

	return &AgentConfig{
		Name:         name,
		Capabilities: capabilities,
		Port:         port,
		InboundPort:  inboundPort,
	}
}

// Start starts the test agent
func (a *TestAgent) Start() error {
	log.Printf("Starting agent: %s", a.config.Name)

	// Initialize Tailscale server
	a.tsServer = &tsnet.Server{
		Hostname: a.config.Name,
		AuthKey:  os.Getenv("TS_AUTHKEY"),
		Dir:      "/var/lib/tailscale/" + a.config.Name,
	}

	// Get listener for HTTP server
	ln, err := a.tsServer.Listen("tcp", fmt.Sprintf(":%d", a.config.Port))
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	// Start HTTP server
	mux := http.NewServeMux()
	a.setupRoutes(mux)

	server := &http.Server{
		Handler: mux,
	}

	go func() {
		log.Printf("HTTP server listening on :%d", a.config.Port)
		if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Update phone book periodically
	go a.updatePhoneBook()

	log.Printf("Agent %s started successfully", a.config.Name)
	log.Printf("Capabilities: %s", strings.Join(a.config.Capabilities, ", "))

	return nil
}

// Stop stops the test agent
func (a *TestAgent) Stop() {
	if a.tsServer != nil {
		a.tsServer.Close()
	}
	log.Println("Agent stopped")
}

// setupRoutes sets up HTTP routes
func (a *TestAgent) setupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", a.handleHealth)
	mux.HandleFunc("/phonebook", a.handlePhoneBook)
	mux.HandleFunc("/a2a/inbound", a.handleInboundMessage)
	mux.HandleFunc("/a2a/send", a.handleSendMessage)
	mux.HandleFunc("/agents", a.handleListAgents)
	mux.HandleFunc("/messages", a.handleListMessages)
	mux.HandleFunc("/status", a.handleStatus)
}

// handleHealth handles health check requests
func (a *TestAgent) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "healthy",
		"agent":  a.config.Name,
		"uptime": time.Since(a.startTime).String(),
	})
}

// handlePhoneBook handles phone book requests
func (a *TestAgent) handlePhoneBook(w http.ResponseWriter, r *http.Request) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agents": a.phoneBook,
		"count":  len(a.phoneBook),
	})
}

// handleInboundMessage handles incoming A2A messages
func (a *TestAgent) handleInboundMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var msg Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "Invalid message format", http.StatusBadRequest)
		return
	}

	a.mu.Lock()
	a.messages = append(a.messages, msg)
	a.mu.Unlock()

	log.Printf("Received message from %s: %s", msg.Source, msg.Type)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "received",
		"id":     msg.ID,
	})
}

// handleSendMessage handles outgoing A2A messages
func (a *TestAgent) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Dest    string      `json:"dest"`
		Message MessageBody `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Create message
	msg := Message{
		ID:        uuid.New().String(),
		Type:      "request",
		Source:    a.config.Name,
		Dest:      req.Dest,
		Timestamp: time.Now(),
		Body:      req.Message,
	}

	log.Printf("Sending message to %s: %s", req.Dest, msg.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "sent",
		"id":     msg.ID,
	})
}

// handleListAgents handles agent listing requests
func (a *TestAgent) handleListAgents(w http.ResponseWriter, r *http.Request) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Filter by capability if requested
	capFilter := r.URL.Query().Get("capability")
	filtered := make([]PhoneBookEntry, 0)

	for _, entry := range a.phoneBook {
		if capFilter != "" {
			hasCap := false
			for _, cap := range entry.Capabilities {
				if cap == capFilter {
					hasCap = true
					break
				}
			}
			if !hasCap {
				continue
			}
		}
		filtered = append(filtered, entry)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agents": filtered,
		"count":  len(filtered),
	})
}

// handleListMessages handles message listing requests
func (a *TestAgent) handleListMessages(w http.ResponseWriter, r *http.Request) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}

	messages := a.messages
	if len(messages) > limit {
		messages = messages[len(messages)-limit:]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"messages": messages,
		"count":    len(messages),
	})
}

// handleStatus handles status requests
func (a *TestAgent) handleStatus(w http.ResponseWriter, r *http.Request) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"name":         a.config.Name,
		"capabilities": a.config.Capabilities,
		"uptime":       time.Since(a.startTime).String(),
		"messages":     len(a.messages),
		"phonebook":    len(a.phoneBook),
		"status":       "online",
	})
}

// updatePhoneBook periodically updates the phone book
func (a *TestAgent) updatePhoneBook() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// In a real implementation, this would scan the tailnet
		// For testing, we'll just maintain our own entry
		a.mu.Lock()
		
		found := false
		for i, entry := range a.phoneBook {
			if entry.Name == a.config.Name {
				a.phoneBook[i].Online = true
				a.phoneBook[i].IP = "127.0.0.1"
				found = true
				break
			}
		}

		if !found {
			a.phoneBook = append(a.phoneBook, PhoneBookEntry{
				Name:         a.config.Name,
				NodeID:       uuid.New().String()[:16],
				Tailnet:      "test.tailnet",
				Capabilities: a.config.Capabilities,
				Online:       true,
				IP:           "127.0.0.1",
			})
		}

		a.mu.Unlock()
		log.Printf("Phone book updated: %d agents", len(a.phoneBook))
	}
}
