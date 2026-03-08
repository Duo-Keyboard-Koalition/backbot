package models

import (
	"fmt"
	"log"
	"os"
	"sync"
)

// ConsoleTUINotifier implements TUINotifier for console-based TUI
type ConsoleTUINotifier struct {
	mu            sync.Mutex
	bufferStatus  int
	currentStatus string
	notifications []string
	maxNotifications int
}

// ConsoleTUINotifierConfig configures the console TUI notifier
type ConsoleTUINotifierConfig struct {
	MaxNotifications int
}

// DefaultConsoleTUINotifierConfig returns default configuration
func DefaultConsoleTUINotifierConfig() *ConsoleTUINotifierConfig {
	return &ConsoleTUINotifierConfig{
		MaxNotifications: 10,
	}
}

// NewConsoleTUINotifier creates a new console TUI notifier
func NewConsoleTUINotifier(config *ConsoleTUINotifierConfig) *ConsoleTUINotifier {
	if config == nil {
		config = DefaultConsoleTUINotifierConfig()
	}

	return &ConsoleTUINotifier{
		notifications: make([]string, 0),
		maxNotifications: config.MaxNotifications,
	}
}

// NotifyAgentTriggered notifies that the agent has been triggered
func (n *ConsoleTUINotifier) NotifyAgentTriggered(reason string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	msg := fmt.Sprintf("🚀 Agent triggered: %s", reason)
	n.addNotification(msg)
	log.Printf("[tui] %s", msg)
	
	// Print to stderr to not interfere with normal output
	fmt.Fprintf(os.Stderr, "\n%s\n", msg)
	
	return nil
}

// NotifyAgentStopped notifies that the agent has stopped
func (n *ConsoleTUINotifier) NotifyAgentStopped(reason string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	msg := fmt.Sprintf("🛑 Agent stopped: %s", reason)
	n.addNotification(msg)
	log.Printf("[tui] %s", msg)
	
	fmt.Fprintf(os.Stderr, "\n%s\n", msg)
	
	return nil
}

// UpdateBufferStatus updates the buffer status display
func (n *ConsoleTUINotifier) UpdateBufferStatus(pending int, status string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.bufferStatus = pending
	n.currentStatus = status

	// Log status changes
	log.Printf("[tui] buffer: %d pending, status: %s", pending, status)
	
	return nil
}

// GetStatus returns the current TUI status
func (n *ConsoleTUINotifier) GetStatus() map[string]interface{} {
	n.mu.Lock()
	defer n.mu.Unlock()

	return map[string]interface{}{
		"buffer_pending": n.bufferStatus,
		"status":         n.currentStatus,
		"notifications":  n.getRecentNotifications(),
	}
}

// GetNotifications returns recent notifications
func (n *ConsoleTUINotifier) GetNotifications() []string {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.getRecentNotifications()
}

// addNotification adds a notification to the list
func (n *ConsoleTUINotifier) addNotification(msg string) {
	n.notifications = append(n.notifications, msg)
	
	// Trim old notifications if exceeding max
	if len(n.notifications) > n.maxNotifications {
		n.notifications = n.notifications[len(n.notifications)-n.maxNotifications:]
	}
}

// getRecentNotifications returns recent notifications (caller must hold lock)
func (n *ConsoleTUINotifier) getRecentNotifications() []string {
	if len(n.notifications) > n.maxNotifications {
		return n.notifications[len(n.notifications)-n.maxNotifications:]
	}
	return n.notifications
}

// ClearNotifications clears all notifications
func (n *ConsoleTUINotifier) ClearNotifications() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.notifications = make([]string, 0)
}

// SimpleTUINotifier is a simpler implementation that just logs
type SimpleTUINotifier struct{}

// NewSimpleTUINotifier creates a simple TUI notifier
func NewSimpleTUINotifier() *SimpleTUINotifier {
	return &SimpleTUINotifier{}
}

// NotifyAgentTriggered logs the trigger event
func (n *SimpleTUINotifier) NotifyAgentTriggered(reason string) error {
	log.Printf("[tui] Agent triggered: %s", reason)
	return nil
}

// NotifyAgentStopped logs the stop event
func (n *SimpleTUINotifier) NotifyAgentStopped(reason string) error {
	log.Printf("[tui] Agent stopped: %s", reason)
	return nil
}

// UpdateBufferStatus logs buffer status updates
func (n *SimpleTUINotifier) UpdateBufferStatus(pending int, status string) error {
	log.Printf("[tui] Buffer: %d pending (%s)", pending, status)
	return nil
}
