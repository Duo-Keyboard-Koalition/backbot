package models

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// AgentTriggerState represents the current state of the agent trigger
type AgentTriggerState string

const (
	// TriggerStateIdle - buffer is empty, agent not running
	TriggerStateIdle AgentTriggerState = "idle"
	// TriggerStateActive - buffer has messages, agent is running
	TriggerStateActive AgentTriggerState = "active"
	// TriggerStateStopping - buffer draining, agent stopping
	TriggerStateStopping AgentTriggerState = "stopping"
)

// TriggerEvent represents events that trigger state changes
type TriggerEvent string

const (
	// EventBufferEmpty - buffer transitioned to empty (0 messages)
	EventBufferEmpty TriggerEvent = "buffer_empty"
	// EventBufferHasMessages - buffer transitioned to non-empty (1+ messages)
	EventBufferHasMessages TriggerEvent = "buffer_has_messages"
	// EventAgentStopped - agent has stopped
	EventAgentStopped TriggerEvent = "agent_stopped"
	// EventAgentStarted - agent has started
	EventAgentStarted TriggerEvent = "agent_started"
)

// BufferMonitor defines the interface for monitoring buffer state
type BufferMonitor interface {
	GetPendingCount() (int, error)
	IsRunning() bool
}

// TUINotifier defines the interface for notifying the TUI
type TUINotifier interface {
	NotifyAgentTriggered(reason string) error
	NotifyAgentStopped(reason string) error
	UpdateBufferStatus(pending int, status string) error
}

// AgentTriggerService manages agent triggering based on buffer state
type AgentTriggerService struct {
	// monitor provides access to buffer state
	monitor BufferMonitor

	// notifier sends notifications to the TUI
	notifier TUINotifier

	// startAgentFunc is called to start the agent
	startAgentFunc func(ctx context.Context) error

	// stopAgentFunc is called to stop the agent
	stopAgentFunc func() error

	// mu protects concurrent access
	mu sync.RWMutex

	// state is the current trigger state
	state AgentTriggerState

	// agentRunning indicates if the agent is currently running
	agentRunning bool

	// agentCtx is the context for the running agent
	agentCtx context.Context

	// agentCancel is the cancel function for the agent context
	agentCancel context.CancelFunc

	// stopChan signals the monitor loop to stop
	stopChan chan struct{}

	// checkInterval controls how often to check buffer state
	checkInterval time.Duration

	// lastBufferCount tracks the previous buffer count for edge detection
	lastBufferCount int
}

// AgentTriggerServiceConfig configures the trigger service
type AgentTriggerServiceConfig struct {
	// CheckInterval controls how often to check buffer state
	CheckInterval time.Duration
}

// DefaultAgentTriggerServiceConfig returns default configuration
func DefaultAgentTriggerServiceConfig() *AgentTriggerServiceConfig {
	return &AgentTriggerServiceConfig{
		CheckInterval: 1 * time.Second,
	}
}

// NewAgentTriggerService creates a new agent trigger service
func NewAgentTriggerService(
	monitor BufferMonitor,
	notifier TUINotifier,
	startAgentFunc func(ctx context.Context) error,
	stopAgentFunc func() error,
	config *AgentTriggerServiceConfig,
) (*AgentTriggerService, error) {
	if config == nil {
		config = DefaultAgentTriggerServiceConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &AgentTriggerService{
		monitor:        monitor,
		notifier:       notifier,
		startAgentFunc: startAgentFunc,
		stopAgentFunc:  stopAgentFunc,
		state:          TriggerStateIdle,
		agentCtx:       ctx,
		agentCancel:    cancel,
		stopChan:       make(chan struct{}),
		checkInterval:  config.CheckInterval,
		lastBufferCount: 0,
	}, nil
}

// Start begins monitoring the buffer and triggering the agent
func (s *AgentTriggerService) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("[trigger] starting agent trigger service (interval=%v)", s.checkInterval)
	go s.monitorLoop()
}

// Stop halts the trigger service
func (s *AgentTriggerService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("[trigger] stopping agent trigger service")
	close(s.stopChan)

	// Stop agent if running
	if s.agentRunning {
		s.stopAgent()
	}
}

// monitorLoop continuously monitors buffer state
func (s *AgentTriggerService) monitorLoop() {
	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.checkBufferAndTrigger()
		}
	}
}

