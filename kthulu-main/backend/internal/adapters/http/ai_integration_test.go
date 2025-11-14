package adapterhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"backend/internal/ai"
	"backend/internal/usecase"
)

// TestAIEndpoints_Integration tests all AI endpoints working together
func TestAIEndpoints_Integration(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockClient := ai.NewMockClientWithCache(100, 1*time.Minute)
	aiUC := usecase.NewAIUseCase(mockClient)
	handler := NewAIHandler(aiUC, logger)

	tests := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		expectedKeys   []string
	}{
		{
			name:           "GET /api/v1/ai/providers",
			method:         "GET",
			path:           "/api/v1/ai/providers",
			expectedStatus: http.StatusOK,
			expectedKeys:   []string{"providers"},
		},
		{
			name:   "POST /api/v1/ai/suggest",
			method: "POST",
			path:   "/api/v1/ai/suggest",
			body: map[string]interface{}{
				"prompt":          "Test prompt",
				"include_context": false,
			},
			expectedStatus: http.StatusOK,
			expectedKeys:   []string{"result", "model", "provider"},
		},
		{
			name:           "GET /api/v1/ai/models",
			method:         "GET",
			path:           "/api/v1/ai/models",
			expectedStatus: http.StatusOK,
			expectedKeys:   []string{"models"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			var err error

			if tt.method == "GET" {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			} else {
				body, _ := json.Marshal(tt.body)
				req = httptest.NewRequest(tt.method, tt.path, bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
			}

			w := httptest.NewRecorder()

			// Route the request
			switch tt.path {
			case "/api/v1/ai/suggest":
				handler.suggest(w, req)
			case "/api/v1/ai/providers":
				handler.getProviders(w, req)
			case "/api/v1/ai/models":
				handler.getModels(w, req)
			}

			assert.Equal(t, tt.expectedStatus, w.Code, "Unexpected status code")

			var result map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &result)
			require.NoError(t, err, "Failed to unmarshal response")

			for _, key := range tt.expectedKeys {
				assert.Contains(t, result, key, "Missing key in response: %s", key)
			}
		})
	}
}

// TestAISuggestionFlow tests the complete suggestion flow
func TestAISuggestionFlow(t *testing.T) {
	logger := zap.NewNop()
	mockClient := ai.NewMockClientWithCache(100, 1*time.Minute)
	aiUC := usecase.NewAIUseCase(mockClient)
	handler := NewAIHandler(aiUC, logger)

	// Create request
	reqBody := map[string]interface{}{
		"prompt":          "Suggest rate limiting middleware",
		"include_context": true,
		"project_path":    ".",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/ai/suggest", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.suggest(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var result suggestResponse
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)

	assert.NotEmpty(t, result.Result)
	assert.Equal(t, "gpt-4", result.Model)
	assert.Equal(t, "litellm", result.Provider)
}

// TestProviderSwitching tests switching between providers
func TestProviderSwitching(t *testing.T) {
	logger := zap.NewNop()
	mockClient := ai.NewMockClientWithCache(100, 1*time.Minute)
	aiUC := usecase.NewAIUseCase(mockClient)
	handler := NewAIHandler(aiUC, logger)

	validProviders := []string{"litellm", "gemini"}

	for _, provider := range validProviders {
		reqBody := map[string]interface{}{
			"provider": provider,
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/ai/provider", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler.setProvider(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var result setProviderResponse
		err := json.Unmarshal(w.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, "ok", result.Status)
	}
}

// TestLiteLLMClientIntegration tests LiteLLM client with mock server
func TestLiteLLMClientIntegration(t *testing.T) {
	// Create a mock LiteLLM server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/chat/completions" && r.Method == "POST" {
			response := ai.LiteLLMResponse{
				ID: "test-id",
				Choices: []struct {
					Message struct {
						Content string `json:"content"`
					} `json:"message"`
				}{
					{
						Message: struct {
							Content string `json:"content"`
						}{
							Content: "Generated response from LiteLLM",
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		} else if r.URL.Path == "/v1/models" && r.Method == "GET" {
			response := map[string]interface{}{
				"data": []map[string]string{
					{"id": "gpt-4"},
					{"id": "gpt-3.5-turbo"},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer mockServer.Close()

	// Test the client
	client := ai.NewLiteLLMClient(ai.LiteLLMConfig{
		BaseURL: mockServer.URL,
		Timeout: 5 * time.Second,
	}, 1*time.Minute)

	ctx := context.Background()

	// Test GenerateText
	result, err := client.GenerateText(ctx, "Test prompt")
	require.NoError(t, err)
	assert.NotEmpty(t, result)

	// Test caching (second call should use cache)
	result2, err := client.GenerateText(ctx, "Test prompt")
	require.NoError(t, err)
	assert.Equal(t, result, result2)
}

// TestCacheIntegration tests the caching behavior across requests
func TestCacheIntegration(t *testing.T) {
	logger := zap.NewNop()
	mockClient := ai.NewMockClientWithCache(100, 1*time.Minute)
	aiUC := usecase.NewAIUseCase(mockClient)
	handler := NewAIHandler(aiUC, logger)

	prompt := "Cache test prompt"
	reqBody := map[string]interface{}{
		"prompt":          prompt,
		"include_context": false,
	}

	// First request
	body1, _ := json.Marshal(reqBody)
	req1 := httptest.NewRequest("POST", "/api/v1/ai/suggest", bytes.NewReader(body1))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	handler.suggest(w1, req1)

	var result1 suggestResponse
	json.Unmarshal(w1.Body.Bytes(), &result1)

	// Second request with same prompt (should use cache)
	body2, _ := json.Marshal(reqBody)
	req2 := httptest.NewRequest("POST", "/api/v1/ai/suggest", bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	handler.suggest(w2, req2)

	var result2 suggestResponse
	json.Unmarshal(w2.Body.Bytes(), &result2)

	// Results should be identical (from cache)
	assert.Equal(t, result1.Result, result2.Result)
}
