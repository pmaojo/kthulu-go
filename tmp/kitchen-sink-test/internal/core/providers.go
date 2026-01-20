package core

import (
    "fmt"
    "log"
    "os"
    "go.uber.org/fx"
    "gorm.io/gorm"
    "gorm.io/driver/sqlite"
    "path/filepath"
    	authCore "kitchen-sink-test/internal/modules/auth/core"
    	calendarCore "kitchen-sink-test/internal/modules/calendar/core"
    	inventoryCore "kitchen-sink-test/internal/modules/inventory/core"
    	productCore "kitchen-sink-test/internal/modules/product/core"
    	userCore "kitchen-sink-test/internal/modules/user/core"
    	organizationCore "kitchen-sink-test/internal/modules/organization/core"
    	contactCore "kitchen-sink-test/internal/modules/contact/core"
    	invoiceCore "kitchen-sink-test/internal/modules/invoice/core"
    	verifactuCore "kitchen-sink-test/internal/modules/verifactu/core"
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
    if err := db.AutoMigrate(&authCore.Auth{}, &calendarCore.Calendar{}, &inventoryCore.Inventory{}, &productCore.Product{}, &userCore.User{}, &organizationCore.Organization{}, &contactCore.Contact{}, &invoiceCore.Invoice{}, &verifactuCore.Verifactu{}); err != nil {
        return nil, fmt.Errorf("auto-migrate failed: %w", err)
    }
    return db, nil
    }
    dbPath := getEnv("SQLITE_PATH", "data/kitchen-sink-test.db")
    if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
        return nil, err
    }
    log.Printf("Using SQLite database at %s", dbPath)
    db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})

    if err != nil { return nil, err }

    // if .AutoMigrateCall
    // 
    // Auto-migrate all domain models
    if err := db.AutoMigrate(&authCore.Auth{}, &calendarCore.Calendar{}, &inventoryCore.Inventory{}, &productCore.Product{}, &userCore.User{}, &organizationCore.Organization{}, &contactCore.Contact{}, &invoiceCore.Invoice{}, &verifactuCore.Verifactu{}); err != nil {
        return nil, fmt.Errorf("auto-migrate failed: %w", err)
    }
    return db, nil
    // else
    // return db, nil
    // end
    return db, nil
}

func getEnv(key, fallback string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return fallback
}
