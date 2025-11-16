// @kthulu:test:handlers:product
package handlers

import (
"encoding/json"
"fmt"
"net/http"
"net/http/httptest"
"strings"
"testing"

"github.com/gorilla/mux"
"my-kthulu-app/internal/adapters/http/modules/product/domain"
)

type fakeProductService struct {
store  map[uint]*domain.Product
nextID uint
}

func newFakeProductService() *fakeProductService {
return &fakeProductService{store: make(map[uint]*domain.Product), nextID: 1}
}

func (s *fakeProductService) CreateProduct(entity *domain.Product) error {
if entity.ID == 0 {
entity.ID = s.nextID
s.nextID++
}
s.store[entity.ID] = entity
return nil
}

func (s *fakeProductService) GetProductByID(id uint) (*domain.Product, error) {
if entity, ok := s.store[id]; ok {
return entity, nil
}
return nil, fmt.Errorf("not found")
}

func (s *fakeProductService) UpdateProduct(entity *domain.Product) error {
s.store[entity.ID] = entity
return nil
}

func (s *fakeProductService) DeleteProduct(id uint) error {
delete(s.store, id)
return nil
}

func (s *fakeProductService) ListProducts() ([]*domain.Product, error) {
items := make([]*domain.Product, 0, len(s.store))
for _, entity := range s.store {
items = append(items, entity)
}
return items, nil
}

func TestProductHandlerCRUD(t *testing.T) {
service := newFakeProductService()
handler := NewProductHandler(service)
router := mux.NewRouter()
handler.RegisterRoutes(router)

createReq := httptest.NewRequest(http.MethodPost, "/product", strings.NewReader(`{}`))
createRec := httptest.NewRecorder()
router.ServeHTTP(createRec, createReq)
if createRec.Code != http.StatusOK {
t.Fatalf("expected 200 got %d", createRec.Code)
}
var created domain.Product
if err := json.NewDecoder(createRec.Body).Decode(&created); err != nil {
t.Fatalf("decode failed: %v", err)
}

listReq := httptest.NewRequest(http.MethodGet, "/product", nil)
listRec := httptest.NewRecorder()
router.ServeHTTP(listRec, listReq)
if listRec.Code != http.StatusOK {
t.Fatalf("expected 200 got %d", listRec.Code)
}

getReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/product/%d", created.ID), nil)
getRec := httptest.NewRecorder()
router.ServeHTTP(getRec, getReq)
if getRec.Code != http.StatusOK {
t.Fatalf("expected 200 got %d", getRec.Code)
}

updateReq := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/product/%d", created.ID), strings.NewReader(`{}`))
updateRec := httptest.NewRecorder()
router.ServeHTTP(updateRec, updateReq)
if updateRec.Code != http.StatusOK {
t.Fatalf("expected 200 got %d", updateRec.Code)
}

deleteReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/product/%d", created.ID), nil)
deleteRec := httptest.NewRecorder()
router.ServeHTTP(deleteRec, deleteReq)
if deleteRec.Code != http.StatusNoContent {
t.Fatalf("expected 204 got %d", deleteRec.Code)
}
}
