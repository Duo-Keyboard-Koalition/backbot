// Package wal provides a write-ahead log for message persistence.
package wal

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Common errors
var (
	ErrNotFound       = errors.New("entry not found")
	ErrCorrupted      = errors.New("log file corrupted")
	ErrInvalidOffset  = errors.New("invalid offset")
	ErrClosed         = errors.New("WAL is closed")
	ErrSegmentFull    = errors.New("segment is full")
)

// WAL configuration
const (
	DefaultSegmentSize    = 64 * 1024 * 1024 // 64MB per segment
	DefaultIndexInterval  = 4096             // Index entry every 4KB
	DefaultSyncInterval   = time.Second
	MaxMessageSize        = 10 * 1024 * 1024 // 10MB max message
)

// Message format:
// ┌─────────────────────────────────────────┐
// │ Offset (8 bytes)                        │
// │ Size (4 bytes)                          │
// │ CRC (4 bytes)                           │
// │ Key Length (4 bytes)                    │
// │ Key (variable)                          │
// │ Value (variable)                        │
// │ Timestamp (8 bytes)                     │
// └─────────────────────────────────────────┘

const (
	headerSize = 8 + 4 + 4 + 4 + 8 // offset + size + crc + keyLen + timestamp = 28
)

// WAL defines the write-ahead log interface
type WAL interface {
	// Append writes a message and returns its offset
	Append(topic string, partition int, data []byte) (int64, error)

	// AppendWithKey writes a message with a key
	AppendWithKey(topic string, partition int, key string, data []byte) (int64, error)

	// Read reads a message at the given offset
	Read(topic string, partition int, offset int64) ([]byte, error)

	// ReadWithMetadata reads a message with full metadata
	ReadWithMetadata(topic string, partition int, offset int64) (*LogEntry, error)

	// ReadRange reads messages in offset range
	ReadRange(topic string, partition int, start, end int64) ([]*LogEntry, error)

	// Truncate removes messages before offset
	Truncate(topic string, partition int, offset int64) error

	// Sync flushes pending writes to disk
	Sync() error

	// Close closes the WAL
	Close() error
}

// LogEntry represents a single log entry
type LogEntry struct {
	Offset    int64
	Key       string
	Value     []byte
	Timestamp time.Time
	CRC       uint32
}

// walImpl implements the WAL interface
type walImpl struct {
	mu          sync.RWMutex
	baseDir     string
	segments    map[string]*segmentManager
	config      *Config
	closed      bool
	crcTable    *crc32.Table
}

// Config configures the WAL
type Config struct {
	SegmentSize   int64
	IndexInterval int
	SyncInterval  time.Duration
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		SegmentSize:   DefaultSegmentSize,
		IndexInterval: DefaultIndexInterval,
		SyncInterval:  DefaultSyncInterval,
	}
}

// New creates a new WAL
func New(baseDir string) (*walImpl, error) {
	return NewWithConfig(baseDir, DefaultConfig())
}

// NewWithConfig creates a new WAL with custom configuration
func NewWithConfig(baseDir string, config *Config) (*walImpl, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create WAL directory: %w", err)
	}

	w := &walImpl{
		baseDir:  baseDir,
		segments: make(map[string]*segmentManager),
		config:   config,
		crcTable: crc32.MakeTable(crc32.Castagnoli),
	}

	// Recover existing segments
	if err := w.recover(); err != nil {
		return nil, fmt.Errorf("failed to recover WAL: %w", err)
	}

	return w, nil
}

// recover loads existing segments
func (w *walImpl) recover() error {
	// List topic directories
	entries, err := os.ReadDir(w.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		topicName := entry.Name()

		// List partition directories
		partitionDir := filepath.Join(w.baseDir, topicName)
		partitionEntries, err := os.ReadDir(partitionDir)
		if err != nil {
			continue
		}

		for _, pEntry := range partitionEntries {
			if !pEntry.IsDir() {
				continue
			}

			var partitionID int
			fmt.Sscanf(pEntry.Name(), "%d", &partitionID)

			// Load segment manager for this partition
			key := fmt.Sprintf("%s:%d", topicName, partitionID)
			sm, err := newSegmentManager(partitionDir, partitionID, w.config)
			if err != nil {
				return fmt.Errorf("failed to load segment for %s: %w", key, err)
			}

			w.segments[key] = sm
		}
	}

	return nil
}

// getSegmentManager gets or creates a segment manager for a topic/partition
func (w *walImpl) getSegmentManager(topic string, partition int) (*segmentManager, error) {
	key := fmt.Sprintf("%s:%d", topic, partition)

	if sm, exists := w.segments[key]; exists {
		return sm, nil
	}

	// Create directory structure
	partitionDir := filepath.Join(w.baseDir, topic, fmt.Sprintf("%d", partition))
	if err := os.MkdirAll(partitionDir, 0755); err != nil {
		return nil, err
	}

	sm, err := newSegmentManager(partitionDir, partition, w.config)
	if err != nil {
		return nil, err
	}

	w.segments[key] = sm
	return sm, nil
}

// Append appends a message to the WAL
func (w *walImpl) Append(topic string, partition int, data []byte) (int64, error) {
	return w.AppendWithKey(topic, partition, "", data)
}

