// Package config provides configuration for the taila2a service.
package config

import (
	"encoding/json"
	"os"
	"time"
)

// Default configuration values
const (
	DefaultEventBusDir        = "data/eventbus"
	DefaultWALDir             = "data/eventbus/wal"
	DefaultTailFSDir          = "data/tailfs"
	DefaultTailFSChunkDir     = "data/tailfs/chunks"
	DefaultTailFSIncomingDir  = "data/tailfs/incoming"
	DefaultHTTPPort           = 8001
	DefaultDefaultPartitions  = 4
	DefaultWALEnabled         = true
	DefaultSyncInterval       = time.Second
	DefaultSegmentSize        = 64 * 1024 * 1024 // 64MB
	DefaultSessionTimeout     = 30 * time.Second
	DefaultHeartbeatInterval  = 3 * time.Second
)

// Config holds the complete configuration for taila2a
type Config struct {
	// Service configuration
	Service ServiceConfig `json:"service"`

	// EventBus configuration
	EventBus EventBusConfig `json:"eventbus"`

	// TailFS configuration
	TailFS TailFSConfig `json:"tailfs"`

	// Tailscale configuration
	Tailscale TailscaleConfig `json:"tailscale"`

	// Logging configuration
	Logging LoggingConfig `json:"logging"`
}

// ServiceConfig holds service-level configuration
type ServiceConfig struct {
	// Name of the service
	Name string `json:"name"`

	// AgentID is the unique identifier for this agent
	AgentID string `json:"agent_id"`

	// HTTPPort is the port for the HTTP server
	HTTPPort int `json:"http_port"`

	// GRPCPort is the port for gRPC server (optional)
	GRPCPort int `json:"grpc_port,omitempty"`
}

// EventBusConfig holds event bus configuration
type EventBusConfig struct {
	// Enabled enables the event bus
	Enabled bool `json:"enabled"`

	// DataDir is the directory for event bus data
	DataDir string `json:"data_dir"`

	// WAL configuration
	WAL WALConfig `json:"wal"`

	// Consumer group configuration
	ConsumerGroup ConsumerGroupConfig `json:"consumer_group"`

	// DefaultPartitions is the default number of partitions for new topics
	DefaultPartitions int `json:"default_partitions"`
}

// WALConfig holds write-ahead log configuration
type WALConfig struct {
	// Enabled enables WAL persistence
	Enabled bool `json:"enabled"`

	// Dir is the directory for WAL storage
	Dir string `json:"dir"`

	// SegmentSize is the maximum size of a segment file
	SegmentSize int64 `json:"segment_size"`

	// SyncInterval is how often to sync WAL to disk
	SyncInterval time.Duration `json:"sync_interval"`

	// SyncIntervalStr is the string representation for JSON
	SyncIntervalStr string `json:"sync_interval_str,omitempty"`
}

// ConsumerGroupConfig holds consumer group configuration
type ConsumerGroupConfig struct {
	// SessionTimeout is the timeout for consumer sessions
	SessionTimeout time.Duration `json:"session_timeout"`

	// SessionTimeoutStr is the string representation for JSON
	SessionTimeoutStr string `json:"session_timeout_str,omitempty"`

	// HeartbeatInterval is the interval between heartbeats
	HeartbeatInterval time.Duration `json:"heartbeat_interval"`

	// HeartbeatIntervalStr is the string representation for JSON
	HeartbeatIntervalStr string `json:"heartbeat_interval_str,omitempty"`
}

// TailFSConfig holds TailFS configuration
type TailFSConfig struct {
	// Enabled enables TailFS
	Enabled bool `json:"enabled"`

	// DataDir is the base directory for TailFS data
	DataDir string `json:"data_dir"`

	// ChunkDir is the directory for file chunks
	ChunkDir string `json:"chunk_dir"`

	// IncomingDir is the directory for incoming files
	IncomingDir string `json:"incoming_dir"`

	// MaxChunkSize is the maximum size of a file chunk
	MaxChunkSize int64 `json:"max_chunk_size"`

	// CompressionEnabled enables compression
	CompressionEnabled bool `json:"compression_enabled"`

	// EncryptionEnabled enables encryption
	EncryptionEnabled bool `json:"encryption_enabled"`
}

