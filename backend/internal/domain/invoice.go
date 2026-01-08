// @kthulu:module:invoices
package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

// Domain errors for invoice module
var (
	ErrInvoiceNotFound      = errors.New("invoice not found")
	ErrInvoiceAlreadyExists = errors.New("invoice already exists")
	ErrInvalidInvoiceNumber = errors.New("invalid invoice number")
	ErrInvalidInvoiceType   = errors.New("invalid invoice type")
	ErrInvalidInvoiceStatus = errors.New("invalid invoice status")
	ErrInvoiceItemNotFound  = errors.New("invoice item not found")
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrInvalidPaymentMethod = errors.New("invalid payment method")
	ErrInvalidAmount        = errors.New("invalid amount")
	ErrInvoiceNotEditable   = errors.New("invoice is not editable")
	ErrInsufficientPayment  = errors.New("payment amount exceeds balance due")
)

// InvoiceType represents the type of invoice
type InvoiceType string

const (
	InvoiceTypeInvoice    InvoiceType = "invoice"
	InvoiceTypeQuote      InvoiceType = "quote"
	InvoiceTypeCreditNote InvoiceType = "credit_note"
	InvoiceTypeProforma   InvoiceType = "proforma"
)

// InvoiceStatus represents the status of an invoice
type InvoiceStatus string

const (
	InvoiceStatusDraft     InvoiceStatus = "draft"
	InvoiceStatusSent      InvoiceStatus = "sent"
	InvoiceStatusViewed    InvoiceStatus = "viewed"
	InvoiceStatusPartial   InvoiceStatus = "partial"
	InvoiceStatusPaid      InvoiceStatus = "paid"
	InvoiceStatusOverdue   InvoiceStatus = "overdue"
	InvoiceStatusCancelled InvoiceStatus = "canceled"
)

// PaymentMethod represents the method of payment
type PaymentMethod string

const (
	PaymentMethodCash         PaymentMethod = "cash"
	PaymentMethodCheck        PaymentMethod = "check"
	PaymentMethodCreditCard   PaymentMethod = "credit_card"
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
	PaymentMethodPayPal       PaymentMethod = "paypal"
	PaymentMethodStripe       PaymentMethod = "stripe"
	PaymentMethodOther        PaymentMethod = "other"
)

// Invoice represents an invoice in the system
type Invoice struct {
	ID              uint          `json:"id"`
	OrganizationID  uint          `json:"organizationId" validate:"required"`
	ContactID       uint          `json:"contactId" validate:"required"`
	InvoiceNumber   string        `json:"invoiceNumber" validate:"required,min=1,max=50"`
	Type            InvoiceType   `json:"type" validate:"required,oneof=invoice quote credit_note proforma"`
	Status          InvoiceStatus `json:"status" validate:"required,oneof=draft sent viewed partial paid overdue canceled"`
	Currency        string        `json:"currency" validate:"required,len=3"`
	ExchangeRate    float64       `json:"exchangeRate" validate:"min=0"`
	Subtotal        float64       `json:"subtotal" validate:"min=0"`
	TaxAmount       float64       `json:"taxAmount" validate:"min=0"`
	DiscountAmount  float64       `json:"discountAmount" validate:"min=0"`
	TotalAmount     float64       `json:"totalAmount" validate:"min=0"`
	PaidAmount      float64       `json:"paidAmount" validate:"min=0"`
	BalanceDue      float64       `json:"balanceDue"`
	IssueDate       time.Time     `json:"issueDate" validate:"required"`
	DueDate         *time.Time    `json:"dueDate,omitempty"`
	PaymentTerms    string        `json:"paymentTerms,omitempty" validate:"max=50"`
	Notes           string        `json:"notes,omitempty"`
	TermsConditions string        `json:"termsConditions,omitempty"`
	CreatedBy       uint          `json:"createdBy" validate:"required"`
	CreatedAt       time.Time     `json:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt"`

	// Related entities (loaded separately)
	Items    []InvoiceItem `json:"items,omitempty"`
	Payments []Payment     `json:"payments,omitempty"`
}

