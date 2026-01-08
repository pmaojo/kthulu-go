package main

import (
	"log"

	"github.com/pmaojo/kthulu-go/backend/core"
	db "github.com/pmaojo/kthulu-go/backend/internal/infrastructure/db"

	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	cfg, err := core.NewConfig()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	// Create logger
	logger, err := core.NewLogger(cfg)
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Sync()
	zapLogger := core.GetZapLogger(logger)

	// Open sql.DB
	sqlDB, err := core.NewDB(cfg, zapLogger)
	if err != nil {
		logger.Fatal("failed to open database", "error", err)
	}
	defer core.CloseDB(sqlDB, zapLogger)

	// Wrap with GORM DB
	gormDB, err := core.NewGormDB(sqlDB, cfg)
	if err != nil {
		logger.Fatal("failed to create gorm db", "error", err)
	}

	// Run migrations using gormigrate
	m := gormigrate.New(gormDB, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "auto-migrate-initial",
			Migrate: func(tx *gorm.DB) error {
				return db.AutoMigrateModels(tx)
			},
			Rollback: func(tx *gorm.DB) error {
				return nil
			},
		},
	})

	if err := m.Migrate(); err != nil {
		logger.Fatal("migration failed", "error", err)
	}

	logger.Info("gorm migrations completed successfully")
}
