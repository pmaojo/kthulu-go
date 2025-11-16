package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"

	"github.com/pmaojo/kthulu-go/backend/internal/infrastructure/observability"
)

type failingResponseWriter struct {
	header http.Header
	code   int
}

func (f *failingResponseWriter) Header() http.Header {
	if f.header == nil {
		f.header = make(http.Header)
	}
	return f.header
}

func (f *failingResponseWriter) Write([]byte) (int, error) {
	return 0, errors.New("write error")
}

func (f *failingResponseWriter) WriteHeader(statusCode int) {
	f.code = statusCode
}

func TestHealthCheckHandler_EncodeFailure(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectPing()

	handler := HealthCheckHandler(db, &observability.ZapLogger{zap.NewNop()}, "test-version")

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := &failingResponseWriter{}

	handler(w, req)

	if w.code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, w.code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}
