package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// runAIPCommand handles AIP CLI commands
func runAIPCommand(args []string) {
	if len(args) == 0 {
		printAIPUsage()
		os.Exit(1)
	}

	subcommand := args[0]

	switch subcommand {
	case "list":
		runAIPList(args[1:])
	case "pending":
		runAIPPending(args[1:])
	case "approve":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Error: agent_id required")
			fmt.Fprintln(os.Stderr, "Usage: taila2a aip approve <agent_id>")
			os.Exit(1)
		}
		runAIPApprove(args[1])
	case "reject":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Error: agent_id required")
			fmt.Fprintln(os.Stderr, "Usage: taila2a aip reject <agent_id>")
			os.Exit(1)
		}
		runAIPReject(args[1])
	case "remove":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Error: agent_id required")
			fmt.Fprintln(os.Stderr, "Usage: taila2a aip remove <agent_id>")
			os.Exit(1)
		}
		runAIPRemove(args[1])
	case "info":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Error: agent_id required")
			fmt.Fprintln(os.Stderr, "Usage: taila2a aip info <agent_id>")
			os.Exit(1)
		}
		runAIPInfo(args[1])
	case "registry":
		runAIPRegistry()
	default:
		fmt.Fprintf(os.Stderr, "Unknown AIP command: %s\n\n", subcommand)
		printAIPUsage()
		os.Exit(1)
	}
}

func printAIPUsage() {
	fmt.Println("AIP (Agent Identification Protocol) Commands")
	fmt.Println()
	fmt.Println("Usage: taila2a aip <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  list                 List all registered agents")
	fmt.Println("  pending              List pending agent registrations")
	fmt.Println("  approve <agent_id>   Approve a pending registration")
	fmt.Println("  reject <agent_id>    Reject a pending registration")
	fmt.Println("  remove <agent_id>    Remove an agent from registry")
	fmt.Println("  info <agent_id>      Show detailed agent information")
	fmt.Println("  registry             Show raw registry file contents")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  taila2a aip list")
	fmt.Println("  taila2a aip pending")
	fmt.Println("  taila2a aip approve darci-python-001")
	fmt.Println("  taila2a aip info darci-python-001")
}

// runAIPList lists all registered agents
func runAIPList(args []string) {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	registry, err := NewAgentRegistry(cfg.StateDir, cfg.BridgeName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading registry: %v\n", err)
		os.Exit(1)
	}

	agents := registry.GetApprovedAgents()
	if len(agents) == 0 {
		fmt.Println("No approved agents registered")
		return
	}

	fmt.Printf("Registered Agents (%d):\n\n", len(agents))
	for _, agent := range agents {
		status := "✓"
		if agent.Status == StatusOffline {
			status = "○"
		}
		fmt.Printf("  %s %s (%s)\n", status, agent.AgentID, agent.AgentType)
		fmt.Printf("     Endpoints: %s\n", agent.Endpoints.Primary)
		fmt.Printf("     Capabilities: %v\n", agent.Capabilities)
		fmt.Printf("     Last Heartbeat: %s\n", agent.LastHeartbeat.Format(time.RFC3339))
		fmt.Println()
	}
}

// runAIPPending lists pending agent registrations
func runAIPPending(args []string) {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	registry, err := NewAgentRegistry(cfg.StateDir, cfg.BridgeName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading registry: %v\n", err)
		os.Exit(1)
	}

	pending := registry.GetPendingAgents()
	if len(pending) == 0 {
		fmt.Println("No pending registrations")
		return
	}

	fmt.Printf("Pending Registrations (%d):\n\n", len(pending))
	for _, agent := range pending {
		fmt.Printf("  Agent ID: %s\n", agent.AgentID)
		fmt.Printf("  Type: %s\n", agent.AgentType)
		fmt.Printf("  Version: %s\n", agent.AgentVersion)
		fmt.Printf("  Registered: %s\n", agent.RegisteredAt.Format(time.RFC3339))
		fmt.Printf("  Hostname: %s\n", agent.Metadata.Hostname)
		fmt.Printf("  Endpoints: %s\n", agent.Endpoints.Primary)
		fmt.Printf("  Capabilities: %v\n", agent.Capabilities)
		fmt.Println()
		fmt.Printf("  To approve: taila2a aip approve %s\n", agent.AgentID)
		fmt.Printf("  To reject:  taila2a aip reject %s\n", agent.AgentID)
		fmt.Println()
	}
}

