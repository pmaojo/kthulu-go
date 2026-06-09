package mcp_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pmaojo/kthulu-go/internal/adapters/mcp"
	"github.com/stretchr/testify/require"
)

func TestFileSystemServiceWriteReadEdit(t *testing.T) {
	dir := t.TempDir()
	service := mcp.NewFileSystemService()

	_, err := service.WriteFile(dir, mcp.WriteFileArgs{Path: "pkg/example.txt", Content: "hello world\nsecond line\n"})
	require.NoError(t, err)

	content, err := service.ReadFile(dir, mcp.ReadFileArgs{Path: "pkg/example.txt"})
	require.NoError(t, err)
	require.Contains(t, content, "hello world")

	sliced, err := service.ReadFile(dir, mcp.ReadFileArgs{Path: "pkg/example.txt", Offset: 2, Limit: 1})
	require.NoError(t, err)
	require.Contains(t, sliced, "second line")
	require.NotContains(t, sliced, "hello world")

	summary, err := service.EditFile(dir, mcp.EditFileArgs{Path: "pkg/example.txt", OldString: "hello world", NewString: "goodbye world"})
	require.NoError(t, err)
	require.Contains(t, summary, "line 1")

	content, err = service.ReadFile(dir, mcp.ReadFileArgs{Path: "pkg/example.txt"})
	require.NoError(t, err)
	require.Contains(t, content, "goodbye world")
}

func TestFileSystemServiceEditRequiresUniqueMatch(t *testing.T) {
	dir := t.TempDir()
	service := mcp.NewFileSystemService()

	_, err := service.WriteFile(dir, mcp.WriteFileArgs{Path: "dup.txt", Content: "same\nsame\n"})
	require.NoError(t, err)

	_, err = service.EditFile(dir, mcp.EditFileArgs{Path: "dup.txt", OldString: "same", NewString: "diff"})
	require.ErrorContains(t, err, "matches 2 times")

	summary, err := service.EditFile(dir, mcp.EditFileArgs{Path: "dup.txt", OldString: "same", NewString: "diff", ReplaceAll: true})
	require.NoError(t, err)
	require.Contains(t, summary, "Replaced 2 occurrence(s)")
}

func TestFileSystemServiceListMoveDelete(t *testing.T) {
	dir := t.TempDir()
	service := mcp.NewFileSystemService()

	_, err := service.WriteFile(dir, mcp.WriteFileArgs{Path: "a/one.txt", Content: "1"})
	require.NoError(t, err)
	_, err = service.WriteFile(dir, mcp.WriteFileArgs{Path: "b/two.txt", Content: "2"})
	require.NoError(t, err)

	listing, err := service.ListDir(dir, mcp.ListDirArgs{Recursive: true})
	require.NoError(t, err)
	require.Contains(t, listing, filepath.Join("a", "one.txt"))
	require.Contains(t, listing, filepath.Join("b", "two.txt"))

	_, err = service.MoveFile(dir, mcp.MoveFileArgs{Source: "a/one.txt", Destination: "c/renamed.txt"})
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(dir, "c", "renamed.txt"))

	_, err = service.DeleteFile(dir, mcp.DeleteFileArgs{Path: "c/renamed.txt"})
	require.NoError(t, err)
	require.NoFileExists(t, filepath.Join(dir, "c", "renamed.txt"))

	// Non-empty directories require recursive=true.
	_, err = service.DeleteFile(dir, mcp.DeleteFileArgs{Path: "b"})
	require.Error(t, err)
	_, err = service.DeleteFile(dir, mcp.DeleteFileArgs{Path: "b", Recursive: true})
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(dir, "b"))
	require.True(t, os.IsNotExist(err))
}

func TestFileSystemServiceRefusesToDeleteRoot(t *testing.T) {
	dir := t.TempDir()
	service := mcp.NewFileSystemService()

	_, err := service.DeleteFile(dir, mcp.DeleteFileArgs{Path: ".", Recursive: true})
	require.ErrorContains(t, err, "project root")
}
