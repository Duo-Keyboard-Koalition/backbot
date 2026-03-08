//go:build integration

package filetransfer

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

// FileTransferIntegrationTestSuite provides integration tests for file transfer
type FileTransferIntegrationTestSuite struct {
	suite.Suite
	agent1URL string
	agent2URL string
	agent3URL string
	client    *http.Client
}

// FileTransferRequest represents a file transfer request
type FileTransferRequest struct {
	ID            string `json:"id"`
	FilePath      string `json:"file_path"`
	DestAgentName string `json:"dest_agent_name"`
	FileName      string `json:"file_name"`
	FileSize      int64  `json:"file_size"`
	Compress      bool   `json:"compress"`
	Encrypt       bool   `json:"encrypt"`
}

// TransferProgress represents transfer progress
type TransferProgress struct {
	TransferID    string  `json:"transfer_id"`
	Status        string  `json:"status"`
	BytesSent     int64   `json:"bytes_sent"`
	BytesTotal    int64   `json:"bytes_total"`
	PercentComplete float64 `json:"percent_complete"`
	BytesPerSecond int64  `json:"bytes_per_second"`
	ETASeconds    int64   `json:"eta_seconds"`
}

// SetupSuite runs once before all tests
func (s *FileTransferIntegrationTestSuite) SetupSuite() {
	s.agent1URL = getEnv("AGENT1_URL", "http://localhost:8081")
	s.agent2URL = getEnv("AGENT2_URL", "http://localhost:8082")
	s.agent3URL = getEnv("AGENT3_URL", "http://localhost:8083")

	s.client = &http.Client{
		Timeout: 5 * time.Minute,
	}
}

// SetupTest runs before each test
func (s *FileTransferIntegrationTestSuite) SetupTest() {
	s.waitForAgents()
}

// waitForAgents waits for all agents to be healthy
func (s *FileTransferIntegrationTestSuite) waitForAgents() {
	agents := []string{s.agent1URL, s.agent2URL, s.agent3URL}
	
	for _, url := range agents {
		s.Require().Eventually(func() bool {
			resp, err := s.client.Get(url + "/health")
			if err != nil {
				return false
			}
			defer resp.Body.Close()
			return resp.StatusCode == http.StatusOK
		}, 2*time.Minute, 5*time.Second, "Agent %s did not become ready", url)
	}
}

// TestSmallFileTransfer tests transferring a small file
func (s *FileTransferIntegrationTestSuite) TestSmallFileTransfer() {
	// Create test file (100 KB)
	fileSize := int64(100 * 1024)
	fileContent := make([]byte, fileSize)
	_, err := rand.Read(fileContent)
	s.Require().NoError(err)

	// Upload file to agent1
	fileID, err := s.uploadFile(s.agent1URL, "test-small.bin", fileContent)
	s.Require().NoError(err)

	// Initiate transfer to agent2
	transferReq := FileTransferRequest{
		ID:            uuid.New().String(),
		FilePath:      fileID,
		DestAgentName: "agent2",
		FileName:      "test-small.bin",
		FileSize:      fileSize,
		Compress:      false,
		Encrypt:       false,
	}

	transferID, err := s.initiateTransfer(s.agent1URL, transferReq)
	s.Require().NoError(err)

	// Wait for transfer to complete
	s.waitForTransferComplete(s.agent1URL, transferID, 30*time.Second)

	// Verify file received on agent2
	receivedContent, err := s.downloadFile(s.agent2URL, "test-small.bin")
	s.Require().NoError(err)
	s.Assert().Equal(fileContent, receivedContent)
}

