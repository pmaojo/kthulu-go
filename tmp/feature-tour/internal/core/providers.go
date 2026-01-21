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
    	authCore "feature-tour/internal/modules/auth/core"
    	productCore "feature-tour/internal/modules/product/core"
    	contactCore "feature-tour/internal/modules/contact/core"
    	mailCore "feature-tour/internal/modules/mail/core"
    	storageCore "feature-tour/internal/modules/storage/core"
    	eventsCore "feature-tour/internal/modules/events/core"
    	organizationCore "feature-tour/internal/modules/organization/core"
    	invoiceCore "feature-tour/internal/modules/invoice/core"
    	cacheCore "feature-tour/internal/modules/cache/core"
    	schedulerCore "feature-tour/internal/modules/scheduler/core"
    	userCore "feature-tour/internal/modules/user/core"
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
    if err := db.AutoMigrate(&authCore.Auth{}, &productCore.Product{}, &contactCore.Contact{}, &mailCore.Mail{}, &storageCore.Storage{}, &eventsCore.Events{}, &organizationCore.Organization{}, &invoiceCore.Invoice{}, &cacheCore.Cache{}, &schedulerCore.Scheduler{}, &userCore.User{}); err != nil {
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
    if err := db.AutoMigrate(&authCore.Auth{}, &productCore.Product{}, &contactCore.Contact{}, &mailCore.Mail{}, &storageCore.Storage{}, &eventsCore.Events{}, &organizationCore.Organization{}, &invoiceCore.Invoice{}, &cacheCore.Cache{}, &schedulerCore.Scheduler{}, &userCore.User{}); err != nil {
        return nil, fmt.Errorf("auto-migrate failed: %w", err)
    }
    return db, nil
    }
    dbPath := getEnv("SQLITE_PATH", "data/feature-tour.db")
    if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
        return nil, err
    }
    log.Printf("Using SQLite database at %s", dbPath)
    db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})

    if err != nil { return nil, err }
    
    // Auto-migrate all domain models
    if err := db.AutoMigrate(&authCore.Auth{}, &productCore.Product{}, &contactCore.Contact{}, &mailCore.Mail{}, &storageCore.Storage{}, &eventsCore.Events{}, &organizationCore.Organization{}, &invoiceCore.Invoice{}, &cacheCore.Cache{}, &schedulerCore.Scheduler{}, &userCore.User{}); err != nil {
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
