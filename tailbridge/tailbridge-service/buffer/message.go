package buffer

import (
	"time"

	"github.com/codejedi-ai/kaggle-for-tensors/tailscale-app/protocol"
)

// MessageStatus represents the current state of a buffered message
type MessageStatus string

const (
	// StatusPending - message waiting to be delivered
	StatusPending MessageStatus = "pending"
	// StatusRetrying - message delivery failed, scheduled for retry
	StatusRetrying MessageStatus = "retrying"
	// StatusDelivered - message successfully delivered
	StatusDelivered MessageStatus = "delivered"
	// StatusFailed - message delivery failed permanently
	StatusFailed MessageStatus = "failed"
)

// Message represents a buffered message with delivery metadata
type Message struct {
	// ID is a unique identifier for this message
	ID string `json:"id"`

	// CreatedAt is when the message was first enqueued
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the message was last modified
	UpdatedAt time.Time `json:"updated_at"`

	// NextRetryAt is when the next delivery attempt should be made
	NextRetryAt time.Time `json:"next_retry_at,omitempty"`

	// Status is the current delivery status
	Status MessageStatus `json:"status"`

	// RetryCount tracks how many delivery attempts have been made
	RetryCount int `json:"retry_count"`

	// MaxRetries is the maximum number of retry attempts
	MaxRetries int `json:"max_retries"`

	// LastError contains the error message from the last failed attempt
	LastError string `json:"last_error,omitempty"`

	// Envelope contains the actual message payload
	Envelope protocol.Envelope `json:"envelope"`

	// SourceListen is the local listen address of the source bridge
	SourceListen string `json:"source_listen,omitempty"`

	// CallbackURL is an optional URL to notify on delivery status change
	CallbackURL string `json:"callback_url,omitempty"`
}

// Envelope wraps a protocol.Envelope with buffer metadata for transport
type Envelope struct {
	// Message is the buffered message being transported
	Message Message `json:"message"`

	// TransportType indicates how this envelope is being transported
	TransportType string `json:"transport_type"`

	// Timestamp is when this envelope was created
	Timestamp time.Time `json:"timestamp"`
}

// BufferStats provides statistics about the buffer state
type BufferStats struct {
	// TotalMessages is the total number of messages in the buffer
	TotalMessages int `json:"total_messages"`

	// PendingCount is the number of messages waiting to be delivered
	PendingCount int `json:"pending_count"`

	// RetryingCount is the number of messages being retried
	RetryingCount int `json:"retrying_count"`

	// FailedCount is the number of messages that failed permanently
	FailedCount int `json:"failed_count"`

	// DeliveredCount is the number of messages successfully delivered
	DeliveredCount int `json:"delivered_count"`

	// OldestMessage is the creation time of the oldest pending message
	OldestMessage time.Time `json:"oldest_message,omitempty"`

	// NewestMessage is the creation time of the newest message
	NewestMessage time.Time `json:"newest_message,omitempty"`
}