// TestLargeFileTransfer tests transferring a large file (10 MB)
func (s *FileTransferIntegrationTestSuite) TestLargeFileTransfer() {
	// Create test file (10 MB)
	fileSize := int64(10 * 1024 * 1024)
	fileContent := make([]byte, fileSize)
	// Fill with pattern instead of random for speed
	for i := range fileContent {
		fileContent[i] = byte(i % 256)
	}

	// Upload file to agent1
	fileID, err := s.uploadFile(s.agent1URL, "test-large.bin", fileContent)
	s.Require().NoError(err)

	// Initiate transfer to agent3
	transferReq := FileTransferRequest{
		ID:            uuid.New().String(),
		FilePath:      fileID,
		DestAgentName: "agent3",
		FileName:      "test-large.bin",
		FileSize:      fileSize,
		Compress:      true,
		Encrypt:       false,
	}

	transferID, err := s.initiateTransfer(s.agent1URL, transferReq)
	s.Require().NoError(err)

	// Wait for transfer to complete
	s.waitForTransferComplete(s.agent1URL, transferID, 2*time.Minute)

	// Verify file received on agent3
	receivedContent, err := s.downloadFile(s.agent3URL, "test-large.bin")
	s.Require().NoError(err)
	s.Assert().Equal(fileContent, receivedContent)
}

// TestTransferProgress tests progress tracking during transfer
func (s *FileTransferIntegrationTestSuite) TestTransferProgress() {
	// Create test file (1 MB)
	fileSize := int64(1 * 1024 * 1024)
	fileContent := make([]byte, fileSize)
	for i := range fileContent {
		fileContent[i] = byte(i % 256)
	}

	// Upload file
	fileID, err := s.uploadFile(s.agent1URL, "test-progress.bin", fileContent)
	s.Require().NoError(err)

	// Initiate transfer
	transferReq := FileTransferRequest{
		ID:            uuid.New().String(),
		FilePath:      fileID,
		DestAgentName: "agent2",
		FileName:      "test-progress.bin",
		FileSize:      fileSize,
	}

	transferID, err := s.initiateTransfer(s.agent1URL, transferReq)
	s.Require().NoError(err)

	// Monitor progress
	progress := s.getTransferProgress(s.agent1URL, transferID)
	s.Require().NotNil(progress)
	s.Assert().NotEmpty(progress.TransferID)
	s.Assert().GreaterOrEqual(progress.PercentComplete, 0.0)
	s.Assert().LessOrEqual(progress.PercentComplete, 100.0)
}

// TestConcurrentTransfers tests multiple simultaneous file transfers
func (s *FileTransferIntegrationTestSuite) TestConcurrentTransfers() {
	fileCount := 3
	fileSize := int64(512 * 1024) // 512 KB each

	transferIDs := make([]string, fileCount)
	fileContents := make([][]byte, fileCount)

	// Start multiple transfers
	for i := 0; i < fileCount; i++ {
		fileContent := make([]byte, fileSize)
		for j := range fileContent {
			fileContent[j] = byte((i*fileSize + j) % 256)
		}
		fileContents[i] = fileContent

		fileID, err := s.uploadFile(s.agent1URL, fmt.Sprintf("test-concurrent-%d.bin", i), fileContent)
		s.Require().NoError(err)

		transferReq := FileTransferRequest{
			ID:            uuid.New().String(),
			FilePath:      fileID,
			DestAgentName: "agent3",
			FileName:      fmt.Sprintf("test-concurrent-%d.bin", i),
			FileSize:      fileSize,
		}

		transferID, err := s.initiateTransfer(s.agent1URL, transferReq)
		s.Require().NoError(err)
		transferIDs[i] = transferID
	}

	// Wait for all transfers to complete
	for i, transferID := range transferIDs {
		s.waitForTransferComplete(s.agent1URL, transferID, 1*time.Minute)
		
		// Verify file received
		receivedContent, err := s.downloadFile(s.agent3URL, fmt.Sprintf("test-concurrent-%d.bin", i))
		s.Require().NoError(err)
		s.Assert().Equal(fileContents[i], receivedContent)
	}
}

