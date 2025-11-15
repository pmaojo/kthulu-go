// @kthulu:module:products
package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pmaojo/kthulu-go/backend/core"
	"github.com/pmaojo/kthulu-go/backend/internal/domain"
	"github.com/pmaojo/kthulu-go/backend/internal/repository"
)

// ProductRepository implements the product repository interface using GORM
type ProductRepository struct {
	db     *sql.DB
	logger core.Logger
}

// NewProductRepository creates a new product repository instance
func NewProductRepository(db *sql.DB, logger core.Logger) repository.ProductRepository {
	return &ProductRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new product
func (r *ProductRepository) Create(ctx context.Context, product *domain.Product) error {
	query := `
		INSERT INTO products (
			organization_id, sku, name, description, category, brand, 
			unit_of_measure, weight, dimensions, barcode, tax_rate, 
			is_active, is_trackable, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		) RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		product.OrganizationID, product.SKU, product.Name, product.Description,
		product.Category, product.Brand, product.UnitOfMeasure, product.Weight,
		product.Dimensions, product.Barcode, product.TaxRate, product.IsActive,
		product.IsTrackable, product.CreatedAt, product.UpdatedAt,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return domain.ErrProductAlreadyExists
		}
		r.logger.Error("Failed to create product", "error", err, "sku", product.SKU)
		return fmt.Errorf("failed to create product: %w", err)
	}

	r.logger.Info("Product created successfully", "productId", product.ID, "sku", product.SKU)
	return nil
}

// GetByID retrieves a product by ID within an organization
func (r *ProductRepository) GetByID(ctx context.Context, organizationID, productID uint) (*domain.Product, error) {
	query := `
		SELECT id, organization_id, sku, name, description, category, brand,
			   unit_of_measure, weight, dimensions, barcode, tax_rate,
			   is_active, is_trackable, created_at, updated_at
		FROM products 
		WHERE id = $1 AND organization_id = $2`

	product := &domain.Product{}
	err := r.db.QueryRowContext(ctx, query, productID, organizationID).Scan(
		&product.ID, &product.OrganizationID, &product.SKU, &product.Name,
		&product.Description, &product.Category, &product.Brand,
		&product.UnitOfMeasure, &product.Weight, &product.Dimensions,
		&product.Barcode, &product.TaxRate, &product.IsActive,
		&product.IsTrackable, &product.CreatedAt, &product.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrProductNotFound
		}
		r.logger.Error("Failed to get product by ID", "error", err, "productId", productID)
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return product, nil
}

// GetBySKU retrieves a product by SKU within an organization
func (r *ProductRepository) GetBySKU(ctx context.Context, organizationID uint, sku string) (*domain.Product, error) {
	query := `
		SELECT id, organization_id, sku, name, description, category, brand,
			   unit_of_measure, weight, dimensions, barcode, tax_rate,
			   is_active, is_trackable, created_at, updated_at
		FROM products 
		WHERE sku = $1 AND organization_id = $2`

	product := &domain.Product{}
	err := r.db.QueryRowContext(ctx, query, sku, organizationID).Scan(
		&product.ID, &product.OrganizationID, &product.SKU, &product.Name,
		&product.Description, &product.Category, &product.Brand,
		&product.UnitOfMeasure, &product.Weight, &product.Dimensions,
		&product.Barcode, &product.TaxRate, &product.IsActive,
		&product.IsTrackable, &product.CreatedAt, &product.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrProductNotFound
		}
		r.logger.Error("Failed to get product by SKU", "error", err, "sku", sku)
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return product, nil
}

// Update updates an existing product
func (r *ProductRepository) Update(ctx context.Context, product *domain.Product) error {
	query := `
		UPDATE products SET 
			name = $2, description = $3, category = $4, brand = $5,
			unit_of_measure = $6, weight = $7, dimensions = $8, barcode = $9,
			tax_rate = $10, is_active = $11, is_trackable = $12, updated_at = $13
		WHERE id = $1 AND organization_id = $14`

	result, err := r.db.ExecContext(ctx, query,
		product.ID, product.Name, product.Description, product.Category,
		product.Brand, product.UnitOfMeasure, product.Weight, product.Dimensions,
		product.Barcode, product.TaxRate, product.IsActive, product.IsTrackable,
		time.Now(), product.OrganizationID,
	)

	if err != nil {
		r.logger.Error("Failed to update product", "error", err, "productId", product.ID)
		return fmt.Errorf("failed to update product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrProductNotFound
	}

	r.logger.Info("Product updated successfully", "productId", product.ID)
	return nil
}

// Delete deletes a product
func (r *ProductRepository) Delete(ctx context.Context, organizationID, productID uint) error {
	query := `DELETE FROM products WHERE id = $1 AND organization_id = $2`

	result, err := r.db.ExecContext(ctx, query, productID, organizationID)
	if err != nil {
		r.logger.Error("Failed to delete product", "error", err, "productId", productID)
		return fmt.Errorf("failed to delete product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrProductNotFound
	}

	r.logger.Info("Product deleted successfully", "productId", productID)
	return nil
}

// List retrieves products with filtering and pagination
func (r *ProductRepository) List(ctx context.Context, organizationID uint, filters repository.ProductFilters) ([]*domain.Product, int64, error) {
	// Validate filters
	if err := filters.Validate(); err != nil {
		return nil, 0, err
	}

	// Build WHERE clause
	whereClause, args := r.buildWhereClause(organizationID, filters)

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM products %s", whereClause)
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		r.logger.Error("Failed to count products", "error", err)
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	// Main query with pagination and sorting
	orderClause := fmt.Sprintf("ORDER BY %s %s", filters.SortBy, strings.ToUpper(filters.SortOrder))
	limitClause := fmt.Sprintf("LIMIT %d OFFSET %d", filters.PageSize, filters.GetOffset())

	query := fmt.Sprintf(`
		SELECT id, organization_id, sku, name, description, category, brand,
			   unit_of_measure, weight, dimensions, barcode, tax_rate,
			   is_active, is_trackable, created_at, updated_at
		FROM products %s %s %s`, whereClause, orderClause, limitClause)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to list products", "error", err)
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}
	defer rows.Close()

	var products []*domain.Product
	for rows.Next() {
		product := &domain.Product{}
		err := rows.Scan(
			&product.ID, &product.OrganizationID, &product.SKU, &product.Name,
			&product.Description, &product.Category, &product.Brand,
			&product.UnitOfMeasure, &product.Weight, &product.Dimensions,
			&product.Barcode, &product.TaxRate, &product.IsActive,
			&product.IsTrackable, &product.CreatedAt, &product.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan product", "error", err)
			return nil, 0, fmt.Errorf("failed to scan product: %w", err)
		}

		// Load related data if requested
		if filters.IncludeVariants {
			variants, err := r.GetVariantsByProductID(ctx, product.ID)
			if err != nil {
				r.logger.Error("Failed to load product variants", "error", err, "productId", product.ID)
			} else {
				// Convert []*domain.ProductVariant to []domain.ProductVariant
				product.Variants = make([]domain.ProductVariant, len(variants))
				for i, variant := range variants {
					product.Variants[i] = *variant
				}
			}
		}

		if filters.IncludePrices {
			prices, err := r.GetPricesByProductID(ctx, product.ID)
			if err != nil {
				r.logger.Error("Failed to load product prices", "error", err, "productId", product.ID)
			} else {
				// Convert []*domain.ProductPrice to []domain.ProductPrice
				product.Prices = make([]domain.ProductPrice, len(prices))
				for i, price := range prices {
					product.Prices[i] = *price
				}
			}
		}

		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate products: %w", err)
	}

	return products, total, nil
}

// buildWhereClause builds the WHERE clause for product filtering
func (r *ProductRepository) buildWhereClause(organizationID uint, filters repository.ProductFilters) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	// Organization filter (always required)
	conditions = append(conditions, fmt.Sprintf("organization_id = $%d", argIndex))
	args = append(args, organizationID)
	argIndex++

	// Category filter
	if filters.Category != "" {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIndex))
		args = append(args, filters.Category)
		argIndex++
	}

	// Brand filter
	if filters.Brand != "" {
		conditions = append(conditions, fmt.Sprintf("brand = $%d", argIndex))
		args = append(args, filters.Brand)
		argIndex++
	}

	// Active filter
	if filters.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *filters.IsActive)
		argIndex++
	}

	// Trackable filter
	if filters.IsTrackable != nil {
		conditions = append(conditions, fmt.Sprintf("is_trackable = $%d", argIndex))
		args = append(args, *filters.IsTrackable)
		argIndex++
	}

	// Search filter (name, SKU, description)
	if filters.Search != "" {
		searchPattern := "%" + strings.ToLower(filters.Search) + "%"
		conditions = append(conditions, fmt.Sprintf("(LOWER(name) LIKE $%d OR LOWER(sku) LIKE $%d OR LOWER(description) LIKE $%d)", argIndex, argIndex, argIndex))
		args = append(args, searchPattern)
		argIndex++
	}

	// Date range filters
	if filters.CreatedFrom != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, *filters.CreatedFrom)
		argIndex++
	}

	if filters.CreatedTo != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, *filters.CreatedTo)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	return whereClause, args
}

// CreateVariant creates a new product variant
func (r *ProductRepository) CreateVariant(ctx context.Context, variant *domain.ProductVariant) error {
	attributesJSON, err := json.Marshal(variant.Attributes)
	if err != nil {
		return fmt.Errorf("failed to marshal attributes: %w", err)
	}

	query := `
		INSERT INTO product_variants (
			product_id, sku, name, description, attributes, weight, 
			dimensions, barcode, is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		) RETURNING id, created_at, updated_at`

	err = r.db.QueryRowContext(ctx, query,
		variant.ProductID, variant.SKU, variant.Name, variant.Description,
		attributesJSON, variant.Weight, variant.Dimensions, variant.Barcode,
		variant.IsActive, variant.CreatedAt, variant.UpdatedAt,
	).Scan(&variant.ID, &variant.CreatedAt, &variant.UpdatedAt)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return domain.ErrVariantAlreadyExists
		}
		r.logger.Error("Failed to create product variant", "error", err, "sku", variant.SKU)
		return fmt.Errorf("failed to create product variant: %w", err)
	}

	r.logger.Info("Product variant created successfully", "variantId", variant.ID, "sku", variant.SKU)
	return nil
}

// GetVariantByID retrieves a product variant by ID
func (r *ProductRepository) GetVariantByID(ctx context.Context, productID, variantID uint) (*domain.ProductVariant, error) {
	query := `
		SELECT id, product_id, sku, name, description, attributes, weight,
			   dimensions, barcode, is_active, created_at, updated_at
		FROM product_variants 
		WHERE id = $1 AND product_id = $2`

	variant := &domain.ProductVariant{}
	var attributesJSON []byte

	err := r.db.QueryRowContext(ctx, query, variantID, productID).Scan(
		&variant.ID, &variant.ProductID, &variant.SKU, &variant.Name,
		&variant.Description, &attributesJSON, &variant.Weight,
		&variant.Dimensions, &variant.Barcode, &variant.IsActive,
		&variant.CreatedAt, &variant.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrVariantNotFound
		}
		r.logger.Error("Failed to get product variant by ID", "error", err, "variantId", variantID)
		return nil, fmt.Errorf("failed to get product variant: %w", err)
	}

	// Unmarshal attributes
	if len(attributesJSON) > 0 {
		if err := json.Unmarshal(attributesJSON, &variant.Attributes); err != nil {
			r.logger.Error("Failed to unmarshal variant attributes", "error", err, "variantId", variantID)
			variant.Attributes = make(map[string]interface{})
		}
	}

	return variant, nil
}

// GetVariantBySKU retrieves a product variant by SKU
func (r *ProductRepository) GetVariantBySKU(ctx context.Context, sku string) (*domain.ProductVariant, error) {
	query := `
		SELECT id, product_id, sku, name, description, attributes, weight,
			   dimensions, barcode, is_active, created_at, updated_at
		FROM product_variants 
		WHERE sku = $1`

	variant := &domain.ProductVariant{}
	var attributesJSON []byte

	err := r.db.QueryRowContext(ctx, query, sku).Scan(
		&variant.ID, &variant.ProductID, &variant.SKU, &variant.Name,
		&variant.Description, &attributesJSON, &variant.Weight,
		&variant.Dimensions, &variant.Barcode, &variant.IsActive,
		&variant.CreatedAt, &variant.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrVariantNotFound
		}
		r.logger.Error("Failed to get product variant by SKU", "error", err, "sku", sku)
		return nil, fmt.Errorf("failed to get product variant: %w", err)
	}

	// Unmarshal attributes
	if len(attributesJSON) > 0 {
		if err := json.Unmarshal(attributesJSON, &variant.Attributes); err != nil {
			r.logger.Error("Failed to unmarshal variant attributes", "error", err, "sku", sku)
			variant.Attributes = make(map[string]interface{})
		}
	}

	return variant, nil
}

// GetVariantsByProductID retrieves all variants for a product
func (r *ProductRepository) GetVariantsByProductID(ctx context.Context, productID uint) ([]*domain.ProductVariant, error) {
	query := `
		SELECT id, product_id, sku, name, description, attributes, weight,
			   dimensions, barcode, is_active, created_at, updated_at
		FROM product_variants 
		WHERE product_id = $1
		ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, productID)
	if err != nil {
		r.logger.Error("Failed to get product variants", "error", err, "productId", productID)
		return nil, fmt.Errorf("failed to get product variants: %w", err)
	}
	defer rows.Close()

	var variants []*domain.ProductVariant
	for rows.Next() {
		variant := &domain.ProductVariant{}
		var attributesJSON []byte

		err := rows.Scan(
			&variant.ID, &variant.ProductID, &variant.SKU, &variant.Name,
			&variant.Description, &attributesJSON, &variant.Weight,
			&variant.Dimensions, &variant.Barcode, &variant.IsActive,
			&variant.CreatedAt, &variant.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan product variant", "error", err)
			return nil, fmt.Errorf("failed to scan product variant: %w", err)
		}

		// Unmarshal attributes
		if len(attributesJSON) > 0 {
			if err := json.Unmarshal(attributesJSON, &variant.Attributes); err != nil {
				r.logger.Error("Failed to unmarshal variant attributes", "error", err, "variantId", variant.ID)
				variant.Attributes = make(map[string]interface{})
			}
		}

		variants = append(variants, variant)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate product variants: %w", err)
	}

	return variants, nil
}

