// @kthulu:database:factory
package database

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewDB creates a database connection based on environment.
// Uses DATABASE_URL for Postgres (production/Vercel), falls back to SQLite (local dev).
func NewDB() (*gorm.DB, error) {
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// Production: Use Postgres via DATABASE_URL (Neon, Supabase, etc.)
	if url := os.Getenv("DATABASE_URL"); url != "" {
		db, err := gorm.Open(postgres.Open(url), config)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to postgres: %w", err)
		}
		fmt.Println("📦 Connected to Postgres (DATABASE_URL)")
		return db, nil
	}

	// Local development: Use SQLite
	dbPath := os.Getenv("SQLITE_PATH")
	if dbPath == "" {
		dbPath = "kitchen-sink-test.db"
	}
	
	db, err := gorm.Open(sqlite.Open(dbPath), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to sqlite: %w", err)
	}
	fmt.Printf("📦 Connected to SQLite (%s)\n", dbPath)
	return db, nil
}