// TestTransferWithCompression tests file transfer with compression enabled
func (s *FileTransferIntegrationTestSuite) TestTransferWithCompression() {
	// Create compressible file (text data)
	fileSize := int64(500 * 1024)
	fileContent := make([]byte, fileSize)
	// Fill with repeating pattern for good compression
	pattern := []byte("This is compressible text data. ")
	for i := 0; i < int(fileSize); i += len(pattern) {
		copy(fileContent[i:], pattern)
	}

	// Upload file
	fileID, err := s.uploadFile(s.agent1URL, "test-compress.txt", fileContent)
	s.Require().NoError(err)

	// Initiate transfer with compression
	transferReq := FileTransferRequest{
		ID:            uuid.New().String(),
		FilePath:      fileID,
		DestAgentName: "agent2",
		FileName:      "test-compress.txt",
		FileSize:      fileSize,
		Compress:      true,
	}

	transferID, err := s.initiateTransfer(s.agent1URL, transferReq)
	s.Require().NoError(err)

	// Wait for transfer
	s.waitForTransferComplete(s.agent1URL, transferID, 1*time.Minute)

	// Verify file received (decompressed automatically)
	receivedContent, err := s.downloadFile(s.agent2URL, "test-compress.txt")
	s.Require().NoError(err)
	s.Assert().Equal(fileContent, receivedContent)
}

// TestTransferToMultipleReceivers tests sending same file to multiple agents
func (s *FileTransferIntegrationTestSuite) TestTransferToMultipleReceivers() {
	// Create test file
	fileSize := int64(256 * 1024)
	fileContent := make([]byte, fileSize)
	for i := range fileContent {
		fileContent[i] = byte(i % 256)
	}

	// Upload file to agent1
	fileID, err := s.uploadFile(s.agent1URL, "test-multi.bin", fileContent)
	s.Require().NoError(err)

	// Transfer to agent2
	transferReq1 := FileTransferRequest{
		ID:            uuid.New().String(),
		FilePath:      fileID,
		DestAgentName: "agent2",
		FileName:      "test-multi.bin",
		FileSize:      fileSize,
	}
	transferID1, err := s.initiateTransfer(s.agent1URL, transferReq1)
	s.Require().NoError(err)

	// Transfer to agent3
	transferReq2 := FileTransferRequest{
		ID:            uuid.New().String(),
		FilePath:      fileID,
		DestAgentName: "agent3",
		FileName:      "test-multi.bin",
		FileSize:      fileSize,
	}
	transferID2, err := s.initiateTransfer(s.agent1URL, transferReq2)
	s.Require().NoError(err)

	// Wait for both transfers
	s.waitForTransferComplete(s.agent1URL, transferID1, 1*time.Minute)
	s.waitForTransferComplete(s.agent1URL, transferID2, 1*time.Minute)

	// Verify both received
	received2, err := s.downloadFile(s.agent2URL, "test-multi.bin")
	s.Require().NoError(err)
	s.Assert().Equal(fileContent, received2)

	received3, err := s.downloadFile(s.agent3URL, "test-multi.bin")
	s.Require().NoError(err)
	s.Assert().Equal(fileContent, received3)
}

// TestTransferStatus tests transfer status endpoint
func (s *FileTransferIntegrationTestSuite) TestTransferStatus() {
	// Create test file
	fileSize := int64(100 * 1024)
	fileContent := make([]byte, fileSize)
	for i := range fileContent {
		fileContent[i] = byte(i % 256)
	}

	fileID, err := s.uploadFile(s.agent1URL, "test-status.bin", fileContent)
	s.Require().NoError(err)

	transferReq := FileTransferRequest{
		ID:            uuid.New().String(),
		FilePath:      fileID,
		DestAgentName: "agent2",
		FileName:      "test-status.bin",
		FileSize:      fileSize,
	}

	transferID, err := s.initiateTransfer(s.agent1URL, transferReq)
	s.Require().NoError(err)

	// Get transfer history
	history := s.getTransferHistory(s.agent1URL)
	s.Require().NotEmpty(history)

	// Find our transfer in history
	found := false
	for _, entry := range history {
		if entry["transfer_id"] == transferID {
			found = true
			break
		}
	}
	s.Assert().True(found, "Transfer not found in history")
}

