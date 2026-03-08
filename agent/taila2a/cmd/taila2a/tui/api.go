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
 
	var result struct {
		Notifications []Notification `json:"notifications"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Notifications, nil
}
