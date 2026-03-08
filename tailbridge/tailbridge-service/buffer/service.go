package buffer

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/codejedi-ai/kaggle-for-tensors/tailscale-app/protocol"
)

// DeliveryFunc is a function that attempts to deliver a message
type DeliveryFunc func(ctx context.Context, msg *Message) error

// BufferService manages the message buffer with persistence and retry logic
type BufferService struct {
	// store is the persistent message store
	store *PersistentStore

	// strategy handles retry logic
	strategy *RetryStrategy

	// deliverFunc is called to attempt message delivery
	deliverFunc DeliveryFunc

	// httpClient is used for delivery attempts
	httpClient *http.Client

	// peerInboundPort is the port for outbound peer connections
	peerInboundPort int

	// mu protects concurrent access
	mu sync.RWMutex

	// running indicates if the background processor is active
	running bool

	// stopChan signals the background processor to stop
	stopChan chan struct{}

	// processInterval controls how often to check for messages to retry
	processInterval time.Duration
}

// BufferServiceConfig configures the buffer service
type BufferServiceConfig struct {
	// DataDir is the directory for persistent storage
	DataDir string

	// RetryConfig configures retry behavior
	RetryConfig *RetryConfig

	// ProcessInterval controls background processing frequency
	ProcessInterval time.Duration

	// HTTPTimeout is the timeout for delivery attempts
	HTTPTimeout time.Duration

	// PeerInboundPort is the port for outbound peer connections
	PeerInboundPort int
}

// DefaultBufferServiceConfig returns the default configuration
func DefaultBufferServiceConfig() *BufferServiceConfig {
	return &BufferServiceConfig{
		DataDir:         "./buffer_data",
		RetryConfig:     DefaultRetryConfig(),
		ProcessInterval: 5 * time.Second,
		HTTPTimeout:     20 * time.Second,
		PeerInboundPort: 8001,
	}
}

// NewBufferService creates a new buffer service
func NewBufferService(config *BufferServiceConfig, deliverFunc DeliveryFunc) (*BufferService, error) {
	if config == nil {
		config = DefaultBufferServiceConfig()
	}

	store, err := NewPersistentStore(config.DataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create store: %w", err)
	}

	httpClient := &http.Client{
		Timeout: config.HTTPTimeout,
	}

	bs := &BufferService{
		store:           store,
		strategy:        NewRetryStrategy(config.RetryConfig),
		deliverFunc:     deliverFunc,
		httpClient:      httpClient,
		peerInboundPort: config.PeerInboundPort,
		processInterval: config.ProcessInterval,
		stopChan:        make(chan struct{}),
	}

	return bs, nil
}

// Enqueue adds a message to the buffer for delivery
func (bs *BufferService) Enqueue(env protocol.Envelope, sourceListen string) (*Message, error) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	msg := &Message{
		ID:           uuid.New().String(),
		Envelope:     env,
		SourceListen: sourceListen,
		MaxRetries:   bs.strategy.config.MaxRetries,
	}

	bs.strategy.PrepareForInitialSend(msg)

	if err := bs.store.Save(msg); err != nil {
		return nil, fmt.Errorf("failed to persist message: %w", err)
	}

	log.Printf("[buffer] enqueued message %s for %s", msg.ID, env.DestNode)
	return msg, nil
}

// Start begins background processing of the buffer
func (bs *BufferService) Start(ctx context.Context) {
	bs.mu.Lock()
	if bs.running {
		bs.mu.Unlock()
		return
	}
	bs.running = true
	bs.mu.Unlock()

	go bs.backgroundProcessor(ctx)
	log.Printf("[buffer] background processor started (interval=%v)", bs.processInterval)
}

// Stop halts background processing
func (bs *BufferService) Stop() {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	if !bs.running {
		return
	}

	bs.running = false
	close(bs.stopChan)
	log.Printf("[buffer] background processor stopped")
}

// backgroundProcessor periodically checks for messages to retry
func (bs *BufferService) backgroundProcessor(ctx context.Context) {
	ticker := time.NewTicker(bs.processInterval)
	defer ticker.Stop()

	for {
		select {
		case <-bs.stopChan:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			bs.processDueMessages()
		}
	}
}

// processDueMessages finds and processes messages due for retry
func (bs *BufferService) processDueMessages() {
	bs.mu.RLock()
	dueMessages, err := bs.store.GetDueForRetry(time.Now())
	bs.mu.RUnlock()

	if err != nil {
		log.Printf("[buffer] failed to get due messages: %v", err)
		return
	}

	for _, msg := range dueMessages {
		go bs.attemptDelivery(msg)
	}
}

