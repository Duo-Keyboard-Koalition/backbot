package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/codejedi-ai/kaggle-for-tensors/tailscale-app/internal/models"
)

// TriggerController handles agent trigger lifecycle management
type TriggerController struct {
	triggerSvc *models.AgentTriggerService
	notifier   *models.ConsoleTUINotifier
}

// NewTriggerController creates a new trigger controller
func NewTriggerController(
	triggerSvc *models.AgentTriggerService,
	notifier *models.ConsoleTUINotifier,
) *TriggerController {
	return &TriggerController{
		triggerSvc: triggerSvc,
		notifier:   notifier,
	}
}

// HandleTriggerStatus returns the current trigger status
func (c *TriggerController) HandleTriggerStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := c.triggerSvc.GetStatus()
	tuiStatus := c.notifier.GetStatus()

	response := map[string]interface{}{
		"trigger": status,
		"tui":     tuiStatus,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// HandleTriggerManual triggers the agent manually
func (c *TriggerController) HandleTriggerManual(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := c.triggerSvc.ManualTrigger(); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Agent triggered manually",
	})
}

// HandleTriggerStop stops the agent manually
func (c *TriggerController) HandleTriggerStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := c.triggerSvc.ManualStop(); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Agent stopped manually",
	})
}

// HandleNotifications returns recent TUI notifications
func (c *TriggerController) HandleNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	notifications := c.notifier.GetNotifications()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"notifications": notifications,
		"count":         len(notifications),
	})
}

// HandleClearNotifications clears all notifications
func (c *TriggerController) HandleClearNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	c.notifier.ClearNotifications()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Notifications cleared",
	})
}
