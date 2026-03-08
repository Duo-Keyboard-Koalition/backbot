package tui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Taila2aClient struct {
	baseURL string
	client  *http.Client
}

func NewClient(port int) *Taila2aClient {
	return &Taila2aClient{
		baseURL: fmt.Sprintf("http://localhost:%d", port),
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *Taila2aClient) FetchAgents() ([]Agent, error) {
	resp, err := c.client.Get(c.baseURL + "/agents")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Agents []Agent `json:"agents"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Agents, nil
}

func (c *Taila2aClient) FetchNotifications() ([]Notification, error) {
	resp, err := c.client.Get(c.baseURL + "/trigger/notifications")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// The server returns either structured objects or plain strings.
	// Decode each element as raw JSON and handle both forms.
	var result struct {
		Notifications []json.RawMessage `json:"notifications"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	notifications := make([]Notification, 0, len(result.Notifications))
	for _, raw := range result.Notifications {
		// Try structured notification first.
		var n Notification
		if err := json.Unmarshal(raw, &n); err == nil && n.Message != "" {
			notifications = append(notifications, n)
			continue
		}
		// Fall back to plain string.
		var s string
		if err := json.Unmarshal(raw, &s); err == nil && s != "" {
			notifications = append(notifications, Notification{
				Timestamp: time.Now().Format("15:04:05"),
				Level:     "INFO",
				Message:   s,
			})
		}
	}
	return notifications, nil
}
