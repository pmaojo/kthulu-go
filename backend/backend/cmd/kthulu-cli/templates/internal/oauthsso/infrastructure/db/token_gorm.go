package db

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ory/fosite"
	"gorm.io/gorm"

	"backend/internal/modules/oauthsso/repository"
)

const (
	tokenKindToken = "token"
	tokenKindJTI   = "jti"
)

// TokenModel maps the oauth_tokens table.
type TokenModel struct {
	Signature string `gorm:"primaryKey"`
	Request   string
	Kind      string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// TableName returns the table name.
func (TokenModel) TableName() string { return "oauth_tokens" }

// TokenRepository implements repository.TokenRepository using Gorm.
type TokenRepository struct {
	db *gorm.DB
}

// NewTokenRepository creates a new repository instance.
func NewTokenRepository(db *gorm.DB) repository.TokenRepository {
	return &TokenRepository{db: db}
}

// CreateToken persists the requester using the signature.
func (r *TokenRepository) CreateToken(ctx context.Context, signature string, requester fosite.Requester) error {
	data, err := json.Marshal(requester)
	if err != nil {
		return err
	}
	model := &TokenModel{Signature: signature, Request: string(data), Kind: tokenKindToken}
	return r.db.WithContext(ctx).Create(model).Error
}

// GetToken retrieves the requester associated with the signature.
func (r *TokenRepository) GetToken(ctx context.Context, signature string) (fosite.Requester, error) {
	var model TokenModel
	if err := r.db.WithContext(ctx).Where("signature = ? AND kind = ?", signature, tokenKindToken).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fosite.ErrNotFound
		}
		return nil, err
	}
	var req fosite.Request
	if err := json.Unmarshal([]byte(model.Request), &req); err != nil {
		return nil, err
	}
	return &req, nil
}

// DeleteToken removes the requester associated with the signature.
func (r *TokenRepository) DeleteToken(ctx context.Context, signature string) error {
	return r.db.WithContext(ctx).Where("signature = ? AND kind = ?", signature, tokenKindToken).Delete(&TokenModel{}).Error
}
