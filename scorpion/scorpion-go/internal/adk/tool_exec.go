package adk

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// ExecTool executes shell commands.
type ExecTool struct {
	Timeout     int
	WorkingDir  string
	DenyPatterns []string
}

func (ExecTool) Name() string { return "exec" }
func (ExecTool) Description() string {
	return "Execute a shell command and return its output. Use with caution."
}
func (t ExecTool) Run(ctx context.Context, input map[string]string) (string, error) {
	cmdStr := strings.TrimSpace(input["command"])
	if cmdStr == "" {
		return "", fmt.Errorf("command is required")
	}

	// Check for dangerous commands
	dangerous := []string{"rm -rf", "del /f", "format", "mkfs", "diskpart", "shutdown", "reboot"}
	for _, d := range dangerous {
		if strings.Contains(strings.ToLower(cmdStr), d) {
			return "", fmt.Errorf("dangerous command blocked: %s", d)
		}
	}

	workingDir := t.WorkingDir
	if input["working_dir"] != "" {
		workingDir = input["working_dir"]
	}

	timeout := time.Duration(t.Timeout) * time.Second
	if timeout == 0 {
		timeout = 60 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", cmdStr)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", cmdStr)
	}

	if workingDir != "" {
		cmd.Dir = workingDir
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("command timed out after %v", timeout)
		}
		return string(output), fmt.Errorf("command failed: %w", err)
	}

	return string(output), nil
}
