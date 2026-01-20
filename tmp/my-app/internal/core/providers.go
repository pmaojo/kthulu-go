package core

import (
    "fmt"
    "log"
    "os"
    "go.uber.org/fx"
    "gorm.io/gorm"
    "gorm.io/driver/sqlite"
    "gorm.io/driver/postgres"
    "path/filepath"
    	organizationCore "my-app/internal/modules/organization/core"
    	authCore "my-app/internal/modules/auth/core"
    	productCore "my-app/internal/modules/product/core"
    	userCore "my-app/internal/modules/user/core"
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
        if err != nil { return nil, err }
        
    // Auto-migrate all domain models
    if err := db.AutoMigrate(&organizationCore.Organization{}, &authCore.Auth{}, &productCore.Product{}, &userCore.User{}); err != nil {
        return nil, fmt.Errorf("auto-migrate failed: %w", err)
    }
    return db, nil
    }

    // Vercel/Production: Check for DATABASE_URL first (Neon, Supabase, etc.)
    if url := os.Getenv("DATABASE_URL"); url != "" {
        log.Println("Connecting to PostgreSQL via DATABASE_URL")
        db, err := gorm.Open(postgres.Open(url), &gorm.Config{})
        if err != nil { return nil, fmt.Errorf("failed to connect to postgres: %w", err) }
        
    // Auto-migrate all domain models
    if err := db.AutoMigrate(&organizationCore.Organization{}, &authCore.Auth{}, &productCore.Product{}, &userCore.User{}); err != nil {
        return nil, fmt.Errorf("auto-migrate failed: %w", err)
    }
    return db, nil
    }
    dbPath := getEnv("SQLITE_PATH", "data/my-app.db")
    if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
        return nil, err
    }
    log.Printf("Using SQLite database at %s", dbPath)
    db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})

    if err != nil { return nil, err }
    
    // Auto-migrate all domain models
    if err := db.AutoMigrate(&organizationCore.Organization{}, &authCore.Auth{}, &productCore.Product{}, &userCore.User{}); err != nil {
        return nil, fmt.Errorf("auto-migrate failed: %w", err)
    }
    return db, nil
}

func getEnv(key, fallback string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return fallback
}
