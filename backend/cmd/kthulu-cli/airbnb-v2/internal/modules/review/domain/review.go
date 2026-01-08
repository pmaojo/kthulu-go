// @kthulu:domain:review
package domain

import (
	"time"
	
)

// SearchFilter represents search criteria
type SearchFilter struct {
	Query string
	Limit int
	Offset int
}

// Review represents a review entity
type Review struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	BookingID string `json:"booking_i_d" gorm:"booking_i_d"`
	Rating int `json:"rating" gorm:"rating"`
	Comment string `json:"comment" gorm:"comment"`
}

// TableName overrides the table name used by Review to `Reviews`
func (Review) TableName() string {
	return "Reviews"
}

// ReviewRepository defines the repository interface
type ReviewRepository interface {
	Create(entity *Review) error
	GetByID(id uint) (*Review, error)
	Update(entity *Review) error
	Delete(id uint) error
	List(filter SearchFilter) ([]*Review, error)
}

// ReviewService defines the service interface  
type ReviewService interface {
	CreateReview(entity *Review) error
	GetReviewByID(id uint) (*Review, error)
	UpdateReview(entity *Review) error
	DeleteReview(id uint) error
	ListReviews(filter SearchFilter) ([]*Review, error)
}
