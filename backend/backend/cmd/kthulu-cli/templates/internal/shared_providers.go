// @kthulu:core
package modules

import (
	"go.uber.org/fx"

	"backend/internal/infrastructure/db"
	"backend/internal/infrastructure/notifier"
	"backend/internal/infrastructure/queues"
	"backend/internal/infrastructure/storage"
	"backend/internal/repository"
)

// SharedRepositoryProviders returns fx.Options for all shared repository implementations.
// This centralizes repository provisioning to avoid duplication across modules.
func SharedRepositoryProviders() fx.Option {
	return fx.Options(
		// Core domain repositories
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

		// Organization domain repositories
		fx.Provide(
			fx.Annotate(
				db.NewOrganizationRepository,
				fx.As(new(repository.OrganizationRepository)),
			),
			fx.Annotate(
				db.NewOrganizationUserRepository,
				fx.As(new(repository.OrganizationUserRepository)),
			),
			fx.Annotate(
				db.NewInvitationRepository,
				fx.As(new(repository.InvitationRepository)),
			),
		),

		// Contact domain repositories
		fx.Provide(
			fx.Annotate(
				db.NewContactRepository,
				fx.As(new(repository.ContactRepository)),
			),
		),

		// Product domain repositories
		fx.Provide(
			fx.Annotate(
				db.NewProductRepository,
				fx.As(new(repository.ProductRepository)),
			),
		),

		// Invoice domain repositories
		fx.Provide(
			fx.Annotate(
				db.NewInvoiceRepository,
				fx.As(new(repository.InvoiceRepository)),
			),
		),

		// Inventory domain repositories
		fx.Provide(
			fx.Annotate(
				db.NewInventoryRepository,
				fx.As(new(repository.InventoryRepository)),
			),
		),

		// Calendar domain repositories
		fx.Provide(
			fx.Annotate(
				db.NewCalendarRepository,
				fx.As(new(repository.CalendarRepository)),
			),
		),

		// Notification providers
		fx.Provide(
			fx.Annotated{
				Name:   "smtpProvider",
				Target: notifier.NewSMTPProvider,
			},
		),

		// Storage providers
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

// SharedServiceProviders returns fx.Options for shared service implementations.
// This can be extended for other cross-cutting concerns like caching, metrics, etc.
func SharedServiceProviders() fx.Option {
	return fx.Options(
		queues.Module,
		// Add shared services here as needed
		// fx.Provide(cache.NewRedisCache),
		// fx.Provide(metrics.NewPrometheusMetrics),
	)
}