// UpdateVariant updates an existing product variant
func (r *ProductRepository) UpdateVariant(ctx context.Context, variant *domain.ProductVariant) error {
	attributesJSON, err := json.Marshal(variant.Attributes)
	if err != nil {
		return fmt.Errorf("failed to marshal attributes: %w", err)
	}

	query := `
		UPDATE product_variants SET 
			name = $2, description = $3, attributes = $4, weight = $5,
			dimensions = $6, barcode = $7, is_active = $8, updated_at = $9
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		variant.ID, variant.Name, variant.Description, attributesJSON,
		variant.Weight, variant.Dimensions, variant.Barcode,
		variant.IsActive, time.Now(),
	)

	if err != nil {
		r.logger.Error("Failed to update product variant", "error", err, "variantId", variant.ID)
		return fmt.Errorf("failed to update product variant: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrVariantNotFound
	}

	r.logger.Info("Product variant updated successfully", "variantId", variant.ID)
	return nil
}

// DeleteVariant deletes a product variant
func (r *ProductRepository) DeleteVariant(ctx context.Context, productID, variantID uint) error {
	query := `DELETE FROM product_variants WHERE id = $1 AND product_id = $2`

	result, err := r.db.ExecContext(ctx, query, variantID, productID)
	if err != nil {
		r.logger.Error("Failed to delete product variant", "error", err, "variantId", variantID)
		return fmt.Errorf("failed to delete product variant: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrVariantNotFound
	}

	r.logger.Info("Product variant deleted successfully", "variantId", variantID)
	return nil
}

// CreatePrice creates a new product price
func (r *ProductRepository) CreatePrice(ctx context.Context, price *domain.ProductPrice) error {
	query := `
		INSERT INTO product_prices (
			product_id, product_variant_id, price_type, currency, amount,
			min_quantity, max_quantity, valid_from, valid_until,
			is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		) RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		price.ProductID, price.ProductVariantID, price.PriceType,
		price.Currency, price.Amount, price.MinQuantity, price.MaxQuantity,
		price.ValidFrom, price.ValidUntil, price.IsActive,
		price.CreatedAt, price.UpdatedAt,
	).Scan(&price.ID, &price.CreatedAt, &price.UpdatedAt)

	if err != nil {
		r.logger.Error("Failed to create product price", "error", err)
		return fmt.Errorf("failed to create product price: %w", err)
	}

	r.logger.Info("Product price created successfully", "priceId", price.ID)
	return nil
}

