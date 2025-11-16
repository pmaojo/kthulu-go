package common

import (
	"testing"
	"time"
)

func TestParseAndFormatDateTime(t *testing.T) {
	now := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	formatted := FormatDateTime(now)
	if formatted != "2024-01-02 03:04:05" {
		t.Fatalf("unexpected format: %s", formatted)
	}
	parsed, err := ParseDateTime(formatted)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if !parsed.Equal(now) {
		t.Fatalf("expected %v, got %v", now, parsed)
	}
}

func TestNowFormat(t *testing.T) {
	tNow := Now()
	if _, err := ParseDateTime(tNow); err != nil {
		t.Fatalf("Now returned invalid format: %v", err)
	}
}
