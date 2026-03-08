# Task 032: Unit Tests - TailFS

## Priority
**P0** - Critical

## Status
⏳ Pending

## Objective
Write comprehensive unit tests for TailFS core components to ensure reliable file transfers.

## Background
File transfer code must be thoroughly tested to prevent data loss and ensure reliability. This task covers unit testing of all TailFS components.

## Requirements

### Coverage Goals
- [ ] Overall coverage > 80%
- [ ] Transfer logic > 90%
- [ ] All public APIs tested
- [ ] Error paths covered

### Components to Test

#### Models (internal/models/)
- [ ] file_transfer.go - Transfer data structures
- [ ] Test all struct methods
- [ ] Test JSON marshaling

#### Services (internal/services/)
- [ ] transfer_service.go - Core transfer logic
- [ ] chunk_protocol.go - Chunking (when implemented)
- [ ] resume.go - Resume support (when implemented)

### Test Infrastructure
- [ ] Mock file system
- [ ] Test file generators
- [ ] Network mocks
- [ ] Progress tracking mocks

## Technical Specification

### Mock File System
```go
// MockFS for testing without disk I/O
type MockFS struct {
    Files map[string][]byte
    Mu    sync.Mutex
}

func (m *MockFS) Create(path string) (io.WriteCloser, error) {
    m.Files[path] = []byte{}
    return &MockFile{Path: path, FS: m}, nil
}

func (m *MockFS) Open(path string) (io.ReadCloser, error) {
    data, ok := m.Files[path]
    if !ok {
        return nil, os.ErrNotExist
    }
    return &MockFile{Path: path, Data: data, FS: m}, nil
}

// Mock file for testing
type MockFile struct {
    Path string
    Data []byte
    Pos  int64
    FS   *MockFS
}
```

### Test File Generator
```go
// Generate test files of various sizes
func GenerateTestFile(t *testing.T, size int64) (string, func()) {
    dir := t.TempDir()
    path := filepath.Join(dir, "testfile")
    
    f, err := os.Create(path)
    if err != nil {
        t.Fatal(err)
    }
    
    // Write deterministic data for hash verification
    hash := sha256.New()
    w := io.MultiWriter(f, hash)
    
    remaining := size
    buf := make([]byte, 1024*1024) // 1MB buffer
    for remaining > 0 {
        n := int64(len(buf))
        if n > remaining {
            n = remaining
        }
        // Fill with deterministic pattern
        for i := int64(0); i < n; i++ {
            buf[i] = byte((size + i) % 256)
        }
        if _, err := w.Write(buf[:n]); err != nil {
            t.Fatal(err)
        }
        remaining -= n
    }
    
    expectedHash := hex.EncodeToString(hash.Sum(nil))
    
    cleanup := func() {
        os.RemoveAll(dir)
    }
    
    return path, cleanup
}
```

### Test Patterns for Transfers
```go
func TestTransferService(t *testing.T) {
    tests := []struct {
        name        string
        fileSize    int64
        chunkSize   int64
        expectError bool
    }{
        {"small file", 1024, 1024*1024, false},
        {"exact chunk", 1024 * 1024, 1024 * 1024, false},
        {"multiple chunks", 10 * 1024 * 1024, 1024 * 1024, false},
        {"large file", 100 * 1024 * 1024, 1024 * 1024, false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            src, cleanup := GenerateTestFile(t, tt.fileSize)
            defer cleanup()
            
            config := &models.FileTransferConfig{
                ChunkSize: tt.chunkSize,
            }
            
            svc, err := services.NewFileTransferService(config)
            if err != nil {
                t.Fatal(err)
            }
            
            req := &models.FileTransferRequest{
                ID:       uuid.New().String(),
                FilePath: src,
            }
            
            err = svc.SendFile(context.Background(), req)
            if tt.expectError && err == nil {
                t.Error("expected error, got nil")
            }
            if !tt.expectError && err != nil {
                t.Errorf("unexpected error: %v", err)
            }
        })
    }
}
```

## Acceptance Criteria

### Coverage
- [ ] models/ > 85%
- [ ] services/ > 80%
- [ ] Overall > 80%

### Quality
- [ ] All tests pass consistently
- [ ] No flaky tests
- [ ] Tests run in < 60 seconds
- [ ] Clear test names
- [ ] Good error messages

## Testing Requirements

### Unit Tests
- [ ] File chunking tests
- [ ] Transfer state tests
- [ ] Progress tracking tests
- [ ] Error handling tests
- [ ] Hash verification tests

### Integration Tests (within unit test scope)
- [ ] End-to-end transfer tests
- [ ] Resume tests (when implemented)

## Implementation Steps

1. **Phase 1: Test Infrastructure** (1 day)
   - Set up test helpers
   - Create mock file system
   - Define test data generators

2. **Phase 2: Models Tests** (1 day)
   - Test file_transfer.go
   - Test JSON marshaling
   - Test helper functions

3. **Phase 3: Service Tests** (3 days)
   - Test transfer_service.go
   - Test chunking (when ready)
   - Test resume (when ready)

4. **Phase 4: Coverage Improvement** (1 day)
   - Identify gaps
   - Add missing tests
   - Improve assertions

## Files to Create

### Test Files
- `tail-agent-file-send/internal/models/file_transfer_test.go`
- `tail-agent-file-send/internal/services/transfer_service_test.go`
- `tail-agent-file-send/internal/services/chunk_protocol_test.go` (when ready)
- `tail-agent-file-send/internal/services/resume_test.go` (when ready)

### Test Helpers
- `tail-agent-file-send/internal/testutil/mocks.go`
- `tail-agent-file-send/internal/testutil/testdata.go`
- `tail-agent-file-send/internal/testutil/mock_fs.go`

## References
- [Go Testing Package](https://pkg.go.dev/testing)
- [Testify](https://github.com/stretchr/testify) (optional)
- [Task 011: Chunk Protocol](task_011_tailfs_chunk_protocol.md)

## Assignment
**Agent:** [Unassigned]  
**Assigned:** [Date]  
**Due:** [Date]  

## Progress Log
- [Date]: [Update]