// GetPriceByID retrieves a product price by ID
func (r *ProductRepository) GetPriceByID(ctx context.Context, priceID uint) (*domain.ProductPrice, error) {
	query := `
		SELECT id, product_id, product_variant_id, price_type, currency, amount,
			   min_quantity, max_quantity, valid_from, valid_until,
			   is_active, created_at, updated_at
		FROM product_prices 
		WHERE id = $1`

	price := &domain.ProductPrice{}
	err := r.db.QueryRowContext(ctx, query, priceID).Scan(
		&price.ID, &price.ProductID, &price.ProductVariantID, &price.PriceType,
		&price.Currency, &price.Amount, &price.MinQuantity, &price.MaxQuantity,
		&price.ValidFrom, &price.ValidUntil, &price.IsActive,
		&price.CreatedAt, &price.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrPriceNotFound
		}
		r.logger.Error("Failed to get product price by ID", "error", err, "priceId", priceID)
		return nil, fmt.Errorf("failed to get product price: %w", err)
	}

	return price, nil
}

// GetPricesByProductID retrieves all prices for a product
func (r *ProductRepository) GetPricesByProductID(ctx context.Context, productID uint) ([]*domain.ProductPrice, error) {
	query := `
		SELECT id, product_id, product_variant_id, price_type, currency, amount,
			   min_quantity, max_quantity, valid_from, valid_until,
			   is_active, created_at, updated_at
		FROM product_prices 
		WHERE product_id = $1
		ORDER BY price_type, min_quantity`

	return r.queryPrices(ctx, query, productID)
}

