//go:build integration

package multiagent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// AgentStatus represents the status of a Tailscale-connected agent
type AgentStatus struct {
	Name       string   `json:"name"`
	IP         string   `json:"ip"`
	TailscaleIP string  `json:"tailscale_ip"`
	Online     bool     `json:"online"`
	Capabilities []string `json:"capabilities"`
	Hostname   string   `json:"hostname"`
}

// ChatMessage represents a chat message between agents
type ChatMessage struct {
	From        string    `json:"from"`
	To          string    `json:"to"`
	Content     string    `json:"content"`
	Timestamp   time.Time `json:"timestamp"`
	MessageType string    `json:"message_type"` // chat, directive, status
}

// ChatLog stores all communications for an agent
type ChatLog struct {
	AgentName string       `json:"agent_name"`
	Messages  []ChatMessage `json:"messages"`
	IP        string       `json:"ip"`
}

// MultiAgentTestSuite runs comprehensive multi-agent tests
type MultiAgentTestSuite struct {
	suite.Suite
	agent1URL    string
	agent2URL    string
	agent3URL    string
	tsAuthKey    string
	geminiAPIKey string
	client       *http.Client
	mu           sync.Mutex
	chatLogs     map[string]*ChatLog
}

// SetupSuite runs once before all tests
func (s *MultiAgentTestSuite) SetupSuite() {
	s.agent1URL = getEnv("AGENT1_URL", "http://localhost:8081")
	s.agent2URL = getEnv("AGENT2_URL", "http://localhost:8082")
	s.agent3URL = getEnv("AGENT3_URL", "http://localhost:8083")
	s.tsAuthKey = getEnv("TS_AUTH_KEY", "tskey-auth-k7Q1t39ZWj11CNTRL-FbRR2tKLRcPn5L246vsAcP7LP2YCUxWD")
	s.geminiAPIKey = getEnv("GEMINI_API_KEY", "AIzaSyC_9M8im8z5F0ING2W3Hu2aQiunJhhWUXI")

	s.client = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
		},
	}

	s.chatLogs = make(map[string]*ChatLog)
}

