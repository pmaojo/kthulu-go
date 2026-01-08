// @kthulu:module:user
package adapterhttp

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/pmaojo/kthulu-go/backend/core"
	"github.com/pmaojo/kthulu-go/backend/internal/adapters/http/middleware"
	"github.com/pmaojo/kthulu-go/backend/internal/domain"
	"github.com/pmaojo/kthulu-go/backend/internal/usecase"
)

// UserHandler exposes user profile endpoints.
type UserHandler struct {
	user         *usecase.UserUseCase
	tokenManager core.TokenManager
	log          *zap.SugaredLogger
}

// NewUserHandler constructs UserHandler with required dependencies.
func NewUserHandler(user *usecase.UserUseCase, tokenManager core.TokenManager, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		user:         user,
		tokenManager: tokenManager,
		log:          logger.Sugar(),
	}
}

// RegisterRoutes attaches user routes to the router.
func (h *UserHandler) RegisterRoutes(r chi.Router) {
	// Protected routes that require authentication
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(h.tokenManager))
		r.Get("/users/me", instrumentHandler("user.getProfile", h.getProfile))
		r.Patch("/users/me", instrumentHandler("user.updateProfile", h.updateProfile))
	})
}

type updateProfileRequest struct {
	Email           *string `json:"email,omitempty"`
	Password        *string `json:"password,omitempty"`
	CurrentPassword *string `json:"currentPassword,omitempty"`
}

// getProfile godoc
// @Summary Get current user profile
// @Description Returns the authenticated user's profile information
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.User "User profile retrieved successfully"
// @Failure 401 {object} map[string]string "Unauthorized - invalid or missing token"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/me [get]
func (h *UserHandler) getProfile(w http.ResponseWriter, r *http.Request) {
	// Use context-aware logger
	logger := middleware.GetSugaredLogger(r.Context())

	// Extract user ID from context (set by auth middleware)
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		logger.Errorw("Failed to get user ID from context", "error", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	logger.Infow("Get profile request", "userId", userID)

	// Get user profile
	user, err := h.user.GetProfile(r.Context(), userID)
	if err != nil {
		logger.Errorw("Failed to get user profile", "userId", userID, "error", err)
		if err == domain.ErrUserNotFound {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	logger.Infow("Profile retrieved successfully", "userId", userID)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)
}

// updateProfile godoc
// @Summary Update current user profile
// @Description Updates the authenticated user's profile information
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body updateProfileRequest true "Profile update details"
// @Success 200 {object} domain.User "Profile updated successfully"
// @Failure 400 {object} map[string]string "Invalid request or current password"
// @Failure 401 {object} map[string]string "Unauthorized - invalid or missing token"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 409 {object} map[string]string "Email already exists"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/me [patch]
func (h *UserHandler) updateProfile(w http.ResponseWriter, r *http.Request) {
	// Use context-aware logger
	logger := middleware.GetSugaredLogger(r.Context())

	// Extract user ID from context (set by auth middleware)
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		logger.Errorw("Failed to get user ID from context", "error", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Errorw("Failed to decode update profile request", "userId", userID, "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	logger.Infow("Update profile request", "userId", userID)

	// Convert to use case request
	updateReq := usecase.UpdateProfileRequest{
		Email:           req.Email,
		Password:        req.Password,
		CurrentPassword: req.CurrentPassword,
	}

	// Update user profile
	user, err := h.user.UpdateProfile(r.Context(), userID, updateReq)
	if err != nil {
		logger.Errorw("Failed to update user profile", "userId", userID, "error", err)
		if err == domain.ErrUserNotFound {
			w.WriteHeader(http.StatusNotFound)
		} else if err == domain.ErrUserAlreadyExists {
			w.WriteHeader(http.StatusConflict)
		} else if err.Error() == "invalid current password" {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	logger.Infow("Profile updated successfully", "userId", userID)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)
}
