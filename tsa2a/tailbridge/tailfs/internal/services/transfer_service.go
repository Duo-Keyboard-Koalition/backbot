package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/codejedi-ai/kaggle-for-tensors/tailscale-app/tail-agent-file-send/internal/models"
)

// FileTransferService manages file transfers between agents
type FileTransferService struct {
	config      *models.FileTransferConfig
	transfers   map[string]*transferState
	mu          sync.RWMutex
	stopChan    chan struct{}

	// Callbacks
	onProgress  func(*models.FileTransferProgress)
	onComplete  func(*models.FileTransferHistoryEntry)
}

type transferState struct {
	request   *models.FileTransferRequest
	progress  *models.FileTransferProgress
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewFileTransferService creates a new file transfer service
func NewFileTransferService(config *models.FileTransferConfig) (*FileTransferService, error) {
	if config == nil {
		config = models.DefaultFileTransferConfig()
	}
	
	return &FileTransferService{
		config:    config,
		transfers: make(map[string]*transferState),
		stopChan:  make(chan struct{}),
	}, nil
}

// Start begins the file transfer service
func (s *FileTransferService) Start() {
	// Start cleanup goroutine
	go s.cleanupLoop()
}

// Stop halts the file transfer service
func (s *FileTransferService) Stop() {
	close(s.stopChan)
	
	// Cancel all active transfers
	s.mu.Lock()
	for _, state := range s.transfers {
		if state.cancel != nil {
			state.cancel()
		}
	}
	s.mu.Unlock()
}

// SendFile initiates a file transfer to a remote agent
func (s *FileTransferService) SendFile(ctx context.Context, req *models.FileTransferRequest) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if file exists
	info, err := os.Stat(req.FilePath)
	if err != nil {
		return "", fmt.Errorf("file not found: %w", err)
	}

	req.FileSize = info.Size()
	req.CreatedAt = time.Now()

	// Create transfer state
	tctx, cancel := context.WithTimeout(ctx, s.config.TransferTimeout)
	state := &transferState{
		request: req,
		cancel:  cancel,
		progress: &models.FileTransferProgress{
			TransferID: req.ID,
			Status:     models.StatusPending,
			BytesTotal: req.FileSize,
		},
	}

	s.transfers[req.ID] = state

	// Start transfer in background
	go s.executeTransfer(tctx, state)

	return req.ID, nil
}

// GetProgress returns the progress of a transfer
func (s *FileTransferService) GetProgress(transferID string) (*models.FileTransferProgress, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	state, exists := s.transfers[transferID]
	if !exists {
		return nil, false
	}
	
	return state.progress, true
}

// CancelTransfer cancels an ongoing transfer
func (s *FileTransferService) CancelTransfer(transferID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	state, exists := s.transfers[transferID]
	if !exists {
		return fmt.Errorf("transfer not found: %s", transferID)
	}
	
	if state.cancel != nil {
		state.cancel()
	}
	
	state.progress.Status = models.StatusCancelled
	return nil
}

// GetActiveTransfers returns all active transfers
func (s *FileTransferService) GetActiveTransfers() []models.FileTransferProgress {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	result := make([]models.FileTransferProgress, 0)
	for _, state := range s.transfers {
		if state.progress.Status == models.StatusSending || state.progress.Status == models.StatusReceiving {
			result = append(result, *state.progress)
		}
	}
	return result
}

// GetTransferHistory returns completed/failed transfers
func (s *FileTransferService) GetTransferHistory() []models.FileTransferHistoryEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	result := make([]models.FileTransferHistoryEntry, 0)
	for _, state := range s.transfers {
		if state.progress.Status == models.StatusComplete || state.progress.Status == models.StatusFailed {
			entry := s.progressToHistory(state.progress)
			result = append(result, entry)
		}
	}
	return result
}

