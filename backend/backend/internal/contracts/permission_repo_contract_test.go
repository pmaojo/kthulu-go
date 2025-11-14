package contracts

import (
	"backend/internal/infrastructure/db"
	"backend/internal/repository"
)

// Ensure db.PermissionRepository satisfies repository.PermissionRepository at compile time.
var _ repository.PermissionRepository = (*db.PermissionRepository)(nil)
