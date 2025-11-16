// @kthulu:test:service:user
package service

import (
"testing"

"my-kthulu-app/internal/adapters/http/modules/user/domain"
)

type fakeUserRepository struct {
store  map[uint]*domain.User
nextID uint
}

func newFakeUserRepository() *fakeUserRepository {
return &fakeUserRepository{
store:  make(map[uint]*domain.User),
nextID: 1,
}
}

func (r *fakeUserRepository) Create(entity *domain.User) error {
if entity.ID == 0 {
entity.ID = r.nextID
r.nextID++
}
r.store[entity.ID] = entity
return nil
}

func (r *fakeUserRepository) GetByID(id uint) (*domain.User, error) {
return r.store[id], nil
}

func (r *fakeUserRepository) Update(entity *domain.User) error {
r.store[entity.ID] = entity
return nil
}

func (r *fakeUserRepository) Delete(id uint) error {
delete(r.store, id)
return nil
}

func (r *fakeUserRepository) List() ([]*domain.User, error) {
items := make([]*domain.User, 0, len(r.store))
for _, item := range r.store {
items = append(items, item)
}
return items, nil
}

func TestUserServiceCRUD(t *testing.T) {
repo := newFakeUserRepository()
service := NewUserService(repo)
entity := &domain.User{}
if err := service.CreateUser(entity); err != nil {
t.Fatalf("create failed: %v", err)
}
if entity.ID == 0 {
t.Fatal("expected ID to be set")
}
if _, err := service.GetUserByID(entity.ID); err != nil {
t.Fatalf("get failed: %v", err)
}
if err := service.UpdateUser(entity); err != nil {
t.Fatalf("update failed: %v", err)
}
items, err := service.ListUsers()
if err != nil {
t.Fatalf("list failed: %v", err)
}
if len(items) != 1 {
t.Fatalf("expected 1 item got %d", len(items))
}
if err := service.DeleteUser(entity.ID); err != nil {
t.Fatalf("delete failed: %v", err)
}
}
