package cmd

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/pmaojo/kthulu-go/backend/core"
)

// migrateCmd represents the migrate command group
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Manage database migrations",
}

func init() {
	// migrateCmd is added in root.go

	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
	migrateCmd.AddCommand(migrateResetCmd)
	migrateCmd.AddCommand(migrateStatusCmd)
	migrateCmd.AddCommand(migrateVersionCmd)
	migrateCmd.AddCommand(migrateValidateCmd)
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Apply all pending migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		return withDB(func(db *sql.DB, logger *zap.Logger) error {
			return core.Migrate(db, logger)
		})
	},
}

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Revert the last migration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return withDB(func(db *sql.DB, logger *zap.Logger) error {
			return core.MigrateDown(db, logger)
		})
	},
}

var migrateResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset database and reapply all migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		return withDB(func(db *sql.DB, logger *zap.Logger) error {
			return core.ResetDatabase(db, logger)
		})
	},
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current database version",
	RunE: func(cmd *cobra.Command, args []string) error {
		return withDB(func(db *sql.DB, logger *zap.Logger) error {
			v, err := core.GetMigrationStatus(db, logger)
			if err != nil {
				return err
			}
			fmt.Printf("Current database version: %d\n", v)
			return nil
		})
	},
}

var migrateVersionCmd = &cobra.Command{
	Use:   "version [target]",
	Short: "Migrate database to a specific version",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid version number: %w", err)
		}
		return withDB(func(db *sql.DB, logger *zap.Logger) error {
			return core.MigrateToVersion(db, target, logger)
		})
	},
}

var migrateValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate all migrations are correct",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := core.NewConfig()
		if err != nil {
			return err
		}
		l, err := core.NewLogger(cfg)
		if err != nil {
			return err
		}
		defer l.Sync()
		logger := core.GetZapLogger(l)
		if err := core.ValidateMigrations(logger); err != nil {
			return err
		}
		fmt.Println("All migrations are valid")
		return nil
	},
}

// withDB loads configuration, logger and database, then executes the given function.
func withDB(fn func(db *sql.DB, logger *zap.Logger) error) error {
	cfg, err := core.NewConfig()
	if err != nil {
		return err
	}
	l, err := core.NewLogger(cfg)
	if err != nil {
		return err
	}
	defer l.Sync()
	logger := core.GetZapLogger(l)
	db, err := core.NewDB(cfg, logger)
	if err != nil {
		return err
	}
	defer core.CloseDB(db, logger)
	if err := fn(db, logger); err != nil {
		return err
	}
	l.Info("Migration command completed successfully")
	return nil
}
