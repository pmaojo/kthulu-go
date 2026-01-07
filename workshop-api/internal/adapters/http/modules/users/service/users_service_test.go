// @kthulu:test:service:users
package service

import (
"testing"

"github.com/example/workshop-api/internal/adapters/http/modules/users/domain"
)

type fakeUsersRepository struct {
store  map[uint]*domain.Users
nextID uint
}

func newFakeUsersRepository() *fakeUsersRepository {
return &fakeUsersRepository{
store:  make(map[uint]*domain.Users),
nextID: 1,
}
}

func (r *fakeUsersRepository) Create(entity *domain.Users) error {
if entity.ID == 0 {
entity.ID = r.nextID
r.nextID++
}
r.store[entity.ID] = entity
return nil
}

func (r *fakeUsersRepository) GetByID(id uint) (*domain.Users, error) {
return r.store[id], nil
}

func (r *fakeUsersRepository) Update(entity *domain.Users) error {
r.store[entity.ID] = entity
return nil
}

func (r *fakeUsersRepository) Delete(id uint) error {
delete(r.store, id)
return nil
}

func (r *fakeUsersRepository) List() ([]*domain.Users, error) {
items := make([]*domain.Users, 0, len(r.store))
for _, item := range r.store {
items = append(items, item)
}
return items, nil
}

func TestUsersServiceCRUD(t *testing.T) {
repo := newFakeUsersRepository()
service := NewUsersService(repo)
entity := &domain.Users{}
if err := service.CreateUsers(entity); err != nil {
t.Fatalf("create failed: %v", err)
}
if entity.ID == 0 {
t.Fatal("expected ID to be set")
}
if _, err := service.GetUsersByID(entity.ID); err != nil {
t.Fatalf("get failed: %v", err)
}
if err := service.UpdateUsers(entity); err != nil {
t.Fatalf("update failed: %v", err)
}
items, err := service.ListUserses()
if err != nil {
t.Fatalf("list failed: %v", err)
}
if len(items) != 1 {
t.Fatalf("expected 1 item got %d", len(items))
}
if err := service.DeleteUsers(entity.ID); err != nil {
t.Fatalf("delete failed: %v", err)
}
}
