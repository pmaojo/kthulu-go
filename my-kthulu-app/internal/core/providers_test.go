// @kthulu:test:core
package core

import "testing"

func TestGetEnv(t *testing.T) {
t.Setenv("KTHULU_SAMPLE", "value")
if got := getEnv("KTHULU_SAMPLE", "fallback"); got != "value" {
t.Fatalf("expected value got %s", got)
}
if got := getEnv("MISSING", "fallback"); got != "fallback" {
t.Fatalf("expected fallback got %s", got)
}
}

func TestNewDatabaseTestMode(t *testing.T) {
t.Setenv("KTHULU_TEST_MODE", "1")
db, err := NewDatabase()
if err != nil {
t.Fatalf("expected sqlite database, got error: %v", err)
}
if db == nil {
t.Fatal("expected database instance")
}
}
