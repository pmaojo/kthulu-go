package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	mcp_golang "github.com/metoro-io/mcp-golang"
)

const maxBufferedWatchEvents = 1000

// FileWatchService provides real-time file watching backed by fsnotify.
// Because MCP is request/response, events are buffered per watch and
// retrieved by polling watch_events.
type FileWatchService struct {
	mu      sync.Mutex
	watches map[string]*activeWatch
	nextID  int
}

type activeWatch struct {
	id      string
	path    string
	watcher *fsnotify.Watcher

	mu      sync.Mutex
	events  []watchEvent
	dropped int
	started time.Time
}

type watchEvent struct {
	Time time.Time
	Op   string
	Path string
}

// NewFileWatchService creates a new FileWatchService.
func NewFileWatchService() *FileWatchService {
	return &FileWatchService{watches: make(map[string]*activeWatch)}
}

// GetTools returns all watch tools bound to the working directory.
func (s *FileWatchService) GetTools(workingDir string) []RegisteredTool {
	return []RegisteredTool{
		s.startTool(workingDir),
		s.eventsTool(),
		s.stopTool(),
		s.listTool(workingDir),
	}
}

// WatchStartArgs defines arguments for starting a watch.
type WatchStartArgs struct {
	Path      string `json:"path,omitempty" jsonschema:"description=File or directory to watch (default: project root)"`
	Recursive bool   `json:"recursive,omitempty" jsonschema:"description=Watch subdirectories too (directories created later are picked up automatically)"`
}

// WatchEventsArgs defines arguments for polling watch events.
type WatchEventsArgs struct {
	ID   string `json:"id" jsonschema:"description=Watch ID returned by watch_start,required"`
	Keep bool   `json:"keep,omitempty" jsonschema:"description=Keep events in the buffer instead of clearing them after reading"`
}

// WatchStopArgs defines arguments for stopping a watch.
type WatchStopArgs struct {
	ID string `json:"id" jsonschema:"description=Watch ID returned by watch_start,required"`
}

// WatchListArgs defines arguments for listing watches (none required).
type WatchListArgs struct{}

func (s *FileWatchService) startTool(workingDir string) RegisteredTool {
	return RegisteredTool{
		Name:        "watch_start",
		Description: "Start watching a file or directory for changes (create, write, rename, delete). Returns a watch ID; poll watch_events to collect changes.",
		Handler: func(ctx context.Context, args WatchStartArgs) (*mcp_golang.ToolResponse, error) {
			result, err := s.Start(workingDir, args)
			if err != nil {
				return nil, err
			}
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
		},
	}
}

func (s *FileWatchService) eventsTool() RegisteredTool {
	return RegisteredTool{
		Name:        "watch_events",
		Description: "Retrieve file change events collected by a watch since the last poll. Events are cleared after reading unless keep=true.",
		Handler: func(ctx context.Context, args WatchEventsArgs) (*mcp_golang.ToolResponse, error) {
			result, err := s.Events(args)
			if err != nil {
				return nil, err
			}
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
		},
	}
}

func (s *FileWatchService) stopTool() RegisteredTool {
	return RegisteredTool{
		Name:        "watch_stop",
		Description: "Stop a file watch and discard its buffered events.",
		Handler: func(ctx context.Context, args WatchStopArgs) (*mcp_golang.ToolResponse, error) {
			result, err := s.Stop(args.ID)
			if err != nil {
				return nil, err
			}
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
		},
	}
}

func (s *FileWatchService) listTool(workingDir string) RegisteredTool {
	return RegisteredTool{
		Name:        "watch_list",
		Description: "List active file watches with their IDs, paths, and pending event counts.",
		Handler: func(ctx context.Context, _ WatchListArgs) (*mcp_golang.ToolResponse, error) {
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(s.List(workingDir))), nil
		},
	}
}

// Start begins watching a path and returns a summary with the new watch ID.
func (s *FileWatchService) Start(workingDir string, args WatchStartArgs) (string, error) {
	watchPath := args.Path
	if strings.TrimSpace(watchPath) == "" {
		watchPath = "."
	}
	path, err := resolveWorkspacePath(workingDir, watchPath)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("unable to access %s: %w", watchPath, err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return "", fmt.Errorf("failed to create watcher: %w", err)
	}

	dirsWatched := 0
	if info.IsDir() && args.Recursive {
		err = filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if !d.IsDir() {
				return nil
			}
			if shouldSkipDir(d.Name()) {
				return filepath.SkipDir
			}
			if addErr := watcher.Add(p); addErr == nil {
				dirsWatched++
			}
			return nil
		})
		if err != nil {
			watcher.Close()
			return "", fmt.Errorf("failed to register watch directories: %w", err)
		}
	} else {
		if err := watcher.Add(path); err != nil {
			watcher.Close()
			return "", fmt.Errorf("failed to watch %s: %w", watchPath, err)
		}
		dirsWatched = 1
	}

	s.mu.Lock()
	s.nextID++
	id := fmt.Sprintf("watch-%d", s.nextID)
	watch := &activeWatch{
		id:      id,
		path:    path,
		watcher: watcher,
		started: time.Now(),
	}
	s.watches[id] = watch
	s.mu.Unlock()

	go s.consume(watch, args.Recursive)

	return fmt.Sprintf("Started %s on %s (%d director%s watched). Poll watch_events with id=%s to collect changes.",
		id, watchPath, dirsWatched, pluralYIes(dirsWatched), id), nil
}

