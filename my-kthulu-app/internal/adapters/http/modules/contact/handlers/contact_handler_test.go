// @kthulu:test:handlers:contact
package handlers

import (
"encoding/json"
"fmt"
"net/http"
"net/http/httptest"
"strings"
"testing"

"github.com/gorilla/mux"
"my-kthulu-app/internal/adapters/http/modules/contact/domain"
)

type fakeContactService struct {
store  map[uint]*domain.Contact
nextID uint
}

func newFakeContactService() *fakeContactService {
return &fakeContactService{store: make(map[uint]*domain.Contact), nextID: 1}
}

func (s *fakeContactService) CreateContact(entity *domain.Contact) error {
if entity.ID == 0 {
entity.ID = s.nextID
s.nextID++
}
s.store[entity.ID] = entity
return nil
}

func (s *fakeContactService) GetContactByID(id uint) (*domain.Contact, error) {
if entity, ok := s.store[id]; ok {
return entity, nil
}
return nil, fmt.Errorf("not found")
}

func (s *fakeContactService) UpdateContact(entity *domain.Contact) error {
s.store[entity.ID] = entity
return nil
}

func (s *fakeContactService) DeleteContact(id uint) error {
delete(s.store, id)
return nil
}

func (s *fakeContactService) ListContacts() ([]*domain.Contact, error) {
items := make([]*domain.Contact, 0, len(s.store))
for _, entity := range s.store {
items = append(items, entity)
}
return items, nil
}

func TestContactHandlerCRUD(t *testing.T) {
service := newFakeContactService()
handler := NewContactHandler(service)
router := mux.NewRouter()
handler.RegisterRoutes(router)

createReq := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader(`{}`))
createRec := httptest.NewRecorder()
router.ServeHTTP(createRec, createReq)
if createRec.Code != http.StatusOK {
t.Fatalf("expected 200 got %d", createRec.Code)
}
var created domain.Contact
if err := json.NewDecoder(createRec.Body).Decode(&created); err != nil {
t.Fatalf("decode failed: %v", err)
}

listReq := httptest.NewRequest(http.MethodGet, "/contact", nil)
listRec := httptest.NewRecorder()
router.ServeHTTP(listRec, listReq)
if listRec.Code != http.StatusOK {
t.Fatalf("expected 200 got %d", listRec.Code)
}

getReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/contact/%d", created.ID), nil)
getRec := httptest.NewRecorder()
router.ServeHTTP(getRec, getReq)
if getRec.Code != http.StatusOK {
t.Fatalf("expected 200 got %d", getRec.Code)
}

updateReq := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/contact/%d", created.ID), strings.NewReader(`{}`))
updateRec := httptest.NewRecorder()
router.ServeHTTP(updateRec, updateReq)
if updateRec.Code != http.StatusOK {
t.Fatalf("expected 200 got %d", updateRec.Code)
}

deleteReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/contact/%d", created.ID), nil)
deleteRec := httptest.NewRecorder()
router.ServeHTTP(deleteRec, deleteReq)
if deleteRec.Code != http.StatusNoContent {
t.Fatalf("expected 204 got %d", deleteRec.Code)
}
}
