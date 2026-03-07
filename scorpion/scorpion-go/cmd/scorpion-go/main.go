package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Duo-Keyboard-Koalition/AuraFlow/scorpion-go/internal/adk"
)

func main() {
	workspace := envOr("SCORPION_GO_WORKSPACE", ".")
	base, err := filepath.Abs(workspace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve workspace: %v\n", err)
		os.Exit(1)
	}

	// Create message store for communication
	msgStore := adk.NewMessageStore()

	tools := adk.NewToolRegistry()

	// Core tools
	tools.Register(adk.TimeNowTool{})
	tools.Register(adk.ListDirTool{BaseDir: base})
	tools.Register(adk.ReadFileTool{BaseDir: base})

	// File operations
	tools.Register(adk.WriteFileTool{BaseDir: base})
	tools.Register(adk.EditFileTool{BaseDir: base})

	// Shell execution
	tools.Register(adk.ExecTool{
		Timeout:    60,
		WorkingDir: base,
	})

	// Web tools (require BRAVE_API_KEY for search)
	tools.Register(adk.WebSearchTool{
		APIKey:     os.Getenv("BRAVE_API_KEY"),
		MaxResults: 5,
	})
	tools.Register(adk.WebFetchTool{})

	// Communication
	tools.Register(adk.MessageTool{Store: msgStore})

	agent := adk.NewAgent(
		adk.RuleModel{},
		tools,
		"You are Scorpion-Go ADK. A capable AI assistant with file system, shell, and web access. "+
			"Prefer using tools to accomplish tasks. Always read files before editing. "+
			"Use web_search for current information, web_fetch for page content. "+
			"Use message to send updates to the user.",
		12,
	)

	fmt.Printf("🐈 scorpion-go adk ready (workspace=%s)\n", base)
	fmt.Printf("tools: %s\n", strings.Join(tools.Names(), ", "))
	fmt.Println("type 'exit' to quit")
	fmt.Println()
	fmt.Println("Quick commands:")
	fmt.Println("  /time              - Show current time")
	fmt.Println("  /ls [path]         - List directory")
	fmt.Println("  /cat <path>        - Read file")
	fmt.Println("  /write <path>      - Write to file (interactive)")
	fmt.Println("  /exec <command>    - Execute shell command")
	fmt.Println("  /search <query>    - Web search (requires BRAVE_API_KEY)")
	fmt.Println("  /fetch <url>       - Fetch web page")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	history := make([]adk.Message, 0)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "exit" || line == "quit" || line == "/exit" || line == "/quit" {
			fmt.Println("Goodbye!")
			break
		}

		// Handle slash commands
		userInput := line
		if strings.HasPrefix(line, "/") {
			userInput = handleSlashCommand(line, base)
			if userInput == "" {
				continue
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		reply, runs, runErr := agent.RunTurn(ctx, history, userInput)
		cancel()

		if runErr != nil {
			fmt.Printf("error: %v\n", runErr)
			continue
		}

		history = append(history, adk.Message{Role: "user", Content: userInput}, reply)

		// Print tool results
		for _, r := range runs {
			status := "ok"
			if r.IsError {
				status = "error"
			}
			fmt.Printf("  [tool:%s][%s]\n", r.Name, status)
		}

		// Print assistant response
		fmt.Printf("🐈 %s\n", reply.Content)

		// Print any pending messages
		msgs := msgStore.GetAndClear()
		for _, msg := range msgs {
			fmt.Printf("📬 %s\n", msg.Content)
		}
		fmt.Println()
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "read failed: %v\n", err)
		os.Exit(1)
	}
}

// handleSlashCommand processes slash commands and returns the equivalent natural language input
func handleSlashCommand(cmd, baseDir string) string {
	parts := strings.SplitN(cmd, " ", 2)
	command := parts[0]
	arg := ""
	if len(parts) > 1 {
		arg = strings.TrimSpace(parts[1])
	}

	switch command {
	case "/time":
		return "What time is it?"
	case "/ls":
		if arg == "" {
			arg = "."
		}
		return fmt.Sprintf("list_dir path=%s", arg)
	case "/cat":
		if arg == "" {
			fmt.Println("Usage: /cat <filepath>")
			return ""
		}
		return fmt.Sprintf("read_file path=%s", arg)
	case "/write":
		if arg == "" {
			fmt.Println("Usage: /write <filepath>")
			return ""
		}
		fmt.Print("Enter content (end with empty line): ")
		var content strings.Builder
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				break
			}
			content.WriteString(line)
			content.WriteString("\n")
		}
		return fmt.Sprintf("write_file path=%s content=%s", arg, content.String())
	case "/exec":
		if arg == "" {
			fmt.Println("Usage: /exec <command>")
			return ""
		}
		return fmt.Sprintf("exec command=%s", arg)
	case "/search":
		if arg == "" {
			fmt.Println("Usage: /search <query>")
			return ""
		}
		return fmt.Sprintf("web_search query=%s", arg)
	case "/fetch":
		if arg == "" {
			fmt.Println("Usage: /fetch <url>")
			return ""
		}
		return fmt.Sprintf("web_fetch url=%s", arg)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Available: /time, /ls, /cat, /write, /exec, /search, /fetch")
		return ""
	}
}

func envOr(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}
