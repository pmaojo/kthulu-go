// @kthulu:test:service:organization
package service

import (
"testing"

"my-kthulu-app/internal/adapters/http/modules/organization/domain"
)

type fakeOrganizationRepository struct {
store  map[uint]*domain.Organization
nextID uint
}

func newFakeOrganizationRepository() *fakeOrganizationRepository {
return &fakeOrganizationRepository{
store:  make(map[uint]*domain.Organization),
nextID: 1,
}
}

func (r *fakeOrganizationRepository) Create(entity *domain.Organization) error {
if entity.ID == 0 {
entity.ID = r.nextID
r.nextID++
}
r.store[entity.ID] = entity
return nil
}

func (r *fakeOrganizationRepository) GetByID(id uint) (*domain.Organization, error) {
return r.store[id], nil
}

func (r *fakeOrganizationRepository) Update(entity *domain.Organization) error {
r.store[entity.ID] = entity
return nil
}

func (r *fakeOrganizationRepository) Delete(id uint) error {
delete(r.store, id)
return nil
}

func (r *fakeOrganizationRepository) List() ([]*domain.Organization, error) {
items := make([]*domain.Organization, 0, len(r.store))
for _, item := range r.store {
items = append(items, item)
}
return items, nil
}

func TestOrganizationServiceCRUD(t *testing.T) {
repo := newFakeOrganizationRepository()
service := NewOrganizationService(repo)
entity := &domain.Organization{}
if err := service.CreateOrganization(entity); err != nil {
t.Fatalf("create failed: %v", err)
}
if entity.ID == 0 {
t.Fatal("expected ID to be set")
}
if _, err := service.GetOrganizationByID(entity.ID); err != nil {
t.Fatalf("get failed: %v", err)
}
if err := service.UpdateOrganization(entity); err != nil {
t.Fatalf("update failed: %v", err)
}
items, err := service.ListOrganizations()
if err != nil {
t.Fatalf("list failed: %v", err)
}
if len(items) != 1 {
t.Fatalf("expected 1 item got %d", len(items))
}
if err := service.DeleteOrganization(entity.ID); err != nil {
t.Fatalf("delete failed: %v", err)
}
}
