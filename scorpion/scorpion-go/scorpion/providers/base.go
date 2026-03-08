package providers

import (
	"context"

	"scorpion-go/internal/adk"
)

// LLMProvider is the interface that all LLM providers must implement.
type LLMProvider interface {
	// Name returns the provider name.
	Name() string

	// Initialize initializes the provider.
	Initialize(ctx context.Context) error

	// Chat sends a chat request and returns the response.
	Chat(ctx context.Context, messages []adk.Message, tools []adk.Tool) (*LLMResponse, error)

	// ChatStream sends a streaming chat request.
	ChatStream(ctx context.Context, messages []adk.Message, tools []adk.Tool) (<-chan StreamChunk, error)

	// IsReady returns whether the provider is ready.
	IsReady() bool
}

// LLMResponse represents an LLM response.
type LLMResponse struct {
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	Done      bool       `json:"done"`
	Model     string     `json:"model"`
	Usage     *Usage     `json:"usage,omitempty"`
}

// ToolCall represents a tool call from the LLM.
type ToolCall struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// Usage represents token usage.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamChunk represents a streaming response chunk.
type StreamChunk struct {
	Content   string     `json:"content,omitempty"`
	ToolCall  *ToolCall  `json:"tool_call,omitempty"`
	Done      bool       `json:"done,omitempty"`
	Error     error      `json:"error,omitempty"`
}

// ProviderConfig holds common provider configuration.
type ProviderConfig struct {
	APIKey     string `json:"api_key"`
	BaseURL    string `json:"base_url,omitempty"`
	Model      string `json:"model"`
	Timeout    int    `json:"timeout,omitempty"`
	MaxTokens  int    `json:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

// BaseProvider provides common functionality for providers.
type BaseProvider struct {
	name   string
	config *ProviderConfig
	ready  bool
}

// NewBaseProvider creates a new base provider.
func NewBaseProvider(name string, config *ProviderConfig) *BaseProvider {
	return &BaseProvider{
		name:   name,
		config: config,
	}
}

// Name returns the provider name.
func (b *BaseProvider) Name() string {
	return b.name
}

// IsReady returns whether the provider is ready.
func (b *BaseProvider) IsReady() bool {
	return b.ready
}

// SetReady sets the ready state.
func (b *BaseProvider) SetReady(ready bool) {
	b.ready = ready
}

// GetConfig returns the provider configuration.
func (b *BaseProvider) GetConfig() *ProviderConfig {
	return b.config
}
