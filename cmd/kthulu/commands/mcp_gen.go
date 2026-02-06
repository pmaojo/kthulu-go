package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pmaojo/kthulu-go/internal/adapters/mcp"
	"github.com/spf13/cobra"
)

var mcpGenCmd = &cobra.Command{
	Use:    "mcp-schema-gen",
	Short:  "Generate MCP tool schemas from commands",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Generating MCP tools schema...")

		code, err := mcp.GenerateToolsCode(rootCmd)
		if err != nil {
			return fmt.Errorf("generation failed: %w", err)
		}

		// Determine output path. We assume we are in the repo root or can find it.
		// For now, let's hardcode relative to CWD if possible, or assume execution from root.
		outputPath := "internal/adapters/mcp/generated_tools.gen.go"

		// Ensure directory exists
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return err
		}

		if err := os.WriteFile(outputPath, []byte(code), 0644); err != nil {
			return fmt.Errorf("write failed: %w", err)
		}

		fmt.Printf("✅ Generated %s\n", outputPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(mcpGenCmd)
}
