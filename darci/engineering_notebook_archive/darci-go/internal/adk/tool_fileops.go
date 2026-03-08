package adk

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WriteFileTool writes content to a file.
type WriteFileTool struct {
	BaseDir   string
}

func (WriteFileTool) Name() string { return "write_file" }
func (WriteFileTool) Description() string {
	return "Write content to a file. Creates parent directories if needed."
}
func (t WriteFileTool) Run(ctx context.Context, input map[string]string) (string, error) {
	path := strings.TrimSpace(input["path"])
	content := input["content"]

	if path == "" {
		return "", fmt.Errorf("path is required")
	}

	full, err := safePath(t.BaseDir, path)
	if err != nil {
		return "", err
	}

	// Create parent directories
	dir := filepath.Dir(full)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directories: %w", err)
	}

	// Write file
	if err := os.WriteFile(full, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), path), nil
}

// EditFileTool edits a file with a diff.
type EditFileTool struct {
	BaseDir string
}

func (EditFileTool) Name() string { return "edit_file" }
func (EditFileTool) Description() string {
	return "Edit a file by replacing old_text with new_text. Returns diff."
}
func (t EditFileTool) Run(ctx context.Context, input map[string]string) (string, error) {
	path := strings.TrimSpace(input["path"])
	oldText := input["old_text"]
	newText := input["new_text"]

	if path == "" {
		return "", fmt.Errorf("path is required")
	}
	if oldText == "" {
		return "", fmt.Errorf("old_text is required")
	}

	full, err := safePath(t.BaseDir, path)
	if err != nil {
		return "", err
	}

	// Read current content
	content, err := os.ReadFile(full)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	current := string(content)
	if !strings.Contains(current, oldText) {
		return "", fmt.Errorf("old_text not found in file")
	}

	// Replace
	updated := strings.Replace(current, oldText, newText, 1)

	// Write back
	if err := os.WriteFile(full, []byte(updated), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// Generate simple diff
	diff := fmt.Sprintf("Edited %s:\n- %s\n+ %s", path, 
		truncateString(oldText, 100), 
		truncateString(newText, 100))

	return diff, nil
}

func truncateString(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", "\\n")
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}
