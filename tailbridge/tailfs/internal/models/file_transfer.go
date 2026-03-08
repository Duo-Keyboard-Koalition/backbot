package models

import (
	"encoding/json"
	"time"
)

// FileTransferStatus represents the status of a file transfer
type FileTransferStatus string

const (
	StatusPending   FileTransferStatus = "pending"
	StatusSending   FileTransferStatus = "sending"
	StatusReceiving FileTransferStatus = "receiving"
	StatusComplete  FileTransferStatus = "complete"
	StatusFailed    FileTransferStatus = "failed"
	StatusCancelled FileTransferStatus = "cancelled"
)

// FileTransferRequest represents a request to send a file
type FileTransferRequest struct {
	// ID is the unique transfer identifier
	ID string `json:"id"`
	
	// Source agent information
	SourceAgentID   string `json:"source_agent_id"`
	SourceAgentName string `json:"source_agent_name"`
	
	// Destination agent information
	DestAgentID     string `json:"dest_agent_id"`
	DestAgentName   string `json:"dest_agent_name"`
	
	// File information
	FileName        string `json:"file_name"`
	FilePath        string `json:"file_path,omitempty"`
	FileSize        int64  `json:"file_size"`
	FileHash        string `json:"file_hash,omitempty"`
	FileType        string `json:"file_type,omitempty"`
	
	// Transfer options
	Compress        bool   `json:"compress,omitempty"`
	Encrypt         bool   `json:"encrypt,omitempty"`
	Verify          bool   `json:"verify,omitempty"`
	
	// Metadata
	CreatedAt       time.Time `json:"created_at"`
	ExpiresAt       time.Time `json:"expires_at,omitempty"`
	
	// Callback
	CallbackURL     string `json:"callback_url,omitempty"`
}

// FileTransferResponse represents the response to a transfer request
type FileTransferResponse struct {
	// Request ID
	RequestID       string `json:"request_id"`
	
	// Acceptance
	Accepted        bool   `json:"accepted"`
	Message         string `json:"message,omitempty"`
	
	// Destination info
	DestPath        string `json:"dest_path,omitempty"`
	
	// Timing
	RespondedAt     time.Time `json:"responded_at"`
}

// FileTransferChunk represents a chunk of file data
type FileTransferChunk struct {
	// Transfer ID
	TransferID      string `json:"transfer_id"`
	
	// Chunk info
	ChunkIndex      int    `json:"chunk_index"`
	TotalChunks     int    `json:"total_chunks"`
	ChunkSize       int    `json:"chunk_size"`
	
	// Data
	Data            []byte `json:"data"`
	
	// Verification
	ChunkHash       string `json:"chunk_hash,omitempty"`
}

// FileTransferProgress represents transfer progress
type FileTransferProgress struct {
	TransferID      string             `json:"transfer_id"`
	Status          FileTransferStatus `json:"status"`

	// Progress
	BytesSent       int64              `json:"bytes_sent"`
	BytesTotal      int64              `json:"bytes_total"`
	PercentComplete float64            `json:"percent_complete"`
	TotalChunks     int64              `json:"total_chunks,omitempty"`

	// Speed
	BytesPerSecond  float64            `json:"bytes_per_second"`
	ETASeconds      int64              `json:"eta_seconds"`

	// Timing
	StartedAt       *time.Time         `json:"started_at,omitempty"`
	CompletedAt     *time.Time         `json:"completed_at,omitempty"`

	// Errors
	Error           string             `json:"error,omitempty"`
}

// FileTransferHistoryEntry represents a completed transfer in history
type FileTransferHistoryEntry struct {
	ID              string             `json:"id"`
	SourceAgent     string             `json:"source_agent"`
	DestAgent       string             `json:"dest_agent"`
	FileName        string             `json:"file_name"`
	FileSize        int64              `json:"file_size"`
	Status          FileTransferStatus `json:"status"`
	StartedAt       time.Time          `json:"started_at"`
	CompletedAt     time.Time          `json:"completed_at"`
	DurationSeconds float64            `json:"duration_seconds"`
	AvgSpeed        float64            `json:"avg_speed"`
}

// FileTransferConfig configures the file transfer service
type FileTransferConfig struct {
	// ChunkSize is the size of each transfer chunk (default: 1MB)
	ChunkSize int64
	
	// MaxConcurrentTransfers limits simultaneous transfers
	MaxConcurrentTransfers int
	
	// TransferTimeout is the timeout for individual transfers
	TransferTimeout time.Duration
	
	// MaxRetries is the number of retry attempts
	MaxRetries int
	
	// DownloadDir is the default directory for received files
	DownloadDir string
	
	// EnableCompression enables compression for transfers
	EnableCompression bool
	
	// EnableEncryption enables encryption for transfers
	EnableEncryption bool
}

// DefaultFileTransferConfig returns default configuration
func DefaultFileTransferConfig() *FileTransferConfig {
	return &FileTransferConfig{
		ChunkSize:            1024 * 1024, // 1MB
		MaxConcurrentTransfers: 3,
		TransferTimeout:      30 * time.Minute,
		MaxRetries:           3,
		DownloadDir:          "~/Downloads/tailfs",
		EnableCompression:    true,
		EnableEncryption:     true,
	}
}

// MarshalJSON implements custom JSON marshaling for progress
func (p FileTransferProgress) MarshalJSON() ([]byte, error) {
	type Alias FileTransferProgress
	return json.Marshal(&struct {
		*Alias
		PercentComplete string `json:"percent_complete_str"`
	}{
		Alias:           (*Alias)(&p),
		PercentComplete: formatPercent(p.PercentComplete),
	})
}

func formatPercent(p float64) string {
	if p < 0 {
		return "0.0%"
	}
	if p > 100 {
		return "100.0%"
	}
	// Simple formatting
	intPart := int(p)
	decPart := int((p - float64(intPart)) * 10)
	return string(rune('0'+intPart/10)) + "." + string(rune('0'+intPart%10)) + string(rune('0'+decPart)) + "%"
}
