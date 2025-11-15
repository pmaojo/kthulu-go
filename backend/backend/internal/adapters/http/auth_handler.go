package adapterhttp

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/kthulu/kthulu-go/backend/internal/adapters/http/middleware"
	"github.com/kthulu/kthulu-go/backend/internal/domain"
	"github.com/kthulu/kthulu-go/backend/internal/usecase"
)

// AuthHandler exposes authentication endpoints.
type AuthUseCase interface {
	Login(ctx context.Context, req usecase.LoginRequest) (*usecase.AuthResponse, error)
	Register(ctx context.Context, req usecase.RegisterRequest) (*usecase.AuthResponse, error)
	Refresh(ctx context.Context, req usecase.RefreshRequest) (*usecase.AuthResponse, error)
	Confirm(ctx context.Context, req usecase.ConfirmRequest) (*usecase.AuthResponse, error)
	Logout(ctx context.Context, req usecase.LogoutRequest) error
	ResendConfirmation(ctx context.Context, req usecase.ResendConfirmationRequest) error
}

// AuthHandler exposes authentication endpoints.
type AuthHandler struct {
	auth AuthUseCase
	log  *zap.SugaredLogger
}

// NewAuthHandler constructs AuthHandler with required dependencies.
func NewAuthHandler(auth AuthUseCase, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		auth: auth,
		log:  logger.Sugar(),
	}
}

// RegisterRoutes attaches auth routes to the router.
func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Post("/auth/login", instrumentHandler("auth.login", h.login))
	r.Post("/auth/register", instrumentHandler("auth.register", h.register))
	r.Post("/auth/confirm", instrumentHandler("auth.confirm", h.confirm))
	r.Post("/auth/refresh", instrumentHandler("auth.refresh", h.refresh))
	r.Post("/auth/logout", instrumentHandler("auth.logout", h.logout))
	r.Post("/auth/resend-confirmation", instrumentHandler("auth.resendConfirmation", h.resendConfirmation))
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	RoleID   uint   `json:"roleId,omitempty"`
}