// AppendWithKey appends a message with a key to the WAL
func (w *walImpl) AppendWithKey(topic string, partition int, key string, data []byte) (int64, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return -1, ErrClosed
	}

	sm, err := w.getSegmentManager(topic, partition)
	if err != nil {
		return -1, err
	}

	entry := &LogEntry{
		Key:       key,
		Value:     data,
		Timestamp: time.Now().UTC(),
	}

	offset, err := sm.append(entry)
	if err != nil {
		return -1, err
	}

	entry.Offset = offset

	return offset, nil
}

// Read reads a message at the given offset
func (w *walImpl) Read(topic string, partition int, offset int64) ([]byte, error) {
	entry, err := w.ReadWithMetadata(topic, partition, offset)
	if err != nil {
		return nil, err
	}
	return entry.Value, nil
}

// ReadWithMetadata reads a message with full metadata
func (w *walImpl) ReadWithMetadata(topic string, partition int, offset int64) (*LogEntry, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if w.closed {
		return nil, ErrClosed
	}

	key := fmt.Sprintf("%s:%d", topic, partition)
	sm, exists := w.segments[key]
	if !exists {
		return nil, ErrNotFound
	}

	return sm.read(offset)
}

// ReadRange reads messages in an offset range
func (w *walImpl) ReadRange(topic string, partition int, start, end int64) ([]*LogEntry, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if w.closed {
		return nil, ErrClosed
	}

	key := fmt.Sprintf("%s:%d", topic, partition)
	sm, exists := w.segments[key]
	if !exists {
		return nil, ErrNotFound
	}

	return sm.readRange(start, end)
}

// Truncate removes messages before the given offset
func (w *walImpl) Truncate(topic string, partition int, offset int64) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return ErrClosed
	}

	key := fmt.Sprintf("%s:%d", topic, partition)
	sm, exists := w.segments[key]
	if !exists {
		return nil
	}

	return sm.truncate(offset)
}

// Sync flushes all pending writes to disk
func (w *walImpl) Sync() error {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if w.closed {
		return ErrClosed
	}

	for _, sm := range w.segments {
		if err := sm.sync(); err != nil {
			return err
		}
	}

	return nil
}

// Close closes the WAL
func (w *walImpl) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return nil
	}

	w.closed = true

	var lastErr error
	for _, sm := range w.segments {
		if err := sm.close(); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// computeCRC computes CRC32 checksum
func (w *walImpl) computeCRC(data []byte) uint32 {
	return crc32.Checksum(data, w.crcTable)
}

// verifyCRC verifies CRC32 checksum
func (w *walImpl) verifyCRC(data []byte, expected uint32) bool {
	return crc32.Checksum(data, w.crcTable) == expected
}

// encodeEntry encodes a log entry to bytes
func (w *walImpl) encodeEntry(entry *LogEntry) ([]byte, error) {
	keyBytes := []byte(entry.Key)
	keyLen := uint32(len(keyBytes))
	valueLen := uint32(len(entry.Value))

	// Calculate total size
	totalSize := headerSize + int(keyLen) + int(valueLen)
	buf := make([]byte, totalSize)

	pos := 0

	// Offset (8 bytes)
	binary.BigEndian.PutUint64(buf[pos:], uint64(entry.Offset))
	pos += 8

	// Size (4 bytes)
	binary.BigEndian.PutUint32(buf[pos:], valueLen)
	pos += 4

	// CRC will be computed after filling the rest
	crcPos := pos
	pos += 4

	// Key length (4 bytes)
	binary.BigEndian.PutUint32(buf[pos:], keyLen)
	pos += 4

	// Timestamp (8 bytes)
	binary.BigEndian.PutUint64(buf[pos:], uint64(entry.Timestamp.UnixNano()))
	pos += 8

	// Key (variable)
	copy(buf[pos:], keyBytes)
	pos += int(keyLen)

	// Value (variable)
	copy(buf[pos:], entry.Value)
	pos += int(valueLen)

	// Compute CRC over everything after the CRC field
	crc := w.computeCRC(buf[crcPos+4:])
	binary.BigEndian.PutUint32(buf[crcPos:], crc)

	return buf, nil
}

// decodeEntry decodes a log entry from bytes
func (w *walImpl) decodeEntry(data []byte) (*LogEntry, error) {
	if len(data) < headerSize {
		return nil, ErrCorrupted
	}

	pos := 0

	// Offset (8 bytes)
	offset := int64(binary.BigEndian.Uint64(data[pos:]))
	pos += 8

	// Size (4 bytes)
	valueLen := binary.BigEndian.Uint32(data[pos:])
	pos += 4

	// CRC (4 bytes)
	storedCRC := binary.BigEndian.Uint32(data[pos:])
	pos += 4

	// Key length (4 bytes)
	keyLen := binary.BigEndian.Uint32(data[pos:])
	pos += 4

	// Timestamp (8 bytes)
	timestampNano := binary.BigEndian.Uint64(data[pos:])
	timestamp := time.Unix(0, int64(timestampNano)).UTC()
	pos += 8

	// Key (variable)
	key := string(data[pos : pos+int(keyLen)])
	pos += int(keyLen)

	// Value (variable)
	value := data[pos : pos+int(valueLen)]

	// Verify CRC (computed over everything after the CRC field: keyLen + timestamp + key + value)
	computedCRC := w.computeCRC(data[16:]) // Skip offset(8) + size(4) + crc(4) = 16
	if computedCRC != storedCRC {
		return nil, ErrCorrupted
	}

	return &LogEntry{
		Offset:    offset,
		Key:       key,
		Value:     value,
		Timestamp: timestamp,
		CRC:       storedCRC,
	}, nil
}
