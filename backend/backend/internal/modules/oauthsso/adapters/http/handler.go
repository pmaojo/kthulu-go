package oauthhttp

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"backend/internal/modules/auth"
	"backend/internal/modules/oauthsso/usecase"
)

// Handler exposes OAuth SSO endpoints.
type Handler struct {
	uc *usecase.OAuthUseCase
}

// NewHandler creates a new Handler instance.
func NewHandler(uc *usecase.OAuthUseCase) *Handler {
	return &Handler{uc: uc}
}

// RegisterRoutes registers OAuth routes under /oauth prefix.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/oauth", func(r chi.Router) {
		r.Get("/sso", h.handleSSO)
	})
}

func (h *Handler) handleSSO(w http.ResponseWriter, r *http.Request) {
	tokens, err := auth.SSOLogin(r.Context(), h.uc, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(tokens)
}
