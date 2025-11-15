// @kthulu:module:invoices
package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/pmaojo/kthulu-go/backend/core"
	"github.com/pmaojo/kthulu-go/backend/internal/domain"
	"github.com/pmaojo/kthulu-go/backend/internal/repository"
)

const invoiceColumns = "id, organization_id, contact_id, invoice_number, type, status, currency, exchange_rate, subtotal, tax_amount, discount_amount, total_amount, paid_amount, balance_due, issue_date, due_date, payment_terms, notes, terms_conditions, created_by, created_at, updated_at"

type scanner interface {
	Scan(dest ...any) error
}

func scanInvoice(s scanner, invoice *domain.Invoice) error {
	return s.Scan(
		&invoice.ID, &invoice.OrganizationID, &invoice.ContactID,
		&invoice.InvoiceNumber, &invoice.Type, &invoice.Status,
		&invoice.Currency, &invoice.ExchangeRate, &invoice.Subtotal,
		&invoice.TaxAmount, &invoice.DiscountAmount, &invoice.TotalAmount,
		&invoice.PaidAmount, &invoice.BalanceDue, &invoice.IssueDate,
		&invoice.DueDate, &invoice.PaymentTerms, &invoice.Notes,
		&invoice.TermsConditions, &invoice.CreatedBy, &invoice.CreatedAt, &invoice.UpdatedAt,
	)
}

// InvoiceRepository implements the invoice repository interface using SQL
type InvoiceRepository struct {
	db     *sql.DB
	logger core.Logger
}

// NewInvoiceRepository creates a new invoice repository instance
func NewInvoiceRepository(db *sql.DB, logger core.Logger) repository.InvoiceRepository {
	return &InvoiceRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new invoice
func (r *InvoiceRepository) Create(ctx context.Context, invoice *domain.Invoice) error {
	query := fmt.Sprintf(`
                INSERT INTO invoices (
                        organization_id, contact_id, invoice_number, type, status, currency,
                        exchange_rate, subtotal, tax_amount, discount_amount, total_amount,
                        paid_amount, balance_due, issue_date, due_date, payment_terms,
                        notes, terms_conditions, created_by, created_at, updated_at
                ) VALUES (
                        $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21
                ) RETURNING %s`, invoiceColumns)

	row := r.db.QueryRowContext(ctx, query,
		invoice.OrganizationID, invoice.ContactID, invoice.InvoiceNumber,
		invoice.Type, invoice.Status, invoice.Currency, invoice.ExchangeRate,
		invoice.Subtotal, invoice.TaxAmount, invoice.DiscountAmount,
		invoice.TotalAmount, invoice.PaidAmount, invoice.BalanceDue,
		invoice.IssueDate, invoice.DueDate, invoice.PaymentTerms,
		invoice.Notes, invoice.TermsConditions, invoice.CreatedBy,
		invoice.CreatedAt, invoice.UpdatedAt,
	)
	err := scanInvoice(row, invoice)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return domain.ErrInvoiceAlreadyExists
		}
		r.logger.Error("Failed to create invoice", "error", err, "invoiceNumber", invoice.InvoiceNumber)
		return fmt.Errorf("failed to create invoice: %w", err)
	}

	r.logger.Info("Invoice created successfully", "invoiceId", invoice.ID, "invoiceNumber", invoice.InvoiceNumber)
	return nil
}

// GetByID retrieves an invoice by ID within an organization
func (r *InvoiceRepository) GetByID(ctx context.Context, organizationID, invoiceID uint) (*domain.Invoice, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM invoices WHERE id = $1 AND organization_id = $2",
		invoiceColumns,
	)

	invoice := &domain.Invoice{}
	err := scanInvoice(r.db.QueryRowContext(ctx, query, invoiceID, organizationID), invoice)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrInvoiceNotFound
		}
		r.logger.Error("Failed to get invoice by ID", "error", err, "invoiceId", invoiceID)
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	return invoice, nil
}

// GetByNumber retrieves an invoice by number within an organization
func (r *InvoiceRepository) GetByNumber(ctx context.Context, organizationID uint, invoiceNumber string) (*domain.Invoice, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM invoices WHERE invoice_number = $1 AND organization_id = $2",
		invoiceColumns,
	)

	invoice := &domain.Invoice{}
	err := scanInvoice(r.db.QueryRowContext(ctx, query, invoiceNumber, organizationID), invoice)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrInvoiceNotFound
		}
		r.logger.Error("Failed to get invoice by number", "error", err, "invoiceNumber", invoiceNumber)
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	return invoice, nil
}

