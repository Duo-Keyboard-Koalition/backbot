package views

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/codejedi-ai/kaggle-for-tensors/tailscale-app/internal/models"
)

// JSONView renders data as JSON
type JSONView struct{}

// NewJSONView creates a new JSON view
func NewJSONView() *JSONView {
	return &JSONView{}
}

// Render renders data as JSON response
func (v *JSONView) Render(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode response: %v", err), http.StatusInternalServerError)
	}
}

// Error renders an error response
func (v *JSONView) Error(w http.ResponseWriter, message string, status int) {
	v.Render(w, map[string]interface{}{
		"error":     message,
		"timestamp": time.Now().UTC(),
	}, status)
}

// Success renders a success response
func (v *JSONView) Success(w http.ResponseWriter, data interface{}) {
	v.Render(w, map[string]interface{}{
		"success":   true,
		"data":      data,
		"timestamp": time.Now().UTC(),
	}, http.StatusOK)
}

// AgentListView renders a list of agents
type AgentListView struct{}

// NewAgentListView creates a new agent list view
func NewAgentListView() *AgentListView {
	return &AgentListView{}
}

// Render renders the agent list
func (v *AgentListView) Render(w http.ResponseWriter, agents []models.AgentInfo) {
	response := map[string]interface{}{
		"agents": agents,
		"count":  len(agents),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode response: %v", err), http.StatusInternalServerError)
	}
}

// EnvelopeView renders envelope-related responses
type EnvelopeView struct{}

// NewEnvelopeView creates a new envelope view
func NewEnvelopeView() *EnvelopeView {
	return &EnvelopeView{}
}

// Render renders an envelope response
func (v *EnvelopeView) Render(w http.ResponseWriter, env models.Envelope, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(env); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode envelope: %v", err), http.StatusInternalServerError)
	}
}

// CommandResponseView renders command execution responses
type CommandResponseView struct{}

// NewCommandResponseView creates a new command response view
func NewCommandResponseView() *CommandResponseView {
	return &CommandResponseView{}
}

// Render renders a command response
func (v *CommandResponseView) Render(w http.ResponseWriter, resp models.Response) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode response: %v", err), http.StatusInternalServerError)
	}
}

// StatusView renders status information
type StatusView struct{}

// NewStatusView creates a new status view
func NewStatusView() *StatusView {
	return &StatusView{}
}

// Render renders status information
func (v *StatusView) Render(w http.ResponseWriter, status map[string]interface{}) {
	response := map[string]interface{}{
		"status":    "ok",
		"data":      status,
		"timestamp": time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode response: %v", err), http.StatusInternalServerError)
	}
}
