package main

import (
	"fmt"
	"log"
	"os"

	"darci-go/darci/agent"
)

const Version = "1.0.0"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version":
			fmt.Printf("darci-go version %s\n", Version)
			return
		case "help":
			printUsage()
			return
		}
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("[darci] starting DarCI Go Agent v%s", Version)

	// Load configuration from environment
	config := agent.LoadConfigFromEnv()

	// Validate required config
	if config.Secret == "" {
		log.Fatal("[darci] error: DARCI_AGENT_SECRET environment variable is required")
	}

	// Create and run agent
	agt, err := agent.NewAgent(config)
	if err != nil {
		log.Fatalf("[darci] failed to create agent: %v", err)
	}

	if err := agt.Run(); err != nil {
		log.Fatalf("[darci] agent error: %v", err)
	}
}

func printUsage() {
	fmt.Println("DarCI Go Agent - Intelligent agent with AIP protocol")
	fmt.Printf("Version: %s\n\n", Version)
	fmt.Println("Usage:")
	fmt.Println("  darci [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  darci version    Show version information")
	fmt.Println("  darci help       Show this help message")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  DARCI_AGENT_ID      Agent identifier (default: darci-go-001)")
	fmt.Println("  DARCI_AGENT_SECRET  Shared secret for HMAC authentication (required)")
	fmt.Println("  DARCI_BRIDGE_URL    Bridge URL (default: http://127.0.0.1:8080)")
	fmt.Println("  DARCI_LISTEN_ADDR   Listen address (default: :9090)")
	fmt.Println("  DARCI_CAPABILITIES  Comma-separated capabilities")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Run with defaults")
	fmt.Println("  export DARCI_AGENT_SECRET=my-secret-key")
	fmt.Println("  darci")
	fmt.Println()
	fmt.Println("  # Run with custom agent ID")
	fmt.Println("  export DARCI_AGENT_ID=my-agent")
	fmt.Println("  export DARCI_AGENT_SECRET=my-secret-key")
	fmt.Println("  darci")
	fmt.Println()
	fmt.Println("  # Run with custom bridge")
	fmt.Println("  export DARCI_BRIDGE_URL=http://bridge-alpha:8001")
	fmt.Println("  export DARCI_AGENT_SECRET=my-secret-key")
	fmt.Println("  darci")
}
