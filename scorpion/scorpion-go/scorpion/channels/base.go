package channels

import (
	"context"

	"scorpion-go/scorpion/bus"
)

// BaseChannel is the interface that all chat channels must implement.
type BaseChannel interface {
	// Name returns the channel name.
	Name() string

	// Initialize initializes the channel.
	Initialize(ctx context.Context) error

	// Start starts the channel (blocking).
	Start(ctx context.Context) error

	// Stop stops the channel.
	Stop() error

	// Send sends a message through the channel.
	Send(ctx context.Context, msg *bus.OutboundMessage) error

	// IsRunning returns whether the channel is running.
	IsRunning() bool
}

// ChannelConfig holds common channel configuration.
type ChannelConfig struct {
	Enabled bool     `json:"enabled"`
	AllowFrom []string `json:"allow_from"`
}

// BaseChannelImpl provides common functionality for channels.
type BaseChannelImpl struct {
	name    string
	config  *ChannelConfig
	running bool
	bus     *bus.MessageBus
}

// NewBaseChannel creates a new base channel.
func NewBaseChannel(name string, config *ChannelConfig, messageBus *bus.MessageBus) *BaseChannelImpl {
	return &BaseChannelImpl{
		name:   name,
		config: config,
		bus:    messageBus,
	}
}

// Name returns the channel name.
func (b *BaseChannelImpl) Name() string {
	return b.name
}

// IsRunning returns whether the channel is running.
func (b *BaseChannelImpl) IsRunning() bool {
	return b.running
}

// SetRunning sets the running state.
func (b *BaseChannelImpl) SetRunning(running bool) {
	b.running = running
}

// IsAllowed checks if a sender is allowed.
func (b *BaseChannelImpl) IsAllowed(senderID string) bool {
	if len(b.config.AllowFrom) == 0 {
		return true // Empty allow list means everyone is allowed
	}

	for _, allowed := range b.config.AllowFrom {
		if allowed == senderID || allowed == "*" {
			return true
		}
	}
	return false
}

// PublishInbound publishes an inbound message to the bus.
func (b *BaseChannelImpl) PublishInbound(msg *bus.InboundMessage) error {
	if b.bus != nil {
		return b.bus.PublishInbound(msg)
	}
	return nil
}

// PublishOutbound publishes an outbound message to the bus.
func (b *BaseChannelImpl) PublishOutbound(msg *bus.OutboundMessage) error {
	if b.bus != nil {
		return b.bus.PublishOutbound(msg)
	}
	return nil
}
