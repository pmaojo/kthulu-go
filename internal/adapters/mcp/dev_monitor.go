package mcp

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	mcp_golang "github.com/metoro-io/mcp-golang"
)

// DevMonitorService exposes application runtime data as MCP resources.
type DevMonitorService struct {
	httpLogs *LogBuffer
	dbLogs   *LogBuffer
	testLogs *LogBuffer
	mu       sync.RWMutex
}

// NewDevMonitorService creates a monitor service with default buffer sizes.
func NewDevMonitorService() *DevMonitorService {
	return &DevMonitorService{
		httpLogs: NewLogBuffer(100),
		dbLogs:   NewLogBuffer(100),
		testLogs: NewLogBuffer(50),
	}
}

// RegisterResources adds the monitoring resources to the MCP server.
func (s *DevMonitorService) RegisterResources(server *mcp_golang.Server) error {
	resources := []struct {
		uri         string
		name        string
		description string
		getter      func() []LogEntry
	}{
		{
			uri:         "kthulu://monitor/http",
			name:        "HTTP Request Logs",
			description: "Recent HTTP requests with method, path, status, and timing",
			getter:      func() []LogEntry { return s.httpLogs.Last(50) },
		},
		{
			uri:         "kthulu://monitor/db",
			name:        "Database Query Logs",
			description: "Recent database queries with SQL, rows affected, and timing",
			getter:      func() []LogEntry { return s.dbLogs.Last(50) },
		},
		{
			uri:         "kthulu://monitor/tests",
			name:        "Test Results",
			description: "Latest test run results with pass/fail status",
			getter:      func() []LogEntry { return s.testLogs.All() },
		},
	}

	for _, res := range resources {
		uri := res.uri // capture for closure
		getter := res.getter // capture for closure
		mimeType := "application/json"
		handler := func(ctx context.Context) (*mcp_golang.ResourceResponse, error) {
			entries := getter()
			data, err := json.MarshalIndent(entries, "", "  ")
			if err != nil {
				return nil, err
			}
			return mcp_golang.NewResourceResponse(
				mcp_golang.NewTextEmbeddedResource(uri, string(data), mimeType),
			), nil
		}

		if err := server.RegisterResource(res.uri, res.name, res.description, mimeType, handler); err != nil {
			return err
		}
	}

	return nil
}

// PushHTTP adds an HTTP log entry.
func (s *DevMonitorService) PushHTTP(method, path, status, duration string) {
	s.httpLogs.Push(LogEntry{
		Timestamp: time.Now(),
		Type:      "http",
		Data: map[string]string{
			"method":   method,
			"path":     path,
			"status":   status,
			"duration": duration,
		},
	})
}

// PushDB adds a database query log entry.
func (s *DevMonitorService) PushDB(sql, rows, duration string) {
	s.dbLogs.Push(LogEntry{
		Timestamp: time.Now(),
		Type:      "db",
		Data: map[string]string{
			"sql":      sql,
			"rows":     rows,
			"duration": duration,
		},
	})
}

// PushTest adds a test result log entry.
func (s *DevMonitorService) PushTest(name, status, output string) {
	s.testLogs.Push(LogEntry{
		Timestamp: time.Now(),
		Type:      "test",
		Data: map[string]string{
			"name":   name,
			"status": status,
			"output": output,
		},
	})
}

// ClearAll clears all log buffers.
func (s *DevMonitorService) ClearAll() {
	s.httpLogs.Clear()
	s.dbLogs.Clear()
	s.testLogs.Clear()
}

// Stats returns current buffer counts.
func (s *DevMonitorService) Stats() map[string]int {
	return map[string]int{
		"http":  s.httpLogs.Count(),
		"db":    s.dbLogs.Count(),
		"tests": s.testLogs.Count(),
	}
}
