// @kthulu:domain:services
package domain

import "time"

// SearchFilter represents search criteria
type SearchFilter struct {
	Query string
	Limit int
	Offset int
}

// Services represents a services entity
type Services struct {
        ID        uint      `json:"id" gorm:"primaryKey"`
        CreatedAt time.Time `json:"created_at"`
        UpdatedAt time.Time `json:"updated_at"`

        Name string `json:"name" gorm:"name"`
        Description string `json:"description" gorm:"description"`
        Price float64 `json:"price" gorm:"price"`
        Duration int `json:"duration" gorm:"duration"`

}

// TableName overrides the table name used by User to `servicess`
func (Services) TableName() string {
	return "servicess"
}

// ServicesRepository defines the repository interface
type ServicesRepository interface {
        Create(entity *Services) error
        GetByID(id uint) (*Services, error)
        Update(entity *Services) error
        Delete(id uint) error
        List(filter SearchFilter) ([]*Services, error)
}

// ServicesService defines the service interface
type ServicesService interface {
        CreateServices(entity *Services) error
        GetServicesByID(id uint) (*Services, error)
        UpdateServices(entity *Services) error
        DeleteServices(id uint) error
        ListServicess(filter SearchFilter) ([]*Services, error)
}
