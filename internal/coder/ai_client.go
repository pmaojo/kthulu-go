package coder

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// AIConfig holds configuration for the AI provider
type AIConfig struct {
	Model   string
	APIBase string
	APIKey  string
}

// DefaultAIConfig returns defaults prioritizing OpenRouter
func DefaultAIConfig(model string) AIConfig {
	// Check for OpenRouter API key
	if apiKey := os.Getenv("OPENROUTER_API_KEY"); apiKey != "" {
		return AIConfig{
			Model:   model,
			APIBase: "https://openrouter.ai/api/v1",
			APIKey:  apiKey,
		}
	}
	
	// Check for Groq API key
	if apiKey := os.Getenv("GROQ_API_KEY"); apiKey != "" {
		return AIConfig{
			Model:   model,
			APIBase: "https://api.groq.com/openai/v1",
			APIKey:  apiKey,
		}
	}
	
	// Check for generic OpenAI key
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		return AIConfig{
			Model:   model,
			APIBase: "https://api.openai.com/v1",
			APIKey:  apiKey,
		}
	}

	// Fallback/Error state - user will need to configure keys
	return AIConfig{
		Model:   model,
		APIBase: "https://openrouter.ai/api/v1",
	}
}

// AIClient handles communication with AI providers
type AIClient struct {
	config     AIConfig
	httpClient *http.Client
}

// NewAIClient creates a new client
func NewAIClient(config AIConfig) *AIClient {
	return &AIClient{
		config: config,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

// ChatMessage represents a message in the conversation
type ChatMessage struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"` // For tool role messages
}

// ChatRequest is the OpenAI-compatible request format
type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Stream      bool          `json:"stream"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Tools       []Tool        `json:"tools,omitempty"`
	ToolChoice  interface{}   `json:"tool_choice,omitempty"` // "auto", "none", or specific tool
}

// Tool represents a tool definition in the request
type Tool struct {
	Type     string      `json:"type"`
	Function FunctionDef `json:"function"`
}

// FunctionDef represents the function definition inside a tool
type FunctionDef struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Parameters  interface{} `json:"parameters"` // JSON Schema
}

// ToolCall represents a tool call from the assistant
type ToolCall struct {
	Index    *int         `json:"index,omitempty"`
	ID       string       `json:"id,omitempty"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// FunctionCall represents the function details in a tool call
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ChatEvent represents a streaming event
type ChatEvent struct {
	Type         string     // "content", "tool_calls", "done", "error", "finish"
	Content      string
	ToolCalls    []ToolCall
	FinishReason string
	Error        error
}

// StreamChoice represents a choice in streaming response
type StreamChoice struct {
	Delta        StreamDelta `json:"delta"`
	Index        int         `json:"index"`
	FinishReason *string     `json:"finish_reason"`
}

// StreamDelta represents a streaming chunk
type StreamDelta struct {
	Content   string     `json:"content,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// StreamResponse is a single SSE chunk from the API
type StreamResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []StreamChoice `json:"choices"`
}

// StreamChat sends a chat request and streams the response
func (c *AIClient) StreamChat(ctx context.Context, messages []ChatMessage, tools []Tool) <-chan ChatEvent {
	eventChan := make(chan ChatEvent, 100)

	go func() {
		defer close(eventChan)

		// Build request
		reqBody := ChatRequest{
			Model:     c.config.Model,
			Messages:  messages,
			Stream:    true,
			MaxTokens: 1024,
		}

		if len(tools) > 0 {
			reqBody.Tools = tools
			reqBody.ToolChoice = "auto"
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			eventChan <- ChatEvent{Type: "error", Error: fmt.Errorf("failed to marshal request: %w", err)}
			return
		}


		// Create HTTP request
		req, err := http.NewRequestWithContext(ctx, "POST",
			c.config.APIBase+"/chat/completions",
			strings.NewReader(string(jsonBody)))
		if err != nil {
			eventChan <- ChatEvent{Type: "error", Error: fmt.Errorf("failed to create request: %w", err)}
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "text/event-stream")
		
		// Required for OpenRouter to identify the app
		req.Header.Set("HTTP-Referer", "https://github.com/pmaojo/kthulu-go")
		req.Header.Set("X-Title", "Kthulu Coder")

		if c.config.APIKey != "" {
			req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
		}

		// Send request
		resp, err := c.httpClient.Do(req)
		if err != nil {
			eventChan <- ChatEvent{Type: "error", Error: fmt.Errorf("request failed: %w", err)}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			eventChan <- ChatEvent{Type: "error", Error: fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))}
			return
		}


		// Parse SSE stream
		scanner := bufio.NewScanner(resp.Body)
		lineCount := 0
		for scanner.Scan() {
			line := scanner.Text()
			lineCount++


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

				// Extract content from choices
				for _, choice := range chunk.Choices {
					// Handle text content
					if choice.Delta.Content != "" {
						select {
						case eventChan <- ChatEvent{Type: "content", Content: choice.Delta.Content}:
						case <-ctx.Done():
							return
						}
					}

					// Handle tool calls
					if len(choice.Delta.ToolCalls) > 0 {
						select {
						case eventChan <- ChatEvent{Type: "tool_calls", ToolCalls: choice.Delta.ToolCalls}:
						case <-ctx.Done():
							return
						}
					}

					// Handle finish reason
					if choice.FinishReason != nil {
						select {
						case eventChan <- ChatEvent{Type: "finish", FinishReason: *choice.FinishReason}:
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			eventChan <- ChatEvent{Type: "error", Error: fmt.Errorf("stream read error: %w", err)}
		}
	}()

	return eventChan
}
