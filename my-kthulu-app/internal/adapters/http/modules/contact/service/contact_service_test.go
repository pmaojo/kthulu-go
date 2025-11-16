// @kthulu:test:service:contact
package service

import (
"testing"

"my-kthulu-app/internal/adapters/http/modules/contact/domain"
)

type fakeContactRepository struct {
store  map[uint]*domain.Contact
nextID uint
}

func newFakeContactRepository() *fakeContactRepository {
return &fakeContactRepository{
store:  make(map[uint]*domain.Contact),
nextID: 1,
}
}

func (r *fakeContactRepository) Create(entity *domain.Contact) error {
if entity.ID == 0 {
entity.ID = r.nextID
r.nextID++
}
r.store[entity.ID] = entity
return nil
}

func (r *fakeContactRepository) GetByID(id uint) (*domain.Contact, error) {
return r.store[id], nil
}

func (r *fakeContactRepository) Update(entity *domain.Contact) error {
r.store[entity.ID] = entity
return nil
}

func (r *fakeContactRepository) Delete(id uint) error {
delete(r.store, id)
return nil
}

func (r *fakeContactRepository) List() ([]*domain.Contact, error) {
items := make([]*domain.Contact, 0, len(r.store))
for _, item := range r.store {
items = append(items, item)
}
return items, nil
}

func TestContactServiceCRUD(t *testing.T) {
repo := newFakeContactRepository()
service := NewContactService(repo)
entity := &domain.Contact{}
if err := service.CreateContact(entity); err != nil {
t.Fatalf("create failed: %v", err)
}
if entity.ID == 0 {
t.Fatal("expected ID to be set")
}
if _, err := service.GetContactByID(entity.ID); err != nil {
t.Fatalf("get failed: %v", err)
}
if err := service.UpdateContact(entity); err != nil {
t.Fatalf("update failed: %v", err)
}
items, err := service.ListContacts()
if err != nil {
t.Fatalf("list failed: %v", err)
}
if len(items) != 1 {
t.Fatalf("expected 1 item got %d", len(items))
}
if err := service.DeleteContact(entity.ID); err != nil {
t.Fatalf("delete failed: %v", err)
}
}
