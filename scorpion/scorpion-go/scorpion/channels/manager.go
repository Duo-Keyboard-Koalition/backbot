package channels

import (
	"context"
	"fmt"
	"sync"

	"scorpion-go/scorpion/bus"
)

// ChannelManager manages multiple chat channels.
type ChannelManager struct {
	mu       sync.RWMutex
	channels map[string]BaseChannel
	bus      *bus.MessageBus
}

// NewChannelManager creates a new channel manager.
func NewChannelManager(messageBus *bus.MessageBus) *ChannelManager {
	return &ChannelManager{
		channels: make(map[string]BaseChannel),
		bus:      messageBus,
	}
}

// Register registers a channel.
func (cm *ChannelManager) Register(channel BaseChannel) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	name := channel.Name()
	if _, exists := cm.channels[name]; exists {
		return fmt.Errorf("channel %q already registered", name)
	}

	cm.channels[name] = channel
	return nil
}

// Get retrieves a channel by name.
func (cm *ChannelManager) Get(name string) (BaseChannel, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	channel, ok := cm.channels[name]
	return channel, ok
}

// List returns all registered channels.
func (cm *ChannelManager) List() []BaseChannel {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	channels := make([]BaseChannel, 0, len(cm.channels))
	for _, channel := range cm.channels {
		channels = append(channels, channel)
	}
	return channels
}

// StartAll starts all registered channels.
func (cm *ChannelManager) StartAll(ctx context.Context) error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for _, channel := range cm.channels {
		if err := channel.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize channel %s: %w", channel.Name(), err)
		}

		go func(ch BaseChannel) {
			if err := ch.Start(ctx); err != nil {
				// Log error but continue with other channels
				fmt.Printf("Channel %s error: %v\n", ch.Name(), err)
			}
		}(channel)
	}

	return nil
}

// StopAll stops all channels.
func (cm *ChannelManager) StopAll() error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var lastErr error
	for _, channel := range cm.channels {
		if err := channel.Stop(); err != nil {
			lastErr = fmt.Errorf("failed to stop channel %s: %w", channel.Name(), err)
		}
	}

	return lastErr
}

// Send sends a message through a specific channel.
func (cm *ChannelManager) Send(ctx context.Context, channelName string, msg *bus.OutboundMessage) error {
	cm.mu.RLock()
	channel, ok := cm.channels[channelName]
	cm.mu.RUnlock()

	if !ok {
		return fmt.Errorf("channel %q not found", channelName)
	}

	return channel.Send(ctx, msg)
}

// Broadcast sends a message to all channels.
func (cm *ChannelManager) Broadcast(ctx context.Context, msg *bus.OutboundMessage) error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var lastErr error
	for _, channel := range cm.channels {
		if err := channel.Send(ctx, msg); err != nil {
			lastErr = fmt.Errorf("failed to send to channel %s: %w", channel.Name(), err)
		}
	}

	return lastErr
}
