package agent

import (
	"context"
	"sync"

	"darci-go/darci/bus"
	"darci-go/internal/adk"
)

// AdkAgentLoop manages the agent execution loop.
type AdkAgentLoop struct {
	mu      sync.RWMutex
	agent   *adk.Agent
	context *ContextBuilder
	memory  *MemoryStore
	skills  *SkillsLoader
	Bus     *bus.MessageBus
	Tools   *adk.ToolRegistry
}

// NewAdkAgentLoop creates a new agent loop.
func NewAdkAgentLoop(model adk.Model, tools *adk.ToolRegistry, systemPrompt string) *AdkAgentLoop {
	if tools == nil {
		tools = adk.NewToolRegistry()
	}
	return &AdkAgentLoop{
		agent:   adk.NewAgent(model, tools, systemPrompt, 8),
		context: NewContextBuilder(),
		memory:  NewMemoryStore(),
		skills:  NewSkillsLoader(),
		Bus:     bus.NewMessageBus(100),
		Tools:   tools,
	}
}

// Initialize initializes the agent loop.
func (l *AdkAgentLoop) Initialize(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if err := l.skills.Load(); err != nil {
		return err
	}

	// Skills are already tools, no need to re-register
	// The tools are registered in the tool registry passed to NewAdkAgentLoop

	return nil
}

// Run executes the agent loop with the given input.
func (l *AdkAgentLoop) Run(ctx context.Context, input string) (adk.Message, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Add user input to context
	l.context.AddUserMessage(input)

	// Run the agent
	msg, toolResults, err := l.agent.RunTurn(ctx, l.context.Build(), input)
	if err != nil {
		return adk.Message{}, err
	}

	// Add assistant response to context
	l.context.AddAssistantMessage(msg.Content)

	// Update memory if needed
	if len(toolResults) == 0 {
		l.memory.AddShortTerm(input, msg.Content)
	}

	return msg, nil
}

// GetContext returns the current context.
func (l *AdkAgentLoop) GetContext() []adk.Message {
	return l.context.Build()
}

// ClearContext clears the context.
func (l *AdkAgentLoop) ClearContext() {
	l.context.Clear()
}

// GetAgent returns the underlying agent.
func (l *AdkAgentLoop) GetAgent() *adk.Agent {
	return l.agent
}