// runAIPApprove approves a pending registration
func runAIPApprove(agentID string) {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	registry, err := NewAgentRegistry(cfg.StateDir, cfg.BridgeName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading registry: %v\n", err)
		os.Exit(1)
	}

	if err := registry.ApproveAgent(agentID); err != nil {
		fmt.Fprintf(os.Stderr, "Error approving agent: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Agent %s approved successfully\n", agentID)
	fmt.Println()
	fmt.Println("The agent can now:")
	fmt.Println("  - Send heartbeats to /aip/heartbeat")
	fmt.Println("  - Communicate with paired peer bridges")
	fmt.Println("  - Receive tasks from other agents")
}

// runAIPReject rejects a pending registration
func runAIPReject(agentID string) {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	registry, err := NewAgentRegistry(cfg.StateDir, cfg.BridgeName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading registry: %v\n", err)
		os.Exit(1)
	}

	if err := registry.RejectAgent(agentID); err != nil {
		fmt.Fprintf(os.Stderr, "Error rejecting agent: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Agent %s rejected\n", agentID)
}

// runAIPRemove removes an agent from the registry
func runAIPRemove(agentID string) {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	registry, err := NewAgentRegistry(cfg.StateDir, cfg.BridgeName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading registry: %v\n", err)
		os.Exit(1)
	}

	if err := registry.RemoveAgent(agentID); err != nil {
		fmt.Fprintf(os.Stderr, "Error removing agent: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Agent %s removed from registry\n", agentID)
}

// runAIPInfo shows detailed information about an agent
func runAIPInfo(agentID string) {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	registry, err := NewAgentRegistry(cfg.StateDir, cfg.BridgeName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading registry: %v\n", err)
		os.Exit(1)
	}

	agent, exists := registry.GetAgent(agentID)
	if !exists {
		fmt.Fprintf(os.Stderr, "Agent %s not found\n", agentID)
		os.Exit(1)
	}

	fmt.Printf("Agent Information: %s\n\n", agentID)
	fmt.Printf("  Type:       %s\n", agent.AgentType)
	fmt.Printf("  Version:    %s\n", agent.AgentVersion)
	fmt.Printf("  Status:     %s\n", agent.Status)
	fmt.Printf("  Registered: %s\n", agent.RegisteredAt.Format(time.RFC3339))
	if agent.ApprovedAt != nil {
		fmt.Printf("  Approved:   %s\n", agent.ApprovedAt.Format(time.RFC3339))
	}
	fmt.Printf("  Last Beat:  %s\n", agent.LastHeartbeat.Format(time.RFC3339))
	fmt.Println()
	fmt.Printf("  Endpoints:\n")
	fmt.Printf("    Primary: %s\n", agent.Endpoints.Primary)
	if agent.Endpoints.Health != "" {
		fmt.Printf("    Health:  %s\n", agent.Endpoints.Health)
	}
	fmt.Println()
	fmt.Printf("  Capabilities: %v\n", agent.Capabilities)
	fmt.Println()
	fmt.Printf("  Metadata:\n")
	fmt.Printf("    Hostname: %s\n", agent.Metadata.Hostname)
	if agent.Metadata.OS != "" {
		fmt.Printf("    OS:       %s\n", agent.Metadata.OS)
	}
	fmt.Printf("    Tags:     %v\n", agent.Metadata.Tags)
}

// runAIPRegistry shows the raw registry file
func runAIPRegistry() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	registryPath := filepath.Join(cfg.StateDir, "registry.json")
	data, err := os.ReadFile(registryPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Registry file does not exist")
			fmt.Println("No agents have been registered yet")
			return
		}
		fmt.Fprintf(os.Stderr, "Error reading registry: %v\n", err)
		os.Exit(1)
	}

	// Pretty print the JSON
	var prettyJSON map[string]interface{}
	if err := json.Unmarshal(data, &prettyJSON); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing registry: %v\n", err)
		os.Exit(1)
	}

	prettyData, err := json.MarshalIndent(prettyJSON, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting registry: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Registry File: %s\n\n", registryPath)
	fmt.Println(string(prettyData))
}

// HTTP-based AIP commands (when bridge is running)

func getBridgeURL() string {
	cfg, err := loadConfig()
	if err != nil {
		return ""
	}
	return fmt.Sprintf("http://127.0.0.1:%d", cfg.InboundPort)
}

func runAIPListHTTP() {
	baseURL := getBridgeURL()
	if baseURL == "" {
		fmt.Fprintln(os.Stderr, "Error: could not load config")
		os.Exit(1)
	}

	resp, err := http.Get(fmt.Sprintf("%s/aip/agents", baseURL))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to bridge: %v\n", err)
		fmt.Fprintln(os.Stderr, "Make sure the bridge is running: taila2a run")
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(os.Stderr, "Error: %s\n", string(body))
		os.Exit(1)
	}

	var agents []RegisteredAgent
	if err := json.NewDecoder(resp.Body).Decode(&agents); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		os.Exit(1)
	}

	if len(agents) == 0 {
		fmt.Println("No agents registered")
		return
	}

	fmt.Printf("Registered Agents (%d):\n\n", len(agents))
	for _, agent := range agents {
		fmt.Printf("  %s (%s) - %s\n", agent.AgentID, agent.AgentType, agent.Status)
	}
}

