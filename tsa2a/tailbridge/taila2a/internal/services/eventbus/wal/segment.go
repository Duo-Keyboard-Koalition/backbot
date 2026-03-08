package wal

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// segmentManager manages log segments for a single partition
type segmentManager struct {
	mu            sync.RWMutex
	partitionDir  string
	partitionID   int
	segments      []*logSegment
	activeSegment *logSegment
	config        *Config
	baseOffset    int64
	nextOffset    int64
}

// logSegment represents a single log segment file
type logSegment struct {
	id          int
	baseOffset  int64
	file        *os.File
	index       *offsetIndex
	size        int64
	maxSize     int64
	isActive    bool
}

// offsetIndex provides fast offset lookup
type offsetIndex struct {
	file       *os.File
	entries    []indexEntry
	maxEntries int
}

// indexEntry is a single index entry
type indexEntry struct {
	relativeOffset int64
	position       int64
}

const (
	indexEntrySize = 16 // 8 bytes offset + 8 bytes position
)

// newSegmentManager creates a new segment manager
func newSegmentManager(partitionDir string, partitionID int, config *Config) (*segmentManager, error) {
	sm := &segmentManager{
		partitionDir: partitionDir,
		partitionID:  partitionID,
		segments:     make([]*logSegment, 0),
		config:       config,
	}

	// Load existing segments
	if err := sm.loadSegments(); err != nil {
		return nil, err
	}

	// Create active segment if none exists
	if sm.activeSegment == nil {
		if err := sm.createSegment(); err != nil {
			return nil, err
		}
	}

	return sm, nil
}

// loadSegments loads existing segment files
func (sm *segmentManager) loadSegments() error {
	files, err := filepath.Glob(filepath.Join(sm.partitionDir, "*.log"))
	if err != nil {
		return err
	}

	for _, file := range files {
		segment, err := sm.loadSegment(file)
		if err != nil {
			continue // Skip corrupted segments
		}

		sm.segments = append(sm.segments, segment)

		if segment.baseOffset+int64(segment.index.len()) > sm.baseOffset {
			sm.baseOffset = segment.baseOffset
			sm.nextOffset = segment.baseOffset + int64(segment.index.len())
		}
	}

	// Set last segment as active
	if len(sm.segments) > 0 {
		last := sm.segments[len(sm.segments)-1]
		last.isActive = true
		sm.activeSegment = last
	}

	return nil
}

// loadSegment loads a single segment file
func (sm *segmentManager) loadSegment(path string) (*logSegment, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	// Extract base offset from filename
	var baseOffset int64
	fmt.Sscanf(filepath.Base(path), "%020d.log", &baseOffset)

	// Get file size
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	// Scan file to count entries and build index
	data, err := os.ReadFile(path)
	if err != nil {
		file.Close()
		return nil, err
	}

	entryCount := 0
	pos := int64(0)
	for pos < info.Size() {
		if pos+28 > info.Size() {
			break
		}

		valueLen := binary.BigEndian.Uint32(data[pos+8 : pos+12])
		keyLen := binary.BigEndian.Uint32(data[pos+16 : pos+20])
		totalSize := 28 + int64(keyLen) + int64(valueLen)

		if pos+totalSize > info.Size() {
			break
		}

		// Add to index at intervals
		if entryCount%sm.config.IndexInterval == 0 {
			sm.createIndex(baseOffset) // Ensure index exists
		}

		entryCount++
		pos += totalSize
	}

	// Load or create index
	index, err := sm.loadIndex(baseOffset)
	if err != nil {
		file.Close()
		return nil, err
	}

	return &logSegment{
		id:         int(baseOffset),
		baseOffset: baseOffset,
		file:       file,
		index:      index,
		size:       info.Size(),
		maxSize:    sm.config.SegmentSize,
		isActive:   false,
	}, nil
}

