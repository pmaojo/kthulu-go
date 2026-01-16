package coder

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// LiteLLMConfig holds configuration for the LiteLLM sidecar
type LiteLLMConfig struct {
	Model   string
	Port    int
	APIBase string
	APIKey  string // Optional: for direct provider access
}

// DefaultLiteLLMConfig returns sensible defaults
func DefaultLiteLLMConfig(model string) LiteLLMConfig {
	return LiteLLMConfig{
		Model:   model,
		Port:    4000,
		APIBase: "http://localhost:4000",
	}
}

// LiteLLMClient handles communication with LiteLLM or direct API
type LiteLLMClient struct {
	config     LiteLLMConfig
	httpClient *http.Client
	cmd        *exec.Cmd
	mu         sync.Mutex
	running    bool
}

// NewLiteLLMClient creates a new client
func NewLiteLLMClient(config LiteLLMConfig) *LiteLLMClient {
	return &LiteLLMClient{
		config: config,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // Long timeout for streaming
		},
	}
}

// ChatMessage represents a message in the conversation
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest is the OpenAI-compatible request format
type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Stream      bool          `json:"stream"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

// StreamDelta represents a streaming chunk
type StreamDelta struct {
	Content string `json:"content,omitempty"`
}

// StreamChoice represents a choice in streaming response
type StreamChoice struct {
	Delta        StreamDelta `json:"delta"`
	Index        int         `json:"index"`
	FinishReason *string     `json:"finish_reason"`
}

// StreamResponse is a single SSE chunk from the API
type StreamResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []StreamChoice `json:"choices"`
}

// StartSidecar starts the LiteLLM proxy as a background process
func (c *LiteLLMClient) StartSidecar(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return nil
	}

	// Check if LiteLLM is already running
	if c.healthCheck() {
		c.running = true
		return nil
	}

	// Check if litellm is installed
	_, err := exec.LookPath("litellm")
	if err != nil {
		return fmt.Errorf("litellm not found. Please install Python and run: pip install litellm")
	}

	// Start LiteLLM proxy
	c.cmd = exec.CommandContext(ctx, "litellm",
		"--model", c.config.Model,
		"--port", fmt.Sprintf("%d", c.config.Port),
	)

	// Redirect output to null (or we could log it)
	c.cmd.Stdout = io.Discard
	c.cmd.Stderr = io.Discard

	if err := c.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start LiteLLM: %w", err)
	}

	// Wait for health check
	for i := 0; i < 30; i++ {
		time.Sleep(500 * time.Millisecond)
		if c.healthCheck() {
			c.running = true
			return nil
		}
	}

	return fmt.Errorf("LiteLLM failed to start within timeout")
}

// StopSidecar gracefully stops the LiteLLM proxy
func (c *LiteLLMClient) StopSidecar() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cmd != nil && c.cmd.Process != nil {
		c.cmd.Process.Signal(os.Interrupt)
		c.cmd.Wait()
	}
	c.running = false
}

// healthCheck tests if the API is responding
func (c *LiteLLMClient) healthCheck() bool {
	resp, err := c.httpClient.Get(c.config.APIBase + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// StreamChat sends a chat request and streams the response
func (c *LiteLLMClient) StreamChat(ctx context.Context, messages []ChatMessage) (<-chan string, <-chan error) {
	textChan := make(chan string, 100)
	errChan := make(chan error, 1)

	go func() {
		defer close(textChan)
		defer close(errChan)

		// Build request
		reqBody := ChatRequest{
			Model:    c.config.Model,
			Messages: messages,
			Stream:   true,
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			errChan <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		// Create HTTP request
		req, err := http.NewRequestWithContext(ctx, "POST",
			c.config.APIBase+"/v1/chat/completions",
			strings.NewReader(string(jsonBody)))
		if err != nil {
			errChan <- fmt.Errorf("failed to create request: %w", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "text/event-stream")

		// Add API key if configured (for direct provider access)
		if c.config.APIKey != "" {
			req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
		}

		// Send request
		resp, err := c.httpClient.Do(req)
		if err != nil {
			errChan <- fmt.Errorf("request failed: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			errChan <- fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
			return
		}

		// Parse SSE stream
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			// Skip empty lines and comments
			if line == "" || strings.HasPrefix(line, ":") {
				continue
			}

			// Parse SSE data
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")

				// Check for stream end
				if data == "[DONE]" {
					return
				}

				// Parse JSON chunk
				var chunk StreamResponse
				if err := json.Unmarshal([]byte(data), &chunk); err != nil {
					continue // Skip malformed chunks
				}

				// Extract content from delta
				for _, choice := range chunk.Choices {
					if choice.Delta.Content != "" {
						select {
						case textChan <- choice.Delta.Content:
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			errChan <- fmt.Errorf("stream read error: %w", err)
		}
	}()

	return textChan, errChan
}

// DirectChat bypasses LiteLLM and calls the provider directly
// Useful when LiteLLM is not available
func (c *LiteLLMClient) DirectChat(ctx context.Context, messages []ChatMessage, provider string) (<-chan string, <-chan error) {
	// Map provider to API base
	apiBase := map[string]string{
		"google":    "https://generativelanguage.googleapis.com/v1beta",
		"openai":    "https://api.openai.com/v1",
		"anthropic": "https://api.anthropic.com/v1",
	}[provider]

	if apiBase == "" {
		errChan := make(chan error, 1)
		errChan <- fmt.Errorf("unsupported provider: %s", provider)
		close(errChan)
		return nil, errChan
	}

	// Override API base for direct access
	originalBase := c.config.APIBase
	c.config.APIBase = apiBase
	defer func() { c.config.APIBase = originalBase }()

	return c.StreamChat(ctx, messages)
}
