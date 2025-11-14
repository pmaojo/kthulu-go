// @kthulu:core
package storage

import (
	"context"
	"fmt"
	"sync"
	"time"

	"backend/internal/repository"
)

// tokenEntry represents a stored token with metadata
type tokenEntry struct {
	UserID    uint
	ExpiresAt time.Time
	Revoked   bool
}

// MemoryTokenStorage implements TokenStorage using in-memory storage
// This is useful for development, testing, and single-instance deployments
type MemoryTokenStorage struct {
	tokens map[string]*tokenEntry
	mutex  sync.RWMutex
}

// NewMemoryTokenStorage creates a new memory-based token storage
func NewMemoryTokenStorage() repository.TokenStorage {
	storage := &MemoryTokenStorage{
		tokens: make(map[string]*tokenEntry),
	}

	// Start cleanup goroutine
	go storage.cleanupLoop()

	return storage
}

// StoreToken stores a token with user ID and TTL
func (m *MemoryTokenStorage) StoreToken(ctx context.Context, tokenID string, userID uint, ttl time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.tokens[tokenID] = &tokenEntry{
		UserID:    userID,
		ExpiresAt: time.Now().Add(ttl),
		Revoked:   false,
	}

	return nil
}

// GetTokenUser retrieves the user ID associated with a token
func (m *MemoryTokenStorage) GetTokenUser(ctx context.Context, tokenID string) (uint, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	entry, exists := m.tokens[tokenID]
	if !exists {
		return 0, fmt.Errorf("token not found")
	}

	if entry.Revoked {
		return 0, fmt.Errorf("token revoked")
	}

	if time.Now().After(entry.ExpiresAt) {
		return 0, fmt.Errorf("token expired")
	}

	return entry.UserID, nil
}

// RevokeToken marks a token as revoked
func (m *MemoryTokenStorage) RevokeToken(ctx context.Context, tokenID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	entry, exists := m.tokens[tokenID]
	if !exists {
		return nil // Already doesn't exist
	}

	entry.Revoked = true
	return nil
}

// RevokeAllUserTokens revokes all tokens for a specific user
func (m *MemoryTokenStorage) RevokeAllUserTokens(ctx context.Context, userID uint) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, entry := range m.tokens {
		if entry.UserID == userID {
			entry.Revoked = true
		}
	}

	return nil
}

// IsTokenRevoked checks if a token is revoked
func (m *MemoryTokenStorage) IsTokenRevoked(ctx context.Context, tokenID string) (bool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	entry, exists := m.tokens[tokenID]
	if !exists {
		return true, nil // Non-existent tokens are considered revoked
	}

	return entry.Revoked, nil
}

// GetTokenExpiry returns the expiry time of a token
func (m *MemoryTokenStorage) GetTokenExpiry(ctx context.Context, tokenID string) (time.Time, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	entry, exists := m.tokens[tokenID]
	if !exists {
		return time.Time{}, fmt.Errorf("token not found")
	}

	return entry.ExpiresAt, nil
}

// ExtendToken extends the expiry time of a token
func (m *MemoryTokenStorage) ExtendToken(ctx context.Context, tokenID string, ttl time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	entry, exists := m.tokens[tokenID]
	if !exists {
		return fmt.Errorf("token not found")
	}

	entry.ExpiresAt = time.Now().Add(ttl)
	return nil
}

// CleanupExpiredTokens removes expired tokens and returns the count
func (m *MemoryTokenStorage) CleanupExpiredTokens(ctx context.Context) (int, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	count := 0

	for tokenID, entry := range m.tokens {
		if now.After(entry.ExpiresAt) {
			delete(m.tokens, tokenID)
			count++
		}
	}

	return count, nil
}

// cleanupLoop runs periodic cleanup of expired tokens
func (m *MemoryTokenStorage) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		ctx := context.Background()
		if count, err := m.CleanupExpiredTokens(ctx); err == nil && count > 0 {
			// Could log cleanup results if logger was available
			_ = count
		}
	}
}
