// @kthulu:core
package core

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	_ "modernc.org/sqlite"
)

// NewDB initializes a database connection using the provided configuration.
// Supports both SQLite (default, Optimal-style) and PostgreSQL.
// It configures connection pooling, validates the connection, and implements retry logic.
func NewDB(cfg *Config, logger *zap.Logger) (*sql.DB, error) {
	logger.Info("Initializing database connection",
		zap.String("driver", cfg.Database.Driver),
		zap.String("url", maskDatabaseURL(cfg.Database.URL)),
		zap.Int("max_open_conns", cfg.Database.MaxOpenConns),
		zap.Int("max_idle_conns", cfg.Database.MaxIdleConns),
		zap.Duration("conn_max_lifetime", cfg.Database.ConnMaxLifetime),
	)

	var db *sql.DB
	var err error

	switch cfg.Database.Driver {
	case "sqlite":
		db, err = sql.Open("sqlite", cfg.Database.URL)
	case "postgres":
		db, err = sql.Open("pgx", cfg.Database.URL)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Database.Driver)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	// SQLite: limit concurrent connections to 1 for write safety
	if cfg.Database.Driver == "sqlite" {
		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(1)
	} else {
		db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
		db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	}
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	// Test the connection with retry logic
	if err := connectWithRetry(db, logger, 5, 2*time.Second); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to establish database connection after retries: %w", err)
	}

	logger.Info("Database connection established successfully")
	return db, nil
}

// HealthCheck performs a database health check
func HealthCheck(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return db.PingContext(ctx)
}

// HealthCheckDetailed performs a detailed database health check with metrics
func HealthCheckDetailed(db *sql.DB, logger *zap.Logger) error {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := db.PingContext(ctx)
	duration := time.Since(start)

	if err != nil {
		logger.Error("Database health check failed",
			zap.Error(err),
			zap.Duration("duration", duration),
		)
		return err
	}

	// Get connection pool stats
	stats := db.Stats()
	logger.Debug("Database health check successful",
		zap.Duration("ping_duration", duration),
		zap.Int("open_connections", stats.OpenConnections),
		zap.Int("in_use", stats.InUse),
		zap.Int("idle", stats.Idle),
		zap.Int64("wait_count", stats.WaitCount),
		zap.Duration("wait_duration", stats.WaitDuration),
		zap.Int64("max_idle_closed", stats.MaxIdleClosed),
		zap.Int64("max_idle_time_closed", stats.MaxIdleTimeClosed),
		zap.Int64("max_lifetime_closed", stats.MaxLifetimeClosed),
	)

	return nil
}

// connectWithRetry attempts to connect to the database with retry logic
func connectWithRetry(db *sql.DB, logger *zap.Logger, maxRetries int, delay time.Duration) error {
	for i := 0; i < maxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		err := db.PingContext(ctx)
		cancel()

		if err == nil {
			if i > 0 {
				logger.Info("Database connection established after retries", zap.Int("attempts", i+1))
			}
			return nil
		}

		logger.Warn("Database connection attempt failed",
			zap.Int("attempt", i+1),
			zap.Int("max_retries", maxRetries),
			zap.Error(err),
		)

		if i < maxRetries-1 {
			logger.Info("Retrying database connection", zap.Duration("delay", delay))
			time.Sleep(delay)
			delay *= 2 // Exponential backoff
		}
	}

	return fmt.Errorf("failed to connect to database after %d attempts", maxRetries)
}

// maskDatabaseURL masks sensitive information in database URL for logging
func maskDatabaseURL(url string) string {
	// Simple masking - in production you might want more sophisticated masking
	if len(url) > 20 {
		return url[:10] + "***" + url[len(url)-7:]
	}
	return "***"
}

// GetConnectionStats returns current database connection statistics
func GetConnectionStats(db *sql.DB) sql.DBStats {
	return db.Stats()
}

// CloseDB closes the database connection gracefully
func CloseDB(db *sql.DB, logger *zap.Logger) error {
	logger.Info("Closing database connection")

	if err := db.Close(); err != nil {
		logger.Error("Failed to close database connection", zap.Error(err))
		return err
	}

	logger.Info("Database connection closed successfully")
	return nil
}
