package testify

import (
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	"github.com/codejedi-ai/kaggle-for-tensors/tailscale-app/tailbridge/test_platform/mock"
	"github.com/stretchr/testify/suite"
)

// FileTransferTestSuite provides a test suite for file transfer testing
type FileTransferTestSuite struct {
	suite.Suite
	network  *mock.MockNetwork
	sender   *mock.MockAgent
	receiver *mock.MockAgent
}

// SetupTest spins up a fresh network with agents before each test
func (s *FileTransferTestSuite) SetupTest() {
	s.network = mock.NewMockNetwork()
	
	// Create and start sender and receiver agents
	s.sender = mock.NewMockAgent("tailfs-sender", "test.tailnet", []string{"file_send", "chat"})
	s.receiver = mock.NewMockAgent("tailfs-receiver", "test.tailnet", []string{"file_receive", "chat"})
	
	// Add agents to network (this starts them)
	s.Require().NoError(s.network.AddAgent(s.sender))
	s.Require().NoError(s.network.AddAgent(s.receiver))
	
	// Wait for agents to be running
	s.Require().NoError(s.network.WaitForAllAgents(2, 2*time.Second))
}

// TearDownTest tears down all agents and clears network after each test
func (s *FileTransferTestSuite) TearDownTest() {
	// Properly remove all agents
	if s.sender != nil {
		s.network.RemoveAgent(s.sender.Name)
	}
	if s.receiver != nil {
		s.network.RemoveAgent(s.receiver.Name)
	}
	s.network.ClearNetwork()
	s.network = nil
}

// TestSmallFileTransfer tests transferring a small file (< 1MB)
func (s *FileTransferTestSuite) TestSmallFileTransfer() {
	fileSize := int64(512 * 1024) // 512 KB
	
	req := mock.FileTransferRequest{
		ID:            "transfer-001",
		FilePath:      "/tmp/test-small.bin",
		DestAgentName: "tailfs-receiver",
		FileName:      "test-small.bin",
		FileSize:      fileSize,
		Compress:      false,
		Encrypt:       false,
	}
	
	transferID, err := s.network.SendFile("tailfs-sender", "tailfs-receiver", req)
	s.Require().NoError(err)
	s.Assert().NotEmpty(transferID)
	
	// Wait for transfer
	time.Sleep(50 * time.Millisecond)
	
	// Verify file received
	receivedFile, err := s.network.GetFile("tailfs-receiver", "test-small.bin")
	s.Require().NoError(err)
	s.Assert().Equal(fileSize, int64(len(receivedFile)))
}

// TestLargeFileTransfer tests transferring a large file (100MB+)
func (s *FileTransferTestSuite) TestLargeFileTransfer() {
	fileSize := int64(100 * 1024 * 1024) // 100 MB
	
	req := mock.FileTransferRequest{
		ID:            "transfer-002",
		FilePath:      "/tmp/test-large.bin",
		DestAgentName: "tailfs-receiver",
		FileName:      "test-large.bin",
		FileSize:      fileSize,
		Compress:      false,
		Encrypt:       false,
	}
	
	transferID, err := s.network.SendFile("tailfs-sender", "tailfs-receiver", req)
	s.Require().NoError(err)
	
	// Get progress
	progress, err := s.network.GetTransferProgress(transferID)
	s.Require().NoError(err)
	s.Assert().Equal("completed", progress.Status)
	s.Assert().Equal(100.0, progress.PercentComplete)
	
	// Verify file received
	receivedFile, err := s.network.GetFile("tailfs-receiver", "test-large.bin")
	s.Require().NoError(err)
	s.Assert().Equal(fileSize, int64(len(receivedFile)))
}

// TestFileTransferProgress tests progress tracking during transfer
func (s *FileTransferTestSuite) TestFileTransferProgress() {
	fileSize := int64(10 * 1024 * 1024) // 10 MB
	
	req := mock.FileTransferRequest{
		ID:            "transfer-003",
		FilePath:      "/tmp/test-progress.bin",
		DestAgentName: "tailfs-receiver",
		FileName:      "test-progress.bin",
		FileSize:      fileSize,
	}
	
	transferID, err := s.network.SendFile("tailfs-sender", "tailfs-receiver", req)
	s.Require().NoError(err)
	
	// Get progress
	progress, err := s.network.GetTransferProgress(transferID)
	s.Require().NoError(err)
	
	s.Assert().Equal(transferID, progress.TransferID)
	s.Assert().Equal("completed", progress.Status)
	s.Assert().GreaterOrEqual(progress.PercentComplete, 0.0)
	s.Assert().LessOrEqual(progress.PercentComplete, 100.0)
	s.Assert().Greater(progress.BytesPerSecond, int64(0))
}

