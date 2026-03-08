package config

import (
	"os"
	"path/filepath"
)

// DarciConfig holds DarCI-specific configuration
type DarciConfig struct {
	BridgeLocalURL     string
	SentinelPort       int
	StateDir           string
	NotebookDir        string
	DiscoveryIntervalS int
}

// DefaultDarciConfig returns a DarciConfig with default values
func DefaultDarciConfig() *DarciConfig {
	homeDir, _ := os.UserHomeDir()
	
	return &DarciConfig{
		BridgeLocalURL:     "http://localhost:8080",
		SentinelPort:       8000,
		StateDir:           filepath.Join(homeDir, ".darci"),
		NotebookDir:        "darci/engineering_notebook",
		DiscoveryIntervalS: 30,
	}
}
