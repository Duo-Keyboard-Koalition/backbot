package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// DefaultConfigPath is the default configuration file path.
const DefaultConfigPath = "~/.scorpion-go/config.json"

// DefaultWorkspacePath is the default workspace path.
const DefaultWorkspacePath = "~/.scorpion-go/workspace"

// LoadConfig loads the configuration from the default or specified path.
func LoadConfig(path string) (*Config, error) {
	if path == "" {
		path = DefaultConfigPath
	}

	// Expand tilde
	expanded, err := expandTilde(path)
	if err != nil {
		return nil, fmt.Errorf("failed to expand path: %w", err)
	}

	// Read file
	data, err := os.ReadFile(expanded)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Apply defaults
	config.applyDefaults()

	return &config, nil
}

// SaveConfig saves the configuration to the specified path.
func SaveConfig(config *Config, path string) error {
	if path == "" {
		path = DefaultConfigPath
	}

	// Expand tilde
	expanded, err := expandTilde(path)
	if err != nil {
		return fmt.Errorf("failed to expand path: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(expanded)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write file
	if err := os.WriteFile(expanded, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetConfigPath returns the configuration file path.
func GetConfigPath() (string, error) {
	return expandTilde(DefaultConfigPath)
}

// GetWorkspacePath returns the workspace path.
func GetWorkspacePath() (string, error) {
	return expandTilde(DefaultWorkspacePath)
}

// DefaultConfig returns a default configuration.
func DefaultConfig() *Config {
	return &Config{
		Providers: ProvidersConfig{
			Gemini: &GeminiConfig{
				Model: "gemini-2.5-flash",
			},
		},
		Channels: ChannelsConfig{},
		Tools: ToolsConfig{
			MCPServers:        make(map[string]MCPServerConfig),
			RestrictToWorkspace: false,
			ToolTimeout:       30,
		},
	}
}

// applyDefaults applies default values to the configuration.
func (c *Config) applyDefaults() {
	if c.Providers.Gemini == nil {
		c.Providers.Gemini = &GeminiConfig{}
	}
	if c.Providers.Gemini.Model == "" {
		c.Providers.Gemini.Model = "gemini-2.5-flash"
	}
	if c.Tools.MCPServers == nil {
		c.Tools.MCPServers = make(map[string]MCPServerConfig)
	}
	if c.Tools.ToolTimeout == 0 {
		c.Tools.ToolTimeout = 30
	}
}

// expandTilde expands tilde in the path to the home directory.
func expandTilde(path string) (string, error) {
	if path == "" {
		return "", nil
	}

	if path[0] != '~' {
		return path, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	if len(path) == 1 {
		return homeDir, nil
	}

	if path[1] != '/' {
		return path, nil // Invalid tilde path
	}

	return filepath.Join(homeDir, path[2:]), nil
}

// EnsureDir ensures the directory exists.
func EnsureDir(path string) error {
	expanded, err := expandTilde(path)
	if err != nil {
		return err
	}

	return os.MkdirAll(expanded, 0755)
}
