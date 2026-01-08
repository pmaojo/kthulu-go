// @kthulu:repository:auth
package repository

import (
	"gorm.io/gorm"
	"demo-app/internal/modules/auth/domain"
)

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) domain.AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) Create(entity *domain.Auth) error {
	return r.db.Create(entity).Error
}

func (r *AuthRepository) GetByID(id uint) (*domain.Auth, error) {
	var entity domain.Auth
	err := r.db.First(&entity, id).Error
	return &entity, err
}

func (r *AuthRepository) Update(entity *domain.Auth) error {
	return r.db.Save(entity).Error
}

func (r *AuthRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Auth{}, id).Error
}

func (r *AuthRepository) List(filter domain.SearchFilter) ([]*domain.Auth, error) {
	var entities []*domain.Auth
	query := r.db.Model(&domain.Auth{})

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
