// @kthulu:core
package core

import (
	"database/sql"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewGormDB creates a GORM database instance from the existing sql.DB connection
func NewGormDB(sqlDB *sql.DB, cfg *Config) (*gorm.DB, error) {
	// Configure GORM logger level based on environment
	var logLevel logger.LogLevel
	if cfg.IsDevelopment() {
		logLevel = logger.Info
	} else {
		logLevel = logger.Error
	}

	// Create GORM config
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	}

	var gormDB *gorm.DB
	var err error

	// Create GORM DB using the existing sql.DB connection based on driver
	switch cfg.Database.Driver {
	case "sqlite":
		gormDB, err = gorm.Open(sqlite.New(sqlite.Config{
			Conn: sqlDB,
		}), gormConfig)
	case "postgres":
		gormDB, err = gorm.Open(postgres.New(postgres.Config{
			Conn: sqlDB,
		}), gormConfig)
	default:
		return nil, fmt.Errorf("unsupported database driver for GORM: %s", cfg.Database.Driver)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create GORM database: %w", err)
	}

	return gormDB, nil
}
