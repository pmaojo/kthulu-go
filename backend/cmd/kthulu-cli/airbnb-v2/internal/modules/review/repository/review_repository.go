// @kthulu:repository:review
package repository

import (
	"gorm.io/gorm"
	"airbnb-v2/internal/modules/review/domain"
)

type ReviewRepository struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) domain.ReviewRepository {
	return &ReviewRepository{db: db}
}

func (r *ReviewRepository) Create(entity *domain.Review) error {
	return r.db.Create(entity).Error
}

func (r *ReviewRepository) GetByID(id uint) (*domain.Review, error) {
	var entity domain.Review
	err := r.db.First(&entity, id).Error
	return &entity, err
}

func (r *ReviewRepository) Update(entity *domain.Review) error {
	return r.db.Save(entity).Error
}

func (r *ReviewRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Review{}, id).Error
}

func (r *ReviewRepository) List(filter domain.SearchFilter) ([]*domain.Review, error) {
	var entities []*domain.Review
	query := r.db.Model(&domain.Review{})

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