// Update updates an existing invoice
func (r *InvoiceRepository) Update(ctx context.Context, invoice *domain.Invoice) error {
	query := `
		UPDATE invoices SET 
			contact_id = $2, type = $3, status = $4, currency = $5,
			exchange_rate = $6, subtotal = $7, tax_amount = $8, discount_amount = $9,
			total_amount = $10, paid_amount = $11, balance_due = $12,
			issue_date = $13, due_date = $14, payment_terms = $15,
			notes = $16, terms_conditions = $17, updated_at = $18
		WHERE id = $1 AND organization_id = $19`

	result, err := r.db.ExecContext(ctx, query,
		invoice.ID, invoice.ContactID, invoice.Type, invoice.Status,
		invoice.Currency, invoice.ExchangeRate, invoice.Subtotal,
		invoice.TaxAmount, invoice.DiscountAmount, invoice.TotalAmount,
		invoice.PaidAmount, invoice.BalanceDue, invoice.IssueDate,
		invoice.DueDate, invoice.PaymentTerms, invoice.Notes,
		invoice.TermsConditions, time.Now(), invoice.OrganizationID,
	)

	if err != nil {
		r.logger.Error("Failed to update invoice", "error", err, "invoiceId", invoice.ID)
		return fmt.Errorf("failed to update invoice: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrInvoiceNotFound
	}

	r.logger.Info("Invoice updated successfully", "invoiceId", invoice.ID)
	return nil
}

// Delete deletes an invoice
func (r *InvoiceRepository) Delete(ctx context.Context, organizationID, invoiceID uint) error {
	query := `DELETE FROM invoices WHERE id = $1 AND organization_id = $2`

	result, err := r.db.ExecContext(ctx, query, invoiceID, organizationID)
	if err != nil {
		r.logger.Error("Failed to delete invoice", "error", err, "invoiceId", invoiceID)
		return fmt.Errorf("failed to delete invoice: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrInvoiceNotFound
	}

	r.logger.Info("Invoice deleted successfully", "invoiceId", invoiceID)
	return nil
}

// List retrieves invoices with filtering and pagination
func (r *InvoiceRepository) List(ctx context.Context, organizationID uint, filters repository.InvoiceFilters) ([]*domain.Invoice, int64, error) {
	// Validate filters
	if err := filters.Validate(); err != nil {
		return nil, 0, err
	}

	// Build WHERE clause
	whereClause, args := r.buildInvoiceWhereClause(organizationID, filters)

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM invoices %s", whereClause)
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		r.logger.Error("Failed to count invoices", "error", err)
		return nil, 0, fmt.Errorf("failed to count invoices: %w", err)
	}

	// Main query with pagination and sorting
	orderClause := fmt.Sprintf("ORDER BY %s %s", filters.SortBy, strings.ToUpper(filters.SortOrder))
	limitClause := fmt.Sprintf("LIMIT %d OFFSET %d", filters.PageSize, filters.GetOffset())

	query := fmt.Sprintf(
		"SELECT %s FROM invoices %s %s %s",
		invoiceColumns, whereClause, orderClause, limitClause,
	)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to list invoices", "error", err)
		return nil, 0, fmt.Errorf("failed to list invoices: %w", err)
	}
	defer rows.Close()

	var invoices []*domain.Invoice
	for rows.Next() {
		invoice := &domain.Invoice{}
		err := scanInvoice(rows, invoice)
		if err != nil {
			r.logger.Error("Failed to scan invoice", "error", err)
			return nil, 0, fmt.Errorf("failed to scan invoice: %w", err)
		}

		// Load related data if requested
		if filters.IncludeItems {
			items, err := r.GetItemsByInvoiceID(ctx, invoice.ID)
			if err != nil {
				r.logger.Error("Failed to load invoice items", "error", err, "invoiceId", invoice.ID)
			} else {
				// Convert []*domain.InvoiceItem to []domain.InvoiceItem
				invoice.Items = make([]domain.InvoiceItem, len(items))
				for i, item := range items {
					invoice.Items[i] = *item
				}
			}
		}

		if filters.IncludePayments {
			payments, err := r.GetPaymentsByInvoiceID(ctx, invoice.ID)
			if err != nil {
				r.logger.Error("Failed to load invoice payments", "error", err, "invoiceId", invoice.ID)
			} else {
				// Convert []*domain.Payment to []domain.Payment
				invoice.Payments = make([]domain.Payment, len(payments))
				for i, payment := range payments {
					invoice.Payments[i] = *payment
				}
			}
		}

		invoices = append(invoices, invoice)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate invoices: %w", err)
	}

	return invoices, total, nil
}

// buildInvoiceWhereClause builds the WHERE clause for invoice filtering
func (r *InvoiceRepository) buildInvoiceWhereClause(organizationID uint, filters repository.InvoiceFilters) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	// Organization filter (always required)
	conditions = append(conditions, fmt.Sprintf("organization_id = $%d", argIndex))
	args = append(args, organizationID)
	argIndex++

	// Contact filter
	if filters.ContactID != nil {
		conditions = append(conditions, fmt.Sprintf("contact_id = $%d", argIndex))
		args = append(args, *filters.ContactID)
		argIndex++
	}

	// Type filter
	if filters.Type != nil {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, *filters.Type)
		argIndex++
	}

	// Status filter
	if filters.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *filters.Status)
		argIndex++
	}

	// Currency filter
	if filters.Currency != "" {
		conditions = append(conditions, fmt.Sprintf("currency = $%d", argIndex))
		args = append(args, filters.Currency)
		argIndex++
	}

	// Search filter (invoice number)
	if filters.Search != "" {
		searchPattern := "%" + strings.ToLower(filters.Search) + "%"
		conditions = append(conditions, fmt.Sprintf("LOWER(invoice_number) LIKE $%d", argIndex))
		args = append(args, searchPattern)
		argIndex++
	}

	// Date range filters
	if filters.IssuedFrom != nil {
		conditions = append(conditions, fmt.Sprintf("issue_date >= $%d", argIndex))
		args = append(args, *filters.IssuedFrom)
		argIndex++
	}

	if filters.IssuedTo != nil {
		conditions = append(conditions, fmt.Sprintf("issue_date <= $%d", argIndex))
		args = append(args, *filters.IssuedTo)
		argIndex++
	}

	if filters.DueFrom != nil {
		conditions = append(conditions, fmt.Sprintf("due_date >= $%d", argIndex))
		args = append(args, *filters.DueFrom)
		argIndex++
	}

	if filters.DueTo != nil {
		conditions = append(conditions, fmt.Sprintf("due_date <= $%d", argIndex))
		args = append(args, *filters.DueTo)
		argIndex++
	}

	// Amount range filters
	if filters.MinAmount != nil {
		conditions = append(conditions, fmt.Sprintf("total_amount >= $%d", argIndex))
		args = append(args, *filters.MinAmount)
		argIndex++
	}

	if filters.MaxAmount != nil {
		conditions = append(conditions, fmt.Sprintf("total_amount <= $%d", argIndex))
		args = append(args, *filters.MaxAmount)
		argIndex++
	}

	// Overdue filter
	if filters.IsOverdue != nil && *filters.IsOverdue {
		conditions = append(conditions, "due_date < NOW() AND balance_due > 0 AND status NOT IN ('paid', 'canceled')")
	}

	// Created by filter
	if filters.CreatedBy != nil {
		conditions = append(conditions, fmt.Sprintf("created_by = $%d", argIndex))
		args = append(args, *filters.CreatedBy)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	return whereClause, args
}