// TailscaleConfig holds Tailscale configuration
type TailscaleConfig struct {
	// Enabled enables Tailscale integration
	Enabled bool `json:"enabled"`

	// AuthKey is the Tailscale auth key (optional, can use env var)
	AuthKey string `json:"auth_key,omitempty"`

	// ControlURL is the Tailscale control server URL (optional)
	ControlURL string `json:"control_url,omitempty"`

	// Hostname is the hostname for this node
	Hostname string `json:"hostname,omitempty"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	// Level is the log level (debug, info, warn, error)
	Level string `json:"level"`

	// Format is the log format (json, text)
	Format string `json:"format"`

	// Output is the log output (stdout, stderr, file)
	Output string `json:"output"`

	// File is the log file path (if output is file)
	File string `json:"file,omitempty"`
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	return &Config{
		Service: ServiceConfig{
			Name:       "taila2a",
			AgentID:    "",
			HTTPPort:   DefaultHTTPPort,
			GRPCPort:   0,
		},
		EventBus: EventBusConfig{
			Enabled:           true,
			DataDir:           DefaultEventBusDir,
			DefaultPartitions: DefaultDefaultPartitions,
			WAL: WALConfig{
				Enabled:       DefaultWALEnabled,
				Dir:           DefaultWALDir,
				SegmentSize:   DefaultSegmentSize,
				SyncInterval:  DefaultSyncInterval,
			},
			ConsumerGroup: ConsumerGroupConfig{
				SessionTimeout:    DefaultSessionTimeout,
				HeartbeatInterval: DefaultHeartbeatInterval,
			},
		},
		TailFS: TailFSConfig{
			Enabled:            true,
			DataDir:            DefaultTailFSDir,
			ChunkDir:           DefaultTailFSChunkDir,
			IncomingDir:        DefaultTailFSIncomingDir,
			MaxChunkSize:       4 * 1024 * 1024, // 4MB
			CompressionEnabled: false,
			EncryptionEnabled:  false,
		},
		Tailscale: TailscaleConfig{
			Enabled: true,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		},
	}
}

// Load loads configuration from a file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := DefaultConfig()
	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	// Apply defaults for zero values
	config.applyDefaults()

	return config, nil
}

// Save saves configuration to a file
func (c *Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// applyDefaults applies default values for any zero-value fields
func (c *Config) applyDefaults() {
	// Service defaults
	if c.Service.Name == "" {
		c.Service.Name = "taila2a"
	}
	if c.Service.HTTPPort == 0 {
		c.Service.HTTPPort = DefaultHTTPPort
	}

	// EventBus defaults
	if c.EventBus.DataDir == "" {
		c.EventBus.DataDir = DefaultEventBusDir
	}
	if c.EventBus.DefaultPartitions == 0 {
		c.EventBus.DefaultPartitions = DefaultDefaultPartitions
	}

	// WAL defaults
	if c.EventBus.WAL.Dir == "" {
		c.EventBus.WAL.Dir = DefaultWALDir
	}
	if c.EventBus.WAL.SegmentSize == 0 {
		c.EventBus.WAL.SegmentSize = DefaultSegmentSize
	}
	if c.EventBus.WAL.SyncInterval == 0 {
		c.EventBus.WAL.SyncInterval = DefaultSyncInterval
	}

	// Consumer group defaults
	if c.EventBus.ConsumerGroup.SessionTimeout == 0 {
		c.EventBus.ConsumerGroup.SessionTimeout = DefaultSessionTimeout
	}
	if c.EventBus.ConsumerGroup.HeartbeatInterval == 0 {
		c.EventBus.ConsumerGroup.HeartbeatInterval = DefaultHeartbeatInterval
	}

	// TailFS defaults
	if c.TailFS.DataDir == "" {
		c.TailFS.DataDir = DefaultTailFSDir
	}
	if c.TailFS.ChunkDir == "" {
		c.TailFS.ChunkDir = DefaultTailFSChunkDir
	}
	if c.TailFS.IncomingDir == "" {
		c.TailFS.IncomingDir = DefaultTailFSIncomingDir
	}
	if c.TailFS.MaxChunkSize == 0 {
		c.TailFS.MaxChunkSize = 4 * 1024 * 1024
	}

	// Logging defaults
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.Format == "" {
		c.Logging.Format = "text"
	}
	if c.Logging.Output == "" {
		c.Logging.Output = "stdout"
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate service config
	if c.Service.HTTPPort < 1 || c.Service.HTTPPort > 65535 {
		return ErrInvalidPort
	}

	// Validate eventbus config
	if c.EventBus.DefaultPartitions < 1 {
		return ErrInvalidPartitions
	}

	// Validate WAL config
	if c.EventBus.WAL.SegmentSize < 1024*1024 {
		return ErrInvalidSegmentSize
	}

	// Validate TailFS config
	if c.TailFS.MaxChunkSize < 1024 {
		return ErrInvalidChunkSize
	}

	return nil
}

// Common errors
var (
	ErrInvalidPort        = &ConfigError{"invalid port number"}
	ErrInvalidPartitions  = &ConfigError{"invalid number of partitions"}
	ErrInvalidSegmentSize = &ConfigError{"segment size too small"}
	ErrInvalidChunkSize   = &ConfigError{"chunk size too small"}
)

// ConfigError represents a configuration error
type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string {
	return e.Message
}