// attemptDelivery tries to deliver a message
func (bs *BufferService) attemptDelivery(msg *Message) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	// Refresh message from store
	current, err := bs.store.Get(msg.ID)
	if err != nil {
		log.Printf("[buffer] message %s not found, skipping", msg.ID)
		return
	}

	// Skip if already delivered or failed
	if current.Status == StatusDelivered || current.Status == StatusFailed {
		return
	}

	log.Printf("[buffer] attempting delivery of message %s to %s (attempt %d/%d)",
		msg.ID, msg.Envelope.DestNode, msg.RetryCount+1, msg.MaxRetries)

	err = bs.deliverFunc(context.Background(), msg)
	if err != nil {
		log.Printf("[buffer] delivery failed for message %s: %v", msg.ID, err)

		if bs.strategy.ShouldRetry(current) {
			bs.strategy.UpdateForRetry(current, err)
			if saveErr := bs.store.Save(current); saveErr != nil {
				log.Printf("[buffer] failed to save retry state for %s: %v", msg.ID, saveErr)
			}
			log.Printf("[buffer] message %s scheduled for retry at %v", msg.ID, current.NextRetryAt)
		} else {
			bs.strategy.MarkAsFailed(current, err)
			if saveErr := bs.store.Save(current); saveErr != nil {
				log.Printf("[buffer] failed to save failed state for %s: %v", msg.ID, saveErr)
			}
			log.Printf("[buffer] message %s marked as failed (max retries exceeded)", msg.ID)
		}
	} else {
		bs.strategy.MarkAsDelivered(current)
		if saveErr := bs.store.Save(current); saveErr != nil {
			log.Printf("[buffer] failed to save delivered state for %s: %v", msg.ID, saveErr)
		}
		log.Printf("[buffer] message %s delivered successfully", msg.ID)
	}
}

// DeliverImmediately attempts immediate delivery of a message
func (bs *BufferService) DeliverImmediately(msg *Message) error {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	err := bs.deliverFunc(context.Background(), msg)
	if err != nil {
		if bs.strategy.ShouldRetry(msg) {
			bs.strategy.UpdateForRetry(msg, err)
		} else {
			bs.strategy.MarkAsFailed(msg, err)
		}
		if saveErr := bs.store.Save(msg); saveErr != nil {
			log.Printf("[buffer] failed to save message state: %v", saveErr)
		}
		return err
	}

	bs.strategy.MarkAsDelivered(msg)
	if saveErr := bs.store.Save(msg); saveErr != nil {
		log.Printf("[buffer] failed to save delivered state: %v", saveErr)
	}
	return nil
}

// GetStats returns buffer statistics
func (bs *BufferService) GetStats() BufferStats {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	return bs.store.Stats()
}

// GetMessage retrieves a message by ID
func (bs *BufferService) GetMessage(id string) (*Message, error) {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	return bs.store.Get(id)
}

// GetAllMessages retrieves all messages
func (bs *BufferService) GetAllMessages() ([]*Message, error) {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	return bs.store.GetAll()
}

// GetPendingMessages retrieves all pending messages
func (bs *BufferService) GetPendingMessages() ([]*Message, error) {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	return bs.store.GetByStatus(StatusPending)
}

// GetFailedMessages retrieves all failed messages
func (bs *BufferService) GetFailedMessages() ([]*Message, error) {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	return bs.store.GetByStatus(StatusFailed)
}

// RetryFailedMessage marks a failed message for retry
func (bs *BufferService) RetryFailedMessage(id string) error {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	msg, err := bs.store.Get(id)
	if err != nil {
		return err
	}

	if msg.Status != StatusFailed {
		return fmt.Errorf("message %s is not in failed state", id)
	}

	msg.Status = StatusPending
	msg.RetryCount = 0
	msg.LastError = ""
	msg.UpdatedAt = time.Now()
	msg.NextRetryAt = time.Now()

	return bs.store.Save(msg)
}

// ClearDelivered removes all delivered messages from the buffer
func (bs *BufferService) ClearDelivered() (int, error) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	delivered, err := bs.store.GetByStatus(StatusDelivered)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, msg := range delivered {
		if err := bs.store.Delete(msg.ID); err != nil {
			log.Printf("[buffer] failed to delete delivered message %s: %v", msg.ID, err)
			continue
		}
		count++
	}

	log.Printf("[buffer] cleared %d delivered messages", count)
	return count, nil
}

// IsRunning returns whether the buffer service is running
func (bs *BufferService) IsRunning() bool {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	return bs.running
}
