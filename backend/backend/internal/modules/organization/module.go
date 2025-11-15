// @kthulu:module:organization
package organization

import (
	"go.uber.org/fx"

	"github.com/kthulu/kthulu-go/backend/internal/infrastructure/db"
	"github.com/kthulu/kthulu-go/backend/internal/repository"
)

// Module provides fx.Options for organization (multi-tenancy) module.
// Includes organization management, user relationships, and invitations.
var Module = fx.Options(
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