// consume drains watcher events into the per-watch buffer.
func (s *FileWatchService) consume(watch *activeWatch, recursive bool) {
	for {
		select {
		case event, ok := <-watch.watcher.Events:
			if !ok {
				return
			}

			// Pick up directories created after the watch started.
			if recursive && event.Has(fsnotify.Create) {
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() && !shouldSkipDir(filepath.Base(event.Name)) {
					_ = watch.watcher.Add(event.Name)
				}
			}

			watch.mu.Lock()
			if len(watch.events) >= maxBufferedWatchEvents {
				watch.events = watch.events[1:]
				watch.dropped++
			}
			watch.events = append(watch.events, watchEvent{
				Time: time.Now(),
				Op:   event.Op.String(),
				Path: event.Name,
			})
			watch.mu.Unlock()
		case _, ok := <-watch.watcher.Errors:
			if !ok {
				return
			}
		}
	}
}

// Events returns buffered events for a watch.
func (s *FileWatchService) Events(args WatchEventsArgs) (string, error) {
	watch, err := s.lookup(args.ID)
	if err != nil {
		return "", err
	}

	watch.mu.Lock()
	events := append([]watchEvent(nil), watch.events...)
	dropped := watch.dropped
	if !args.Keep {
		watch.events = nil
		watch.dropped = 0
	}
	watch.mu.Unlock()

	if len(events) == 0 {
		return fmt.Sprintf("No events for %s since last poll", args.ID), nil
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%d event(s) for %s:\n", len(events), args.ID))
	for _, event := range events {
		builder.WriteString(fmt.Sprintf("%s %s %s\n", event.Time.Format(time.RFC3339), event.Op, event.Path))
	}
	if dropped > 0 {
		builder.WriteString(fmt.Sprintf("(%d older event(s) dropped because the buffer overflowed)\n", dropped))
	}
	return truncateOutput(builder.String()), nil
}

// Stop closes a watch and removes it from the registry.
func (s *FileWatchService) Stop(id string) (string, error) {
	s.mu.Lock()
	watch, ok := s.watches[strings.TrimSpace(id)]
	if ok {
		delete(s.watches, watch.id)
	}
	s.mu.Unlock()

	if !ok {
		return "", fmt.Errorf("unknown watch ID %q", id)
	}
	if err := watch.watcher.Close(); err != nil {
		return "", fmt.Errorf("failed to stop %s: %w", watch.id, err)
	}
	return fmt.Sprintf("Stopped %s (%s)", watch.id, watch.path), nil
}

// List describes all active watches.
func (s *FileWatchService) List(workingDir string) string {
	s.mu.Lock()
	watches := make([]*activeWatch, 0, len(s.watches))
	for _, watch := range s.watches {
		watches = append(watches, watch)
	}
	s.mu.Unlock()

	if len(watches) == 0 {
		return "No active watches. Use watch_start to create one."
	}

	sort.Slice(watches, func(i, j int) bool { return watches[i].id < watches[j].id })

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%d active watch(es):\n", len(watches)))
	for _, watch := range watches {
		watch.mu.Lock()
		pending := len(watch.events)
		watch.mu.Unlock()

		display := watch.path
		if rel, err := filepath.Rel(workingDir, watch.path); err == nil && !strings.HasPrefix(rel, "..") {
			display = rel
		}
		builder.WriteString(fmt.Sprintf("• %s: %s (started %s, %d pending event(s))\n",
			watch.id, display, watch.started.Format(time.RFC3339), pending))
	}
	return builder.String()
}

// StopAll closes every active watch; used on server shutdown and in tests.
func (s *FileWatchService) StopAll() {
	s.mu.Lock()
	watches := s.watches
	s.watches = make(map[string]*activeWatch)
	s.mu.Unlock()

	for _, watch := range watches {
		_ = watch.watcher.Close()
	}
}

func (s *FileWatchService) lookup(id string) (*activeWatch, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	watch, ok := s.watches[strings.TrimSpace(id)]
	if !ok {
		return nil, fmt.Errorf("unknown watch ID %q (use watch_list to see active watches)", id)
	}
	return watch, nil
}

func pluralYIes(n int) string {
	if n == 1 {
		return "y"
	}
	return "ies"
}
