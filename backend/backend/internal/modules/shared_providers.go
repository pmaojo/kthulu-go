// @kthulu:core
package modules

import (
	"go.uber.org/fx"

	"github.com/kthulu/kthulu-go/backend/internal/infrastructure/db"
	"github.com/kthulu/kthulu-go/backend/internal/infrastructure/storage"
	"github.com/kthulu/kthulu-go/backend/internal/repository"
)

// CoreRepositoryProviders returns fx.Options for CORE essential repositories.
// These are ALWAYS included in every Kthulu project - no exceptions.
// @kthulu:core
func CoreRepositoryProviders() fx.Option {
	return fx.Options(
		// Essential authentication and authorization
		fx.Provide(
			fx.Annotate(
				db.NewUserRepository,
				fx.As(new(repository.UserRepository)),
			),
			fx.Annotate(
				db.NewRoleRepository,
				fx.As(new(repository.RoleRepository)),
			),
			fx.Annotate(
				db.NewRefreshTokenRepository,
				fx.As(new(repository.RefreshTokenRepository)),
			),
		),

		// Essential storage for sessions and tokens
		fx.Provide(
			fx.Annotate(
				func() repository.TokenStorage {
					// For now, use memory storage. In production, this could be Redis
					return storage.NewMemoryTokenStorage()
				},
				fx.As(new(repository.TokenStorage)),
			),
		),
	)
}

// SharedRepositoryProviders returns fx.Options for all shared repository implementations.
// DEPRECATED: Use CoreRepositoryProviders() + individual module providers instead.
// This centralizes repository provisioning to avoid duplication across modules.
func SharedRepositoryProviders() fx.Option {
	return CoreRepositoryProviders()
}

// SharedServiceProviders returns fx.Options for shared service implementations.
// DEPRECATED: Use individual module providers instead.
// This can be extended for other cross-cutting concerns like caching, metrics, etc.
func SharedServiceProviders() fx.Option {
	return fx.Options(
	// Empty - moved to individual modules
	// Previously: queues.Module, notifier providers, etc.
	)
}
