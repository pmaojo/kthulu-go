//go:build !genai
// +build !genai

package ai

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type mockClient struct {
	ttl   time.Duration
	cache map[string]string
	mx    sync.RWMutex
}

// NewGeminiClient returns a mock client when the real SDK is not compiled in.
func NewGeminiClient(model string, ttl time.Duration) (Client, error) {
	return &mockClient{ttl: ttl, cache: make(map[string]string)}, nil
}

func (m *mockClient) Close() error {
	return nil
}

func (m *mockClient) GenerateText(ctx context.Context, prompt string) (string, error) {
	m.mx.RLock()
	if v, ok := m.cache[prompt]; ok {
		m.mx.RUnlock()
		return v, nil
	}
	m.mx.RUnlock()

	// Deterministic mock response
	out := fmt.Sprintf("[mocked ai] suggestion for: %s", prompt)
	m.mx.Lock()
	m.cache[prompt] = out
	m.mx.Unlock()

	return out, nil
}
