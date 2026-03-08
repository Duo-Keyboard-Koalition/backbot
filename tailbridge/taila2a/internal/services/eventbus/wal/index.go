package wal

import (
	"encoding/binary"
	"io"
	"os"
	"sync"
	"time"
)

// timeIndexEntry represents a time-based index entry
type timeIndexEntry struct {
	timestamp int64
	offset    int64
}

const timeIndexEntrySize = 16 // 8 bytes timestamp + 8 bytes offset

// timeIndex provides time-based offset lookup
type timeIndex struct {
	mu       sync.RWMutex
	file     *os.File
	entries  []timeIndexEntry
	maxSize  int64
	currentSize int64
}

// newTimeIndex creates a new time index
func newTimeIndex(path string, maxSize int64) (*timeIndex, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	idx := &timeIndex{
		file:    file,
		entries: make([]timeIndexEntry, 0),
		maxSize: maxSize,
	}

	// Load existing entries
	if err := idx.load(); err != nil {
		file.Close()
		return nil, err
	}

	return idx, nil
}

// load loads existing index entries
func (idx *timeIndex) load() error {
	info, err := idx.file.Stat()
	if err != nil {
		return err
	}

	numEntries := info.Size() / timeIndexEntrySize
	idx.currentSize = info.Size()

	for i := int64(0); i < numEntries; i++ {
		buf := make([]byte, timeIndexEntrySize)
		_, err := idx.file.ReadAt(buf, i*timeIndexEntrySize)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		entry := timeIndexEntry{
			timestamp: int64(binary.BigEndian.Uint64(buf[0:8])),
			offset:    int64(binary.BigEndian.Uint64(buf[8:16])),
		}
		idx.entries = append(idx.entries, entry)
	}

	return nil
}

// add adds a time index entry
func (idx *timeIndex) add(timestamp int64, offset int64) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Only add entry if timestamp is different from last entry
	if len(idx.entries) > 0 {
		lastEntry := idx.entries[len(idx.entries)-1]
		if lastEntry.timestamp == timestamp {
			return nil
		}
	}

	entry := timeIndexEntry{
		timestamp: timestamp,
		offset:    offset,
	}

	idx.entries = append(idx.entries, entry)
	idx.currentSize += timeIndexEntrySize

	return nil
}

// lookup finds the offset for a timestamp
func (idx *timeIndex) lookup(timestamp int64) int64 {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if len(idx.entries) == 0 {
		return -1
	}

	// Binary search for the largest timestamp <= target
	left, right := 0, len(idx.entries)-1
	result := idx.entries[0].offset

	for left <= right {
		mid := (left + right) / 2
		entry := idx.entries[mid]

		if entry.timestamp <= timestamp {
			result = entry.offset
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	return result
}

// sync flushes the index to disk
func (idx *timeIndex) sync() error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	if idx.file == nil {
		return nil
	}

	// Rewrite entire index
	data := make([]byte, len(idx.entries)*timeIndexEntrySize)
	for i, entry := range idx.entries {
		pos := i * timeIndexEntrySize
		binary.BigEndian.PutUint64(data[pos:], uint64(entry.timestamp))
		binary.BigEndian.PutUint64(data[pos+8:], uint64(entry.offset))
	}

	if err := idx.file.Truncate(0); err != nil {
		return err
	}

	if _, err := idx.file.WriteAt(data, 0); err != nil {
		return err
	}

	return idx.file.Sync()
}

// close closes the time index
func (idx *timeIndex) close() error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	if idx.file != nil {
		return idx.file.Close()
	}
	return nil
}

// LogMetadata holds metadata about a log segment
type LogMetadata struct {
	BaseOffset    int64
	EndOffset     int64
	Size          int64
	CreatedAt     int64
	LastModified  int64
	MessageCount  int64
}

// LogSegmentInfo provides information about a log segment
type LogSegmentInfo struct {
	Name          string
	BaseOffset    int64
	Size          int64
	IndexSize     int64
	TimeIndexSize int64
	CreatedAt     int64
}

// SegmentCleaner handles log segment cleanup and compaction
type SegmentCleaner struct {
	mu            sync.Mutex
	retentionMs   int64
	retentionSize int64
}

// NewSegmentCleaner creates a new segment cleaner
func NewSegmentCleaner(retentionMs, retentionSize int64) *SegmentCleaner {
	return &SegmentCleaner{
		retentionMs:   retentionMs,
		retentionSize: retentionSize,
	}
}

// Clean removes old segments based on retention policy
func (c *SegmentCleaner) Clean(segments []*logSegment) ([]*logSegment, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(segments) == 0 {
		return segments, nil
	}

	now := time.Now().UnixNano()
	cutoff := now - c.retentionMs

	// Find segments to delete
	toDelete := make([]*logSegment, 0)
	remaining := make([]*logSegment, 0)

	for _, segment := range segments {
		// Check retention time
		if c.retentionMs > 0 {
			info, err := segment.file.Stat()
			if err == nil {
				modTime := info.ModTime().UnixNano()
				if modTime < cutoff {
					toDelete = append(toDelete, segment)
					continue
				}
			}
		}

		remaining = append(remaining, segment)
	}

	// Check retention size
	if c.retentionSize > 0 {
		totalSize := int64(0)
		for i := len(remaining) - 1; i >= 0; i-- {
			segment := remaining[i]
			info, err := segment.file.Stat()
			if err == nil {
				totalSize += info.Size()
				if totalSize > c.retentionSize {
					toDelete = append(toDelete, segment)
					remaining = append(remaining[:i], remaining[i+1:]...)
				}
			}
		}
	}

	// Delete old segments
	for _, segment := range toDelete {
		segment.close()
		os.Remove(segment.file.Name())
		if segment.index != nil {
			os.Remove(segment.index.file.Name())
		}
	}

	return remaining, nil
}

// Compactor handles log compaction
type Compactor struct {
	mu sync.Mutex
}

// NewCompactor creates a new compactor
func NewCompactor() *Compactor {
	return &Compactor{}
}

// Compact performs log compaction
// For now, this is a placeholder - full compaction requires
// understanding message keys and keeping only the latest value per key
func (c *Compactor) Compact(segment *logSegment) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Placeholder - compaction to be implemented
	// This would involve:
	// 1. Reading all messages
	// 2. Keeping only the latest value for each key
	// 3. Writing compacted messages to a new segment
	// 4. Swapping the segments

	return nil
}
