package middleware

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"backend/internal/observability"
)

func TestGenerateRequestIDRandReadError(t *testing.T) {
	originalRandRead := randRead
	randRead = func([]byte) (int, error) {
		return 0, errors.New("rand error")
	}
	defer func() { randRead = originalRandRead }()

	core, logs := observer.New(zap.ErrorLevel)
	logger := &observability.ZapLogger{zap.New(core)}

	id := generateRequestID(logger)

	if _, err := uuid.Parse(id); err != nil {
		t.Fatalf("expected UUID fallback, got %s", id)
	}

	if logs.Len() != 1 {
		t.Fatalf("expected 1 log entry, got %d", logs.Len())
	}
	if logs.All()[0].Message != "failed to generate random request ID" {
		t.Fatalf("unexpected log message: %s", logs.All()[0].Message)
	}
}
