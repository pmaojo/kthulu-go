package secure

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Handler provides HTTP endpoints for security operations.
type Handler struct{}

// NewHandler creates a new security handler.
func NewHandler() *Handler {
	return &Handler{}
}

// RegisterRoutes registers security routes on the router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/secure/scan", h.scan)
	r.Post("/secure/patch", h.patch)
}

// scan executes the vulnerability scan.
func (h *Handler) scan(w http.ResponseWriter, r *http.Request) {
	vulns, err := Scan(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(vulns)
}

// PatchRequest represents the patch request payload.
type PatchRequest struct {
	Module        string `json:"module"`
	SecureVersion string `json:"secureVersion"`
}

// PatchResponse represents the patch operation result.
type PatchResponse struct {
	Status string `json:"status"`
}

// patch applies a secure version patch to a module.
func (h *Handler) patch(w http.ResponseWriter, r *http.Request) {
	var req PatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Module == "" || req.SecureVersion == "" {
		http.Error(w, "module and secureVersion required", http.StatusBadRequest)
		return
	}
	if err := Patch(req.Module, req.SecureVersion); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(PatchResponse{Status: "patched"})
}
