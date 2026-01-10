// @kthulu:core
package repository

import (
	"context"
	"time"
)

// Storage defines a generic storage interface that can be implemented
// by different storage backends (database, cache, memory, etc.)
type Storage interface {
	// Generic operations
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	// Batch operations
	GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error)
	SetMultiple(ctx context.Context, items map[string]interface{}, ttl time.Duration) error
	DeleteMultiple(ctx context.Context, keys []string) error

	// Pattern operations
	Keys(ctx context.Context, pattern string) ([]string, error)
	DeletePattern(ctx context.Context, pattern string) error

	// Health check
	Ping(ctx context.Context) error
	Close() error
}

// TokenStorage defines specialized storage interface for tokens
type TokenStorage interface {
	// Token-specific operations
	StoreToken(ctx context.Context, tokenID string, userID uint, ttl time.Duration) error
	GetTokenUser(ctx context.Context, tokenID string) (uint, error)
	RevokeToken(ctx context.Context, tokenID string) error
	RevokeAllUserTokens(ctx context.Context, userID uint) error
	IsTokenRevoked(ctx context.Context, tokenID string) (bool, error)

	// Token metadata
	GetTokenExpiry(ctx context.Context, tokenID string) (time.Time, error)
	ExtendToken(ctx context.Context, tokenID string, ttl time.Duration) error

	// Cleanup operations
	CleanupExpiredTokens(ctx context.Context) (int, error)
}

// CacheStorage defines caching-specific operations
type CacheStorage interface {
	Storage

	// Cache-specific operations
	Increment(ctx context.Context, key string, delta int64) (int64, error)
	Decrement(ctx context.Context, key string, delta int64) (int64, error)
	SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error)
	GetTTL(ctx context.Context, key string) (time.Duration, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
}

// PaginationParams defines parameters for paginated queries
type PaginationParams struct {
	Page     int    `json:"page" validate:"min=1"`
	PageSize int    `json:"pageSize" validate:"min=1,max=100"`
	SortBy   string `json:"sortBy,omitempty"`
	SortDir  string `json:"sortDir,omitempty" validate:"omitempty,oneof=asc desc"`
}

// PaginationResult contains paginated results with metadata
type PaginationResult[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	TotalPages int   `json:"totalPages"`
	HasNext    bool  `json:"hasNext"`
	HasPrev    bool  `json:"hasPrev"`
}

// NewPaginationParams creates pagination parameters with defaults
func NewPaginationParams(page, pageSize int, sortBy, sortDir string) PaginationParams {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	if sortDir != "asc" && sortDir != "desc" {
		sortDir = "asc"
	}

	return PaginationParams{
		Page:     page,
		PageSize: pageSize,
		SortBy:   sortBy,
		SortDir:  sortDir,
	}
}

// CalculateOffset calculates the SQL offset for pagination
func (p PaginationParams) CalculateOffset() int {
	return (p.Page - 1) * p.PageSize
}

// NewPaginationResult creates a paginated result
func NewPaginationResult[T any](data []T, total int64, params PaginationParams) PaginationResult[T] {
	totalPages := int((total + int64(params.PageSize) - 1) / int64(params.PageSize))

	return PaginationResult[T]{
		Data:       data,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
		HasNext:    params.Page < totalPages,
		HasPrev:    params.Page > 1,
	}
}
