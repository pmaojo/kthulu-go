// @kthulu:repository:users
package repository

import (
	"gorm.io/gorm"
	"github.com/example/workshop-api/internal/adapters/http/modules/users/domain"
)

type UsersRepository struct {
	db *gorm.DB
}

func NewUsersRepository(db *gorm.DB) domain.UsersRepository {
	return &UsersRepository{db: db}
}

func (r *UsersRepository) Create(entity *domain.Users) error {
	return r.db.Create(entity).Error
}

func (r *UsersRepository) GetByID(id uint) (*domain.Users, error) {
	var entity domain.Users
	err := r.db.First(&entity, id).Error
	return &entity, err
}

func (r *UsersRepository) Update(entity *domain.Users) error {
	return r.db.Save(entity).Error
}

func (r *UsersRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Users{}, id).Error
}

func (r *UsersRepository) List() ([]*domain.Users, error) {
	var entities []*domain.Users
	err := r.db.Find(&entities).Error
	return entities, err
}