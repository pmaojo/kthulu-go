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

// LiteLLMConfig holds LiteLLM server configuration
type LiteLLMConfig struct {
	BaseURL string
	Timeout time.Duration
}

// LiteLLMClient wraps LiteLLM API calls
type LiteLLMClient struct {
	baseURL    string
	httpClient *http.Client
	cache      *LRUCache
}

// LiteLLMRequest represents a chat completion request for LiteLLM
type LiteLLMRequest struct {
	Model       string                   `json:"model"`
	Messages    []map[string]interface{} `json:"messages"`
	Temperature float32                  `json:"temperature,omitempty"`
	MaxTokens   int                      `json:"max_tokens,omitempty"`
	TopP        float32                  `json:"top_p,omitempty"`
}

// LiteLLMResponse represents a chat completion response from LiteLLM
type LiteLLMResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
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

// NewLiteLLMClient creates a new LiteLLM client
func NewLiteLLMClient(config LiteLLMConfig, cacheTTL time.Duration) *LiteLLMClient {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.BaseURL == "" {
		config.BaseURL = "http://localhost:4000"
	}

	return &LiteLLMClient{
		baseURL: config.BaseURL,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		cache: NewLRUCache(100, cacheTTL),
	}
}

// GenerateText generates text using LiteLLM with optional caching
func (c *LiteLLMClient) GenerateText(ctx context.Context, prompt string) (string, error) {
	// Check cache first
	if cached, ok := c.cache.Get(prompt); ok {
		return cached.Response, nil
	}

	// Build request
	req := LiteLLMRequest{
		Model: "gpt-4",
		Messages: []map[string]interface{}{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		Temperature: 0.7,
		MaxTokens:   2000,
		TopP:        0.95,
	}

	// Execute request
	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("litellm returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result LiteLLMResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("litellm error: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	output := result.Choices[0].Message.Content

	// Cache the result
	c.cache.Set(prompt, &CacheEntry{
		Prompt:   prompt,
		Response: output,
		Tags:     []string{"litellm", "gpt-4"},
		Model:    "gpt-4",
	})

	return output, nil
}

// Close closes the client
func (c *LiteLLMClient) Close() error {
	c.cache.Clear()
	return nil
}
