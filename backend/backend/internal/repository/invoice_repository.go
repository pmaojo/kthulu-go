// @kthulu:module:invoices
package repository

import (
	"context"
	"time"

	"github.com/kthulu/kthulu-go/backend/internal/domain"
)

// InvoiceRepository defines the interface for invoice data operations
type InvoiceRepository interface {
	// Invoice operations
	Create(ctx context.Context, invoice *domain.Invoice) error
	GetByID(ctx context.Context, organizationID, invoiceID uint) (*domain.Invoice, error)
	GetByNumber(ctx context.Context, organizationID uint, invoiceNumber string) (*domain.Invoice, error)
	Update(ctx context.Context, invoice *domain.Invoice) error
	Delete(ctx context.Context, organizationID, invoiceID uint) error
	List(ctx context.Context, organizationID uint, filters InvoiceFilters) ([]*domain.Invoice, int64, error)
	ListPaginated(ctx context.Context, organizationID uint, params PaginationParams) (PaginationResult[*domain.Invoice], error)
	SearchPaginated(ctx context.Context, organizationID uint, query string, params PaginationParams) (PaginationResult[*domain.Invoice], error)

	// Invoice item operations
	CreateItem(ctx context.Context, item *domain.InvoiceItem) error
	GetItemByID(ctx context.Context, invoiceID, itemID uint) (*domain.InvoiceItem, error)
	GetItemsByInvoiceID(ctx context.Context, invoiceID uint) ([]*domain.InvoiceItem, error)
	UpdateItem(ctx context.Context, item *domain.InvoiceItem) error
	DeleteItem(ctx context.Context, invoiceID, itemID uint) error
	BulkCreateItems(ctx context.Context, items []*domain.InvoiceItem) error
	BulkUpdateItems(ctx context.Context, items []*domain.InvoiceItem) error
	BulkDeleteItems(ctx context.Context, invoiceID uint, itemIDs []uint) error

	// Payment operations
	CreatePayment(ctx context.Context, payment *domain.Payment) error
	GetPaymentByID(ctx context.Context, organizationID, paymentID uint) (*domain.Payment, error)
	GetPaymentsByInvoiceID(ctx context.Context, invoiceID uint) ([]*domain.Payment, error)
	UpdatePayment(ctx context.Context, payment *domain.Payment) error
	DeletePayment(ctx context.Context, organizationID, paymentID uint) error
	ListPayments(ctx context.Context, organizationID uint, filters PaymentFilters) ([]*domain.Payment, int64, error)

	// Bulk operations
	BulkCreate(ctx context.Context, invoices []*domain.Invoice) error
	BulkUpdate(ctx context.Context, invoices []*domain.Invoice) error
	BulkDelete(ctx context.Context, organizationID uint, invoiceIDs []uint) error
	BulkUpdateStatus(ctx context.Context, organizationID uint, invoiceIDs []uint, status domain.InvoiceStatus) error

	// Statistics and analytics
	GetInvoiceStats(ctx context.Context, organizationID uint) (*InvoiceStats, error)
	GetRevenueStats(ctx context.Context, organizationID uint, from, to time.Time) (*RevenueStats, error)
	GetOverdueInvoices(ctx context.Context, organizationID uint) ([]*domain.Invoice, error)
	GetUpcomingDueInvoices(ctx context.Context, organizationID uint, days int) ([]*domain.Invoice, error)

	// Number generation
	GenerateInvoiceNumber(ctx context.Context, organizationID uint, invoiceType domain.InvoiceType) (string, error)
}

// InvoiceFilters represents filters for invoice listing
type InvoiceFilters struct {
	ContactID  *uint                 `json:"contactId,omitempty"`
	Type       *domain.InvoiceType   `json:"type,omitempty"`
	Status     *domain.InvoiceStatus `json:"status,omitempty"`
	Currency   string                `json:"currency,omitempty"`
	Search     string                `json:"search,omitempty"`     // Search in invoice number, contact name
	IssuedFrom *string               `json:"issuedFrom,omitempty"` // ISO date string
	IssuedTo   *string               `json:"issuedTo,omitempty"`   // ISO date string
	DueFrom    *string               `json:"dueFrom,omitempty"`    // ISO date string
	DueTo      *string               `json:"dueTo,omitempty"`      // ISO date string
	MinAmount  *float64              `json:"minAmount,omitempty"`
	MaxAmount  *float64              `json:"maxAmount,omitempty"`
	IsOverdue  *bool                 `json:"isOverdue,omitempty"`
	CreatedBy  *uint                 `json:"createdBy,omitempty"`

	// Pagination
	Page     int `json:"page" validate:"min=1"`
	PageSize int `json:"pageSize" validate:"min=1,max=100"`

	// Sorting
	SortBy    string `json:"sortBy,omitempty"`    // invoice_number, issue_date, due_date, total_amount, status, created_at
	SortOrder string `json:"sortOrder,omitempty"` // asc, desc

	// Include related data
	IncludeItems    bool `json:"includeItems,omitempty"`
	IncludePayments bool `json:"includePayments,omitempty"`
	IncludeContact  bool `json:"includeContact,omitempty"`
}

