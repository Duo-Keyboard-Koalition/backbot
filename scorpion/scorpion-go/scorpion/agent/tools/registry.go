package tools

import (
	"context"
	"fmt"
	"sync"

	"scorpion-go/internal/adk"
)

// Registry manages tool registration.
type Registry struct {
	mu    sync.RWMutex
	tools map[string]adk.Tool
}

// NewRegistry creates a new tool registry.
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]adk.Tool),
	}
}

// Register registers a tool.
func (r *Registry) Register(tool adk.Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if tool.Name() == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	r.tools[tool.Name()] = tool
	return nil
}

// Get retrieves a tool by name.
func (r *Registry) Get(name string) (adk.Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, ok := r.tools[name]
	return tool, ok
}

// List returns all registered tools.
func (r *Registry) List() []adk.Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]adk.Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// Execute executes a tool.
func (r *Registry) Execute(ctx context.Context, name string, args map[string]string) (string, error) {
	r.mu.RLock()
	tool, ok := r.tools[name]
	r.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("tool %q not found", name)
	}

	return tool.Run(ctx, args)
}

// Unregister removes a tool.
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tools, name)
}
