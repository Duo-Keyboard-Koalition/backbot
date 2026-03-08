package main

import (
	"fmt"
	"os"
)

const Version = "0.3.0"

func main() {
	if len(os.Args) < 2 {
		// Default: run bridge
		runBridge()
		return
	}

	command := os.Args[1]

	switch command {
	case "init":
		if err := runInit(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "run":
		runBridge()
	case "version":
		fmt.Printf("taila2a bridge version %s\n", Version)
	case "aip":
		runAIPCommand(os.Args[2:])
	case "secrets":
		runSecretsCommand(os.Args[2:])
	case "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("TailA2A Bridge - Agent-to-Agent communication over Tailscale")
	fmt.Printf("Version: %s\n\n", Version)
	fmt.Println("Usage:")
	fmt.Println("  taila2a init              Initialize configuration interactively")
	fmt.Println("  taila2a run               Start the bridge (default)")
	fmt.Println("  taila2a version           Show version information")
	fmt.Println("  taila2a aip <subcommand>  Agent Identification Protocol commands")
	fmt.Println("  taila2a secrets <cmd>     Manage agent secrets")
	fmt.Println("  taila2a help              Show this help message")
	fmt.Println()
	fmt.Println("AIP Commands:")
	fmt.Println("  taila2a aip list          List all registered agents")
	fmt.Println("  taila2a aip pending       List pending agent registrations")
	fmt.Println("  taila2a aip approve <id>  Approve a pending registration")
	fmt.Println("  taila2a aip reject <id>   Reject a pending registration")
	fmt.Println("  taila2a aip remove <id>   Remove an agent from registry")
	fmt.Println()
	fmt.Println("Secrets Commands:")
	fmt.Println("  taila2a secrets generate <agent_id>  Generate new secret for agent")
	fmt.Println("  taila2a secrets list                 List all agents with secrets")
	fmt.Println("  taila2a secrets remove <agent_id>    Remove agent secret")
	fmt.Println()
	fmt.Println("Configuration:")
	fmt.Println("  Config file: ~/.tailtalkie/config.json")
	fmt.Println("  Registry:    ~/.tailtalkie/state/registry.json")
	fmt.Println("  Secrets:     ~/.tailtalkie/state/agent_secrets.json")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  taila2a init                        # Interactive setup")
	fmt.Println("  taila2a run                         # Start bridge with config")
	fmt.Println("  taila2a aip list                    # List registered agents")
	fmt.Println("  taila2a secrets generate agent-001  # Generate secret for agent")
	fmt.Println("  taila2a aip approve agent-001       # Approve agent registration")
}
