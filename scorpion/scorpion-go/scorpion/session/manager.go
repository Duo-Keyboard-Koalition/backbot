package session

import (
	"sync"
	"time"

	"scorpion-go/internal/adk"
)

// Session represents a user session.
type Session struct {
	ID           string      `json:"id"`
	Channel      string      `json:"channel"`
	ChatID       string      `json:"chat_id"`
	SenderID     string      `json:"sender_id"`
	Context      []adk.Message `json:"context"`
	Memory       SessionMemory `json:"memory"`
	CreatedAt    time.Time   `json:"created_at"`
	LastActiveAt time.Time   `json:"last_active_at"`
	MessageCount int         `json:"message_count"`
}

// SessionMemory holds session-specific memory.
type SessionMemory struct {
	ShortTerm []string `json:"short_term"`
	LongTerm  []string `json:"long_term"`
	Summary   string   `json:"summary,omitempty"`
}

// SessionManager manages user sessions.
type SessionManager struct {
	mu          sync.RWMutex
	sessions    map[string]*Session
	maxAge      time.Duration
	maxSessions int
}

// SessionManagerConfig holds session manager configuration.
type SessionManagerConfig struct {
	MaxAge      time.Duration `json:"max_age"`
	MaxSessions int           `json:"max_sessions"`
}

// NewSessionManager creates a new session manager.
func NewSessionManager(config *SessionManagerConfig) *SessionManager {
	if config == nil {
		config = &SessionManagerConfig{}
	}

	maxAge := config.MaxAge
	if maxAge <= 0 {
		maxAge = 24 * time.Hour
	}

	maxSessions := config.MaxSessions
	if maxSessions <= 0 {
		maxSessions = 1000
	}

	return &SessionManager{
		sessions:    make(map[string]*Session),
		maxAge:      maxAge,
		maxSessions: maxSessions,
	}
}

// GetOrCreate gets an existing session or creates a new one.
func (sm *SessionManager) GetOrCreate(channel, chatID, senderID string) *Session {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	key := sm.makeKey(channel, chatID)

	if session, ok := sm.sessions[key]; ok {
		session.LastActiveAt = time.Now()
		return session
	}

	session := &Session{
		ID:           key,
		Channel:      channel,
		ChatID:       chatID,
		SenderID:     senderID,
		Context:      make([]adk.Message, 0),
		Memory:       SessionMemory{ShortTerm: make([]string, 0), LongTerm: make([]string, 0)},
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
		MessageCount: 0,
	}

	sm.sessions[key] = session
	sm.evictIfNeeded()

	return session
}

// Get gets a session by key.
func (sm *SessionManager) Get(key string) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, ok := sm.sessions[key]
	return session, ok
}

// GetByChannelChatID gets a session by channel and chat ID.
func (sm *SessionManager) GetByChannelChatID(channel, chatID string) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	key := sm.makeKey(channel, chatID)
	session, ok := sm.sessions[key]
	return session, ok
}

// AddMessage adds a message to a session's context.
func (sm *SessionManager) AddMessage(key, role, content string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, ok := sm.sessions[key]
	if !ok {
		return ErrSessionNotFound
	}

	session.Context = append(session.Context, adk.Message{
		Role:    role,
		Content: content,
	})
	session.MessageCount++
	session.LastActiveAt = time.Now()

	return nil
}

// AddToShortTerm adds a memory to the session's short-term memory.
func (sm *SessionManager) AddToShortTerm(key, memory string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, ok := sm.sessions[key]
	if !ok {
		return ErrSessionNotFound
	}

	session.Memory.ShortTerm = append(session.Memory.ShortTerm, memory)
	return nil
}

// AddToLongTerm adds a memory to the session's long-term memory.
func (sm *SessionManager) AddToLongTerm(key, memory string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, ok := sm.sessions[key]
	if !ok {
		return ErrSessionNotFound
	}

	session.Memory.LongTerm = append(session.Memory.LongTerm, memory)
	return nil
}

// ClearContext clears a session's context.
func (sm *SessionManager) ClearContext(key string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, ok := sm.sessions[key]
	if !ok {
		return ErrSessionNotFound
	}

	session.Context = make([]adk.Message, 0)
	return nil
}

// Delete deletes a session.
func (sm *SessionManager) Delete(key string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, key)
}

// List lists all sessions.
func (sm *SessionManager) List() []*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessions := make([]*Session, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// Count returns the number of sessions.
func (sm *SessionManager) Count() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.sessions)
}

// Cleanup cleans up expired sessions.
func (sm *SessionManager) Cleanup() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	for key, session := range sm.sessions {
		if now.Sub(session.LastActiveAt) > sm.maxAge {
			delete(sm.sessions, key)
		}
	}
}

// makeKey creates a session key from channel and chat ID.
func (sm *SessionManager) makeKey(channel, chatID string) string {
	return channel + ":" + chatID
}

// evictIfNeeded evicts old sessions if the limit is reached.
func (sm *SessionManager) evictIfNeeded() {
	if len(sm.sessions) <= sm.maxSessions {
		return
	}

	var oldestKey string
	var oldestTime time.Time

	for key, session := range sm.sessions {
		if oldestKey == "" || session.LastActiveAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = session.LastActiveAt
		}
	}

	if oldestKey != "" {
		delete(sm.sessions, oldestKey)
	}
}

// Session errors
var (
	ErrSessionNotFound = &sessionError{"session not found"}
)

type sessionError struct {
	message string
}

func (e *sessionError) Error() string {
	return e.message
}