// CreateItem creates a new invoice item
func (r *InvoiceRepository) CreateItem(ctx context.Context, item *domain.InvoiceItem) error {
	query := `
		INSERT INTO invoice_items (
			invoice_id, product_id, product_variant_id, description, quantity,
			unit_price, discount_percent, discount_amount, tax_rate, tax_amount,
			line_total, sort_order, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		) RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		item.InvoiceID, item.ProductID, item.ProductVariantID, item.Description,
		item.Quantity, item.UnitPrice, item.DiscountPercent, item.DiscountAmount,
		item.TaxRate, item.TaxAmount, item.LineTotal, item.SortOrder,
		item.CreatedAt, item.UpdatedAt,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)

	if err != nil {
		r.logger.Error("Failed to create invoice item", "error", err, "invoiceId", item.InvoiceID)
		return fmt.Errorf("failed to create invoice item: %w", err)
	}

	r.logger.Info("Invoice item created successfully", "itemId", item.ID, "invoiceId", item.InvoiceID)
	return nil
}

// GetItemByID retrieves an invoice item by ID
func (r *InvoiceRepository) GetItemByID(ctx context.Context, invoiceID, itemID uint) (*domain.InvoiceItem, error) {
	query := `
		SELECT id, invoice_id, product_id, product_variant_id, description,
			   quantity, unit_price, discount_percent, discount_amount,
			   tax_rate, tax_amount, line_total, sort_order, created_at, updated_at
		FROM invoice_items 
		WHERE id = $1 AND invoice_id = $2`

	item := &domain.InvoiceItem{}
	err := r.db.QueryRowContext(ctx, query, itemID, invoiceID).Scan(
		&item.ID, &item.InvoiceID, &item.ProductID, &item.ProductVariantID,
		&item.Description, &item.Quantity, &item.UnitPrice, &item.DiscountPercent,
		&item.DiscountAmount, &item.TaxRate, &item.TaxAmount, &item.LineTotal,
		&item.SortOrder, &item.CreatedAt, &item.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrInvoiceItemNotFound
		}
		r.logger.Error("Failed to get invoice item by ID", "error", err, "itemId", itemID)
		return nil, fmt.Errorf("failed to get invoice item: %w", err)
	}

	return item, nil
}

// GetItemsByInvoiceID retrieves all items for an invoice
func (r *InvoiceRepository) GetItemsByInvoiceID(ctx context.Context, invoiceID uint) ([]*domain.InvoiceItem, error) {
	query := `
		SELECT id, invoice_id, product_id, product_variant_id, description,
			   quantity, unit_price, discount_percent, discount_amount,
			   tax_rate, tax_amount, line_total, sort_order, created_at, updated_at
		FROM invoice_items 
		WHERE invoice_id = $1
		ORDER BY sort_order ASC, id ASC`

	rows, err := r.db.QueryContext(ctx, query, invoiceID)
	if err != nil {
		r.logger.Error("Failed to get invoice items", "error", err, "invoiceId", invoiceID)
		return nil, fmt.Errorf("failed to get invoice items: %w", err)
	}
	defer rows.Close()

	var items []*domain.InvoiceItem
	for rows.Next() {
		item := &domain.InvoiceItem{}
		err := rows.Scan(
			&item.ID, &item.InvoiceID, &item.ProductID, &item.ProductVariantID,
			&item.Description, &item.Quantity, &item.UnitPrice, &item.DiscountPercent,
			&item.DiscountAmount, &item.TaxRate, &item.TaxAmount, &item.LineTotal,
			&item.SortOrder, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan invoice item", "error", err)
			return nil, fmt.Errorf("failed to scan invoice item: %w", err)
		}

		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate invoice items: %w", err)
	}

	return items, nil
}

// UpdateItem updates an existing invoice item
func (r *InvoiceRepository) UpdateItem(ctx context.Context, item *domain.InvoiceItem) error {
	query := `
		UPDATE invoice_items SET 
			product_id = $2, product_variant_id = $3, description = $4,
			quantity = $5, unit_price = $6, discount_percent = $7,
			discount_amount = $8, tax_rate = $9, tax_amount = $10,
			line_total = $11, sort_order = $12, updated_at = $13
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		item.ID, item.ProductID, item.ProductVariantID, item.Description,
		item.Quantity, item.UnitPrice, item.DiscountPercent, item.DiscountAmount,
		item.TaxRate, item.TaxAmount, item.LineTotal, item.SortOrder, time.Now(),
	)

	if err != nil {
		r.logger.Error("Failed to update invoice item", "error", err, "itemId", item.ID)
		return fmt.Errorf("failed to update invoice item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrInvoiceItemNotFound
	}

	r.logger.Info("Invoice item updated successfully", "itemId", item.ID)
	return nil
}

