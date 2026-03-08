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

var discoverySvc *DiscoveryService

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

	// Initialize discovery service
	discoverySvc, err = NewDiscoveryService(srv)
	if err != nil {
		log.Fatalf("Failed to create discovery service: %v", err)
	}
	// Start discovery with 30-second interval
	discoverySvc.Start(30 * time.Second)

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
	mux.HandleFunc("/inbound", makeInboundHandler(cfg.BridgeName, cfg.LocalAgentURL, localClient))
	mux.HandleFunc("/agents", makeAgentsHandler())
	mux.HandleFunc("/buffer/stats", makeBufferStatsHandler())
	mux.HandleFunc("/buffer/messages", makeBufferMessagesHandler())
	mux.HandleFunc("/buffer/retry", makeBufferRetryHandler())
	mux.HandleFunc("/buffer/clear", makeBufferClearHandler())

	httpSrv := &http.Server{
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("bridge node %q listening on tailnet :%d (state=%s)", cfg.BridgeName, cfg.InboundPort, cfg.StateDir)
	log.Printf("discovery service started - agents available at /agents endpoint")
	log.Printf("buffer service initialized - stats available at /buffer/stats")
	if err := httpSrv.Serve(ln); err != nil && err != http.ErrServerClosed {
		log.Fatalf("bridge serve failed: %v", err)
	}
}
