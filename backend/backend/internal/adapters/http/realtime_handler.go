package adapterhttp

import (
	"github.com/go-chi/chi/v5"

	adapterrealtime "backend/internal/adapters/realtime"
)

// RealtimeHandler registers WebSocket routes.
type RealtimeHandler struct {
	ws *adapterrealtime.Handler
}

// NewRealtimeHandler creates a new RealtimeHandler.
func NewRealtimeHandler(ws *adapterrealtime.Handler) *RealtimeHandler {
	return &RealtimeHandler{ws: ws}
}

// RegisterRoutes registers the realtime routes.
func (h *RealtimeHandler) RegisterRoutes(r chi.Router) {
	r.Handle("/ws", h.ws)
}
