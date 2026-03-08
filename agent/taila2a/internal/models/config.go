package models

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	configDirName  = ".taila2a"
	configFileName = "config.json"
)

// DefaultConfig returns a Config with default values
func DefaultConfig() Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "~"
	}

	return Config{
		Name:            "agnes-default",
		StateDir:        filepath.Join(homeDir, configDirName, "state"),
		AuthKey:         "",
		LocalAgentURL:   "http://127.0.0.1:9090/api",
		PeerInboundPort: 8001,
		InboundPort:     8001,
		LocalListen:     "127.0.0.1:8080",
	}
}

// LoadConfig loads configuration from ~/.taila2a/config.json
func LoadConfig() (Config, error) {
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
	if cfg.Name == "" {
		return cfg, fmt.Errorf("name is required in config")
	}

	return cfg, nil
}

// EnsureConfigDir creates the config directory if it doesn't exist
func EnsureConfigDir() (string, error) {
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

// SaveConfig writes the configuration to ~/.taila2a/config.json
func SaveConfig(cfg Config) error {
	configDir, err := EnsureConfigDir()
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
