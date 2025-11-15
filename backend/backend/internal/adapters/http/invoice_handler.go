// @kthulu:module:invoices
package adapterhttp

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/kthulu/kthulu-go/backend/internal/domain"
	"github.com/kthulu/kthulu-go/backend/internal/repository"
	"github.com/kthulu/kthulu-go/backend/internal/usecase"
)

// InvoiceHandler handles HTTP requests for invoice operations
type InvoiceHandler struct {
	invoiceUseCase *usecase.InvoiceUseCase
	validator      *validator.Validate
	logger         *zap.Logger
}

// NewInvoiceHandler creates a new invoice handler
func NewInvoiceHandler(invoiceUseCase *usecase.InvoiceUseCase, logger *zap.Logger) *InvoiceHandler {
	return &InvoiceHandler{
		invoiceUseCase: invoiceUseCase,
		validator:      validator.New(),
		logger:         logger,
	}
}

// RegisterRoutes registers invoice routes
func (h *InvoiceHandler) RegisterRoutes(r chi.Router) {
	r.Route("/invoices", func(r chi.Router) {
		r.Post("/", h.CreateInvoice)
		r.Get("/", h.ListInvoices)
		r.Get("/stats", h.GetInvoiceStats)
		r.Get("/overdue", h.GetOverdueInvoices)
		r.Get("/{invoiceId}", h.GetInvoice)
		r.Put("/{invoiceId}", h.UpdateInvoice)
		r.Delete("/{invoiceId}", h.DeleteInvoice)
		r.Patch("/{invoiceId}/status", h.SetInvoiceStatus)

		// Invoice item routes
		r.Post("/{invoiceId}/items", h.CreateInvoiceItem)
		r.Get("/{invoiceId}/items", h.GetInvoiceItems)
		r.Put("/{invoiceId}/items/{itemId}", h.UpdateInvoiceItem)
		r.Delete("/{invoiceId}/items/{itemId}", h.DeleteInvoiceItem)

		// Payment routes
		r.Post("/{invoiceId}/payments", h.CreatePayment)
		r.Get("/{invoiceId}/payments", h.GetInvoicePayments)
	})

	r.Route("/payments", func(r chi.Router) {
		r.Get("/", h.ListPayments)
		r.Get("/{paymentId}", h.GetPayment)
		r.Put("/{paymentId}", h.UpdatePayment)
		r.Delete("/{paymentId}", h.DeletePayment)
	})
}

// CreateInvoice creates a new invoice
// @Summary Create a new invoice
// @Description Create a new invoice in the organization
// @Tags invoices
// @Accept json
// @Produce json
// @Param organizationId header string true "Organization ID"
// @Param invoice body usecase.CreateInvoiceRequest true "Invoice data"
// @Success 201 {object} domain.Invoice
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /invoices [post]
func (h *InvoiceHandler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	organizationID := h.getOrganizationID(r)
	if organizationID == 0 {
		h.writeError(w, http.StatusBadRequest, "missing organization ID", nil)
		return
	}

	var req usecase.CreateInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Set organization ID from header
	req.OrganizationID = organizationID

	if err := h.validator.Struct(req); err != nil {
		h.writeError(w, http.StatusBadRequest, "validation failed", err)
		return
	}

	invoice, err := h.invoiceUseCase.CreateInvoice(r.Context(), req)
	if err != nil {
		switch err {
		case domain.ErrInvoiceAlreadyExists:
			h.writeError(w, http.StatusConflict, "invoice already exists", err)
		default:
			h.logger.Error("Failed to create invoice", zap.Error(err))
			h.writeError(w, http.StatusInternalServerError, "failed to create invoice", err)
		}
		return
	}

	h.writeJSON(w, http.StatusCreated, invoice)
}

// GetInvoice retrieves an invoice by ID
// @Summary Get an invoice by ID
// @Description Retrieve an invoice by its ID
// @Tags invoices
// @Produce json
// @Param organizationId header string true "Organization ID"
// @Param invoiceId path string true "Invoice ID"
// @Success 200 {object} domain.Invoice
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /invoices/{invoiceId} [get]
func (h *InvoiceHandler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	organizationID := h.getOrganizationID(r)
	if organizationID == 0 {
		h.writeError(w, http.StatusBadRequest, "missing organization ID", nil)
		return
	}

	invoiceID, err := h.getUintParam(r, "invoiceId")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid invoice ID", err)
		return
	}

	invoice, err := h.invoiceUseCase.GetInvoice(r.Context(), organizationID, invoiceID)
	if err != nil {
		switch err {
		case domain.ErrInvoiceNotFound:
			h.writeError(w, http.StatusNotFound, "invoice not found", err)
		default:
			h.logger.Error("Failed to get invoice", zap.Error(err))
			h.writeError(w, http.StatusInternalServerError, "failed to get invoice", err)
		}
		return
	}

	h.writeJSON(w, http.StatusOK, invoice)
}

