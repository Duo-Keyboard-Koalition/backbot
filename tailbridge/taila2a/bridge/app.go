package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"tailscale.com/tsnet"

	"github.com/codejedi-ai/kaggle-for-tensors/tailscale-app/buffer"
)

var (
	discoverySvc *DiscoveryService
	registry     *AgentRegistry
	aipHandlers  *AIPHandlers
	handshakeSvc *HandshakeService
	bufferSvc    *buffer.BufferService
)

func runBridge() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	srv := &tsnet.Server{
		Hostname: cfg.BridgeName,
		Dir:      cfg.StateDir,
		AuthKey:  cfg.AuthKey,
	}
	defer srv.Close()

	// Initialize agent registry (AIP - Agent Identification Protocol)
	// This replaces auto-discovery with explicit agent registration
	registry, err = NewAgentRegistry(cfg.StateDir, cfg.BridgeName)
	if err != nil {
		log.Fatalf("Failed to create agent registry: %v", err)
	}
	log.Printf("[aip] agent registry initialized (bridge=%s)", cfg.BridgeName)

	// Initialize handshake service for challenge-response agent verification
	// Load agent secrets from config (in production, use secure secret management)
	agentSecrets := loadAgentSecrets(cfg.StateDir)
	handshakeSvc = NewHandshakeService(cfg.BridgeName, agentSecrets)
	handshakeSvc.StartCleanup(1*time.Minute, make(chan struct{}))
	log.Printf("[handshake] service initialized with %d agent secrets", len(agentSecrets))

	// Initialize AIP HTTP handlers
	aipHandlers = NewAIPHandlers(registry)
	// Add handshake handler for agent-side endpoint
	handshakeHandler := NewHandshakeHandler(handshakeSvc, "", "") // Agent ID/secret set per-agent

	tailnetClient := tsHTTPClient(srv, 20*time.Second)
	localClient := &http.Client{Timeout: 20 * time.Second}

	// Initialize buffer service
	bufferDir := filepath.Join(cfg.StateDir, "buffer")
	bufferConfig := &buffer.BufferServiceConfig{
		DataDir:         bufferDir,
		RetryConfig:     buffer.DefaultRetryConfig(),
		ProcessInterval: 5 * time.Second,
		HTTPTimeout:     20 * time.Second,
		PeerInboundPort: cfg.PeerInboundPort,
	}

	deliverFunc := makeBufferDeliveryFunc(cfg.BridgeName, cfg.PeerInboundPort, tailnetClient)
	bufferSvc, err = buffer.NewBufferService(bufferConfig, deliverFunc)
	if err != nil {
		log.Fatalf("Failed to create buffer service: %v", err)
	}

	// Start buffer background processor
	ctx := context.Background()
	bufferSvc.Start(ctx)

	go runOutboundServer(cfg.LocalListen, cfg.BridgeName, cfg.PeerInboundPort, tailnetClient)
	go logSelfTailscaleIPs(srv)

	ln, err := srv.Listen("tcp", fmt.Sprintf(":%d", cfg.InboundPort))
	if err != nil {
		log.Fatalf("bridge tailnet listen failed: %v", err)
	}

	mux := http.NewServeMux()

	// Original endpoints
	mux.HandleFunc("/inbound", makeInboundHandler(cfg.BridgeName, cfg.LocalAgentURL, localClient))
	mux.HandleFunc("/agents", makeAgentsHandler())

	// Buffer service endpoints
	mux.HandleFunc("/buffer/stats", makeBufferStatsHandler())
	mux.HandleFunc("/buffer/messages", makeBufferMessagesHandler())
	mux.HandleFunc("/buffer/retry", makeBufferRetryHandler())
	mux.HandleFunc("/buffer/clear", makeBufferClearHandler())

	// AIP (Agent Identification Protocol) endpoints
	// These replace auto-discovery with explicit registration
	mux.HandleFunc("/aip/register", aipHandlers.HandleRegister)
	mux.HandleFunc("/aip/heartbeat", aipHandlers.HandleHeartbeat)
	mux.HandleFunc("/aip/pair", aipHandlers.HandlePair)
	mux.HandleFunc("/aip/agents", aipHandlers.HandleListAgents)
	mux.HandleFunc("/aip/approve/", aipHandlers.HandleApproveAgent)
	mux.HandleFunc("/aip/reject/", aipHandlers.HandleRejectAgent)

	// tsA2A Handshake endpoints
	// Challenge-response verification for agent identification
	mux.HandleFunc("/aip/handshake", handshakeHandler.HandleHandshake)
	mux.HandleFunc("/aip/handshake-probe", handshakeHandler.HandleHandshakeProbe)

	httpSrv := &http.Server{
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("bridge node %q listening on tailnet :%d (state=%s)", cfg.BridgeName, cfg.InboundPort, cfg.StateDir)
	log.Printf("[aip] Agent Identification Protocol enabled - register agents at /aip/register")
	log.Printf("[aip] list agents: GET /aip/agents, approve: POST /aip/approve/{agent_id}")
	log.Printf("[handshake] tsA2A handshake enabled at /aip/handshake")
	log.Printf("[discovery] legacy mode: passive only - no network port scanning")
	log.Printf("buffer service initialized - stats available at /buffer/stats")
	if err := httpSrv.Serve(ln); err != nil && err != http.ErrServerClosed {
		log.Fatalf("bridge serve failed: %v", err)
	}
}