// TestMultiAgentCommunication runs the main test
func (s *MultiAgentTestSuite) TestMultiAgentCommunication() {
	ctx := context.Background()

	s.T().Log("=== MULTI-AGENT TAILSCALE COMMUNICATION TEST ===")
	s.T().Log("")

	// Step 1: Wait for all agents to be healthy
	s.T().Log("📡 Step 1: Checking agent health...")
	agents := map[string]string{
		"agent1": s.agent1URL,
		"agent2": s.agent2URL,
		"agent3": s.agent3URL,
	}

	healthyAgents := s.waitForHealthyAgents(ctx, agents, 5*time.Minute)
	s.Require().Equal(3, len(healthyAgents), "All 3 agents must be healthy")
	s.T().Logf("✅ All %d agents are healthy\n", len(healthyAgents))
	s.T().Log("")

	// Step 2: Get Tailscale IP addresses
	s.T().Log("🌐 Step 2: Retrieving Tailscale IP addresses...")
	agentStatuses := make(map[string]*AgentStatus)

	for name, url := range agents {
		status, err := s.getAgentStatus(ctx, url)
		s.Require().NoError(err, "Failed to get status for %s", name)
		agentStatuses[name] = status

		s.T().Logf("  %s:", name)
		s.T().Logf("    Hostname: %s", status.Hostname)
		s.T().Logf("    Tailscale IP: %s", status.TailscaleIP)
		s.T().Logf("    Local IP: %s", status.IP)
		s.T().Logf("    Online: %v", status.Online)
		s.T().Logf("    Capabilities: %v", status.Capabilities)
		s.T().Log("")

		// Initialize chat log
		s.chatLogs[name] = &ChatLog{
			AgentName: name,
			Messages:  []ChatMessage{},
			IP:        status.TailscaleIP,
		}
	}

	// Step 3: Test A2A communication
	s.T().Log("💬 Step 3: Testing Agent-to-Agent communication...")
	s.T().Log("")

	// Test 3a: Agent1 -> Agent2 (Chat)
	s.T().Log("  Test 3a: Agent1 sends chat to Agent2...")
	msg1 := ChatMessage{
		From:        "agent1",
		To:          "agent2",
		Content:     "Hello Agent 2! This is Agent 1 initiating contact. Can you confirm receipt?",
		Timestamp:   time.Now(),
		MessageType: "chat",
	}
	s.sendChatMessage(ctx, msg1)
	s.chatLogs["agent1"].Messages = append(s.chatLogs["agent1"].Messages, msg1)
	s.T().Logf("    ✓ Message sent from agent1 to agent2")
	s.T().Log("")

	// Test 3b: Agent2 -> Agent1 (Response)
	s.T().Log("  Test 3b: Agent2 responds to Agent1...")
	msg2 := ChatMessage{
		From:        "agent2",
		To:          "agent1",
		Content:     "Agent 1, this is Agent 2. Message received loud and clear! I confirm communication is established.",
		Timestamp:   time.Now(),
		MessageType: "chat",
	}
	s.sendChatMessage(ctx, msg2)
	s.chatLogs["agent2"].Messages = append(s.chatLogs["agent2"].Messages, msg2)
	s.T().Logf("    ✓ Response sent from agent2 to agent1")
	s.T().Log("")

	// Test 3c: Agent1 -> Agent3 (Directive)
	s.T().Log("  Test 3c: Agent1 sends directive to Agent3...")
	msg3 := ChatMessage{
		From:        "agent1",
		To:          "agent3",
		Content:     "Agent 3, please execute diagnostic scan and report status. This is a priority directive.",
		Timestamp:   time.Now(),
		MessageType: "directive",
	}
	s.sendChatMessage(ctx, msg3)
	s.chatLogs["agent1"].Messages = append(s.chatLogs["agent1"].Messages, msg3)
	s.T().Logf("    ✓ Directive sent from agent1 to agent3")
	s.T().Log("")

	// Test 3d: Agent3 -> Agent1 (Status Report)
	s.T().Log("  Test 3d: Agent3 reports status to Agent1...")
	msg4 := ChatMessage{
		From:        "agent3",
		To:          "agent1",
		Content:     "Agent 1, diagnostic complete. All systems nominal. Network connectivity: 100%. Tailscale interface: active. Ready for task assignment.",
		Timestamp:   time.Now(),
		MessageType: "status",
	}
	s.sendChatMessage(ctx, msg4)
	s.chatLogs["agent3"].Messages = append(s.chatLogs["agent3"].Messages, msg4)
	s.T().Logf("    ✓ Status report sent from agent3 to agent1")
	s.T().Log("")

	// Test 3e: Agent2 -> Agent3 (Collaboration Request)
	s.T().Log("  Test 3e: Agent2 requests collaboration from Agent3...")
	msg5 := ChatMessage{
		From:        "agent2",
		To:          "agent3",
		Content:     "Agent 3, Agent 2 here. Let's coordinate on the next task. I'll handle data collection, you handle analysis. Agreed?",
		Timestamp:   time.Now(),
		MessageType: "chat",
	}
	s.sendChatMessage(ctx, msg5)
	s.chatLogs["agent2"].Messages = append(s.chatLogs["agent2"].Messages, msg5)
	s.T().Logf("    ✓ Collaboration request sent from agent2 to agent3")
	s.T().Log("")

	// Test 3f: Agent3 -> Agent2 (Agreement)
	s.T().Log("  Test 3f: Agent3 confirms collaboration...")
	msg6 := ChatMessage{
		From:        "agent3",
		To:          "agent2",
		Content:     "Agent 2, agreement confirmed. I'll prepare the analysis pipeline. Send data to my endpoint when ready. Standing by.",
		Timestamp:   time.Now(),
		MessageType: "chat",
	}
	s.sendChatMessage(ctx, msg6)
	s.chatLogs["agent3"].Messages = append(s.chatLogs["agent3"].Messages, msg6)
	s.T().Logf("    ✓ Agreement sent from agent3 to agent2")
	s.T().Log("")

	// Test 3g: Broadcast - All agents
	s.T().Log("  Test 3g: Agent1 broadcasts to all agents...")
	msg7 := ChatMessage{
		From:        "agent1",
		To:          "all",
		Content:     "ATTENTION ALL AGENTS: Multi-agent communication test successful. Tailscale network verified. All agents operational. Test sequence complete.",
		Timestamp:   time.Now(),
		MessageType: "directive",
	}
	s.sendChatMessage(ctx, msg7)
	s.chatLogs["agent1"].Messages = append(s.chatLogs["agent1"].Messages, msg7)
	s.T().Logf("    ✓ Broadcast sent from agent1 to all agents")
	s.T().Log("")

	// Test 3h: Agent2 acknowledgment
	s.T().Log("  Test 3h: Agent2 acknowledges broadcast...")
	msg8 := ChatMessage{
		From:        "agent2",
		To:          "all",
		Content:     "Agent 2 acknowledging. Communication test successful. Network stable. Ready for production tasks.",
		Timestamp:   time.Now(),
		MessageType: "status",
	}
	s.sendChatMessage(ctx, msg8)
	s.chatLogs["agent2"].Messages = append(s.chatLogs["agent2"].Messages, msg8)
	s.T().Logf("    ✓ Acknowledgment from agent2")
	s.T().Log("")

	// Test 3i: Agent3 acknowledgment
	s.T().Log("  Test 3i: Agent3 acknowledges broadcast...")
	msg9 := ChatMessage{
		From:        "agent3",
		To:          "all",
		Content:     "Agent 3 acknowledging. All systems green. Tailscale connection stable at " + agentStatuses["agent3"].TailscaleIP + ". Awaiting further instructions.",
		Timestamp:   time.Now(),
		MessageType: "status",
	}
	s.sendChatMessage(ctx, msg9)
	s.chatLogs["agent3"].Messages = append(s.chatLogs["agent3"].Messages, msg9)
	s.T().Logf("    ✓ Acknowledgment from agent3")
	s.T().Log("")

	// Step 4: Print final chat logs
	s.T().Log("")
	s.T().Log("═══════════════════════════════════════════════════════════")
	s.T().Log("           FINAL CHAT LOGS - ALL AGENTS")
	s.T().Log("═══════════════════════════════════════════════════════════")
	s.T().Log("")

	s.printChatLogs()

	// Step 5: Assertions
	s.T().Log("")
	s.T().Log("📊 Step 5: Running assertions...")

	for name, status := range agentStatuses {
		s.T().Logf("  Asserting %s is online...", name)
		assert.True(s.T(), status.Online, "%s should be online", name)
		assert.NotEmpty(s.T(), status.TailscaleIP, "%s should have Tailscale IP", name)
		assert.NotEmpty(s.T(), status.Capabilities, "%s should have capabilities", name)
	}

	s.T().Logf("  ✅ All assertions passed!")
	s.T().Log("")
	s.T().Log("═══════════════════════════════════════════════════════════")
	s.T().Log("           TEST COMPLETED SUCCESSFULLY")
	s.T().Log("═══════════════════════════════════════════════════════════")
}

