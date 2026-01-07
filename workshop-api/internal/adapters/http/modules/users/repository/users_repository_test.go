// @kthulu:test:repository:users
package repository

import (
"testing"

"gorm.io/driver/sqlite"
"gorm.io/gorm"

"github.com/example/workshop-api/internal/adapters/http/modules/users/domain"
)

func TestUsersRepositoryCRUD(t *testing.T) {
db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
if err != nil {
t.Fatalf("failed to open sqlite: %v", err)
}
if err := db.AutoMigrate(&domain.Users{}); err != nil {
t.Fatalf("failed to migrate: %v", err)
}
repo := NewUsersRepository(db)
entity := &domain.Users{}
if err := repo.Create(entity); err != nil {
t.Fatalf("create failed: %v", err)
}
fetched, err := repo.GetByID(entity.ID)
if err != nil {
t.Fatalf("get failed: %v", err)
}
if fetched.ID != entity.ID {
t.Fatalf("expected ID %d got %d", entity.ID, fetched.ID)
}
if err := repo.Update(entity); err != nil {
t.Fatalf("update failed: %v", err)
}
items, err := repo.List()
if err != nil {
t.Fatalf("list failed: %v", err)
}
if len(items) != 1 {
t.Fatalf("expected 1 item got %d", len(items))
}
if err := repo.Delete(entity.ID); err != nil {
t.Fatalf("delete failed: %v", err)
}
}
