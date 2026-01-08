package adapterhttp

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pmaojo/kthulu-go/backend/internal/domain"
	"github.com/pmaojo/kthulu-go/backend/internal/usecase"
	"go.uber.org/zap"
)

// stubAuthUseCase provides minimal implementations for AuthUseCase.
type stubAuthUseCase struct{}

func (s *stubAuthUseCase) Login(ctx context.Context, req usecase.LoginRequest) (*usecase.AuthResponse, error) {
	return &usecase.AuthResponse{User: &domain.User{ID: 1}}, nil
}
func (s *stubAuthUseCase) Register(ctx context.Context, req usecase.RegisterRequest) (*usecase.AuthResponse, error) {
	return &usecase.AuthResponse{User: &domain.User{ID: 1}}, nil
}
func (s *stubAuthUseCase) Refresh(ctx context.Context, req usecase.RefreshRequest) (*usecase.AuthResponse, error) {
	return &usecase.AuthResponse{User: &domain.User{ID: 1}}, nil
}
func (s *stubAuthUseCase) Confirm(ctx context.Context, req usecase.ConfirmRequest) (*usecase.AuthResponse, error) {
	return &usecase.AuthResponse{User: &domain.User{ID: 1}}, nil
}
func (s *stubAuthUseCase) Logout(ctx context.Context, req usecase.LogoutRequest) error {
	return nil
}
func (s *stubAuthUseCase) ResendConfirmation(ctx context.Context, req usecase.ResendConfirmationRequest) error {
	return nil
}

// failingWriter fails on the first write to simulate encoding errors.
type failingWriter struct {
	header http.Header
	status int
	body   bytes.Buffer
	fail   bool
}

func newFailingWriter() *failingWriter {
	return &failingWriter{header: http.Header{}, fail: true}
}

func (f *failingWriter) Header() http.Header { return f.header }

func (f *failingWriter) Write(b []byte) (int, error) {
	if f.fail {
		f.fail = false
		return 0, errors.New("write error")
	}
	return f.body.Write(b)
}

func (f *failingWriter) WriteHeader(status int) { f.status = status }

func (f *failingWriter) BodyString() string { return f.body.String() }

func TestAuthHandler_LoginEncodeError(t *testing.T) {
	h := NewAuthHandler(&stubAuthUseCase{}, zap.NewNop())
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{"email":"a@b.com","password":"pw"}`))
	w := newFailingWriter()
	h.login(w, req)
	if w.status != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.status)
	}
	if !strings.Contains(w.BodyString(), "Failed to encode response") {
		t.Fatalf("unexpected body: %s", w.BodyString())
	}
}

func TestAuthHandler_RegisterEncodeError(t *testing.T) {
	h := NewAuthHandler(&stubAuthUseCase{}, zap.NewNop())
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{"email":"a@b.com","password":"pw"}`))
	w := newFailingWriter()
	h.register(w, req)
	if w.status != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.status)
	}
	if !strings.Contains(w.BodyString(), "Failed to encode response") {
		t.Fatalf("unexpected body: %s", w.BodyString())
	}
}

func TestAuthHandler_RefreshEncodeError(t *testing.T) {
	h := NewAuthHandler(&stubAuthUseCase{}, zap.NewNop())
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBufferString(`{"refreshToken":"tok"}`))
	w := newFailingWriter()
	h.refresh(w, req)
	if w.status != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.status)
	}
	if !strings.Contains(w.BodyString(), "Failed to encode response") {
		t.Fatalf("unexpected body: %s", w.BodyString())
	}
}

func TestAuthHandler_ConfirmEncodeError(t *testing.T) {
	h := NewAuthHandler(&stubAuthUseCase{}, zap.NewNop())
	req := httptest.NewRequest(http.MethodPost, "/auth/confirm", bytes.NewBufferString(`{"email":"a@b.com","confirmationCode":"123"}`))
	w := newFailingWriter()
	h.confirm(w, req)
	if w.status != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.status)
	}
	if !strings.Contains(w.BodyString(), "Failed to encode response") {
		t.Fatalf("unexpected body: %s", w.BodyString())
	}
}