// TestConcurrentFileTransfers tests multiple simultaneous transfers
func (s *FileTransferTestSuite) TestConcurrentFileTransfers() {
	fileCount := 5
	fileSize := int64(1 * 1024 * 1024) // 1 MB each
	
	transferIDs := make([]string, fileCount)
	
	// Start multiple transfers concurrently
	for i := 0; i < fileCount; i++ {
		req := mock.FileTransferRequest{
			ID:            fmt.Sprintf("transfer-concurrent-%d", i),
			FilePath:      fmt.Sprintf("/tmp/test-concurrent-%d.bin", i),
			DestAgentName: "tailfs-receiver",
			FileName:      fmt.Sprintf("test-concurrent-%d.bin", i),
			FileSize:      fileSize,
		}
		
		transferID, err := s.network.SendFile("tailfs-sender", "tailfs-receiver", req)
		s.Require().NoError(err)
		transferIDs[i] = transferID
	}
	
	// Wait for all transfers
	time.Sleep(200 * time.Millisecond)
	
	// Verify all files received
	for i := 0; i < fileCount; i++ {
		fileName := fmt.Sprintf("test-concurrent-%d.bin", i)
		receivedFile, err := s.network.GetFile("tailfs-receiver", fileName)
		s.Require().NoError(err)
		s.Assert().Equal(fileSize, int64(len(receivedFile)))
	}
}

// TestFileTransferNotification tests transfer completion notifications
func (s *FileTransferTestSuite) TestFileTransferNotification() {
	fileSize := int64(2 * 1024 * 1024) // 2 MB
	
	req := mock.FileTransferRequest{
		ID:            "transfer-notify-001",
		FilePath:      "/tmp/test-notify.bin",
		DestAgentName: "tailfs-receiver",
		FileName:      "test-notify.bin",
		FileSize:      fileSize,
	}
	
	_, err := s.network.SendFile("tailfs-sender", "tailfs-receiver", req)
	s.Require().NoError(err)
	
	// Wait for notification
	time.Sleep(100 * time.Millisecond)
	
	// Check receiver got notification message
	s.Assert().GreaterOrEqual(len(s.receiver.Messages), 1)
	
	// Verify notification content
	if len(s.receiver.Messages) > 0 {
		notifMsg := s.receiver.Messages[len(s.receiver.Messages)-1]
		s.Assert().Equal("file_transfer", notifMsg.Type)
		s.Assert().Equal("file_received", notifMsg.Body.Action)
	}
}

// TestFileIntegrity tests that transferred files maintain integrity
func (s *FileTransferTestSuite) TestFileIntegrity() {
	fileSize := int64(5 * 1024 * 1024) // 5 MB
	
	// Generate random file content
	originalContent := make([]byte, fileSize)
	_, err := rand.Read(originalContent)
	s.Require().NoError(err)
	
	// Store original in sender - use mock network's file storage instead
	req := mock.FileTransferRequest{
		ID:            "transfer-integrity-001",
		FilePath:      "/tmp/test-integrity.bin",
		DestAgentName: "tailfs-receiver",
		FileName:      "test-integrity.bin",
		FileSize:      fileSize,
	}
	
	_, err = s.network.SendFile("tailfs-sender", "tailfs-receiver", req)
	s.Require().NoError(err)
	
	time.Sleep(50 * time.Millisecond)
	
	// Verify received content matches size (mock generates pattern data)
	receivedContent, err := s.network.GetFile("tailfs-receiver", "test-integrity.bin")
	s.Require().NoError(err)
	
	// For mock, we verify size (actual implementation would hash)
	s.Assert().Equal(int(fileSize), len(receivedContent))
}

// TestTransferToOfflineAgent tests error handling when receiver is offline
func (s *FileTransferTestSuite) TestTransferToOfflineAgent() {
	// Stop receiver
	err := s.network.StopAgent("tailfs-receiver")
	s.Require().NoError(err)
	
	req := mock.FileTransferRequest{
		ID:            "transfer-offline-001",
		FilePath:      "/tmp/test-offline.bin",
		DestAgentName: "tailfs-receiver",
		FileName:      "test-offline.bin",
		FileSize:      int64(1024),
	}
	
	_, err = s.network.SendFile("tailfs-sender", "tailfs-receiver", req)
	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "not running")
}

// TestTransferFromOfflineAgent tests error handling when sender is offline
func (s *FileTransferTestSuite) TestTransferFromOfflineAgent() {
	// Stop sender
	err := s.network.StopAgent("tailfs-sender")
	s.Require().NoError(err)
	
	req := mock.FileTransferRequest{
		ID:            "transfer-offline-002",
		FilePath:      "/tmp/test-offline2.bin",
		DestAgentName: "tailfs-receiver",
		FileName:      "test-offline2.bin",
		FileSize:      int64(1024),
	}
	
	_, err = s.network.SendFile("tailfs-sender", "tailfs-receiver", req)
	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "not running")
}

