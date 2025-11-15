package contracts

import (
	"github.com/kthulu/kthulu-go/backend/internal/infrastructure/db"
	"github.com/kthulu/kthulu-go/backend/internal/repository"
)

// Ensure db.RoleRepository satisfies repository.RoleRepository at compile time.
var _ repository.RoleRepository = (*db.RoleRepository)(nil)
