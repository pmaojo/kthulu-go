package repository

import (
	"context"
	"fmt"
	"sync"

	"nhooyr.io/websocket"
)

// ConnectionRepository manages active WebSocket connections.
type ConnectionRepository interface {
	// Add stores a new connection and returns its identifier.
	Add(ctx context.Context, conn *websocket.Conn) (string, error)
	// Remove deletes a connection by its identifier.
	Remove(ctx context.Context, id string) error
	// List retrieves all active connections.
	List(ctx context.Context) ([]*websocket.Conn, error)
}

// InMemoryConnectionRepository provides an in-memory implementation.
type InMemoryConnectionRepository struct {
	mu    sync.RWMutex
	conns map[string]*websocket.Conn
	seq   uint64
}

// NewInMemoryConnectionRepository creates a new in-memory repository.
func NewInMemoryConnectionRepository() ConnectionRepository {
	return &InMemoryConnectionRepository{
		conns: make(map[string]*websocket.Conn),
	}
}

// Add stores the connection and returns a generated identifier.
func (r *InMemoryConnectionRepository) Add(ctx context.Context, conn *websocket.Conn) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	id := fmt.Sprintf("conn-%d", r.seq)
	r.conns[id] = conn
	return id, nil
}

// Remove deletes the connection by identifier.
func (r *InMemoryConnectionRepository) Remove(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.conns, id)
	return nil
}

// List returns all active connections.
func (r *InMemoryConnectionRepository) List(ctx context.Context) ([]*websocket.Conn, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	conns := make([]*websocket.Conn, 0, len(r.conns))
	for _, c := range r.conns {
		conns = append(conns, c)
	}
	return conns, nil
}
