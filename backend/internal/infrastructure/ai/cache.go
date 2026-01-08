package ai

import (
	"container/list"
	"context"
	"sync"
	"time"
)

// CacheEntry represents a cached response with metadata
type CacheEntry struct {
	Prompt    string
	Response  string
	Tags      []string
	CreatedAt time.Time
	ExpiresAt time.Time
	Model     string
}

// LRUCache implements a thread-safe LRU cache with TTL and tag-based queries
type LRUCache struct {
	maxSize int
	ttl     time.Duration
	mu      sync.RWMutex
	cache   map[string]*list.Element
	lru     *list.List
	tags    map[string][]string // tags -> list of cache keys
}

type cacheNode struct {
	key   string
	entry *CacheEntry
}

// NewLRUCache creates a new LRU cache with the given capacity and TTL
func NewLRUCache(maxSize int, ttl time.Duration) *LRUCache {
	return &LRUCache{
		maxSize: maxSize,
		ttl:     ttl,
		cache:   make(map[string]*list.Element),
		lru:     list.New(),
		tags:    make(map[string][]string),
	}
}

// Set stores or updates a cache entry
func (c *LRUCache) Set(key string, entry *CacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If entry exists, remove it
	if elem, exists := c.cache[key]; exists {
		c.lru.Remove(elem)
		delete(c.cache, key)
	}

	// Add new entry
	entry.CreatedAt = time.Now()
	entry.ExpiresAt = entry.CreatedAt.Add(c.ttl)
	node := &cacheNode{key: key, entry: entry}
	elem := c.lru.PushFront(node)
	c.cache[key] = elem

	// Index tags
	for _, tag := range entry.Tags {
		c.tags[tag] = append(c.tags[tag], key)
	}

	// Evict oldest if over capacity
	if c.lru.Len() > c.maxSize {
		oldest := c.lru.Back()
		if oldest != nil {
			c.lru.Remove(oldest)
			oldNode := oldest.Value.(*cacheNode)
			delete(c.cache, oldNode.key)
			c.removeTagIndexes(oldNode.key)
		}
	}
}

// Get retrieves a cache entry if not expired
func (c *LRUCache) Get(key string) (*CacheEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	elem, exists := c.cache[key]
	if !exists {
		return nil, false
	}

	node := elem.Value.(*cacheNode)
	if time.Now().After(node.entry.ExpiresAt) {
		return nil, false
	}

	// Move to front (most recently used)
	c.lru.MoveToFront(elem)
	return node.entry, true
}

// GetByTag retrieves all non-expired entries with a given tag
func (c *LRUCache) GetByTag(tag string) []*CacheEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var results []*CacheEntry
	keys := c.tags[tag]
	now := time.Now()

	for _, key := range keys {
		if elem, exists := c.cache[key]; exists {
			node := elem.Value.(*cacheNode)
			if now.Before(node.entry.ExpiresAt) {
				results = append(results, node.entry)
			}
		}
	}
	return results
}

// Clear removes all entries
func (c *LRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*list.Element)
	c.lru.Init()
	c.tags = make(map[string][]string)
}

// removeTagIndexes removes a key from all tag indexes
func (c *LRUCache) removeTagIndexes(key string) {
	for tag, keys := range c.tags {
		filtered := make([]string, 0)
		for _, k := range keys {
			if k != key {
				filtered = append(filtered, k)
			}
		}
		if len(filtered) > 0 {
			c.tags[tag] = filtered
		} else {
			delete(c.tags, tag)
		}
	}
}

// MockClientWithCache is a mock AI client for testing with advanced caching
type MockClientWithCache struct {
	cache *LRUCache
	mu    sync.RWMutex
}

// NewMockClientWithCache creates a new mock client with advanced cache
func NewMockClientWithCache(cacheSize int, ttl time.Duration) Client {
	return &MockClientWithCache{
		cache: NewLRUCache(cacheSize, ttl),
	}
}

// GenerateText generates a deterministic mock response with cache support
func (m *MockClientWithCache) GenerateText(ctx context.Context, prompt string) (string, error) {
	// Check cache
	if entry, ok := m.cache.Get(prompt); ok {
		return entry.Response, nil
	}

	// Generate mock response
	response := "[mock ai] suggestion for: " + prompt

	// Store in cache
	m.cache.Set(prompt, &CacheEntry{
		Prompt:   prompt,
		Response: response,
		Tags:     []string{"mock", "suggestion"},
		Model:    "mock",
	})

	return response, nil
}

// Close is a no-op for the mock client
func (m *MockClientWithCache) Close() error {
	return nil
}
