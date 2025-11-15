package adapterhttp

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/pmaojo/kthulu-go/backend/internal/usecase"
)

// AIHandler exposes AI endpoints
type AIHandler struct {
	ai  *usecase.AIUseCase
	log *zap.SugaredLogger
}

// NewAIHandler constructor
func NewAIHandler(ai *usecase.AIUseCase, logger *zap.Logger) *AIHandler {
	return &AIHandler{ai: ai, log: logger.Sugar()}
}

// RegisterRoutes attaches AI routes
func (h *AIHandler) RegisterRoutes(r chi.Router) {
	r.Post("/api/v1/ai/suggest", h.suggest)
	r.Get("/api/v1/ai/providers", h.getProviders)
	r.Post("/api/v1/ai/provider", h.setProvider)
	r.Get("/api/v1/ai/models", h.getModels)
}

type suggestRequest struct {
	Prompt         string `json:"prompt"`
	IncludeContext bool   `json:"include_context"`
	ProjectPath    string `json:"project_path,omitempty"`
	Model          string `json:"model,omitempty"`
	Provider       string `json:"provider,omitempty"`
}

type suggestResponse struct {
	Result   string `json:"result"`
	Model    string `json:"model"`
	Provider string `json:"provider"`
}

type providersResponse struct {
	Providers []struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Enabled bool   `json:"enabled"`
	} `json:"providers"`
}

type setProviderRequest struct {
	Provider string `json:"provider"`
}

type setProviderResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type modelsResponse struct {
	Models []string `json:"models"`
}

func (h *AIHandler) suggest(w http.ResponseWriter, r *http.Request) {
	var req suggestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Errorw("Invalid AI suggest request", "error", err)
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	res, err := h.ai.Suggest(ctx, req.Prompt, req.IncludeContext, req.ProjectPath)
	if err != nil {
		h.log.Errorw("AI suggest failed", "error", err)
		http.Error(w, "AI suggest failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suggestResponse{
		Result:   res,
		Model:    "gpt-4",
		Provider: "litellm",
	})
}

func (h *AIHandler) getProviders(w http.ResponseWriter, r *http.Request) {
	providers := providersResponse{
		Providers: []struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Enabled bool   `json:"enabled"`
		}{
			{ID: "litellm", Name: "LiteLLM (Multi-provider)", Enabled: true},
			{ID: "gemini", Name: "Google Gemini", Enabled: true},
			{ID: "openai", Name: "OpenAI", Enabled: false},
			{ID: "anthropic", Name: "Anthropic Claude", Enabled: false},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(providers)
}

func (h *AIHandler) setProvider(w http.ResponseWriter, r *http.Request) {
	var req setProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(setProviderResponse{
			Status: "error",
			Error:  "invalid request",
		})
		return
	}

	// TODO: Implement provider switching logic
	// This would require injecting the AIService and updating it
	validProviders := map[string]bool{
		"litellm":   true,
		"gemini":    true,
		"openai":    false, // TODO: implement
		"anthropic": false, // TODO: implement
	}

	if !validProviders[req.Provider] {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(setProviderResponse{
			Status: "error",
			Error:  "invalid or disabled provider",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(setProviderResponse{Status: "ok"})
}

func (h *AIHandler) getModels(w http.ResponseWriter, r *http.Request) {
	// Return common models available through LiteLLM
	models := modelsResponse{
		Models: []string{
			"gpt-4",
			"gpt-4-turbo",
			"gpt-3.5-turbo",
			"claude-3-opus",
			"claude-3-sonnet",
			"claude-3-haiku",
			"gemini-pro",
			"mistral-7b",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models)
}
