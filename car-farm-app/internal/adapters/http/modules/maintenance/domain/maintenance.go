// @kthulu:domain:maintenance
package domain

import (
	"time"

	carsDomain "car-farm-app/internal/adapters/http/modules/cars/domain"

)

// SearchFilter represents search criteria
type SearchFilter struct {
	Query string
	Limit int
	Offset int
}

// Maintenance represents a maintenance entity
type Maintenance struct {
        ID        uint      `json:"id" gorm:"primaryKey"`
        CreatedAt time.Time `json:"created_at"`
        UpdatedAt time.Time `json:"updated_at"`

        Description string `json:"description" gorm:"description"`
        Cost float64 `json:"cost" gorm:"cost"`
        Date time.Time `json:"date" gorm:"date"`
        CarID uint `json:"car_id" gorm:"car_id"`
        Car *carsDomain.Cars `json:"car,omitempty" gorm:"foreignKey:CarID"`

}

// TableName overrides the table name used by User to `Maintenances`
func (Maintenance) TableName() string {
	return "Maintenances"
}

// MaintenanceRepository defines the repository interface
type MaintenanceRepository interface {
        Create(entity *Maintenance) error
        GetByID(id uint) (*Maintenance, error)
        Update(entity *Maintenance) error
        Delete(id uint) error
        List(filter SearchFilter) ([]*Maintenance, error)
}

// MaintenanceService defines the service interface
type MaintenanceService interface {
        CreateMaintenance(entity *Maintenance) error
        GetMaintenanceByID(id uint) (*Maintenance, error)
        UpdateMaintenance(entity *Maintenance) error
        DeleteMaintenance(id uint) error
        ListMaintenances(filter SearchFilter) ([]*Maintenance, error)
}
