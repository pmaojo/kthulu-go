package adapterrealtime

import (
	"net/http"

	"nhooyr.io/websocket"

	"github.com/pmaojo/kthulu-go/backend/internal/repository"
)

// Handler upgrades HTTP connections to WebSocket and tracks clients.
type Handler struct {
	repo repository.ConnectionRepository
}

// NewHandler creates a WebSocket handler.
func NewHandler(repo repository.ConnectionRepository) *Handler {
	return &Handler{repo: repo}
}

// ServeHTTP handles the WebSocket handshake and lifecycle.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	ctx := r.Context()
	id, err := h.repo.Add(ctx, c)
	if err != nil {
		c.Close(websocket.StatusInternalError, "store failed")
		return
	}
	defer func() {
		h.repo.Remove(ctx, id)
		c.Close(websocket.StatusNormalClosure, "closed")
	}()

	for {
		if _, _, err := c.Read(ctx); err != nil {
			break
		}
	}
}
