// @kthulu:core:cache
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

// Cache represents a cache entity
type Cache struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	Name string `json:"name" gorm:"name"`
}

// TableName overrides the table name used by Cache to `Caches`
func (Cache) TableName() string {
	return "Caches"
}

// CacheRepository defines the repository interface
type CacheRepository interface {
	Create(entity *Cache) error
	GetByID(id uint) (*Cache, error)
	Update(entity *Cache) error
	Delete(id uint) error
	List(filter SearchFilter) ([]*Cache, error)
}

// CacheService defines the service interface  
type CacheService interface {
	CreateCache(entity *Cache) error
	GetCacheByID(id uint) (*Cache, error)
	UpdateCache(entity *Cache) error
	DeleteCache(id uint) error
	ListCaches(filter SearchFilter) ([]*Cache, error)
}
