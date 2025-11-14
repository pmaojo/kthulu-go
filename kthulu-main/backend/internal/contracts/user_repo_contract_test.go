package contracts

import (
	"backend/internal/infrastructure/db"
	"backend/internal/repository"
)

// Ensure db.UserRepository satisfies repository.UserRepository at compile time.
var _ repository.UserRepository = (*db.UserRepository)(nil)
