// @kthulu:repository:user
package repository

import (
	"gorm.io/gorm"
	"test-app/internal/modules/user/domain"
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

func (r *UserRepository) List(filter domain.SearchFilter) ([]*domain.User, error) {
	var entities []*domain.User
	query := r.db.Model(&domain.User{})

	if filter.Query != "" {
		// Basic search implementation
		// query = query.Where("name LIKE ?", "%"+filter.Query+"%")
	}

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	err := query.Find(&entities).Error
	return entities, err
}
