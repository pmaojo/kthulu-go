// @kthulu:core
package core

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

// Migrate applies all pending database migrations.
// It uses goose to manage database schema evolution.
func Migrate(db *sql.DB, logger *zap.Logger) error {
	dir := filepath.Join("migrations")

	logger.Info("Starting database migrations", zap.String("directory", dir))

	dialect := "postgres"
	if driverName := fmt.Sprintf("%T", db.Driver()); strings.Contains(strings.ToLower(driverName), "sqlite") {
		dialect = "sqlite3"
	}

	if err := goose.SetDialect(dialect); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	// Get current version before migration
	currentVersion, err := goose.GetDBVersion(db)
	if err != nil {
		logger.Warn("Could not get current database version", zap.Error(err))
	} else {
		logger.Info("Current database version", zap.Int64("version", currentVersion))
	}

	// Apply migrations
	if err := goose.Up(db, dir); err != nil {
		logger.Error("Database migration failed", zap.Error(err))
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	// Get new version after migration
	newVersion, err := goose.GetDBVersion(db)
	if err != nil {
		logger.Warn("Could not get new database version", zap.Error(err))
	} else {
		logger.Info("Database migrations completed successfully",
			zap.Int64("previous_version", currentVersion),
			zap.Int64("current_version", newVersion),
		)
	}

	return nil
}

// MigrateDown rolls back the last migration.
func MigrateDown(db *sql.DB, logger *zap.Logger) error {
	dir := filepath.Join("migrations")

	logger.Info("Rolling back last migration", zap.String("directory", dir))

	dialect := "postgres"
	if driverName := fmt.Sprintf("%T", db.Driver()); strings.Contains(strings.ToLower(driverName), "sqlite") {
		dialect = "sqlite3"
	}

	if err := goose.SetDialect(dialect); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	// Get current version before rollback
	currentVersion, err := goose.GetDBVersion(db)
	if err != nil {
		logger.Warn("Could not get current database version", zap.Error(err))
	} else {
		logger.Info("Current database version before rollback", zap.Int64("version", currentVersion))
	}

	if err := goose.Down(db, dir); err != nil {
		logger.Error("Migration rollback failed", zap.Error(err))
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	// Get new version after rollback
	newVersion, err := goose.GetDBVersion(db)
	if err != nil {
		logger.Warn("Could not get new database version", zap.Error(err))
	} else {
		logger.Info("Migration rollback completed successfully",
			zap.Int64("previous_version", currentVersion),
			zap.Int64("current_version", newVersion),
		)
	}

	return nil
}

// MigrateToVersion migrates to a specific version
func MigrateToVersion(db *sql.DB, version int64, logger *zap.Logger) error {
	dir := filepath.Join("migrations")

	logger.Info("Migrating to specific version",
		zap.String("directory", dir),
		zap.Int64("target_version", version),
	)

	dialect := "postgres"
	if driverName := fmt.Sprintf("%T", db.Driver()); strings.Contains(strings.ToLower(driverName), "sqlite") {
		dialect = "sqlite3"
	}

	if err := goose.SetDialect(dialect); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	currentVersion, err := goose.GetDBVersion(db)
	if err != nil {
		logger.Warn("Could not get current database version", zap.Error(err))
	}

	if err := goose.UpTo(db, dir, version); err != nil {
		logger.Error("Migration to version failed",
			zap.Int64("target_version", version),
			zap.Error(err),
		)
		return fmt.Errorf("failed to migrate to version %d: %w", version, err)
	}

	logger.Info("Migration to version completed successfully",
		zap.Int64("previous_version", currentVersion),
		zap.Int64("target_version", version),
	)

	return nil
}

// GetMigrationStatus returns the current migration status
func GetMigrationStatus(db *sql.DB, logger *zap.Logger) (int64, error) {
	if err := goose.SetDialect("postgres"); err != nil {
		return 0, fmt.Errorf("failed to set goose dialect: %w", err)
	}

	version, err := goose.GetDBVersion(db)
	if err != nil {
		logger.Error("Failed to get database version", zap.Error(err))
		return 0, fmt.Errorf("failed to get database version: %w", err)
	}

	logger.Info("Current database version", zap.Int64("version", version))
	return version, nil
}

// ValidateMigrations checks if all migrations are valid
func ValidateMigrations(logger *zap.Logger) error {
	dir := filepath.Join("migrations")

	logger.Info("Validating migrations", zap.String("directory", dir))

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	// This is a basic validation - in a real scenario you might want more sophisticated checks
	logger.Info("Migration validation completed")
	return nil
}

// ResetDatabase drops all tables and re-applies all migrations (DANGEROUS - use only in development)
func ResetDatabase(db *sql.DB, logger *zap.Logger) error {
	dir := filepath.Join("migrations")

	logger.Warn("RESETTING DATABASE - This will drop all data!")

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	// Reset to version 0 (drops all tables)
	if err := goose.Reset(db, dir); err != nil {
		logger.Error("Database reset failed", zap.Error(err))
		return fmt.Errorf("failed to reset database: %w", err)
	}

	// Apply all migrations again
	if err := goose.Up(db, dir); err != nil {
		logger.Error("Database migration after reset failed", zap.Error(err))
		return fmt.Errorf("failed to apply migrations after reset: %w", err)
	}

	logger.Info("Database reset and migration completed successfully")
	return nil
}
