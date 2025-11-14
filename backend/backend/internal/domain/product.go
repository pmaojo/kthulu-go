// @kthulu:module:products
package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

// Domain errors for product module
var (
	ErrProductNotFound      = errors.New("product not found")
	ErrProductAlreadyExists = errors.New("product already exists")
	ErrInvalidSKU           = errors.New("invalid SKU")
	ErrVariantNotFound      = errors.New("product variant not found")
	ErrVariantAlreadyExists = errors.New("product variant already exists")
	ErrPriceNotFound        = errors.New("product price not found")
	ErrInvalidPrice         = errors.New("invalid price")
	ErrInvalidPriceType     = errors.New("invalid price type")
	ErrInvalidCurrency      = errors.New("invalid currency")
	ErrInvalidQuantityRange = errors.New("invalid quantity range")
)

// PriceType represents the type of price
type PriceType string

const (
	PriceTypeBase      PriceType = "base"
	PriceTypeSale      PriceType = "sale"
	PriceTypeWholesale PriceType = "wholesale"
	PriceTypeRetail    PriceType = "retail"
	PriceTypeCost      PriceType = "cost"
)

// Product represents a product in the catalog
type Product struct {
	ID             uint      `json:"id"`
	OrganizationID uint      `json:"organizationId" validate:"required"`
	SKU            string    `json:"sku" validate:"required,min=1,max=100"`
	Name           string    `json:"name" validate:"required,min=1,max=200"`
	Description    string    `json:"description,omitempty"`
	Category       string    `json:"category,omitempty" validate:"max=100"`
	Brand          string    `json:"brand,omitempty" validate:"max=100"`
	UnitOfMeasure  string    `json:"unitOfMeasure" validate:"required,max=20"`
	Weight         *float64  `json:"weight,omitempty" validate:"omitempty,min=0"`
	Dimensions     string    `json:"dimensions,omitempty" validate:"max=100"`
	Barcode        string    `json:"barcode,omitempty" validate:"max=100"`
	TaxRate        float64   `json:"taxRate" validate:"min=0,max=1"`
	IsActive       bool      `json:"isActive"`
	IsTrackable    bool      `json:"isTrackable"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`

	// Related entities (loaded separately)
	Variants []ProductVariant `json:"variants,omitempty"`
	Prices   []ProductPrice   `json:"prices,omitempty"`
}

