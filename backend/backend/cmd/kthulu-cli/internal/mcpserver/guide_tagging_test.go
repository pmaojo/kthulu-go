package mcpserver_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pmaojo/kthulu-go/backend/cmd/kthulu-cli/internal/mcpserver"
	"github.com/pmaojo/kthulu-go/backend/cmd/kthulu-cli/internal/parser"
	"github.com/stretchr/testify/require"
)

func TestGuideTaggingServiceBuildGuideHighlightsUntaggedFiles(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()

	tagged := `package user

// @kthulu:module:user
func Existing() {}
`
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "user.go"), []byte(tagged), 0o644))

	untagged := `package order

type OrderService struct{}

func (OrderService) HandleOrder() {}
`
	require.NoError(t, os.MkdirAll(filepath.Join(projectDir, "order"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "order", "service.go"), []byte(untagged), 0o644))

	service := mcpserver.NewGuideTaggingService(parser.NewTagParser(nil))
	guide, err := service.BuildGuide(projectDir)
	require.NoError(t, err)
	require.Contains(t, guide, "order/service.go")
	require.Contains(t, guide, "@kthulu:module:order")
	require.Contains(t, guide, "OrderService")
}
