package oauthhttp

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/pmaojo/kthulu-go/backend/internal/modules/oauthsso/usecase"
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
	w.WriteHeader(http.StatusNotImplemented)
}