func runAIPApproveHTTP(agentID string) {
	baseURL := getBridgeURL()
	if baseURL == "" {
		fmt.Fprintln(os.Stderr, "Error: could not load config")
		os.Exit(1)
	}

	resp, err := http.Post(fmt.Sprintf("%s/aip/approve/%s", baseURL, agentID), "application/json", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to bridge: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(os.Stderr, "Error: %s\n", string(body))
		os.Exit(1)
	}

	fmt.Printf("✓ Agent %s approved\n", agentID)
}

// runSecretsCommand handles secrets management CLI commands
func runSecretsCommand(args []string) {
	if len(args) == 0 {
		printSecretsUsage()
		os.Exit(1)
	}

	subcommand := args[0]

	switch subcommand {
	case "generate":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Error: agent_id required")
			fmt.Fprintln(os.Stderr, "Usage: taila2a secrets generate <agent_id>")
			os.Exit(1)
		}
		runSecretsGenerate(args[1])
	case "list":
		runSecretsList()
	case "remove":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Error: agent_id required")
			fmt.Fprintln(os.Stderr, "Usage: taila2a secrets remove <agent_id>")
			os.Exit(1)
		}
		runSecretsRemove(args[1])
	case "show":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Error: agent_id required")
			fmt.Fprintln(os.Stderr, "Usage: taila2a secrets show <agent_id>")
			os.Exit(1)
		}
		runSecretsShow(args[1])
	default:
		fmt.Fprintf(os.Stderr, "Unknown secrets command: %s\n\n", subcommand)
		printSecretsUsage()
		os.Exit(1)
	}
}

func printSecretsUsage() {
	fmt.Println("Secrets Management Commands")
	fmt.Println()
	fmt.Println("Usage: taila2a secrets <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  generate <agent_id>   Generate new secret for agent")
	fmt.Println("  list                  List all agents with secrets")
	fmt.Println("  remove <agent_id>     Remove agent secret")
	fmt.Println("  show <agent_id>       Show agent secret (for initial setup)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  taila2a secrets generate darci-python-001")
	fmt.Println("  taila2a secrets list")
	fmt.Println("  taila2a secrets show darci-python-001")
}

// runSecretsGenerate generates a new secret for an agent
func runSecretsGenerate(agentID string) {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	secret, err := GenerateAgentSecret(cfg.StateDir, agentID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating secret: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Generated secret for agent: %s\n\n", agentID)
	fmt.Println("Secret:")
	fmt.Printf("  %s\n\n", secret)
	fmt.Println("IMPORTANT: Copy this secret and store it securely!")
	fmt.Println("Configure your agent with this secret, then run:")
	fmt.Printf("  taila2a aip approve %s\n\n", agentID)
	fmt.Println("The secret will be hidden in future 'show' commands for security.")
}

// runSecretsList lists all agents with secrets
func runSecretsList() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	agentIDs := ListAgentSecrets(cfg.StateDir)
	if len(agentIDs) == 0 {
		fmt.Println("No agent secrets configured")
		return
	}

	fmt.Printf("Agent Secrets (%d):\n\n", len(agentIDs))
	for _, agentID := range agentIDs {
		fmt.Printf("  • %s\n", agentID)
	}
	fmt.Println()
	fmt.Println("Use 'taila2a secrets show <agent_id>' to view a secret")
}

// runSecretsRemove removes an agent secret
func runSecretsRemove(agentID string) {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	if err := RemoveAgentSecret(cfg.StateDir, agentID); err != nil {
		fmt.Fprintf(os.Stderr, "Error removing secret: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Removed secret for agent: %s\n", agentID)
}

// runSecretsShow shows an agent secret
func runSecretsShow(agentID string) {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	secret, exists := GetAgentSecret(cfg.StateDir, agentID)
	if !exists {
		fmt.Fprintf(os.Stderr, "No secret found for agent: %s\n", agentID)
		os.Exit(1)
	}

	fmt.Printf("Secret for agent: %s\n\n", agentID)
	fmt.Printf("  %s\n\n", secret)
	fmt.Println("Keep this secret secure and never share it!")
}
