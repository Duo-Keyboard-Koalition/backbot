package adk

import (
	"context"
	"fmt"
	"strings"
)

// RuleModel is a zero-dependency fallback model that keeps scorpion-go self-contained.
type RuleModel struct{}

func (RuleModel) Respond(_ context.Context, state SessionState) (ModelResponse, error) {
	if len(state.Messages) == 0 {
		return ModelResponse{AssistantMessage: "No input provided.", Done: true}, nil
	}
	latest := state.Messages[len(state.Messages)-1]
	if latest.Role != "user" {
		return ModelResponse{AssistantMessage: "Waiting for user input.", Done: true}, nil
	}

	text := strings.TrimSpace(latest.Content)
	lower := strings.ToLower(text)

	if len(state.LastToolRuns) > 0 {
		last := state.LastToolRuns[len(state.LastToolRuns)-1]
		if last.IsError {
			return ModelResponse{AssistantMessage: fmt.Sprintf("Tool %s failed: %s", last.Name, last.Output), Done: true}, nil
		}
		return ModelResponse{AssistantMessage: fmt.Sprintf("Tool %s result:\n%s", last.Name, last.Output), Done: true}, nil
	}

	if strings.HasPrefix(lower, "/time") {
		return ModelResponse{ToolCalls: []ToolCall{{Name: "time_now", Input: map[string]string{}}}}, nil
	}
	if strings.HasPrefix(lower, "/ls") {
		path := strings.TrimSpace(strings.TrimPrefix(text, "/ls"))
		return ModelResponse{ToolCalls: []ToolCall{{Name: "list_dir", Input: map[string]string{"path": path}}}}, nil
	}
	if strings.HasPrefix(lower, "/cat") {
		path := strings.TrimSpace(strings.TrimPrefix(text, "/cat"))
		return ModelResponse{ToolCalls: []ToolCall{{Name: "read_file", Input: map[string]string{"path": path}}}}, nil
	}

	return ModelResponse{
		AssistantMessage: "Self-contained Scorpion-Go ADK is running. Try /time, /ls <path>, or /cat <path>.",
		Done:             true,
	}, nil
}
