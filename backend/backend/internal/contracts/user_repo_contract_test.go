package contracts

import (
	"github.com/kthulu/kthulu-go/backend/internal/infrastructure/db"
	"github.com/kthulu/kthulu-go/backend/internal/repository"
)

// Ensure db.UserRepository satisfies repository.UserRepository at compile time.
var _ repository.UserRepository = (*db.UserRepository)(nil)
