package agent

import (
	"sync"

	"scorpion-go/internal/adk"
)

// ContextBuilder builds and manages conversation context.
type ContextBuilder struct {
	mu       sync.RWMutex
	messages []adk.Message
	system   string
}

// NewContextBuilder creates a new context builder.
func NewContextBuilder() *ContextBuilder {
	return &ContextBuilder{
		messages: make([]adk.Message, 0),
	}
}

// WithSystemPrompt sets the system prompt.
func (cb *ContextBuilder) WithSystemPrompt(prompt string) *ContextBuilder {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.system = prompt
	return cb
}

// AddMessage adds a message to the context.
func (cb *ContextBuilder) AddMessage(role, content string) *ContextBuilder {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.messages = append(cb.messages, adk.Message{
		Role:    role,
		Content: content,
	})
	return cb
}

// AddUserMessage adds a user message.
func (cb *ContextBuilder) AddUserMessage(content string) *ContextBuilder {
	return cb.AddMessage("user", content)
}

// AddAssistantMessage adds an assistant message.
func (cb *ContextBuilder) AddAssistantMessage(content string) *ContextBuilder {
	return cb.AddMessage("assistant", content)
}

// Build returns the built context.
func (cb *ContextBuilder) Build() []adk.Message {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	result := make([]adk.Message, 0, len(cb.messages)+1)
	if cb.system != "" {
		result = append(result, adk.Message{
			Role:    "system",
			Content: cb.system,
		})
	}
	return append(result, cb.messages...)
}

// Clear clears the context.
func (cb *ContextBuilder) Clear() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.messages = make([]adk.Message, 0)
	cb.system = ""
}

// Len returns the number of messages in the context.
func (cb *ContextBuilder) Len() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return len(cb.messages)
}
