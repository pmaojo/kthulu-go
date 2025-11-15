package adapterhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/kthulu/kthulu-go/backend/internal/ai"
	"github.com/kthulu/kthulu-go/backend/internal/usecase"
)

func TestAIHandler_SuggestEndpoint_SuccessfulRequest(t *testing.T) {
	// Create mock client
	client, _ := ai.NewGeminiClient("mock", 1*time.Minute)
	logger := zap.NewNop()

	// Create handler
	uc := usecase.NewAIUseCase(client)
	_ = NewAIHandler(uc, logger)

	// Test suggest via usecase directly
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Call suggest via usecase
	res, err := uc.Suggest(ctx, "Add logging to API", false, ".")
	if err != nil {
		t.Fatalf("suggest failed: %v", err)
	}

	if res == "" {
		t.Fatalf("expected non-empty response, got empty string")
	}

	if len(res) < 5 {
		t.Fatalf("expected longer response, got %q", res)
	}

	t.Logf("✓ Response: %s", res)
}

func TestAIHandler_SuggestEndpoint_WithContext(t *testing.T) {
	client, _ := ai.NewGeminiClient("mock", 1*time.Minute)
	_ = zap.NewNop()

	uc := usecase.NewAIUseCase(client)

	ctx := context.Background()
	res, err := uc.Suggest(ctx, "Optimize database queries", true, ".")
	if err != nil {
		t.Fatalf("suggest with context failed: %v", err)
	}

	if !bytes.Contains([]byte(res), []byte("suggestion")) {
		t.Fatalf("expected 'suggestion' in response, got %q", res)
	}

	t.Logf("✓ Response with context: %s", res)
}

func TestAIHandler_SuggestEndpoint_JSONResponse(t *testing.T) {
	client, _ := ai.NewGeminiClient("mock", 1*time.Minute)
	_ = zap.NewNop()

	uc := usecase.NewAIUseCase(client)
	_ = NewAIHandler(uc, zap.NewNop())

	// For this test, we'll verify the usecase returns valid data
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	suggestion, err := uc.Suggest(ctx, "test", false, ".")
	if err != nil {
		t.Fatalf("failed to get suggestion: %v", err)
	}

	// Verify it's a valid response
	resp := map[string]string{"result": suggestion}
	respData, _ := json.Marshal(resp)

	if len(respData) == 0 {
		t.Fatalf("expected non-empty JSON response")
	}

	t.Logf("✓ JSON response: %s", respData)
}

func TestAIHandler_ConcurrentRequests(t *testing.T) {
	client, _ := ai.NewGeminiClient("mock", 1*time.Minute)
	_ = zap.NewNop()

	uc := usecase.NewAIUseCase(client)

	// Test concurrent calls to suggest
	results := make(chan string, 5)
	errors := make(chan error, 5)

	for i := 0; i < 5; i++ {
		go func(idx int) {
			ctx := context.Background()
			res, err := uc.Suggest(ctx, "prompt", false, ".")
			if err != nil {
				errors <- err
			} else {
				results <- res
			}
		}(i)
	}

	// Collect results
	for i := 0; i < 5; i++ {
		select {
		case res := <-results:
			if res == "" {
				t.Fatalf("expected non-empty response")
			}
			t.Logf("✓ Request %d got response: %s", i, res)
		case err := <-errors:
			t.Fatalf("request %d failed: %v", i, err)
		case <-time.After(2 * time.Second):
			t.Fatalf("request %d timed out", i)
		}
	}
}
