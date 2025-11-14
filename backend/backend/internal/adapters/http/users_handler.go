package adapterhttp

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"backend/internal/modules/users"
)

// UsersHandler exposes registration and login endpoints.
type UsersHandler struct {
	svc *users.AuthService
	log *zap.SugaredLogger
}

// NewUsersHandler creates a new UsersHandler.
func NewUsersHandler(svc *users.AuthService, logger *zap.Logger) *UsersHandler {
	return &UsersHandler{svc: svc, log: logger.Sugar()}
}

// RegisterRoutes registers user authentication routes.
func (h *UsersHandler) RegisterRoutes(r chi.Router) {
	r.Post("/users/register", h.register)
	r.Post("/users/login", h.login)
	r.Get("/users/{id}", h.getByID)
}

type userRegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *UsersHandler) register(w http.ResponseWriter, r *http.Request) {
	var req userRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user, err := h.svc.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(user)
}

type userLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *UsersHandler) login(w http.ResponseWriter, r *http.Request) {
	var req userLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)
}

func (h *UsersHandler) getByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user, err := h.svc.GetByID(r.Context(), uint(id))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)
}
