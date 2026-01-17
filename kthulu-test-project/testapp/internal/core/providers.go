package core

import (
    "fmt"
    "log"
    "os"
    "go.uber.org/fx"
    "gorm.io/gorm"
    "gorm.io/driver/sqlite"
    "path/filepath"
    	userCore "testapp/internal/modules/user/core"
    	authCore "testapp/internal/modules/auth/core"
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
    if err := db.AutoMigrate(&userCore.User{}, &authCore.Auth{}); err != nil {
        return nil, fmt.Errorf("auto-migrate failed: %w", err)
    }
    return db, nil
    }
    dbPath := getEnv("SQLITE_PATH", "data/testapp.db")
    if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
        return nil, err
    }
    log.Printf("Using SQLite database at %s", dbPath)
    db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})

    if err != nil { return nil, err }
    
    // Auto-migrate all domain models
    if err := db.AutoMigrate(&userCore.User{}, &authCore.Auth{}); err != nil {
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
