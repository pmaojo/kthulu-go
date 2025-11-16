// @kthulu:test:handlers:auth
package handlers

import (
"encoding/json"
"fmt"
"net/http"
"net/http/httptest"
"strings"
"testing"

"github.com/gorilla/mux"
"my-kthulu-app/internal/adapters/http/modules/auth/domain"
)

type fakeAuthService struct {
store  map[uint]*domain.Auth
nextID uint
}

func newFakeAuthService() *fakeAuthService {
return &fakeAuthService{store: make(map[uint]*domain.Auth), nextID: 1}
}

func (s *fakeAuthService) CreateAuth(entity *domain.Auth) error {
if entity.ID == 0 {
entity.ID = s.nextID
s.nextID++
}
s.store[entity.ID] = entity
return nil
}

func (s *fakeAuthService) GetAuthByID(id uint) (*domain.Auth, error) {
if entity, ok := s.store[id]; ok {
return entity, nil
}
return nil, fmt.Errorf("not found")
}

func (s *fakeAuthService) UpdateAuth(entity *domain.Auth) error {
s.store[entity.ID] = entity
return nil
}

func (s *fakeAuthService) DeleteAuth(id uint) error {
delete(s.store, id)
return nil
}

func (s *fakeAuthService) ListAuths() ([]*domain.Auth, error) {
items := make([]*domain.Auth, 0, len(s.store))
for _, entity := range s.store {
items = append(items, entity)
}
return items, nil
}

func TestAuthHandlerCRUD(t *testing.T) {
service := newFakeAuthService()
handler := NewAuthHandler(service)
router := mux.NewRouter()
handler.RegisterRoutes(router)

createReq := httptest.NewRequest(http.MethodPost, "/auth", strings.NewReader(`{}`))
createRec := httptest.NewRecorder()
router.ServeHTTP(createRec, createReq)
if createRec.Code != http.StatusOK {
t.Fatalf("expected 200 got %d", createRec.Code)
}
var created domain.Auth
if err := json.NewDecoder(createRec.Body).Decode(&created); err != nil {
t.Fatalf("decode failed: %v", err)
}

listReq := httptest.NewRequest(http.MethodGet, "/auth", nil)
listRec := httptest.NewRecorder()
router.ServeHTTP(listRec, listReq)
if listRec.Code != http.StatusOK {
t.Fatalf("expected 200 got %d", listRec.Code)
}

getReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/auth/%d", created.ID), nil)
getRec := httptest.NewRecorder()
router.ServeHTTP(getRec, getReq)
if getRec.Code != http.StatusOK {
t.Fatalf("expected 200 got %d", getRec.Code)
}

updateReq := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/auth/%d", created.ID), strings.NewReader(`{}`))
updateRec := httptest.NewRecorder()
router.ServeHTTP(updateRec, updateReq)
if updateRec.Code != http.StatusOK {
t.Fatalf("expected 200 got %d", updateRec.Code)
}

deleteReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/auth/%d", created.ID), nil)
deleteRec := httptest.NewRecorder()
router.ServeHTTP(deleteRec, deleteReq)
if deleteRec.Code != http.StatusNoContent {
t.Fatalf("expected 204 got %d", deleteRec.Code)
}
}