// DeleteItem deletes an invoice item
func (r *InvoiceRepository) DeleteItem(ctx context.Context, invoiceID, itemID uint) error {
	query := `DELETE FROM invoice_items WHERE id = $1 AND invoice_id = $2`

	result, err := r.db.ExecContext(ctx, query, itemID, invoiceID)
	if err != nil {
		r.logger.Error("Failed to delete invoice item", "error", err, "itemId", itemID)
		return fmt.Errorf("failed to delete invoice item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrInvoiceItemNotFound
	}

	r.logger.Info("Invoice item deleted successfully", "itemId", itemID)
	return nil
}

// BulkCreateItems creates multiple invoice items in a single transaction
func (r *InvoiceRepository) BulkCreateItems(ctx context.Context, items []*domain.InvoiceItem) error {
	if len(items) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO invoice_items (
			invoice_id, product_id, product_variant_id, description, quantity,
			unit_price, discount_percent, discount_amount, tax_rate, tax_amount,
			line_total, sort_order, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		) RETURNING id, created_at, updated_at`

	for _, item := range items {
		err := tx.QueryRowContext(ctx, query,
			item.InvoiceID, item.ProductID, item.ProductVariantID, item.Description,
			item.Quantity, item.UnitPrice, item.DiscountPercent, item.DiscountAmount,
			item.TaxRate, item.TaxAmount, item.LineTotal, item.SortOrder,
			item.CreatedAt, item.UpdatedAt,
		).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)

		if err != nil {
			r.logger.Error("Failed to bulk create invoice item", "error", err, "invoiceId", item.InvoiceID)
			return fmt.Errorf("failed to bulk create invoice item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit bulk create transaction: %w", err)
	}

	r.logger.Info("Bulk created invoice items successfully", "count", len(items))
	return nil
}

// BulkUpdateItems updates multiple invoice items in a single transaction
func (r *InvoiceRepository) BulkUpdateItems(ctx context.Context, items []*domain.InvoiceItem) error {
	if len(items) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		UPDATE invoice_items SET 
			product_id = $2, product_variant_id = $3, description = $4,
			quantity = $5, unit_price = $6, discount_percent = $7,
			discount_amount = $8, tax_rate = $9, tax_amount = $10,
			line_total = $11, sort_order = $12, updated_at = $13
		WHERE id = $1`

	for _, item := range items {
		result, err := tx.ExecContext(ctx, query,
			item.ID, item.ProductID, item.ProductVariantID, item.Description,
			item.Quantity, item.UnitPrice, item.DiscountPercent, item.DiscountAmount,
			item.TaxRate, item.TaxAmount, item.LineTotal, item.SortOrder, time.Now(),
		)

		if err != nil {
			r.logger.Error("Failed to bulk update invoice item", "error", err, "itemId", item.ID)
			return fmt.Errorf("failed to bulk update invoice item %d: %w", item.ID, err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected for item %d: %w", item.ID, err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("invoice item %d not found", item.ID)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit bulk update transaction: %w", err)
	}

	r.logger.Info("Bulk updated invoice items successfully", "count", len(items))
	return nil
}

// BulkDeleteItems deletes multiple invoice items in a single transaction
func (r *InvoiceRepository) BulkDeleteItems(ctx context.Context, invoiceID uint, itemIDs []uint) error {
	if len(itemIDs) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Build placeholders for IN clause
	placeholders := make([]string, len(itemIDs))
	args := make([]interface{}, len(itemIDs)+1)
	args[0] = invoiceID

	for i, id := range itemIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = id
	}

	query := fmt.Sprintf("DELETE FROM invoice_items WHERE invoice_id = $1 AND id IN (%s)", strings.Join(placeholders, ","))

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to bulk delete invoice items", "error", err)
		return fmt.Errorf("failed to bulk delete invoice items: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit bulk delete transaction: %w", err)
	}

	r.logger.Info("Bulk deleted invoice items successfully", "count", rowsAffected, "requested", len(itemIDs))
	return nil
}

// CreatePayment creates a new payment
func (r *InvoiceRepository) CreatePayment(ctx context.Context, payment *domain.Payment) error {
	query := `
		INSERT INTO payments (
			organization_id, invoice_id, payment_method, reference_number,
			amount, currency, exchange_rate, payment_date, notes,
			created_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		) RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		payment.OrganizationID, payment.InvoiceID, payment.PaymentMethod,
		payment.ReferenceNumber, payment.Amount, payment.Currency,
		payment.ExchangeRate, payment.PaymentDate, payment.Notes,
		payment.CreatedBy, payment.CreatedAt, payment.UpdatedAt,
	).Scan(&payment.ID, &payment.CreatedAt, &payment.UpdatedAt)

	if err != nil {
		r.logger.Error("Failed to create payment", "error", err, "invoiceId", payment.InvoiceID)
		return fmt.Errorf("failed to create payment: %w", err)
	}

	r.logger.Info("Payment created successfully", "paymentId", payment.ID, "invoiceId", payment.InvoiceID)
	return nil
}

// GetPaymentByID retrieves a payment by ID
func (r *InvoiceRepository) GetPaymentByID(ctx context.Context, organizationID, paymentID uint) (*domain.Payment, error) {
	query := `
		SELECT id, organization_id, invoice_id, payment_method, reference_number,
			   amount, currency, exchange_rate, payment_date, notes,
			   created_by, created_at, updated_at
		FROM payments 
		WHERE id = $1 AND organization_id = $2`

	payment := &domain.Payment{}
	err := r.db.QueryRowContext(ctx, query, paymentID, organizationID).Scan(
		&payment.ID, &payment.OrganizationID, &payment.InvoiceID,
		&payment.PaymentMethod, &payment.ReferenceNumber, &payment.Amount,
		&payment.Currency, &payment.ExchangeRate, &payment.PaymentDate,
		&payment.Notes, &payment.CreatedBy, &payment.CreatedAt, &payment.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrPaymentNotFound
		}
		r.logger.Error("Failed to get payment by ID", "error", err, "paymentId", paymentID)
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	return payment, nil
}

// GetPaymentsByInvoiceID retrieves all payments for an invoice
func (r *InvoiceRepository) GetPaymentsByInvoiceID(ctx context.Context, invoiceID uint) ([]*domain.Payment, error) {
	query := `
		SELECT id, organization_id, invoice_id, payment_method, reference_number,
			   amount, currency, exchange_rate, payment_date, notes,
			   created_by, created_at, updated_at
		FROM payments 
		WHERE invoice_id = $1
		ORDER BY payment_date DESC, created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, invoiceID)
	if err != nil {
		r.logger.Error("Failed to get payments for invoice", "error", err, "invoiceId", invoiceID)
		return nil, fmt.Errorf("failed to get payments: %w", err)
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		payment := &domain.Payment{}
		err := rows.Scan(
			&payment.ID, &payment.OrganizationID, &payment.InvoiceID,
			&payment.PaymentMethod, &payment.ReferenceNumber, &payment.Amount,
			&payment.Currency, &payment.ExchangeRate, &payment.PaymentDate,
			&payment.Notes, &payment.CreatedBy, &payment.CreatedAt, &payment.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan payment", "error", err)
			return nil, fmt.Errorf("failed to scan payment: %w", err)
		}

		payments = append(payments, payment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate payments: %w", err)
	}

	return payments, nil
}

