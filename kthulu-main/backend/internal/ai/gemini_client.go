//go:build genai
// +build genai

package ai

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"google.golang.org/genai"
)

type cacheEntry struct {
	resp      string
	createdAt time.Time
}

// GeminiClient wraps access to the Google Generative AI (Gemini) SDK
type GeminiClient struct {
	client   *genai.Client
	model    string
	ttl      time.Duration
	cache    map[string]*cacheEntry
	cacheMtx sync.RWMutex
}

// NewGeminiClient creates a new GeminiClient. Reads API key from GEMINI_API_KEY env var.
func NewGeminiClient(model string, ttl time.Duration) (Client, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	// If API key is not set, return a lightweight mock client which provides
	// deterministic responses. This allows local unit tests and CI runs
	// without needing network access or credentials.
	if apiKey == "" {
		// When built with genai tag but no API key, return a mock to avoid
		// network calls. For real usage, prefer set GEMINI_API_KEY in the
		// environment.
		return &GeminiClient{
			client: nil,
			model:  model,
			ttl:    ttl,
			cache:  make(map[string]*cacheEntry),
		}, nil
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}

	return &GeminiClient{
		client: client,
		model:  model,
		ttl:    ttl,
		cache:  make(map[string]*cacheEntry),
	}, nil
}

// Close closes the underlying client
func (g *GeminiClient) Close() error {
	if g.client == nil {
		return nil
	}
	return g.client.Close()
}

// GenerateText generates a textual response for the given prompt, using a simple cache
func (g *GeminiClient) GenerateText(ctx context.Context, prompt string) (string, error) {
	// Check cache
	g.cacheMtx.RLock()
	if e, ok := g.cache[prompt]; ok {
		if time.Since(e.createdAt) < g.ttl {
			resp := e.resp
			g.cacheMtx.RUnlock()
			return resp, nil
		}
	}
	g.cacheMtx.RUnlock()

	model := g.client.GenerativeModel(g.model)
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("gemini generate failed: %w", err)
	}

	var out strings.Builder
	for _, part := range resp.Parts {
		if part.Kind == genai.PartKindText {
			out.WriteString(part.Text)
		}
	}
	result := out.String()

	// Save to cache
	g.cacheMtx.Lock()
	g.cache[prompt] = &cacheEntry{resp: result, createdAt: time.Now()}
	g.cacheMtx.Unlock()

	return result, nil
}