// TestTransferFileIntegrity tests that file integrity is maintained
func (s *FileTransferIntegrationTestSuite) TestTransferFileIntegrity() {
	// Create test file with known hash pattern
	fileSize := int64(1 * 1024 * 1024)
	fileContent := make([]byte, fileSize)
	
	// Create deterministic pattern
	for i := range fileContent {
		fileContent[i] = byte((i * 7) % 256)
	}

	fileID, err := s.uploadFile(s.agent1URL, "test-integrity.bin", fileContent)
	s.Require().NoError(err)

	transferReq := FileTransferRequest{
		ID:            uuid.New().String(),
		FilePath:      fileID,
		DestAgentName: "agent3",
		FileName:      "test-integrity.bin",
		FileSize:      fileSize,
	}

	transferID, err := s.initiateTransfer(s.agent1URL, transferReq)
	s.Require().NoError(err)

	s.waitForTransferComplete(s.agent1URL, transferID, 1*time.Minute)

	// Verify integrity
	receivedContent, err := s.downloadFile(s.agent3URL, "test-integrity.bin")
	s.Require().NoError(err)
	
	// Verify byte-for-byte match
	s.Assert().Equal(len(fileContent), len(receivedContent))
	s.Assert().Equal(fileContent, receivedContent)
}

// TestTransferVerySmallFile tests edge case of very small file
func (s *FileTransferIntegrationTestSuite) TestTransferVerySmallFile() {
	fileContent := []byte("Hello, World!")

	fileID, err := s.uploadFile(s.agent1URL, "test-tiny.txt", fileContent)
	s.Require().NoError(err)

	transferReq := FileTransferRequest{
		ID:            uuid.New().String(),
		FilePath:      fileID,
		DestAgentName: "agent2",
		FileName:      "test-tiny.txt",
		FileSize:      int64(len(fileContent)),
	}

	transferID, err := s.initiateTransfer(s.agent1URL, transferReq)
	s.Require().NoError(err)

	s.waitForTransferComplete(s.agent1URL, transferID, 30*time.Second)

	receivedContent, err := s.downloadFile(s.agent2URL, "test-tiny.txt")
	s.Require().NoError(err)
	s.Assert().Equal(fileContent, receivedContent)
}

// Helper functions

func (s *FileTransferIntegrationTestSuite) uploadFile(agentURL, filename string, content []byte) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", err
	}
	_, err = part.Write(content)
	if err != nil {
		return "", err
	}
	
	err = writer.Close()
	if err != nil {
		return "", err
	}

	resp, err := s.client.Post(agentURL+"/files/upload", writer.FormDataContentType(), body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		FileID string `json:"file_id"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	return result.FileID, nil
}

func (s *FileTransferIntegrationTestSuite) initiateTransfer(agentURL string, req FileTransferRequest) (string, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	resp, err := s.client.Post(agentURL+"/transfer/initiate", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		TransferID string `json:"transfer_id"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	return result.TransferID, nil
}

func (s *FileTransferIntegrationTestSuite) waitForTransferComplete(agentURL, transferID string, timeout time.Duration) {
	s.Eventually(func() bool {
		progress := s.getTransferProgress(agentURL, transferID)
		if progress == nil {
			return false
		}
		return progress.Status == "completed" || progress.Status == "failed"
	}, timeout, 500*time.Millisecond)
}

func (s *FileTransferIntegrationTestSuite) getTransferProgress(agentURL, transferID string) *TransferProgress {
	resp, err := s.client.Get(fmt.Sprintf("%s/transfer/progress?id=%s", agentURL, transferID))
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var progress TransferProgress
	json.NewDecoder(resp.Body).Decode(&progress)

	return &progress
}

func (s *FileTransferIntegrationTestSuite) getTransferHistory(agentURL string) []map[string]interface{} {
	resp, err := s.client.Get(agentURL + "/transfer/history")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var result struct {
		History []map[string]interface{} `json:"history"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	return result.History
}

func (s *FileTransferIntegrationTestSuite) downloadFile(agentURL, filename string) ([]byte, error) {
	resp, err := s.client.Get(fmt.Sprintf("%s/files/download?name=%s", agentURL, filename))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TestFileTransferIntegration runs all file transfer integration tests
func TestFileTransferIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	suite.Run(t, new(FileTransferIntegrationTestSuite))
}
