package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"darci-go/darci/agent"
	"darci-go/darci/config"
	"darci-go/darci/agent/tools"
	"darci-go/darci/state"
	"darci-go/internal/adk"
)

func main() {
	// Load configuration
	darciConfig := config.DefaultDarciConfig()

	// Initialize task store
	store, err := state.NewTaskStore(darciConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize task store: %v\n", err)
		os.Exit(1)
	}

	// Create tool registry and agent loop with RuleModel (rule-based fallback)
	toolRegistry := adk.NewToolRegistry()
	model := adk.RuleModel{}
	agentLoop := agent.NewAdkAgentLoop(model, toolRegistry, "You are DarCI, a DARCI project manager for AI agents.")

	// Register DarCI tools
	registry := tools.RegisterDarciTools(agentLoop, darciConfig, store)
	_ = registry // Keep registry for potential cleanup

	fmt.Println("DarCI ready — DARCI project manager for AI agents")
	fmt.Printf("State: %s\n", darciConfig.StateDir)
	fmt.Printf("Bridge: %s\n", darciConfig.BridgeLocalURL)
	fmt.Println("Type your command. Ctrl-C to exit.\n")

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("darci> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("\nDarCI shutting down.")
			break
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println("\nDarCI shutting down.")
			break
		}

		ctx := context.Background()
		response, err := agentLoop.Run(ctx, input)
		if err != nil {
			fmt.Printf("Error: %v\n\n", err)
			continue
		}

		fmt.Printf("\n%s\n\n", response.Content)
	}
}

func envOr(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
