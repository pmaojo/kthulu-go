package adapterhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"go.uber.org/zap"

	"github.com/pmaojo/kthulu-go/backend/internal/adapters/http/middleware"
	"github.com/pmaojo/kthulu-go/backend/internal/domain"
	"github.com/pmaojo/kthulu-go/backend/internal/domain/repository"
	"github.com/pmaojo/kthulu-go/backend/internal/infrastructure/observability"
	"github.com/pmaojo/kthulu-go/backend/internal/usecase"
)

// mockContactRepository implements repository.ContactRepository for testing
type mockContactRepository struct {
	GetByIDFunc           func(ctx context.Context, organizationID, contactID uint) (*domain.Contact, error)
	GetAddressByIDFunc    func(ctx context.Context, contactID, addressID uint) (*domain.ContactAddress, error)
	UpdateAddressFunc     func(ctx context.Context, address *domain.ContactAddress) error
	DeleteAddressFunc     func(ctx context.Context, contactID, addressID uint) error
	SetPrimaryAddressFunc func(ctx context.Context, contactID, addressID uint) error
	GetPhoneByIDFunc      func(ctx context.Context, contactID, phoneID uint) (*domain.ContactPhone, error)
	UpdatePhoneFunc       func(ctx context.Context, phone *domain.ContactPhone) error
	DeletePhoneFunc       func(ctx context.Context, contactID, phoneID uint) error
	SetPrimaryPhoneFunc   func(ctx context.Context, contactID, phoneID uint) error
}

func (m *mockContactRepository) Create(ctx context.Context, contact *domain.Contact) error {
	return nil
}
func (m *mockContactRepository) GetByID(ctx context.Context, organizationID, contactID uint) (*domain.Contact, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, organizationID, contactID)
	}
	return &domain.Contact{ID: contactID, OrganizationID: organizationID}, nil
}
func (m *mockContactRepository) GetByEmail(ctx context.Context, organizationID uint, email string) (*domain.Contact, error) {
	return nil, domain.ErrContactNotFound
}
func (m *mockContactRepository) Update(ctx context.Context, contact *domain.Contact) error {
	return nil
}
func (m *mockContactRepository) Delete(ctx context.Context, organizationID, contactID uint) error {
	return nil
}
func (m *mockContactRepository) List(ctx context.Context, organizationID uint, filters repository.ContactFilters) ([]*domain.Contact, int64, error) {
	return nil, 0, nil
}
func (m *mockContactRepository) CreateAddress(ctx context.Context, address *domain.ContactAddress) error {
	return nil
}
func (m *mockContactRepository) GetAddressesByContactID(ctx context.Context, contactID uint) ([]*domain.ContactAddress, error) {
	return nil, nil
}
func (m *mockContactRepository) GetAddressByID(ctx context.Context, contactID, addressID uint) (*domain.ContactAddress, error) {
	if m.GetAddressByIDFunc != nil {
		return m.GetAddressByIDFunc(ctx, contactID, addressID)
	}
	return &domain.ContactAddress{ID: addressID, ContactID: contactID}, nil
}
func (m *mockContactRepository) UpdateAddress(ctx context.Context, address *domain.ContactAddress) error {
	if m.UpdateAddressFunc != nil {
		return m.UpdateAddressFunc(ctx, address)
	}
	return nil
}
func (m *mockContactRepository) DeleteAddress(ctx context.Context, contactID, addressID uint) error {
	if m.DeleteAddressFunc != nil {
		return m.DeleteAddressFunc(ctx, contactID, addressID)
	}
	return nil
}
func (m *mockContactRepository) SetPrimaryAddress(ctx context.Context, contactID, addressID uint) error {
	if m.SetPrimaryAddressFunc != nil {
		return m.SetPrimaryAddressFunc(ctx, contactID, addressID)
	}
	return nil
}
func (m *mockContactRepository) CreatePhone(ctx context.Context, phone *domain.ContactPhone) error {
	return nil
}
func (m *mockContactRepository) GetPhonesByContactID(ctx context.Context, contactID uint) ([]*domain.ContactPhone, error) {
	return nil, nil
}
func (m *mockContactRepository) GetPhoneByID(ctx context.Context, contactID, phoneID uint) (*domain.ContactPhone, error) {
	if m.GetPhoneByIDFunc != nil {
		return m.GetPhoneByIDFunc(ctx, contactID, phoneID)
	}
	return &domain.ContactPhone{ID: phoneID, ContactID: contactID}, nil
}
func (m *mockContactRepository) UpdatePhone(ctx context.Context, phone *domain.ContactPhone) error {
	if m.UpdatePhoneFunc != nil {
		return m.UpdatePhoneFunc(ctx, phone)
	}
	return nil
}
func (m *mockContactRepository) DeletePhone(ctx context.Context, contactID, phoneID uint) error {
	if m.DeletePhoneFunc != nil {
		return m.DeletePhoneFunc(ctx, contactID, phoneID)
	}
	return nil
}
func (m *mockContactRepository) SetPrimaryPhone(ctx context.Context, contactID, phoneID uint) error {
	if m.SetPrimaryPhoneFunc != nil {
		return m.SetPrimaryPhoneFunc(ctx, contactID, phoneID)
	}
	return nil
}
func (m *mockContactRepository) BulkCreate(ctx context.Context, contacts []*domain.Contact) error {
	return nil
}
func (m *mockContactRepository) BulkUpdate(ctx context.Context, contacts []*domain.Contact) error {
	return nil
}
func (m *mockContactRepository) BulkDelete(ctx context.Context, organizationID uint, contactIDs []uint) error {
	return nil
}
func (m *mockContactRepository) GetContactStats(ctx context.Context, organizationID uint) (*repository.ContactStats, error) {
	return nil, nil
}

