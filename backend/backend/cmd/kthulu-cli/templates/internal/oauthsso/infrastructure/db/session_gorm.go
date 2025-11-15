package db

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ory/fosite"
	"gorm.io/gorm"

	"github.com/pmaojo/kthulu-go/backend/internal/modules/oauthsso/repository"
)

// SessionModel maps the oauth_sessions table.
type SessionModel struct {
	Signature string `gorm:"primaryKey"`
	Request   string
	CreatedAt time.Time
}

// TableName returns the table name.
func (SessionModel) TableName() string { return "oauth_sessions" }

// SessionRepository implements repository.SessionRepository using Gorm.
type SessionRepository struct {
	db *gorm.DB
}

// NewSessionRepository creates a new repository instance.
func NewSessionRepository(db *gorm.DB) repository.SessionRepository {
	return &SessionRepository{db: db}
}

// CreateSession stores the requester under the signature.
func (r *SessionRepository) CreateSession(ctx context.Context, signature string, requester fosite.Requester) error {
	data, err := json.Marshal(requester)
	if err != nil {
		return err
	}
	model := &SessionModel{Signature: signature, Request: string(data)}
	return r.db.WithContext(ctx).Create(model).Error
}

// GetSession retrieves the session for the given signature.
func (r *SessionRepository) GetSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	var model SessionModel
	if err := r.db.WithContext(ctx).First(&model, "signature = ?", signature).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fosite.ErrNotFound
		}
		return nil, err
	}
	var req fosite.Request
	if err := json.Unmarshal([]byte(model.Request), &req); err != nil {
		return nil, err
	}
	if session != nil {
		req.Session = session
	}
	return &req, nil
}

// DeleteSession removes the session for the given signature.
func (r *SessionRepository) DeleteSession(ctx context.Context, signature string) error {
	return r.db.WithContext(ctx).Delete(&SessionModel{}, "signature = ?", signature).Error
}
