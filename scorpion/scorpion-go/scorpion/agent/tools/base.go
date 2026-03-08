package tools

import (
	"context"
	"fmt"

	"scorpion-go/internal/adk"
)

// BaseTool provides common functionality for all tools.
type BaseTool struct {
	Name        string
	Description string
	Parameters  map[string]interface{}
}

// ToolDefinition defines a tool's metadata.
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ToolExecutor executes a tool.
type ToolExecutor interface {
	Execute(ctx context.Context, args map[string]interface{}) (interface{}, error)
}

// ToolFunc is a function-based tool executor.
type ToolFunc func(ctx context.Context, args map[string]interface{}) (interface{}, error)

// Execute implements ToolExecutor for ToolFunc.
func (f ToolFunc) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	return f(ctx, args)
}

// ToTool converts a ToolExecutor to an adk.Tool.
func ToTool(name, description string, executor ToolExecutor) adk.Tool {
	return &toolAdapter{
		name:        name,
		description: description,
		executor:    executor,
	}
}

// ToToolFunc converts a function to an adk.Tool.
func ToToolFunc(name, description string, handler ToolFunc) adk.Tool {
	return &toolAdapter{
		name:        name,
		description: description,
		executor:    handler,
	}
}

// toolAdapter adapts ToolExecutor to adk.Tool interface.
type toolAdapter struct {
	name        string
	description string
	executor    ToolExecutor
}

func (t *toolAdapter) Name() string { return t.name }
func (t *toolAdapter) Description() string { return t.description }
func (t *toolAdapter) Run(ctx context.Context, input map[string]string) (string, error) {
	// Convert map[string]string to map[string]interface{}
	args := make(map[string]interface{}, len(input))
	for k, v := range input {
		args[k] = v
	}
	result, err := t.executor.Execute(ctx, args)
	if err != nil {
		return "", err
	}
	// Convert result to string
	if str, ok := result.(string); ok {
		return str, nil
	}
	return fmt.Sprintf("%v", result), nil
}
