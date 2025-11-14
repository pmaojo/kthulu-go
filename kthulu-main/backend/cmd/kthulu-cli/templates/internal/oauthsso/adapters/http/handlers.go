// @kthulu:module:oauth
package oauthhttp

import (
	"encoding/json"
	"net/http"

	"github.com/ory/fosite"

	"backend/internal/modules/oauthsso/usecase"
)

// Handlers groups OAuth2 HTTP handlers.
type Handlers struct {
	provider fosite.OAuth2Provider
	uc       *usecase.OAuthUseCase
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(provider fosite.OAuth2Provider, uc *usecase.OAuthUseCase) *Handlers {
	return &Handlers{provider: provider, uc: uc}
}

// authorize godoc
// @Summary OAuth2 authorization endpoint
// @Description Handles OAuth2 authorization requests
// @Tags OAuth2
// @Accept  application/x-www-form-urlencoded
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /oauth/authorize [get]
func (h *Handlers) Authorize(w http.ResponseWriter, r *http.Request) {
	code, err := h.uc.Authorize(r.Context(), r, &fosite.DefaultSession{})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"code": code})
}

// token godoc
// @Summary OAuth2 token endpoint
// @Description Exchanges authorization code or refresh token for tokens
// @Tags OAuth2
// @Accept  application/x-www-form-urlencoded
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]string
// @Router /oauth/token [post]
func (h *Handlers) Token(w http.ResponseWriter, r *http.Request) {
	resp, err := h.uc.Token(r.Context(), r, &fosite.DefaultSession{})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(resp)
}

// userinfo godoc
// @Summary User information endpoint
// @Description Returns claims about the authenticated user
// @Tags OAuth2
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 501 {object} map[string]string
// @Router /oauth/userinfo [get]
func (h *Handlers) UserInfo(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "not implemented"})
}

// jwks godoc
// @Summary JWKS endpoint
// @Description Provides the JSON Web Key Set
// @Tags OAuth2
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 501 {object} map[string]string
// @Router /oauth/jwks.json [get]
func (h *Handlers) JWKS(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "not implemented"})
}

// introspect godoc
// @Summary Token introspection endpoint
// @Description Validates an OAuth2 access or refresh token
// @Tags OAuth2
// @Accept  application/x-www-form-urlencoded
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]string
// @Router /oauth/introspect [post]
func (h *Handlers) Introspect(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("token")
	if token == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "missing token"})
		return
	}
	_, _, err := h.uc.Introspect(r.Context(), token, &fosite.DefaultSession{})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"active": true})
}

// revoke godoc
// @Summary Token revocation endpoint
// @Description Revokes an access or refresh token
// @Tags OAuth2
// @Accept  application/x-www-form-urlencoded
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 501 {object} map[string]string
// @Router /oauth/revoke [post]
func (h *Handlers) Revoke(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "not implemented"})
}
