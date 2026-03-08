package state

import "time"

// SentinelSnapshot represents the state of a Sentinel monitoring session
type SentinelSnapshot struct {
	RiskScore        float64  `json:"risk_score,omitempty"`
	FailureTypes     []string `json:"failure_types,omitempty"`
	StepCount        int      `json:"step_count,omitempty"`
	InterventionType string   `json:"intervention_type,omitempty"`
}

// DarciRoles represents DARCI role assignments for a task
type DarciRoles struct {
	Driver     string `json:"driver,omitempty"`
	Approver   string `json:"approver,omitempty"`
	Responsible string `json:"responsible,omitempty"`
	Informed   string `json:"informed,omitempty"`
}

// Task represents a task in the system
type Task struct {
	ID               string          `json:"id"`
	Title            string          `json:"title"`
	Description      string          `json:"description,omitempty"`
	Status           string          `json:"status,omitempty"`
	Priority         string          `json:"priority,omitempty"`
	Labels           []string        `json:"labels,omitempty"`
	Dependencies     []string        `json:"dependencies,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	Darci            DarciRoles      `json:"darci,omitempty"`
	SentinelSnapshot SentinelSnapshot `json:"sentinel_snapshot,omitempty"`
}

// AgentAssignment represents an agent's assignment to a task
type AgentAssignment struct {
	TaskID    string  `json:"task_id"`
	DARCIROle string  `json:"darci_role"`
	RiskScore float64 `json:"risk_score"`
	Status    string  `json:"status"`
}

// DarciState represents the overall DARCI state
type DarciState struct {
	ActiveMonitors []string     `json:"active_monitors,omitempty"`
	LastDiscovery  *time.Time   `json:"last_discovery,omitempty"`
}

// AgentContext represents the agent context stored in state
type AgentContext struct {
	AgentAssignments map[string]*AgentAssignment `json:"agent_assignments,omitempty"`
	DarciState       *DarciState                 `json:"darci_state,omitempty"`
}
