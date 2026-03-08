//go:build integration

package a2a

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// A2AIntegrationTestSuite provides integration tests for A2A communication
type A2AIntegrationTestSuite struct {
	suite.Suite
	agent1URL string
	agent2URL string
	agent3URL string
	client    *http.Client
}

// Message represents an A2A message
type Message struct {
	ID            string      `json:"id"`
	Type          string      `json:"type"`
	Source        string      `json:"source"`
	Dest          string      `json:"dest"`
	Topic         string      `json:"topic"`
	Timestamp     time.Time   `json:"timestamp"`
	Body          MessageBody `json:"body"`
	CorrelationID string      `json:"correlation_id,omitempty"`
}

// MessageBody contains the message payload
type MessageBody struct {
	Action      string          `json:"action"`
	ContentType string          `json:"content_type"`
	Payload     json.RawMessage `json:"payload"`
}

// SetupSuite runs once before all tests
func (s *A2AIntegrationTestSuite) SetupSuite() {
	s.agent1URL = getEnv("AGENT1_URL", "http://localhost:8081")
	s.agent2URL = getEnv("AGENT2_URL", "http://localhost:8082")
	s.agent3URL = getEnv("AGENT3_URL", "http://localhost:8083")

	s.client = &http.Client{
		Timeout: 30 * time.Second,
	}
}

// SetupTest runs before each test
func (s *A2AIntegrationTestSuite) SetupTest() {
	// Wait for agents to be ready
	s.waitForAgents()
}

// waitForAgents waits for all agents to be healthy
func (s *A2AIntegrationTestSuite) waitForAgents() {
	agents := []string{s.agent1URL, s.agent2URL, s.agent3URL}
	
	for _, url := range agents {
		s.Require().Eventually(func() bool {
			resp, err := s.client.Get(url + "/health")
			if err != nil {
				return false
			}
			defer resp.Body.Close()
			return resp.StatusCode == http.StatusOK
		}, 2*time.Minute, 5*time.Second, "Agent %s did not become ready", url)
	}
}

// TestAgentHealth checks all agents are healthy
func (s *A2AIntegrationTestSuite) TestAgentHealth() {
	agents := map[string]string{
		"agent1": s.agent1URL,
		"agent2": s.agent2URL,
		"agent3": s.agent3URL,
	}

	for name, url := range agents {
		resp, err := s.client.Get(url + "/health")
		s.Require().NoError(err, "Failed to connect to %s", name)
		defer resp.Body.Close()

		s.Assert().Equal(http.StatusOK, resp.StatusCode, "%s health check failed", name)

		var health struct {
			Status string `json:"status"`
			Agent  string `json:"agent"`
		}
		json.NewDecoder(resp.Body).Decode(&health)
		s.Assert().Equal("healthy", health.Status, "%s status not healthy", name)
	}
}

// TestAgentStatus checks agent status endpoints
func (s *A2AIntegrationTestSuite) TestAgentStatus() {
	resp, err := s.client.Get(s.agent1URL + "/status")
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Assert().Equal(http.StatusOK, resp.StatusCode)

	var status map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&status)

	s.Assert().Equal("online", status["status"])
	s.Assert().NotEmpty(status["name"])
	s.Assert().NotEmpty(status["capabilities"])
}

// TestPhoneBookDiscovery tests agent discovery via phone book
func (s *A2AIntegrationTestSuite) TestPhoneBookDiscovery() {
	// Get phone book from agent1
	resp, err := s.client.Get(s.agent1URL + "/phonebook")
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Assert().Equal(http.StatusOK, resp.StatusCode)

	var phonebook struct {
		Agents []map[string]interface{} `json:"agents"`
		Count  int                      `json:"count"`
	}
	json.NewDecoder(resp.Body).Decode(&phonebook)

	s.Assert().Greater(phonebook.Count, 0, "Phone book should not be empty")
}

