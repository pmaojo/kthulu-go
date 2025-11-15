// @kthulu:module:products
package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/kthulu/kthulu-go/backend/internal/domain"
	"github.com/kthulu/kthulu-go/backend/internal/repository"

	"go.uber.org/zap"
)

// ProductUseCase handles business logic for product management
type ProductUseCase struct {
	productRepo repository.ProductRepository
	logger      *zap.Logger
}

// NewProductUseCase creates a new product use case
func NewProductUseCase(
	productRepo repository.ProductRepository,
	logger *zap.Logger,
) *ProductUseCase {
	return &ProductUseCase{
		productRepo: productRepo,
		logger:      logger,
	}
}

// CreateProduct creates a new product
func (uc *ProductUseCase) CreateProduct(ctx context.Context, organizationID uint, req CreateProductRequest) (*domain.Product, error) {
	uc.logger.Info("Creating new product",
		zap.Uint("organization_id", organizationID),
		zap.String("sku", req.SKU),
		zap.String("name", req.Name),
	)

	// Check if product with SKU already exists
	existing, err := uc.productRepo.GetBySKU(ctx, organizationID, req.SKU)
	if err == nil && existing != nil {
		return nil, domain.ErrProductAlreadyExists
	}

	// Create new product
	product, err := domain.NewProduct(organizationID, req.SKU, req.Name, req.UnitOfMeasure)
	if err != nil {
		uc.logger.Error("Failed to create product domain object", zap.Error(err))
		return nil, err
	}

	// Update additional fields
	if err := product.UpdateBasicInfo(
		req.Name,
		req.Description,
		req.Category,
		req.Brand,
		req.UnitOfMeasure,
		req.Dimensions,
		req.Barcode,
		req.Weight,
		req.TaxRate,
	); err != nil {
		uc.logger.Error("Failed to update product basic info", zap.Error(err))
		return nil, err
	}

	if req.IsTrackable != nil {
		product.SetTrackable(*req.IsTrackable)
	}

	// Save to repository
	if err := uc.productRepo.Create(ctx, product); err != nil {
		uc.logger.Error("Failed to save product to repository", zap.Error(err))
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	uc.logger.Info("Product created successfully",
		zap.Uint("product_id", product.ID),
		zap.String("display_name", product.GetDisplayName()),
	)

	return product, nil
}

// GetProduct retrieves a product by ID
func (uc *ProductUseCase) GetProduct(ctx context.Context, organizationID, productID uint, includeRelated bool) (*domain.Product, error) {
	product, err := uc.productRepo.GetByID(ctx, organizationID, productID)
	if err != nil {
		uc.logger.Error("Failed to get product",
			zap.Uint("organization_id", organizationID),
			zap.Uint("product_id", productID),
			zap.Error(err),
		)
		return nil, err
	}

	// Load related data if requested
	if includeRelated {
		if err := uc.loadProductRelations(ctx, product); err != nil {
			uc.logger.Warn("Failed to load product relations", zap.Error(err))
			// Don't fail the request, just log the warning
		}
	}

	return product, nil
}

// UpdateProduct updates an existing product
func (uc *ProductUseCase) UpdateProduct(ctx context.Context, organizationID, productID uint, req UpdateProductRequest) (*domain.Product, error) {
	uc.logger.Info("Updating product",
		zap.Uint("organization_id", organizationID),
		zap.Uint("product_id", productID),
	)

	// Get existing product
	product, err := uc.productRepo.GetByID(ctx, organizationID, productID)
	if err != nil {
		return nil, err
	}

	// Update product information
	if err := product.UpdateBasicInfo(
		req.Name,
		req.Description,
		req.Category,
		req.Brand,
		req.UnitOfMeasure,
		req.Dimensions,
		req.Barcode,
		req.Weight,
		req.TaxRate,
	); err != nil {
		uc.logger.Error("Failed to update product basic info", zap.Error(err))
		return nil, err
	}

	if req.IsTrackable != nil {
		product.SetTrackable(*req.IsTrackable)
	}

	// Save changes
	if err := uc.productRepo.Update(ctx, product); err != nil {
		uc.logger.Error("Failed to update product in repository", zap.Error(err))
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	uc.logger.Info("Product updated successfully", zap.Uint("product_id", productID))
	return product, nil
}

// DeleteProduct deletes a product
func (uc *ProductUseCase) DeleteProduct(ctx context.Context, organizationID, productID uint) error {
	uc.logger.Info("Deleting product",
		zap.Uint("organization_id", organizationID),
		zap.Uint("product_id", productID),
	)

	if err := uc.productRepo.Delete(ctx, organizationID, productID); err != nil {
		uc.logger.Error("Failed to delete product", zap.Error(err))
		return fmt.Errorf("failed to delete product: %w", err)
	}

	uc.logger.Info("Product deleted successfully", zap.Uint("product_id", productID))
	return nil
}

// ListProducts retrieves a list of products with filtering and pagination
func (uc *ProductUseCase) ListProducts(ctx context.Context, organizationID uint, filters repository.ProductFilters) (*ProductListResponse, error) {
	// Validate and set defaults for filters
	if err := filters.Validate(); err != nil {
		return nil, err
	}

	products, total, err := uc.productRepo.List(ctx, organizationID, filters)
	if err != nil {
		uc.logger.Error("Failed to list products", zap.Error(err))
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	// Load related data if requested
	if filters.IncludeVariants || filters.IncludePrices {
		for _, product := range products {
			if err := uc.loadProductRelations(ctx, product); err != nil {
				uc.logger.Warn("Failed to load product relations",
					zap.Uint("product_id", product.ID),
					zap.Error(err),
				)
			}
		}
	}

	return &ProductListResponse{
		Products:   products,
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		TotalPages: (total + int64(filters.PageSize) - 1) / int64(filters.PageSize),
	}, nil
}

// SetProductActive sets the active status of a product
func (uc *ProductUseCase) SetProductActive(ctx context.Context, organizationID, productID uint, active bool) error {
	uc.logger.Info("Setting product active status",
		zap.Uint("organization_id", organizationID),
		zap.Uint("product_id", productID),
		zap.Bool("active", active),
	)

	product, err := uc.productRepo.GetByID(ctx, organizationID, productID)
	if err != nil {
		return err
	}

	product.SetActive(active)

	if err := uc.productRepo.Update(ctx, product); err != nil {
		uc.logger.Error("Failed to update product active status", zap.Error(err))
		return fmt.Errorf("failed to update product status: %w", err)
	}

	uc.logger.Info("Product active status updated successfully", zap.Uint("product_id", productID))
	return nil
}

// GetProductStats retrieves product statistics for an organization
func (uc *ProductUseCase) GetProductStats(ctx context.Context, organizationID uint) (*repository.ProductStats, error) {
	stats, err := uc.productRepo.GetProductStats(ctx, organizationID)
	if err != nil {
		uc.logger.Error("Failed to get product stats", zap.Error(err))
		return nil, fmt.Errorf("failed to get product stats: %w", err)
	}

	return stats, nil
}

// CreateProductVariant creates a new product variant
func (uc *ProductUseCase) CreateProductVariant(ctx context.Context, organizationID, productID uint, req CreateVariantRequest) (*domain.ProductVariant, error) {
	uc.logger.Info("Creating product variant",
		zap.Uint("organization_id", organizationID),
		zap.Uint("product_id", productID),
		zap.String("sku", req.SKU),
	)

	// Verify product exists and belongs to organization
	_, err := uc.productRepo.GetByID(ctx, organizationID, productID)
	if err != nil {
		return nil, err
	}

	// Check if variant with SKU already exists
	existing, err := uc.productRepo.GetVariantBySKU(ctx, req.SKU)
	if err == nil && existing != nil {
		return nil, domain.ErrVariantAlreadyExists
	}

	// Create new variant
	variant, err := domain.NewProductVariant(productID, req.SKU, req.Name, req.Attributes)
	if err != nil {
		return nil, err
	}

	// Update additional fields
	if err := variant.UpdateBasicInfo(
		req.Name,
		req.Description,
		req.Dimensions,
		req.Barcode,
		req.Weight,
		req.Attributes,
	); err != nil {
		return nil, err
	}

	if err := uc.productRepo.CreateVariant(ctx, variant); err != nil {
		uc.logger.Error("Failed to create product variant", zap.Error(err))
		return nil, fmt.Errorf("failed to create variant: %w", err)
	}

	uc.logger.Info("Product variant created successfully", zap.Uint("variant_id", variant.ID))
	return variant, nil
}

// CreateProductPrice creates a new product price
func (uc *ProductUseCase) CreateProductPrice(ctx context.Context, organizationID uint, req CreatePriceRequest) (*domain.ProductPrice, error) {
	uc.logger.Info("Creating product price",
		zap.Uint("organization_id", organizationID),
		zap.String("price_type", string(req.PriceType)),
		zap.Float64("amount", req.Amount),
	)

	// Verify product or variant exists and belongs to organization
	if req.ProductID != nil {
		_, err := uc.productRepo.GetByID(ctx, organizationID, *req.ProductID)
		if err != nil {
			return nil, err
		}
	}

	if req.ProductVariantID != nil {
		// For variants, we need to check through the product
		variant, err := uc.productRepo.GetVariantByID(ctx, 0, *req.ProductVariantID) // productID=0 as placeholder
		if err != nil {
			return nil, err
		}
		// Verify the variant's product belongs to the organization
		_, err = uc.productRepo.GetByID(ctx, organizationID, variant.ProductID)
		if err != nil {
			return nil, err
		}
	}

	// Create new price
	price, err := domain.NewProductPrice(
		req.ProductID,
		req.ProductVariantID,
		req.PriceType,
		req.Currency,
		req.Amount,
		req.MinQuantity,
	)
	if err != nil {
		return nil, err
	}

	// Update additional fields
	if err := price.UpdatePrice(
		req.Amount,
		req.MinQuantity,
		req.MaxQuantity,
		req.ValidFrom,
		req.ValidUntil,
	); err != nil {
		return nil, err
	}

	if err := uc.productRepo.CreatePrice(ctx, price); err != nil {
		uc.logger.Error("Failed to create product price", zap.Error(err))
		return nil, fmt.Errorf("failed to create price: %w", err)
	}

	uc.logger.Info("Product price created successfully", zap.Uint("price_id", price.ID))
	return price, nil
}

// GetEffectivePrice gets the effective price for a product or variant
func (uc *ProductUseCase) GetEffectivePrice(ctx context.Context, organizationID uint, req GetEffectivePriceRequest) (*domain.ProductPrice, error) {
	// Verify product or variant exists and belongs to organization
	if req.ProductID != nil {
		_, err := uc.productRepo.GetByID(ctx, organizationID, *req.ProductID)
		if err != nil {
			return nil, err
		}
	}

	if req.ProductVariantID != nil {
		variant, err := uc.productRepo.GetVariantByID(ctx, 0, *req.ProductVariantID)
		if err != nil {
			return nil, err
		}
		_, err = uc.productRepo.GetByID(ctx, organizationID, variant.ProductID)
		if err != nil {
			return nil, err
		}
	}

	at := time.Now()
	if req.At != nil {
		at = *req.At
	}

	price, err := uc.productRepo.GetEffectivePrice(
		ctx,
		req.ProductID,
		req.ProductVariantID,
		req.PriceType,
		req.Quantity,
		at,
	)
	if err != nil {
		return nil, err
	}

	return price, nil
}

// loadProductRelations loads variants and prices for a product
func (uc *ProductUseCase) loadProductRelations(ctx context.Context, product *domain.Product) error {
	// Load variants
	variants, err := uc.productRepo.GetVariantsByProductID(ctx, product.ID)
	if err != nil {
		return err
	}
	// Convert []*domain.ProductVariant to []domain.ProductVariant
	product.Variants = make([]domain.ProductVariant, len(variants))
	for i, variant := range variants {
		product.Variants[i] = *variant

		// Load prices for each variant
		variantPrices, err := uc.productRepo.GetPricesByVariantID(ctx, variant.ID)
		if err == nil {
			product.Variants[i].Prices = make([]domain.ProductPrice, len(variantPrices))
			for j, price := range variantPrices {
				product.Variants[i].Prices[j] = *price
			}
		}
	}

	// Load prices for the product itself
	prices, err := uc.productRepo.GetPricesByProductID(ctx, product.ID)
	if err != nil {
		return err
	}
	// Convert []*domain.ProductPrice to []domain.ProductPrice
	product.Prices = make([]domain.ProductPrice, len(prices))
	for i, price := range prices {
		product.Prices[i] = *price
	}

	return nil
}

// Request/Response DTOs

// CreateProductRequest represents a request to create a new product
type CreateProductRequest struct {
	SKU           string   `json:"sku" validate:"required,min=1,max=100"`
	Name          string   `json:"name" validate:"required,min=1,max=200"`
	Description   string   `json:"description,omitempty"`
	Category      string   `json:"category,omitempty" validate:"max=100"`
	Brand         string   `json:"brand,omitempty" validate:"max=100"`
	UnitOfMeasure string   `json:"unitOfMeasure" validate:"required,max=20"`
	Weight        *float64 `json:"weight,omitempty" validate:"omitempty,min=0"`
	Dimensions    string   `json:"dimensions,omitempty" validate:"max=100"`
	Barcode       string   `json:"barcode,omitempty" validate:"max=100"`
	TaxRate       float64  `json:"taxRate" validate:"min=0,max=1"`
	IsTrackable   *bool    `json:"isTrackable,omitempty"`
}

// UpdateProductRequest represents a request to update a product
type UpdateProductRequest struct {
	Name          string   `json:"name" validate:"required,min=1,max=200"`
	Description   string   `json:"description,omitempty"`
	Category      string   `json:"category,omitempty" validate:"max=100"`
	Brand         string   `json:"brand,omitempty" validate:"max=100"`
	UnitOfMeasure string   `json:"unitOfMeasure" validate:"required,max=20"`
	Weight        *float64 `json:"weight,omitempty" validate:"omitempty,min=0"`
	Dimensions    string   `json:"dimensions,omitempty" validate:"max=100"`
	Barcode       string   `json:"barcode,omitempty" validate:"max=100"`
	TaxRate       float64  `json:"taxRate" validate:"min=0,max=1"`
	IsTrackable   *bool    `json:"isTrackable,omitempty"`
}

// CreateVariantRequest represents a request to create a product variant
type CreateVariantRequest struct {
	SKU         string                 `json:"sku" validate:"required,min=1,max=100"`
	Name        string                 `json:"name" validate:"required,min=1,max=200"`
	Description string                 `json:"description,omitempty"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
	Weight      *float64               `json:"weight,omitempty" validate:"omitempty,min=0"`
	Dimensions  string                 `json:"dimensions,omitempty" validate:"max=100"`
	Barcode     string                 `json:"barcode,omitempty" validate:"max=100"`
}

// CreatePriceRequest represents a request to create a product price
type CreatePriceRequest struct {
	ProductID        *uint            `json:"productId,omitempty"`
	ProductVariantID *uint            `json:"productVariantId,omitempty"`
	PriceType        domain.PriceType `json:"priceType" validate:"required,oneof=base sale wholesale retail cost"`
	Currency         string           `json:"currency" validate:"required,len=3"`
	Amount           float64          `json:"amount" validate:"required,min=0"`
	MinQuantity      int              `json:"minQuantity" validate:"min=1"`
	MaxQuantity      *int             `json:"maxQuantity,omitempty" validate:"omitempty,min=1"`
	ValidFrom        *time.Time       `json:"validFrom,omitempty"`
	ValidUntil       *time.Time       `json:"validUntil,omitempty"`
}

// GetEffectivePriceRequest represents a request to get effective price
type GetEffectivePriceRequest struct {
	ProductID        *uint            `json:"productId,omitempty"`
	ProductVariantID *uint            `json:"productVariantId,omitempty"`
	PriceType        domain.PriceType `json:"priceType" validate:"required,oneof=base sale wholesale retail cost"`
	Quantity         int              `json:"quantity" validate:"min=1"`
	At               *time.Time       `json:"at,omitempty"`
}

// ProductListResponse represents a paginated list of products
type ProductListResponse struct {
	Products   []*domain.Product `json:"products"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"pageSize"`
	TotalPages int64             `json:"totalPages"`
}