func TestContactHandler_AddressRoutes(t *testing.T) {
	repo := &mockContactRepository{
		GetByIDFunc: func(ctx context.Context, orgID, contactID uint) (*domain.Contact, error) {
			return &domain.Contact{ID: contactID, OrganizationID: orgID}, nil
		},
		GetAddressByIDFunc: func(ctx context.Context, contactID, addressID uint) (*domain.ContactAddress, error) {
			return &domain.ContactAddress{ID: addressID, ContactID: contactID}, nil
		},
	}
	uc := usecase.NewContactUseCase(repo, observability.NewLoggerFromZap(zap.NewNop()))
	handler := NewContactHandler(uc, observability.NewLoggerFromZap(zap.NewNop()))

	router := chi.NewRouter()
	router.Use(middleware.OrganizationContextMiddleware)
	handler.RegisterRoutes(router)

	// Update address
	body := bytes.NewBufferString(`{"type":"billing","addressLine1":"123 Main","city":"Town","country":"US","isPrimary":false}`)
	req := httptest.NewRequest(http.MethodPatch, "/contacts/1/addresses/2", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Organization-ID", "1")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var addr domain.ContactAddress
	if err := json.NewDecoder(w.Body).Decode(&addr); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if addr.ID != 2 || addr.ContactID != 1 || addr.AddressLine1 != "123 Main" {
		t.Fatalf("unexpected address: %+v", addr)
	}

	// Delete address
	req = httptest.NewRequest(http.MethodDelete, "/contacts/1/addresses/2", nil)
	req.Header.Set("X-Organization-ID", "1")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}

	// Set primary address
	req = httptest.NewRequest(http.MethodPost, "/contacts/1/addresses/2/set-primary", nil)
	req.Header.Set("X-Organization-ID", "1")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestContactHandler_PhoneRoutes(t *testing.T) {
	repo := &mockContactRepository{
		GetByIDFunc: func(ctx context.Context, orgID, contactID uint) (*domain.Contact, error) {
			return &domain.Contact{ID: contactID, OrganizationID: orgID}, nil
		},
		GetPhoneByIDFunc: func(ctx context.Context, contactID, phoneID uint) (*domain.ContactPhone, error) {
			return &domain.ContactPhone{ID: phoneID, ContactID: contactID}, nil
		},
	}
	uc := usecase.NewContactUseCase(repo, observability.NewLoggerFromZap(zap.NewNop()))
	handler := NewContactHandler(uc, observability.NewLoggerFromZap(zap.NewNop()))

	router := chi.NewRouter()
	router.Use(middleware.OrganizationContextMiddleware)
	handler.RegisterRoutes(router)

	// Update phone
	body := bytes.NewBufferString(`{"type":"work","number":"123456","isPrimary":false}`)
	req := httptest.NewRequest(http.MethodPatch, "/contacts/1/phones/2", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Organization-ID", "1")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var phone domain.ContactPhone
	if err := json.NewDecoder(w.Body).Decode(&phone); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if phone.ID != 2 || phone.ContactID != 1 || phone.Number != "123456" {
		t.Fatalf("unexpected phone: %+v", phone)
	}

	// Delete phone
	req = httptest.NewRequest(http.MethodDelete, "/contacts/1/phones/2", nil)
	req.Header.Set("X-Organization-ID", "1")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}

	// Set primary phone
	req = httptest.NewRequest(http.MethodPost, "/contacts/1/phones/2/set-primary", nil)
	req.Header.Set("X-Organization-ID", "1")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}