// loadIndex loads an index file
func (sm *segmentManager) loadIndex(baseOffset int64) (*offsetIndex, error) {
	indexPath := filepath.Join(sm.partitionDir, fmt.Sprintf("%020d.index", baseOffset))

	file, err := os.OpenFile(indexPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	idx := &offsetIndex{
		file:       file,
		entries:    make([]indexEntry, 0),
		maxEntries: int(sm.config.SegmentSize) / indexEntrySize,
	}

	// Read existing index entries
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	numEntries := info.Size() / indexEntrySize
	for i := int64(0); i < numEntries; i++ {
		buf := make([]byte, indexEntrySize)
		_, err := file.ReadAt(buf, i*indexEntrySize)
		if err != nil {
			if err == io.EOF {
				break
			}
			file.Close()
			return nil, err
		}

		entry := indexEntry{
			relativeOffset: int64(binary.BigEndian.Uint64(buf[0:8])),
			position:       int64(binary.BigEndian.Uint64(buf[8:16])),
		}
		idx.entries = append(idx.entries, entry)
	}

	return idx, nil
}

// createSegment creates a new log segment
func (sm *segmentManager) createSegment() error {
	baseOffset := sm.nextOffset

	filename := fmt.Sprintf("%020d.log", baseOffset)
	path := filepath.Join(sm.partitionDir, filename)

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	// Create index
	index, err := sm.createIndex(baseOffset)
	if err != nil {
		file.Close()
		return err
	}

	segment := &logSegment{
		id:         int(baseOffset),
		baseOffset: baseOffset,
		file:       file,
		index:      index,
		size:       0,
		maxSize:    sm.config.SegmentSize,
		isActive:   true,
	}

	// Deactivate previous segment
	if sm.activeSegment != nil {
		sm.activeSegment.isActive = false
	}

	sm.segments = append(sm.segments, segment)
	sm.activeSegment = segment

	return nil
}

// createIndex creates a new index file
func (sm *segmentManager) createIndex(baseOffset int64) (*offsetIndex, error) {
	indexPath := filepath.Join(sm.partitionDir, fmt.Sprintf("%020d.index", baseOffset))

	file, err := os.OpenFile(indexPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}

	return &offsetIndex{
		file:       file,
		entries:    make([]indexEntry, 0),
		maxEntries: int(sm.config.SegmentSize) / indexEntrySize,
	}, nil
}

// append appends an entry to the active segment
func (sm *segmentManager) append(entry *LogEntry) (int64, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.activeSegment == nil {
		if err := sm.createSegment(); err != nil {
			return -1, err
		}
	}

	// Encode entry
	wal := &walImpl{crcTable: crc32.MakeTable(crc32.Castagnoli)}
	data, err := wal.encodeEntry(entry)
	if err != nil {
		return -1, err
	}

	// Check if segment is full
	if sm.activeSegment.size+int64(len(data)) > sm.activeSegment.maxSize {
		if err := sm.createSegment(); err != nil {
			return -1, err
		}
	}

	// Write to file
	position := sm.activeSegment.size
	_, err = sm.activeSegment.file.Write(data)
	if err != nil {
		return -1, err
	}

	sm.activeSegment.size += int64(len(data))

	// Add to index
	relativeOffset := sm.nextOffset - sm.activeSegment.baseOffset
	if len(sm.activeSegment.index.entries)%sm.config.IndexInterval == 0 {
		sm.activeSegment.index.add(relativeOffset, position)
	}

	offset := sm.nextOffset
	sm.nextOffset++

	return offset, nil
}

// read reads an entry at the given offset
func (sm *segmentManager) read(offset int64) (*LogEntry, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Find segment
	segment := sm.findSegment(offset)
	if segment == nil {
		return nil, ErrNotFound
	}

	// Scan from beginning of segment to find the entry
	return sm.scanForEntry(segment, offset)
}

// scanForEntry scans a segment for an entry at the given offset
func (sm *segmentManager) scanForEntry(segment *logSegment, targetOffset int64) (*LogEntry, error) {
	// Read the entire file and scan
	data, err := os.ReadFile(segment.file.Name())
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, ErrNotFound
	}

	pos := int64(0)
	for pos < int64(len(data)) {
		// Read header (28 bytes: offset(8) + size(4) + crc(4) + keyLen(4) + timestamp(8))
		if pos+28 > int64(len(data)) {
			break
		}

		entryOffset := int64(binary.BigEndian.Uint64(data[pos : pos+8]))
		valueLen := binary.BigEndian.Uint32(data[pos+8 : pos+12])
		keyLen := binary.BigEndian.Uint32(data[pos+16 : pos+20])
		
		// Sanity check
		if keyLen > 1024*1024 || valueLen > 10*1024*1024 {
			break
		}
		
		// Calculate total entry size
		totalSize := 28 + int64(keyLen) + int64(valueLen)
		
		if pos+totalSize > int64(len(data)) {
			break
		}
		
		if entryOffset == targetOffset {
			// Read full entry
			entryData := data[pos : pos+totalSize]
			wal := &walImpl{crcTable: crc32.MakeTable(crc32.Castagnoli)}
			return wal.decodeEntry(entryData)
		}
		
		pos += totalSize
	}

	return nil, ErrNotFound
}

