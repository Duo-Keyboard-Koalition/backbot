package tools

import (
	"darci-go/darci/agent"
	"darci-go/darci/config"
	"darci-go/darci/state"
)

// RegisterDarciTools wires all DarCI tools into the AdkAgentLoop
func RegisterDarciTools(agentLoop *agent.AdkAgentLoop, cfg *config.DarciConfig, store *state.TaskStore) *SentinelMonitorRegistry {
	registry := NewSentinelMonitorRegistry()
	sendTool := NewSendDarciMessageTool(cfg)

	// Register tools with the agent loop's tool registry
	agentLoop.Tools.Register(ToAdkTool(NewTaskCreateTool(store)))
	agentLoop.Tools.Register(ToAdkTool(NewTaskUpdateTool(store)))
	agentLoop.Tools.Register(ToAdkTool(NewTaskQueryTool(store)))
	agentLoop.Tools.Register(ToAdkTool(NewStatusReportTool(store)))
	agentLoop.Tools.Register(ToAdkTool(NewAssignTaskTool(store, sendTool)))
	agentLoop.Tools.Register(ToAdkTool(NewDiscoverAgentsTool(cfg, store)))
	agentLoop.Tools.Register(ToAdkTool(sendTool))
	agentLoop.Tools.Register(ToAdkTool(NewMonitorAgentTool(cfg, store, agentLoop.Bus, registry)))
	agentLoop.Tools.Register(ToAdkTool(NewNotebookCreateTool(cfg)))
	agentLoop.Tools.Register(ToAdkTool(NewNotebookAppendTool(cfg)))

	return registry
}
