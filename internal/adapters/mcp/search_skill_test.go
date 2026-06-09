package mcp_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pmaojo/kthulu-go/internal/adapters/mcp"
	"github.com/stretchr/testify/require"
)

func writeTestFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

func TestSearchServiceCodeSearch(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "main.go", "package main\n\nfunc HandleRequest() {}\n")
	writeTestFile(t, dir, "internal/util.go", "package internal\n\nfunc handleResponse() {}\n")
	writeTestFile(t, dir, "notes.md", "handle nothing here\n")

	service := mcp.NewSearchService()

	result, err := service.CodeSearch(dir, mcp.CodeSearchArgs{Pattern: `func Handle\w+`, Include: "*.go"})
	require.NoError(t, err)
	require.Contains(t, result, "main.go:3")
	require.NotContains(t, result, "util.go")
	require.NotContains(t, result, "notes.md")

	// Case-insensitive search widens the net.
	result, err = service.CodeSearch(dir, mcp.CodeSearchArgs{Pattern: `func handle\w+`, Include: "*.go", CaseInsensitive: true})
	require.NoError(t, err)
	require.Contains(t, result, "main.go:3")
	require.Contains(t, result, filepath.Join("internal", "util.go")+":3")
}

func TestSearchServiceCodeSearchSkipsVCSDirs(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, ".git/config", "match-me\n")
	writeTestFile(t, dir, "real.txt", "match-me\n")

	service := mcp.NewSearchService()
	result, err := service.CodeSearch(dir, mcp.CodeSearchArgs{Pattern: "match-me"})
	require.NoError(t, err)
	require.Contains(t, result, "real.txt")
	require.NotContains(t, result, ".git")
}

func TestSearchServiceFileGlob(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "cmd/app/main.go", "package main\n")
	writeTestFile(t, dir, "internal/core/service_test.go", "package core\n")
	writeTestFile(t, dir, "README.md", "# readme\n")

	service := mcp.NewSearchService()

	result, err := service.FileGlob(dir, mcp.FileGlobArgs{Pattern: "**/*.go"})
	require.NoError(t, err)
	require.Contains(t, result, "cmd/app/main.go")
	require.Contains(t, result, "internal/core/service_test.go")
	require.NotContains(t, result, "README.md")

	result, err = service.FileGlob(dir, mcp.FileGlobArgs{Pattern: "**/*_test.go"})
	require.NoError(t, err)
	require.Contains(t, result, "service_test.go")
	require.NotContains(t, result, "main.go")

	result, err = service.FileGlob(dir, mcp.FileGlobArgs{Pattern: "*.{md,txt}"})
	require.NoError(t, err)
	require.Contains(t, result, "README.md")
}
