package channels

import (
	"context"
	"fmt"
	"sync"

	"darci-go/darci/bus"
)

// SlackChannel implements the Slack chat channel using Socket Mode.
type SlackChannel struct {
	*BaseChannelImpl
	mu     sync.RWMutex
	config *SlackConfig
	stopCh chan struct{}
}

// SlackConfig holds Slack channel configuration.
type SlackConfig struct {
	ChannelConfig
	Mode               string   `json:"mode"` // "socket" supported
	WebhookPath        string   `json:"webhook_path"`
	BotToken           string   `json:"bot_token"`
	AppToken           string   `json:"app_token"`
	UserTokenReadOnly  bool     `json:"user_token_read_only"`
	ReplyInThread      bool     `json:"reply_in_thread"`
	ReactEmoji         string   `json:"react_emoji"`
	GroupPolicy        string   `json:"group_policy"` // "mention", "open", "allowlist"
	GroupAllowFrom     []string `json:"group_allow_from"`
	DM                 SlackDMConfig `json:"dm"`
}

// SlackDMConfig holds Slack DM policy configuration.
type SlackDMConfig struct {
	Enabled   bool     `json:"enabled"`
	Policy    string   `json:"policy"` // "open" or "allowlist"
	AllowFrom []string `json:"allow_from"`
}

// NewSlackChannel creates a new Slack channel.
func NewSlackChannel(config *SlackConfig, messageBus *bus.MessageBus) *SlackChannel {
	return &SlackChannel{
		BaseChannelImpl: NewBaseChannel("slack", &config.ChannelConfig, messageBus),
		config:          config,
		stopCh:          make(chan struct{}),
	}
}

// Initialize initializes the Slack channel.
func (s *SlackChannel) Initialize(ctx context.Context) error {
	if s.config.BotToken == "" {
		return fmt.Errorf("slack bot token is required")
	}

	if s.config.AppToken == "" {
		return fmt.Errorf("slack app token is required for socket mode")
	}

	// In a full implementation, this would:
	// 1. Create a Slack client with Socket Mode
	// 2. Set up event handlers
	// 3. Configure group policy

	return nil
}

// Start starts the Slack channel (blocking).
func (s *SlackChannel) Start(ctx context.Context) error {
	s.SetRunning(true)
	defer s.SetRunning(false)

	// In a full implementation, this would:
	// 1. Start Socket Mode listener
	// 2. Handle events (messages, mentions, reactions)
	// 3. Convert to InboundMessage and publish to bus
	// 4. Listen to bus for OutboundMessage and send to Slack

	<-s.stopCh
	return nil
}

// Stop stops the Slack channel.
func (s *SlackChannel) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.IsRunning() {
		return nil
	}

	close(s.stopCh)
	return nil
}

// Send sends a message through the Slack channel.
func (s *SlackChannel) Send(ctx context.Context, msg *bus.OutboundMessage) error {
	if !s.IsRunning() {
		return fmt.Errorf("channel not running")
	}

	// In a full implementation, this would:
	// 1. Use the Slack API to send the message
	// 2. Handle thread replies
	// 3. Handle media attachments
	// 4. Add reactions

	return nil
}

// ShouldRespondToMessage checks if the bot should respond to a message based on group policy.
func (s *SlackChannel) ShouldRespondToMessage(channelID, messageType string, mentionsBot bool) bool {
	// DM policy
	if s.config.DM.Enabled && messageType == "dm" {
		return true
	}

	// Group policy
	switch s.config.GroupPolicy {
	case "open":
		return true
	case "mention":
		return mentionsBot
	case "allowlist":
		for _, allowed := range s.config.GroupAllowFrom {
			if allowed == channelID {
				return true
			}
		}
		return false
	default:
		return mentionsBot
	}
}
