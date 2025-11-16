// @kthulu:repository:auth
package repository

import (
	"gorm.io/gorm"
	"my-kthulu-app/internal/adapters/http/modules/auth/domain"
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

func (r *AuthRepository) List() ([]*domain.Auth, error) {
	var entities []*domain.Auth
	err := r.db.Find(&entities).Error
	return entities, err
}