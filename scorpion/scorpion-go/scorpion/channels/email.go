package channels

import (
	"context"
	"fmt"
	"sync"
	"time"

	"scorpion-go/scorpion/bus"
)

// EmailChannel implements the Email chat channel using IMAP/SMTP.
type EmailChannel struct {
	*BaseChannelImpl
	mu      sync.RWMutex
	config  *EmailConfig
	stopCh  chan struct{}
	lastUID string // Last processed email UID
}

// EmailConfig holds Email channel configuration.
type EmailConfig struct {
	ChannelConfig
	ConsentGranted   bool   `json:"consent_granted"`
	IMAPHost         string `json:"imap_host"`
	IMAPPort         int    `json:"imap_port"`
	IMAPUsername     string `json:"imap_username"`
	IMAPPassword     string `json:"imap_password"`
	IMAPMailbox      string `json:"imap_mailbox"`
	IMAPUseSSL       bool   `json:"imap_use_ssl"`
	SMTPHost         string `json:"smtp_host"`
	SMTPPort         int    `json:"smtp_port"`
	SMTPUsername     string `json:"smtp_username"`
	SMTPPassword     string `json:"smtp_password"`
	SMTPUseTLS       bool   `json:"smtp_use_tls"`
	SMTPUseSSL       bool   `json:"smtp_use_ssl"`
	FromAddress      string `json:"from_address"`
	AutoReplyEnabled bool   `json:"auto_reply_enabled"`
	PollIntervalSecs int    `json:"poll_interval_seconds"`
	MarkSeen         bool   `json:"mark_seen"`
	MaxBodyChars     int    `json:"max_body_chars"`
	SubjectPrefix    string `json:"subject_prefix"`
}

// NewEmailChannel creates a new Email channel.
func NewEmailChannel(config *EmailConfig, messageBus *bus.MessageBus) *EmailChannel {
	if config.PollIntervalSecs <= 0 {
		config.PollIntervalSecs = 30
	}
	if config.MaxBodyChars <= 0 {
		config.MaxBodyChars = 12000
	}
	if config.SubjectPrefix == "" {
		config.SubjectPrefix = "Re: "
	}

	return &EmailChannel{
		BaseChannelImpl: NewBaseChannel("email", &config.ChannelConfig, messageBus),
		config:          config,
		stopCh:          make(chan struct{}),
	}
}

// Initialize initializes the Email channel.
func (e *EmailChannel) Initialize(ctx context.Context) error {
	if !e.config.ConsentGranted {
		return fmt.Errorf("email consent must be granted")
	}

	if e.config.IMAPHost == "" || e.config.SMTPHost == "" {
		return fmt.Errorf("email hosts are required")
	}

	// In a full implementation, this would:
	// 1. Create IMAP client for receiving emails
	// 2. Create SMTP client for sending emails
	// 3. Test connections

	return nil
}

// Start starts the Email channel (blocking).
func (e *EmailChannel) Start(ctx context.Context) error {
	e.SetRunning(true)
	defer e.SetRunning(false)

	// Start polling loop
	ticker := time.NewTicker(time.Duration(e.config.PollIntervalSecs) * time.Second)
	defer ticker.Stop()

	// Initial poll
	e.pollEmails(ctx)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-e.stopCh:
			return nil
		case <-ticker.C:
			e.pollEmails(ctx)
		}
	}
}

// Stop stops the Email channel.
func (e *EmailChannel) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.IsRunning() {
		return nil
	}

	close(e.stopCh)
	return nil
}

// Send sends an email through the channel.
func (e *EmailChannel) Send(ctx context.Context, msg *bus.OutboundMessage) error {
	if !e.IsRunning() {
		return fmt.Errorf("channel not running")
	}

	// In a full implementation, this would:
	// 1. Use SMTP to send the email
	// 2. Set proper headers (In-Reply-To, References)
	// 3. Handle attachments

	return nil
}

// pollEmails polls for new emails.
func (e *EmailChannel) pollEmails(ctx context.Context) {
	// In a full implementation, this would:
	// 1. Connect to IMAP server
	// 2. Select mailbox
	// 3. Search for unseen messages
	// 4. Fetch and parse messages
	// 5. Convert to InboundMessage and publish to bus
	// 6. Mark messages as seen if configured
	// 7. Send auto-replies if enabled
}

// checkSenderAllowed checks if a sender is allowed.
func (e *EmailChannel) checkSenderAllowed(sender string) bool {
	if len(e.config.AllowFrom) == 0 {
		return true
	}

	for _, allowed := range e.config.AllowFrom {
		if allowed == sender || allowed == "*" {
			return true
		}
	}
	return false
}
