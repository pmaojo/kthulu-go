// @kthulu:test:handlers:users
package handlers

import (
"encoding/json"
"fmt"
"net/http"
"net/http/httptest"
"strings"
"testing"

"github.com/gorilla/mux"
"github.com/example/workshop-api/internal/adapters/http/modules/users/domain"
)

type fakeUsersService struct {
store  map[uint]*domain.Users
nextID uint
}

func newFakeUsersService() *fakeUsersService {
return &fakeUsersService{store: make(map[uint]*domain.Users), nextID: 1}
}

func (s *fakeUsersService) CreateUsers(entity *domain.Users) error {
if entity.ID == 0 {
entity.ID = s.nextID
s.nextID++
}
s.store[entity.ID] = entity
return nil
}

func (s *fakeUsersService) GetUsersByID(id uint) (*domain.Users, error) {
if entity, ok := s.store[id]; ok {
return entity, nil
}
return nil, fmt.Errorf("not found")
}

func (s *fakeUsersService) UpdateUsers(entity *domain.Users) error {
s.store[entity.ID] = entity
return nil
}

func (s *fakeUsersService) DeleteUsers(id uint) error {
delete(s.store, id)
return nil
}

func (s *fakeUsersService) ListUserses() ([]*domain.Users, error) {
items := make([]*domain.Users, 0, len(s.store))
for _, entity := range s.store {
items = append(items, entity)
}
return items, nil
}

func TestUsersHandlerCRUD(t *testing.T) {
service := newFakeUsersService()
handler := NewUsersHandler(service)
router := mux.NewRouter()
handler.RegisterRoutes(router)

createReq := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{}`))
createRec := httptest.NewRecorder()
router.ServeHTTP(createRec, createReq)
if createRec.Code != http.StatusOK {
t.Fatalf("expected 200 got %d", createRec.Code)
}
var created domain.Users
if err := json.NewDecoder(createRec.Body).Decode(&created); err != nil {
t.Fatalf("decode failed: %v", err)
}

listReq := httptest.NewRequest(http.MethodGet, "/users", nil)
listRec := httptest.NewRecorder()
router.ServeHTTP(listRec, listReq)
if listRec.Code != http.StatusOK {
t.Fatalf("expected 200 got %d", listRec.Code)
}

getReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users/%d", created.ID), nil)
getRec := httptest.NewRecorder()
router.ServeHTTP(getRec, getReq)
if getRec.Code != http.StatusOK {
t.Fatalf("expected 200 got %d", getRec.Code)
}

updateReq := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/users/%d", created.ID), strings.NewReader(`{}`))
updateRec := httptest.NewRecorder()
router.ServeHTTP(updateRec, updateReq)
if updateRec.Code != http.StatusOK {
t.Fatalf("expected 200 got %d", updateRec.Code)
}

deleteReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%d", created.ID), nil)
deleteRec := httptest.NewRecorder()
router.ServeHTTP(deleteRec, deleteReq)
if deleteRec.Code != http.StatusNoContent {
t.Fatalf("expected 204 got %d", deleteRec.Code)
}
}
