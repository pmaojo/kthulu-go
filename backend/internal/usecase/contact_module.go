// @kthulu:module:contacts
package usecase

import (
	"github.com/pmaojo/kthulu-go/backend/internal/domain/repository"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// ContactModule provides the contact use case
var ContactModule = fx.Module("contact",
	fx.Provide(NewContactUseCase),
)

// ContactUseCaseParams defines the dependencies for ContactUseCase
type ContactUseCaseParams struct {
	fx.In
	ContactRepo repository.ContactRepository
	Logger      *zap.Logger
}
