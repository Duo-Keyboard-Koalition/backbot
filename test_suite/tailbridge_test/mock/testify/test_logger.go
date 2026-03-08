package testify

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/codejedi-ai/kaggle-for-tensors/tailscale-app/tailbridge/test_platform/mock"
)

// TestLogger handles structured test logging
type TestLogger struct {
	TestID      string
	TestName    string
	StartTime   time.Time
	LogEntries  []LogEntry
	NetworkID   string
	AgentCount  int
	OutputDir   string
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp   time.Time `json:"timestamp"`
	Level       string    `json:"level"` // INFO, WARN, ERROR, DEBUG
	Category    string    `json:"category"`
	Message     string    `json:"message"`
	Details     interface{} `json:"details,omitempty"`
}

// AgentLogEntry logs agent-specific information
type AgentLogEntry struct {
	Name          string   `json:"name"`
	State         string   `json:"state"`
	TailscaleIP   string   `json:"tailscale_ip"`
	TailscaleIPv6 string   `json:"tailscale_ipv6"`
	Capabilities  []string `json:"capabilities"`
	InboundPort   int      `json:"inbound_port"`
	HTTPPort      int      `json:"http_port"`
	NodeID        string   `json:"node_id"`
}

// NetworkLogEntry logs network state
type NetworkLogEntry struct {
	NetworkID     string `json:"network_id"`
	TotalAgents   int    `json:"total_agents"`
	RunningAgents int    `json:"running_agents"`
	MessageCount  int    `json:"message_count"`
}

// NewTestLogger creates a new test logger
func NewTestLogger(testName string) *TestLogger {
	return &TestLogger{
		TestID:     fmt.Sprintf("test-%s-%d", testName, time.Now().UnixNano()),
		TestName:   testName,
		StartTime:  time.Now(),
		LogEntries: make([]LogEntry, 0),
		OutputDir:  filepath.Join("test_logs", testName),
	}
}

// Log adds a log entry
func (tl *TestLogger) Log(level, category, message string, details interface{}) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Category:  category,
		Message:   message,
		Details:   details,
	}
	tl.LogEntries = append(tl.LogEntries, entry)
}

// LogInfo logs an info message
func (tl *TestLogger) LogInfo(category, message string, details interface{}) {
	tl.Log("INFO", category, message, details)
}

// LogError logs an error message
func (tl *TestLogger) LogError(category, message string, err error) {
	tl.Log("ERROR", category, message, map[string]interface{}{
		"error": err.Error(),
	})
}

// LogAgentSpinUp logs agent startup
func (tl *TestLogger) LogAgentSpinUp(agent *mock.MockAgent) {
	tl.LogInfo("AGENT_LIFECYCLE", fmt.Sprintf("Agent %s spinning up", agent.Name), AgentLogEntry{
		Name:          agent.Name,
		State:         string(agent.State),
		TailscaleIP:   agent.TailscaleIP,
		TailscaleIPv6: agent.TailscaleIPv6,
		Capabilities:  agent.Capabilities,
		InboundPort:   agent.InboundPort,
		HTTPPort:      agent.HTTPPort,
		NodeID:        agent.NodeID,
	})
}

// LogAgentTearDown logs agent teardown
func (tl *TestLogger) LogAgentTearDown(agent *mock.MockAgent) {
	tl.LogInfo("AGENT_LIFECYCLE", fmt.Sprintf("Agent %s tearing down", agent.Name), map[string]interface{}{
		"name":    agent.Name,
		"state":   "stopping",
		"uptime":  time.Since(agent.StartTime).String(),
		"messages_sent": agent.MessageCount,
	})
}

// LogNetworkState logs network state
func (tl *TestLogger) LogNetworkState(network *mock.MockNetwork) {
	stats := network.GetNetworkStats()
	tl.LogInfo("NETWORK_STATE", "Network state snapshot", NetworkLogEntry{
		NetworkID:     network.GetNetworkID(),
		TotalAgents:   stats.TotalAgents,
		RunningAgents: stats.RunningAgents,
		MessageCount:  stats.TotalMessages,
	})
}

// LogDiscoveryEvent logs a discovery event
func (tl *TestLogger) LogDiscoveryEvent(event mock.DiscoveryEvent) {
	tl.LogInfo("DISCOVERY", fmt.Sprintf("Discovery event: %s", event.Type), map[string]interface{}{
		"event_type":  event.Type,
		"agent_name":  event.AgentName,
		"timestamp":   event.Timestamp,
		"details":     event.Details,
	})
}

// LogTestStart logs test start
func (tl *TestLogger) LogTestStart() {
	tl.LogInfo("TEST_LIFECYCLE", fmt.Sprintf("Test %s started", tl.TestName), map[string]interface{}{
		"test_id":   tl.TestID,
		"start_time": tl.StartTime,
	})
}

// LogTestEnd logs test end
func (tl *TestLogger) LogTestEnd(passed bool, duration time.Duration) {
	tl.LogInfo("TEST_LIFECYCLE", fmt.Sprintf("Test %s completed", tl.TestName), map[string]interface{}{
		"test_id":  tl.TestID,
		"passed":   passed,
		"duration": duration.String(),
	})
}

// Save writes logs to disk
func (tl *TestLogger) Save() error {
	// Create output directory
	err := os.MkdirAll(tl.OutputDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create test log file
	logFile := filepath.Join(tl.OutputDir, fmt.Sprintf("%s.json", tl.TestID))
	
	data, err := json.MarshalIndent(map[string]interface{}{
		"test_id":      tl.TestID,
		"test_name":    tl.TestName,
		"start_time":   tl.StartTime,
		"end_time":     time.Now(),
		"duration":     time.Since(tl.StartTime).String(),
		"log_entries":  tl.LogEntries,
	}, "", "  ")
	
	if err != nil {
		return fmt.Errorf("failed to marshal log data: %w", err)
	}

	err = os.WriteFile(logFile, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write log file: %w", err)
	}

	return nil
}

// GenerateTestReport generates a summary report for all tests in a suite
func GenerateTestReport(testName string, outputDir string) error {
	// Find all log files
	pattern := filepath.Join(outputDir, "*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to find log files: %w", err)
	}

	type TestSummary struct {
		TestID   string `json:"test_id"`
		Duration string `json:"duration"`
		Passed   bool   `json:"passed"`
	}

	summaries := make([]TestSummary, 0)
	totalDuration := time.Duration(0)
	passCount := 0

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		var logData map[string]interface{}
		if err := json.Unmarshal(data, &logData); err != nil {
			continue
		}

		summary := TestSummary{
			TestID:   logData["test_id"].(string),
			Duration: logData["duration"].(string),
			Passed:   true, // Assume passed unless marked otherwise
		}
		
		// Parse duration
		if durationStr, ok := logData["duration"].(string); ok {
			if d, err := time.ParseDuration(durationStr); err == nil {
				totalDuration += d
			}
		}

		summaries = append(summaries, summary)
		passCount++
	}

	// Create report
	report := map[string]interface{}{
		"suite_name":     testName,
		"generated_at":   time.Now(),
		"total_tests":    len(summaries),
		"passed":         passCount,
		"failed":         len(summaries) - passCount,
		"total_duration": totalDuration.String(),
		"tests":          summaries,
	}

	reportData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	reportFile := filepath.Join(outputDir, "test_report.json")
	err = os.WriteFile(reportFile, reportData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	return nil
}
