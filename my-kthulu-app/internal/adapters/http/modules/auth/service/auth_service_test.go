// @kthulu:test:service:auth
package service

import (
"testing"

"my-kthulu-app/internal/adapters/http/modules/auth/domain"
)

type fakeAuthRepository struct {
store  map[uint]*domain.Auth
nextID uint
}

func newFakeAuthRepository() *fakeAuthRepository {
return &fakeAuthRepository{
store:  make(map[uint]*domain.Auth),
nextID: 1,
}
}

func (r *fakeAuthRepository) Create(entity *domain.Auth) error {
if entity.ID == 0 {
entity.ID = r.nextID
r.nextID++
}
r.store[entity.ID] = entity
return nil
}

func (r *fakeAuthRepository) GetByID(id uint) (*domain.Auth, error) {
return r.store[id], nil
}

func (r *fakeAuthRepository) Update(entity *domain.Auth) error {
r.store[entity.ID] = entity
return nil
}

func (r *fakeAuthRepository) Delete(id uint) error {
delete(r.store, id)
return nil
}

func (r *fakeAuthRepository) List() ([]*domain.Auth, error) {
items := make([]*domain.Auth, 0, len(r.store))
for _, item := range r.store {
items = append(items, item)
}
return items, nil
}

func TestAuthServiceCRUD(t *testing.T) {
repo := newFakeAuthRepository()
service := NewAuthService(repo)
entity := &domain.Auth{}
if err := service.CreateAuth(entity); err != nil {
t.Fatalf("create failed: %v", err)
}
if entity.ID == 0 {
t.Fatal("expected ID to be set")
}
if _, err := service.GetAuthByID(entity.ID); err != nil {
t.Fatalf("get failed: %v", err)
}
if err := service.UpdateAuth(entity); err != nil {
t.Fatalf("update failed: %v", err)
}
items, err := service.ListAuths()
if err != nil {
t.Fatalf("list failed: %v", err)
}
if len(items) != 1 {
t.Fatalf("expected 1 item got %d", len(items))
}
if err := service.DeleteAuth(entity.ID); err != nil {
t.Fatalf("delete failed: %v", err)
}
}
