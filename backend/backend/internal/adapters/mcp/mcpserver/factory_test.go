package mcpserver_test

import (
	"testing"

	"github.com/pmaojo/kthulu-go/backend/internal/adapters/mcp/mcpserver"
	"github.com/pmaojo/kthulu-go/backend/internal/adapters/cli/parser"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestToolFactoryBuildsCommandAndAnalysisTools(t *testing.T) {
	root := &cobra.Command{Use: "kthulu"}
	statusCmd := &cobra.Command{Use: "status", Run: func(cmd *cobra.Command, args []string) {}}
	root.AddCommand(statusCmd)

	executor := &stubExecutor{}
	tagParser := parser.NewTagParser(nil)

	factory := mcpserver.NewToolFactory(root, executor, tagParser)
	tools := factory.BuildTools("/tmp", nil)

	names := make(map[string]struct{})
	for _, tool := range tools {
		names[tool.Name] = struct{}{}
	}

	require.Contains(t, names, "status")
	require.Contains(t, names, "guide_tagging")
	require.Contains(t, names, "project_overview")
}
