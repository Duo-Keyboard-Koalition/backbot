package bus

import (
	"sync"
)

// MessageBus is a message bus for decoupled communication.
type MessageBus struct {
	mu            sync.RWMutex
	inboundChan   chan *InboundMessage
	outboundChan  chan *OutboundMessage
	subscribers   map[string][]chan *InboundMessage
	closed        bool
}

// NewMessageBus creates a new message bus.
func NewMessageBus(bufferSize int) *MessageBus {
	if bufferSize <= 0 {
		bufferSize = 100
	}

	return &MessageBus{
		inboundChan:  make(chan *InboundMessage, bufferSize),
		outboundChan: make(chan *OutboundMessage, bufferSize),
		subscribers:  make(map[string][]chan *InboundMessage),
	}
}

// PublishInbound publishes an inbound message.
func (mb *MessageBus) PublishInbound(msg *InboundMessage) error {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	if mb.closed {
		return ErrBusClosed
	}

	mb.inboundChan <- msg

	// Notify subscribers for this channel
	if subs, ok := mb.subscribers[msg.Channel]; ok {
		for _, sub := range subs {
			select {
			case sub <- msg:
			default:
				// Drop message if subscriber is slow
			}
		}
	}

	return nil
}

// PublishOutbound publishes an outbound message.
func (mb *MessageBus) PublishOutbound(msg *OutboundMessage) error {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	if mb.closed {
		return ErrBusClosed
	}

	mb.outboundChan <- msg
	return nil
}

// Subscribe subscribes to messages from a specific channel.
func (mb *MessageBus) Subscribe(channel string) (<-chan *InboundMessage, func()) {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	subChan := make(chan *InboundMessage, 100)
	mb.subscribers[channel] = append(mb.subscribers[channel], subChan)

	unsubscribe := func() {
		mb.mu.Lock()
		defer mb.mu.Unlock()
		subs := mb.subscribers[channel]
		for i, sub := range subs {
			if sub == subChan {
				mb.subscribers[channel] = append(subs[:i], subs[i+1:]...)
				close(subChan)
				break
			}
		}
	}

	return subChan, unsubscribe
}

// GetInboundChannel returns the inbound message channel.
func (mb *MessageBus) GetInboundChannel() <-chan *InboundMessage {
	return mb.inboundChan
}

// GetOutboundChannel returns the outbound message channel.
func (mb *MessageBus) GetOutboundChannel() <-chan *OutboundMessage {
	return mb.outboundChan
}

// Close closes the message bus.
func (mb *MessageBus) Close() {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	if mb.closed {
		return
	}

	mb.closed = true
	close(mb.inboundChan)
	close(mb.outboundChan)

	// Close all subscriber channels
	for _, subs := range mb.subscribers {
		for _, sub := range subs {
			close(sub)
		}
	}
}

// IsClosed returns whether the bus is closed.
func (mb *MessageBus) IsClosed() bool {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	return mb.closed
}

// Bus errors
var (
	ErrBusClosed = &busError{"message bus is closed"}
)

type busError struct {
	message string
}

func (e *busError) Error() string {
	return e.message
}
