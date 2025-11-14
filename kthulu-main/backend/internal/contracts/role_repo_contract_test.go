package contracts

import (
	"backend/internal/infrastructure/db"
	"backend/internal/repository"
)

// Ensure db.RoleRepository satisfies repository.RoleRepository at compile time.
var _ repository.RoleRepository = (*db.RoleRepository)(nil)