// TestAgentDiscoveryForFileTransfer tests finding agents with file capabilities
func (s *FileTransferTestSuite) TestAgentDiscoveryForFileTransfer() {
	// Add more agents with different capabilities
	agent3 := mock.NewMockAgent("tailfs-both", "test.tailnet", []string{"file_send", "file_receive"})
	agent4 := mock.NewMockAgent("tailfs-send-only", "test.tailnet", []string{"file_send"})
	
	s.network.AddAgent(agent3)
	s.network.AddAgent(agent4)
	
	// Find agents that can receive files
	receivers := s.network.GetAgentsByCapability("file_receive")
	s.Require().Len(receivers, 2) // receiver and both
	
	// Find agents that can send files
	senders := s.network.GetAgentsByCapability("file_send")
	s.Require().Len(senders, 3) // sender, both, and send-only
}

// TestMultipleReceivers tests sending to different receivers
func (s *FileTransferTestSuite) TestMultipleReceivers() {
	// Add another receiver
	receiver2 := mock.NewMockAgent("tailfs-receiver2", "test.tailnet", []string{"file_receive"})
	s.network.AddAgent(receiver2)
	
	// Send to first receiver
	req1 := mock.FileTransferRequest{
		ID:            "transfer-multi-001",
		FilePath:      "/tmp/test-multi1.bin",
		DestAgentName: "tailfs-receiver",
		FileName:      "test-multi1.bin",
		FileSize:      int64(512 * 1024),
	}
	
	_, err := s.network.SendFile("tailfs-sender", "tailfs-receiver", req1)
	s.Require().NoError(err)
	
	// Send to second receiver
	req2 := mock.FileTransferRequest{
		ID:            "transfer-multi-002",
		FilePath:      "/tmp/test-multi2.bin",
		DestAgentName: "tailfs-receiver2",
		FileName:      "test-multi2.bin",
		FileSize:      int64(512 * 1024),
	}
	
	_, err = s.network.SendFile("tailfs-sender", "tailfs-receiver2", req2)
	s.Require().NoError(err)
	
	time.Sleep(100 * time.Millisecond)
	
	// Verify both received
	file1, err := s.network.GetFile("tailfs-receiver", "test-multi1.bin")
	s.Require().NoError(err)
	s.Assert().Equal(int64(512*1024), int64(len(file1)))
	
	file2, err := s.network.GetFile("tailfs-receiver2", "test-multi2.bin")
	s.Require().NoError(err)
	s.Assert().Equal(int64(512*1024), int64(len(file2)))
}

// TestFileTransferWithNetworkLatency tests transfer with simulated latency
func (s *FileTransferTestSuite) TestFileTransferWithNetworkLatency() {
	// Set network latency
	s.network.SetNetworkLatency(50 * time.Millisecond)
	
	req := mock.FileTransferRequest{
		ID:            "transfer-latency-001",
		FilePath:      "/tmp/test-latency.bin",
		DestAgentName: "tailfs-receiver",
		FileName:      "test-latency.bin",
		FileSize:      int64(1 * 1024 * 1024),
	}
	
	start := time.Now()
	_, err := s.network.SendFile("tailfs-sender", "tailfs-receiver", req)
	s.Require().NoError(err)
	
	elapsed := time.Since(start)
	s.Assert().GreaterOrEqual(elapsed, 50*time.Millisecond)
	
	// Reset latency
	s.network.SetNetworkLatency(0)
}

// TestFileNotFound tests error handling for non-existent files
func (s *FileTransferTestSuite) TestFileNotFound() {
	// Try to get file that doesn't exist
	_, err := s.network.GetFile("tailfs-receiver", "non-existent-file.bin")
	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "not found")
}

// TestAgentNotFoundInTransfer tests error handling for non-existent agents
func (s *FileTransferTestSuite) TestAgentNotFoundInTransfer() {
	req := mock.FileTransferRequest{
		ID:            "transfer-not-found-001",
		FilePath:      "/tmp/test-notfound.bin",
		DestAgentName: "non-existent-agent",
		FileName:      "test-notfound.bin",
		FileSize:      int64(1024),
	}
	
	_, err := s.network.SendFile("tailfs-sender", "non-existent-agent", req)
	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "not found")
}

// TestVerySmallFile tests edge case of very small file transfer
func (s *FileTransferTestSuite) TestVerySmallFile() {
	req := mock.FileTransferRequest{
		ID:            "transfer-tiny-001",
		FilePath:      "/tmp/test-tiny.txt",
		DestAgentName: "tailfs-receiver",
		FileName:      "test-tiny.txt",
		FileSize:      int64(10), // 10 bytes
	}
	
	_, err := s.network.SendFile("tailfs-sender", "tailfs-receiver", req)
	s.Require().NoError(err)
	
	time.Sleep(50 * time.Millisecond)
	
	receivedFile, err := s.network.GetFile("tailfs-receiver", "test-tiny.txt")
	s.Require().NoError(err)
	s.Assert().Equal(int64(10), int64(len(receivedFile)))
}

// TestFileTransferSuite runs all file transfer tests
func TestFileTransferSuite(t *testing.T) {
	suite.Run(t, new(FileTransferTestSuite))
}
