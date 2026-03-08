// Package heartbeat provides periodic agent wake-up functionality.
//
// The heartbeat service periodically checks a HEARTBEAT.md file in the workspace
// and uses an LLM to decide whether there are active tasks to execute.
//
// Usage:
//
//	service := heartbeat.NewHeartbeatService(heartbeat.HeartbeatServiceConfig{
//	    Workspace: "/path/to/workspace",
//	    Provider:  provider,
//	    Model:     "gemini-2.5-flash",
//	    Interval:  30 * time.Minute,
//	    Enabled:   true,
//	})
//
//	if err := service.Start(); err != nil {
//	    log.Fatal(err)
//	}
//	defer service.Stop()
package heartbeat
