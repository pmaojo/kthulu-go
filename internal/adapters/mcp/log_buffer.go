package mcp

import (
	"sync"
	"time"
)

// LogEntry represents a single log event from the application.
type LogEntry struct {
	Timestamp time.Time         `json:"timestamp"`
	Type      string            `json:"type"` // "http", "db", "test"
	Raw       string            `json:"raw,omitempty"`
	Data      map[string]string `json:"data,omitempty"`
}

// LogBuffer is a thread-safe ring buffer for log entries.
type LogBuffer struct {
	entries []LogEntry
	maxSize int
	head    int
	count   int
	mu      sync.RWMutex
}

// NewLogBuffer creates a buffer with the given capacity.
func NewLogBuffer(maxSize int) *LogBuffer {
	if maxSize <= 0 {
		maxSize = 100
	}
	return &LogBuffer{
		entries: make([]LogEntry, maxSize),
		maxSize: maxSize,
	}
}

// Push adds an entry to the buffer, evicting the oldest if full.
func (b *LogBuffer) Push(entry LogEntry) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.entries[b.head] = entry
	b.head = (b.head + 1) % b.maxSize
	if b.count < b.maxSize {
		b.count++
	}
}

// All returns all entries in chronological order (oldest first).
func (b *LogBuffer) All() []LogEntry {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.count == 0 {
		return nil
	}

	result := make([]LogEntry, b.count)
	start := 0
	if b.count == b.maxSize {
		start = b.head
	}

	for i := 0; i < b.count; i++ {
		idx := (start + i) % b.maxSize
		result[i] = b.entries[idx]
	}
	return result
}

// Since returns entries after the given timestamp.
func (b *LogBuffer) Since(t time.Time) []LogEntry {
	all := b.All()
	var result []LogEntry
	for _, entry := range all {
		if entry.Timestamp.After(t) {
			result = append(result, entry)
		}
	}
	return result
}

// Last returns the N most recent entries (newest first).
func (b *LogBuffer) Last(n int) []LogEntry {
	all := b.All()
	if len(all) == 0 {
		return nil
	}
	if n > len(all) {
		n = len(all)
	}
	// Reverse to get newest first
	result := make([]LogEntry, n)
	for i := 0; i < n; i++ {
		result[i] = all[len(all)-1-i]
	}
	return result
}

// Clear removes all entries from the buffer.
func (b *LogBuffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.head = 0
	b.count = 0
}

// Count returns the number of entries in the buffer.
func (b *LogBuffer) Count() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.count
}
