package mcp_test

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/pmaojo/kthulu-go/internal/adapters/mcp"
	"github.com/stretchr/testify/require"
)

func startWatch(t *testing.T, service *mcp.FileWatchService, dir string, args mcp.WatchStartArgs) string {
	t.Helper()
	summary, err := service.Start(dir, args)
	require.NoError(t, err)

	id := regexp.MustCompile(`watch-\d+`).FindString(summary)
	require.NotEmpty(t, id, "watch ID missing from summary: %s", summary)
	return id
}

func TestFileWatchServiceCollectsEvents(t *testing.T) {
	dir := t.TempDir()
	service := mcp.NewFileWatchService()
	defer service.StopAll()

	id := startWatch(t, service, dir, mcp.WatchStartArgs{})

	require.NoError(t, os.WriteFile(filepath.Join(dir, "new.txt"), []byte("hello"), 0o644))

	require.Eventually(t, func() bool {
		events, err := service.Events(mcp.WatchEventsArgs{ID: id, Keep: true})
		return err == nil && regexp.MustCompile(`CREATE|WRITE`).MatchString(events)
	}, 5*time.Second, 50*time.Millisecond, "expected a CREATE/WRITE event for new.txt")

	// Reading without keep clears the buffer.
	_, err := service.Events(mcp.WatchEventsArgs{ID: id})
	require.NoError(t, err)
	events, err := service.Events(mcp.WatchEventsArgs{ID: id})
	require.NoError(t, err)
	require.Contains(t, events, "No events")
}

func TestFileWatchServiceRecursivePicksUpNewDirs(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sub"), 0o755))

	service := mcp.NewFileWatchService()
	defer service.StopAll()

	id := startWatch(t, service, dir, mcp.WatchStartArgs{Recursive: true})

	require.NoError(t, os.WriteFile(filepath.Join(dir, "sub", "inner.txt"), []byte("x"), 0o644))

	require.Eventually(t, func() bool {
		events, err := service.Events(mcp.WatchEventsArgs{ID: id, Keep: true})
		return err == nil && regexp.MustCompile(`inner\.txt`).MatchString(events)
	}, 5*time.Second, 50*time.Millisecond, "expected an event from the nested directory")
}

func TestFileWatchServiceStopAndList(t *testing.T) {
	dir := t.TempDir()
	service := mcp.NewFileWatchService()
	defer service.StopAll()

	id := startWatch(t, service, dir, mcp.WatchStartArgs{})

	listing := service.List(dir)
	require.Contains(t, listing, id)

	summary, err := service.Stop(id)
	require.NoError(t, err)
	require.Contains(t, summary, "Stopped "+id)

	_, err = service.Events(mcp.WatchEventsArgs{ID: id})
	require.ErrorContains(t, err, "unknown watch ID")

	require.Contains(t, service.List(dir), "No active watches")
}
