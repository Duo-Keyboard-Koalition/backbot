package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"darci-go/internal/adk"
	"darci-go/darci/agent"
)

// App represents the CLI application.
type App struct {
	Version string
	Name    string
}

// NewApp creates a new CLI application.
func NewApp() *App {
	return &App{
		Name:    "scorpion",
		Version: "0.1.0",
	}
}

// Run runs the CLI application.
func (a *App) Run(args []string) error {
	if len(args) < 1 {
		a.printUsage()
		return nil
	}

	command := args[0]
	switch command {
	case "agent":
		return a.runAgent(args[1:])
	case "gateway":
		return a.runGateway(args[1:])
	case "status":
		return a.runStatus(args[1:])
	case "onboard":
		return a.runOnboard(args[1:])
	case "version", "-v", "--version":
		fmt.Printf("%s version %s\n", a.Name, a.Version)
		return nil
	case "help", "-h", "--help":
		a.printUsage()
		return nil
	default:
		fmt.Printf("Unknown command: %s\n", command)
		a.printUsage()
		return fmt.Errorf("unknown command: %s", command)
	}
}

// runAgent runs the agent in interactive mode.
func (a *App) runAgent(args []string) error {
	// Parse flags
	_ = false // noMarkdown (unused for now)
	_ = false // logs (unused for now)
	message := ""

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--no-markdown":
			// noMarkdown = true
		case "--logs":
			// logs = true
		case "-m", "--message":
			if i+1 < len(args) {
				message = args[i+1]
				i++
			}
		}
	}

	// Create agent with rule-based model and tool registry
	model := adk.RuleModel{}
	tools := adk.NewToolRegistry()
	
	// Register built-in tools
	tools.Register(adk.TimeNowTool{})
	tools.Register(adk.ListDirTool{BaseDir: "."})
	tools.Register(adk.ReadFileTool{BaseDir: "."})
	tools.Register(adk.WriteFileTool{BaseDir: "."})
	tools.Register(adk.EditFileTool{BaseDir: "."})
	tools.Register(adk.ExecTool{})
	tools.Register(adk.WebSearchTool{})
	tools.Register(adk.WebFetchTool{})
	tools.Register(adk.MessageTool{})
	
	loop := agent.NewAdkAgentLoop(model, tools, "You are Scorpion, a helpful AI assistant.")
	if err := loop.Initialize(context.Background()); err != nil {
		return fmt.Errorf("failed to initialize agent: %w", err)
	}

	// Single message mode
	if message != "" {
		response, err := loop.Run(context.Background(), message)
		if err != nil {
			return err
		}
		fmt.Println(response.Content)
		return nil
	}

	// Interactive mode
	fmt.Println("Scorpion Agent - Interactive Mode")
	fmt.Println("Type 'exit', 'quit', or Ctrl+D to exit")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Check for exit commands
		if a.isExitCommand(input) {
			break
		}

		// Run agent
		response, err := loop.Run(context.Background(), input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Println(response.Content)
	}

	return nil
}

// runGateway runs the gateway for chat channels.
func (a *App) runGateway(args []string) error {
	fmt.Println("Starting Scorpion Gateway...")
	// In a full implementation, this would:
	// 1. Load configuration
	// 2. Initialize channels
	// 3. Start message bus
	// 4. Run all enabled channels
	// 5. Process messages

	return nil
}

// runStatus shows the status.
func (a *App) runStatus(args []string) error {
	fmt.Println("Scorpion Status")
	fmt.Println("Version:", a.Version)
	fmt.Println("Status: OK")
	return nil
}

// runOnboard initializes the configuration and workspace.
func (a *App) runOnboard(args []string) error {
	fmt.Println("Scorpion Onboarding")
	fmt.Println("Initializing configuration and workspace...")
	// In a full implementation, this would:
	// 1. Create ~/.darci directory
	// 2. Create config.json template
	// 3. Create workspace directory
	// 4. Create template files

	return nil
}

// isExitCommand checks if the input is an exit command.
func (a *App) isExitCommand(input string) bool {
	commands := []string{"exit", "quit", "/exit", "/quit", ":q"}
	input = strings.ToLower(strings.TrimSpace(input))
	for _, cmd := range commands {
		if input == cmd {
			return true
		}
	}
	return false
}

// printUsage prints the usage information.
func (a *App) printUsage() {
	fmt.Printf("Usage: %s <command> [options]\n\n", a.Name)
	fmt.Println("Commands:")
	fmt.Println("  agent     Chat with the agent")
	fmt.Println("  gateway   Start the gateway")
	fmt.Println("  status    Show status")
	fmt.Println("  onboard   Initialize config & workspace")
	fmt.Println("  version   Show version")
	fmt.Println("\nOptions:")
	fmt.Println("  -h, --help     Show help")
	fmt.Println("  -v, --version  Show version")
	fmt.Println("\nAgent options:")
	fmt.Println("  -m, --message  Single message mode")
	fmt.Println("  --no-markdown  Disable markdown formatting")
	fmt.Println("  --logs         Show runtime logs")
}