// PaymentFilters represents filters for payment listing
type PaymentFilters struct {
	InvoiceID     *uint                 `json:"invoiceId,omitempty"`
	PaymentMethod *domain.PaymentMethod `json:"paymentMethod,omitempty"`
	Currency      string                `json:"currency,omitempty"`
	Search        string                `json:"search,omitempty"`      // Search in reference number
	PaymentFrom   *string               `json:"paymentFrom,omitempty"` // ISO date string
	PaymentTo     *string               `json:"paymentTo,omitempty"`   // ISO date string
	MinAmount     *float64              `json:"minAmount,omitempty"`
	MaxAmount     *float64              `json:"maxAmount,omitempty"`
	CreatedBy     *uint                 `json:"createdBy,omitempty"`

	// Pagination
	Page     int `json:"page" validate:"min=1"`
	PageSize int `json:"pageSize" validate:"min=1,max=100"`

	// Sorting
	SortBy    string `json:"sortBy,omitempty"`    // payment_date, amount, payment_method, created_at
	SortOrder string `json:"sortOrder,omitempty"` // asc, desc
}

// InvoiceStats represents invoice statistics for an organization
type InvoiceStats struct {
	TotalInvoices       int64   `json:"totalInvoices"`
	DraftInvoices       int64   `json:"draftInvoices"`
	SentInvoices        int64   `json:"sentInvoices"`
	PaidInvoices        int64   `json:"paidInvoices"`
	OverdueInvoices     int64   `json:"overdueInvoices"`
	CancelledInvoices   int64   `json:"canceledInvoices"`
	TotalRevenue        float64 `json:"totalRevenue"`
	PaidRevenue         float64 `json:"paidRevenue"`
	OutstandingAmount   float64 `json:"outstandingAmount"`
	OverdueAmount       float64 `json:"overdueAmount"`
	AverageInvoiceValue float64 `json:"averageInvoiceValue"`
	AveragePaymentTime  float64 `json:"averagePaymentTime"` // Days
}

// RevenueStats represents revenue statistics for a time period
type RevenueStats struct {
	Period              string    `json:"period"`
	StartDate           time.Time `json:"startDate"`
	EndDate             time.Time `json:"endDate"`
	TotalRevenue        float64   `json:"totalRevenue"`
	PaidRevenue         float64   `json:"paidRevenue"`
	OutstandingAmount   float64   `json:"outstandingAmount"`
	InvoiceCount        int64     `json:"invoiceCount"`
	PaymentCount        int64     `json:"paymentCount"`
	AverageInvoiceValue float64   `json:"averageInvoiceValue"`
	Currency            string    `json:"currency"`
}

// DefaultInvoiceFilters returns default filters for invoice listing
func DefaultInvoiceFilters() InvoiceFilters {
	return InvoiceFilters{
		Page:      1,
		PageSize:  20,
		SortBy:    "created_at",
		SortOrder: "desc",
	}
}

// Validate validates the invoice filters
func (f *InvoiceFilters) Validate() error {
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
	return nil
}

// GetOffset returns the offset for pagination
func (f *InvoiceFilters) GetOffset() int {
	return (f.Page - 1) * f.PageSize
}

// DefaultPaymentFilters returns default filters for payment listing
func DefaultPaymentFilters() PaymentFilters {
	return PaymentFilters{
		Page:      1,
		PageSize:  20,
		SortBy:    "payment_date",
		SortOrder: "desc",
	}
}

// Validate validates the payment filters
func (f *PaymentFilters) Validate() error {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 100 {
		f.PageSize = 20
	}
	if f.SortBy == "" {
		f.SortBy = "payment_date"
	}
	if f.SortOrder != "asc" && f.SortOrder != "desc" {
		f.SortOrder = "desc"
	}
	return nil
}

// GetOffset returns the offset for pagination
func (f *PaymentFilters) GetOffset() int {
	return (f.Page - 1) * f.PageSize
}