// ProductVariant represents a variant of a product
type ProductVariant struct {
	ID          uint                   `json:"id"`
	ProductID   uint                   `json:"productId" validate:"required"`
	SKU         string                 `json:"sku" validate:"required,min=1,max=100"`
	Name        string                 `json:"name" validate:"required,min=1,max=200"`
	Description string                 `json:"description,omitempty"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
	Weight      *float64               `json:"weight,omitempty" validate:"omitempty,min=0"`
	Dimensions  string                 `json:"dimensions,omitempty" validate:"max=100"`
	Barcode     string                 `json:"barcode,omitempty" validate:"max=100"`
	IsActive    bool                   `json:"isActive"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`

	// Related entities (loaded separately)
	Prices []ProductPrice `json:"prices,omitempty"`
}

// ProductPrice represents pricing information for a product or variant
type ProductPrice struct {
	ID               uint       `json:"id"`
	ProductID        *uint      `json:"productId,omitempty"`
	ProductVariantID *uint      `json:"productVariantId,omitempty"`
	PriceType        PriceType  `json:"priceType" validate:"required,oneof=base sale wholesale retail cost"`
	Currency         string     `json:"currency" validate:"required,len=3"`
	Amount           float64    `json:"amount" validate:"required,min=0"`
	MinQuantity      int        `json:"minQuantity" validate:"min=1"`
	MaxQuantity      *int       `json:"maxQuantity,omitempty" validate:"omitempty,min=1"`
	ValidFrom        *time.Time `json:"validFrom,omitempty"`
	ValidUntil       *time.Time `json:"validUntil,omitempty"`
	IsActive         bool       `json:"isActive"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

// NewProduct creates a new product with validation
func NewProduct(organizationID uint, sku, name, unitOfMeasure string) (*Product, error) {
	product := &Product{
		OrganizationID: organizationID,
		SKU:            strings.TrimSpace(sku),
		Name:           strings.TrimSpace(name),
		UnitOfMeasure:  strings.TrimSpace(unitOfMeasure),
		TaxRate:        0.0,
		IsActive:       true,
		IsTrackable:    true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := product.Validate(); err != nil {
		return nil, err
	}

	return product, nil
}

// Validate validates the product data
func (p *Product) Validate() error {
	validate := validator.New()
	if err := validate.Struct(p); err != nil {
		return err
	}

	// Business rule: SKU must be unique within organization (handled by repository)
	if p.SKU == "" {
		return ErrInvalidSKU
	}

	return nil
}

// UpdateBasicInfo updates the basic product information
func (p *Product) UpdateBasicInfo(name, description, category, brand, unitOfMeasure, dimensions, barcode string, weight *float64, taxRate float64) error {
	p.Name = strings.TrimSpace(name)
	p.Description = strings.TrimSpace(description)
	p.Category = strings.TrimSpace(category)
	p.Brand = strings.TrimSpace(brand)
	p.UnitOfMeasure = strings.TrimSpace(unitOfMeasure)
	p.Dimensions = strings.TrimSpace(dimensions)
	p.Barcode = strings.TrimSpace(barcode)
	p.Weight = weight
	p.TaxRate = taxRate
	p.UpdatedAt = time.Now()

	return p.Validate()
}

// SetActive sets the active status of the product
func (p *Product) SetActive(active bool) {
	p.IsActive = active
	p.UpdatedAt = time.Now()
}

// SetTrackable sets the trackable status of the product
func (p *Product) SetTrackable(trackable bool) {
	p.IsTrackable = trackable
	p.UpdatedAt = time.Now()
}

// GetDisplayName returns the display name for the product
func (p *Product) GetDisplayName() string {
	if p.Brand != "" {
		return p.Brand + " " + p.Name
	}
	return p.Name
}

// NewProductVariant creates a new product variant with validation
func NewProductVariant(productID uint, sku, name string, attributes map[string]interface{}) (*ProductVariant, error) {
	variant := &ProductVariant{
		ProductID:  productID,
		SKU:        strings.TrimSpace(sku),
		Name:       strings.TrimSpace(name),
		Attributes: attributes,
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := variant.Validate(); err != nil {
		return nil, err
	}

	return variant, nil
}

// Validate validates the product variant data
func (pv *ProductVariant) Validate() error {
	validate := validator.New()
	if err := validate.Struct(pv); err != nil {
		return err
	}

	if pv.SKU == "" {
		return ErrInvalidSKU
	}

	return nil
}

// UpdateBasicInfo updates the basic variant information
func (pv *ProductVariant) UpdateBasicInfo(name, description, dimensions, barcode string, weight *float64, attributes map[string]interface{}) error {
	pv.Name = strings.TrimSpace(name)
	pv.Description = strings.TrimSpace(description)
	pv.Dimensions = strings.TrimSpace(dimensions)
	pv.Barcode = strings.TrimSpace(barcode)
	pv.Weight = weight
	pv.Attributes = attributes
	pv.UpdatedAt = time.Now()

	return pv.Validate()
}

// SetActive sets the active status of the variant
func (pv *ProductVariant) SetActive(active bool) {
	pv.IsActive = active
	pv.UpdatedAt = time.Now()
}

// GetAttributeValue gets a specific attribute value
func (pv *ProductVariant) GetAttributeValue(key string) (interface{}, bool) {
	if pv.Attributes == nil {
		return nil, false
	}
	value, exists := pv.Attributes[key]
	return value, exists
}

// SetAttributeValue sets a specific attribute value
func (pv *ProductVariant) SetAttributeValue(key string, value interface{}) {
	if pv.Attributes == nil {
		pv.Attributes = make(map[string]interface{})
	}
	pv.Attributes[key] = value
	pv.UpdatedAt = time.Now()
}

// NewProductPrice creates a new product price with validation
func NewProductPrice(productID *uint, variantID *uint, priceType PriceType, currency string, amount float64, minQuantity int) (*ProductPrice, error) {
	// Validate that either productID or variantID is set, but not both
	if (productID == nil && variantID == nil) || (productID != nil && variantID != nil) {
		return nil, errors.New("either product ID or variant ID must be set, but not both")
	}

	price := &ProductPrice{
		ProductID:        productID,
		ProductVariantID: variantID,
		PriceType:        priceType,
		Currency:         strings.ToUpper(strings.TrimSpace(currency)),
		Amount:           amount,
		MinQuantity:      minQuantity,
		IsActive:         true,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := price.Validate(); err != nil {
		return nil, err
	}

	return price, nil
}

// Validate validates the product price data
func (pp *ProductPrice) Validate() error {
	validate := validator.New()
	if err := validate.Struct(pp); err != nil {
		return err
	}

	if pp.Amount < 0 {
		return ErrInvalidPrice
	}

	if pp.Currency == "" || len(pp.Currency) != 3 {
		return ErrInvalidCurrency
	}

	if pp.MaxQuantity != nil && *pp.MaxQuantity < pp.MinQuantity {
		return ErrInvalidQuantityRange
	}

	if pp.ValidFrom != nil && pp.ValidUntil != nil && pp.ValidUntil.Before(*pp.ValidFrom) {
		return errors.New("valid until date must be after valid from date")
	}

	return nil
}

// UpdatePrice updates the price information
func (pp *ProductPrice) UpdatePrice(amount float64, minQuantity int, maxQuantity *int, validFrom, validUntil *time.Time) error {
	pp.Amount = amount
	pp.MinQuantity = minQuantity
	pp.MaxQuantity = maxQuantity
	pp.ValidFrom = validFrom
	pp.ValidUntil = validUntil
	pp.UpdatedAt = time.Now()

	return pp.Validate()
}

// SetActive sets the active status of the price
func (pp *ProductPrice) SetActive(active bool) {
	pp.IsActive = active
	pp.UpdatedAt = time.Now()
}

// IsValidAt checks if the price is valid at a specific time
func (pp *ProductPrice) IsValidAt(t time.Time) bool {
	if !pp.IsActive {
		return false
	}

	if pp.ValidFrom != nil && t.Before(*pp.ValidFrom) {
		return false
	}

	if pp.ValidUntil != nil && t.After(*pp.ValidUntil) {
		return false
	}

	return true
}

// IsValidForQuantity checks if the price is valid for a specific quantity
func (pp *ProductPrice) IsValidForQuantity(quantity int) bool {
	if quantity < pp.MinQuantity {
		return false
	}

	if pp.MaxQuantity != nil && quantity > *pp.MaxQuantity {
		return false
	}

	return true
}

// GetEffectivePrice calculates the effective price for a given quantity and time
func (pp *ProductPrice) GetEffectivePrice(quantity int, t time.Time) (float64, bool) {
	if !pp.IsValidAt(t) || !pp.IsValidForQuantity(quantity) {
		return 0, false
	}

	return pp.Amount, true
}
