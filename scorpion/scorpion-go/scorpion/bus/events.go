package bus

import (
	"time"
)

// InboundMessage represents a message received from a chat channel.
type InboundMessage struct {
	Channel            string            `json:"channel"`
	SenderID           string            `json:"sender_id"`
	ChatID             string            `json:"chat_id"`
	Content            string            `json:"content"`
	Timestamp          time.Time         `json:"timestamp"`
	Media              []string          `json:"media,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	SessionKeyOverride string            `json:"session_key_override,omitempty"`
}

// SessionKey returns the unique key for session identification.
func (m *InboundMessage) SessionKey() string {
	if m.SessionKeyOverride != "" {
		return m.SessionKeyOverride
	}
	return m.Channel + ":" + m.ChatID
}

// OutboundMessage represents a message to send to a chat channel.
type OutboundMessage struct {
	Channel   string                 `json:"channel"`
	ChatID    string                 `json:"chat_id"`
	Content   string                 `json:"content"`
	ReplyTo   string                 `json:"reply_to,omitempty"`
	Media     []string               `json:"media,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NewInboundMessage creates a new inbound message.
func NewInboundMessage(channel, senderID, chatID, content string) *InboundMessage {
	return &InboundMessage{
		Channel:   channel,
		SenderID:  senderID,
		ChatID:    chatID,
		Content:   content,
		Timestamp: time.Now(),
		Media:     make([]string, 0),
		Metadata:  make(map[string]interface{}),
	}
}

// NewOutboundMessage creates a new outbound message.
func NewOutboundMessage(channel, chatID, content string) *OutboundMessage {
	return &OutboundMessage{
		Channel:  channel,
		ChatID:   chatID,
		Content:  content,
		Media:    make([]string, 0),
		Metadata: make(map[string]interface{}),
	}
}

// WithReplyTo sets the reply-to message ID.
func (m *OutboundMessage) WithReplyTo(replyTo string) *OutboundMessage {
	m.ReplyTo = replyTo
	return m
}

// WithMedia adds media URLs to the message.
func (m *OutboundMessage) WithMedia(media ...string) *OutboundMessage {
	m.Media = append(m.Media, media...)
	return m
}

// WithMetadata adds metadata to the message.
func (m *OutboundMessage) WithMetadata(key string, value interface{}) *OutboundMessage {
	m.Metadata[key] = value
	return m
}