// InvoiceItem represents an item in an invoice
type InvoiceItem struct {
	ID               uint      `json:"id"`
	InvoiceID        uint      `json:"invoiceId" validate:"required"`
	ProductID        *uint     `json:"productId,omitempty"`
	ProductVariantID *uint     `json:"productVariantId,omitempty"`
	Description      string    `json:"description" validate:"required,min=1,max=500"`
	Quantity         float64   `json:"quantity" validate:"required,min=0"`
	UnitPrice        float64   `json:"unitPrice" validate:"required,min=0"`
	DiscountPercent  float64   `json:"discountPercent" validate:"min=0,max=1"`
	DiscountAmount   float64   `json:"discountAmount" validate:"min=0"`
	TaxRate          float64   `json:"taxRate" validate:"min=0,max=1"`
	TaxAmount        float64   `json:"taxAmount" validate:"min=0"`
	LineTotal        float64   `json:"lineTotal" validate:"min=0"`
	SortOrder        int       `json:"sortOrder"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// Payment represents a payment made against an invoice
type Payment struct {
	ID              uint          `json:"id"`
	OrganizationID  uint          `json:"organizationId" validate:"required"`
	InvoiceID       uint          `json:"invoiceId" validate:"required"`
	PaymentMethod   PaymentMethod `json:"paymentMethod" validate:"required,oneof=cash check credit_card bank_transfer paypal stripe other"`
	ReferenceNumber string        `json:"referenceNumber,omitempty" validate:"max=100"`
	Amount          float64       `json:"amount" validate:"required,min=0"`
	Currency        string        `json:"currency" validate:"required,len=3"`
	ExchangeRate    float64       `json:"exchangeRate" validate:"min=0"`
	PaymentDate     time.Time     `json:"paymentDate" validate:"required"`
	Notes           string        `json:"notes,omitempty"`
	CreatedBy       uint          `json:"createdBy" validate:"required"`
	CreatedAt       time.Time     `json:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt"`
}

