package ai

import (
	"context"
	"testing"
	"time"
)

func TestLRUCache_Set_and_Get(t *testing.T) {
	cache := NewLRUCache(2, 1*time.Minute)

	entry := &CacheEntry{
		Prompt:   "test",
		Response: "response",
		Tags:     []string{"test"},
		Model:    "mock",
	}

	cache.Set("test", entry)

	if retrieved, ok := cache.Get("test"); !ok {
		t.Fatalf("expected entry to be cached")
	} else if retrieved.Response != "response" {
		t.Fatalf("expected 'response', got %s", retrieved.Response)
	}
}

func TestLRUCache_Expiry(t *testing.T) {
	cache := NewLRUCache(10, 100*time.Millisecond)

	entry := &CacheEntry{
		Prompt:   "test",
		Response: "response",
		Model:    "mock",
	}

	cache.Set("test", entry)

	// Should be in cache initially
	if _, ok := cache.Get("test"); !ok {
		t.Fatalf("expected entry to be cached")
	}

	// Wait for expiry
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	if _, ok := cache.Get("test"); ok {
		t.Fatalf("expected entry to be expired")
	}
}

func TestLRUCache_GetByTag(t *testing.T) {
	cache := NewLRUCache(10, 1*time.Minute)

	entries := []struct {
		key  string
		tags []string
	}{
		{"key1", []string{"feature", "ai"}},
		{"key2", []string{"feature", "auth"}},
		{"key3", []string{"ai"}},
	}

	for _, e := range entries {
		cache.Set(e.key, &CacheEntry{
			Prompt:   e.key,
			Response: "resp",
			Tags:     e.tags,
			Model:    "mock",
		})
	}

	results := cache.GetByTag("ai")
	if len(results) != 2 {
		t.Fatalf("expected 2 entries with tag 'ai', got %d", len(results))
	}
}

func TestLRUCache_Eviction(t *testing.T) {
	cache := NewLRUCache(2, 1*time.Minute)

	cache.Set("key1", &CacheEntry{Prompt: "p1", Response: "r1", Model: "mock"})
	cache.Set("key2", &CacheEntry{Prompt: "p2", Response: "r2", Model: "mock"})
	cache.Set("key3", &CacheEntry{Prompt: "p3", Response: "r3", Model: "mock"})

	// key1 should be evicted (oldest)
	if _, ok := cache.Get("key1"); ok {
		t.Fatalf("expected key1 to be evicted")
	}

	// key2 and key3 should remain
	if _, ok := cache.Get("key2"); !ok {
		t.Fatalf("expected key2 to be cached")
	}
	if _, ok := cache.Get("key3"); !ok {
		t.Fatalf("expected key3 to be cached")
	}
}

func TestMockClientWithCache_GenerateText(t *testing.T) {
	client := NewMockClientWithCache(10, 1*time.Minute)

	res1, err := client.GenerateText(context.Background(), "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Same prompt should return cached result
	res2, err := client.GenerateText(context.Background(), "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res1 != res2 {
		t.Fatalf("expected same result from cache, got %s vs %s", res1, res2)
	}

	if err := client.Close(); err != nil {
		t.Fatalf("unexpected error on close: %v", err)
	}
}
