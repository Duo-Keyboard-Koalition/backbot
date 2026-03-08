package tools

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// ShellTools provides shell command execution tools.
type ShellTools struct {
	workspace   string
	timeout     time.Duration
	pathAppend  string
	restrictToWorkspace bool
}

// NewShellTools creates new shell tools.
func NewShellTools(workspace string, restrictToWorkspace bool) *ShellTools {
	return &ShellTools{
		workspace:           workspace,
		timeout:             30 * time.Second,
		restrictToWorkspace: restrictToWorkspace,
	}
}

// ExecuteShell runs a shell command.
func (s *ShellTools) ExecuteShell(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	command, ok := args["command"].(string)
	if !ok {
		return nil, fmt.Errorf("command is required")
	}

	// Security check for workspace restriction
	if s.restrictToWorkspace {
		if err := s.checkCommandSafety(command); err != nil {
			return nil, err
		}
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Determine shell based on OS
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}

	// Set working directory
	cmd.Dir = s.workspace

	// Append to PATH if configured
	if s.pathAppend != "" {
		env := cmd.Environ()
		env = append(env, fmt.Sprintf("PATH=%s:%s", s.pathAppend, getEnv(env, "PATH")))
		cmd.Env = env
	}

	// Execute and capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("command timed out after %v", s.timeout)
		}
		return nil, fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// checkCommandSafety checks if a command is safe to run.
func (s *ShellTools) checkCommandSafety(command string) error {
	// Block dangerous commands
	dangerous := []string{
		"rm -rf /",
		"rm -rf /*",
		"mkfs",
		"dd if=/dev/zero",
		":(){:|:&};:",
		"chmod -R 777 /",
		"chown -R",
	}

	for _, dangerousCmd := range dangerous {
		if strings.Contains(command, dangerousCmd) {
			return fmt.Errorf("dangerous command blocked: %s", dangerousCmd)
		}
	}

	return nil
}

// getEnv gets an environment variable from env slice.
func getEnv(env []string, key string) string {
	prefix := key + "="
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			return e[len(prefix):]
		}
	}
	return ""
}
