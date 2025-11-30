package debugui

import (
	"os"
	"testing"
	"time"
)

func TestParseLog_Gin(t *testing.T) {
	line := "[GIN] 2023/10/26 - 12:00:00 | 200 | 1.2ms | ::1 | GET /api/v1/users"
	entry := parseLog(line)

	if entry.Type != LogTypeHTTP {
		t.Errorf("Expected LogTypeHTTP, got %v", entry.Type)
	}
	if entry.Method != "GET" {
		t.Errorf("Expected GET, got %s", entry.Method)
	}
	if entry.Status != "200" {
		t.Errorf("Expected 200, got %s", entry.Status)
	}
	if entry.Path != "/api/v1/users" {
		t.Errorf("Expected /api/v1/users, got %s", entry.Path)
	}
}

func TestParseLog_GORM(t *testing.T) {
	line := "2023/10/26 12:00:00 /path/to/file.go:123 [1.2ms] [rows:1] SELECT * FROM users"
	entry := parseLog(line)

	if entry.Type != LogTypeDB {
		t.Errorf("Expected LogTypeDB, got %v", entry.Type)
	}
	if entry.Rows != "1" {
		t.Errorf("Expected 1 row, got %s", entry.Rows)
	}
	if entry.SQL != "SELECT * FROM users" {
		t.Errorf("Expected SQL, got %s", entry.SQL)
	}
}

func TestAppendToFile(t *testing.T) {
	m := Model{}
	entry := LogEntry{
		Timestamp: time.Now(),
		Type:      LogTypeHTTP,
		Raw:       "test log",
	}

	// Clean up before/after
	os.Remove("debug_events.jsonl")
	defer os.Remove("debug_events.jsonl")

	m.appendToFile(entry)

	content, err := os.ReadFile("debug_events.jsonl")
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if len(content) == 0 {
		t.Error("File is empty")
	}
}
