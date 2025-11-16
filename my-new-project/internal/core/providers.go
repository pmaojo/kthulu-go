package core

import (
	"fmt"
	"log"
	"os"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"path/filepath"
	"gorm.io/driver/sqlite"
)

func CoreRepositoryProviders() fx.Option {
        return fx.Options(
                fx.Provide(NewDatabase),
        )
}

func NewDatabase() (*gorm.DB, error) {
if os.Getenv("KTHULU_TEST_MODE") == "1" {
log.Println("Using in-memory SQLite database for tests")
return gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
}
		dbPath := fmt.Sprintf("%s", getEnv("SQLITE_PATH", "data/my-new-project.db"))
		if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
			return nil, err
		}
		log.Printf("Using SQLite database at %s", dbPath)
		return gorm.Open(sqlite.Open(dbPath), &gorm.Config{})

}

func getEnv(key, fallback string) string {
        if value := os.Getenv(key); value != "" {
                return value
        }
        return fallback
}
