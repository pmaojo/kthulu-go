// @kthulu:core:storage
package core

import (
	"time"
	
)

// SearchFilter represents search criteria
type SearchFilter struct {
	Query string
	Limit int
	Offset int
}

// Storage represents a storage entity
type Storage struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	Name string `json:"name" gorm:"name"`
}

// TableName overrides the table name used by Storage to `Storages`
func (Storage) TableName() string {
	return "Storages"
}

// StorageRepository defines the repository interface
type StorageRepository interface {
	Create(entity *Storage) error
	GetByID(id uint) (*Storage, error)
	Update(entity *Storage) error
	Delete(id uint) error
	List(filter SearchFilter) ([]*Storage, error)
}

// StorageService defines the service interface  
type StorageService interface {
	CreateStorage(entity *Storage) error
	GetStorageByID(id uint) (*Storage, error)
	UpdateStorage(entity *Storage) error
	DeleteStorage(id uint) error
	ListStorages(filter SearchFilter) ([]*Storage, error)
}
