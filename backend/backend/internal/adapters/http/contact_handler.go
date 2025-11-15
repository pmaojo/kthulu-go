// @kthulu:module:contacts
package adapterhttp

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/kthulu/kthulu-go/backend/core"
	"github.com/kthulu/kthulu-go/backend/internal/adapters/http/middleware"
	"github.com/kthulu/kthulu-go/backend/internal/domain"
	"github.com/kthulu/kthulu-go/backend/internal/repository"
	"github.com/kthulu/kthulu-go/backend/internal/usecase"
)

// ContactHandler handles HTTP requests for contact management
type ContactHandler struct {
	contactUC *usecase.ContactUseCase
	validator *validator.Validate
	logger    core.Logger
}

// NewContactHandler creates a new ContactHandler
func NewContactHandler(
	contactUC *usecase.ContactUseCase,
	logger core.Logger,
) *ContactHandler {
	return &ContactHandler{
		contactUC: contactUC,
		validator: validator.New(),
		logger:    logger,
	}
}

// RegisterRoutes registers contact routes
func (h *ContactHandler) RegisterRoutes(r chi.Router) {
	r.Route("/contacts", func(r chi.Router) {
		r.Post("/", h.CreateContact)
		r.Get("/", h.ListContacts)
		r.Get("/stats", h.GetContactStats)

		r.Route("/{contactId}", func(r chi.Router) {
			r.Get("/", h.GetContact)
			r.Patch("/", h.UpdateContact)
			r.Delete("/", h.DeleteContact)
			r.Patch("/status", h.SetContactStatus)
			r.Post("/convert-to-customer", h.ConvertLeadToCustomer)

			// Address management
			r.Post("/addresses", h.AddContactAddress)
			r.Route("/addresses/{addressId}", func(r chi.Router) {
				r.Patch("/", h.UpdateContactAddress)
				r.Delete("/", h.DeleteContactAddress)
				r.Post("/set-primary", h.SetPrimaryAddress)
			})

			// Phone management
			r.Post("/phones", h.AddContactPhone)
			r.Route("/phones/{phoneId}", func(r chi.Router) {
				r.Patch("/", h.UpdateContactPhone)
				r.Delete("/", h.DeleteContactPhone)
				r.Post("/set-primary", h.SetPrimaryPhone)
			})
		})
	})
}

// CreateContact creates a new contact
// @Summary Create a new contact
// @Description Create a new contact for the organization
// @Tags @kthulu:module:contacts
// @Accept json
// @Produce json
// @Param X-Organization-ID header string true "Organization ID"
// @Param contact body usecase.CreateContactRequest true "Contact data"
// @Success 201 {object} domain.Contact
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /contacts [post]
func (h *ContactHandler) CreateContact(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get organization ID from context (set by middleware)
	organizationID, err := getOrganizationIDFromContext(ctx)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid organization ID", err)
		return
	}

	var req usecase.CreateContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Validation failed", err)
		return
	}

	contact, err := h.contactUC.CreateContact(ctx, organizationID, req)
	if err != nil {
		if err == domain.ErrContactAlreadyExists {
			h.writeErrorResponse(w, http.StatusConflict, "Contact already exists", err)
			return
		}
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create contact", err)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, contact)
}

// GetContact retrieves a contact by ID
// @Summary Get a contact
// @Description Get a contact by ID
// @Tags @kthulu:module:contacts
// @Produce json
// @Param X-Organization-ID header string true "Organization ID"
// @Param contactId path int true "Contact ID"
// @Success 200 {object} domain.Contact
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /contacts/{contactId} [get]
func (h *ContactHandler) GetContact(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	organizationID, err := getOrganizationIDFromContext(ctx)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid organization ID", err)
		return
	}

	contactID, err := strconv.ParseUint(chi.URLParam(r, "contactId"), 10, 32)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid contact ID", err)
		return
	}

	contact, err := h.contactUC.GetContact(ctx, organizationID, uint(contactID))
	if err != nil {
		if err == domain.ErrContactNotFound {
			h.writeErrorResponse(w, http.StatusNotFound, "Contact not found", err)
			return
		}
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get contact", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, contact)
}

