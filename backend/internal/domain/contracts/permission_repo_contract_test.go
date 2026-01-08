package contracts

import (
	"github.com/pmaojo/kthulu-go/backend/internal/infrastructure/db"
	"github.com/pmaojo/kthulu-go/backend/internal/domain/repository"
)

// Ensure db.PermissionRepository satisfies repository.PermissionRepository at compile time.
var _ repository.PermissionRepository = (*db.PermissionRepository)(nil)
