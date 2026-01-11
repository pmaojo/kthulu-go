package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

var migrateCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new migration file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		return createMigrationFile(name, "")
	},
}

func init() {
	migrateCmd.AddCommand(migrateCreateCmd)
}

func createMigrationFile(name, content string) error {
	// Robust detection of migrations directory
	dir := "migrations"
	
	// Check if we are in the project root
	if _, err := os.Stat("go.mod"); err == nil {
		dir = "migrations"
	} else {
		// Check parent directory
		if _, err := os.Stat("../go.mod"); err == nil {
			dir = "../migrations"
		} else {
			// Check two levels up (e.g. from a module subdirectory)
			if _, err := os.Stat("../../go.mod"); err == nil {
				dir = "../../migrations"
			}
		}
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("%s_%s.sql", timestamp, name)
	path := filepath.Join(dir, filename)

	if content == "" {
		content = "-- +goose Up\n-- SQL in section 'Up' is executed when this migration is applied\n\n-- +goose Down\n-- SQL section 'Down' is executed when this migration is rolled back\n"
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create migration file: %w", err)
	}

	fmt.Printf("Created migration file: %s\n", path)
	return nil
}
