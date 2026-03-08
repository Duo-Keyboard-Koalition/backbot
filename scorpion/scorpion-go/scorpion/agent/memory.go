package agent

import (
	"sync"
	"time"
)

// MemoryStore manages agent memory.
type MemoryStore struct {
	mu         sync.RWMutex
	shortTerm  []MemoryItem
	longTerm   []MemoryItem
	maxShort   int
	maxLong    int
}

// MemoryItem represents a single memory entry.
type MemoryItem struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

// NewMemoryStore creates a new memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		shortTerm: make([]MemoryItem, 0),
		longTerm:  make([]MemoryItem, 0),
		maxShort:  100,
		maxLong:   1000,
	}
}

// AddShortTerm adds a short-term memory.
func (m *MemoryStore) AddShortTerm(key, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	item := MemoryItem{
		Key:       key,
		Value:     value,
		Timestamp: time.Now(),
	}

	m.shortTerm = append(m.shortTerm, item)

	// Trim if exceeds max
	if len(m.shortTerm) > m.maxShort {
		m.shortTerm = m.shortTerm[len(m.shortTerm)-m.maxShort:]
	}
}

// AddLongTerm adds a long-term memory.
func (m *MemoryStore) AddLongTerm(key, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	item := MemoryItem{
		Key:       key,
		Value:     value,
		Timestamp: time.Now(),
	}

	m.longTerm = append(m.longTerm, item)

	// Trim if exceeds max
	if len(m.longTerm) > m.maxLong {
		m.longTerm = m.longTerm[len(m.longTerm)-m.maxLong:]
	}
}

// GetShortTerm returns recent short-term memories.
func (m *MemoryStore) GetShortTerm(limit int) []MemoryItem {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit > len(m.shortTerm) {
		return append([]MemoryItem(nil), m.shortTerm...)
	}
	return append([]MemoryItem(nil), m.shortTerm[len(m.shortTerm)-limit:]...)
}

// GetLongTerm returns long-term memories.
func (m *MemoryStore) GetLongTerm(limit int) []MemoryItem {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit > len(m.longTerm) {
		return append([]MemoryItem(nil), m.longTerm...)
	}
	return append([]MemoryItem(nil), m.longTerm[len(m.longTerm)-limit:]...)
}

// Search searches memories by keyword.
func (m *MemoryStore) Search(keyword string, limit int) []MemoryItem {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []MemoryItem
	for _, item := range m.shortTerm {
		if contains(item.Key, keyword) || contains(item.Value, keyword) {
			results = append(results, item)
			if len(results) >= limit {
				return results
			}
		}
	}
	for _, item := range m.longTerm {
		if contains(item.Key, keyword) || contains(item.Value, keyword) {
			results = append(results, item)
			if len(results) >= limit {
				return results
			}
		}
	}
	return results
}

// GetContext returns a summary of memory for context.
func (m *MemoryStore) GetContext() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.shortTerm) == 0 {
		return ""
	}

	// Return last few memories as context
	recent := m.shortTerm
	if len(recent) > 5 {
		recent = recent[len(recent)-5:]
	}

	context := "Recent context:\n"
	for _, item := range recent {
		context += "- " + item.Value + "\n"
	}
	return context
}

// Clear clears all memories.
func (m *MemoryStore) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shortTerm = make([]MemoryItem, 0)
	m.longTerm = make([]MemoryItem, 0)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