// UpdateContact updates an existing contact
// @Summary Update a contact
// @Description Update an existing contact
// @Tags @kthulu:module:contacts
// @Accept json
// @Produce json
// @Param X-Organization-ID header string true "Organization ID"
// @Param contactId path int true "Contact ID"
// @Param contact body usecase.UpdateContactRequest true "Contact data"
// @Success 200 {object} domain.Contact
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /contacts/{contactId} [patch]
func (h *ContactHandler) UpdateContact(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	organizationID, err := getOrganizationIDFromContext(ctx)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid organization ID", err)
		return
	}

	contactID, err := strconv.ParseUint(chi.URLParam(r, "contactId"), 10, 32)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid contact ID", err)
		return
	}

	var req usecase.UpdateContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Validation failed", err)
		return
	}

	contact, err := h.contactUC.UpdateContact(ctx, organizationID, uint(contactID), req)
	if err != nil {
		if err == domain.ErrContactNotFound {
			h.writeErrorResponse(w, http.StatusNotFound, "Contact not found", err)
			return
		}
		if err == domain.ErrContactAlreadyExists {
			h.writeErrorResponse(w, http.StatusConflict, "Contact with email already exists", err)
			return
		}
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to update contact", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, contact)
}

// DeleteContact deletes a contact
// @Summary Delete a contact
// @Description Delete a contact by ID
// @Tags @kthulu:module:contacts
// @Param X-Organization-ID header string true "Organization ID"
// @Param contactId path int true "Contact ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /contacts/{contactId} [delete]
func (h *ContactHandler) DeleteContact(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	organizationID, err := getOrganizationIDFromContext(ctx)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid organization ID", err)
		return
	}

	contactID, err := strconv.ParseUint(chi.URLParam(r, "contactId"), 10, 32)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid contact ID", err)
		return
	}

	if err := h.contactUC.DeleteContact(ctx, organizationID, uint(contactID)); err != nil {
		if err == domain.ErrContactNotFound {
			h.writeErrorResponse(w, http.StatusNotFound, "Contact not found", err)
			return
		}
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete contact", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListContacts lists contacts with filtering and pagination
// @Summary List contacts
// @Description List contacts with filtering and pagination
// @Tags @kthulu:module:contacts
// @Produce json
// @Param X-Organization-ID header string true "Organization ID"
// @Param type query string false "Contact type" Enums(customer,supplier,lead,partner)
// @Param isActive query boolean false "Filter by active status"
// @Param search query string false "Search in name, email, company"
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Param sortBy query string false "Sort field" default(created_at)
// @Param sortOrder query string false "Sort order" Enums(asc,desc) default(desc)
// @Success 200 {object} usecase.ContactListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /contacts [get]
func (h *ContactHandler) ListContacts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	organizationID, err := getOrganizationIDFromContext(ctx)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid organization ID", err)
		return
	}

	filters := h.parseContactFilters(r)

	response, err := h.contactUC.ListContacts(ctx, organizationID, filters)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list contacts", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// SetContactStatus sets the active status of a contact
