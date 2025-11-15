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
	Short: "Gestiona las migraciones de la base de datos",
}

func init() {
	rootCmd.AddCommand(migrateCmd)

	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
	migrateCmd.AddCommand(migrateResetCmd)
	migrateCmd.AddCommand(migrateStatusCmd)
	migrateCmd.AddCommand(migrateVersionCmd)
	migrateCmd.AddCommand(migrateValidateCmd)
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Aplica todas las migraciones pendientes",
	RunE: func(cmd *cobra.Command, args []string) error {
		return withDB(func(db *sql.DB, logger *zap.Logger) error {
			return core.Migrate(db, logger)
		})
	},
}

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Revierte la última migración",
	RunE: func(cmd *cobra.Command, args []string) error {
		return withDB(func(db *sql.DB, logger *zap.Logger) error {
			return core.MigrateDown(db, logger)
		})
	},
}

var migrateResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Resetea la base de datos y reaplica todas las migraciones",
	RunE: func(cmd *cobra.Command, args []string) error {
		return withDB(func(db *sql.DB, logger *zap.Logger) error {
			return core.ResetDatabase(db, logger)
		})
	},
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Muestra la versión actual de la base de datos",
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
	Short: "Migra la base de datos a una versión específica",
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
	Short: "Valida que todas las migraciones sean correctas",
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
