// @kthulu:module:invoices
package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"backend/core"
	"backend/internal/domain"
	"backend/internal/repository"
)

// InvoiceUseCase orchestrates invoice management workflows
type InvoiceUseCase struct {
	invoices repository.InvoiceRepository
	logger   core.Logger
}

// NewInvoiceUseCase creates a new invoice use case instance
func NewInvoiceUseCase(
	invoices repository.InvoiceRepository,
	logger core.Logger,
) *InvoiceUseCase {
	return &InvoiceUseCase{
		invoices: invoices,
		logger:   logger,
	}
}

// CreateInvoiceRequest contains the data needed to create a new invoice
type CreateInvoiceRequest struct {
	OrganizationID  uint                       `json:"organizationId" validate:"required"`
	ContactID       uint                       `json:"contactId" validate:"required"`
	Type            domain.InvoiceType         `json:"type" validate:"required,oneof=invoice quote credit_note proforma"`
	Currency        string                     `json:"currency" validate:"required,len=3"`
	IssueDate       time.Time                  `json:"issueDate" validate:"required"`
	DueDate         *time.Time                 `json:"dueDate,omitempty"`
	PaymentTerms    string                     `json:"paymentTerms,omitempty" validate:"max=50"`
	Notes           string                     `json:"notes,omitempty"`
	TermsConditions string                     `json:"termsConditions,omitempty"`
	CreatedBy       uint                       `json:"createdBy" validate:"required"`
	Items           []CreateInvoiceItemRequest `json:"items,omitempty"`
}

// CreateInvoiceItemRequest contains the data needed to create an invoice item
type CreateInvoiceItemRequest struct {
	ProductID        *uint   `json:"productId,omitempty"`
	ProductVariantID *uint   `json:"productVariantId,omitempty"`
	Description      string  `json:"description" validate:"required,min=1,max=500"`
	Quantity         float64 `json:"quantity" validate:"required,min=0"`
	UnitPrice        float64 `json:"unitPrice" validate:"required,min=0"`
	DiscountPercent  float64 `json:"discountPercent" validate:"min=0,max=1"`
	TaxRate          float64 `json:"taxRate" validate:"min=0,max=1"`
}

// UpdateInvoiceRequest contains the data needed to update an invoice
type UpdateInvoiceRequest struct {
	ContactID       uint       `json:"contactId" validate:"required"`
	DueDate         *time.Time `json:"dueDate,omitempty"`
	PaymentTerms    string     `json:"paymentTerms,omitempty" validate:"max=50"`
	Notes           string     `json:"notes,omitempty"`
	TermsConditions string     `json:"termsConditions,omitempty"`
}

// CreatePaymentRequest contains the data needed to create a payment
type CreatePaymentRequest struct {
	OrganizationID  uint                 `json:"organizationId" validate:"required"`
	InvoiceID       uint                 `json:"invoiceId" validate:"required"`
	PaymentMethod   domain.PaymentMethod `json:"paymentMethod" validate:"required,oneof=cash check credit_card bank_transfer paypal stripe other"`
	ReferenceNumber string               `json:"referenceNumber,omitempty" validate:"max=100"`
	Amount          float64              `json:"amount" validate:"required,min=0"`
	Currency        string               `json:"currency" validate:"required,len=3"`
	PaymentDate     time.Time            `json:"paymentDate" validate:"required"`
	Notes           string               `json:"notes,omitempty"`
	CreatedBy       uint                 `json:"createdBy" validate:"required"`
}

// InvoiceListResponse represents a paginated list of invoices
type InvoiceListResponse struct {
	Invoices   []*domain.Invoice `json:"invoices"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"pageSize"`
	TotalPages int64             `json:"totalPages"`
}

// PaymentListResponse represents a paginated list of payments
type PaymentListResponse struct {
	Payments   []*domain.Payment `json:"payments"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"pageSize"`
	TotalPages int64             `json:"totalPages"`
}