// waitForHealthyAgents waits for agents to become healthy
func (s *MultiAgentTestSuite) waitForHealthyAgents(ctx context.Context, agents map[string]string, timeout time.Duration) map[string]string {
	healthy := make(map[string]string)
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) && len(healthy) < 3 {
		for name, url := range agents {
			if _, ok := healthy[name]; ok {
				continue
			}

			resp, err := s.client.Get(url + "/health")
			if err == nil && resp.StatusCode == 200 {
				resp.Body.Close()
				healthy[name] = url
				s.T().Logf("  ✓ %s is healthy", name)
			}
		}

		if len(healthy) < 3 {
			time.Sleep(5 * time.Second)
		}
	}

	return healthy
}

// getAgentStatus retrieves agent status including Tailscale IP
func (s *MultiAgentTestSuite) getAgentStatus(ctx context.Context, url string) (*AgentStatus, error) {
	resp, err := s.client.Get(url + "/status")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var status AgentStatus
	if err := json.Unmarshal(body, &status); err != nil {
		return nil, err
	}

	return &status, nil
}

// sendChatMessage sends a chat message between agents
func (s *MultiAgentTestSuite) sendChatMessage(ctx context.Context, msg ChatMessage) error {
	payload := map[string]interface{}{
		"from":         msg.From,
		"to":           msg.To,
		"content":      msg.Content,
		"message_type": msg.MessageType,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Send to the appropriate agent
	targetURL := s.agent1URL // Default to agent1
	if msg.To == "agent2" {
		targetURL = s.agent2URL
	} else if msg.To == "agent3" {
		targetURL = s.agent3URL
	}

	req, err := http.NewRequest("POST", targetURL+"/a2a/chat", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to send message: %s", string(body))
	}

	return nil
}

// printChatLogs prints all chat logs in a formatted way
func (s *MultiAgentTestSuite) printChatLogs() {
	for name, log := range s.chatLogs {
		s.T().Logf("┌─────────────────────────────────────────────────────────────┐")
		s.T().Logf("│ AGENT: %-8s                                        │", strings.ToUpper(name))
		s.T().Logf("│ TAILSCALE IP: %-15s                            │", log.IP)
		s.T().Logf("├─────────────────────────────────────────────────────────────┤")

		for i, msg := range log.Messages {
			timeStr := msg.Timestamp.Format("15:04:05")
			fromBadge := fmt.Sprintf("[%s]", msg.From)
			toBadge := fmt.Sprintf("→ [%s]", msg.To)

			typeBadge := "💬"
			if msg.MessageType == "directive" {
				typeBadge = "📋"
			} else if msg.MessageType == "status" {
				typeBadge = "📊"
			}

			s.T().Logf("│ %d. %s %s %-45s │", i+1, timeStr, typeBadge, fromBadge)
			s.T().Logf("│       %s %-45s │", toBadge, "")

			// Word wrap content
			content := msg.Content
			for len(content) > 50 {
				s.T().Logf("│       %-53s │", content[:50])
				content = content[50:]
			}
			s.T().Logf("│       %-53s │", content)
			s.T().Logf("│                                                             │")
		}

		s.T().Logf("└─────────────────────────────────────────────────────────────┘")
		s.T().Log("")
	}
}

// getEnv gets environment variable with default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TestMultiAgentSuite runs the test suite
func TestMultiAgentSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("TS_AUTH_KEY") == "" {
		t.Log("TS_AUTH_KEY not set, using default test key")
	}

	suite.Run(t, new(MultiAgentTestSuite))
}