// TestAgentListing tests filtering agents by capability
func (s *A2AIntegrationTestSuite) TestAgentListing() {
	// Get agents with file_send capability
	resp, err := s.client.Get(s.agent1URL + "/agents?capability=file_send")
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Assert().Equal(http.StatusOK, resp.StatusCode)

	var result struct {
		Agents []map[string]interface{} `json:"agents"`
		Count  int                      `json:"count"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	// Should have at least one agent with file_send capability
	s.Assert().Greater(result.Count, 0)
}

// TestSendMessage tests sending a message between agents
func (s *A2AIntegrationTestSuite) TestSendMessage() {
	payload, _ := json.Marshal(map[string]string{
		"message": "Hello from integration test!",
		"test":    "TestSendMessage",
	})

	msg := Message{
		ID:        uuid.New().String(),
		Type:      "request",
		Source:    "test-client",
		Dest:      "agent2",
		Topic:     "agent.requests",
		Timestamp: time.Now(),
		Body: MessageBody{
			Action:      "chat",
			ContentType: "application/json",
			Payload:     payload,
		},
	}

	// Send message to agent2
	resp, err := s.postJSON(s.agent2URL+"/a2a/inbound", msg)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Assert().Equal(http.StatusAccepted, resp.StatusCode)
}

// TestBidirectionalMessaging tests two-way communication
func (s *A2AIntegrationTestSuite) TestBidirectionalMessaging() {
	// Send ping from agent1 to agent2
	pingPayload, _ := json.Marshal(map[string]string{
		"type": "ping",
		"data": "test",
	})

	pingMsg := Message{
		ID:        uuid.New().String(),
		Type:      "request",
		Source:    "agent1",
		Dest:      "agent2",
		Timestamp: time.Now(),
		Body: MessageBody{
			Action:  "ping",
			Payload: pingPayload,
		},
	}

	resp, err := s.postJSON(s.agent2URL+"/a2a/inbound", pingMsg)
	s.Require().NoError(err)
	resp.Body.Close()

	// Send pong back
	pongPayload, _ := json.Marshal(map[string]string{
		"type": "pong",
		"data": "response",
	})

	pongMsg := Message{
		ID:        uuid.New().String(),
		Type:      "response",
		Source:    "agent2",
		Dest:      "agent1",
		Timestamp: time.Now(),
		Body: MessageBody{
			Action:  "pong",
			Payload: pongPayload,
		},
	}

	resp, err = s.postJSON(s.agent1URL+"/a2a/inbound", pongMsg)
	s.Require().NoError(err)
	resp.Body.Close()

	// Verify messages were received
	s.verifyMessageCount(s.agent1URL, 1)
	s.verifyMessageCount(s.agent2URL, 1)
}

// TestMessageCorrelation tests request/response correlation
func (s *A2AIntegrationTestSuite) TestMessageCorrelation() {
	correlationID := uuid.New().String()

	// Send request
	requestPayload, _ := json.Marshal(map[string]string{
		"command": "execute_test",
	})

	requestMsg := Message{
		ID:            uuid.New().String(),
		Type:          "request",
		CorrelationID: correlationID,
		Source:        "agent1",
		Dest:          "agent3",
		Timestamp:     time.Now(),
		Body: MessageBody{
			Action:  "execute",
			Payload: requestPayload,
		},
	}

	resp, err := s.postJSON(s.agent3URL+"/a2a/inbound", requestMsg)
	s.Require().NoError(err)
	resp.Body.Close()

	// Send response with same correlation ID
	responsePayload, _ := json.Marshal(map[string]string{
		"result": "success",
	})

	responseMsg := Message{
		ID:            uuid.New().String(),
		Type:          "response",
		CorrelationID: correlationID,
		Source:        "agent3",
		Dest:          "agent1",
		Timestamp:     time.Now(),
		Body: MessageBody{
			Action:  "result",
			Payload: responsePayload,
		},
	}

	resp, err = s.postJSON(s.agent1URL+"/a2a/inbound", responseMsg)
	s.Require().NoError(err)
	resp.Body.Close()

	// Verify correlation ID is preserved
	messages := s.getMessages(s.agent1URL, 10)
	found := false
	for _, msg := range messages {
		if msg["correlation_id"] == correlationID {
			found = true
			break
		}
	}
	s.Assert().True(found, "Correlation ID not found in messages")
}

// TestConcurrentMessaging tests concurrent message sending
func (s *A2AIntegrationTestSuite) TestConcurrentMessaging() {
	messageCount := 20
	done := make(chan bool, messageCount)

	for i := 0; i < messageCount; i++ {
		go func(idx int) {
			msg := Message{
				ID:        uuid.New().String(),
				Type:      "request",
				Source:    "test-client",
				Dest:      "agent2",
				Timestamp: time.Now(),
				Body: MessageBody{
					Action:  "test",
					Payload: json.RawMessage(fmt.Sprintf(`{"index": %d}`, idx)),
				},
			}

			resp, err := s.postJSON(s.agent2URL+"/a2a/inbound", msg)
			if err == nil {
				resp.Body.Close()
			}
			done <- true
		}(i)
	}

	// Wait for all messages to be sent
	for i := 0; i < messageCount; i++ {
		<-done
	}

	// Give time for processing
	time.Sleep(2 * time.Second)

	// Verify messages were received
	messages := s.getMessages(s.agent2URL, messageCount)
	s.Assert().GreaterOrEqual(len(messages), messageCount)
}

// TestMessageTypes tests different message types
func (s *A2AIntegrationTestSuite) TestMessageTypes() {
	messageTypes := []string{"request", "response", "event", "broadcast"}

	for _, msgType := range messageTypes {
		msg := Message{
			ID:        uuid.New().String(),
			Type:      msgType,
			Source:    "test-client",
			Dest:      "agent3",
			Timestamp: time.Now(),
			Body: MessageBody{
				Action:  "test_type",
				Payload: json.RawMessage(fmt.Sprintf(`{"type": "%s"}`, msgType)),
			},
		}

		resp, err := s.postJSON(s.agent3URL+"/a2a/inbound", msg)
		s.Require().NoError(err)
		resp.Body.Close()
		s.Assert().Equal(http.StatusAccepted, resp.StatusCode, "Failed for message type: %s", msgType)
	}
}

// TestLargeMessage tests sending large messages
func (s *A2AIntegrationTestSuite) TestLargeMessage() {
	// Create large payload (100KB)
	largeData := make([]byte, 100*1024)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"data": largeData,
		"type": "large_message",
	})

	msg := Message{
		ID:        uuid.New().String(),
		Type:      "request",
		Source:    "test-client",
		Dest:      "agent1",
		Timestamp: time.Now(),
		Body: MessageBody{
			Action:      "large_transfer",
			ContentType: "application/json",
			Payload:     payload,
		},
	}

	resp, err := s.postJSON(s.agent1URL+"/a2a/inbound", msg)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Assert().Equal(http.StatusAccepted, resp.StatusCode)
}

// TestInvalidMessage tests error handling for invalid messages
func (s *A2AIntegrationTestSuite) TestInvalidMessage() {
	// Send invalid JSON
	resp, err := s.client.Post(s.agent1URL+"/a2a/inbound", "application/json", 
		bytes.NewReader([]byte("invalid json")))
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Assert().Equal(http.StatusBadRequest, resp.StatusCode)
}

// TestMessageWithMetadata tests messages with metadata
func (s *A2AIntegrationTestSuite) TestMessageWithMetadata() {
	payload := Message{
		ID:        uuid.New().String(),
		Type:      "request",
		Source:    "test-client",
		Dest:      "agent2",
		Topic:     "test.topic",
		Timestamp: time.Now(),
		Body: MessageBody{
			Action:  "test_metadata",
			Payload: json.RawMessage(`{"test": "data"}`),
		},
	}

	resp, err := s.postJSON(s.agent2URL+"/a2a/inbound", payload)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Assert().Equal(http.StatusAccepted, resp.StatusCode)
}

// Helper functions

func (s *A2AIntegrationTestSuite) postJSON(url string, data interface{}) (*http.Response, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return s.client.Do(req)
}

func (s *A2AIntegrationTestSuite) getMessages(agentURL string, limit int) []map[string]interface{} {
	resp, err := s.client.Get(fmt.Sprintf("%s/messages?limit=%d", agentURL, limit))
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var result struct {
		Messages []map[string]interface{} `json:"messages"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	return result.Messages
}

func (s *A2AIntegrationTestSuite) verifyMessageCount(agentURL string, minCount int) {
	s.Eventually(func() bool {
		messages := s.getMessages(agentURL, 100)
		return len(messages) >= minCount
	}, 10*time.Second, 500*time.Millisecond)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TestA2AIntegration runs all A2A integration tests
func TestA2AIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	suite.Run(t, new(A2AIntegrationTestSuite))
}