type confirmRequest struct {
	Email            string `json:"email"`
	ConfirmationCode string `json:"confirmationCode"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type logoutRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type resendConfirmationRequest struct {
	Email string `json:"email"`
}

// login godoc
// @Summary Authenticate user
// @Description Authenticates a user and returns JWT tokens
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body loginRequest true "Login credentials"
// @Success 200 {object} usecase.AuthResponse "Authentication successful"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Invalid credentials or unconfirmed account"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/login [post]
func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
	// Use context-aware logger
	logger := middleware.GetSugaredLogger(r.Context())

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Errorw("Failed to decode login request", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	logger.Infow("Login attempt", "email", req.Email)

	// Convert to use case request
	loginReq := usecase.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	response, err := h.auth.Login(r.Context(), loginReq)
	if err != nil {
		logger.Errorw("Login failed", "email", req.Email, "error", err)
		if err == domain.ErrUserNotFound || err == domain.ErrUserNotConfirmed {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	logger.Infow("Login successful", "email", req.Email, "user_id", response.User.ID)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Errorw("Failed to encode login response", "error", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// register godoc
// @Summary Register a new user
// @Description Creates a new user account and sends a confirmation email
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body registerRequest true "Registration details"
// @Success 201 {object} usecase.AuthResponse "User registered successfully"
// @Failure 400 {object} map[string]string "Invalid request or user already exists"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/register [post]
func (h *AuthHandler) register(w http.ResponseWriter, r *http.Request) {
	// Use context-aware logger
	logger := middleware.GetSugaredLogger(r.Context())

	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Errorw("Failed to decode register request", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	logger.Infow("Registration attempt", "email", req.Email)

	// Convert to use case request
	registerReq := usecase.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
		RoleID:   req.RoleID,
	}

	response, err := h.auth.Register(r.Context(), registerReq)
	if err != nil {
		logger.Errorw("Registration failed", "email", req.Email, "error", err)
		if err == domain.ErrUserAlreadyExists || err == domain.ErrInvalidRole {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	logger.Infow("Registration successful", "email", req.Email, "user_id", response.User.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Errorw("Failed to encode register response", "error", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// refresh godoc
// @Summary Refresh access token
// @Description Generates a new access token using a valid refresh token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body refreshRequest true "Refresh token"
// @Success 200 {object} usecase.AuthResponse "Token refreshed successfully"
// @Failure 401 {object} map[string]string "Invalid or expired refresh token"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/refresh [post]
func (h *AuthHandler) refresh(w http.ResponseWriter, r *http.Request) {
	// Use context-aware logger
	logger := middleware.GetSugaredLogger(r.Context())

	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Errorw("Failed to decode refresh request", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	logger.Infow("Token refresh attempt")

	// Convert to use case request
	refreshReq := usecase.RefreshRequest{
		RefreshToken: req.RefreshToken,
	}

	response, err := h.auth.Refresh(r.Context(), refreshReq)
	if err != nil {
		logger.Errorw("Token refresh failed", "error", err)
		if err == domain.ErrTokenExpired || err == domain.ErrInvalidToken {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	logger.Infow("Token refresh successful")

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Errorw("Failed to encode refresh response", "error", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// confirm godoc
// @Summary Confirm user email
// @Description Confirms a user's email address using the confirmation code
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body confirmRequest true "Email confirmation details"
// @Success 200 {object} usecase.AuthResponse "Email confirmed successfully"
// @Failure 400 {object} map[string]string "Invalid confirmation code"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/confirm [post]
func (h *AuthHandler) confirm(w http.ResponseWriter, r *http.Request) {
	// Use context-aware logger
	logger := middleware.GetSugaredLogger(r.Context())

	var req confirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Errorw("Failed to decode confirm request", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	logger.Infow("Email confirmation attempt", "email", req.Email)

	// Convert to use case request
	confirmReq := usecase.ConfirmRequest{
		Email:            req.Email,
		ConfirmationCode: req.ConfirmationCode,
	}

	response, err := h.auth.Confirm(r.Context(), confirmReq)
	if err != nil {
		logger.Errorw("Email confirmation failed", "email", req.Email, "error", err)
		if err == domain.ErrUserNotFound {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
		return
	}

	logger.Infow("Email confirmation successful", "email", req.Email, "user_id", response.User.ID)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Errorw("Failed to encode confirm response", "error", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// logout godoc
// @Summary Logout user
// @Description Invalidates the user's refresh token
// @Tags Authentication
// @Accept json
// @Param request body logoutRequest true "Refresh token to invalidate"
// @Success 204 "Logout successful"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/logout [post]
func (h *AuthHandler) logout(w http.ResponseWriter, r *http.Request) {
	// Use context-aware logger
	logger := middleware.GetSugaredLogger(r.Context())

	var req logoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Errorw("Failed to decode logout request", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	logger.Infow("Logout attempt")

	// Convert to use case request
	logoutReq := usecase.LogoutRequest{
		RefreshToken: req.RefreshToken,
	}

	err := h.auth.Logout(r.Context(), logoutReq)
	if err != nil {
		logger.Errorw("Logout failed", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger.Infow("Logout successful")

	w.WriteHeader(http.StatusNoContent)
}

// resendConfirmation godoc
// @Summary Resend confirmation email
// @Description Resends the email confirmation code to the user
// @Tags Authentication
// @Accept json
// @Param request body resendConfirmationRequest true "Email address"
// @Success 204 "Confirmation email sent"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/resend-confirmation [post]
func (h *AuthHandler) resendConfirmation(w http.ResponseWriter, r *http.Request) {
	// Use context-aware logger
	logger := middleware.GetSugaredLogger(r.Context())

	var req resendConfirmationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Errorw("Failed to decode resend confirmation request", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	logger.Infow("Resend confirmation attempt", "email", req.Email)

	// Convert to use case request
	resendReq := usecase.ResendConfirmationRequest{
		Email: req.Email,
	}

	err := h.auth.ResendConfirmation(r.Context(), resendReq)
	if err != nil {
		logger.Errorw("Resend confirmation failed", "email", req.Email, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger.Infow("Resend confirmation successful", "email", req.Email)

	w.WriteHeader(http.StatusNoContent)
}
