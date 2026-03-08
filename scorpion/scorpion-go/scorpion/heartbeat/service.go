// Package heartbeat provides periodic agent wake-up functionality.
package heartbeat

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"scorpion-go/scorpion/providers"

	"log/slog"
)

// heartbeatTool implements the heartbeat decision tool.
type heartbeatTool struct{}

func (h heartbeatTool) Name() string { return "heartbeat" }
func (h heartbeatTool) Description() string {
	return "Report heartbeat decision after reviewing tasks."
}
func (h heartbeatTool) Run(_ context.Context, input map[string]string) (string, error) {
	action := strings.TrimSpace(input["action"])
	tasks := strings.TrimSpace(input["tasks"])

	if action != "skip" && action != "run" {
		action = "skip"
	}

	result := map[string]string{
		"action": action,
		"tasks":  tasks,
	}

	data, err := json.Marshal(result)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// ExecuteFunc is a callback function type for executing tasks.
type ExecuteFunc func(ctx context.Context, tasks string) (string, error)

// NotifyFunc is a callback function type for notifying results.
type NotifyFunc func(ctx context.Context, response string) error

// HeartbeatService provides periodic heartbeat functionality.
type HeartbeatService struct {
	mu            sync.RWMutex
	workspace     string
	provider      providers.LLMProvider
	model         string
	onExecute     ExecuteFunc
	onNotify      NotifyFunc
	interval      time.Duration
	enabled       bool
	running       bool
	cancelFunc    context.CancelFunc
	logger        *slog.Logger
}

// HeartbeatServiceConfig holds configuration for the heartbeat service.
type HeartbeatServiceConfig struct {
	Workspace string
	Provider  providers.LLMProvider
	Model     string
	OnExecute ExecuteFunc
	OnNotify  NotifyFunc
	Interval  time.Duration
	Enabled   bool
}

// NewHeartbeatService creates a new heartbeat service.
func NewHeartbeatService(cfg HeartbeatServiceConfig) *HeartbeatService {
	if cfg.Interval <= 0 {
		cfg.Interval = 30 * time.Minute // Default 30 minutes
	}
	if cfg.Workspace == "" {
		cfg.Workspace = filepath.Join(os.Getenv("HOME"), ".scorpion-go", "workspace")
	}

	return &HeartbeatService{
		workspace:  cfg.Workspace,
		provider:   cfg.Provider,
		model:      cfg.Model,
		onExecute:  cfg.OnExecute,
		onNotify:   cfg.OnNotify,
		interval:   cfg.Interval,
		enabled:    cfg.Enabled,
		logger:     slog.Default(),
	}
}

// heartbeatFile returns the path to the heartbeat file.
func (h *HeartbeatService) heartbeatFile() string {
	return filepath.Join(h.workspace, "HEARTBEAT.md")
}

// readHeartbeatFile reads the heartbeat file content.
func (h *HeartbeatService) readHeartbeatFile() (string, error) {
	path := h.heartbeatFile()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("failed to read heartbeat file: %w", err)
	}
	return string(data), nil
}

// decide asks the LLM to decide whether to skip or run based on heartbeat content.
func (h *HeartbeatService) decide(ctx context.Context, content string) (action string, tasks string, err error) {
	// For now, use a simple heuristic since we don't have full LLM integration
	// In production, this would call the LLM with the heartbeat tool
	
	// Simple heuristic: if content contains "run" or "active", trigger execution
	contentLower := strings.ToLower(content)
	
	if strings.Contains(contentLower, "run") || 
	   strings.Contains(contentLower, "active") ||
	   strings.Contains(contentLower, "task") ||
	   strings.Contains(contentLower, "todo") {
		// Extract a summary of tasks
		lines := strings.Split(content, "\n")
		var taskLines []string
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
				taskLines = append(taskLines, trimmed)
			}
		}
		tasks = strings.Join(taskLines, " ")
		action = "run"
	} else {
		action = "skip"
	}

	return action, tasks, nil
}

// Start starts the heartbeat service.
func (h *HeartbeatService) Start() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.enabled {
		h.logger.Info("Heartbeat service disabled")
		return nil
	}

	if h.running {
		h.logger.Warn("Heartbeat service already running")
		return nil
	}

	h.running = true
	ctx, cancel := context.WithCancel(context.Background())
	h.cancelFunc = cancel

	go h.runLoop(ctx)

	h.logger.Info("Heartbeat service started", "interval", h.interval)
	return nil
}

// Stop stops the heartbeat service.
func (h *HeartbeatService) Stop() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.running {
		return nil
	}

	h.running = false
	if h.cancelFunc != nil {
		h.cancelFunc()
		h.cancelFunc = nil
	}

	h.logger.Info("Heartbeat service stopped")
	return nil
}

// runLoop runs the main heartbeat loop.
func (h *HeartbeatService) runLoop(ctx context.Context) {
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := h.tick(ctx); err != nil {
				h.logger.Error("Heartbeat tick error", "error", err)
			}
		}
	}
}

// tick executes a single heartbeat tick.
func (h *HeartbeatService) tick(ctx context.Context) error {
	content, err := h.readHeartbeatFile()
	if err != nil {
		h.logger.Debug("Heartbeat file read error", "error", err)
		return nil
	}

	if content == "" {
		h.logger.Debug("Heartbeat file empty or missing")
		return nil
	}

	h.logger.Info("Heartbeat: checking for tasks...")

	action, tasks, err := h.decide(ctx, content)
	if err != nil {
		return err
	}

	if action != "run" {
		h.logger.Info("Heartbeat: OK (nothing to report)")
		return nil
	}

	h.logger.Info("Heartbeat: tasks found, executing...")

	if h.onExecute == nil {
		return nil
	}

	response, err := h.onExecute(ctx, tasks)
	if err != nil {
		h.logger.Error("Heartbeat execution failed", "error", err)
		return err
	}

	if response != "" && h.onNotify != nil {
		h.logger.Info("Heartbeat: completed, delivering response")
		if err := h.onNotify(ctx, response); err != nil {
			h.logger.Error("Heartbeat notify error", "error", err)
		}
	}

	return nil
}

// TriggerNow manually triggers a heartbeat immediately.
func (h *HeartbeatService) TriggerNow(ctx context.Context) (string, error) {
	content, err := h.readHeartbeatFile()
	if err != nil {
		return "", err
	}

	if content == "" {
		return "", nil
	}

	action, tasks, err := h.decide(ctx, content)
	if err != nil {
		return "", err
	}

	if action != "run" || h.onExecute == nil {
		return "", nil
	}

	return h.onExecute(ctx, tasks)
}

// IsRunning returns whether the heartbeat service is running.
func (h *HeartbeatService) IsRunning() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.running
}

// SetEnabled enables or disables the heartbeat service.
func (h *HeartbeatService) SetEnabled(enabled bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.enabled = enabled
}

// IsEnabled returns whether the heartbeat service is enabled.
func (h *HeartbeatService) IsEnabled() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.enabled
}
