package tools

import (
	"context"
	"fmt"
)

// ManageTools provides management tools.
type ManageTools struct{}

// NewManageTools creates new manage tools.
func NewManageTools() *ManageTools {
	return &ManageTools{}
}

// ListTools lists available tools.
func (m *ManageTools) ListTools(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	return []string{
		"read_file",
		"write_file",
		"edit_file",
		"list_dir",
		"execute_shell",
		"fetch_url",
		"search_web",
		"send_message",
	}, nil
}

// GetToolInfo gets information about a tool.
func (m *ManageTools) GetToolInfo(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	toolName, ok := args["tool_name"].(string)
	if !ok {
		return nil, fmt.Errorf("tool_name is required")
	}

	return fmt.Sprintf("Tool info for: %s", toolName), nil
}
