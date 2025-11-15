// @kthulu:module:org
package adapterhttp

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/kthulu/kthulu-go/backend/core"
	"github.com/kthulu/kthulu-go/backend/internal/domain"
	"github.com/kthulu/kthulu-go/backend/internal/usecase"
)

// OrganizationHandler handles HTTP requests for organization management
type OrganizationHandler struct {
	organizationUC *usecase.OrganizationUseCase
	validator      *validator.Validate
	logger         core.Logger
}

// NewOrganizationHandler creates a new OrganizationHandler
func NewOrganizationHandler(
	organizationUC *usecase.OrganizationUseCase,
	logger core.Logger,
) *OrganizationHandler {
	return &OrganizationHandler{
		organizationUC: organizationUC,
		validator:      validator.New(),
		logger:         logger,
	}
}

// RegisterRoutes registers organization routes
func (h *OrganizationHandler) RegisterRoutes(r chi.Router) {
	r.Route("/organizations", func(r chi.Router) {
		r.Post("/", h.CreateOrganization)
		r.Get("/", h.ListUserOrganizations)

		r.Route("/{organizationId}", func(r chi.Router) {
			r.Get("/", h.GetOrganization)
			r.Patch("/", h.UpdateOrganization)
			r.Post("/invitations", h.InviteUser)
		})
	})

	r.Route("/invitations", func(r chi.Router) {
		r.Post("/{token}/accept", h.AcceptInvitation)
	})
}

// CreateOrganizationRequest represents the request to create an organization
type CreateOrganizationRequest struct {
	Name        string                  `json:"name" validate:"required,min=2,max=100"`
	Slug        string                  `json:"slug" validate:"required,min=2,max=50,alphanum"`
	Description string                  `json:"description,omitempty" validate:"max=500"`
	Type        domain.OrganizationType `json:"type" validate:"required,oneof=company nonprofit personal education"`
	Domain      string                  `json:"domain,omitempty" validate:"omitempty,fqdn"`
	Website     string                  `json:"website,omitempty" validate:"omitempty,url"`
	Phone       string                  `json:"phone,omitempty" validate:"omitempty,e164"`
	Address     string                  `json:"address,omitempty" validate:"max=200"`
	City        string                  `json:"city,omitempty" validate:"max=100"`
	State       string                  `json:"state,omitempty" validate:"max=100"`
	Country     string                  `json:"country,omitempty" validate:"max=100"`
	PostalCode  string                  `json:"postalCode,omitempty" validate:"max=20"`
}