// UpdatePayment updates an existing payment
func (r *InvoiceRepository) UpdatePayment(ctx context.Context, payment *domain.Payment) error {
	query := `
		UPDATE payments SET 
			payment_method = $2, reference_number = $3, amount = $4,
			currency = $5, exchange_rate = $6, payment_date = $7,
			notes = $8, updated_at = $9
		WHERE id = $1 AND organization_id = $10`

	result, err := r.db.ExecContext(ctx, query,
		payment.ID, payment.PaymentMethod, payment.ReferenceNumber,
		payment.Amount, payment.Currency, payment.ExchangeRate,
		payment.PaymentDate, payment.Notes, time.Now(), payment.OrganizationID,
	)

	if err != nil {
		r.logger.Error("Failed to update payment", "error", err, "paymentId", payment.ID)
		return fmt.Errorf("failed to update payment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrPaymentNotFound
	}

	r.logger.Info("Payment updated successfully", "paymentId", payment.ID)
	return nil
}

// DeletePayment deletes a payment
func (r *InvoiceRepository) DeletePayment(ctx context.Context, organizationID, paymentID uint) error {
	query := `DELETE FROM payments WHERE id = $1 AND organization_id = $2`

	result, err := r.db.ExecContext(ctx, query, paymentID, organizationID)
	if err != nil {
		r.logger.Error("Failed to delete payment", "error", err, "paymentId", paymentID)
		return fmt.Errorf("failed to delete payment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrPaymentNotFound
	}

	r.logger.Info("Payment deleted successfully", "paymentId", paymentID)
	return nil
}

// ListPayments retrieves payments with filtering and pagination
func (r *InvoiceRepository) ListPayments(ctx context.Context, organizationID uint, filters repository.PaymentFilters) ([]*domain.Payment, int64, error) {
	// Validate filters
	if err := filters.Validate(); err != nil {
		return nil, 0, err
	}

	// Build WHERE clause
	whereClause, args := r.buildPaymentWhereClause(organizationID, filters)

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM payments %s", whereClause)
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		r.logger.Error("Failed to count payments", "error", err)
		return nil, 0, fmt.Errorf("failed to count payments: %w", err)
	}

	// Main query with pagination and sorting
	orderClause := fmt.Sprintf("ORDER BY %s %s", filters.SortBy, strings.ToUpper(filters.SortOrder))
	limitClause := fmt.Sprintf("LIMIT %d OFFSET %d", filters.PageSize, filters.GetOffset())

	query := fmt.Sprintf(`
		SELECT id, organization_id, invoice_id, payment_method, reference_number,
			   amount, currency, exchange_rate, payment_date, notes,
			   created_by, created_at, updated_at
		FROM payments %s %s %s`, whereClause, orderClause, limitClause)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to list payments", "error", err)
		return nil, 0, fmt.Errorf("failed to list payments: %w", err)
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		payment := &domain.Payment{}
		err := rows.Scan(
			&payment.ID, &payment.OrganizationID, &payment.InvoiceID,
			&payment.PaymentMethod, &payment.ReferenceNumber, &payment.Amount,
			&payment.Currency, &payment.ExchangeRate, &payment.PaymentDate,
			&payment.Notes, &payment.CreatedBy, &payment.CreatedAt, &payment.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan payment", "error", err)
			return nil, 0, fmt.Errorf("failed to scan payment: %w", err)
		}

		payments = append(payments, payment)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate payments: %w", err)
	}

	return payments, total, nil
}

// buildPaymentWhereClause builds the WHERE clause for payment filtering
func (r *InvoiceRepository) buildPaymentWhereClause(organizationID uint, filters repository.PaymentFilters) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	// Organization filter (always required)
	conditions = append(conditions, fmt.Sprintf("organization_id = $%d", argIndex))
	args = append(args, organizationID)
	argIndex++

	// Invoice filter
	if filters.InvoiceID != nil {
		conditions = append(conditions, fmt.Sprintf("invoice_id = $%d", argIndex))
		args = append(args, *filters.InvoiceID)
		argIndex++
	}

	// Payment method filter
	if filters.PaymentMethod != nil {
		conditions = append(conditions, fmt.Sprintf("payment_method = $%d", argIndex))
		args = append(args, *filters.PaymentMethod)
		argIndex++
	}

	// Currency filter
	if filters.Currency != "" {
		conditions = append(conditions, fmt.Sprintf("currency = $%d", argIndex))
		args = append(args, filters.Currency)
		argIndex++
	}

	// Search filter (reference number)
	if filters.Search != "" {
		searchPattern := "%" + strings.ToLower(filters.Search) + "%"
		conditions = append(conditions, fmt.Sprintf("LOWER(reference_number) LIKE $%d", argIndex))
		args = append(args, searchPattern)
		argIndex++
	}

	// Date range filters
	if filters.PaymentFrom != nil {
		conditions = append(conditions, fmt.Sprintf("payment_date >= $%d", argIndex))
		args = append(args, *filters.PaymentFrom)
		argIndex++
	}

	if filters.PaymentTo != nil {
		conditions = append(conditions, fmt.Sprintf("payment_date <= $%d", argIndex))
		args = append(args, *filters.PaymentTo)
		argIndex++
	}

	// Amount range filters
	if filters.MinAmount != nil {
		conditions = append(conditions, fmt.Sprintf("amount >= $%d", argIndex))
		args = append(args, *filters.MinAmount)
		argIndex++
	}

	if filters.MaxAmount != nil {
		conditions = append(conditions, fmt.Sprintf("amount <= $%d", argIndex))
		args = append(args, *filters.MaxAmount)
		argIndex++
	}

	// Created by filter
	if filters.CreatedBy != nil {
		conditions = append(conditions, fmt.Sprintf("created_by = $%d", argIndex))
		args = append(args, *filters.CreatedBy)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	return whereClause, args
}

