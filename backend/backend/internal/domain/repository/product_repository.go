// @kthulu:module:products
package repository

import (
	"context"
	"time"

	"github.com/pmaojo/kthulu-go/backend/internal/domain"
)

// ProductRepository defines the interface for product data operations
type ProductRepository interface {
	// Product operations
	Create(ctx context.Context, product *domain.Product) error
	GetByID(ctx context.Context, organizationID, productID uint) (*domain.Product, error)
	GetBySKU(ctx context.Context, organizationID uint, sku string) (*domain.Product, error)
	Update(ctx context.Context, product *domain.Product) error
	Delete(ctx context.Context, organizationID, productID uint) error
	List(ctx context.Context, organizationID uint, filters ProductFilters) ([]*domain.Product, int64, error)
	ListPaginated(ctx context.Context, organizationID uint, params PaginationParams) (PaginationResult[*domain.Product], error)
	SearchPaginated(ctx context.Context, organizationID uint, query string, params PaginationParams) (PaginationResult[*domain.Product], error)

	// Variant operations
	CreateVariant(ctx context.Context, variant *domain.ProductVariant) error
	GetVariantByID(ctx context.Context, productID, variantID uint) (*domain.ProductVariant, error)
	GetVariantBySKU(ctx context.Context, sku string) (*domain.ProductVariant, error)
	GetVariantsByProductID(ctx context.Context, productID uint) ([]*domain.ProductVariant, error)
	UpdateVariant(ctx context.Context, variant *domain.ProductVariant) error
	DeleteVariant(ctx context.Context, productID, variantID uint) error

	// Price operations
	CreatePrice(ctx context.Context, price *domain.ProductPrice) error
	GetPriceByID(ctx context.Context, priceID uint) (*domain.ProductPrice, error)
	GetPricesByProductID(ctx context.Context, productID uint) ([]*domain.ProductPrice, error)
	GetPricesByVariantID(ctx context.Context, variantID uint) ([]*domain.ProductPrice, error)
	GetEffectivePrice(ctx context.Context, productID *uint, variantID *uint, priceType domain.PriceType, quantity int, at time.Time) (*domain.ProductPrice, error)
	UpdatePrice(ctx context.Context, price *domain.ProductPrice) error
	DeletePrice(ctx context.Context, priceID uint) error

	// Bulk operations
	BulkCreate(ctx context.Context, products []*domain.Product) error
	BulkUpdate(ctx context.Context, products []*domain.Product) error
	BulkDelete(ctx context.Context, organizationID uint, productIDs []uint) error

	// Statistics and analytics
	GetProductStats(ctx context.Context, organizationID uint) (*ProductStats, error)
	GetCategoriesWithCounts(ctx context.Context, organizationID uint) ([]CategoryCount, error)
	GetBrandsWithCounts(ctx context.Context, organizationID uint) ([]BrandCount, error)
}

// ProductFilters represents filters for product listing
type ProductFilters struct {
	Category    string  `json:"category,omitempty"`
	Brand       string  `json:"brand,omitempty"`
	IsActive    *bool   `json:"isActive,omitempty"`
	IsTrackable *bool   `json:"isTrackable,omitempty"`
	Search      string  `json:"search,omitempty"`      // Search in name, SKU, description
	CreatedFrom *string `json:"createdFrom,omitempty"` // ISO date string
	CreatedTo   *string `json:"createdTo,omitempty"`   // ISO date string

	// Price filtering
	MinPrice *float64 `json:"minPrice,omitempty"`
	MaxPrice *float64 `json:"maxPrice,omitempty"`
	Currency string   `json:"currency,omitempty"`

	// Pagination
	Page     int `json:"page" validate:"min=1"`
	PageSize int `json:"pageSize" validate:"min=1,max=100"`

	// Sorting
	SortBy    string `json:"sortBy,omitempty"`    // name, sku, category, brand, created_at, updated_at
	SortOrder string `json:"sortOrder,omitempty"` // asc, desc

	// Include related data
	IncludeVariants bool `json:"includeVariants,omitempty"`
	IncludePrices   bool `json:"includePrices,omitempty"`
}

// ProductStats represents product statistics for an organization
type ProductStats struct {
	TotalProducts     int64   `json:"totalProducts"`
	ActiveProducts    int64   `json:"activeProducts"`
	InactiveProducts  int64   `json:"inactiveProducts"`
	TrackableProducts int64   `json:"trackableProducts"`
	TotalVariants     int64   `json:"totalVariants"`
	TotalCategories   int64   `json:"totalCategories"`
	TotalBrands       int64   `json:"totalBrands"`
	RecentProducts    int64   `json:"recentProducts"` // Products created in last 30 days
	AveragePrice      float64 `json:"averagePrice"`
	HighestPrice      float64 `json:"highestPrice"`
	LowestPrice       float64 `json:"lowestPrice"`
}

// CategoryCount represents a category with its product count
type CategoryCount struct {
	Category string `json:"category"`
	Count    int64  `json:"count"`
}

// BrandCount represents a brand with its product count
type BrandCount struct {
	Brand string `json:"brand"`
	Count int64  `json:"count"`
}

// DefaultProductFilters returns default filters for product listing
func DefaultProductFilters() ProductFilters {
	active := true
	return ProductFilters{
		IsActive:  &active,
		Page:      1,
		PageSize:  20,
		SortBy:    "created_at",
		SortOrder: "desc",
		Currency:  "USD",
	}
}

// Validate validates the product filters
func (f *ProductFilters) Validate() error {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 100 {
		f.PageSize = 20
	}
	if f.SortBy == "" {
		f.SortBy = "created_at"
	}
	if f.SortOrder != "asc" && f.SortOrder != "desc" {
		f.SortOrder = "desc"
	}
	if f.Currency == "" {
		f.Currency = "USD"
	}
	return nil
}

// GetOffset returns the offset for pagination
func (f *ProductFilters) GetOffset() int {
	return (f.Page - 1) * f.PageSize
}