// GetPricesByVariantID retrieves all prices for a product variant
func (r *ProductRepository) GetPricesByVariantID(ctx context.Context, variantID uint) ([]*domain.ProductPrice, error) {
	query := `
		SELECT id, product_id, product_variant_id, price_type, currency, amount,
			   min_quantity, max_quantity, valid_from, valid_until,
			   is_active, created_at, updated_at
		FROM product_prices 
		WHERE product_variant_id = $1
		ORDER BY price_type, min_quantity`

	return r.queryPrices(ctx, query, variantID)
}

// GetEffectivePrice retrieves the effective price for a product or variant
func (r *ProductRepository) GetEffectivePrice(ctx context.Context, productID *uint, variantID *uint, priceType domain.PriceType, quantity int, at time.Time) (*domain.ProductPrice, error) {
	var query string
	var args []interface{}

	if productID != nil {
		query = `
			SELECT id, product_id, product_variant_id, price_type, currency, amount,
				   min_quantity, max_quantity, valid_from, valid_until,
				   is_active, created_at, updated_at
			FROM product_prices 
			WHERE product_id = $1 AND price_type = $2 AND is_active = true
			  AND min_quantity <= $3 AND (max_quantity IS NULL OR max_quantity >= $3)
			  AND (valid_from IS NULL OR valid_from <= $4)
			  AND (valid_until IS NULL OR valid_until >= $4)
			ORDER BY min_quantity DESC
			LIMIT 1`
		args = []interface{}{*productID, priceType, quantity, at}
	} else if variantID != nil {
		query = `
			SELECT id, product_id, product_variant_id, price_type, currency, amount,
				   min_quantity, max_quantity, valid_from, valid_until,
				   is_active, created_at, updated_at
			FROM product_prices 
			WHERE product_variant_id = $1 AND price_type = $2 AND is_active = true
			  AND min_quantity <= $3 AND (max_quantity IS NULL OR max_quantity >= $3)
			  AND (valid_from IS NULL OR valid_from <= $4)
			  AND (valid_until IS NULL OR valid_until >= $4)
			ORDER BY min_quantity DESC
			LIMIT 1`
		args = []interface{}{*variantID, priceType, quantity, at}
	} else {
		return nil, fmt.Errorf("either product ID or variant ID must be provided")
	}

	price := &domain.ProductPrice{}
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&price.ID, &price.ProductID, &price.ProductVariantID, &price.PriceType,
		&price.Currency, &price.Amount, &price.MinQuantity, &price.MaxQuantity,
		&price.ValidFrom, &price.ValidUntil, &price.IsActive,
		&price.CreatedAt, &price.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrPriceNotFound
		}
		r.logger.Error("Failed to get effective price", "error", err)
		return nil, fmt.Errorf("failed to get effective price: %w", err)
	}

	return price, nil
}

