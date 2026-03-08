package agent

import (
	"context"
	"sync"

	"darci-go/internal/adk"
)

// Subagent represents a subagent instance.
type Subagent struct {
	mu      sync.RWMutex
	id      string
	agent   *adk.Agent
	enabled bool
}

// SubagentConfig holds subagent configuration.
type SubagentConfig struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Model       adk.Model   `json:"-"`
	Tools       *adk.ToolRegistry `json:"-"`
	Enabled     bool        `json:"enabled"`
}

// SubagentManager manages subagents.
type SubagentManager struct {
	mu        sync.RWMutex
	subagents map[string]*Subagent
}

// NewSubagentManager creates a new subagent manager.
func NewSubagentManager() *SubagentManager {
	return &SubagentManager{
		subagents: make(map[string]*Subagent),
	}
}

// Create creates a new subagent.
func (m *SubagentManager) Create(ctx context.Context, config *SubagentConfig) (*Subagent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if config.ID == "" {
		return nil, ErrSubagentIDRequired
	}

	tools := config.Tools
	if tools == nil {
		tools = adk.NewToolRegistry()
	}

	subagent := &Subagent{
		id:      config.ID,
		agent:   adk.NewAgent(config.Model, tools, config.Description, 8),
		enabled: config.Enabled,
	}

	m.subagents[config.ID] = subagent
	return subagent, nil
}

// Get retrieves a subagent by ID.
func (m *SubagentManager) Get(id string) (*Subagent, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	subagent, ok := m.subagents[id]
	return subagent, ok
}

// List returns all subagents.
func (m *SubagentManager) List() []*Subagent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	subagents := make([]*Subagent, 0, len(m.subagents))
	for _, subagent := range m.subagents {
		subagents = append(subagents, subagent)
	}
	return subagents
}

// Remove removes a subagent.
func (m *SubagentManager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.subagents, id)
}

// Execute executes a subagent with the given input.
func (m *SubagentManager) Execute(ctx context.Context, id, input string) (adk.Message, error) {
	m.mu.RLock()
	subagent, ok := m.subagents[id]
	m.mu.RUnlock()

	if !ok {
		return adk.Message{}, ErrSubagentNotFound
	}

	if !subagent.enabled {
		return adk.Message{}, ErrSubagentDisabled
	}

	history := []adk.Message{}
	msg, _, err := subagent.agent.RunTurn(ctx, history, input)
	return msg, err
}

// Enable enables a subagent.
func (m *SubagentManager) Enable(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	subagent, ok := m.subagents[id]
	if !ok {
		return ErrSubagentNotFound
	}

	subagent.enabled = true
	return nil
}

// Disable disables a subagent.
func (m *SubagentManager) Disable(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	subagent, ok := m.subagents[id]
	if !ok {
		return ErrSubagentNotFound
	}

	subagent.enabled = false
	return nil
}

// Subagent errors
var (
	ErrSubagentIDRequired = &subagentError{"subagent ID is required"}
	ErrSubagentNotFound   = &subagentError{"subagent not found"}
	ErrSubagentDisabled   = &subagentError{"subagent is disabled"}
)

type subagentError struct {
	message string
}

func (e *subagentError) Error() string {
	return e.message
}

// Subagent methods

// Run executes the subagent with the given input.
func (s *Subagent) Run(ctx context.Context, input string) (adk.Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.enabled {
		return adk.Message{}, ErrSubagentDisabled
	}

	history := []adk.Message{}
	msg, _, err := s.agent.RunTurn(ctx, history, input)
	return msg, err
}

// IsEnabled returns whether the subagent is enabled.
func (s *Subagent) IsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.enabled
}

// GetID returns the subagent ID.
func (s *Subagent) GetID() string {
	return s.id
}
