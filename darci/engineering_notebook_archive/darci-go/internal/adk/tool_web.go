package adk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// WebSearchTool searches the web using Brave Search API.
type WebSearchTool struct {
	APIKey     string
	MaxResults int
}

func (WebSearchTool) Name() string { return "web_search" }
func (WebSearchTool) Description() string {
	return "Search the web using Brave Search API. Returns titles, URLs, and snippets."
}
func (t WebSearchTool) Run(ctx context.Context, input map[string]string) (string, error) {
	query := strings.TrimSpace(input["query"])
	if query == "" {
		return "", fmt.Errorf("query is required")
	}

	apiKey := t.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("BRAVE_API_KEY")
	}
	if apiKey == "" {
		return "", fmt.Errorf("BRAVE_API_KEY not set. Configure it in environment or config")
	}

	count := t.MaxResults
	if count == 0 {
		count = 5
	}
	if count > 10 {
		count = 10
	}
	if count < 1 {
		count = 1
	}

	braveURL := "https://api.search.brave.com/res/v1/web/search"
	params := url.Values{}
	params.Set("q", query)
	params.Set("count", fmt.Sprintf("%d", count))

	req, err := http.NewRequestWithContext(ctx, "GET", braveURL+"?"+params.Encode(), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Subscription-Token", apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error: %s", string(body))
	}

	var result struct {
		Web struct {
			Results []struct {
				Title       string `json:"title"`
				URL         string `json:"url"`
				Description string `json:"description"`
			} `json:"results"`
		} `json:"web"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	results := result.Web.Results
	if len(results) == 0 {
		return fmt.Sprintf("No results found for: %s", query), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Search results for: %s\n\n", query))
	for i, r := range results {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, r.Title))
		sb.WriteString(fmt.Sprintf("   URL: %s\n", r.URL))
		if r.Description != "" {
			sb.WriteString(fmt.Sprintf("   %s\n\n", r.Description))
		}
	}

	return sb.String(), nil
}

// WebFetchTool fetches web page content.
type WebFetchTool struct{}

func (WebFetchTool) Name() string { return "web_fetch" }
func (WebFetchTool) Description() string {
	return "Fetch and extract content from a web page URL. Returns text content."
}
func (t WebFetchTool) Run(ctx context.Context, input map[string]string) (string, error) {
	urlStr := strings.TrimSpace(input["url"])
	if urlStr == "" {
		return "", fmt.Errorf("url is required")
	}

	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		return "", fmt.Errorf("URL must start with http:// or https://")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; DarCI-Bot/1.0)")

	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Simple HTML to text conversion
	content := stripHTMLTags(string(body))
	content = normalizeWhitespace(content)

	if len(content) > 10000 {
		content = content[:10000] + "\n\n[Content truncated to 10000 characters]"
	}

	return content, nil
}

// stripHTMLTags removes HTML tags from text.
func stripHTMLTags(html string) string {
	// Remove script and style blocks
	result := html
	result = removeBetween(result, "<script", "</script>")
	result = removeBetween(result, "<style", "</style>")

	// Remove all tags
	var sb strings.Builder
	inTag := false
	for _, r := range result {
		if r == '<' {
			inTag = true
		} else if r == '>' {
			inTag = false
		} else if !inTag {
			sb.WriteRune(r)
		}
	}

	// Decode common HTML entities
	text := sb.String()
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")

	return text
}

// removeBetween removes content between start and end markers.
func removeBetween(s, start, end string) string {
	for {
		startIdx := strings.Index(strings.ToLower(s), strings.ToLower(start))
		if startIdx == -1 {
			break
		}
		endIdx := strings.Index(strings.ToLower(s[startIdx:]), strings.ToLower(end))
		if endIdx == -1 {
			break
		}
		s = s[:startIdx] + s[startIdx+endIdx+len(end):]
	}
	return s
}

// normalizeWhitespace normalizes whitespace in text.
func normalizeWhitespace(s string) string {
	var sb strings.Builder
	lastWasSpace := false
	for _, r := range s {
		isSpace := r == ' ' || r == '\t'
		if isSpace {
			if !lastWasSpace {
				sb.WriteRune(' ')
			}
			lastWasSpace = true
		} else {
			sb.WriteRune(r)
			lastWasSpace = false
		}
	}
	return strings.TrimSpace(sb.String())
}
