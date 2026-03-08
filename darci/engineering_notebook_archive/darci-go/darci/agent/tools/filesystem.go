package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FilesystemTools provides file system related tools.
type FilesystemTools struct {
	workspace string
}

// NewFilesystemTools creates new filesystem tools.
func NewFilesystemTools(workspace string) *FilesystemTools {
	return &FilesystemTools{
		workspace: workspace,
	}
}

// ReadFile reads a file's contents.
func (f *FilesystemTools) ReadFile(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path is required")
	}

	// Security: ensure path is within workspace
	safePath, err := f.safePath(path)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(safePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}

// WriteFile writes content to a file.
func (f *FilesystemTools) WriteFile(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path is required")
	}

	content, ok := args["content"].(string)
	if !ok {
		return nil, fmt.Errorf("content is required")
	}

	safePath, err := f.safePath(path)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(safePath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(safePath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return "File written successfully", nil
}

// ListDir lists directory contents.
func (f *FilesystemTools) ListDir(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path is required")
	}

	safePath, err := f.safePath(path)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(safePath)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}

	var result []string
	for _, entry := range entries {
		result = append(result, entry.Name())
	}

	return result, nil
}

// EditFile edits a file by replacing old content with new content.
func (f *FilesystemTools) EditFile(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path is required")
	}

	oldStr, ok := args["old_string"].(string)
	if !ok {
		return nil, fmt.Errorf("old_string is required")
	}

	newStr, ok := args["new_string"].(string)
	if !ok {
		return nil, fmt.Errorf("new_string is required")
	}

	safePath, err := f.safePath(path)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(safePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	newContent := strings.Replace(string(content), oldStr, newStr, 1)
	if newContent == string(content) {
		return nil, fmt.Errorf("old_string not found in file")
	}

	if err := os.WriteFile(safePath, []byte(newContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return "File edited successfully", nil
}

// safePath ensures the path is within the workspace.
func (f *FilesystemTools) safePath(path string) (string, error) {
	// Clean the path
	cleanPath := filepath.Clean(path)

	// If it's already absolute, check if it's within workspace
	if filepath.IsAbs(cleanPath) {
		if !strings.HasPrefix(cleanPath, f.workspace) {
			return "", fmt.Errorf("path outside workspace: %s", path)
		}
		return cleanPath, nil
	}

	// Relative path - join with workspace
	return filepath.Join(f.workspace, cleanPath), nil
}
