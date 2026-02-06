package mcp_test

import (
	"testing"

	"github.com/pmaojo/kthulu-go/internal/adapters/mcp"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestGenerateToolsCode(t *testing.T) {
	root := &cobra.Command{Use: "app"}

	cmd1 := &cobra.Command{
		Use:   "create [name]",
		Short: "Create a resource",
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	root.AddCommand(cmd1)

	cmd2 := &cobra.Command{
		Use:   "deploy [target] [flags...]",
		Short: "Deploy",
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	cmd2.Flags().Bool("force", false, "Force deploy")
	root.AddCommand(cmd2)

	code, err := mcp.GenerateToolsCode(root)
	require.NoError(t, err)

	// Check content
	require.Contains(t, code, "type CreateArgs struct")
	require.Contains(t, code, "Name string")
	require.Contains(t, code, "type DeployArgs struct")
	require.Contains(t, code, "Force")
	require.Contains(t, code, "bool")
	require.Contains(t, code, "NewCreateArgsTool")
	require.Contains(t, code, "NewDeployArgsTool")
}
