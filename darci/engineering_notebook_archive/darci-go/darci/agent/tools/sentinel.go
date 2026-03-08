package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"darci-go/darci/bus"
	"darci-go/darci/config"
	"darci-go/darci/state"

	"github.com/gorilla/websocket"
)

// SentinelSnapshot represents the state of a Sentinel monitoring session
type SentinelSnapshot struct {
	RiskScore       float64  `json:"risk_score"`
	FailureTypes    []string `json:"failure_types"`
	StepCount       int      `json:"step_count"`
	InterventionType string  `json:"intervention_type,omitempty"`
}

// SentinelMonitorRegistry tracks running WebSocket monitor tasks (one per agent)
type SentinelMonitorRegistry struct {
	mu      sync.RWMutex
	tasks   map[string]context.CancelFunc
	cancel  context.CancelFunc
	ctx     context.Context
}

func NewSentinelMonitorRegistry() *SentinelMonitorRegistry {
	ctx, cancel := context.WithCancel(context.Background())
	return &SentinelMonitorRegistry{
		tasks:  make(map[string]context.CancelFunc),
		cancel: cancel,
		ctx:    ctx,
	}
}

func (r *SentinelMonitorRegistry) IsMonitoring(nodeName string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.tasks[nodeName]
	return exists
}

func (r *SentinelMonitorRegistry) Start(nodeName string, cancel context.CancelFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks[nodeName] = cancel
}

func (r *SentinelMonitorRegistry) Stop(nodeName string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if cancel, exists := r.tasks[nodeName]; exists {
		cancel()
		delete(r.tasks, nodeName)
	}
}

func (r *SentinelMonitorRegistry) StopAll() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, cancel := range r.tasks {
		cancel()
	}
	r.tasks = make(map[string]context.CancelFunc)
}

func (r *SentinelMonitorRegistry) Active() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	active := make([]string, 0, len(r.tasks))
	for nodeName := range r.tasks {
		active = append(active, nodeName)
	}
	return active
}

// MonitorAgentTool starts monitoring a Responsible agent's Sentinel risk stream
type MonitorAgentTool struct {
	config   *config.DarciConfig
	store    *state.TaskStore
	messageBus *bus.MessageBus
	registry *SentinelMonitorRegistry
}

func NewMonitorAgentTool(
	cfg *config.DarciConfig,
	store *state.TaskStore,
	bus *bus.MessageBus,
	registry *SentinelMonitorRegistry,
) *MonitorAgentTool {
	return &MonitorAgentTool{
		config:     cfg,
		store:      store,
		messageBus: bus,
		registry:   registry,
	}
}

func (t *MonitorAgentTool) Name() string {
	return "monitor_agent"
}

func (t *MonitorAgentTool) Description() string {
	return "Start monitoring a Responsible agent's Sentinel risk stream. Runs in the background — alerts you automatically when risk_score >= 0.5 (Approver signal) or when a HALT intervention fires. Requires a task_id to track state."
}

func (t *MonitorAgentTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"node_name": map[string]interface{}{
				"type":        "string",
				"description": "Tailnet name of the agent to monitor",
			},
			"sentinel_url": map[string]interface{}{
				"type":        "string",
				"description": "WebSocket URL of the agent's Sentinel backend (e.g. ws://192.168.x.x:8000/ws/run)",
			},
			"goal": map[string]interface{}{
				"type":        "string",
				"description": "The agent's current goal to send to Sentinel",
			},
			"task_id": map[string]interface{}{
				"type":        "string",
				"description": "DarCI task ID this agent is working on",
			},
			"api_key": map[string]interface{}{
				"type":        "string",
				"description": "Optional Gemini API key for the Sentinel backend",
			},
		},
		"required": []string{"node_name", "sentinel_url", "goal", "task_id"},
	}
}

func (t *MonitorAgentTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	nodeName, _ := args["node_name"].(string)
	sentinelURL, _ := args["sentinel_url"].(string)
	goal, _ := args["goal"].(string)
	taskID, _ := args["task_id"].(string)
	apiKey, _ := args["api_key"].(string)

	if nodeName == "" || sentinelURL == "" || goal == "" || taskID == "" {
		return "", fmt.Errorf("node_name, sentinel_url, goal, and task_id are required")
	}

	if t.registry.IsMonitoring(nodeName) {
		return fmt.Sprintf("Already monitoring '%s'. Use monitor_agent with a different node, or wait.", nodeName), nil
	}

	task, err := t.store.Get(taskID)
	if err != nil || task == nil {
		return fmt.Sprintf("Error: task %s not found. Create it first with task_create.", taskID), nil
	}

	// Create a new context for this monitor
	monitorCtx, cancel := context.WithCancel(ctx)

	go t.monitorLoop(monitorCtx, nodeName, sentinelURL, goal, taskID, apiKey, cancel)

	t.registry.Start(nodeName, cancel)

	active := t.registry.Active()
	return fmt.Sprintf("Monitoring started for '%s' on task %s.\nSentinel URL: %s\nActive monitors: %v",
		nodeName, taskID, sentinelURL, active), nil
}