// @Summary Set contact status
// @Description Set the active status of a contact
// @Tags @kthulu:module:contacts
// @Accept json
// @Param X-Organization-ID header string true "Organization ID"
// @Param contactId path int true "Contact ID"
// @Param status body object{active:boolean} true "Status data"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /contacts/{contactId}/status [patch]
func (h *ContactHandler) SetContactStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	organizationID, err := getOrganizationIDFromContext(ctx)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid organization ID", err)
		return
	}

	contactID, err := strconv.ParseUint(chi.URLParam(r, "contactId"), 10, 32)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid contact ID", err)
		return
	}

	var req struct {
		Active bool `json:"active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.contactUC.SetContactActive(ctx, organizationID, uint(contactID), req.Active); err != nil {
		if err == domain.ErrContactNotFound {
			h.writeErrorResponse(w, http.StatusNotFound, "Contact not found", err)
			return
		}
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to update contact status", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ConvertLeadToCustomer converts a lead to a customer
// @Summary Convert lead to customer
// @Description Convert a lead contact to a customer
// @Tags @kthulu:module:contacts
// @Produce json
// @Param X-Organization-ID header string true "Organization ID"
// @Param contactId path int true "Contact ID"
// @Success 200 {object} domain.Contact
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /contacts/{contactId}/convert-to-customer [post]
func (h *ContactHandler) ConvertLeadToCustomer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	organizationID, err := getOrganizationIDFromContext(ctx)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid organization ID", err)
		return
	}

	contactID, err := strconv.ParseUint(chi.URLParam(r, "contactId"), 10, 32)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid contact ID", err)
		return
	}

	contact, err := h.contactUC.ConvertLeadToCustomer(ctx, organizationID, uint(contactID))
	if err != nil {
		if err == domain.ErrContactNotFound {
			h.writeErrorResponse(w, http.StatusNotFound, "Contact not found", err)
			return
		}
		h.writeErrorResponse(w, http.StatusBadRequest, "Failed to convert lead", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, contact)
}

// GetContactStats retrieves contact statistics
// @Summary Get contact statistics
// @Description Get contact statistics for the organization
// @Tags @kthulu:module:contacts
// @Produce json
// @Param X-Organization-ID header string true "Organization ID"
// @Success 200 {object} repository.ContactStats
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /contacts/stats [get]
func (h *ContactHandler) GetContactStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	organizationID, err := getOrganizationIDFromContext(ctx)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid organization ID", err)
		return
	}

	stats, err := h.contactUC.GetContactStats(ctx, organizationID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get contact stats", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, stats)
}

// AddContactAddress adds an address to a contact
// @Summary Add contact address
// @Description Add an address to a contact
// @Tags @kthulu:module:contacts
// @Accept json
// @Produce json
// @Param X-Organization-ID header string true "Organization ID"
// @Param contactId path int true "Contact ID"
// @Param address body usecase.CreateAddressRequest true "Address data"
// @Success 201 {object} domain.ContactAddress
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /contacts/{contactId}/addresses [post]
func (h *ContactHandler) AddContactAddress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	organizationID, err := getOrganizationIDFromContext(ctx)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid organization ID", err)
		return
	}

	contactID, err := strconv.ParseUint(chi.URLParam(r, "contactId"), 10, 32)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid contact ID", err)
		return
	}

	var req usecase.CreateAddressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Validation failed", err)
		return
	}

	address, err := h.contactUC.AddContactAddress(ctx, organizationID, uint(contactID), req)
	if err != nil {
		if err == domain.ErrContactNotFound {
			h.writeErrorResponse(w, http.StatusNotFound, "Contact not found", err)
			return
		}
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to add address", err)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, address)
}

// AddContactPhone adds a phone number to a contact
// @Summary Add contact phone
// @Description Add a phone number to a contact
// @Tags @kthulu:module:contacts
// @Accept json
// @Produce json
// @Param X-Organization-ID header string true "Organization ID"
// @Param contactId path int true "Contact ID"
// @Param phone body usecase.CreatePhoneRequest true "Phone data"
// @Success 201 {object} domain.ContactPhone
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /contacts/{contactId}/phones [post]
func (h *ContactHandler) AddContactPhone(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	organizationID, err := getOrganizationIDFromContext(ctx)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid organization ID", err)
		return
	}

	contactID, err := strconv.ParseUint(chi.URLParam(r, "contactId"), 10, 32)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid contact ID", err)
		return
	}

	var req usecase.CreatePhoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Validation failed", err)
		return
	}

	phone, err := h.contactUC.AddContactPhone(ctx, organizationID, uint(contactID), req)
	if err != nil {
		if err == domain.ErrContactNotFound {
			h.writeErrorResponse(w, http.StatusNotFound, "Contact not found", err)
			return
		}
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to add phone", err)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, phone)
}

func (h *ContactHandler) UpdateContactAddress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	organizationID, err := getOrganizationIDFromContext(ctx)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid organization ID", err)
		return
	}

	contactID, err := strconv.ParseUint(chi.URLParam(r, "contactId"), 10, 32)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid contact ID", err)
		return
	}

	addressID, err := strconv.ParseUint(chi.URLParam(r, "addressId"), 10, 32)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid address ID", err)
		return
	}

	var req usecase.UpdateAddressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Validation failed", err)
		return
	}

	address, err := h.contactUC.UpdateContactAddress(ctx, organizationID, uint(contactID), uint(addressID), req)
	if err != nil {
		if err == domain.ErrContactNotFound || err == domain.ErrAddressNotFound {
			h.writeErrorResponse(w, http.StatusNotFound, "Address not found", err)
			return
		}
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to update address", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, address)
}

func (h *ContactHandler) DeleteContactAddress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	organizationID, err := getOrganizationIDFromContext(ctx)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid organization ID", err)
		return
	}

	contactID, err := strconv.ParseUint(chi.URLParam(r, "contactId"), 10, 32)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid contact ID", err)
		return
	}

	addressID, err := strconv.ParseUint(chi.URLParam(r, "addressId"), 10, 32)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid address ID", err)
		return
	}

	if err := h.contactUC.DeleteContactAddress(ctx, organizationID, uint(contactID), uint(addressID)); err != nil {
		if err == domain.ErrContactNotFound || err == domain.ErrAddressNotFound {
			h.writeErrorResponse(w, http.StatusNotFound, "Address not found", err)
			return
		}
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete address", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ContactHandler) SetPrimaryAddress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	organizationID, err := getOrganizationIDFromContext(ctx)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid organization ID", err)
		return
	}

	contactID, err := strconv.ParseUint(chi.URLParam(r, "contactId"), 10, 32)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid contact ID", err)
		return
	}

	addressID, err := strconv.ParseUint(chi.URLParam(r, "addressId"), 10, 32)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid address ID", err)
		return
	}

	if err := h.contactUC.SetPrimaryAddress(ctx, organizationID, uint(contactID), uint(addressID)); err != nil {
		if err == domain.ErrContactNotFound || err == domain.ErrAddressNotFound {
			h.writeErrorResponse(w, http.StatusNotFound, "Address not found", err)
			return
		}
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to set primary address", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ContactHandler) UpdateContactPhone(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	organizationID, err := getOrganizationIDFromContext(ctx)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid organization ID", err)
		return
	}

	contactID, err := strconv.ParseUint(chi.URLParam(r, "contactId"), 10, 32)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid contact ID", err)
		return
	}

	phoneID, err := strconv.ParseUint(chi.URLParam(r, "phoneId"), 10, 32)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid phone ID", err)
		return
	}

	var req usecase.UpdatePhoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Validation failed", err)
		return
	}

	phone, err := h.contactUC.UpdateContactPhone(ctx, organizationID, uint(contactID), uint(phoneID), req)
	if err != nil {
		if err == domain.ErrContactNotFound || err == domain.ErrPhoneNotFound {
			h.writeErrorResponse(w, http.StatusNotFound, "Phone not found", err)
			return
		}
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to update phone", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, phone)
}

func (h *ContactHandler) DeleteContactPhone(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	organizationID, err := getOrganizationIDFromContext(ctx)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid organization ID", err)
		return
	}

	contactID, err := strconv.ParseUint(chi.URLParam(r, "contactId"), 10, 32)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid contact ID", err)
		return
	}

	phoneID, err := strconv.ParseUint(chi.URLParam(r, "phoneId"), 10, 32)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid phone ID", err)
		return
	}

	if err := h.contactUC.DeleteContactPhone(ctx, organizationID, uint(contactID), uint(phoneID)); err != nil {
		if err == domain.ErrContactNotFound || err == domain.ErrPhoneNotFound {
			h.writeErrorResponse(w, http.StatusNotFound, "Phone not found", err)
			return
		}
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete phone", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ContactHandler) SetPrimaryPhone(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	organizationID, err := getOrganizationIDFromContext(ctx)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid organization ID", err)
		return
	}

	contactID, err := strconv.ParseUint(chi.URLParam(r, "contactId"), 10, 32)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid contact ID", err)
		return
	}

	phoneID, err := strconv.ParseUint(chi.URLParam(r, "phoneId"), 10, 32)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid phone ID", err)
		return
	}

	if err := h.contactUC.SetPrimaryPhone(ctx, organizationID, uint(contactID), uint(phoneID)); err != nil {
		if err == domain.ErrContactNotFound || err == domain.ErrPhoneNotFound {
			h.writeErrorResponse(w, http.StatusNotFound, "Phone not found", err)
			return
		}
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to set primary phone", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper methods

func (h *ContactHandler) parseContactFilters(r *http.Request) repository.ContactFilters {
	filters := repository.DefaultContactFilters()

	if contactType := r.URL.Query().Get("type"); contactType != "" {
		filters.Type = domain.ContactType(contactType)
	}

	if isActiveStr := r.URL.Query().Get("isActive"); isActiveStr != "" {
		if isActive := isActiveStr == "true"; isActiveStr == "true" || isActiveStr == "false" {
			filters.IsActive = &isActive
		}
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

	if sortOrder := r.URL.Query().Get("sortOrder"); sortOrder == "asc" || sortOrder == "desc" {
		filters.SortOrder = sortOrder
	}

	filters.Validate()
	return filters
}

func (h *ContactHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *ContactHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	h.logger.Error("HTTP error", map[string]interface{}{
		"status_code": statusCode,
		"message":     message,
		"error":       err,
	})

	response := ErrorResponse{
		Error:   message,
		Code:    statusCode,
		Details: nil,
	}

	if err != nil {
		response.Details = err.Error()
	}

	h.writeJSONResponse(w, statusCode, response)
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string      `json:"error"`
	Code    int         `json:"code"`
	Details interface{} `json:"details,omitempty"`
}

// Helper function to get organization ID from context
// This should be set by middleware
func getOrganizationIDFromContext(ctx context.Context) (uint, error) {
	if orgID, ok := ctx.Value(middleware.OrganizationIDKey).(uint); ok && orgID != 0 {
		return orgID, nil
	}
	return 0, domain.ErrOrganizationNotFound
}