// readEntryAt reads an entry at a specific position
func (sm *segmentManager) readEntryAt(segment *logSegment, position int64, targetOffset int64) (*LogEntry, error) {
	// Read header
	header := make([]byte, 16) // offset + size
	_, err := segment.file.ReadAt(header, position)
	if err != nil {
		return nil, err
	}

	entryOffset := int64(binary.BigEndian.Uint64(header[0:8]))
	if entryOffset != targetOffset {
		return nil, ErrNotFound
	}

	valueLen := binary.BigEndian.Uint32(header[8:12])

	// Read full entry
	totalSize := 16 + 4 + 4 + 8 + int(valueLen) // header + crc + keyLen + timestamp + key + value
	data := make([]byte, totalSize)
	_, err = segment.file.ReadAt(data, position)
	if err != nil {
		return nil, err
	}

	// Decode
	wal := &walImpl{crcTable: crc32.MakeTable(crc32.Castagnoli)}
	return wal.decodeEntry(data)
}

// readRange reads entries in an offset range
func (sm *segmentManager) readRange(start, end int64) ([]*LogEntry, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var entries []*LogEntry

	for offset := start; offset <= end; offset++ {
		entry, err := sm.read(offset)
		if err != nil {
			if err == ErrNotFound {
				// Continue to next offset
				continue
			}
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// truncate removes entries before the given offset
func (sm *segmentManager) truncate(offset int64) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Find and remove old segments
	newSegments := make([]*logSegment, 0)
	for _, segment := range sm.segments {
		if segment.baseOffset+int64(segment.index.len()) <= offset {
			// This segment is entirely before the truncation point
			segment.close()
			os.Remove(segment.file.Name())
			os.Remove(segment.index.file.Name())
		} else {
			newSegments = append(newSegments, segment)
		}
	}

	sm.segments = newSegments
	if len(sm.segments) > 0 {
		sm.activeSegment = sm.segments[len(sm.segments)-1]
	}

	return nil
}

// sync flushes pending writes to disk
func (sm *segmentManager) sync() error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for _, segment := range sm.segments {
		if err := segment.sync(); err != nil {
			return err
		}
	}

	return nil
}

// close closes the segment manager
func (sm *segmentManager) close() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, segment := range sm.segments {
		segment.close()
	}

	return nil
}

// findSegment finds the segment containing the offset
func (sm *segmentManager) findSegment(offset int64) *logSegment {
	for i := len(sm.segments) - 1; i >= 0; i-- {
		segment := sm.segments[i]
		if offset >= segment.baseOffset {
			return segment
		}
	}
	return nil
}

// logSegment methods

// sync flushes the segment to disk
func (s *logSegment) sync() error {
	if err := s.file.Sync(); err != nil {
		return err
	}
	return s.index.sync()
}

// close closes the segment
func (s *logSegment) close() error {
	var err error
	if s.file != nil {
		err = s.file.Close()
	}
	if s.index != nil {
		s.index.close()
	}
	return err
}

// offsetIndex methods

// add adds an entry to the index
func (idx *offsetIndex) add(relativeOffset int64, position int64) {
	idx.entries = append(idx.entries, indexEntry{
		relativeOffset: relativeOffset,
		position:       position,
	})
}

// len returns the number of entries
func (idx *offsetIndex) len() int {
	return len(idx.entries)
}

// lookup finds the position for an offset
func (idx *offsetIndex) lookup(relativeOffset int64) int64 {
	if len(idx.entries) == 0 {
		return -1
	}

	// Binary search
	left, right := 0, len(idx.entries)-1
	result := int64(-1)

	for left <= right {
		mid := (left + right) / 2
		entry := idx.entries[mid]

		if entry.relativeOffset <= relativeOffset {
			result = entry.position
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	return result
}

// sync flushes the index to disk
func (idx *offsetIndex) sync() error {
	if idx.file == nil {
		return nil
	}

	// Write entries
	data := make([]byte, len(idx.entries)*indexEntrySize)
	for i, entry := range idx.entries {
		pos := i * indexEntrySize
		binary.BigEndian.PutUint64(data[pos:], uint64(entry.relativeOffset))
		binary.BigEndian.PutUint64(data[pos+8:], uint64(entry.position))
	}

	_, err := idx.file.Write(data)
	if err != nil {
		return err
	}

	return idx.file.Sync()
}

// close closes the index
func (idx *offsetIndex) close() error {
	if idx.file != nil {
		return idx.file.Close()
	}
	return nil
}
