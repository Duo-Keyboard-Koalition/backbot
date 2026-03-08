package main

import (
	"fmt"
	"os"
)

const Version = "0.2.0"

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
		fmt.Printf("walkie-talkie-bridge version %s\n", Version)
	case "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Walkie-Talkie Bridge - Peer-to-peer agent communication")
	fmt.Printf("Version: %s\n\n", Version)
	fmt.Println("Usage:")
	fmt.Println("  bridge init    Initialize configuration interactively")
	fmt.Println("  bridge run     Start the bridge (default)")
	fmt.Println("  bridge version Show version information")
	fmt.Println("  bridge help    Show this help message")
	fmt.Println()
	fmt.Println("Configuration:")
	fmt.Println("  Config file: ~/.tailtalkie/config.json")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  bridge init              # Interactive setup")
	fmt.Println("  bridge run               # Start bridge with config")
	fmt.Println("  bridge version           # Show version")
	fmt.Println("  go run ./bridge init     # Initialize via go run")
	fmt.Println("  go run ./bridge run      # Run via go run")
}
