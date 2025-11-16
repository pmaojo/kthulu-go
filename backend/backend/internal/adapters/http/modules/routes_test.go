package modules

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type mockRegistrar struct{ called bool }

func (m *mockRegistrar) RegisterRoutes(r chi.Router) {
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	m.called = true
}

func TestRouteRegistry_RegisterAllRoutes(t *testing.T) {
	rr := NewRouteRegistry(zap.NewNop())
	mr := &mockRegistrar{}
	rr.Register(mr)

	router := chi.NewRouter()
	rr.RegisterAllRoutes(router)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if !mr.called {
		t.Fatalf("expected registrar to be called")
	}
	if w.Code != http.StatusOK || w.Body.String() != "ok" {
		t.Fatalf("unexpected response: %d %s", w.Code, w.Body.String())
	}
}

func TestRouteRegistry_Register_NilRegistrar(t *testing.T) {
	rr := NewRouteRegistry(zap.NewNop())
	rr.Register(nil)
	if len(rr.registrars) != 0 {
		t.Fatalf("expected no registrars, got %d", len(rr.registrars))
	}
}

func TestRouteRegistry_RegisterAllRoutes_NilRegistrar(t *testing.T) {
	rr := NewRouteRegistry(zap.NewNop())
	rr.registrars = append(rr.registrars, nil)

	router := chi.NewRouter()
	rr.RegisterAllRoutes(router)
}
