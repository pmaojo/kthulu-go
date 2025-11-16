// @kthulu:test:handlers:user
package handlers

import (
"encoding/json"
"fmt"
"net/http"
"net/http/httptest"
"strings"
"testing"

"github.com/gorilla/mux"
"my-new-project/internal/adapters/http/modules/user/domain"
)

type fakeUserService struct {
store  map[uint]*domain.User
nextID uint
}

func newFakeUserService() *fakeUserService {
return &fakeUserService{store: make(map[uint]*domain.User), nextID: 1}
}

func (s *fakeUserService) CreateUser(entity *domain.User) error {
if entity.ID == 0 {
entity.ID = s.nextID
s.nextID++
}
s.store[entity.ID] = entity
return nil
}

func (s *fakeUserService) GetUserByID(id uint) (*domain.User, error) {
if entity, ok := s.store[id]; ok {
return entity, nil
}
return nil, fmt.Errorf("not found")
}

func (s *fakeUserService) UpdateUser(entity *domain.User) error {
s.store[entity.ID] = entity
return nil
}

func (s *fakeUserService) DeleteUser(id uint) error {
delete(s.store, id)
return nil
}

func (s *fakeUserService) ListUsers() ([]*domain.User, error) {
items := make([]*domain.User, 0, len(s.store))
for _, entity := range s.store {
items = append(items, entity)
}
return items, nil
}

func TestUserHandlerCRUD(t *testing.T) {
service := newFakeUserService()
handler := NewUserHandler(service)
router := mux.NewRouter()
handler.RegisterRoutes(router)

createReq := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(`{}`))
createRec := httptest.NewRecorder()
router.ServeHTTP(createRec, createReq)
if createRec.Code != http.StatusOK {
t.Fatalf("expected 200 got %d", createRec.Code)
}
var created domain.User
if err := json.NewDecoder(createRec.Body).Decode(&created); err != nil {
t.Fatalf("decode failed: %v", err)
}

listReq := httptest.NewRequest(http.MethodGet, "/user", nil)
listRec := httptest.NewRecorder()
router.ServeHTTP(listRec, listReq)
if listRec.Code != http.StatusOK {
t.Fatalf("expected 200 got %d", listRec.Code)
}

getReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/user/%d", created.ID), nil)
getRec := httptest.NewRecorder()
router.ServeHTTP(getRec, getReq)
if getRec.Code != http.StatusOK {
t.Fatalf("expected 200 got %d", getRec.Code)
}

updateReq := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/user/%d", created.ID), strings.NewReader(`{}`))
updateRec := httptest.NewRecorder()
router.ServeHTTP(updateRec, updateReq)
if updateRec.Code != http.StatusOK {
t.Fatalf("expected 200 got %d", updateRec.Code)
}

deleteReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/user/%d", created.ID), nil)
deleteRec := httptest.NewRecorder()
router.ServeHTTP(deleteRec, deleteReq)
if deleteRec.Code != http.StatusNoContent {
t.Fatalf("expected 204 got %d", deleteRec.Code)
}
}
