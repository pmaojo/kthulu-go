package core

import (
	"log"
	"os"
	"path/filepath"

	"go.uber.org/fx"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func CoreRepositoryProviders() fx.Option {
	return fx.Options(
		fx.Provide(NewDatabase),
	)
}

func NewDatabase() (*gorm.DB, error) {
	if os.Getenv("KTHULU_TEST_MODE") == "1" {
		log.Println("Using in-memory SQLite database for tests")
		db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
		if err != nil {
			return nil, err
		}
		return db, nil
	}
	
	dbPath := getEnv("SQLITE_PATH", "data/tournament-app.db")
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, err
	}
	
	log.Printf("Using SQLite database at %s", dbPath)
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
