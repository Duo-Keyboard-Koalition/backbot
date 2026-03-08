package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"scorpion-go/internal/adk"
)

// GeminiProvider implements the Google Gemini provider.
type GeminiProvider struct {
	*BaseProvider
	client  *http.Client
	apiKey  string
	baseURL string
	model   string
}

// GeminiConfig holds Gemini-specific configuration.
type GeminiConfig struct {
	APIKey      string  `json:"api_key"`
	BaseURL     string  `json:"base_url,omitempty"`
	Model       string  `json:"model"`
	Timeout     int     `json:"timeout,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

// NewGeminiProvider creates a new Gemini provider.
func NewGeminiProvider(config *GeminiConfig) *GeminiProvider {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com/v1beta"
	}

	timeout := config.Timeout
	if timeout <= 0 {
		timeout = 30
	}

	model := config.Model
	if model == "" {
		model = "gemini-2.5-flash"
	}

	return &GeminiProvider{
		BaseProvider: NewBaseProvider("gemini", &ProviderConfig{
			APIKey:      config.APIKey,
			BaseURL:     baseURL,
			Model:       model,
			Timeout:     timeout,
			MaxTokens:   config.MaxTokens,
			Temperature: config.Temperature,
		}),
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		apiKey:  config.APIKey,
		baseURL: baseURL,
		model:   model,
	}
}

// Initialize initializes the Gemini provider.
func (g *GeminiProvider) Initialize(ctx context.Context) error {
	if g.apiKey == "" {
		return fmt.Errorf("Gemini API key is required")
	}

	g.SetReady(true)
	return nil
}

// Chat sends a chat request to Gemini.
func (g *GeminiProvider) Chat(ctx context.Context, messages []adk.Message, tools []adk.Tool) (*LLMResponse, error) {
	if !g.IsReady() {
		return nil, fmt.Errorf("provider not ready")
	}

	// Build request
	reqBody, err := g.buildRequest(messages, tools)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", g.baseURL, g.model, g.apiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	// Parse response
	return g.parseResponse(body)
}

// ChatStream sends a streaming chat request to Gemini.
func (g *GeminiProvider) ChatStream(ctx context.Context, messages []adk.Message, tools []adk.Tool) (<-chan StreamChunk, error) {
	if !g.IsReady() {
		return nil, fmt.Errorf("provider not ready")
	}

	// For now, return a single chunk with the full response
	chunks := make(chan StreamChunk, 1)
	go func() {
		defer close(chunks)
		
		response, err := g.Chat(ctx, messages, tools)
		if err != nil {
			chunks <- StreamChunk{Error: err}
			return
		}

		chunks <- StreamChunk{
			Content: response.Content,
			Done:    true,
		}
	}()

	return chunks, nil
}

// buildRequest builds the Gemini API request body.
func (g *GeminiProvider) buildRequest(messages []adk.Message, tools []adk.Tool) ([]byte, error) {
	// Convert messages to Gemini format
	contents := make([]map[string]interface{}, 0, len(messages))
	for _, msg := range messages {
		role := "user"
		if msg.Role == "assistant" || msg.Role == "model" {
			role = "model"
		}
		if msg.Role == "system" {
			// System messages become user messages with special prefix
			role = "user"
		}

		content := map[string]interface{}{
			"role":  role,
			"parts": []map[string]string{{"text": msg.Content}},
		}
		contents = append(contents, content)
	}

	// Build generation config
	generationConfig := map[string]interface{}{
		"temperature": g.BaseProvider.config.Temperature,
	}

	if g.BaseProvider.config.MaxTokens > 0 {
		generationConfig["maxOutputTokens"] = g.BaseProvider.config.MaxTokens
	}

	request := map[string]interface{}{
		"contents":         contents,
		"generationConfig": generationConfig,
	}

	return json.Marshal(request)
}

// parseResponse parses the Gemini API response.
func (g *GeminiProvider) parseResponse(body []byte) (*LLMResponse, error) {
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract content from response
	candidates, ok := result["candidates"].([]interface{})
	if !ok || len(candidates) == 0 {
		return &LLMResponse{
			Content: "",
			Done:    true,
		}, nil
	}

	firstCandidate := candidates[0].(map[string]interface{})
	content, ok := firstCandidate["content"].(map[string]interface{})
	if !ok {
		return &LLMResponse{
			Content: "",
			Done:    true,
		}, nil
	}

	parts, ok := content["parts"].([]interface{})
	if !ok || len(parts) == 0 {
		return &LLMResponse{
			Content: "",
			Done:    true,
		}, nil
	}

	// Extract text from first part
	firstPart := parts[0].(map[string]interface{})
	text, _ := firstPart["text"].(string)

	return &LLMResponse{
		Content: text,
		Done:    true,
		Model:   g.model,
	}, nil
}
