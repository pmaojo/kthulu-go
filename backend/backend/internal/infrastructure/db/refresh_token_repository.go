// @kthulu:module:auth
package db

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"backend/internal/domain"
	"backend/internal/repository"

	"gorm.io/gorm"
)

// RefreshTokenModel represents the database model for refresh tokens
type RefreshTokenModel struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null"`
	Token     string    `gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time

	// Association
	User *UserModel `gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for RefreshTokenModel
func (RefreshTokenModel) TableName() string {
	return "refresh_tokens"
}

// ToDomain converts RefreshTokenModel to domain.RefreshToken
func (rt *RefreshTokenModel) ToDomain() (*domain.RefreshToken, error) {
	token := &domain.RefreshToken{
		ID:        rt.ID,
		UserID:    rt.UserID,
		Token:     rt.Token,
		ExpiresAt: rt.ExpiresAt,
		CreatedAt: rt.CreatedAt,
	}

	if rt.User != nil {
		user, err := rt.User.ToDomain()
		if err != nil {
			return nil, err
		}
		token.User = user
	}

	return token, nil
}

// FromDomain converts domain.RefreshToken to RefreshTokenModel
func (rt *RefreshTokenModel) FromDomain(token *domain.RefreshToken) {
	rt.ID = token.ID
	rt.UserID = token.UserID
	rt.Token = token.Token
	rt.ExpiresAt = token.ExpiresAt
	rt.CreatedAt = token.CreatedAt
}

// RefreshTokenRepository provides a database-backed implementation of repository.RefreshTokenRepository.
type RefreshTokenRepository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository creates a new instance bound to a Gorm database.
func NewRefreshTokenRepository(db *gorm.DB) repository.RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Create persists a new refresh token.
func (r *RefreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	model := &RefreshTokenModel{}
	model.FromDomain(token)

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	token.ID = model.ID
	return nil
}

// FindByToken retrieves a refresh token by token value.
func (r *RefreshTokenRepository) FindByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	var model RefreshTokenModel
	err := r.db.WithContext(ctx).Preload("User").Where("token = ?", token).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrTokenNotFound
		}
		return nil, err
	}

	return model.ToDomain()
}

// FindByID retrieves a refresh token by ID.
func (r *RefreshTokenRepository) FindByID(ctx context.Context, id uint) (*domain.RefreshToken, error) {
	var model RefreshTokenModel
	err := r.db.WithContext(ctx).Preload("User").Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrTokenNotFound
		}
		return nil, err
	}

	return model.ToDomain()
}

// Update saves refresh token changes.
func (r *RefreshTokenRepository) Update(ctx context.Context, token *domain.RefreshToken) error {
	model := &RefreshTokenModel{}
	model.FromDomain(token)

	return r.db.WithContext(ctx).Save(model).Error
}

// Delete removes a refresh token by ID.
func (r *RefreshTokenRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&RefreshTokenModel{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrTokenNotFound
	}
	return nil
}

// DeleteByToken removes a refresh token by token value.
func (r *RefreshTokenRepository) DeleteByToken(ctx context.Context, token string) error {
	result := r.db.WithContext(ctx).Where("token = ?", token).Delete(&RefreshTokenModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrTokenNotFound
	}
	return nil
}

// FindByUserID retrieves all refresh tokens for a user.
func (r *RefreshTokenRepository) FindByUserID(ctx context.Context, userID uint) ([]*domain.RefreshToken, error) {
	var models []RefreshTokenModel
	err := r.db.WithContext(ctx).Preload("User").Where("user_id = ?", userID).Find(&models).Error
	if err != nil {
		return nil, err
	}

	tokens := make([]*domain.RefreshToken, len(models))
	for i, model := range models {
		token, err := model.ToDomain()
		if err != nil {
			return nil, err
		}
		tokens[i] = token
	}

	return tokens, nil
}

// DeleteByUserID removes all refresh tokens for a user.
func (r *RefreshTokenRepository) DeleteByUserID(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&RefreshTokenModel{}).Error
}

// CountByUserID returns the number of refresh tokens for a user.
func (r *RefreshTokenRepository) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&RefreshTokenModel{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// DeleteExpired removes all expired refresh tokens.
func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&RefreshTokenModel{})
	return result.RowsAffected, result.Error
}

// DeleteOlderThan removes all refresh tokens older than the specified time.
func (r *RefreshTokenRepository) DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("created_at < ?", cutoff).Delete(&RefreshTokenModel{})
	return result.RowsAffected, result.Error
}

// List retrieves refresh tokens with pagination.
func (r *RefreshTokenRepository) List(ctx context.Context, limit, offset int) ([]*domain.RefreshToken, error) {
	var models []RefreshTokenModel
	err := r.db.WithContext(ctx).Preload("User").Limit(limit).Offset(offset).Find(&models).Error
	if err != nil {
		return nil, err
	}

	tokens := make([]*domain.RefreshToken, len(models))
	for i, model := range models {
		token, err := model.ToDomain()
		if err != nil {
			return nil, err
		}
		tokens[i] = token
	}

	return tokens, nil
}

// Count returns the total number of refresh tokens.
func (r *RefreshTokenRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&RefreshTokenModel{}).Count(&count).Error
	return count, err
}

// FindExpired retrieves all expired refresh tokens.
func (r *RefreshTokenRepository) FindExpired(ctx context.Context) ([]*domain.RefreshToken, error) {
	var models []RefreshTokenModel
	err := r.db.WithContext(ctx).Preload("User").Where("expires_at < ?", time.Now()).Find(&models).Error
	if err != nil {
		return nil, err
	}

	tokens := make([]*domain.RefreshToken, len(models))
	for i, model := range models {
		token, err := model.ToDomain()
		if err != nil {
			return nil, err
		}
		tokens[i] = token
	}

	return tokens, nil
}

// ExistsByToken checks if a refresh token exists with the given token value.
func (r *RefreshTokenRepository) ExistsByToken(ctx context.Context, token string) (bool, error) {
	var count int64
	hashed := hashToken(token)
	err := r.db.WithContext(ctx).Model(&RefreshTokenModel{}).Where("token = ?", hashed).Count(&count).Error
	return count > 0, err
}

// ExistsByID checks if a refresh token exists with the given ID.
func (r *RefreshTokenRepository) ExistsByID(ctx context.Context, id uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&RefreshTokenModel{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// IsValidToken checks if a token exists and is not expired.
func (r *RefreshTokenRepository) IsValidToken(ctx context.Context, token string) (bool, error) {
	var count int64
	hashed := hashToken(token)
	err := r.db.WithContext(ctx).Model(&RefreshTokenModel{}).
		Where("token = ? AND expires_at > ?", hashed, time.Now()).
		Count(&count).Error
	return count > 0, err
}

// hashToken hashes a token string using SHA-256 and returns the hex-encoded hash.
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// Ensure RefreshTokenRepository implements the interface
var _ repository.RefreshTokenRepository = (*RefreshTokenRepository)(nil)
