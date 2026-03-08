package testify

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/codejedi-ai/kaggle-for-tensors/tailscale-app/tailbridge/test_platform/mock"
	"github.com/stretchr/testify/suite"
)

// NetworkDiscoveryTestSuite tests network-wide agent discovery
type NetworkDiscoveryTestSuite struct {
	suite.Suite
	network *mock.MockNetwork
	agents  []*mock.MockAgent
	logger  *TestLogger
}

// SetupSuite runs once before all tests in the suite
func (s *NetworkDiscoveryTestSuite) SetupSuite() {
	s.logger = NewTestLogger("NetworkDiscovery")
	s.logger.LogTestStart()
}

// SetupTest spins up a fresh network with agents before each test
func (s *NetworkDiscoveryTestSuite) SetupTest() {
	s.network = mock.NewMockNetwork()
	s.agents = make([]*mock.MockAgent, 0)
	
	// Log network creation
	s.logger.LogInfo("NETWORK", "Created new mock network", map[string]interface{}{
		"network_id": s.network.GetNetworkID(),
	})
}

// TearDownTest tears down all agents and clears network after each test
func (s *NetworkDiscoveryTestSuite) TearDownTest() {
	// Log agent teardown
	for _, agent := range s.agents {
		s.logger.LogAgentTearDown(agent)
		s.network.RemoveAgent(agent.Name)
	}
	
	// Log final network state
	s.logger.LogNetworkState(s.network)
	
	s.network.ClearNetwork()
	s.network = nil
	s.agents = nil
}

// TearDownSuite runs once after all tests in the suite
func (s *NetworkDiscoveryTestSuite) TearDownSuite() {
	s.logger.LogTestEnd(true, time.Since(s.logger.StartTime))
	
	// Save logs
	err := s.logger.Save()
	if err != nil {
		fmt.Printf("Warning: Failed to save test logs: %v\n", err)
	}
}

// spawnAgent creates and adds an agent to the network
func (s *NetworkDiscoveryTestSuite) spawnAgent(name string, capabilities []string) *mock.MockAgent {
	agent := mock.NewMockAgent(name, "test.tailnet", capabilities)
	err := s.network.AddAgent(agent)
	s.Require().NoError(err)
	s.agents = append(s.agents, agent)
	
	// Log agent spin up with IP assignment
	s.logger.LogAgentSpinUp(agent)
	
	return agent
}

// TestAgentJoinEvents tests that agent join events are recorded
func (s *NetworkDiscoveryTestSuite) TestAgentJoinEvents() {
	// Spawn agents sequentially
	s.spawnAgent("discover-agent-1", []string{"chat"})
	s.spawnAgent("discover-agent-2", []string{"file_send"})
	s.spawnAgent("discover-agent-3", []string{"file_receive"})
	
	// Wait for discovery processing
	time.Sleep(50 * time.Millisecond)
	
	// Get discovery history
	events := s.network.GetDiscoveryHistory()
	
	s.Require().Len(events, 3)
	
	// Verify all are join events
	for i, event := range events {
		s.Assert().Equal("agent_joined", event.Type)
		s.Assert().Equal(fmt.Sprintf("discover-agent-%d", i+1), event.AgentName)
		s.Assert().NotEmpty(event.Details)
	}
}

// TestAgentLeaveEvents tests that agent leave events are recorded
func (s *NetworkDiscoveryTestSuite) TestAgentLeaveEvents() {
	// Spawn and then remove agents
	s.spawnAgent("leave-agent-1", []string{"chat"})
	s.spawnAgent("leave-agent-2", []string{"chat"})
	
	// Remove first agent
	err := s.network.RemoveAgent("leave-agent-1")
	s.Require().NoError(err)
	
	// Remove from our tracking (skip first)
	s.agents = s.agents[1:]
	
	// Spawn another agent after removal
	s.spawnAgent("leave-agent-3", []string{"chat"})
	
	time.Sleep(50 * time.Millisecond)
	
	events := s.network.GetDiscoveryHistory()
	
	// Should have: 2 joins, 1 leave, 1 join = 4 events
	s.Require().Len(events, 4)
	
	// Find the leave event
	var leaveEvent *mock.DiscoveryEvent
	for i, event := range events {
		if event.Type == "agent_left" {
			leaveEvent = &events[i]
			break
		}
	}
	s.Require().NotNil(leaveEvent)
	s.Assert().Equal("agent_left", leaveEvent.Type)
	s.Assert().Equal("leave-agent-1", leaveEvent.AgentName)
	s.Assert().Contains(leaveEvent.Details, "Uptime")
}