// CalculateFileHash calculates SHA256 hash of a file
func CalculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// GetFileSize returns the size of a file
func GetFileSize(filePath string) (int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// GetFileType detects the file type
func GetFileType(filePath string) string {
	ext := filepath.Ext(filePath)
	return ext
}

// EnsureDir creates directory if it doesn't exist
func EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

func (s *FileTransferService) executeTransfer(ctx context.Context, state *transferState) {
	req := state.request
	progress := state.progress
	
	progress.Status = models.StatusSending
	startTime := time.Now()
	progress.StartedAt = &startTime
	
	// Open source file
	file, err := os.Open(req.FilePath)
	if err != nil {
		s.failTransfer(state, fmt.Sprintf("failed to open file: %v", err))
		return
	}
	defer file.Close()
	
	// Calculate chunks
	totalChunks := int((req.FileSize + s.config.ChunkSize - 1) / s.config.ChunkSize)
	progress.TotalChunks = int64(totalChunks)
	
	buf := make([]byte, s.config.ChunkSize)
	var bytesSent int64 = 0
	
	for chunkIndex := 0; chunkIndex < totalChunks; chunkIndex++ {
		select {
		case <-ctx.Done():
			s.failTransfer(state, "transfer cancelled")
			return
		default:
		}
		
		// Read chunk
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			s.failTransfer(state, fmt.Sprintf("read error: %v", err))
			return
		}
		
		if n == 0 {
			break
		}
		
		// Send chunk (placeholder - actual sending via taila2a)
		chunk := &models.FileTransferChunk{
			TransferID: req.ID,
			ChunkIndex: chunkIndex,
			TotalChunks: totalChunks,
			ChunkSize:  n,
			Data:       buf[:n],
		}
		
		// TODO: Send chunk via taila2a protocol
		_ = chunk
		
		bytesSent += int64(n)
		progress.BytesSent = bytesSent
		progress.PercentComplete = float64(bytesSent) * 100.0 / float64(req.FileSize)
		
		// Calculate speed
		elapsed := time.Since(startTime).Seconds()
		if elapsed > 0 {
			progress.BytesPerSecond = float64(bytesSent) / elapsed
			remaining := float64(req.FileSize - bytesSent)
			if progress.BytesPerSecond > 0 {
				progress.ETASeconds = int64(remaining / progress.BytesPerSecond)
			}
		}
		
		// Notify progress
		if s.onProgress != nil {
			s.onProgress(progress)
		}
	}
	
	// Complete
	completedTime := time.Now()
	progress.Status = models.StatusComplete
	progress.CompletedAt = &completedTime
	
	if s.onComplete != nil {
		entry := s.progressToHistory(progress)
		s.onComplete(&entry)
	}
}

func (s *FileTransferService) failTransfer(state *transferState, errMsg string) {
	state.progress.Status = models.StatusFailed
	state.progress.Error = errMsg
	completedTime := time.Now()
	state.progress.CompletedAt = &completedTime
	
	if s.onComplete != nil {
		entry := s.progressToHistory(state.progress)
		s.onComplete(&entry)
	}
}

func (s *FileTransferService) progressToHistory(p *models.FileTransferProgress) models.FileTransferHistoryEntry {
	duration := 0.0
	if p.StartedAt != nil && p.CompletedAt != nil {
		duration = p.CompletedAt.Sub(*p.StartedAt).Seconds()
	}
	
	avgSpeed := 0.0
	if duration > 0 {
		avgSpeed = float64(p.BytesSent) / duration
	}
	
	return models.FileTransferHistoryEntry{
		ID:              p.TransferID,
		SourceAgent:     "",
		DestAgent:       "",
		FileName:        "",
		FileSize:        p.BytesTotal,
		Status:          p.Status,
		StartedAt:       *p.StartedAt,
		CompletedAt:     *p.CompletedAt,
		DurationSeconds: duration,
		AvgSpeed:        avgSpeed,
	}
}

func (s *FileTransferService) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.cleanupOldTransfers()
		}
	}
}

func (s *FileTransferService) cleanupOldTransfers() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	now := time.Now()
	for id, state := range s.transfers {
		if state.progress.CompletedAt != nil {
			if now.Sub(*state.progress.CompletedAt) > 24*time.Hour {
				delete(s.transfers, id)
			}
		}
	}
}
