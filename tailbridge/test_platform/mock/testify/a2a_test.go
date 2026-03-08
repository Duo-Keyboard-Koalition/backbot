package testify

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/codejedi-ai/kaggle-for-tensors/tailscale-app/tailbridge/test_platform/mock"
	"github.com/stretchr/testify/suite"
)

// A2ATestSuite provides a test suite for A2A communication
type A2ATestSuite struct {
	suite.Suite
	network *mock.MockNetwork
	agent1  *mock.MockAgent
	agent2  *mock.MockAgent
	agent3  *mock.MockAgent
}

// SetupTest spins up a fresh network with agents before each test
func (s *A2ATestSuite) SetupTest() {
	s.network = mock.NewMockNetwork()
	
	// Create and start agents
	s.agent1 = mock.NewMockAgent("agent-alpha", "test.tailnet", []string{"chat", "file_send", "command"})
	s.agent2 = mock.NewMockAgent("agent-beta", "test.tailnet", []string{"chat", "file_receive"})
	s.agent3 = mock.NewMockAgent("agent-gamma", "test.tailnet", []string{"chat", "file_send", "file_receive", "stream"})
	
	// Add agents to network (this starts them)
	s.Require().NoError(s.network.AddAgent(s.agent1))
	s.Require().NoError(s.network.AddAgent(s.agent2))
	s.Require().NoError(s.network.AddAgent(s.agent3))
	
	// Wait for all agents to be running
	s.Require().NoError(s.network.WaitForAllAgents(3, 2*time.Second))
}

// TearDownTest tears down all agents and clears network after each test
func (s *A2ATestSuite) TearDownTest() {
	// Properly remove all agents
	if s.agent1 != nil {
		s.network.RemoveAgent(s.agent1.Name)
	}
	if s.agent2 != nil {
		s.network.RemoveAgent(s.agent2.Name)
	}
	if s.agent3 != nil {
		s.network.RemoveAgent(s.agent3.Name)
	}
	s.network.ClearNetwork()
	s.network = nil
}

// TestAgentDiscovery tests agent discovery mechanisms
func (s *A2ATestSuite) TestAgentDiscovery() {
	// Test get all agents
	allAgents := s.network.GetAllAgents()
	s.Require().Len(allAgents, 3)
	
	// Test search by name
	results := s.network.SearchAgents("alpha")
	s.Require().Len(results, 1)
	s.Assert().Equal("agent-alpha", results[0].Name)
	
	// Test search partial
	results = s.network.SearchAgents("agent")
	s.Require().Len(results, 3)
}

// TestCapabilityFiltering tests filtering agents by capability
func (s *A2ATestSuite) TestCapabilityFiltering() {
	// Get agents with file_send capability
	fileSendAgents := s.network.GetAgentsByCapability("file_send")
	s.Require().Len(fileSendAgents, 2)
	
	// Verify correct agents
	names := make(map[string]bool)
	for _, agent := range fileSendAgents {
		names[agent.Name] = true
	}
	s.Assert().True(names["agent-alpha"])
	s.Assert().True(names["agent-gamma"])
	s.Assert().False(names["agent-beta"])
	
	// Get agents with file_receive capability
	fileReceiveAgents := s.network.GetAgentsByCapability("file_receive")
	s.Require().Len(fileReceiveAgents, 2)
}

// TestBasicMessaging tests basic agent-to-agent messaging
func (s *A2ATestSuite) TestBasicMessaging() {
	payload, _ := json.Marshal(map[string]string{"message": "Hello from alpha!"})
	
	msg := mock.Message{
		Type:      "request",
		Dest:      "agent-beta",
		Topic:     "agent.requests",
		Body: mock.MessageBody{
			Action:      "chat",
			ContentType: "application/json",
			Payload:     payload,
		},
	}
	
	err := s.network.Send("agent-alpha", "agent-beta", msg)
	s.Require().NoError(err)
	
	// Wait for message delivery
	time.Sleep(50 * time.Millisecond)
	
	// Check receiver got the message
	s.Assert().Len(s.agent2.Messages, 1)
	s.Assert().Equal("agent-alpha", s.agent2.Messages[0].Source)
	
	var receivedData map[string]string
	json.Unmarshal(s.agent2.Messages[0].Body.Payload, &receivedData)
	s.Assert().Equal("Hello from alpha!", receivedData["message"])
}

