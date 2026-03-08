package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// runInit initializes the configuration file interactively.
func runInit() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("=== Walkie-Talkie Bridge Setup ===")
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

	// Get bridge name
	fmt.Print("Bridge Name (e.g., bridge-alpha): ")
	bridgeName, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read bridge name: %w", err)
	}
	bridgeName = strings.TrimSpace(bridgeName)
	if bridgeName == "" {
		bridgeName = defaultBridgeName
		fmt.Printf("Using default: %s\n", bridgeName)
	}

	// Get local agent URL
	fmt.Print("Local Agent URL (e.g., http://127.0.0.1:9090/api): ")
	localAgentURL, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read local agent URL: %w", err)
	}
	localAgentURL = strings.TrimSpace(localAgentURL)
	if localAgentURL == "" {
		localAgentURL = defaultLocalAgent
		fmt.Printf("Using default: %s\n", localAgentURL)
	}

	// Get inbound port
	fmt.Print("Inbound Port (default: 8001): ")
	inboundPortStr, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read inbound port: %w", err)
	}
	inboundPortStr = strings.TrimSpace(inboundPortStr)
	inboundPort := defaultInboundPort
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
		localListen = defaultLocalListen
		fmt.Printf("Using default: %s\n", localListen)
	}

	// Create config
	cfg := DefaultConfig()
	cfg.AuthKey = authKey
	cfg.BridgeName = bridgeName
	cfg.LocalAgentURL = localAgentURL
	cfg.InboundPort = inboundPort
	cfg.LocalListen = localListen
	cfg.PeerInboundPort = inboundPort // Same as inbound by default

	// Save config
	if err := saveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	fmt.Println("✓ Setup complete!")
	fmt.Println()
	fmt.Println("To start the bridge, run:")
	fmt.Println("  go run ./bridge")
	fmt.Println()
	fmt.Println("To start with a different config, edit ~/.tailtalkie/config.json")

	return nil
}

func parseInt(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}
