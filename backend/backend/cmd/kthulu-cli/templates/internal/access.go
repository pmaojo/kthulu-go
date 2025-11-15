// @kthulu:module:access
package modules

import (
	"go.uber.org/fx"

	"github.com/kthulu/kthulu-go/backend/internal/infrastructure/db"
	"github.com/kthulu/kthulu-go/backend/internal/repository"
	"github.com/kthulu/kthulu-go/backend/internal/usecase"
)

// AccessModule provides access control functionality
var AccessModule = fx.Options(
	// Repositories (PermissionRepository is specific to this module)
	fx.Provide(
		fx.Annotate(
			db.NewPermissionRepository,
			fx.As(new(repository.PermissionRepository)),
		),
	),

	// Use cases
	fx.Provide(
		usecase.NewAccessUseCase,
	),
)
