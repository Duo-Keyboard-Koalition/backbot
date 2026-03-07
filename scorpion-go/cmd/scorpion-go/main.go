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

	tools := adk.NewToolRegistry()
	tools.Register(adk.TimeNowTool{})
	tools.Register(adk.ListDirTool{BaseDir: base})
	tools.Register(adk.ReadFileTool{BaseDir: base})

	agent := adk.NewAgent(
		adk.RuleModel{},
		tools,
		"You are Scorpion-Go ADK. Prefer local self-contained tools first.",
		8,
	)

	fmt.Printf("scorpion-go adk ready (workspace=%s)\n", base)
	fmt.Printf("tools: %s\n", strings.Join(tools.Names(), ", "))
	fmt.Println("type 'exit' to quit")

	scanner := bufio.NewScanner(os.Stdin)
	history := make([]adk.Message, 0)
	for {
		fmt.Print("\n> ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "exit" || line == "quit" {
			break
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		reply, runs, runErr := agent.RunTurn(ctx, history, line)
		cancel()
		if runErr != nil {
			fmt.Printf("error: %v\n", runErr)
			continue
		}

		history = append(history, adk.Message{Role: "user", Content: line}, reply)
		for _, r := range runs {
			status := "ok"
			if r.IsError {
				status = "error"
			}
			fmt.Printf("[tool:%s][%s]\n", r.Name, status)
		}
		fmt.Printf("assistant: %s\n", reply.Content)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "read failed: %v\n", err)
		os.Exit(1)
	}
}

func envOr(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}
