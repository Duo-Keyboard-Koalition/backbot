package buffer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// PersistentStore provides file-based persistence for buffered messages
type PersistentStore struct {
	// dataDir is the directory where message files are stored
	dataDir string

	// mu protects concurrent access to the store
	mu sync.RWMutex

	// index maps message IDs to their file paths
	index map[string]string
}

// NewPersistentStore creates a new persistent message store
func NewPersistentStore(dataDir string) (*PersistentStore, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	store := &PersistentStore{
		dataDir: dataDir,
		index:   make(map[string]string),
	}

	if err := store.rebuildIndex(); err != nil {
		return nil, fmt.Errorf("failed to rebuild index: %w", err)
	}

	return store, nil
}

// rebuildIndex scans the data directory and rebuilds the message index
func (s *PersistentStore) rebuildIndex() error {
	entries, err := os.ReadDir(s.dataDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(s.dataDir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}

		s.index[msg.ID] = filePath
	}

	return nil
}

// Save persists a message to disk
func (s *PersistentStore) Save(msg *Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	fileName := fmt.Sprintf("%s.json", msg.ID)
	filePath := filepath.Join(s.dataDir, fileName)

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write message file: %w", err)
	}

	s.index[msg.ID] = filePath
	return nil
}

// Delete removes a message from the store
func (s *PersistentStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filePath, exists := s.index[id]
	if !exists {
		return fmt.Errorf("message not found: %s", id)
	}

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to remove message file: %w", err)
	}

	delete(s.index, id)
	return nil
}

// Get retrieves a message by ID
func (s *PersistentStore) Get(id string) (*Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filePath, exists := s.index[id]
	if !exists {
		return nil, fmt.Errorf("message not found: %s", id)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read message file: %w", err)
	}

	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return &msg, nil
}

// GetAll retrieves all messages from the store
func (s *PersistentStore) GetAll() ([]*Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	messages := make([]*Message, 0, len(s.index))
	for id := range s.index {
		msg, err := s.Get(id)
		if err != nil {
			continue
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// GetByStatus retrieves all messages with a specific status
func (s *PersistentStore) GetByStatus(status MessageStatus) ([]*Message, error) {
	all, err := s.GetAll()
	if err != nil {
		return nil, err
	}

	filtered := make([]*Message, 0)
	for _, msg := range all {
		if msg.Status == status {
			filtered = append(filtered, msg)
		}
	}

	return filtered, nil
}

// GetDueForRetry retrieves all messages that are due for retry
func (s *PersistentStore) GetDueForRetry(now time.Time) ([]*Message, error) {
	retrying, err := s.GetByStatus(StatusRetrying)
	if err != nil {
		return nil, err
	}

	due := make([]*Message, 0)
	for _, msg := range retrying {
		if !msg.NextRetryAt.IsZero() && msg.NextRetryAt.Before(now) {
			due = append(due, msg)
		}
	}

	return due, nil
}

// Count returns the total number of messages in the store
func (s *PersistentStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.index)
}

// Stats returns statistics about the store
func (s *PersistentStore) Stats() BufferStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := BufferStats{}
	messages, err := s.getAllLocked()
	if err != nil {
		return stats
	}

	stats.TotalMessages = len(messages)

	var oldest, newest time.Time
	for _, msg := range messages {
		switch msg.Status {
		case StatusPending:
			stats.PendingCount++
		case StatusRetrying:
			stats.RetryingCount++
		case StatusFailed:
			stats.FailedCount++
		case StatusDelivered:
			stats.DeliveredCount++
		}

		if oldest.IsZero() || msg.CreatedAt.Before(oldest) {
			oldest = msg.CreatedAt
		}
		if newest.IsZero() || msg.CreatedAt.After(newest) {
			newest = msg.CreatedAt
		}
	}

	stats.OldestMessage = oldest
	stats.NewestMessage = newest

	return stats
}

// getAllLocked retrieves all messages (caller must hold read lock)
func (s *PersistentStore) getAllLocked() ([]*Message, error) {
	messages := make([]*Message, 0, len(s.index))
	for id := range s.index {
		msg, err := s.Get(id)
		if err != nil {
			continue
		}
		messages = append(messages, msg)
	}
	return messages, nil
}
