// @kthulu:store:user
package store

import (
	"gorm.io/gorm"
	"feature-tour/internal/modules/user/core"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) core.UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(entity *core.User) error {
	return r.db.Create(entity).Error
}

func (r *UserRepository) GetByID(id uint) (*core.User, error) {
	var entity core.User
	err := r.db.First(&entity, id).Error
	return &entity, err
}

func (r *UserRepository) Update(entity *core.User) error {
	return r.db.Save(entity).Error
}

func (r *UserRepository) Delete(id uint) error {
	return r.db.Delete(&core.User{}, id).Error
}

func (r *UserRepository) List(filter core.SearchFilter) ([]*core.User, error) {
	var entities []*core.User
	query := r.db.Model(&core.User{})

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
