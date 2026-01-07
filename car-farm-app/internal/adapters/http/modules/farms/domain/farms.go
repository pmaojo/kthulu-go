// @kthulu:domain:farms
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

// Farms represents a farms entity
type Farms struct {
        ID        uint      `json:"id" gorm:"primaryKey"`
        CreatedAt time.Time `json:"created_at"`
        UpdatedAt time.Time `json:"updated_at"`

        Name string `json:"name" gorm:"name"`
        Location string `json:"location" gorm:"location"`
        Capacity int `json:"capacity" gorm:"capacity"`

}

// TableName overrides the table name used by User to `Farms`
func (Farms) TableName() string {
	return "Farms"
}

// FarmsRepository defines the repository interface
type FarmsRepository interface {
        Create(entity *Farms) error
        GetByID(id uint) (*Farms, error)
        Update(entity *Farms) error
        Delete(id uint) error
        List(filter SearchFilter) ([]*Farms, error)
}

// FarmsService defines the service interface
type FarmsService interface {
        CreateFarms(entity *Farms) error
        GetFarmsByID(id uint) (*Farms, error)
        UpdateFarms(entity *Farms) error
        DeleteFarms(id uint) error
        ListFarmss(filter SearchFilter) ([]*Farms, error)
}
