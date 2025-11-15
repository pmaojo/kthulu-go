package realtime

import (
	"context"
	"encoding/json"

	"nhooyr.io/websocket"

	"github.com/pmaojo/kthulu-go/backend/internal/repository"
)

// DomainEvent represents a generic event to broadcast to clients.
type DomainEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Broadcaster defines behavior for broadcasting events to connected clients.
type Broadcaster interface {
	Broadcast(ctx context.Context, event DomainEvent) error
}

// Service implements the Broadcaster interface.
type Service struct {
	repo repository.ConnectionRepository
}

// NewService creates a new real-time broadcasting service.
func NewService(repo repository.ConnectionRepository) *Service {
	return &Service{repo: repo}
}

// Broadcast sends the event to all active connections.
func (s *Service) Broadcast(ctx context.Context, event DomainEvent) error {
	conns, err := s.repo.List(ctx)
	if err != nil {
		return err
	}
	msg, err := json.Marshal(event)
	if err != nil {
		return err
	}
	for _, c := range conns {
		_ = c.Write(ctx, websocket.MessageText, msg)
	}
	return nil
}
