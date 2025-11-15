package modules

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	adapterhttp "github.com/pmaojo/kthulu-go/backend/internal/adapters/http"
	"github.com/pmaojo/kthulu-go/backend/internal/ai"
	"github.com/pmaojo/kthulu-go/backend/internal/usecase"
)

func TestAIHandler_RegisterRoutes(t *testing.T) {
	// Create a mock AI client
	client, err := ai.NewGeminiClient("mock", 0)
	if err != nil {
		t.Fatalf("failed to create mock client: %v", err)
	}

	// Create the AI handler
	logger := zap.NewNop()
	uc := usecase.NewAIUseCase(client)
	handler := adapterhttp.NewAIHandler(uc, logger)

	// Create a router and register routes
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	// Test the /api/v1/ai/suggest endpoint
	body := `{"prompt": "test prompt", "include_context": false, "project_path": "."}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/suggest", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if _, ok := resp["result"]; !ok {
		t.Fatalf("expected 'result' in response, got %v", resp)
	}
}

func TestRouteRegistry_AIHandler_Registered(t *testing.T) {
	// Create components
	client, _ := ai.NewGeminiClient("mock", 0)
	logger := zap.NewNop()
	registry := NewRouteRegistry(logger)

	uc := usecase.NewAIUseCase(client)
	handler := adapterhttp.NewAIHandler(uc, logger)

	// Register the handler
	registry.Register(handler)

	// Verify the handler was registered
	if len(registry.registrars) != 1 {
		t.Fatalf("expected 1 registrar, got %d", len(registry.registrars))
	}

	// Create a router and register all routes
	r := chi.NewRouter()
	registry.RegisterAllRoutes(r)

	// Test that the route exists
	body := `{"prompt": "test", "include_context": false}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/suggest", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code == http.StatusNotFound {
		t.Fatalf("route /api/v1/ai/suggest not found")
	}
}