// BulkCreate creates multiple invoices in a single transaction
func (r *InvoiceRepository) BulkCreate(ctx context.Context, invoices []*domain.Invoice) error {
	if len(invoices) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := fmt.Sprintf(`
                INSERT INTO invoices (
                        organization_id, contact_id, invoice_number, type, status, currency,
                        exchange_rate, subtotal, tax_amount, discount_amount, total_amount,
                        paid_amount, balance_due, issue_date, due_date, payment_terms,
                        notes, terms_conditions, created_by, created_at, updated_at
                ) VALUES (
                        $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21
                ) RETURNING %s`, invoiceColumns)

	for _, invoice := range invoices {
		row := tx.QueryRowContext(ctx, query,
			invoice.OrganizationID, invoice.ContactID, invoice.InvoiceNumber,
			invoice.Type, invoice.Status, invoice.Currency, invoice.ExchangeRate,
			invoice.Subtotal, invoice.TaxAmount, invoice.DiscountAmount,
			invoice.TotalAmount, invoice.PaidAmount, invoice.BalanceDue,
			invoice.IssueDate, invoice.DueDate, invoice.PaymentTerms,
			invoice.Notes, invoice.TermsConditions, invoice.CreatedBy,
			invoice.CreatedAt, invoice.UpdatedAt,
		)
		err := scanInvoice(row, invoice)

		if err != nil {
			r.logger.Error("Failed to bulk create invoice", "error", err, "invoiceNumber", invoice.InvoiceNumber)
			return fmt.Errorf("failed to bulk create invoice %s: %w", invoice.InvoiceNumber, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit bulk create transaction: %w", err)
	}

	r.logger.Info("Bulk created invoices successfully", "count", len(invoices))
	return nil
}

// BulkUpdate updates multiple invoices in a single transaction
func (r *InvoiceRepository) BulkUpdate(ctx context.Context, invoices []*domain.Invoice) error {
	if len(invoices) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		UPDATE invoices SET 
			contact_id = $2, type = $3, status = $4, currency = $5,
			exchange_rate = $6, subtotal = $7, tax_amount = $8, discount_amount = $9,
			total_amount = $10, paid_amount = $11, balance_due = $12,
			issue_date = $13, due_date = $14, payment_terms = $15,
			notes = $16, terms_conditions = $17, updated_at = $18
		WHERE id = $1 AND organization_id = $19`

	for _, invoice := range invoices {
		result, err := tx.ExecContext(ctx, query,
			invoice.ID, invoice.ContactID, invoice.Type, invoice.Status,
			invoice.Currency, invoice.ExchangeRate, invoice.Subtotal,
			invoice.TaxAmount, invoice.DiscountAmount, invoice.TotalAmount,
			invoice.PaidAmount, invoice.BalanceDue, invoice.IssueDate,
			invoice.DueDate, invoice.PaymentTerms, invoice.Notes,
			invoice.TermsConditions, time.Now(), invoice.OrganizationID,
		)

		if err != nil {
			r.logger.Error("Failed to bulk update invoice", "error", err, "invoiceId", invoice.ID)
			return fmt.Errorf("failed to bulk update invoice %d: %w", invoice.ID, err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected for invoice %d: %w", invoice.ID, err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("invoice %d not found or not owned by organization", invoice.ID)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit bulk update transaction: %w", err)
	}

	r.logger.Info("Bulk updated invoices successfully", "count", len(invoices))
	return nil
}

// BulkDelete deletes multiple invoices in a single transaction
func (r *InvoiceRepository) BulkDelete(ctx context.Context, organizationID uint, invoiceIDs []uint) error {
	if len(invoiceIDs) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Build placeholders for IN clause
	placeholders := make([]string, len(invoiceIDs))
	args := make([]interface{}, len(invoiceIDs)+1)
	args[0] = organizationID

	for i, id := range invoiceIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = id
	}

	query := fmt.Sprintf("DELETE FROM invoices WHERE organization_id = $1 AND id IN (%s)", strings.Join(placeholders, ","))

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to bulk delete invoices", "error", err)
		return fmt.Errorf("failed to bulk delete invoices: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit bulk delete transaction: %w", err)
	}

	r.logger.Info("Bulk deleted invoices successfully", "count", rowsAffected, "requested", len(invoiceIDs))
	return nil
}

// BulkUpdateStatus updates the status of multiple invoices
func (r *InvoiceRepository) BulkUpdateStatus(ctx context.Context, organizationID uint, invoiceIDs []uint, status domain.InvoiceStatus) error {
	if len(invoiceIDs) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Build placeholders for IN clause
	placeholders := make([]string, len(invoiceIDs))
	args := make([]interface{}, len(invoiceIDs)+3)
	args[0] = status
	args[1] = time.Now()
	args[2] = organizationID

	for i, id := range invoiceIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+4)
		args[i+3] = id
	}

	query := fmt.Sprintf("UPDATE invoices SET status = $1, updated_at = $2 WHERE organization_id = $3 AND id IN (%s)", strings.Join(placeholders, ","))

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to bulk update invoice status", "error", err)
		return fmt.Errorf("failed to bulk update invoice status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit bulk status update transaction: %w", err)
	}

	r.logger.Info("Bulk updated invoice status successfully", "count", rowsAffected, "status", status)
	return nil
}

// GetInvoiceStats retrieves invoice statistics for an organization
func (r *InvoiceRepository) GetInvoiceStats(ctx context.Context, organizationID uint) (*repository.InvoiceStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_invoices,
			COUNT(CASE WHEN status = 'draft' THEN 1 END) as draft_invoices,
			COUNT(CASE WHEN status = 'sent' THEN 1 END) as sent_invoices,
			COUNT(CASE WHEN status = 'paid' THEN 1 END) as paid_invoices,
			COUNT(CASE WHEN status = 'canceled' THEN 1 END) as canceled_invoices,
			COUNT(CASE WHEN due_date < NOW() AND balance_due > 0 AND status NOT IN ('paid', 'canceled') THEN 1 END) as overdue_invoices,
			COALESCE(SUM(total_amount), 0) as total_revenue,
			COALESCE(SUM(paid_amount), 0) as paid_revenue,
			COALESCE(SUM(balance_due), 0) as outstanding_amount,
			COALESCE(SUM(CASE WHEN due_date < NOW() AND balance_due > 0 AND status NOT IN ('paid', 'canceled') THEN balance_due ELSE 0 END), 0) as overdue_amount,
			COALESCE(AVG(total_amount), 0) as average_invoice_value
		FROM invoices 
		WHERE organization_id = $1`

	stats := &repository.InvoiceStats{}
	err := r.db.QueryRowContext(ctx, query, organizationID).Scan(
		&stats.TotalInvoices, &stats.DraftInvoices, &stats.SentInvoices,
		&stats.PaidInvoices, &stats.CancelledInvoices, &stats.OverdueInvoices,
		&stats.TotalRevenue, &stats.PaidRevenue, &stats.OutstandingAmount,
		&stats.OverdueAmount, &stats.AverageInvoiceValue,
	)

	if err != nil {
		r.logger.Error("Failed to get invoice stats", "error", err, "organizationId", organizationID)
		return nil, fmt.Errorf("failed to get invoice stats: %w", err)
	}

	// Calculate average payment time
	paymentTimeQuery := `
		SELECT COALESCE(AVG(EXTRACT(DAY FROM p.payment_date - i.issue_date)), 0)
		FROM payments p
		JOIN invoices i ON p.invoice_id = i.id
		WHERE i.organization_id = $1 AND i.status = 'paid'`

	err = r.db.QueryRowContext(ctx, paymentTimeQuery, organizationID).Scan(&stats.AveragePaymentTime)
	if err != nil {
		r.logger.Error("Failed to get average payment time", "error", err, "organizationId", organizationID)
		// Don't fail the entire operation for this
		stats.AveragePaymentTime = 0
	}

	return stats, nil
}

// GetRevenueStats retrieves revenue statistics for a time period
func (r *InvoiceRepository) GetRevenueStats(ctx context.Context, organizationID uint, from, to time.Time) (*repository.RevenueStats, error) {
	query := `
		SELECT 
			COALESCE(SUM(total_amount), 0) as total_revenue,
			COALESCE(SUM(paid_amount), 0) as paid_revenue,
			COALESCE(SUM(balance_due), 0) as outstanding_amount,
			COUNT(*) as invoice_count,
			COALESCE(AVG(total_amount), 0) as average_invoice_value
		FROM invoices 
		WHERE organization_id = $1 AND issue_date >= $2 AND issue_date <= $3`

	stats := &repository.RevenueStats{
		Period:    fmt.Sprintf("%s to %s", from.Format("2006-01-02"), to.Format("2006-01-02")),
		StartDate: from,
		EndDate:   to,
		Currency:  "USD", // Default currency, could be made configurable
	}

	err := r.db.QueryRowContext(ctx, query, organizationID, from, to).Scan(
		&stats.TotalRevenue, &stats.PaidRevenue, &stats.OutstandingAmount,
		&stats.InvoiceCount, &stats.AverageInvoiceValue,
	)

	if err != nil {
		r.logger.Error("Failed to get revenue stats", "error", err, "organizationId", organizationID)
		return nil, fmt.Errorf("failed to get revenue stats: %w", err)
	}

	// Get payment count for the period
	paymentCountQuery := `
		SELECT COUNT(*)
		FROM payments p
		JOIN invoices i ON p.invoice_id = i.id
		WHERE i.organization_id = $1 AND p.payment_date >= $2 AND p.payment_date <= $3`

	err = r.db.QueryRowContext(ctx, paymentCountQuery, organizationID, from, to).Scan(&stats.PaymentCount)
	if err != nil {
		r.logger.Error("Failed to get payment count", "error", err, "organizationId", organizationID)
		// Don't fail the entire operation for this
		stats.PaymentCount = 0
	}

	return stats, nil
}

// GetOverdueInvoices retrieves all overdue invoices for an organization
func (r *InvoiceRepository) GetOverdueInvoices(ctx context.Context, organizationID uint) ([]*domain.Invoice, error) {
	query := fmt.Sprintf(`
                SELECT %s
                FROM invoices
                WHERE organization_id = $1
                  AND due_date < NOW()
                  AND balance_due > 0
                  AND status NOT IN ('paid', 'canceled')
                ORDER BY due_date ASC`, invoiceColumns)

	rows, err := r.db.QueryContext(ctx, query, organizationID)
	if err != nil {
		r.logger.Error("Failed to get overdue invoices", "error", err, "organizationId", organizationID)
		return nil, fmt.Errorf("failed to get overdue invoices: %w", err)
	}
	defer rows.Close()

	var invoices []*domain.Invoice
	for rows.Next() {
		invoice := &domain.Invoice{}
		err := scanInvoice(rows, invoice)
		if err != nil {
			r.logger.Error("Failed to scan overdue invoice", "error", err)
			return nil, fmt.Errorf("failed to scan overdue invoice: %w", err)
		}

		invoices = append(invoices, invoice)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate overdue invoices: %w", err)
	}

	return invoices, nil
}

// GetUpcomingDueInvoices retrieves invoices due within the specified number of days
func (r *InvoiceRepository) GetUpcomingDueInvoices(ctx context.Context, organizationID uint, days int) ([]*domain.Invoice, error) {
	query := fmt.Sprintf(`
                SELECT %s
                FROM invoices
                WHERE organization_id = $1
                  AND due_date BETWEEN NOW() AND NOW() + INTERVAL '%d days'
                  AND balance_due > 0
                  AND status NOT IN ('paid', 'canceled')
                ORDER BY due_date ASC`, invoiceColumns, days)

	rows, err := r.db.QueryContext(ctx, query, organizationID)
	if err != nil {
		r.logger.Error("Failed to get upcoming due invoices", "error", err, "organizationId", organizationID)
		return nil, fmt.Errorf("failed to get upcoming due invoices: %w", err)
	}
	defer rows.Close()

	var invoices []*domain.Invoice
	for rows.Next() {
		invoice := &domain.Invoice{}
		err := scanInvoice(rows, invoice)
		if err != nil {
			r.logger.Error("Failed to scan upcoming due invoice", "error", err)
			return nil, fmt.Errorf("failed to scan upcoming due invoice: %w", err)
		}

		invoices = append(invoices, invoice)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate upcoming due invoices: %w", err)
	}

	return invoices, nil
}

// GenerateInvoiceNumber generates a unique invoice number for the organization
func (r *InvoiceRepository) GenerateInvoiceNumber(ctx context.Context, organizationID uint, invoiceType domain.InvoiceType) (string, error) {
	// Get the current year and month
	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	// Get the next sequence number for this organization, type, and period
	query := `
		SELECT COALESCE(MAX(
			CAST(SUBSTRING(invoice_number FROM '[0-9]+$') AS INTEGER)
		), 0) + 1
		FROM invoices 
		WHERE organization_id = $1 
		  AND type = $2 
		  AND EXTRACT(YEAR FROM created_at) = $3 
		  AND EXTRACT(MONTH FROM created_at) = $4`

	var nextNumber int
	err := r.db.QueryRowContext(ctx, query, organizationID, invoiceType, year, month).Scan(&nextNumber)
	if err != nil {
		r.logger.Error("Failed to generate invoice number", "error", err, "organizationId", organizationID)
		return "", fmt.Errorf("failed to generate invoice number: %w", err)
	}

	// Format the invoice number based on type
	var prefix string
	switch invoiceType {
	case domain.InvoiceTypeInvoice:
		prefix = "INV"
	case domain.InvoiceTypeQuote:
		prefix = "QUO"
	case domain.InvoiceTypeCreditNote:
		prefix = "CN"
	case domain.InvoiceTypeProforma:
		prefix = "PRO"
	default:
		prefix = "INV"
	}

	// Format: PREFIX-YYYY-MM-NNNN (e.g., INV-2024-03-0001)
	invoiceNumber := fmt.Sprintf("%s-%04d-%02d-%04d", prefix, year, month, nextNumber)

	r.logger.Info("Generated invoice number", "invoiceNumber", invoiceNumber, "organizationId", organizationID)
	return invoiceNumber, nil
}

// ListPaginated returns paginated invoices for an organization
func (r *InvoiceRepository) ListPaginated(ctx context.Context, organizationID uint, params repository.PaginationParams) (repository.PaginationResult[*domain.Invoice], error) {
	baseQuery := `
		SELECT i.id, i.organization_id, i.contact_id, i.invoice_number, i.invoice_type,
			   i.status, i.currency, i.subtotal, i.tax_amount, i.total_amount,
			   i.issue_date, i.due_date, i.notes, i.created_by, i.created_at, i.updated_at
		FROM invoices i
		WHERE i.organization_id = $1`

	countQuery := `SELECT COUNT(*) FROM invoices WHERE organization_id = $1`

	// Get total count
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, organizationID).Scan(&total); err != nil {
		return repository.PaginationResult[*domain.Invoice]{}, fmt.Errorf("failed to get total count: %w", err)
	}

	// Build paginated query
	allowedSortFields := []string{"invoice_number", "issue_date", "due_date", "total_amount", "status", "created_at", "updated_at"}
	helper := NewPaginationHelper(r.db)
	paginatedQuery := helper.BuildPaginatedQuery(baseQuery, params, allowedSortFields)

	// Execute query
	rows, err := r.db.QueryContext(ctx, paginatedQuery, organizationID)
	if err != nil {
		return repository.PaginationResult[*domain.Invoice]{}, fmt.Errorf("failed to execute paginated query: %w", err)
	}
	defer rows.Close()

	// Scan results
	var invoices []*domain.Invoice
	for rows.Next() {
		invoice := &domain.Invoice{}
		err := rows.Scan(
			&invoice.ID, &invoice.OrganizationID, &invoice.ContactID,
			&invoice.InvoiceNumber, &invoice.Type, &invoice.Status,
			&invoice.Currency, &invoice.Subtotal, &invoice.TaxAmount, &invoice.TotalAmount,
			&invoice.IssueDate, &invoice.DueDate, &invoice.Notes, &invoice.CreatedBy,
			&invoice.CreatedAt, &invoice.UpdatedAt,
		)
		if err != nil {
			return repository.PaginationResult[*domain.Invoice]{}, fmt.Errorf("failed to scan invoice: %w", err)
		}
		invoices = append(invoices, invoice)
	}

	if err := rows.Err(); err != nil {
		return repository.PaginationResult[*domain.Invoice]{}, fmt.Errorf("row iteration error: %w", err)
	}

	return repository.NewPaginationResult(invoices, total, params), nil
}

// SearchPaginated returns paginated invoices matching search query
func (r *InvoiceRepository) SearchPaginated(ctx context.Context, organizationID uint, query string, params repository.PaginationParams) (repository.PaginationResult[*domain.Invoice], error) {
	baseQuery := `
		SELECT i.id, i.organization_id, i.contact_id, i.invoice_number, i.invoice_type,
			   i.status, i.currency, i.subtotal, i.tax_amount, i.total_amount,
			   i.issue_date, i.due_date, i.notes, i.created_by, i.created_at, i.updated_at
		FROM invoices i
		WHERE i.organization_id = $1`

	// Add search conditions
	helper := NewPaginationHelper(r.db)
	searchFields := []string{"i.invoice_number", "i.notes"}
	searchQuery, searchArgs := helper.BuildSearchQuery(baseQuery, searchFields, query)

	// Combine arguments
	args := helper.CombineArgs([]interface{}{organizationID}, searchArgs)

	// Build count query
	countQuery := helper.BuildCountQuery(searchQuery)

	// Get total count
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return repository.PaginationResult[*domain.Invoice]{}, fmt.Errorf("failed to get total count: %w", err)
	}

	// Build paginated query
	allowedSortFields := []string{"i.invoice_number", "i.issue_date", "i.due_date", "i.total_amount", "i.status", "i.created_at", "i.updated_at"}
	paginatedQuery := helper.BuildPaginatedQuery(searchQuery, params, allowedSortFields)

	// Execute query
	rows, err := r.db.QueryContext(ctx, paginatedQuery, args...)
	if err != nil {
		return repository.PaginationResult[*domain.Invoice]{}, fmt.Errorf("failed to execute paginated query: %w", err)
	}
	defer rows.Close()

	// Scan results
	var invoices []*domain.Invoice
	for rows.Next() {
		invoice := &domain.Invoice{}
		err := rows.Scan(
			&invoice.ID, &invoice.OrganizationID, &invoice.ContactID,
			&invoice.InvoiceNumber, &invoice.Type, &invoice.Status,
			&invoice.Currency, &invoice.Subtotal, &invoice.TaxAmount, &invoice.TotalAmount,
			&invoice.IssueDate, &invoice.DueDate, &invoice.Notes, &invoice.CreatedBy,
			&invoice.CreatedAt, &invoice.UpdatedAt,
		)
		if err != nil {
			return repository.PaginationResult[*domain.Invoice]{}, fmt.Errorf("failed to scan invoice: %w", err)
		}
		invoices = append(invoices, invoice)
	}

	if err := rows.Err(); err != nil {
		return repository.PaginationResult[*domain.Invoice]{}, fmt.Errorf("row iteration error: %w", err)
	}

	return repository.NewPaginationResult(invoices, total, params), nil
}