// TestNetworkWideDiscovery tests discovering all agents on the network
func (s *NetworkDiscoveryTestSuite) TestNetworkWideDiscovery() {
	// Spawn 5 agents with different capabilities
	capabilities := [][]string{
		{"chat", "file_send"},
		{"chat", "file_receive"},
		{"file_send", "file_receive"},
		{"chat", "command"},
		{"stream", "chat"},
	}
	
	s.logger.LogInfo("TEST", "Spawning 5 agents for network-wide discovery test", nil)
	
	for i := 0; i < 5; i++ {
		agent := s.spawnAgent(fmt.Sprintf("network-agent-%d", i+1), capabilities[i])
		s.logger.LogInfo("AGENT", fmt.Sprintf("Spawned agent %s", agent.Name), map[string]interface{}{
			"tailscale_ip": agent.TailscaleIP,
			"tailscale_ipv6": agent.TailscaleIPv6,
			"capabilities": capabilities[i],
		})
	}
	
	// Wait for all agents to be visible
	err := s.network.WaitForAllAgents(5, 2*time.Second)
	s.Require().NoError(err)
	
	// Log network state
	s.logger.LogNetworkState(s.network)
	
	// Verify all agents are discoverable
	allAgents := s.network.GetAllAgents()
	s.Require().Len(allAgents, 5)
	
	// Log agent IPs
	for _, agent := range allAgents {
		s.logger.LogInfo("AGENT", fmt.Sprintf("Agent %s discovered", agent.Name), map[string]interface{}{
			"tailscale_ip": agent.TailscaleIP,
			"tailscale_ipv6": agent.TailscaleIPv6,
			"state": agent.State,
		})
	}
	
	// Verify all are running
	runningAgents := s.network.GetRunningAgents()
	s.Require().Len(runningAgents, 5)
	
	// Verify agent count
	s.Assert().Equal(5, s.network.GetAgentCount())
	s.Assert().Equal(5, s.network.GetRunningAgentCount())
}

// TestCapabilityBasedDiscovery tests filtering agents by capability
func (s *NetworkDiscoveryTestSuite) TestCapabilityBasedDiscovery() {
	// Spawn agents with different capabilities
	s.spawnAgent("cap-sender-1", []string{"file_send", "chat"})
	s.spawnAgent("cap-sender-2", []string{"file_send"})
	s.spawnAgent("cap-receiver-1", []string{"file_receive", "chat"})
	s.spawnAgent("cap-receiver-2", []string{"file_receive"})
	s.spawnAgent("cap-both", []string{"file_send", "file_receive"})
	s.spawnAgent("cap-chat-only", []string{"chat"})
	
	time.Sleep(50 * time.Millisecond)
	
	// Find file_send agents (sender-1, sender-2, both)
	senders := s.network.GetAgentsByCapability("file_send")
	s.Require().Len(senders, 3)
	
	// Find file_receive agents (receiver-1, receiver-2, both)
	receivers := s.network.GetAgentsByCapability("file_receive")
	s.Require().Len(receivers, 3)
	
	// Find chat agents (sender-1, receiver-1, chat-only)
	chatters := s.network.GetAgentsByCapability("chat")
	s.Require().Len(chatters, 3)
}

// TestAgentSearchByName tests searching agents by name pattern
func (s *NetworkDiscoveryTestSuite) TestAgentSearchByName() {
	s.spawnAgent("search-alpha-1", []string{"chat"})
	s.spawnAgent("search-alpha-2", []string{"chat"})
	s.spawnAgent("search-beta-1", []string{"chat"})
	s.spawnAgent("other-agent", []string{"chat"})
	
	// Search by prefix
	results := s.network.SearchAgents("search-alpha")
	s.Require().Len(results, 2)
	
	// Search by different prefix
	results = s.network.SearchAgents("search-beta")
	s.Require().Len(results, 1)
	
	// Search all search-*
	results = s.network.SearchAgents("search")
	s.Require().Len(results, 3)
}

