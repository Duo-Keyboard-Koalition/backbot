package adk

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// MessageStore stores messages for delivery.
type MessageStore struct {
	mu       sync.Mutex
	messages []Message
}

func NewMessageStore() *MessageStore {
	return &MessageStore{messages: make([]Message, 0)}
}

func (s *MessageStore) Add(msg Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = append(s.messages, msg)
}

func (s *MessageStore) GetAndClear() []Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	msgs := s.messages
	s.messages = make([]Message, 0)
	return msgs
}

// MessageTool sends messages to users.
type MessageTool struct {
	Store *MessageStore
}

func (MessageTool) Name() string { return "message" }
func (MessageTool) Description() string {
	return "Send a message to the user. Use for proactive communication."
}
func (t MessageTool) Run(ctx context.Context, input map[string]string) (string, error) {
	content := strings.TrimSpace(input["content"])
	if content == "" {
		return "", fmt.Errorf("content is required")
	}

	if t.Store != nil {
		t.Store.Add(Message{Role: "assistant", Content: content})
	}

	return "Message sent successfully", nil
}
