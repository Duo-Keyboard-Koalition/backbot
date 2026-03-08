package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/codejedi-ai/kaggle-for-tensors/tailscale-app/internal/models"
)

// runInit initializes the configuration file interactively
func runInit() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("=== Taila2a Setup ===")
	fmt.Println()

	// Get auth key
	fmt.Print("Tailscale Auth Key (from https://login.tailscale.com/admin/settings/keys): ")
	authKey, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read auth key: %w", err)
	}
	authKey = strings.TrimSpace(authKey)
	if authKey == "" {
		return fmt.Errorf("auth key is required")
	}

	// Get node name
	fmt.Print("Node Name (e.g., taila2a-alpha): ")
	name, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read name: %w", err)
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name = "taila2a-default"
		fmt.Printf("Using default: %s\n", name)
	}

	// Get local agent URL
	fmt.Print("Local Agent URL (e.g., http://127.0.0.1:9090/api): ")
	localAgentURL, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read local agent URL: %w", err)
	}
	localAgentURL = strings.TrimSpace(localAgentURL)
	if localAgentURL == "" {
		localAgentURL = "http://127.0.0.1:9090/api"
		fmt.Printf("Using default: %s\n", localAgentURL)
	}

	// Get inbound port
	fmt.Print("Inbound Port (default: 8001): ")
	inboundPortStr, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read inbound port: %w", err)
	}
	inboundPortStr = strings.TrimSpace(inboundPortStr)
	inboundPort := 8001
	if inboundPortStr != "" {
		if n := parseInt(inboundPortStr); n > 0 {
			inboundPort = n
		} else {
			fmt.Printf("Invalid port, using default: %d\n", inboundPort)
		}
	}

	// Get local listen address
	fmt.Print("Local Listen Address (default: 127.0.0.1:8080): ")
	localListen, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read local listen address: %w", err)
	}
	localListen = strings.TrimSpace(localListen)
	if localListen == "" {
		localListen = "127.0.0.1:8080"
		fmt.Printf("Using default: %s\n", localListen)
	}

	// Create config
	cfg := models.DefaultConfig()
	cfg.AuthKey = authKey
	cfg.Name = name
	cfg.LocalAgentURL = localAgentURL
	cfg.InboundPort = inboundPort
	cfg.LocalListen = localListen
	cfg.PeerInboundPort = inboundPort

	// Save config
	if err := models.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	fmt.Println("✓ Setup complete!")
	fmt.Println()
	fmt.Println("To start taila2a, run:")
	fmt.Println("  go run ./taila2a")
	fmt.Println()
	fmt.Println("To start with a different config, edit ~/.taila2a/config.json")

	return nil
}

func parseInt(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}
