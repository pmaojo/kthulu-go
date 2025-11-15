package usecase

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/kthulu/kthulu-go/backend/internal/ai"
)

func TestAIUseCase_Suggest_WithMock(t *testing.T) {
	client, err := ai.NewGeminiClient("mock", 1*time.Minute)
	if err != nil {
		t.Fatalf("failed to create mock ai client: %v", err)
	}

	uc := NewAIUseCase(client)
	res, err := uc.Suggest(context.Background(), "hello world", false, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res, "mocked ai") {
		t.Fatalf("expected mock response, got %q", res)
	}
}
