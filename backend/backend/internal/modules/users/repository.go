package users

import (
	"context"
	"errors"
	"sync"
)

// Repository errors
var (
	ErrUserNotFound = errors.New("user not found")
	ErrEmailExists  = errors.New("email already exists")
)

// UserRepository defines persistence behavior for users.
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uint) (*User, error)
}

// InMemoryUserRepository provides an in-memory implementation of UserRepository.
type InMemoryUserRepository struct {
	mu      sync.RWMutex
	users   map[uint]*User
	byEmail map[string]*User
	nextID  uint
}

// NewInMemoryUserRepository creates a new repository instance.
func NewInMemoryUserRepository() UserRepository {
	return &InMemoryUserRepository{
		users:   make(map[uint]*User),
		byEmail: make(map[string]*User),
		nextID:  1,
	}
}

func (r *InMemoryUserRepository) Create(ctx context.Context, user *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.byEmail[user.Email]; exists {
		return ErrEmailExists
	}
	user.ID = r.nextID
	r.nextID++
	r.users[user.ID] = user
	r.byEmail[user.Email] = user
	return nil
}

func (r *InMemoryUserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if u, ok := r.byEmail[email]; ok {
		return u, nil
	}
	return nil, ErrUserNotFound
}

func (r *InMemoryUserRepository) FindByID(ctx context.Context, id uint) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if u, ok := r.users[id]; ok {
		return u, nil
	}
	return nil, ErrUserNotFound
}
