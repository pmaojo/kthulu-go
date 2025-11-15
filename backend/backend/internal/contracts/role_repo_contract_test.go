package contracts

import (
	"github.com/pmaojo/kthulu-go/backend/internal/infrastructure/db"
	"github.com/pmaojo/kthulu-go/backend/internal/repository"
)

// Ensure db.RoleRepository satisfies repository.RoleRepository at compile time.
var _ repository.RoleRepository = (*db.RoleRepository)(nil)
