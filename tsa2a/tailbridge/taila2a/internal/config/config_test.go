package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Service.Name != "taila2a" {
		t.Errorf("Expected service name 'taila2a', got %s", cfg.Service.Name)
	}

	if cfg.Service.HTTPPort != DefaultHTTPPort {
		t.Errorf("Expected HTTP port %d, got %d", DefaultHTTPPort, cfg.Service.HTTPPort)
	}

	if cfg.EventBus.DataDir != DefaultEventBusDir {
		t.Errorf("Expected eventbus dir %s, got %s", DefaultEventBusDir, cfg.EventBus.DataDir)
	}

	if cfg.EventBus.WAL.Dir != DefaultWALDir {
		t.Errorf("Expected WAL dir %s, got %s", DefaultWALDir, cfg.EventBus.WAL.Dir)
	}

	if cfg.TailFS.DataDir != DefaultTailFSDir {
		t.Errorf("Expected TailFS dir %s, got %s", DefaultTailFSDir, cfg.TailFS.DataDir)
	}

	if cfg.TailFS.ChunkDir != DefaultTailFSChunkDir {
		t.Errorf("Expected TailFS chunk dir %s, got %s", DefaultTailFSChunkDir, cfg.TailFS.ChunkDir)
	}
}

func TestLoadConfig(t *testing.T) {
	// Create temp config file
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")

	cfg := DefaultConfig()
	cfg.Service.AgentID = "test-agent"
	cfg.Service.HTTPPort = 9000

	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load config
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loaded.Service.AgentID != "test-agent" {
		t.Errorf("Expected agent ID 'test-agent', got %s", loaded.Service.AgentID)
	}

	if loaded.Service.HTTPPort != 9000 {
		t.Errorf("Expected HTTP port 9000, got %d", loaded.Service.HTTPPort)
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := Load("nonexistent.json")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestConfigValidate(t *testing.T) {
	cfg := DefaultConfig()

	// Valid config
	if err := cfg.Validate(); err != nil {
		t.Errorf("Expected no error for valid config: %v", err)
	}

	// Invalid port
	cfg.Service.HTTPPort = 0
	if err := cfg.Validate(); err != ErrInvalidPort {
		t.Errorf("Expected ErrInvalidPort, got %v", err)
	}
	cfg.Service.HTTPPort = DefaultHTTPPort

	// Invalid partitions
	cfg.EventBus.DefaultPartitions = 0
	if err := cfg.Validate(); err != ErrInvalidPartitions {
		t.Errorf("Expected ErrInvalidPartitions, got %v", err)
	}
	cfg.EventBus.DefaultPartitions = DefaultDefaultPartitions

	// Invalid segment size
	cfg.EventBus.WAL.SegmentSize = 100
	if err := cfg.Validate(); err != ErrInvalidSegmentSize {
		t.Errorf("Expected ErrInvalidSegmentSize, got %v", err)
	}
	cfg.EventBus.WAL.SegmentSize = DefaultSegmentSize
}

func TestConfigApplyDefaults(t *testing.T) {
	// Create config with zero values
	cfg := &Config{}
	cfg.applyDefaults()

	if cfg.Service.Name != "taila2a" {
		t.Errorf("Expected default service name")
	}

	if cfg.Service.HTTPPort != DefaultHTTPPort {
		t.Errorf("Expected default HTTP port")
	}

	if cfg.EventBus.DataDir != DefaultEventBusDir {
		t.Errorf("Expected default eventbus dir")
	}

	if cfg.TailFS.DataDir != DefaultTailFSDir {
		t.Errorf("Expected default TailFS dir")
	}
}

func TestConfigSaveLoad(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")

	original := DefaultConfig()
	original.Service.AgentID = "save-test"
	original.EventBus.WAL.Enabled = true
	original.TailFS.CompressionEnabled = true

	// Save
	if err := original.Save(configPath); err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Load
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	// Verify values
	if loaded.Service.AgentID != "save-test" {
		t.Errorf("Agent ID not preserved")
	}

	if !loaded.EventBus.WAL.Enabled {
		t.Errorf("WAL enabled not preserved")
	}

	if !loaded.TailFS.CompressionEnabled {
		t.Errorf("Compression enabled not preserved")
	}
}

func TestConfigDurationFields(t *testing.T) {
	cfg := DefaultConfig()

	// Verify duration fields have correct defaults
	if cfg.EventBus.ConsumerGroup.SessionTimeout != 30*time.Second {
		t.Errorf("Expected 30s session timeout")
	}

	if cfg.EventBus.ConsumerGroup.HeartbeatInterval != 3*time.Second {
		t.Errorf("Expected 3s heartbeat interval")
	}

	if cfg.EventBus.WAL.SyncInterval != 1*time.Second {
		t.Errorf("Expected 1s sync interval")
	}
}

func TestConfigTailFSDefaults(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.TailFS.MaxChunkSize != 4*1024*1024 {
		t.Errorf("Expected 4MB chunk size, got %d", cfg.TailFS.MaxChunkSize)
	}

	if cfg.TailFS.ChunkDir != DefaultTailFSChunkDir {
		t.Errorf("Expected chunk dir %s, got %s", DefaultTailFSChunkDir, cfg.TailFS.ChunkDir)
	}

	if cfg.TailFS.IncomingDir != DefaultTailFSIncomingDir {
		t.Errorf("Expected incoming dir %s, got %s", DefaultTailFSIncomingDir, cfg.TailFS.IncomingDir)
	}
}
