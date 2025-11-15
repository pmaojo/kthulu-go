package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	flagcfg "github.com/pmaojo/kthulu-go/backend/internal/modules/flags"
)

func TestFlagsMiddleware(t *testing.T) {
	cfg := flagcfg.HeaderConfig{"X-Test": "test"}
	mw := FlagsMiddleware(cfg)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if val, ok := GetFlag(r.Context(), "test"); !ok || val != "on" {
			t.Fatalf("expected flag value 'on', got %q", val)
		}
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Test", "on")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
}