// CreateInvoice creates a new invoice
func (uc *InvoiceUseCase) CreateInvoice(ctx context.Context, req CreateInvoiceRequest) (*domain.Invoice, error) {
	uc.logger.Info("Creating new invoice", "organizationId", req.OrganizationID, "contactId", req.ContactID)

	// Generate invoice number
	invoiceNumber, err := uc.invoices.GenerateInvoiceNumber(ctx, req.OrganizationID, req.Type)
	if err != nil {
		uc.logger.Error("Failed to generate invoice number", "error", err)
		return nil, fmt.Errorf("failed to generate invoice number: %w", err)
	}

	// Create invoice domain entity
	invoice, err := domain.NewInvoice(
		req.OrganizationID, req.ContactID, req.CreatedBy,
		invoiceNumber, req.Type, req.Currency, req.IssueDate,
	)
	if err != nil {
		uc.logger.Error("Failed to create invoice domain entity", "error", err)
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	// Update additional properties
	if err := invoice.UpdateBasicInfo(req.ContactID, req.DueDate, req.PaymentTerms, req.Notes, req.TermsConditions); err != nil {
		uc.logger.Error("Failed to update invoice basic info", "error", err)
		return nil, fmt.Errorf("failed to update invoice info: %w", err)
	}

	// Persist invoice
	if err := uc.invoices.Create(ctx, invoice); err != nil {
		uc.logger.Error("Failed to persist invoice", "error", err)
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	// Create invoice items if provided
	if len(req.Items) > 0 {
		for i, itemReq := range req.Items {
			item, err := domain.NewInvoiceItem(
				invoice.ID, itemReq.Description, itemReq.Quantity, itemReq.UnitPrice,
			)
			if err != nil {
				uc.logger.Error("Failed to create invoice item", "error", err, "itemIndex", i)
				return nil, fmt.Errorf("failed to create invoice item %d: %w", i, err)
			}

			// Update item properties
			if err := item.UpdateBasicInfo(
				itemReq.ProductID, itemReq.ProductVariantID, itemReq.Description,
				itemReq.Quantity, itemReq.UnitPrice, itemReq.DiscountPercent, itemReq.TaxRate,
			); err != nil {
				uc.logger.Error("Failed to update invoice item info", "error", err, "itemIndex", i)
				return nil, fmt.Errorf("failed to update invoice item %d: %w", i, err)
			}

			item.SortOrder = i

			// Persist item
			if err := uc.invoices.CreateItem(ctx, item); err != nil {
				uc.logger.Error("Failed to persist invoice item", "error", err, "itemIndex", i)
				return nil, fmt.Errorf("failed to create invoice item %d: %w", i, err)
			}

			invoice.Items = append(invoice.Items, *item)
		}

		// Recalculate totals
		invoice.CalculateTotals()

		// Update invoice with calculated totals
		if err := uc.invoices.Update(ctx, invoice); err != nil {
			uc.logger.Error("Failed to update invoice totals", "error", err)
			return nil, fmt.Errorf("failed to update invoice totals: %w", err)
		}
	}

	uc.logger.Info("Invoice created successfully", "invoiceId", invoice.ID, "invoiceNumber", invoice.InvoiceNumber)
	return invoice, nil
}

// GetInvoice retrieves an invoice by ID
func (uc *InvoiceUseCase) GetInvoice(ctx context.Context, organizationID, invoiceID uint) (*domain.Invoice, error) {
	uc.logger.Info("Getting invoice", "organizationId", organizationID, "invoiceId", invoiceID)

	invoice, err := uc.invoices.GetByID(ctx, organizationID, invoiceID)
	if err != nil {
		if errors.Is(err, domain.ErrInvoiceNotFound) {
			uc.logger.Warn("Invoice not found", "invoiceId", invoiceID, "organizationId", organizationID)
			return nil, domain.ErrInvoiceNotFound
		}
		uc.logger.Error("Failed to get invoice", "error", err, "invoiceId", invoiceID)
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	return invoice, nil
}

// GetInvoiceByNumber retrieves an invoice by number
func (uc *InvoiceUseCase) GetInvoiceByNumber(ctx context.Context, organizationID uint, invoiceNumber string) (*domain.Invoice, error) {
	uc.logger.Info("Getting invoice by number", "organizationId", organizationID, "invoiceNumber", invoiceNumber)

	invoice, err := uc.invoices.GetByNumber(ctx, organizationID, invoiceNumber)
	if err != nil {
		if errors.Is(err, domain.ErrInvoiceNotFound) {
			uc.logger.Warn("Invoice not found", "invoiceNumber", invoiceNumber, "organizationId", organizationID)
			return nil, domain.ErrInvoiceNotFound
		}
		uc.logger.Error("Failed to get invoice by number", "error", err, "invoiceNumber", invoiceNumber)
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	return invoice, nil
}

// UpdateInvoice updates an existing invoice
func (uc *InvoiceUseCase) UpdateInvoice(ctx context.Context, organizationID, invoiceID uint, req UpdateInvoiceRequest) (*domain.Invoice, error) {
	uc.logger.Info("Updating invoice", "organizationId", organizationID, "invoiceId", invoiceID)

	// Get existing invoice
	invoice, err := uc.invoices.GetByID(ctx, organizationID, invoiceID)
	if err != nil {
		if errors.Is(err, domain.ErrInvoiceNotFound) {
			uc.logger.Warn("Invoice not found for update", "invoiceId", invoiceID, "organizationId", organizationID)
			return nil, domain.ErrInvoiceNotFound
		}
		uc.logger.Error("Failed to get invoice for update", "error", err, "invoiceId", invoiceID)
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	// Check if invoice can be edited
	if !invoice.CanEdit() {
		uc.logger.Warn("Attempt to edit non-editable invoice", "invoiceId", invoiceID, "status", invoice.Status)
		return nil, domain.ErrInvoiceNotEditable
	}

	// Update invoice information
	if err := invoice.UpdateBasicInfo(req.ContactID, req.DueDate, req.PaymentTerms, req.Notes, req.TermsConditions); err != nil {
		uc.logger.Error("Failed to update invoice basic info", "error", err, "invoiceId", invoiceID)
		return nil, fmt.Errorf("failed to update invoice info: %w", err)
	}

	// Persist changes
	if err := uc.invoices.Update(ctx, invoice); err != nil {
		uc.logger.Error("Failed to persist invoice update", "error", err, "invoiceId", invoiceID)
		return nil, fmt.Errorf("failed to update invoice: %w", err)
	}

	uc.logger.Info("Invoice updated successfully", "invoiceId", invoiceID)
	return invoice, nil
}

// DeleteInvoice deletes an invoice
func (uc *InvoiceUseCase) DeleteInvoice(ctx context.Context, organizationID, invoiceID uint) error {
	uc.logger.Info("Deleting invoice", "organizationId", organizationID, "invoiceId", invoiceID)

	// Check if invoice exists and can be deleted
	invoice, err := uc.invoices.GetByID(ctx, organizationID, invoiceID)
	if err != nil {
		if errors.Is(err, domain.ErrInvoiceNotFound) {
			uc.logger.Warn("Invoice not found for deletion", "invoiceId", invoiceID, "organizationId", organizationID)
			return domain.ErrInvoiceNotFound
		}
		uc.logger.Error("Failed to get invoice for deletion", "error", err, "invoiceId", invoiceID)
		return fmt.Errorf("failed to get invoice: %w", err)
	}

	// Business rule: only draft invoices can be deleted
	if invoice.Status != domain.InvoiceStatusDraft {
		uc.logger.Warn("Attempt to delete non-draft invoice", "invoiceId", invoiceID, "status", invoice.Status)
		return errors.New("only draft invoices can be deleted")
	}

	// Delete invoice
	if err := uc.invoices.Delete(ctx, organizationID, invoiceID); err != nil {
		uc.logger.Error("Failed to delete invoice", "error", err, "invoiceId", invoiceID)
		return fmt.Errorf("failed to delete invoice: %w", err)
	}

	uc.logger.Info("Invoice deleted successfully", "invoiceId", invoiceID)
	return nil
}

// ListInvoices retrieves invoices with filtering and pagination
func (uc *InvoiceUseCase) ListInvoices(ctx context.Context, organizationID uint, filters repository.InvoiceFilters) (*InvoiceListResponse, error) {
	uc.logger.Info("Listing invoices", "organizationId", organizationID, "filters", filters)

	// Validate and set defaults for filters
	if err := filters.Validate(); err != nil {
		uc.logger.Error("Invalid invoice filters", "error", err, "filters", filters)
		return nil, fmt.Errorf("invalid filters: %w", err)
	}

	invoices, total, err := uc.invoices.List(ctx, organizationID, filters)
	if err != nil {
		uc.logger.Error("Failed to list invoices", "error", err, "organizationId", organizationID)
		return nil, fmt.Errorf("failed to list invoices: %w", err)
	}

	response := &InvoiceListResponse{
		Invoices:   invoices,
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		TotalPages: (total + int64(filters.PageSize) - 1) / int64(filters.PageSize),
	}

	uc.logger.Info("Invoices listed successfully", "organizationId", organizationID, "count", len(invoices), "total", total)
	return response, nil
}

// SetInvoiceStatus sets the status of an invoice
func (uc *InvoiceUseCase) SetInvoiceStatus(ctx context.Context, organizationID, invoiceID uint, status domain.InvoiceStatus) error {
	uc.logger.Info("Setting invoice status", "organizationId", organizationID, "invoiceId", invoiceID, "status", status)

	// Get existing invoice
	invoice, err := uc.invoices.GetByID(ctx, organizationID, invoiceID)
	if err != nil {
		if errors.Is(err, domain.ErrInvoiceNotFound) {
			uc.logger.Warn("Invoice not found for status update", "invoiceId", invoiceID, "organizationId", organizationID)
			return domain.ErrInvoiceNotFound
		}
		uc.logger.Error("Failed to get invoice for status update", "error", err, "invoiceId", invoiceID)
		return fmt.Errorf("failed to get invoice: %w", err)
	}

	// Set status
	if err := invoice.SetStatus(status); err != nil {
		uc.logger.Error("Failed to set invoice status", "error", err, "invoiceId", invoiceID, "status", status)
		return fmt.Errorf("failed to set invoice status: %w", err)
	}

	// Persist changes
	if err := uc.invoices.Update(ctx, invoice); err != nil {
		uc.logger.Error("Failed to persist invoice status update", "error", err, "invoiceId", invoiceID)
		return fmt.Errorf("failed to update invoice status: %w", err)
	}

	uc.logger.Info("Invoice status updated successfully", "invoiceId", invoiceID, "status", status)
	return nil
}

// CreatePayment creates a new payment for an invoice
func (uc *InvoiceUseCase) CreatePayment(ctx context.Context, req CreatePaymentRequest) (*domain.Payment, error) {
	uc.logger.Info("Creating payment", "organizationId", req.OrganizationID, "invoiceId", req.InvoiceID)

	// Get invoice to validate payment
	invoice, err := uc.invoices.GetByID(ctx, req.OrganizationID, req.InvoiceID)
	if err != nil {
		if errors.Is(err, domain.ErrInvoiceNotFound) {
			uc.logger.Warn("Invoice not found for payment", "invoiceId", req.InvoiceID, "organizationId", req.OrganizationID)
			return nil, domain.ErrInvoiceNotFound
		}
		uc.logger.Error("Failed to get invoice for payment", "error", err, "invoiceId", req.InvoiceID)
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	// Validate payment amount
	if req.Amount > invoice.BalanceDue {
		uc.logger.Warn("Payment amount exceeds balance due", "amount", req.Amount, "balanceDue", invoice.BalanceDue)
		return nil, domain.ErrInsufficientPayment
	}

	// Create payment domain entity
	payment, err := domain.NewPayment(
		req.OrganizationID, req.InvoiceID, req.CreatedBy,
		req.PaymentMethod, req.Amount, req.Currency, req.PaymentDate,
	)
	if err != nil {
		uc.logger.Error("Failed to create payment domain entity", "error", err)
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Update additional properties
	if err := payment.UpdateBasicInfo(req.PaymentMethod, req.ReferenceNumber, req.Amount, req.PaymentDate, req.Notes); err != nil {
		uc.logger.Error("Failed to update payment basic info", "error", err)
		return nil, fmt.Errorf("failed to update payment info: %w", err)
	}

	// Persist payment
	if err := uc.invoices.CreatePayment(ctx, payment); err != nil {
		uc.logger.Error("Failed to persist payment", "error", err)
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	uc.logger.Info("Payment created successfully", "paymentId", payment.ID, "invoiceId", req.InvoiceID)
	return payment, nil
}

// GetInvoiceStats retrieves invoice statistics for an organization
func (uc *InvoiceUseCase) GetInvoiceStats(ctx context.Context, organizationID uint) (*repository.InvoiceStats, error) {
	uc.logger.Info("Getting invoice statistics", "organizationId", organizationID)

	stats, err := uc.invoices.GetInvoiceStats(ctx, organizationID)
	if err != nil {
		uc.logger.Error("Failed to get invoice statistics", "error", err, "organizationId", organizationID)
		return nil, fmt.Errorf("failed to get invoice statistics: %w", err)
	}

	uc.logger.Info("Invoice statistics retrieved successfully", "organizationId", organizationID, "totalInvoices", stats.TotalInvoices)
	return stats, nil
}

// GetOverdueInvoices retrieves all overdue invoices for an organization
func (uc *InvoiceUseCase) GetOverdueInvoices(ctx context.Context, organizationID uint) ([]*domain.Invoice, error) {
	uc.logger.Info("Getting overdue invoices", "organizationId", organizationID)

	invoices, err := uc.invoices.GetOverdueInvoices(ctx, organizationID)
	if err != nil {
		uc.logger.Error("Failed to get overdue invoices", "error", err, "organizationId", organizationID)
		return nil, fmt.Errorf("failed to get overdue invoices: %w", err)
	}

	uc.logger.Info("Overdue invoices retrieved successfully", "organizationId", organizationID, "count", len(invoices))
	return invoices, nil
}