// TestBidirectionalMessaging tests two-way communication
func (s *A2ATestSuite) TestBidirectionalMessaging() {
	// Alpha sends to Beta
	msg1 := mock.Message{
		Type: "request",
		Dest: "agent-beta",
		Body: mock.MessageBody{
			Action:  "ping",
			Payload: json.RawMessage(`{"data": "ping"}`),
		},
	}
	
	err := s.network.Send("agent-alpha", "agent-beta", msg1)
	s.Require().NoError(err)
	
	// Beta responds to Alpha
	msg2 := mock.Message{
		Type: "response",
		Dest: "agent-alpha",
		Body: mock.MessageBody{
			Action:  "pong",
			Payload: json.RawMessage(`{"data": "pong"}`),
		},
	}
	
	err = s.network.Send("agent-beta", "agent-alpha", msg2)
	s.Require().NoError(err)
	
	// Verify both agents received messages
	s.Assert().Len(s.agent1.Messages, 1)
	s.Assert().Len(s.agent2.Messages, 1)
}

// TestTopicBasedRouting tests topic-based message routing
func (s *A2ATestSuite) TestTopicBasedRouting() {
	// Publish to topic
	msg := mock.Message{
		Type:  "event",
		Topic: "agent.events",
		Body: mock.MessageBody{
			Action:  "status_update",
			Payload: json.RawMessage(`{"status": "online"}`),
		},
	}
	
	err := s.network.Publish("agent.events", msg)
	s.Require().NoError(err)
	
	// Retrieve messages from topic
	messages := s.network.GetTopicMessages("agent.events", 10)
	s.Require().Len(messages, 1)
	s.Assert().Equal("agent.events", messages[0].Topic)
	s.Assert().Equal("event", messages[0].Type)
}

// TestConsumerGroups tests consumer group functionality
func (s *A2ATestSuite) TestConsumerGroups() {
	// Create consumer group
	cg := s.network.CreateConsumerGroup("test-group", []string{"agent.requests"})
	s.Require().NotNil(cg)
	
	// Subscribe agents
	err := s.network.Subscribe("test-group", []string{"agent.requests"}, s.agent1)
	s.Require().NoError(err)
	
	err = s.network.Subscribe("test-group", []string{"agent.requests"}, s.agent2)
	s.Require().NoError(err)
	
	// Verify consumer group was created
	s.Assert().Equal("test-group", cg.ID)
	s.Assert().Len(cg.Members, 2)
	s.Assert().Equal(2, cg.Generation)  // Generation increments on each subscribe
	
	// Publish message
	msg := mock.Message{
		Type:  "request",
		Topic: "agent.requests",
		Body: mock.MessageBody{
			Action:  "process",
			Payload: json.RawMessage(`{"task": "test"}`),
		},
	}
	
	err = s.network.Publish("agent.requests", msg)
	s.Require().NoError(err)
	
	// Verify via topic messages
	messages := s.network.GetTopicMessages("agent.requests", 10)
	s.Assert().Len(messages, 1)
	s.Assert().Equal("agent.requests", messages[0].Topic)
}

// TestAgentLifecycle tests agent start/stop functionality
func (s *A2ATestSuite) TestAgentLifecycle() {
	// Stop agent
	err := s.network.StopAgent("agent-beta")
	s.Require().NoError(err)
	
	// Verify agent is offline
	phonebook := s.network.GetPhoneBook()
	for _, agent := range phonebook {
		if agent.Name == "agent-beta" {
			s.Assert().False(agent.Online)
		}
	}
	
	// Try to send to stopped agent (should fail)
	msg := mock.Message{
		Type: "request",
		Dest: "agent-beta",
	}
	err = s.network.Send("agent-alpha", "agent-beta", msg)
	s.Assert().Error(err)
	
	// Restart agent
	err = s.network.StartAgent("agent-beta")
	s.Require().NoError(err)
	
	// Verify agent is online
	phonebook = s.network.GetPhoneBook()
	for _, agent := range phonebook {
		if agent.Name == "agent-beta" {
			s.Assert().True(agent.Online)
		}
	}
}

