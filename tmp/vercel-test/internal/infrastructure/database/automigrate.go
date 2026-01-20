// @kthulu:database:automigrate
package database

import (
	"fmt"
	"os"

	"gorm.io/gorm"
)

// AutoMigrate runs database migrations if RUN_MIGRATIONS=true.
// This is designed for serverless deployments where migrations run on first request.
func AutoMigrate(db *gorm.DB) error {
	if os.Getenv("RUN_MIGRATIONS") != "true" {
		return nil
	}

	fmt.Println("🔄 Running auto-migrations...")

	// Add your models here for auto-migration
	// Example:
	// err := db.AutoMigrate(
	//     &user.User{},
	//     &task.Task{},
	// )

	fmt.Println("✅ Migrations complete")
	return nil
}
