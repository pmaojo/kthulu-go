// @kthulu:service:reviews
package service

import (
	"test-app/internal/modules/reviews/domain"
)

type ReviewService struct {
	repo domain.ReviewRepository
}

func NewReviewService(repo domain.ReviewRepository) domain.ReviewService {
	return &ReviewService{repo: repo}
}

func (s *ReviewService) CreateReview(entity *domain.Review) error {
	// Add business logic here
	return s.repo.Create(entity)
}

func (s *ReviewService) GetReviewByID(id uint) (*domain.Review, error) {
	return s.repo.GetByID(id)
}

func (s *ReviewService) UpdateReview(entity *domain.Review) error {
	// Add business logic here
	return s.repo.Update(entity)
}

func (s *ReviewService) DeleteReview(id uint) error {
	// Add business logic here
	return s.repo.Delete(id)
}

func (s *ReviewService) ListReviews(filter domain.SearchFilter) ([]*domain.Review, error) {
	return s.repo.List(filter)
}
