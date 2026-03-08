package adk

import (
	"context"
	"fmt"
	"strings"
)

// RuleModel is a zero-dependency fallback model that keeps darci-go self-contained.
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

	// Handle tool result responses
	if len(state.LastToolRuns) > 0 {
		last := state.LastToolRuns[len(state.LastToolRuns)-1]
		if last.IsError {
			return ModelResponse{AssistantMessage: fmt.Sprintf("Tool %s failed: %s", last.Name, last.Output), Done: true}, nil
		}
		return ModelResponse{AssistantMessage: fmt.Sprintf("Tool %s result:\n%s", last.Name, last.Output), Done: true}, nil
	}

	// Time commands
	if strings.HasPrefix(lower, "/time") || strings.Contains(lower, "what time") || strings.Contains(lower, "current time") || strings.HasPrefix(lower, "time_now") {
		return ModelResponse{ToolCalls: []ToolCall{{Name: "time_now", Input: map[string]string{}}}}, nil
	}

	// List directory commands
	if strings.HasPrefix(lower, "/ls") || strings.HasPrefix(lower, "list_dir") || strings.HasPrefix(lower, "list") {
		path := extractParam(text, "path")
		if path == "" {
			path = extractPath(text, "/ls")
		}
		if path == "" {
			path = "."
		}
		return ModelResponse{ToolCalls: []ToolCall{{Name: "list_dir", Input: map[string]string{"path": path}}}}, nil
	}

	// Read file commands
	if strings.HasPrefix(lower, "/cat") || strings.HasPrefix(lower, "read_file") || (strings.HasPrefix(lower, "read ") && !strings.Contains(lower, "web")) {
		path := extractParam(text, "path")
		if path == "" {
			path = extractPath(text, "/cat")
		}
		if path == "" {
			return ModelResponse{AssistantMessage: "Please specify a file path. Usage: /cat <filepath>", Done: true}, nil
		}
		return ModelResponse{ToolCalls: []ToolCall{{Name: "read_file", Input: map[string]string{"path": path}}}}, nil
	}

	// Write file commands
	if strings.HasPrefix(lower, "/write") || strings.HasPrefix(lower, "write_file") {
		path := extractParam(text, "path")
		content := extractParam(text, "content")
		if path == "" {
			return ModelResponse{AssistantMessage: "Please specify a file path. Usage: /write <filepath>", Done: true}, nil
		}
		return ModelResponse{ToolCalls: []ToolCall{{Name: "write_file", Input: map[string]string{"path": path, "content": content}}}}, nil
	}

	// Execute shell commands
	if strings.HasPrefix(lower, "/exec") || strings.HasPrefix(lower, "exec ") || strings.HasPrefix(lower, "execute") {
		cmd := extractParam(text, "command")
		if cmd == "" {
			return ModelResponse{AssistantMessage: "Please specify a command. Usage: /exec <command>", Done: true}, nil
		}
		return ModelResponse{ToolCalls: []ToolCall{{Name: "exec", Input: map[string]string{"command": cmd}}}}, nil
	}

	// Web search commands
	if strings.HasPrefix(lower, "/search") || strings.HasPrefix(lower, "web_search") || strings.HasPrefix(lower, "search ") {
		query := extractParam(text, "query")
		if query == "" {
			return ModelResponse{AssistantMessage: "Please specify a search query. Usage: /search <query>", Done: true}, nil
		}
		return ModelResponse{ToolCalls: []ToolCall{{Name: "web_search", Input: map[string]string{"query": query}}}}, nil
	}

	// Web fetch commands
	if strings.HasPrefix(lower, "/fetch") || strings.HasPrefix(lower, "web_fetch") || strings.HasPrefix(lower, "fetch ") {
		url := extractParam(text, "url")
		if url == "" {
			return ModelResponse{AssistantMessage: "Please specify a URL. Usage: /fetch <url>", Done: true}, nil
		}
		return ModelResponse{ToolCalls: []ToolCall{{Name: "web_fetch", Input: map[string]string{"url": url}}}}, nil
	}

	// Message/send commands
	if strings.Contains(lower, "send message") || strings.Contains(lower, "tell user") || strings.Contains(lower, "notify") {
		content := text
		if idx := strings.Index(lower, "message"); idx != -1 && idx < len(text)-1 {
			content = text[idx+8:]
		}
		return ModelResponse{ToolCalls: []ToolCall{{Name: "message", Input: map[string]string{"content": strings.TrimSpace(content)}}}}, nil
	}

	// Default response
	return ModelResponse{
		AssistantMessage: "🐈 DarCI-Go ADK is ready! I can help you with:\n" +
			"  • File operations: read, write, edit files\n" +
			"  • Shell commands: execute system commands\n" +
			"  • Web access: search and fetch web pages\n" +
			"  • Time: show current UTC time\n" +
			"\nTry: /time, /ls, /cat <file>, /exec <cmd>, /search <query>, /fetch <url>",
		Done: true,
	}, nil
}

// extractParam extracts a parameter value from text like "command=echo hello"
func extractParam(text, paramName string) string {
	pattern := paramName + "="
	idx := strings.Index(text, pattern)
	if idx == -1 {
		return ""
	}
	rest := text[idx+len(pattern):]
	// Find the end (space or end of string)
	endIdx := strings.IndexAny(rest, " \t\n")
	if endIdx == -1 {
		return rest
	}
	return rest[:endIdx]
}

// extractPath extracts a file path from text after a prefix
func extractPath(text, prefix string) string {
	idx := strings.Index(strings.ToLower(text), strings.ToLower(prefix))
	if idx == -1 {
		return ""
	}
	rest := strings.TrimSpace(text[idx+len(prefix):])
	if rest == "" {
		return ""
	}
	// Handle quoted paths
	if strings.HasPrefix(rest, "\"") || strings.HasPrefix(rest, "'") {
		quote := rest[0:1]
		endIdx := strings.Index(rest[1:], quote)
		if endIdx != -1 {
			return rest[1 : endIdx+1]
		}
	}
	// Skip common words and take the actual path
	skipWords := []string{"the", "file", "at", "to", "from", "in", "contents", "of", "this"}
	parts := strings.Fields(rest)
	for i, part := range parts {
		skip := false
		for _, skipWord := range skipWords {
			if strings.EqualFold(part, skipWord) {
				skip = true
				break
			}
		}
		if !skip {
			return strings.Trim(part, " \t\n\r\"':")
		}
		_ = i
	}
	return ""
}

// extractCommand extracts a command string
func extractCommand(text, prefix string) string {
	idx := strings.Index(strings.ToLower(text), strings.ToLower(prefix))
	if idx == -1 {
		return ""
	}
	return strings.TrimSpace(text[idx+len(prefix):])
}

// extractQuery extracts a search query
func extractQuery(text, prefix string) string {
	idx := strings.Index(strings.ToLower(text), strings.ToLower(prefix))
	if idx == -1 {
		return ""
	}
	rest := strings.TrimSpace(text[idx+len(prefix):])
	// Remove quotes if present
	rest = strings.Trim(rest, "\"'")
	return rest
}

// extractURL extracts a URL
func extractURL(text, prefix string) string {
	idx := strings.Index(strings.ToLower(text), strings.ToLower(prefix))
	if idx == -1 {
		return ""
	}
	rest := strings.TrimSpace(text[idx+len(prefix):])
	// Take first word as URL
	parts := strings.Fields(rest)
	if len(parts) == 0 {
		return ""
	}
	url := strings.Trim(parts[0], " \t\n\r\"'")
	// Add https:// if missing
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}
	return url
}