// NewInvoice creates a new invoice with validation
func NewInvoice(organizationID, contactID, createdBy uint, invoiceNumber string, invoiceType InvoiceType, currency string, issueDate time.Time) (*Invoice, error) {
	invoice := &Invoice{
		OrganizationID: organizationID,
		ContactID:      contactID,
		InvoiceNumber:  strings.TrimSpace(invoiceNumber),
		Type:           invoiceType,
		Status:         InvoiceStatusDraft,
		Currency:       strings.ToUpper(strings.TrimSpace(currency)),
		ExchangeRate:   1.0,
		Subtotal:       0.0,
		TaxAmount:      0.0,
		DiscountAmount: 0.0,
		TotalAmount:    0.0,
		PaidAmount:     0.0,
		BalanceDue:     0.0,
		IssueDate:      issueDate,
		CreatedBy:      createdBy,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := invoice.Validate(); err != nil {
		return nil, err
	}

	return invoice, nil
}

// Validate validates the invoice data
func (i *Invoice) Validate() error {
	validate := validator.New()
	if err := validate.Struct(i); err != nil {
		return err
	}

	// Business rule validations
	if i.InvoiceNumber == "" {
		return ErrInvalidInvoiceNumber
	}

	if i.Currency == "" || len(i.Currency) != 3 {
		return errors.New("currency must be a 3-letter code")
	}

	if i.DueDate != nil && i.DueDate.Before(i.IssueDate) {
		return errors.New("due date cannot be before issue date")
	}

	if i.PaidAmount > i.TotalAmount {
		return errors.New("paid amount cannot exceed total amount")
	}

	return nil
}

// UpdateBasicInfo updates the basic invoice information
func (i *Invoice) UpdateBasicInfo(contactID uint, dueDate *time.Time, paymentTerms, notes, termsConditions string) error {
	i.ContactID = contactID
	i.DueDate = dueDate
	i.PaymentTerms = strings.TrimSpace(paymentTerms)
	i.Notes = strings.TrimSpace(notes)
	i.TermsConditions = strings.TrimSpace(termsConditions)
	i.UpdatedAt = time.Now()

	return i.Validate()
}

// SetStatus sets the invoice status
func (i *Invoice) SetStatus(status InvoiceStatus) error {
	// Business rules for status transitions
	switch i.Status {
	case InvoiceStatusCancelled:
		return errors.New("cannot change status of canceled invoice")
	case InvoiceStatusPaid:
		if status != InvoiceStatusPaid {
			return errors.New("cannot change status of paid invoice")
		}
	}

	i.Status = status
	i.UpdatedAt = time.Now()
	return nil
}

// CanEdit returns true if the invoice can be edited
func (i *Invoice) CanEdit() bool {
	return i.Status == InvoiceStatusDraft
}

// IsOverdue returns true if the invoice is overdue
func (i *Invoice) IsOverdue() bool {
	if i.DueDate == nil || i.Status == InvoiceStatusPaid || i.Status == InvoiceStatusCancelled {
		return false
	}
	return time.Now().After(*i.DueDate) && i.BalanceDue > 0
}

// CalculateTotals recalculates the invoice totals based on items
func (i *Invoice) CalculateTotals() {
	i.Subtotal = 0.0
	i.TaxAmount = 0.0
	i.DiscountAmount = 0.0

	for _, item := range i.Items {
		i.Subtotal += (item.LineTotal - item.TaxAmount)
		i.TaxAmount += item.TaxAmount
		i.DiscountAmount += item.DiscountAmount
	}

	i.TotalAmount = i.Subtotal + i.TaxAmount - i.DiscountAmount
	i.BalanceDue = i.TotalAmount - i.PaidAmount
	i.UpdatedAt = time.Now()
}

// AddItem adds an item to the invoice
func (i *Invoice) AddItem(item *InvoiceItem) error {
	if !i.CanEdit() {
		return ErrInvoiceNotEditable
	}

	item.InvoiceID = i.ID
	item.SortOrder = len(i.Items)
	item.CalculateLineTotal()

	i.Items = append(i.Items, *item)
	i.CalculateTotals()

	return nil
}

// RemoveItem removes an item from the invoice
func (i *Invoice) RemoveItem(itemID uint) error {
	if !i.CanEdit() {
		return ErrInvoiceNotEditable
	}

	for idx, item := range i.Items {
		if item.ID == itemID {
			i.Items = append(i.Items[:idx], i.Items[idx+1:]...)
			i.CalculateTotals()
			return nil
		}
	}

	return ErrInvoiceItemNotFound
}

// NewInvoiceItem creates a new invoice item with validation
func NewInvoiceItem(invoiceID uint, description string, quantity, unitPrice float64) (*InvoiceItem, error) {
	item := &InvoiceItem{
		InvoiceID:       invoiceID,
		Description:     strings.TrimSpace(description),
		Quantity:        quantity,
		UnitPrice:       unitPrice,
		DiscountPercent: 0.0,
		DiscountAmount:  0.0,
		TaxRate:         0.0,
		TaxAmount:       0.0,
		LineTotal:       0.0,
		SortOrder:       0,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	item.CalculateLineTotal()

	if err := item.Validate(); err != nil {
		return nil, err
	}

	return item, nil
}

// Validate validates the invoice item data
func (ii *InvoiceItem) Validate() error {
	validate := validator.New()
	if err := validate.Struct(ii); err != nil {
		return err
	}

	if ii.Description == "" {
		return errors.New("description is required")
	}

	if ii.Quantity <= 0 {
		return errors.New("quantity must be greater than zero")
	}

	if ii.UnitPrice < 0 {
		return errors.New("unit price cannot be negative")
	}

	return nil
}

// UpdateBasicInfo updates the basic item information
func (ii *InvoiceItem) UpdateBasicInfo(productID, productVariantID *uint, description string, quantity, unitPrice, discountPercent, taxRate float64) error {
	ii.ProductID = productID
	ii.ProductVariantID = productVariantID
	ii.Description = strings.TrimSpace(description)
	ii.Quantity = quantity
	ii.UnitPrice = unitPrice
	ii.DiscountPercent = discountPercent
	ii.TaxRate = taxRate
	ii.UpdatedAt = time.Now()

	ii.CalculateLineTotal()
	return ii.Validate()
}

// CalculateLineTotal calculates the line total for the item
func (ii *InvoiceItem) CalculateLineTotal() {
	subtotal := ii.Quantity * ii.UnitPrice

	// Calculate discount
	if ii.DiscountPercent > 0 {
		ii.DiscountAmount = subtotal * ii.DiscountPercent
	}

	discountedSubtotal := subtotal - ii.DiscountAmount

	// Calculate tax
	if ii.TaxRate > 0 {
		ii.TaxAmount = discountedSubtotal * ii.TaxRate
	}

	ii.LineTotal = discountedSubtotal + ii.TaxAmount
}

// NewPayment creates a new payment with validation
func NewPayment(organizationID, invoiceID, createdBy uint, method PaymentMethod, amount float64, currency string, paymentDate time.Time) (*Payment, error) {
	payment := &Payment{
		OrganizationID: organizationID,
		InvoiceID:      invoiceID,
		PaymentMethod:  method,
		Amount:         amount,
		Currency:       strings.ToUpper(strings.TrimSpace(currency)),
		ExchangeRate:   1.0,
		PaymentDate:    paymentDate,
		CreatedBy:      createdBy,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := payment.Validate(); err != nil {
		return nil, err
	}

	return payment, nil
}

// Validate validates the payment data
func (p *Payment) Validate() error {
	validate := validator.New()
	if err := validate.Struct(p); err != nil {
		return err
	}

	if p.Amount <= 0 {
		return ErrInvalidAmount
	}

	if p.Currency == "" || len(p.Currency) != 3 {
		return errors.New("currency must be a 3-letter code")
	}

	return nil
}

// UpdateBasicInfo updates the basic payment information
func (p *Payment) UpdateBasicInfo(method PaymentMethod, referenceNumber string, amount float64, paymentDate time.Time, notes string) error {
	p.PaymentMethod = method
	p.ReferenceNumber = strings.TrimSpace(referenceNumber)
	p.Amount = amount
	p.PaymentDate = paymentDate
	p.Notes = strings.TrimSpace(notes)
	p.UpdatedAt = time.Now()

	return p.Validate()
}

// GetDisplayReference returns a display-friendly reference for the payment
func (p *Payment) GetDisplayReference() string {
	if p.ReferenceNumber != "" {
		return p.ReferenceNumber
	}
	return string(p.PaymentMethod)
}