// UpdatePrice updates an existing product price
func (r *ProductRepository) UpdatePrice(ctx context.Context, price *domain.ProductPrice) error {
	query := `
		UPDATE product_prices SET 
			amount = $2, min_quantity = $3, max_quantity = $4,
			valid_from = $5, valid_until = $6, is_active = $7, updated_at = $8
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		price.ID, price.Amount, price.MinQuantity, price.MaxQuantity,
		price.ValidFrom, price.ValidUntil, price.IsActive, time.Now(),
	)

	if err != nil {
		r.logger.Error("Failed to update product price", "error", err, "priceId", price.ID)
		return fmt.Errorf("failed to update product price: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrPriceNotFound
	}

	r.logger.Info("Product price updated successfully", "priceId", price.ID)
	return nil
}

// DeletePrice deletes a product price
func (r *ProductRepository) DeletePrice(ctx context.Context, priceID uint) error {
	query := `DELETE FROM product_prices WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, priceID)
	if err != nil {
		r.logger.Error("Failed to delete product price", "error", err, "priceId", priceID)
		return fmt.Errorf("failed to delete product price: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrPriceNotFound
	}

	r.logger.Info("Product price deleted successfully", "priceId", priceID)
	return nil
}

// queryPrices is a helper method to query prices
func (r *ProductRepository) queryPrices(ctx context.Context, query string, id uint) ([]*domain.ProductPrice, error) {
	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to query product prices", "error", err)
		return nil, fmt.Errorf("failed to query product prices: %w", err)
	}
	defer rows.Close()

	var prices []*domain.ProductPrice
	for rows.Next() {
		price := &domain.ProductPrice{}
		err := rows.Scan(
			&price.ID, &price.ProductID, &price.ProductVariantID, &price.PriceType,
			&price.Currency, &price.Amount, &price.MinQuantity, &price.MaxQuantity,
			&price.ValidFrom, &price.ValidUntil, &price.IsActive,
			&price.CreatedAt, &price.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan product price", "error", err)
			return nil, fmt.Errorf("failed to scan product price: %w", err)
		}

		prices = append(prices, price)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate product prices: %w", err)
	}

	return prices, nil
}

