package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// WebTools provides web-related tools.
type WebTools struct {
	client  *http.Client
	timeout time.Duration
}

// NewWebTools creates new web tools.
func NewWebTools() *WebTools {
	return &WebTools{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		timeout: 30 * time.Second,
	}
}

// FetchURL fetches content from a URL.
func (w *WebTools) FetchURL(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	url, ok := args["url"].(string)
	if !ok {
		return nil, fmt.Errorf("url is required")
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", "darci-agent/0.1.0")

	// Execute request
	resp, err := w.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	// Read response body
	content, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB limit
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return string(content), nil
}

// Search performs a web search (placeholder).
func (w *WebTools) Search(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	query, ok := args["query"].(string)
	if !ok {
		return nil, fmt.Errorf("query is required")
	}

	// This is a placeholder - in a real implementation, this would call a search API
	return fmt.Sprintf("Search results for: %s", query), nil
}
