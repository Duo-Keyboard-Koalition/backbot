package buffer

import (
	"math"
	"time"
)

// RetryConfig configures the retry behavior for message delivery
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts
	MaxRetries int

	// InitialDelay is the delay before the first retry
	InitialDelay time.Duration

	// MaxDelay is the maximum delay between retries
	MaxDelay time.Duration

	// Multiplier is the factor by which the delay increases
	Multiplier float64
}

// DefaultRetryConfig returns the default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:   5,
		InitialDelay: 1 * time.Second,
		MaxDelay:     5 * time.Minute,
		Multiplier:   2.0,
	}
}

// RetryStrategy provides exponential backoff calculation
type RetryStrategy struct {
	config *RetryConfig
}

// NewRetryStrategy creates a new retry strategy with the given configuration
func NewRetryStrategy(config *RetryConfig) *RetryStrategy {
	if config == nil {
		config = DefaultRetryConfig()
	}
	return &RetryStrategy{config: config}
}

// CalculateNextRetry calculates the next retry time based on the current retry count
func (s *RetryStrategy) CalculateNextRetry(retryCount int) time.Duration {
	if retryCount < 0 {
		retryCount = 0
	}

	// Calculate exponential backoff: initialDelay * multiplier^retryCount
	delay := float64(s.config.InitialDelay) * math.Pow(s.config.Multiplier, float64(retryCount))

	// Cap at max delay
	if delay > float64(s.config.MaxDelay) {
		delay = float64(s.config.MaxDelay)
	}

	return time.Duration(delay)
}

// ShouldRetry determines if a message should be retried based on its current state
func (s *RetryStrategy) ShouldRetry(msg *Message) bool {
	if msg == nil {
		return false
	}

	// Don't retry if we've exceeded max retries
	if msg.RetryCount >= s.config.MaxRetries {
		return false
	}

	// Don't retry if already delivered or permanently failed
	if msg.Status == StatusDelivered || msg.Status == StatusFailed {
		return false
	}

	return true
}

// UpdateForRetry updates a message for the next retry attempt
func (s *RetryStrategy) UpdateForRetry(msg *Message, err error) {
	msg.RetryCount++
	msg.LastError = err.Error()
	msg.Status = StatusRetrying
	msg.UpdatedAt = time.Now()
	msg.NextRetryAt = time.Now().Add(s.CalculateNextRetry(msg.RetryCount))
}

// MarkAsFailed marks a message as permanently failed
func (s *RetryStrategy) MarkAsFailed(msg *Message, err error) {
	msg.Status = StatusFailed
	msg.LastError = err.Error()
	msg.UpdatedAt = time.Now()
}

// MarkAsDelivered marks a message as successfully delivered
func (s *RetryStrategy) MarkAsDelivered(msg *Message) {
	msg.Status = StatusDelivered
	msg.UpdatedAt = time.Now()
}

// PrepareForInitialSend prepares a new message for its first send attempt
func (s *RetryStrategy) PrepareForInitialSend(msg *Message) {
	msg.Status = StatusPending
	msg.RetryCount = 0
	msg.CreatedAt = time.Now()
	msg.UpdatedAt = time.Now()
	msg.NextRetryAt = time.Now() // Ready to send immediately
}
