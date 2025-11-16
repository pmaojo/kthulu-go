// @kthulu:repository:user
package repository

import (
	"gorm.io/gorm"
	"my-kthulu-app/internal/adapters/http/modules/user/domain"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(entity *domain.User) error {
	return r.db.Create(entity).Error
}

func (r *UserRepository) GetByID(id uint) (*domain.User, error) {
	var entity domain.User
	err := r.db.First(&entity, id).Error
	return &entity, err
}

func (r *UserRepository) Update(entity *domain.User) error {
	return r.db.Save(entity).Error
}

func (r *UserRepository) Delete(id uint) error {
	return r.db.Delete(&domain.User{}, id).Error
}

func (r *UserRepository) List() ([]*domain.User, error) {
	var entities []*domain.User
	err := r.db.Find(&entities).Error
	return entities, err
}