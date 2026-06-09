package mcp_test

import (
	"testing"

	"github.com/pmaojo/kthulu-go/internal/adapters/cli/parser"
	"github.com/pmaojo/kthulu-go/internal/adapters/mcp"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestToolFactoryBuildsCommandAndAnalysisTools(t *testing.T) {
	root := &cobra.Command{Use: "kthulu"}
	statusCmd := &cobra.Command{Use: "status", Run: func(cmd *cobra.Command, args []string) {}}
	root.AddCommand(statusCmd)

	executor := &stubExecutor{}
	tagParser := parser.NewTagParser(nil)

	factory := mcp.NewToolFactory(root, executor, tagParser)
	tools := factory.BuildTools("/tmp", nil)

	names := make(map[string]struct{})
	for _, tool := range tools {
		names[tool.Name] = struct{}{}
	}

	require.Contains(t, names, "status")
	require.Contains(t, names, "guide_tagging")
	require.Contains(t, names, "project_overview")

	// Native skills: filesystem, search, AST, database, Go toolchain, watching.
	for _, name := range []string{
		"fs_read", "fs_write", "fs_edit", "fs_list", "fs_move", "fs_delete",
		"code_search", "file_glob",
		"go_outline", "go_find_symbol", "go_symbol_source",
		"db_schema", "db_query",
		"go_test", "go_build", "go_vet",
		"watch_start", "watch_events", "watch_stop", "watch_list",
	} {
		require.Contains(t, names, name)
	}
}