// CreateOrganization godoc
// @Summary Create organization
// @Description Creates a new organization with the authenticated user as admin
// @Tags Organizations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateOrganizationRequest true "Organization details"
// @Success 201 {object} domain.Organization "Organization created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request or validation error"
// @Failure 401 {object} map[string]string "Unauthorized - invalid or missing token"
// @Failure 409 {object} map[string]string "Organization already exists"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /organizations [post]
func (h *OrganizationHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context (set by auth middleware)
	userID, ok := ctx.Value("userID").(uint)
	if !ok {
		h.logger.Error("User ID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req CreateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Invalid JSON in create organization request", "error", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		h.logger.Warn("Validation failed for create organization request", "error", err)
		h.writeValidationError(w, err)
		return
	}

	// Convert to use case request
	ucReq := usecase.CreateOrganizationRequest{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Type:        req.Type,
		Domain:      req.Domain,
		Website:     req.Website,
		Phone:       req.Phone,
		Address:     req.Address,
		City:        req.City,
		State:       req.State,
		Country:     req.Country,
		PostalCode:  req.PostalCode,
	}

	// Create organization
	org, err := h.organizationUC.CreateOrganization(ctx, userID, ucReq)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Return created organization
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(org)
}

// GetOrganization godoc
// @Summary Get organization
// @Description Returns details of a specific organization
// @Tags Organizations
// @Produce json
// @Security BearerAuth
// @Param organizationId path int true "Organization ID"
// @Success 200 {object} domain.Organization "Organization retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid organization ID"
// @Failure 401 {object} map[string]string "Unauthorized - invalid or missing token"
// @Failure 403 {object} map[string]string "User not in organization"
// @Failure 404 {object} map[string]string "Organization not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /organizations/{organizationId} [get]
func (h *OrganizationHandler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userID, ok := ctx.Value("userID").(uint)
	if !ok {
		h.logger.Error("User ID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get organization ID from URL
	organizationIDStr := chi.URLParam(r, "organizationId")
	organizationID, err := strconv.ParseUint(organizationIDStr, 10, 32)
	if err != nil {
		h.logger.Warn("Invalid organization ID in URL", "organizationId", organizationIDStr)
		http.Error(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	// Get organization
	org, err := h.organizationUC.GetOrganization(ctx, userID, uint(organizationID))
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Return organization
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(org)
}

// UpdateOrganizationRequest represents the request to update an organization
type UpdateOrganizationRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	Domain      *string `json:"domain,omitempty" validate:"omitempty,fqdn"`
	Website     *string `json:"website,omitempty" validate:"omitempty,url"`
	Phone       *string `json:"phone,omitempty" validate:"omitempty,e164"`
	Address     *string `json:"address,omitempty" validate:"omitempty,max=200"`
	City        *string `json:"city,omitempty" validate:"omitempty,max=100"`
	State       *string `json:"state,omitempty" validate:"omitempty,max=100"`
	Country     *string `json:"country,omitempty" validate:"omitempty,max=100"`
	PostalCode  *string `json:"postalCode,omitempty" validate:"omitempty,max=20"`
}

// UpdateOrganization handles PATCH /organizations/{organizationId}
func (h *OrganizationHandler) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userID, ok := ctx.Value("userID").(uint)
	if !ok {
		h.logger.Error("User ID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get organization ID from URL
	organizationIDStr := chi.URLParam(r, "organizationId")
	organizationID, err := strconv.ParseUint(organizationIDStr, 10, 32)
	if err != nil {
		h.logger.Warn("Invalid organization ID in URL", "organizationId", organizationIDStr)
		http.Error(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req UpdateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Invalid JSON in update organization request", "error", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		h.logger.Warn("Validation failed for update organization request", "error", err)
		h.writeValidationError(w, err)
		return
	}

	// Convert to use case request
	ucReq := usecase.UpdateOrganizationRequest{
		Name:        req.Name,
		Description: req.Description,
		Domain:      req.Domain,
		Website:     req.Website,
		Phone:       req.Phone,
		Address:     req.Address,
		City:        req.City,
		State:       req.State,
		Country:     req.Country,
		PostalCode:  req.PostalCode,
	}

	// Update organization
	org, err := h.organizationUC.UpdateOrganization(ctx, userID, uint(organizationID), ucReq)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Return updated organization
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(org)
}

// ListUserOrganizations godoc
// @Summary List user organizations
// @Description Returns all organizations the authenticated user belongs to
// @Tags Organizations
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Organizations retrieved successfully"
// @Failure 401 {object} map[string]string "Unauthorized - invalid or missing token"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /organizations [get]
func (h *OrganizationHandler) ListUserOrganizations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userID, ok := ctx.Value("userID").(uint)
	if !ok {
		h.logger.Error("User ID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// List user organizations
	organizations, err := h.organizationUC.ListUserOrganizations(ctx, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Return organizations
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"organizations": organizations,
		"count":         len(organizations),
	})
}

// InviteUserRequest represents the request to invite a user to an organization
type InviteUserRequest struct {
	Email   string                  `json:"email" validate:"required,email"`
	Role    domain.OrganizationRole `json:"role" validate:"required,oneof=admin member guest"`
	Message string                  `json:"message,omitempty" validate:"max=500"`
}

// InviteUser handles POST /organizations/{organizationId}/invitations
func (h *OrganizationHandler) InviteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userID, ok := ctx.Value("userID").(uint)
	if !ok {
		h.logger.Error("User ID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get organization ID from URL
	organizationIDStr := chi.URLParam(r, "organizationId")
	organizationID, err := strconv.ParseUint(organizationIDStr, 10, 32)
	if err != nil {
		h.logger.Warn("Invalid organization ID in URL", "organizationId", organizationIDStr)
		http.Error(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req InviteUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Invalid JSON in invite user request", "error", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		h.logger.Warn("Validation failed for invite user request", "error", err)
		h.writeValidationError(w, err)
		return
	}

	// Convert to use case request
	ucReq := usecase.InviteUserRequest{
		Email:   req.Email,
		Role:    req.Role,
		Message: req.Message,
	}

	// Invite user
	invitation, err := h.organizationUC.InviteUser(ctx, userID, uint(organizationID), ucReq)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Return created invitation
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(invitation)
}

// AcceptInvitation handles POST /invitations/{token}/accept
func (h *OrganizationHandler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userID, ok := ctx.Value("userID").(uint)
	if !ok {
		h.logger.Error("User ID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get token from URL
	token := chi.URLParam(r, "token")
	if token == "" {
		h.logger.Warn("Empty token in accept invitation request")
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}

	// Accept invitation
	org, err := h.organizationUC.AcceptInvitation(ctx, userID, token)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Return organization
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(org)
}

// handleError handles use case errors and converts them to appropriate HTTP responses
func (h *OrganizationHandler) handleError(w http.ResponseWriter, err error) {
	switch err {
	case domain.ErrOrganizationNotFound:
		http.Error(w, "Organization not found", http.StatusNotFound)
	case domain.ErrOrganizationAlreadyExists:
		http.Error(w, "Organization already exists", http.StatusConflict)
	case domain.ErrUserNotInOrganization:
		http.Error(w, "User not in organization", http.StatusForbidden)
	case domain.ErrInsufficientPermissions:
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
	case domain.ErrInvitationNotFound:
		http.Error(w, "Invitation not found", http.StatusNotFound)
	case domain.ErrInvitationExpired:
		http.Error(w, "Invitation expired", http.StatusGone)
	case domain.ErrInvitationAlreadyAccepted:
		http.Error(w, "Invitation already accepted", http.StatusConflict)
	default:
		h.logger.Error("Unhandled error in organization handler", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// writeValidationError writes validation errors as JSON response
func (h *OrganizationHandler) writeValidationError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	validationErrors := make(map[string]string)
	if validationErr, ok := err.(validator.ValidationErrors); ok {
		for _, fieldErr := range validationErr {
			validationErrors[fieldErr.Field()] = fieldErr.Tag()
		}
	}

	response := map[string]interface{}{
		"error":   "Validation failed",
		"details": validationErrors,
	}

	json.NewEncoder(w).Encode(response)
}