// UpdateInvoice updates an existing invoice
// @Summary Update an invoice
// @Description Update an existing invoice's information
// @Tags invoices
// @Accept json
// @Produce json
// @Param organizationId header string true "Organization ID"
// @Param invoiceId path string true "Invoice ID"
// @Param invoice body usecase.UpdateInvoiceRequest true "Updated invoice data"
// @Success 200 {object} domain.Invoice
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /invoices/{invoiceId} [put]
func (h *InvoiceHandler) UpdateInvoice(w http.ResponseWriter, r *http.Request) {
	organizationID := h.getOrganizationID(r)
	if organizationID == 0 {
		h.writeError(w, http.StatusBadRequest, "missing organization ID", nil)
		return
	}

	invoiceID, err := h.getUintParam(r, "invoiceId")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid invoice ID", err)
		return
	}

	var req usecase.UpdateInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		h.writeError(w, http.StatusBadRequest, "validation failed", err)
		return
	}

	invoice, err := h.invoiceUseCase.UpdateInvoice(r.Context(), organizationID, invoiceID, req)
	if err != nil {
		switch err {
		case domain.ErrInvoiceNotFound:
			h.writeError(w, http.StatusNotFound, "invoice not found", err)
		case domain.ErrInvoiceNotEditable:
			h.writeError(w, http.StatusBadRequest, "invoice is not editable", err)
		default:
			h.logger.Error("Failed to update invoice", zap.Error(err))
			h.writeError(w, http.StatusInternalServerError, "failed to update invoice", err)
		}
		return
	}

	h.writeJSON(w, http.StatusOK, invoice)
}

