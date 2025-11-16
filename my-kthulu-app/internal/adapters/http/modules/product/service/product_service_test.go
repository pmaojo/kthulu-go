// @kthulu:test:service:product
package service

import (
"testing"

"my-kthulu-app/internal/adapters/http/modules/product/domain"
)

type fakeProductRepository struct {
store  map[uint]*domain.Product
nextID uint
}

func newFakeProductRepository() *fakeProductRepository {
return &fakeProductRepository{
store:  make(map[uint]*domain.Product),
nextID: 1,
}
}

func (r *fakeProductRepository) Create(entity *domain.Product) error {
if entity.ID == 0 {
entity.ID = r.nextID
r.nextID++
}
r.store[entity.ID] = entity
return nil
}

func (r *fakeProductRepository) GetByID(id uint) (*domain.Product, error) {
return r.store[id], nil
}

func (r *fakeProductRepository) Update(entity *domain.Product) error {
r.store[entity.ID] = entity
return nil
}

func (r *fakeProductRepository) Delete(id uint) error {
delete(r.store, id)
return nil
}

func (r *fakeProductRepository) List() ([]*domain.Product, error) {
items := make([]*domain.Product, 0, len(r.store))
for _, item := range r.store {
items = append(items, item)
}
return items, nil
}

func TestProductServiceCRUD(t *testing.T) {
repo := newFakeProductRepository()
service := NewProductService(repo)
entity := &domain.Product{}
if err := service.CreateProduct(entity); err != nil {
t.Fatalf("create failed: %v", err)
}
if entity.ID == 0 {
t.Fatal("expected ID to be set")
}
if _, err := service.GetProductByID(entity.ID); err != nil {
t.Fatalf("get failed: %v", err)
}
if err := service.UpdateProduct(entity); err != nil {
t.Fatalf("update failed: %v", err)
}
items, err := service.ListProducts()
if err != nil {
t.Fatalf("list failed: %v", err)
}
if len(items) != 1 {
t.Fatalf("expected 1 item got %d", len(items))
}
if err := service.DeleteProduct(entity.ID); err != nil {
t.Fatalf("delete failed: %v", err)
}
}
