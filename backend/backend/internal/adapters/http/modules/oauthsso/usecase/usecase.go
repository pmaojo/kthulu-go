package usecase

import (
	"github.com/pmaojo/kthulu-go/backend/internal/adapters/http/modules/oauthsso/domain"

	"github.com/ory/fosite"
)

// OAuthUseCase provides OAuth SSO operations.
type OAuthUseCase struct {
	cfg     *domain.Config
	storage fosite.Storage
}

// NewOAuthUseCase creates a new OAuthUseCase instance.
func NewOAuthUseCase(cfg *domain.Config, storage fosite.Storage) *OAuthUseCase {
	return &OAuthUseCase{cfg: cfg, storage: storage}
}
