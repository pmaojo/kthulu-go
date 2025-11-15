package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pmaojo/kthulu-go/backend/internal/domain"
	"github.com/pmaojo/kthulu-go/backend/internal/repository"
)

// mockRoleRepository implements repository.RoleRepository for testing.
type mockRoleRepository struct {
	t              *testing.T
	expectedUserID uint
	role           *domain.Role
	called         bool
}

func (m *mockRoleRepository) FindByUserID(ctx context.Context, userID uint) (*domain.Role, error) {
	m.called = true
	if userID != m.expectedUserID {
		m.t.Fatalf("expected userID %d, got %d", m.expectedUserID, userID)
	}
	return m.role, nil
}

// The following methods satisfy the repository.RoleRepository interface but are not used in tests.
func (m *mockRoleRepository) Create(context.Context, *domain.Role) error { return nil }
func (m *mockRoleRepository) FindByID(context.Context, uint) (*domain.Role, error) {
	m.t.Fatalf("FindByID should not be called")
	return nil, nil
}
func (m *mockRoleRepository) FindByName(context.Context, string) (*domain.Role, error) {
	return nil, nil
}
func (m *mockRoleRepository) Update(context.Context, *domain.Role) error         { return nil }
func (m *mockRoleRepository) Delete(context.Context, uint) error                 { return nil }
func (m *mockRoleRepository) List(context.Context) ([]*domain.Role, error)       { return nil, nil }
func (m *mockRoleRepository) Count(context.Context) (int64, error)               { return 0, nil }
func (m *mockRoleRepository) ExistsByName(context.Context, string) (bool, error) { return false, nil }
func (m *mockRoleRepository) ExistsByID(context.Context, uint) (bool, error)     { return false, nil }
func (m *mockRoleRepository) AddPermission(context.Context, uint, uint) error    { return nil }
func (m *mockRoleRepository) RemovePermission(context.Context, uint, uint) error { return nil }
func (m *mockRoleRepository) GetRolePermissions(context.Context, uint) ([]*domain.Permission, error) {
	return nil, nil
}

// Ensure mockRoleRepository satisfies the interface at compile time.
var _ repository.RoleRepository = (*mockRoleRepository)(nil)

func TestRequireRoleUsesFindByUserID(t *testing.T) {
	repo := &mockRoleRepository{
		t:              t,
		expectedUserID: 42,
		role:           &domain.Role{ID: 1, Name: domain.RoleAdmin},
	}

	nextCalled := false
	handler := RequireRole(repo, domain.RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), UserIDKey, uint(42))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req.WithContext(ctx))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if !nextCalled {
		t.Fatalf("next handler was not called")
	}
	if !repo.called {
		t.Fatalf("FindByUserID was not called")
	}
}

func TestRequirePermissionUsesFindByUserID(t *testing.T) {
	repo := &mockRoleRepository{
		t:              t,
		expectedUserID: 7,
		role: &domain.Role{ID: 2, Name: domain.RoleUser, Permissions: []domain.Permission{{
			Resource: "docs",
			Action:   "read",
		}}},
	}

	handler := RequirePermission(repo, "docs", "read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), UserIDKey, uint(7))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req.WithContext(ctx))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if !repo.called {
		t.Fatalf("FindByUserID was not called")
	}
}
