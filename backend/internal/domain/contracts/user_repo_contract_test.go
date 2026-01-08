package contracts

import (
	"github.com/pmaojo/kthulu-go/backend/internal/infrastructure/db"
	"github.com/pmaojo/kthulu-go/backend/internal/domain/repository"
)

// Ensure db.UserRepository satisfies repository.UserRepository at compile time.
var _ repository.UserRepository = (*db.UserRepository)(nil)
