package adk

import "context"

// Message is one chat item in the conversation state.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ToolCall describes one model-requested tool invocation.
type ToolCall struct {
	Name  string            `json:"name"`
	Input map[string]string `json:"input"`
}

// ToolResult is the result of a tool call that gets fed back to the model.
type ToolResult struct {
	Name    string `json:"name"`
	Output  string `json:"output"`
	IsError bool   `json:"isError"`
}

// ModelResponse represents the model output for a single iteration.
type ModelResponse struct {
	AssistantMessage string     `json:"assistantMessage"`
	ToolCalls        []ToolCall `json:"toolCalls"`
	Done             bool       `json:"done"`
}

// Model decides next assistant output and optional tool calls.
type Model interface {
	Respond(ctx context.Context, state SessionState) (ModelResponse, error)
}

// SessionState is the mutable conversation state during one agent turn.
type SessionState struct {
	SystemPrompt string
	Messages     []Message
	LastToolRuns []ToolResult
}