// BulkCreate creates multiple products in a single transaction
func (r *ProductRepository) BulkCreate(ctx context.Context, products []*domain.Product) error {
	if len(products) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO products (
			organization_id, sku, name, description, category, brand, 
			unit_of_measure, weight, dimensions, barcode, tax_rate, 
			is_active, is_trackable, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		) RETURNING id, created_at, updated_at`

	for _, product := range products {
		err := tx.QueryRowContext(ctx, query,
			product.OrganizationID, product.SKU, product.Name, product.Description,
			product.Category, product.Brand, product.UnitOfMeasure, product.Weight,
			product.Dimensions, product.Barcode, product.TaxRate, product.IsActive,
			product.IsTrackable, product.CreatedAt, product.UpdatedAt,
		).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)

		if err != nil {
			r.logger.Error("Failed to bulk create product", "error", err, "sku", product.SKU)
			return fmt.Errorf("failed to bulk create product %s: %w", product.SKU, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit bulk create transaction: %w", err)
	}

	r.logger.Info("Bulk created products successfully", "count", len(products))
	return nil
}

// BulkUpdate updates multiple products in a single transaction
func (r *ProductRepository) BulkUpdate(ctx context.Context, products []*domain.Product) error {
	if len(products) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		UPDATE products SET 
			name = $2, description = $3, category = $4, brand = $5,
			unit_of_measure = $6, weight = $7, dimensions = $8, barcode = $9,
			tax_rate = $10, is_active = $11, is_trackable = $12, updated_at = $13
		WHERE id = $1 AND organization_id = $14`

	for _, product := range products {
		result, err := tx.ExecContext(ctx, query,
			product.ID, product.Name, product.Description, product.Category,
			product.Brand, product.UnitOfMeasure, product.Weight, product.Dimensions,
			product.Barcode, product.TaxRate, product.IsActive, product.IsTrackable,
			time.Now(), product.OrganizationID,
		)

		if err != nil {
			r.logger.Error("Failed to bulk update product", "error", err, "productId", product.ID)
			return fmt.Errorf("failed to bulk update product %d: %w", product.ID, err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected for product %d: %w", product.ID, err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("product %d not found or not owned by organization", product.ID)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit bulk update transaction: %w", err)
	}

	r.logger.Info("Bulk updated products successfully", "count", len(products))
	return nil
}

// BulkDelete deletes multiple products in a single transaction
func (r *ProductRepository) BulkDelete(ctx context.Context, organizationID uint, productIDs []uint) error {
	if len(productIDs) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Build placeholders for IN clause
	placeholders := make([]string, len(productIDs))
	args := make([]interface{}, len(productIDs)+1)
	args[0] = organizationID

	for i, id := range productIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = id
	}

	query := fmt.Sprintf("DELETE FROM products WHERE organization_id = $1 AND id IN (%s)", strings.Join(placeholders, ","))

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to bulk delete products", "error", err)
		return fmt.Errorf("failed to bulk delete products: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit bulk delete transaction: %w", err)
	}

	r.logger.Info("Bulk deleted products successfully", "count", rowsAffected, "requested", len(productIDs))
	return nil
}