// TestMessageCorrelation tests request/response correlation
func (s *A2ATestSuite) TestMessageCorrelation() {
	correlationID := "corr-123"
	
	// Send request with correlation ID
	requestMsg := mock.Message{
		Type:          "request",
		Dest:          "agent-beta",
		CorrelationID: correlationID,
		ReplyTo:       "agent.responses",
		Body: mock.MessageBody{
			Action:  "execute",
			Payload: json.RawMessage(`{"command": "test"}`),
		},
	}
	
	err := s.network.Send("agent-alpha", "agent-beta", requestMsg)
	s.Require().NoError(err)
	
	// Send response with same correlation ID
	responseMsg := mock.Message{
		Type:          "response",
		Dest:          "agent-alpha",
		CorrelationID: correlationID,
		Body: mock.MessageBody{
			Action:  "result",
			Payload: json.RawMessage(`{"success": true}`),
		},
	}
	
	err = s.network.Send("agent-beta", "agent-alpha", responseMsg)
	s.Require().NoError(err)
	
	// Verify correlation
	s.Assert().Equal(correlationID, s.agent1.Messages[0].CorrelationID)
}

// TestNetworkLatency tests message delivery with simulated latency
func (s *A2ATestSuite) TestNetworkLatency() {
	// Set latency
	s.network.SetNetworkLatency(100 * time.Millisecond)
	
	start := time.Now()
	
	msg := mock.Message{
		Type: "request",
		Dest: "agent-beta",
	}
	
	err := s.network.Send("agent-alpha", "agent-beta", msg)
	s.Require().NoError(err)
	
	elapsed := time.Since(start)
	s.Assert().GreaterOrEqual(elapsed, 100*time.Millisecond)
	
	// Reset latency
	s.network.SetNetworkLatency(0)
}

// TestBroadcastMessaging tests broadcasting to multiple agents
func (s *A2ATestSuite) TestBroadcastMessaging() {
	// Broadcast to all via topic
	broadcastMsg := mock.Message{
		Type:  "broadcast",
		Topic: "system.announcements",
		Body: mock.MessageBody{
			Action:  "announce",
			Payload: json.RawMessage(`{"message": "System maintenance in 5 minutes"}`),
		},
	}
	
	err := s.network.Publish("system.announcements", broadcastMsg)
	s.Require().NoError(err)
	
	time.Sleep(100 * time.Millisecond)
	
	// Verify all agents received broadcast via topic
	messages := s.network.GetTopicMessages("system.announcements", 10)
	s.Assert().Len(messages, 1)
	s.Assert().Equal("broadcast", messages[0].Type)
	s.Assert().Equal("announce", messages[0].Body.Action)
}

// TestPhoneBook tests phone book functionality
func (s *A2ATestSuite) TestPhoneBook() {
	phonebook := s.network.GetPhoneBook()
	
	s.Require().Len(phonebook, 3)
	
	// Verify phone book entries
	for _, entry := range phonebook {
		s.Assert().NotEmpty(entry.Name)
		s.Assert().NotEmpty(entry.NodeID)
		s.Assert().Equal("test.tailnet", entry.Tailnet)
		s.Assert().NotEmpty(entry.Capabilities)
		s.Assert().True(entry.Online)
	}
	
	// Verify specific agents
	names := make(map[string]bool)
	for _, entry := range phonebook {
		names[entry.Name] = true
	}
	
	s.Assert().True(names["agent-alpha"])
	s.Assert().True(names["agent-beta"])
	s.Assert().True(names["agent-gamma"])
}

// TestAgentNotFound tests error handling for non-existent agents
func (s *A2ATestSuite) TestAgentNotFound() {
	msg := mock.Message{
		Type: "request",
		Dest: "non-existent-agent",
	}
	
	err := s.network.Send("agent-alpha", "non-existent-agent", msg)
	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "not found")
}

// TestConcurrentMessaging tests concurrent message sending
func (s *A2ATestSuite) TestConcurrentMessaging() {
	var wg sync.WaitGroup
	messageCount := 50
	
	// Send messages concurrently
	for i := 0; i < messageCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			msg := mock.Message{
				Type: "request",
				Dest: "agent-beta",
				Body: mock.MessageBody{
					Action:  "test",
					Payload: json.RawMessage(`{"index": ` + string(rune('0'+idx%10)) + `}`),
				},
			}
			s.network.Send("agent-alpha", "agent-beta", msg)
		}(i)
	}
	
	wg.Wait()
	time.Sleep(100 * time.Millisecond)
	
	// Verify all messages received
	s.Assert().Equal(messageCount, len(s.agent2.Messages))
}

// TestA2AMessaging runs all A2A tests
func TestA2AMessaging(t *testing.T) {
	suite.Run(t, new(A2ATestSuite))
}