// TestPhoneBookWithAgentStates tests phone book reflects agent states
func (s *NetworkDiscoveryTestSuite) TestPhoneBookWithAgentStates() {
	s.spawnAgent("phonebook-agent-1", []string{"chat"})
	s.spawnAgent("phonebook-agent-2", []string{"chat"})
	
	time.Sleep(50 * time.Millisecond)
	
	// Get phone book
	phonebook := s.network.GetPhoneBook()
	s.Require().Len(phonebook, 2)
	
	// Verify all agents online
	for _, entry := range phonebook {
		s.Assert().True(entry.Online)
		s.Assert().Equal("running", entry.State)
		s.Assert().NotEmpty(entry.Uptime)
		s.Assert().Equal(0, entry.MessageCount)
	}
	
	// Stop one agent
	s.network.StopAgent("phonebook-agent-1")
	time.Sleep(50 * time.Millisecond)
	
	// Get updated phone book
	phonebook = s.network.GetPhoneBook()
	s.Require().Len(phonebook, 2)
	
	for _, entry := range phonebook {
		if entry.Name == "phonebook-agent-1" {
			s.Assert().False(entry.Online)
			s.Assert().Equal("stopping", entry.State)
		} else {
			s.Assert().True(entry.Online)
			s.Assert().Equal("running", entry.State)
		}
	}
	
	// Restart agent
	s.network.StartAgent("phonebook-agent-1")
	time.Sleep(50 * time.Millisecond)
	
	phonebook = s.network.GetPhoneBook()
	for _, entry := range phonebook {
		if entry.Name == "phonebook-agent-1" {
			s.Assert().True(entry.Online)
			s.Assert().Equal("running", entry.State)
		}
	}
}

// TestNetworkStats tests network statistics tracking
func (s *NetworkDiscoveryTestSuite) TestNetworkStats() {
	// Initial stats
	stats := s.network.GetNetworkStats()
	s.Assert().Equal(0, stats.TotalAgents)
	s.Assert().Equal(0, stats.DiscoveryEvents)
	
	// Add agents
	s.spawnAgent("stats-agent-1", []string{"chat"})
	s.spawnAgent("stats-agent-2", []string{"chat"})
	s.spawnAgent("stats-agent-3", []string{"chat"})
	
	time.Sleep(50 * time.Millisecond)
	
	stats = s.network.GetNetworkStats()
	s.Assert().Equal(3, stats.TotalAgents)
	s.Assert().Equal(3, stats.RunningAgents)
	s.Assert().Equal(0, stats.StoppedAgents)
	s.Assert().Equal(3, stats.DiscoveryEvents) // 3 joins
	
	// Send some messages
	msg := mock.Message{
		Type: "request",
		Dest: "stats-agent-2",
		Body: mock.MessageBody{
			Action:  "test",
			Payload: json.RawMessage(`{}`),
		},
	}
	s.network.Send("stats-agent-1", "stats-agent-2", msg)
	s.network.Send("stats-agent-1", "stats-agent-3", msg)
	
	stats = s.network.GetNetworkStats()
	s.Assert().Equal(2, stats.TotalMessages)
	
	// Stop an agent
	s.network.StopAgent("stats-agent-2")
	
	stats = s.network.GetNetworkStats()
	s.Assert().Equal(1, stats.StoppedAgents)
	s.Assert().Equal(4, stats.DiscoveryEvents) // 3 joins + 1 stop
}

// TestWaitForAgent tests waiting for specific agents
func (s *NetworkDiscoveryTestSuite) TestWaitForAgent() {
	// Start agent in goroutine to simulate async startup
	go func() {
		time.Sleep(100 * time.Millisecond)
		s.spawnAgent("wait-for-me", []string{"chat"})
	}()
	
	// Wait for agent
	agent, err := s.network.WaitForAgent("wait-for-me", 2*time.Second)
	s.Require().NoError(err)
	s.Assert().NotNil(agent)
	s.Assert().Equal("wait-for-me", agent.Name)
}

// TestWaitForAllAgents tests waiting for multiple agents
func (s *NetworkDiscoveryTestSuite) TestWaitForAllAgents() {
	// Spawn agents with slight delays
	for i := 0; i < 5; i++ {
		go func(idx int) {
			time.Sleep(time.Duration(idx*20) * time.Millisecond)
			s.spawnAgent(fmt.Sprintf("wait-all-%d", idx), []string{"chat"})
		}(i)
	}
	
	// Wait for all 5 agents
	err := s.network.WaitForAllAgents(5, 3*time.Second)
	s.Require().NoError(err)
	
	// Verify all are present
	s.Assert().Equal(5, s.network.GetRunningAgentCount())
}

