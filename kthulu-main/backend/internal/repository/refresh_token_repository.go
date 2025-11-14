// @kthulu:module:auth
package repository

import (
	"context"
	"time"

	"backend/internal/domain"
)

// RefreshTokenRepository defines behavior for refresh token persistence.
type RefreshTokenRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, token *domain.RefreshToken) error
	FindByToken(ctx context.Context, token string) (*domain.RefreshToken, error)
	FindByID(ctx context.Context, id uint) (*domain.RefreshToken, error)
	Update(ctx context.Context, token *domain.RefreshToken) error
	Delete(ctx context.Context, id uint) error
	DeleteByToken(ctx context.Context, token string) error

	// User-specific operations
	FindByUserID(ctx context.Context, userID uint) ([]*domain.RefreshToken, error)
	DeleteByUserID(ctx context.Context, userID uint) error
	CountByUserID(ctx context.Context, userID uint) (int64, error)

	// Cleanup operations
	DeleteExpired(ctx context.Context) (int64, error)
	DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error)

	// Query operations
	List(ctx context.Context, limit, offset int) ([]*domain.RefreshToken, error)
	Count(ctx context.Context) (int64, error)
	FindExpired(ctx context.Context) ([]*domain.RefreshToken, error)

	// Existence checks
	ExistsByToken(ctx context.Context, token string) (bool, error)
	ExistsByID(ctx context.Context, id uint) (bool, error)

	// Validation
	IsValidToken(ctx context.Context, token string) (bool, error)
}
