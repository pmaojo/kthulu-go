// @kthulu:core
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/pmaojo/kthulu-go/backend/core"
)

func main() {
	var (
		action  = flag.String("action", "up", "Migration action: up, down, reset, status, version")
		version = flag.String("version", "", "Target version for migration (optional)")
	)
	flag.Parse()

	// Load configuration
	cfg, err := core.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create logger
	logger, err := core.NewLogger(cfg)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Get zap logger for core functions
	zapLogger := core.GetZapLogger(logger)

	// Connect to database
	db, err := core.NewDB(cfg, zapLogger)
	if err != nil {
		logger.Fatal("Failed to connect to database", "error", err)
	}
	defer core.CloseDB(db, zapLogger)

	// Execute migration action
	switch *action {
	case "up":
		if err := core.Migrate(db, zapLogger); err != nil {
			logger.Fatal("Migration failed", "error", err)
		}
	case "down":
		if err := core.MigrateDown(db, zapLogger); err != nil {
			logger.Fatal("Migration rollback failed", "error", err)
		}
	case "reset":
		if err := core.ResetDatabase(db, zapLogger); err != nil {
			logger.Fatal("Database reset failed", "error", err)
		}
	case "status":
		version, err := core.GetMigrationStatus(db, zapLogger)
		if err != nil {
			logger.Fatal("Failed to get migration status", "error", err)
		}
		fmt.Printf("Current database version: %d\n", version)
	case "version":
		if *version == "" {
			logger.Fatal("Version parameter is required for version action")
		}
		targetVersion, err := strconv.ParseInt(*version, 10, 64)
		if err != nil {
			logger.Fatal("Invalid version number", "version", *version, "error", err)
		}
		if err := core.MigrateToVersion(db, targetVersion, zapLogger); err != nil {
			logger.Fatal("Migration to version failed", "version", targetVersion, "error", err)
		}
	case "validate":
		if err := core.ValidateMigrations(zapLogger); err != nil {
			logger.Fatal("Migration validation failed", "error", err)
		}
		fmt.Println("All migrations are valid")
	default:
		fmt.Printf("Unknown action: %s\n", *action)
		fmt.Println("Available actions: up, down, reset, status, version, validate")
		os.Exit(1)
	}

	logger.Info("Migration command completed successfully")
}
