package db

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/ory/fosite"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&ClientModel{}, &SessionModel{}, &TokenModel{}))
	return db
}

func TestClientRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := NewClientRepository(db)

	redirectURIs, _ := json.Marshal([]string{"http://localhost/callback"})
	scopes, _ := json.Marshal([]string{"scope"})
	client := &ClientModel{ID: "client1", Secret: "secret", RedirectURIs: string(redirectURIs), Scopes: string(scopes)}
	require.NoError(t, db.Create(client).Error)

	fetched, err := repo.GetClient(context.Background(), "client1")
	require.NoError(t, err)
	require.Equal(t, "client1", fetched.GetID())

	ctx := context.Background()
	require.NoError(t, repo.ClientAssertionJWTValid(ctx, "jti1"))
	require.NoError(t, repo.SetClientAssertionJWT(ctx, "jti1", time.Now().Add(time.Hour)))
	err = repo.ClientAssertionJWTValid(ctx, "jti1")
	require.ErrorIs(t, err, fosite.ErrJTIKnown)
}

func TestSessionRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSessionRepository(db)
	ctx := context.Background()

	req := fosite.NewRequest()
	req.Client = nil
	req.Session = nil
	require.NoError(t, repo.CreateSession(ctx, "sig1", req))

	sessData := &fosite.DefaultSession{Subject: "alice"}
	stored, err := repo.GetSession(ctx, "sig1", sessData)
	require.NoError(t, err)
	require.Equal(t, sessData, stored.GetSession())

	require.NoError(t, repo.DeleteSession(ctx, "sig1"))
	_, err = repo.GetSession(ctx, "sig1", &fosite.DefaultSession{})
	require.ErrorIs(t, err, fosite.ErrNotFound)
}

func TestTokenRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTokenRepository(db)
	ctx := context.Background()

	req := fosite.NewRequest()
	req.Client = nil
	req.Session = nil
	req.ID = "req1"
	require.NoError(t, repo.CreateToken(ctx, "sig1", req))

	stored, err := repo.GetToken(ctx, "sig1")
	require.NoError(t, err)
	require.Equal(t, req.ID, stored.GetID())

	require.NoError(t, repo.DeleteToken(ctx, "sig1"))
	_, err = repo.GetToken(ctx, "sig1")
	require.ErrorIs(t, err, fosite.ErrNotFound)
}
