// @kthulu:module:verifactu
package adapterhttp

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"backend/internal/modules/verifactu"
)

// VerifactuHandler exposes VeriFactu HTTP endpoints.
type VerifactuHandler struct {
	service *verifactu.Service
}

// NewVerifactuHandler creates a new handler instance.
func NewVerifactuHandler(service *verifactu.Service) *VerifactuHandler {
	return &VerifactuHandler{service: service}
}

// RegisterRoutes registers VeriFactu routes.
func (h *VerifactuHandler) RegisterRoutes(r chi.Router) {
	r.Route("/verifactu", func(r chi.Router) {
		r.Post("/records/{id}/cancel", h.CancelRecord)
		r.Get("/export", h.ExportRecords)
		r.Get("/config", h.GetConfig)
		r.Post("/config", h.UpdateConfig)
	})
}

// CancelRecord handles record cancellation requests.
func (h *VerifactuHandler) CancelRecord(w http.ResponseWriter, r *http.Request) {
	recordIDStr := chi.URLParam(r, "id")
	recordID, err := strconv.Atoi(recordIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid record ID", err)
		return
	}

	userIDStr := r.Header.Get("X-User-ID")
	userID, _ := strconv.Atoi(userIDStr)

	record, err := h.service.CancelRecord(r.Context(), recordID, userID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to cancel record", err)
		return
	}

	h.writeJSON(w, http.StatusOK, record)
}

// ExportRecords handles export requests.
func (h *VerifactuHandler) ExportRecords(w http.ResponseWriter, r *http.Request) {
	orgIDStr := r.URL.Query().Get("org")
	if orgIDStr == "" {
		h.writeError(w, http.StatusBadRequest, "organization ID required", nil)
		return
	}
	orgID, err := strconv.Atoi(orgIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid organization ID", err)
		return
	}

	data, sig, err := h.service.ExportRecords(r.Context(), orgID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to export records", err)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=verifactu_%d.zip", orgID))
	w.Header().Set("X-Signature", hex.EncodeToString(sig))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// GetConfig returns current VeriFactu configuration values.
func (h *VerifactuHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	cfg, err := h.service.Config(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to load config", err)
		return
	}
	h.writeJSON(w, http.StatusOK, cfg)
}

// UpdateConfig updates the VeriFactu configuration.
func (h *VerifactuHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SIFCode string `json:"sifCode"`
		Mode    string `json:"mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request", err)
		return
	}
	cfg, err := h.service.UpdateConfig(r.Context(), req.SIFCode, req.Mode)
	if err != nil {
		if err == verifactu.ErrModeFrozen {
			h.writeError(w, http.StatusBadRequest, "live mode active", err)
			return
		}
		h.writeError(w, http.StatusInternalServerError, "failed to update config", err)
		return
	}
	h.writeJSON(w, http.StatusOK, cfg)
}

func (h *VerifactuHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *VerifactuHandler) writeError(w http.ResponseWriter, status int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := map[string]interface{}{"error": message, "status": status}
	if err != nil {
		resp["details"] = err.Error()
	}
	json.NewEncoder(w).Encode(resp)
}
