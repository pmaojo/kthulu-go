package db

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ory/fosite"
	"gorm.io/gorm"

	"github.com/pmaojo/kthulu-go/backend/internal/adapters/http/modules/oauthsso/repository"
)

// ClientModel maps the oauth_clients table.
type ClientModel struct {
	ID           string `gorm:"primaryKey"`
	Secret       string `gorm:"not null"`
	RedirectURIs string `gorm:"not null"`
	Scopes       string `gorm:"not null"`
	CreatedAt    time.Time
}

// TableName returns the table name.
func (ClientModel) TableName() string { return "oauth_clients" }

// ClientRepository implements repository.ClientRepository using Gorm.
type ClientRepository struct {
	db *gorm.DB
}

// NewClientRepository creates a new repository instance.
func NewClientRepository(db *gorm.DB) repository.ClientRepository {
	return &ClientRepository{db: db}
}

// GetClient retrieves a client by its ID.
func (r *ClientRepository) GetClient(ctx context.Context, id string) (fosite.Client, error) {
	var model ClientModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fosite.ErrNotFound
		}
		return nil, err
	}

	var redirectURIs []string
	_ = json.Unmarshal([]byte(model.RedirectURIs), &redirectURIs)
	var scopes []string
	_ = json.Unmarshal([]byte(model.Scopes), &scopes)

	client := &fosite.DefaultClient{
		ID:           model.ID,
		Secret:       []byte(model.Secret),
		RedirectURIs: redirectURIs,
		Scopes:       scopes,
	}
	return client, nil
}

// ClientAssertionJWTValid checks if a JTI has already been used.
func (r *ClientRepository) ClientAssertionJWTValid(ctx context.Context, jti string) error {
	var t TokenModel
	err := r.db.WithContext(ctx).
		Where("signature = ? AND kind = ?", jti, tokenKindJTI).
		First(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if t.ExpiresAt.IsZero() || t.ExpiresAt.After(time.Now()) {
		return fosite.ErrJTIKnown
	}
	r.db.WithContext(ctx).Where("signature = ? AND kind = ?", jti, tokenKindJTI).Delete(&TokenModel{})
	return nil
}

// SetClientAssertionJWT stores a JTI with its expiry time.
func (r *ClientRepository) SetClientAssertionJWT(ctx context.Context, jti string, exp time.Time) error {
	model := &TokenModel{Signature: jti, Kind: tokenKindJTI, ExpiresAt: exp}
	return r.db.WithContext(ctx).Create(model).Error
}
