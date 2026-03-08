package channels

import (
	"context"
	"fmt"
	"sync"

	"scorpion-go/scorpion/bus"
)

// TelegramChannel implements the Telegram chat channel.
type TelegramChannel struct {
	*BaseChannelImpl
	mu      sync.RWMutex
	config  *TelegramConfig
	stopCh  chan struct{}
}

// TelegramConfig holds Telegram channel configuration.
type TelegramConfig struct {
	ChannelConfig
	Token          string `json:"token"`
	Proxy          string `json:"proxy"`
	ReplyToMessage bool   `json:"reply_to_message"`
	ReactEmoji     string `json:"react_emoji"`
	Voice          VoiceConfig `json:"voice"`
}

// VoiceConfig holds voice/TTS configuration.
type VoiceConfig struct {
	Enabled bool   `json:"enabled"`
	Voice   string `json:"voice"`
	Always  bool   `json:"always"`
}

// NewTelegramChannel creates a new Telegram channel.
func NewTelegramChannel(config *TelegramConfig, messageBus *bus.MessageBus) *TelegramChannel {
	return &TelegramChannel{
		BaseChannelImpl: NewBaseChannel("telegram", &config.ChannelConfig, messageBus),
		config:          config,
		stopCh:          make(chan struct{}),
	}
}

// Initialize initializes the Telegram channel.
func (t *TelegramChannel) Initialize(ctx context.Context) error {
	if t.config.Token == "" {
		return fmt.Errorf("telegram token is required")
	}

	// In a full implementation, this would:
	// 1. Create a Telegram bot API client
	// 2. Set up webhook or long polling
	// 3. Configure proxy if specified

	return nil
}

// Start starts the Telegram channel (blocking).
func (t *TelegramChannel) Start(ctx context.Context) error {
	t.SetRunning(true)
	defer t.SetRunning(false)

	// In a full implementation, this would:
	// 1. Start long polling or webhook listener
	// 2. Receive messages from Telegram
	// 3. Convert to InboundMessage and publish to bus
	// 4. Listen to bus for OutboundMessage and send to Telegram

	<-t.stopCh
	return nil
}

// Stop stops the Telegram channel.
func (t *TelegramChannel) Stop() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.IsRunning() {
		return nil
	}

	close(t.stopCh)
	return nil
}

// Send sends a message through the Telegram channel.
func (t *TelegramChannel) Send(ctx context.Context, msg *bus.OutboundMessage) error {
	if !t.IsRunning() {
		return fmt.Errorf("channel not running")
	}

	// In a full implementation, this would:
	// 1. Use the Telegram Bot API to send the message
	// 2. Handle media attachments
	// 3. Handle reply-to messages
	// 4. Add reactions if configured

	return nil
}

// GetUpdates polls for new messages (helper method).
func (t *TelegramChannel) GetUpdates(ctx context.Context) error {
	// Placeholder for actual Telegram polling implementation
	return nil
}
