// @kthulu:store:auth
package store

import (
	"gorm.io/gorm"
	"tournament-app/internal/modules/auth/core"
)

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) core.AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) Create(entity *core.Auth) error {
	return r.db.Create(entity).Error
}

func (r *AuthRepository) GetByID(id uint) (*core.Auth, error) {
	var entity core.Auth
	err := r.db.First(&entity, id).Error
	return &entity, err
}

func (r *AuthRepository) Update(entity *core.Auth) error {
	return r.db.Save(entity).Error
}

func (r *AuthRepository) Delete(id uint) error {
	return r.db.Delete(&core.Auth{}, id).Error
}

func (r *AuthRepository) List(filter core.SearchFilter) ([]*core.Auth, error) {
	var entities []*core.Auth
	query := r.db.Model(&core.Auth{})

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
