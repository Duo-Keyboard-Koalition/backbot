package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"darci-go/darci/config"
	"darci-go/darci/state"
)

// DiscoverAgentsTool discovers all agents online on the Tailscale tailnet
type DiscoverAgentsTool struct {
	config *config.DarciConfig
	store  *state.TaskStore
	client *http.Client
}

func NewDiscoverAgentsTool(cfg *config.DarciConfig, store *state.TaskStore) *DiscoverAgentsTool {
	return &DiscoverAgentsTool{
		config: cfg,
		store:  store,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (t *DiscoverAgentsTool) Name() string {
	return "discover_agents"
}

func (t *DiscoverAgentsTool) Description() string {
	return "Discover all agents currently online on the Tailscale tailnet via taila2a. Updates the internal agent registry. Call this before assigning tasks."
}

func (t *DiscoverAgentsTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

func (t *DiscoverAgentsTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	url := fmt.Sprintf("%s/agents", t.config.BridgeLocalURL)

	resp, err := t.client.Get(url)
	if err != nil {
		return fmt.Sprintf("Error: taila2a bridge not reachable at %s. Is it running?", url), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("Error: HTTP %d", resp.StatusCode), nil
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return fmt.Sprintf("Error decoding response: %v", err), nil
	}

	agentsInterface, _ := data["agents"].([]interface{})
	agents := make([]map[string]interface{}, len(agentsInterface))
	for i, a := range agentsInterface {
		if m, ok := a.(map[string]interface{}); ok {
			agents[i] = m
		}
	}

	now := time.Now().UTC()
	ctx_data, err := t.store.GetContext()
	if err != nil || ctx_data == nil {
		ctx_data = &state.AgentContext{
			AgentAssignments: make(map[string]*state.AgentAssignment),
			DarciState: &state.DarciState{
				ActiveMonitors: []string{},
			},
		}
	}

	if ctx_data.DarciState == nil {
		ctx_data.DarciState = &state.DarciState{
			ActiveMonitors: []string{},
		}
	}
	ctx_data.DarciState.LastDiscovery = &now
	t.store.UpdateContext(map[string]interface{}{"darci_state": ctx_data.DarciState})

	if len(agents) == 0 {
		return "No agents discovered on the tailnet.", nil
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("Discovered %d agent(s):\n", len(agents)))
	lines = append(lines, "| Name | IP | Online | Services |")
	lines = append(lines, "|---|---|---|---|")

	for _, agent := range agents {
		name, _ := agent["name"].(string)
		ip, _ := agent["ip"].(string)
		online := "❌"
		if v, ok := agent["online"].(bool); ok && v {
			online = "✅"
		}

		var services []string
		if gateways, ok := agent["gateways"].([]interface{}); ok {
			for _, g := range gateways {
				if gw, ok := g.(map[string]interface{}); ok {
					service, _ := gw["service"].(string)
					port, _ := gw["port"].(string)
					if service != "" && port != "" {
						services = append(services, fmt.Sprintf("%s:%s", service, port))
					}
				}
			}
		}

		servicesStr := "none"
		if len(services) > 0 {
			servicesStr = strings.Join(services, ", ")
		}

		lines = append(lines, fmt.Sprintf("| %s | %s | %s | %s |", name, ip, online, servicesStr))
	}

	return strings.Join(lines, "\n"), nil
}

// SendDarciMessageTool sends a DARCI message to a worker agent via tailbridge
type SendDarciMessageTool struct {
	config *config.DarciConfig
	client *http.Client
}

func NewSendDarciMessageTool(cfg *config.DarciConfig) *SendDarciMessageTool {
	return &SendDarciMessageTool{
		config: cfg,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (t *SendDarciMessageTool) Name() string {
	return "send_darci_message"
}

func (t *SendDarciMessageTool) Description() string {
	return "Send a DARCI message to a worker agent (openclaw, nanobot, sclaw) via tailbridge. Use message_type='darci_directive' to assign or correct work, or 'darci_status_request' to ask what the agent is doing."
}

func (t *SendDarciMessageTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"dest_node": map[string]interface{}{
				"type":        "string",
				"description": "Tailnet node name of the target agent",
			},
			"message_type": map[string]interface{}{
				"type": "string",
				"enum": []string{"darci_directive", "darci_status_request"},
			},
			"payload": map[string]interface{}{
				"type":        "object",
				"description": "Message payload (task_id, goal, priority, etc.)",
			},
		},
		"required": []string{"dest_node", "message_type", "payload"},
	}
}

func (t *SendDarciMessageTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	destNode, _ := args["dest_node"].(string)
	messageType, _ := args["message_type"].(string)
	payload, _ := args["payload"].(map[string]interface{})

	if destNode == "" || messageType == "" {
		return "", fmt.Errorf("dest_node and message_type are required")
	}

	url := fmt.Sprintf("%s/send", t.config.BridgeLocalURL)

	envelope := map[string]interface{}{
		"dest_node": destNode,
		"payload": map[string]interface{}{
			"type": messageType,
		},
	}

	// Merge payload into the payload.type
	if payload != nil {
		for k, v := range payload {
			envelope["payload"].(map[string]interface{})[k] = v
		}
	}

	body, err := json.Marshal(envelope)
	if err != nil {
		return "", fmt.Errorf("Error marshaling request: %v", err)
	}

	resp, err := t.client.Post(url, "application/json", strings.NewReader(string(body)))
	if err != nil {
		return fmt.Sprintf("Error: taila2a bridge not reachable at %s. Is it running?", url), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Sprintf("Error sending to %s: HTTP %d", destNode, resp.StatusCode), nil
	}

	return fmt.Sprintf("Message sent to %s (type: %s)", destNode, messageType), nil
}
