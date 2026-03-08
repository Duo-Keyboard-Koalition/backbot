package agent

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"darci-go/darci/aip"
)

// Agent represents the DarCI Go agent
type Agent struct {
	server     *aip.AgentServer
	config     *Config
	ctx        context.Context
	cancel     context.CancelFunc
}

// Config holds agent configuration
type Config struct {
	AgentID     string
	Secret      string
	BridgeURL   string
	ListenAddr  string
	Capabilities []string
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv() *Config {
	config := &Config{
		AgentID:   getEnv("DARCI_AGENT_ID", "darci-go-001"),
		Secret:    getEnv("DARCI_AGENT_SECRET", ""),
		BridgeURL: getEnv("DARCI_BRIDGE_URL", "http://127.0.0.1:8080"),
		ListenAddr: getEnv("DARCI_LISTEN_ADDR", ":9090"),
		Capabilities: []string{
			"task-execution",
			"notebook",
			"file-ops",
			"shell",
		},
	}

	// Override capabilities if specified
	if caps := getEnv("DARCI_CAPABILITIES", ""); caps != "" {
		config.Capabilities = []string{caps}
	}

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// NewAgent creates a new DarCI agent
func NewAgent(config *Config) (*Agent, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Create agent server
	server, err := aip.NewAgentServer(aip.AgentServerConfig{
		BridgeURL:    config.BridgeURL,
		AgentID:      config.AgentID,
		Secret:       config.Secret,
		ListenAddr:   config.ListenAddr,
		Capabilities: config.Capabilities,
	})
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create agent server: %w", err)
	}

	return &Agent{
		server: server,
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// Start starts the agent
func (a *Agent) Start() error {
	log.Printf("[agent] starting DarCI Go agent: %s", a.config.AgentID)
	log.Printf("[agent] bridge URL: %s", a.config.BridgeURL)
	log.Printf("[agent] listen address: %s", a.config.ListenAddr)
	log.Printf("[agent] capabilities: %v", a.config.Capabilities)

	// Start HTTP server (handles /aip/handshake and /health)
	if err := a.server.Start(); err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	log.Printf("[agent] HTTP server started on %s", a.config.ListenAddr)

	// Wait a moment for HTTP server to be ready
	time.Sleep(100 * time.Millisecond)

	// Register with bridge and start heartbeat
	if err := a.server.RegisterAndStartHeartbeat(a.ctx); err != nil {
		log.Printf("[agent] warning: registration failed: %v", err)
		log.Printf("[agent] will retry registration in background")
		
		// Retry registration in background
		go a.retryRegistration()
	} else {
		log.Printf("[agent] successfully registered with bridge")
	}

	return nil
}

// retryRegistration retries registration with exponential backoff
func (a *Agent) retryRegistration() {
	backoff := 5 * time.Second
	maxBackoff := 5 * time.Minute

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-time.After(backoff):
			log.Printf("[agent] retrying registration...")
			
			// Create new context for registration attempt
			ctx, cancel := context.WithTimeout(a.ctx, 30*time.Second)
			err := a.server.RegisterAndStartHeartbeat(ctx)
			cancel()
			
			if err == nil {
				log.Printf("[agent] registration successful")
				return
			}
			
			log.Printf("[agent] registration retry failed: %v", err)
			
			// Exponential backoff
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}
}

// Run starts the agent and waits for shutdown signal
func (a *Agent) Run() error {
	if err := a.Start(); err != nil {
		return err
	}

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("[agent] received signal %v, shutting down...", sig)

	return a.Stop()
}

// Stop stops the agent
func (a *Agent) Stop() error {
	log.Printf("[agent] stopping agent...")

	// Cancel context
	a.cancel()

	// Stop server
	if err := a.server.Stop(); err != nil {
		log.Printf("[agent] error stopping server: %v", err)
	}

	log.Printf("[agent] agent stopped")
	return nil
}

// GetAgentID returns the agent ID
func (a *Agent) GetAgentID() string {
	return a.config.AgentID
}

// IsApproved returns true if agent is approved by bridge
func (a *Agent) IsApproved() bool {
	return a.server.GetClient().IsApproved()
}