// GetProductStats retrieves product statistics for an organization
func (r *ProductRepository) GetProductStats(ctx context.Context, organizationID uint) (*repository.ProductStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_products,
			COUNT(CASE WHEN is_active = true THEN 1 END) as active_products,
			COUNT(CASE WHEN is_active = false THEN 1 END) as inactive_products,
			COUNT(CASE WHEN is_trackable = true THEN 1 END) as trackable_products,
			COUNT(CASE WHEN created_at >= NOW() - INTERVAL '30 days' THEN 1 END) as recent_products,
			COUNT(DISTINCT category) as total_categories,
			COUNT(DISTINCT brand) as total_brands
		FROM products 
		WHERE organization_id = $1`

	stats := &repository.ProductStats{}
	err := r.db.QueryRowContext(ctx, query, organizationID).Scan(
		&stats.TotalProducts, &stats.ActiveProducts, &stats.InactiveProducts,
		&stats.TrackableProducts, &stats.RecentProducts, &stats.TotalCategories,
		&stats.TotalBrands,
	)

	if err != nil {
		r.logger.Error("Failed to get product stats", "error", err, "organizationId", organizationID)
		return nil, fmt.Errorf("failed to get product stats: %w", err)
	}

	// Get variant count
	variantQuery := `
		SELECT COUNT(*) 
		FROM product_variants pv
		JOIN products p ON pv.product_id = p.id
		WHERE p.organization_id = $1`

	err = r.db.QueryRowContext(ctx, variantQuery, organizationID).Scan(&stats.TotalVariants)
	if err != nil {
		r.logger.Error("Failed to get variant count", "error", err, "organizationId", organizationID)
		// Don't fail the entire operation for this
		stats.TotalVariants = 0
	}

	// Get price statistics
	priceQuery := `
		SELECT 
			COALESCE(AVG(pp.amount), 0) as avg_price,
			COALESCE(MAX(pp.amount), 0) as max_price,
			COALESCE(MIN(pp.amount), 0) as min_price
		FROM product_prices pp
		LEFT JOIN products p ON pp.product_id = p.id
		LEFT JOIN product_variants pv ON pp.product_variant_id = pv.id
		LEFT JOIN products p2 ON pv.product_id = p2.id
		WHERE (p.organization_id = $1 OR p2.organization_id = $1)
		  AND pp.is_active = true
		  AND pp.price_type = 'base'`

	err = r.db.QueryRowContext(ctx, priceQuery, organizationID).Scan(
		&stats.AveragePrice, &stats.HighestPrice, &stats.LowestPrice,
	)

	if err != nil {
		r.logger.Error("Failed to get price stats", "error", err, "organizationId", organizationID)
		// Don't fail the entire operation for this
		stats.AveragePrice = 0
		stats.HighestPrice = 0
		stats.LowestPrice = 0
	}

	return stats, nil
}

// GetCategoriesWithCounts retrieves categories with their product counts
func (r *ProductRepository) GetCategoriesWithCounts(ctx context.Context, organizationID uint) ([]repository.CategoryCount, error) {
	query := `
		SELECT category, COUNT(*) as count
		FROM products 
		WHERE organization_id = $1 AND category IS NOT NULL AND category != ''
		GROUP BY category
		ORDER BY count DESC, category ASC`

	rows, err := r.db.QueryContext(ctx, query, organizationID)
	if err != nil {
		r.logger.Error("Failed to get categories with counts", "error", err, "organizationId", organizationID)
		return nil, fmt.Errorf("failed to get categories with counts: %w", err)
	}
	defer rows.Close()

	var categories []repository.CategoryCount
	for rows.Next() {
		var category repository.CategoryCount
		err := rows.Scan(&category.Category, &category.Count)
		if err != nil {
			r.logger.Error("Failed to scan category count", "error", err)
			return nil, fmt.Errorf("failed to scan category count: %w", err)
		}
		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate categories: %w", err)
	}

	return categories, nil
}

// GetBrandsWithCounts retrieves brands with their product counts
func (r *ProductRepository) GetBrandsWithCounts(ctx context.Context, organizationID uint) ([]repository.BrandCount, error) {
	query := `
		SELECT brand, COUNT(*) as count
		FROM products 
		WHERE organization_id = $1 AND brand IS NOT NULL AND brand != ''
		GROUP BY brand
		ORDER BY count DESC, brand ASC`

	rows, err := r.db.QueryContext(ctx, query, organizationID)
	if err != nil {
		r.logger.Error("Failed to get brands with counts", "error", err, "organizationId", organizationID)
		return nil, fmt.Errorf("failed to get brands with counts: %w", err)
	}
	defer rows.Close()

	var brands []repository.BrandCount
	for rows.Next() {
		var brand repository.BrandCount
		err := rows.Scan(&brand.Brand, &brand.Count)
		if err != nil {
			r.logger.Error("Failed to scan brand count", "error", err)
			return nil, fmt.Errorf("failed to scan brand count: %w", err)
		}
		brands = append(brands, brand)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate brands: %w", err)
	}

	return brands, nil
}

// ListPaginated returns paginated products for an organization
func (r *ProductRepository) ListPaginated(ctx context.Context, organizationID uint, params repository.PaginationParams) (repository.PaginationResult[*domain.Product], error) {
	baseQuery := `
		SELECT p.id, p.organization_id, p.name, p.sku, p.description, p.category, p.brand,
			   p.is_active, p.is_trackable, p.created_at, p.updated_at
		FROM products p
		WHERE p.organization_id = $1`

	countQuery := `SELECT COUNT(*) FROM products WHERE organization_id = $1`

	// Get total count
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, organizationID).Scan(&total); err != nil {
		return repository.PaginationResult[*domain.Product]{}, fmt.Errorf("failed to get total count: %w", err)
	}

	// Build paginated query
	allowedSortFields := []string{"name", "sku", "category", "brand", "created_at", "updated_at"}
	helper := NewPaginationHelper(r.db)
	paginatedQuery := helper.BuildPaginatedQuery(baseQuery, params, allowedSortFields)

	// Execute query
	rows, err := r.db.QueryContext(ctx, paginatedQuery, organizationID)
	if err != nil {
		return repository.PaginationResult[*domain.Product]{}, fmt.Errorf("failed to execute paginated query: %w", err)
	}
	defer rows.Close()

	// Scan results
	var products []*domain.Product
	for rows.Next() {
		product := &domain.Product{}
		err := rows.Scan(
			&product.ID, &product.OrganizationID, &product.Name, &product.SKU,
			&product.Description, &product.Category, &product.Brand,
			&product.IsActive, &product.IsTrackable, &product.CreatedAt, &product.UpdatedAt,
		)
		if err != nil {
			return repository.PaginationResult[*domain.Product]{}, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return repository.PaginationResult[*domain.Product]{}, fmt.Errorf("row iteration error: %w", err)
	}

	return repository.NewPaginationResult(products, total, params), nil
}

// SearchPaginated returns paginated products matching search query
func (r *ProductRepository) SearchPaginated(ctx context.Context, organizationID uint, query string, params repository.PaginationParams) (repository.PaginationResult[*domain.Product], error) {
	baseQuery := `
		SELECT p.id, p.organization_id, p.name, p.sku, p.description, p.category, p.brand,
			   p.is_active, p.is_trackable, p.created_at, p.updated_at
		FROM products p
		WHERE p.organization_id = $1`

	// Add search conditions
	helper := NewPaginationHelper(r.db)
	searchFields := []string{"p.name", "p.sku", "p.description", "p.category", "p.brand"}
	searchQuery, searchArgs := helper.BuildSearchQuery(baseQuery, searchFields, query)

	// Combine arguments
	args := helper.CombineArgs([]interface{}{organizationID}, searchArgs)

	// Build count query
	countQuery := helper.BuildCountQuery(searchQuery)

	// Get total count
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return repository.PaginationResult[*domain.Product]{}, fmt.Errorf("failed to get total count: %w", err)
	}

	// Build paginated query
	allowedSortFields := []string{"p.name", "p.sku", "p.category", "p.brand", "p.created_at", "p.updated_at"}
	paginatedQuery := helper.BuildPaginatedQuery(searchQuery, params, allowedSortFields)

	// Execute query
	rows, err := r.db.QueryContext(ctx, paginatedQuery, args...)
	if err != nil {
		return repository.PaginationResult[*domain.Product]{}, fmt.Errorf("failed to execute paginated query: %w", err)
	}
	defer rows.Close()

	// Scan results
	var products []*domain.Product
	for rows.Next() {
		product := &domain.Product{}
		err := rows.Scan(
			&product.ID, &product.OrganizationID, &product.Name, &product.SKU,
			&product.Description, &product.Category, &product.Brand,
			&product.IsActive, &product.IsTrackable, &product.CreatedAt, &product.UpdatedAt,
		)
		if err != nil {
			return repository.PaginationResult[*domain.Product]{}, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return repository.PaginationResult[*domain.Product]{}, fmt.Errorf("row iteration error: %w", err)
	}

	return repository.NewPaginationResult(products, total, params), nil
}
