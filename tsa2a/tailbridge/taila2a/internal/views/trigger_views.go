package views

import (
	"encoding/json"
	"net/http"
	"time"
)

// TriggerStatusView renders trigger status information
type TriggerStatusView struct{}

// NewTriggerStatusView creates a new trigger status view
func NewTriggerStatusView() *TriggerStatusView {
	return &TriggerStatusView{}
}

// Render renders trigger status
func (v *TriggerStatusView) Render(w http.ResponseWriter, status map[string]interface{}) {
	response := map[string]interface{}{
		"status":    "ok",
		"trigger":   status,
		"timestamp": time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// AgentStatusView renders agent status with buffer information
type AgentStatusView struct{}

// NewAgentStatusView creates a new agent status view
func NewAgentStatusView() *AgentStatusView {
	return &AgentStatusView{}
}

// AgentStatusData contains agent status data for rendering
type AgentStatusData struct {
	TriggerState     string `json:"trigger_state"`
	AgentRunning     bool   `json:"agent_running"`
	BufferPending    int    `json:"buffer_pending"`
	LastBufferCount  int    `json:"last_buffer_count"`
	Notification     string `json:"notification,omitempty"`
}

// Render renders agent status
func (v *AgentStatusView) Render(w http.ResponseWriter, data AgentStatusData) {
	response := map[string]interface{}{
		"success":   true,
		"data":      data,
		"timestamp": time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// NotificationView renders TUI notifications
type NotificationView struct{}

// NewNotificationView creates a new notification view
func NewNotificationView() *NotificationView {
	return &NotificationView{}
}

// NotificationData contains notification data for rendering
type NotificationData struct {
	Notifications []string `json:"notifications"`
	Count         int      `json:"count"`
	Cleared       bool     `json:"cleared,omitempty"`
}

// Render renders notifications
func (v *NotificationView) Render(w http.ResponseWriter, data NotificationData) {
	response := map[string]interface{}{
		"success":   true,
		"data":      data,
		"timestamp": time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// TriggerEventView renders trigger events
type TriggerEventView struct{}

// NewTriggerEventView creates a new trigger event view
func NewTriggerEventView() *TriggerEventView {
	return &TriggerEventView{}
}

// TriggerEventData contains trigger event data
type TriggerEventData struct {
	Event       string    `json:"event"`
	Reason      string    `json:"reason"`
	Timestamp   time.Time `json:"timestamp"`
	AgentState  string    `json:"agent_state"`
	BufferCount int       `json:"buffer_count"`
}

// Render renders a trigger event
func (v *TriggerEventView) Render(w http.ResponseWriter, data TriggerEventData) {
	response := map[string]interface{}{
		"success": true,
		"event":   data,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
