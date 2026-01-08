// @kthulu:domain:cars
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

// Cars represents a cars entity
type Cars struct {
        ID        uint      `json:"id" gorm:"primaryKey"`
        CreatedAt time.Time `json:"created_at"`
        UpdatedAt time.Time `json:"updated_at"`

        Make string `json:"make" gorm:"make"`
        Model string `json:"model" gorm:"model"`
        Year int `json:"year" gorm:"year"`
        Vin string `json:"vin" gorm:"vin"`
        Status string `json:"status" gorm:"status"`

}

// TableName overrides the table name used by User to `Cars`
func (Cars) TableName() string {
	return "Cars"
}

// CarsRepository defines the repository interface
type CarsRepository interface {
        Create(entity *Cars) error
        GetByID(id uint) (*Cars, error)
        Update(entity *Cars) error
        Delete(id uint) error
        List(filter SearchFilter) ([]*Cars, error)
}

// CarsService defines the service interface
type CarsService interface {
        CreateCars(entity *Cars) error
        GetCarsByID(id uint) (*Cars, error)
        UpdateCars(entity *Cars) error
        DeleteCars(id uint) error
        ListCarss(filter SearchFilter) ([]*Cars, error)
}