func (t *MonitorAgentTool) monitorLoop(
	ctx context.Context,
	nodeName, sentinelURL, goal, taskID, apiKey string,
	cancel context.CancelFunc,
) {
	defer cancel()

	conn, _, err := websocket.DefaultDialer.Dial(sentinelURL, nil)
	if err != nil {
		t.messageBus.PublishInbound(&bus.InboundMessage{
			Channel:   "system",
			SenderID:  "sentinel-monitor",
			ChatID:    "darci:direct",
			Content:   fmt.Sprintf("[ERROR] Failed to connect to Sentinel at %s: %v", sentinelURL, err),
		})
		return
	}
	defer conn.Close()

	// Send initial configuration
	initMsg := map[string]interface{}{
		"goal":      goal,
		"api_key":   apiKey,
		"max_steps": 50,
	}
	if err := conn.WriteJSON(initMsg); err != nil {
		t.messageBus.PublishInbound(&bus.InboundMessage{
			Channel:   "system",
			SenderID:  "sentinel-monitor",
			ChatID:    "darci:direct",
			Content:   fmt.Sprintf("[ERROR] Failed to initialize Sentinel: %v", err),
		})
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					t.messageBus.PublishInbound(&bus.InboundMessage{
						Channel:   "system",
						SenderID:  "sentinel-monitor",
						ChatID:    "darci:direct",
						Content:   fmt.Sprintf("[ERROR] Sentinel connection lost for '%s': %v", nodeName, err),
					})
				}
				return
			}

			var event map[string]interface{}
			if err := json.Unmarshal(message, &event); err != nil {
				continue
			}

			eventType, _ := event["type"].(string)

			switch eventType {
			case "step":
				riskScore, _ := event["risk_score"].(float64)
				failureTypesInterface, _ := event["failure_types"].([]interface{})
				stepNum, _ := event["step"].(map[string]interface{})["step_number"].(float64)

				failureTypes := make([]string, len(failureTypesInterface))
				for i, ft := range failureTypesInterface {
					if s, ok := ft.(string); ok {
						failureTypes[i] = s
					}
				}

				t.store.Update(taskID, map[string]interface{}{
					"sentinel_snapshot": SentinelSnapshot{
						RiskScore:    riskScore,
						FailureTypes: failureTypes,
						StepCount:    int(stepNum),
					},
				})
				t.store.SetAgentAssignment(nodeName, taskID, "responsible", riskScore, "")

				if riskScore >= 0.5 {
					t.store.Update(taskID, map[string]interface{}{
						"status": "at_risk",
					})
					t.messageBus.PublishInbound(&bus.InboundMessage{
						Channel:  "system",
						SenderID: "sentinel-monitor",
						ChatID:   "darci:direct",
						Content: fmt.Sprintf(
							"[RISK ALERT] Agent '%s' on task %s: risk_score=%.2f, failures=%v. "+
								"As Driver, send a darci_directive to refocus the agent on its goal.",
							nodeName, taskID, riskScore, failureTypes,
						),
					})
				}

			case "intervention":
				intervention, _ := event["intervention"].(map[string]interface{})
				interventionType, _ := intervention["intervention_type"].(string)

				t.store.Update(taskID, map[string]interface{}{
					"sentinel_snapshot": SentinelSnapshot{
						InterventionType: interventionType,
					},
				})

				if interventionType == "HALT" {
					t.store.Update(taskID, map[string]interface{}{
						"status": "blocked",
					})
					t.messageBus.PublishInbound(&bus.InboundMessage{
						Channel:  "system",
						SenderID: "sentinel-monitor",
						ChatID:   "darci:direct",
						Content: fmt.Sprintf(
							"[HALT] Agent '%s' task %s has been halted by Sentinel (Approver veto). "+
								"Task is now blocked. Notify the user and create a notebook entry documenting this intervention.",
							nodeName, taskID,
						),
					})
				} else {
					t.messageBus.PublishInbound(&bus.InboundMessage{
						Channel:  "system",
						SenderID: "sentinel-monitor",
						ChatID:   "darci:direct",
						Content: fmt.Sprintf(
							"[INTERVENTION] Agent '%s' task %s: Sentinel issued %s. Monitor continues.",
							nodeName, taskID, interventionType,
						),
					})
				}

			case "complete", "timeout":
				status := "completed"
				if eventType == "timeout" {
					status = "timed out"
				}
				t.store.Update(taskID, map[string]interface{}{
					"status": "completed",
				})
				t.messageBus.PublishInbound(&bus.InboundMessage{
					Channel:  "system",
					SenderID: "sentinel-monitor",
					ChatID:   "darci:direct",
					Content: fmt.Sprintf(
						"[COMPLETE] Agent '%s' task %s has finished (%s). "+
							"Create a notebook entry to document the session.",
						nodeName, taskID, status,
					),
				})
				return
			}
		}
	}
}
