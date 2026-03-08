package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	configDirName  = ".tailtalkie"
	configFileName = "config.json"
)

// BridgeConfig contains runtime settings for the bridge bot client.
type BridgeConfig struct {
	BridgeName      string `json:"bridge_name"`
	StateDir        string `json:"state_dir"`
	AuthKey         string `json:"auth_key"`
	LocalAgentURL   string `json:"local_agent_url"`
	PeerInboundPort int    `json:"peer_inbound_port"`
	InboundPort     int    `json:"inbound_port"`
	LocalListen     string `json:"local_listen"`
}

// DefaultConfig returns a BridgeConfig with default values.
func DefaultConfig() BridgeConfig {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "~"
	}

	return BridgeConfig{
		BridgeName:      defaultBridgeName,
		StateDir:        filepath.Join(homeDir, configDirName, "state"),
		AuthKey:         "",
		LocalAgentURL:   defaultLocalAgent,
		PeerInboundPort: defaultPeerInPort,
		InboundPort:     defaultInboundPort,
		LocalListen:     defaultLocalListen,
	}
}

// loadConfig loads configuration from ~/.tailtalkie/config.json
func loadConfig() (BridgeConfig, error) {
	cfg := DefaultConfig()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return cfg, fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, configDirName, configFileName)

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return cfg, fmt.Errorf("config file not found: %s\nRun 'init' command or create config manually", configPath)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return cfg, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	// Validate required fields
	if cfg.AuthKey == "" {
		return cfg, fmt.Errorf("auth_key is required in config")
	}
	if cfg.BridgeName == "" {
		return cfg, fmt.Errorf("bridge_name is required in config")
	}

	return cfg, nil
}

// ensureConfigDir creates the config directory if it doesn't exist.
func ensureConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, configDirName)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return configDir, nil
}

// saveConfig writes the configuration to ~/.tailtalkie/config.json
func saveConfig(cfg BridgeConfig) error {
	configDir, err := ensureConfigDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(configDir, configFileName)

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Config saved to %s\n", configPath)
	return nil
}