// checkBufferAndTrigger checks buffer state and triggers agent if needed
func (s *AgentTriggerService) checkBufferAndTrigger() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get current buffer count
	pendingCount, err := s.monitor.GetPendingCount()
	if err != nil {
		log.Printf("[trigger] failed to get pending count: %v", err)
		pendingCount = 0
	}

	// Update TUI with current buffer status
	if s.notifier != nil {
		status := s.getStateString()
		s.notifier.UpdateBufferStatus(pendingCount, status)
	}

	// Detect edge: 0 -> 1+ (buffer was empty, now has messages)
	if s.lastBufferCount == 0 && pendingCount > 0 {
		log.Printf("[trigger] buffer transition: empty -> has messages (%d)", pendingCount)
		s.triggerAgent()
	}

	// Detect edge: 1+ -> 0 (buffer was non-empty, now empty)
	if s.lastBufferCount > 0 && pendingCount == 0 {
		log.Printf("[trigger] buffer transition: has messages -> empty")
		s.stopAgentIfRunning()
	}

	s.lastBufferCount = pendingCount
}

// triggerAgent starts the agent if not already running
func (s *AgentTriggerService) triggerAgent() {
	if s.agentRunning {
		log.Printf("[trigger] agent already running, skipping trigger")
		return
	}

	log.Printf("[trigger] triggering agent start")

	// Create new context for agent
	ctx, cancel := context.WithCancel(context.Background())
	s.agentCtx = ctx
	s.agentCancel = cancel

	// Notify TUI
	if s.notifier != nil {
		s.notifier.NotifyAgentTriggered("Buffer has messages to process")
	}

	// Start agent in background
	go func() {
		if err := s.startAgentFunc(ctx); err != nil {
			log.Printf("[trigger] agent error: %v", err)
		}
		s.onAgentStopped()
	}()

	s.agentRunning = true
	s.state = TriggerStateActive
}

// stopAgentIfRunning stops the agent if it's running
func (s *AgentTriggerService) stopAgentIfRunning() {
	if !s.agentRunning {
		return
	}

	s.state = TriggerStateStopping
	log.Printf("[trigger] stopping agent (buffer empty)")

	if s.notifier != nil {
		s.notifier.NotifyAgentStopped("Buffer is empty")
	}

	s.stopAgent()
}

// stopAgent calls the stop function
func (s *AgentTriggerService) stopAgent() {
	if s.stopAgentFunc != nil {
		if err := s.stopAgentFunc(); err != nil {
			log.Printf("[trigger] error stopping agent: %v", err)
		}
	}

	// Cancel context
	if s.agentCancel != nil {
		s.agentCancel()
	}

	s.agentRunning = false
	s.state = TriggerStateIdle
}

// onAgentStopped is called when the agent stops
func (s *AgentTriggerService) onAgentStopped() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.agentRunning = false
	s.state = TriggerStateIdle

	log.Printf("[trigger] agent stopped")

	if s.notifier != nil {
		s.notifier.NotifyAgentStopped("Agent completed processing")
	}
}

// GetState returns the current trigger state
func (s *AgentTriggerService) GetState() AgentTriggerState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

// IsAgentRunning returns whether the agent is currently running
func (s *AgentTriggerService) IsAgentRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.agentRunning
}

// GetAgentContext returns the context for the running agent
func (s *AgentTriggerService) GetAgentContext() context.Context {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.agentCtx
}

// ManualTrigger manually triggers the agent
func (s *AgentTriggerService) ManualTrigger() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.agentRunning {
		return fmt.Errorf("agent already running")
	}

	log.Printf("[trigger] manual trigger requested")
	s.triggerAgent()
	return nil
}

// ManualStop manually stops the agent
func (s *AgentTriggerService) ManualStop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.agentRunning {
		return fmt.Errorf("agent not running")
	}

	log.Printf("[trigger] manual stop requested")
	s.stopAgentIfRunning()
	return nil
}

// GetStatus returns the current status as a map
func (s *AgentTriggerService) GetStatus() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"state":          string(s.state),
		"agent_running":  s.agentRunning,
		"last_buffer_count": s.lastBufferCount,
		"check_interval": s.checkInterval.String(),
	}
}

func (s *AgentTriggerService) getStateString() string {
	switch s.state {
	case TriggerStateIdle:
		return "idle"
	case TriggerStateActive:
		return "active"
	case TriggerStateStopping:
		return "stopping"
	default:
		return "unknown"
	}
}