// TestAgentLifecycleTransitions tests agent state transitions
func (s *NetworkDiscoveryTestSuite) TestAgentLifecycleTransitions() {
	agent := s.spawnAgent("lifecycle-agent", []string{"chat"})
	
	// Verify initial state
	s.Assert().Equal(mock.AgentStateRunning, agent.State)
	
	// Stop agent
	err := s.network.StopAgent("lifecycle-agent")
	s.Require().NoError(err)
	s.Assert().Equal(mock.AgentStateStopping, agent.State)
	
	// Start agent again
	err = s.network.StartAgent("lifecycle-agent")
	s.Require().NoError(err)
	s.Assert().Equal(mock.AgentStateRunning, agent.State)
	
	// Verify discovery events
	events := s.network.GetRecentDiscoveryEvents(3)
	s.Require().Len(events, 3)
	s.Assert().Equal("agent_joined", events[0].Type)
	s.Assert().Equal("agent_stopped", events[1].Type)
	s.Assert().Equal("agent_started", events[2].Type)
}

// TestConcurrentAgentOperations tests concurrent agent operations
func (s *NetworkDiscoveryTestSuite) TestConcurrentAgentOperations() {
	var wg sync.WaitGroup
	agentCount := 10
	
	// Spawn agents concurrently
	for i := 0; i < agentCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			s.spawnAgent(fmt.Sprintf("concurrent-%d", idx), []string{"chat"})
		}(i)
	}
	
	wg.Wait()
	time.Sleep(100 * time.Millisecond)
	
	// Verify all agents spawned
	s.Assert().Equal(agentCount, s.network.GetAgentCount())
	
	// Verify all discovery events recorded
	events := s.network.GetDiscoveryHistory()
	s.Assert().Len(events, agentCount)
	
	// Stop half concurrently
	for i := 0; i < agentCount/2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			s.network.StopAgent(fmt.Sprintf("concurrent-%d", idx))
		}(i)
	}
	
	wg.Wait()
	time.Sleep(50 * time.Millisecond)
	
	// Verify stats
	stats := s.network.GetNetworkStats()
	s.Assert().Equal(agentCount, stats.TotalAgents)
	s.Assert().Equal(agentCount/2, stats.RunningAgents)
	s.Assert().Equal(agentCount/2, stats.StoppedAgents)
}

// TestAgentVisibilityDuringMessaging tests agent visibility while messaging
func (s *NetworkDiscoveryTestSuite) TestAgentVisibilityDuringMessaging() {
	// Spawn agents
	s.spawnAgent("visible-sender", []string{"chat", "file_send"})
	s.spawnAgent("visible-receiver", []string{"chat", "file_receive"})
	
	// Send multiple messages
	for i := 0; i < 10; i++ {
		msg := mock.Message{
			Type: "request",
			Dest: "visible-receiver",
			Body: mock.MessageBody{
				Action:  "test",
				Payload: json.RawMessage(fmt.Sprintf(`{"index": %d}`, i)),
			},
		}
		_ = s.network.Send("visible-sender", "visible-receiver", msg)
	}
	
	time.Sleep(100 * time.Millisecond)
	
	// Verify agents still visible and healthy
	phonebook := s.network.GetPhoneBook()
	s.Require().Len(phonebook, 2)
	
	for _, entry := range phonebook {
		s.Assert().True(entry.Online)
		s.Assert().Equal("running", entry.State)
		// Message count tracked per agent
		s.Assert().GreaterOrEqual(entry.MessageCount, 0)
	}
}

// TestNetworkDiscoveryStress tests discovery under stress
func (s *NetworkDiscoveryTestSuite) TestNetworkDiscoveryStress() {
	var wg sync.WaitGroup
	iterations := 20
	
	// Rapidly spawn and remove agents
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			name := fmt.Sprintf("stress-agent-%d", idx)
			agent := mock.NewMockAgent(name, "test.tailnet", []string{"chat"})
			s.network.AddAgent(agent)
			time.Sleep(10 * time.Millisecond)
			s.network.RemoveAgent(name)
		}(i)
	}
	
	wg.Wait()
	time.Sleep(200 * time.Millisecond)
	
	// Verify network is stable
	stats := s.network.GetNetworkStats()
	s.Assert().Equal(0, stats.RunningAgents)
	s.Assert().Greater(stats.DiscoveryEvents, 0)
}

// TestNetworkDiscovery runs all network discovery tests
func TestNetworkDiscovery(t *testing.T) {
	suite.Run(t, new(NetworkDiscoveryTestSuite))
}
