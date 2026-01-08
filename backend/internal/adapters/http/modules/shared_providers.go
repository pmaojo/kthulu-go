// @kthulu:core
package modules

import (
	"sort"

	"go.uber.org/fx"

	"github.com/pmaojo/kthulu-go/backend/internal/infrastructure/db"
	"github.com/pmaojo/kthulu-go/backend/internal/infrastructure/notifier"
	"github.com/pmaojo/kthulu-go/backend/internal/infrastructure/storage"
	"github.com/pmaojo/kthulu-go/backend/internal/domain/repository"
)

// provider identifiers used to map modules to repository providers.
const (
	providerUserRepo         = "user-repo"
	providerRoleRepo         = "role-repo"
	providerPermissionRepo   = "permission-repo"
	providerRefreshTokenRepo = "refresh-token-repo"
	providerTokenStorage     = "token-storage"
	providerOrganizationRepo = "organization-repo"
	providerContactRepo      = "contact-repo"
	providerProductRepo      = "product-repo"
	providerInvoiceRepo      = "invoice-repo"
	providerInventoryRepo    = "inventory-repo"
	providerCalendarRepo     = "calendar-repo"
	providerNotification     = "notification"
)

// providerFactories maps provider identifiers to their Fx option constructors.
var providerFactories = map[string]func() fx.Option{
	providerUserRepo:         UserRepositoryProviders,
	providerRoleRepo:         RoleRepositoryProviders,
	providerPermissionRepo:   PermissionRepositoryProviders,
	providerRefreshTokenRepo: RefreshTokenRepositoryProviders,
	providerTokenStorage:     TokenStorageProviders,
	providerOrganizationRepo: OrganizationRepositoryProviders,
	providerContactRepo:      ContactRepositoryProviders,
	providerProductRepo:      ProductRepositoryProviders,
	providerInvoiceRepo:      InvoiceRepositoryProviders,
	providerInventoryRepo:    InventoryRepositoryProviders,
	providerCalendarRepo:     CalendarRepositoryProviders,
	providerNotification:     NotificationProviders,
}

// moduleProviderMap declares the repositories required by each builtin module.
var moduleProviderMap = map[string][]string{
	"auth":         {providerUserRepo, providerRoleRepo, providerRefreshTokenRepo, providerTokenStorage, providerNotification},
	"user":         {providerUserRepo, providerRoleRepo},
	"access":       {providerUserRepo, providerRoleRepo, providerPermissionRepo},
	"organization": {providerOrganizationRepo, providerUserRepo, providerNotification},
	"contact":      {providerContactRepo},
	"product":      {providerProductRepo},
	"invoice":      {providerInvoiceRepo},
	"inventory":    {providerInventoryRepo, providerProductRepo, providerUserRepo},
	"calendar":     {providerCalendarRepo, providerUserRepo},
}

// CoreRepositoryProviders returns fx.Options for the core repositories.
// These providers cover authentication and authorization primitives.
// @kthulu:core
func CoreRepositoryProviders() fx.Option {
	return fx.Options(
		UserRepositoryProviders(),
		RoleRepositoryProviders(),
		RefreshTokenRepositoryProviders(),
		TokenStorageProviders(),
	)
}

// SharedRepositoryProviders aggregates all repository providers.
// Deprecated: prefer using the granular provider functions directly.
func SharedRepositoryProviders() fx.Option {
	return fx.Options(
		CoreRepositoryProviders(),
		PermissionRepositoryProviders(),
		OrganizationRepositoryProviders(),
		ContactRepositoryProviders(),
		ProductRepositoryProviders(),
		InvoiceRepositoryProviders(),
		InventoryRepositoryProviders(),
		CalendarRepositoryProviders(),
		NotificationProviders(),
	)
}

// UserRepositoryProviders exposes the user repository implementation.
func UserRepositoryProviders() fx.Option {
	return fx.Options(
		fx.Provide(
			fx.Annotate(
				db.NewUserRepository,
				fx.As(new(repository.UserRepository)),
			),
		),
	)
}

// RoleRepositoryProviders exposes the role repository implementation.
func RoleRepositoryProviders() fx.Option {
	return fx.Options(
		fx.Provide(
			fx.Annotate(
				db.NewRoleRepository,
				fx.As(new(repository.RoleRepository)),
			),
		),
	)
}

// PermissionRepositoryProviders exposes the permission repository implementation.
func PermissionRepositoryProviders() fx.Option {
	return fx.Options(
		fx.Provide(
			fx.Annotate(
				db.NewPermissionRepository,
				fx.As(new(repository.PermissionRepository)),
			),
		),
	)
}

// RefreshTokenRepositoryProviders exposes the refresh token repository implementation.
func RefreshTokenRepositoryProviders() fx.Option {
	return fx.Options(
		fx.Provide(
			fx.Annotate(
				db.NewRefreshTokenRepository,
				fx.As(new(repository.RefreshTokenRepository)),
			),
		),
	)
}

// TokenStorageProviders exposes the token storage implementation.
func TokenStorageProviders() fx.Option {
	return fx.Options(
		fx.Provide(
			fx.Annotate(
				func() repository.TokenStorage {
					return storage.NewMemoryTokenStorage()
				},
				fx.As(new(repository.TokenStorage)),
			),
		),
	)
}

// OrganizationRepositoryProviders exposes organization-related repositories.
func OrganizationRepositoryProviders() fx.Option {
	return fx.Options(
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
	)
}

// ContactRepositoryProviders exposes the contact repository implementation.
func ContactRepositoryProviders() fx.Option {
	return fx.Options(
		fx.Provide(
			fx.Annotate(
				db.NewContactRepository,
				fx.As(new(repository.ContactRepository)),
			),
		),
	)
}

// ProductRepositoryProviders exposes the product repository implementation.
func ProductRepositoryProviders() fx.Option {
	return fx.Options(
		fx.Provide(
			fx.Annotate(
				db.NewProductRepository,
				fx.As(new(repository.ProductRepository)),
			),
		),
	)
}

// InvoiceRepositoryProviders exposes the invoice repository implementation.
func InvoiceRepositoryProviders() fx.Option {
	return fx.Options(
		fx.Provide(
			fx.Annotate(
				db.NewInvoiceRepository,
				fx.As(new(repository.InvoiceRepository)),
			),
		),
	)
}

// InventoryRepositoryProviders exposes the inventory repository implementation.
func InventoryRepositoryProviders() fx.Option {
	return fx.Options(
		fx.Provide(
			fx.Annotate(
				db.NewInventoryRepository,
				fx.As(new(repository.InventoryRepository)),
			),
		),
	)
}

// CalendarRepositoryProviders exposes the calendar repository implementation.
func CalendarRepositoryProviders() fx.Option {
	return fx.Options(
		fx.Provide(
			fx.Annotate(
				db.NewCalendarRepository,
				fx.As(new(repository.CalendarRepository)),
			),
		),
	)
}

// NotificationProviders exposes notification infrastructure implementations.
func NotificationProviders() fx.Option {
	return notifier.NotifierModule
}

// SharedServiceProviders remains as a compatibility shim for legacy modules
// that previously relied on this helper. New modules should provide their own
// service dependencies directly.
func SharedServiceProviders() fx.Option {
	return fx.Options()
}

// collectProviderKeys determines the unique provider keys for the provided modules.
func collectProviderKeys(modules []string) []string {
	seen := make(map[string]struct{})
	for _, module := range modules {
		for _, provider := range moduleProviderMap[module] {
			seen[provider] = struct{}{}
		}
	}

	keys := make([]string, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