// DeleteInvoice deletes an invoice
// @Summary Delete an invoice
// @Description Delete an invoice from the system
// @Tags invoices
// @Param organizationId header string true "Organization ID"
// @Param invoiceId path string true "Invoice ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /invoices/{invoiceId} [delete]
func (h *InvoiceHandler) DeleteInvoice(w http.ResponseWriter, r *http.Request) {
	organizationID := h.getOrganizationID(r)
	if organizationID == 0 {
		h.writeError(w, http.StatusBadRequest, "missing organization ID", nil)
		return
	}

	invoiceID, err := h.getUintParam(r, "invoiceId")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid invoice ID", err)
		return
	}

	err = h.invoiceUseCase.DeleteInvoice(r.Context(), organizationID, invoiceID)
	if err != nil {
		switch err {
		case domain.ErrInvoiceNotFound:
			h.writeError(w, http.StatusNotFound, "invoice not found", err)
		default:
			h.logger.Error("Failed to delete invoice", zap.Error(err))
			h.writeError(w, http.StatusInternalServerError, "failed to delete invoice", err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListInvoices retrieves invoices with filtering and pagination
// @Summary List invoices
// @Description Retrieve a paginated list of invoices with optional filtering
// @Tags invoices
// @Produce json
// @Param organizationId header string true "Organization ID"
// @Param page query int false "Page number (default: 1)"
// @Param pageSize query int false "Page size (default: 20, max: 100)"
// @Param contactId query int false "Filter by contact ID"
// @Param type query string false "Filter by invoice type"
// @Param status query string false "Filter by invoice status"
// @Param currency query string false "Filter by currency"
// @Param search query string false "Search in invoice number"
// @Param sortBy query string false "Sort by field"
// @Param sortOrder query string false "Sort order (asc, desc)"
// @Param includeItems query bool false "Include invoice items"
// @Param includePayments query bool false "Include invoice payments"
// @Success 200 {object} usecase.InvoiceListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /invoices [get]
func (h *InvoiceHandler) ListInvoices(w http.ResponseWriter, r *http.Request) {
	organizationID := h.getOrganizationID(r)
	if organizationID == 0 {
		h.writeError(w, http.StatusBadRequest, "missing organization ID", nil)
		return
	}

	filters := h.parseInvoiceFilters(r)

	response, err := h.invoiceUseCase.ListInvoices(r.Context(), organizationID, filters)
	if err != nil {
		h.logger.Error("Failed to list invoices", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "failed to list invoices", err)
		return
	}

	h.writeJSON(w, http.StatusOK, response)
}

// SetInvoiceStatus sets the status of an invoice
// @Summary Set invoice status
// @Description Set the status of an invoice
// @Tags invoices
// @Accept json
// @Param organizationId header string true "Organization ID"
// @Param invoiceId path string true "Invoice ID"
// @Param status body object{status:string} true "Status data"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /invoices/{invoiceId}/status [patch]
func (h *InvoiceHandler) SetInvoiceStatus(w http.ResponseWriter, r *http.Request) {
	organizationID := h.getOrganizationID(r)
	if organizationID == 0 {
		h.writeError(w, http.StatusBadRequest, "missing organization ID", nil)
		return
	}

	invoiceID, err := h.getUintParam(r, "invoiceId")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid invoice ID", err)
		return
	}

	var req struct {
		Status domain.InvoiceStatus `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	err = h.invoiceUseCase.SetInvoiceStatus(r.Context(), organizationID, invoiceID, req.Status)
	if err != nil {
		switch err {
		case domain.ErrInvoiceNotFound:
			h.writeError(w, http.StatusNotFound, "invoice not found", err)
		default:
			h.logger.Error("Failed to set invoice status", zap.Error(err))
			h.writeError(w, http.StatusInternalServerError, "failed to set invoice status", err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetInvoiceStats retrieves invoice statistics
// @Summary Get invoice statistics
// @Description Retrieve statistics for invoices in the organization
// @Tags invoices
// @Produce json
// @Param organizationId header string true "Organization ID"
// @Success 200 {object} repository.InvoiceStats
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /invoices/stats [get]
func (h *InvoiceHandler) GetInvoiceStats(w http.ResponseWriter, r *http.Request) {
	organizationID := h.getOrganizationID(r)
	if organizationID == 0 {
		h.writeError(w, http.StatusBadRequest, "missing organization ID", nil)
		return
	}

	stats, err := h.invoiceUseCase.GetInvoiceStats(r.Context(), organizationID)
	if err != nil {
		h.logger.Error("Failed to get invoice stats", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "failed to get invoice stats", err)
		return
	}

	h.writeJSON(w, http.StatusOK, stats)
}

// GetOverdueInvoices retrieves overdue invoices
// @Summary Get overdue invoices
// @Description Retrieve all overdue invoices for the organization
// @Tags invoices
// @Produce json
// @Param organizationId header string true "Organization ID"
// @Success 200 {array} domain.Invoice
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /invoices/overdue [get]
func (h *InvoiceHandler) GetOverdueInvoices(w http.ResponseWriter, r *http.Request) {
	organizationID := h.getOrganizationID(r)
	if organizationID == 0 {
		h.writeError(w, http.StatusBadRequest, "missing organization ID", nil)
		return
	}

	invoices, err := h.invoiceUseCase.GetOverdueInvoices(r.Context(), organizationID)
	if err != nil {
		h.logger.Error("Failed to get overdue invoices", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "failed to get overdue invoices", err)
		return
	}

	h.writeJSON(w, http.StatusOK, invoices)
}

// CreatePayment creates a new payment for an invoice
// @Summary Create a payment
// @Description Create a new payment for an invoice
// @Tags payments
// @Accept json
// @Produce json
// @Param organizationId header string true "Organization ID"
// @Param invoiceId path string true "Invoice ID"
// @Param payment body usecase.CreatePaymentRequest true "Payment data"
// @Success 201 {object} domain.Payment
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /invoices/{invoiceId}/payments [post]
func (h *InvoiceHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	organizationID := h.getOrganizationID(r)
	if organizationID == 0 {
		h.writeError(w, http.StatusBadRequest, "missing organization ID", nil)
		return
	}

	invoiceID, err := h.getUintParam(r, "invoiceId")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid invoice ID", err)
		return
	}

	var req usecase.CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Set organization ID and invoice ID from URL
	req.OrganizationID = organizationID
	req.InvoiceID = invoiceID

	if err := h.validator.Struct(req); err != nil {
		h.writeError(w, http.StatusBadRequest, "validation failed", err)
		return
	}

	payment, err := h.invoiceUseCase.CreatePayment(r.Context(), req)
	if err != nil {
		switch err {
		case domain.ErrInvoiceNotFound:
			h.writeError(w, http.StatusNotFound, "invoice not found", err)
		case domain.ErrInsufficientPayment:
			h.writeError(w, http.StatusBadRequest, "payment amount exceeds balance due", err)
		default:
			h.logger.Error("Failed to create payment", zap.Error(err))
			h.writeError(w, http.StatusInternalServerError, "failed to create payment", err)
		}
		return
	}

	h.writeJSON(w, http.StatusCreated, payment)
}

// Placeholder implementations for remaining handlers
func (h *InvoiceHandler) CreateInvoiceItem(w http.ResponseWriter, r *http.Request) {
	h.writeError(w, http.StatusNotImplemented, "not implemented", nil)
}

func (h *InvoiceHandler) GetInvoiceItems(w http.ResponseWriter, r *http.Request) {
	h.writeError(w, http.StatusNotImplemented, "not implemented", nil)
}

func (h *InvoiceHandler) UpdateInvoiceItem(w http.ResponseWriter, r *http.Request) {
	h.writeError(w, http.StatusNotImplemented, "not implemented", nil)
}

func (h *InvoiceHandler) DeleteInvoiceItem(w http.ResponseWriter, r *http.Request) {
	h.writeError(w, http.StatusNotImplemented, "not implemented", nil)
}

func (h *InvoiceHandler) GetInvoicePayments(w http.ResponseWriter, r *http.Request) {
	h.writeError(w, http.StatusNotImplemented, "not implemented", nil)
}

func (h *InvoiceHandler) ListPayments(w http.ResponseWriter, r *http.Request) {
	h.writeError(w, http.StatusNotImplemented, "not implemented", nil)
}

func (h *InvoiceHandler) GetPayment(w http.ResponseWriter, r *http.Request) {
	h.writeError(w, http.StatusNotImplemented, "not implemented", nil)
}

func (h *InvoiceHandler) UpdatePayment(w http.ResponseWriter, r *http.Request) {
	h.writeError(w, http.StatusNotImplemented, "not implemented", nil)
}

func (h *InvoiceHandler) DeletePayment(w http.ResponseWriter, r *http.Request) {
	h.writeError(w, http.StatusNotImplemented, "not implemented", nil)
}

// Helper methods

func (h *InvoiceHandler) getOrganizationID(r *http.Request) uint {
	// This should be extracted from JWT token or header
	// For now, we'll use a header value
	if orgIDStr := r.Header.Get("X-Organization-ID"); orgIDStr != "" {
		if orgID, err := strconv.ParseUint(orgIDStr, 10, 32); err == nil {
			return uint(orgID)
		}
	}
	return 0
}

func (h *InvoiceHandler) getUintParam(r *http.Request, param string) (uint, error) {
	paramStr := chi.URLParam(r, param)
	if paramStr == "" {
		return 0, nil
	}

	id, err := strconv.ParseUint(paramStr, 10, 32)
	if err != nil {
		return 0, err
	}

	return uint(id), nil
}

func (h *InvoiceHandler) parseInvoiceFilters(r *http.Request) repository.InvoiceFilters {
	filters := repository.DefaultInvoiceFilters()

	if contactIDStr := r.URL.Query().Get("contactId"); contactIDStr != "" {
		if contactID, err := strconv.ParseUint(contactIDStr, 10, 32); err == nil {
			id := uint(contactID)
			filters.ContactID = &id
		}
	}

	if typeStr := r.URL.Query().Get("type"); typeStr != "" {
		invoiceType := domain.InvoiceType(typeStr)
		filters.Type = &invoiceType
	}

	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		invoiceStatus := domain.InvoiceStatus(statusStr)
		filters.Status = &invoiceStatus
	}

	if currency := r.URL.Query().Get("currency"); currency != "" {
		filters.Currency = currency
	}

	if search := r.URL.Query().Get("search"); search != "" {
		filters.Search = search
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filters.Page = page
		}
	}

	if pageSizeStr := r.URL.Query().Get("pageSize"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 && pageSize <= 100 {
			filters.PageSize = pageSize
		}
	}

	if sortBy := r.URL.Query().Get("sortBy"); sortBy != "" {
		filters.SortBy = sortBy
	}

	if sortOrder := r.URL.Query().Get("sortOrder"); sortOrder != "" {
		filters.SortOrder = sortOrder
	}

	if includeItemsStr := r.URL.Query().Get("includeItems"); includeItemsStr != "" {
		if includeItems, err := strconv.ParseBool(includeItemsStr); err == nil {
			filters.IncludeItems = includeItems
		}
	}

	if includePaymentsStr := r.URL.Query().Get("includePayments"); includePaymentsStr != "" {
		if includePayments, err := strconv.ParseBool(includePaymentsStr); err == nil {
			filters.IncludePayments = includePayments
		}
	}

	return filters
}

func (h *InvoiceHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *InvoiceHandler) writeError(w http.ResponseWriter, status int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := map[string]interface{}{
		"error":  message,
		"status": status,
	}

	if err != nil {
		response["details"] = err.Error()
	}

	json.NewEncoder(w).Encode(response)
}
