package models

import (
	"encoding/json"
	"time"
)

// Envelope represents the message structure for agent communication
type Envelope struct {
	SourceNode string          `json:"source_node"`
	DestNode   string          `json:"dest_node"`
	Payload    json.RawMessage `json:"payload"`
	Timestamp  time.Time       `json:"timestamp,omitempty"`
}

// AgentInfo represents a discovered agent on the tailnet
type AgentInfo struct {
	Name      string        `json:"name"`
	Hostname  string        `json:"hostname"`
	IP        string        `json:"ip"`
	Online    bool          `json:"online"`
	LastSeen  time.Time     `json:"last_seen"`
	Gateways  []GatewayInfo `json:"gateways"`
}

// GatewayInfo represents an open gateway port on an agent
type GatewayInfo struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Service  string `json:"service"`
}

// Config contains runtime settings for agnes
type Config struct {
	Name            string `json:"name"`
	StateDir        string `json:"state_dir"`
	AuthKey         string `json:"auth_key"`
	LocalAgentURL   string `json:"local_agent_url"`
	PeerInboundPort int    `json:"peer_inbound_port"`
	InboundPort     int    `json:"inbound_port"`
	LocalListen     string `json:"local_listen"`
}

// Task represents a task to be executed
type Task struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	Status    string                 `json:"status"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// Command represents a command to be sent to an agent
type Command struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
	Target    string                 `json:"target"`
}

// Response represents a response from an agent
type Response struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}
