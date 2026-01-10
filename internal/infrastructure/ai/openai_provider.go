package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenAIProvider implements the AIProvider interface for OpenAI
type OpenAIProvider struct {
	apiKey    string
	model     string
	baseURL   string
	timeout   time.Duration
	client    *http.Client
	cache     *LRUCache
	cacheSize int
}

// OpenAIRequest represents a request to OpenAI API
type OpenAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float32   `json:"temperature,omitempty"`
}

// OpenAIResponse represents a response from OpenAI API
type OpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int     `json:"index"`
		Message      Message `json:"message"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Message represents a message in the conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey, model string, cacheSize int, timeout time.Duration) *OpenAIProvider {
	cache := NewLRUCache(cacheSize, 24*time.Hour)
	return &OpenAIProvider{
		apiKey:    apiKey,
		model:     model,
		baseURL:   "https://api.openai.com/v1",
		timeout:   timeout,
		client:    &http.Client{Timeout: timeout},
		cache:     cache,
		cacheSize: cacheSize,
	}
}

// GenerateText generates text using OpenAI API
func (p *OpenAIProvider) GenerateText(ctx context.Context, prompt string) (string, error) {
	// Check cache first
	if cached, ok := p.cache.Get(prompt); ok {
		return cached.Response, nil
	}

	// Build request
	req := OpenAIRequest{
		Model: p.model,
		Messages: []Message{
			{
				Role:    "system",
				Content: "You are an expert software architect and code generator. Provide clear, concise, and production-ready suggestions.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   2000,
		Temperature: 0.7,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal openai request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create openai request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to execute openai request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read openai response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openai returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result OpenAIResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal openai response: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("openai error: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices in openai response")
	}

	output := result.Choices[0].Message.Content

	// Cache the result
	p.cache.Set(prompt, &CacheEntry{
		Prompt:   prompt,
		Response: output,
		Tags:     []string{"openai", p.model},
		Model:    p.model,
	})

	return output, nil
}

// Close closes the provider
func (p *OpenAIProvider) Close() error {
	p.cache.Clear()
	return nil
}

// Name returns the name of the provider
func (p *OpenAIProvider) Name() string {
	return "openai"
}

// Model returns the model being used
func (p *OpenAIProvider) Model() string {
	return p.model
}

// SetModel sets the model to use
func (p *OpenAIProvider) SetModel(model string) {
	p.model = model
}

// Health checks if the provider is healthy
func (p *OpenAIProvider) Health(ctx context.Context) error {
	if p.apiKey == "" {
		return fmt.Errorf("OpenAI API key not configured")
	}
	return nil
}
