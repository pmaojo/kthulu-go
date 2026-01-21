// @kthulu:handler:mail
package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	
	"github.com/gorilla/mux"
	"feature-tour/internal/modules/mail/core"
	
)

type MailHandler struct {
	service core.MailService
}

func NewMailHandler(service core.MailService) *MailHandler {
	return &MailHandler{service: service}
}

// RegisterRoutes registers all routes for mail
func (h *MailHandler) RegisterRoutes(router *mux.Router) {
	sub := router.PathPrefix("/mail").Subrouter()
	

	sub.HandleFunc("/", h.List).Methods("GET")
	sub.HandleFunc("/", h.Create).Methods("POST")
	sub.HandleFunc("/{id}", h.GetByID).Methods("GET")
	sub.HandleFunc("/{id}", h.Update).Methods("PUT")
	sub.HandleFunc("/{id}", h.Delete).Methods("DELETE")
}

// Create handles the creation of a new mail
// @Summary      Create a new mail
// @Description  Creates a new mail with the provided data
// @Tags         Mails
// @Accept       json
// @Produce      json
// @Param        input body core.Mail true "Mail object"
// @Success      200  {object}  core.Mail
// @Failure      400  {object}  map[string]string "Invalid input"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Router       /mail [post]
func (h *MailHandler) Create(w http.ResponseWriter, r *http.Request) {
	var entity core.Mail
	if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	if err := h.service.CreateMail(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

// GetByID retrieves a mail by its ID
// @Summary      Get mail
// @Description  Get a mail by its ID
// @Tags         Mails
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Mail ID"
	// @Success      200  {object}  core.Mail
// @Failure      400  {object}  map[string]string "Invalid ID"
// @Failure      404  {object}  map[string]string "Mail not found"
// @Router       /mail/{id} [get]
func (h *MailHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	entity, err := h.service.GetMailByID(uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

// Update handles the update of a mail
func (h *MailHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	var entity core.Mail
	if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	entity.ID = uint(id)
	if err := h.service.UpdateMail(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

// Delete handles the deletion of a mail
func (h *MailHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	if err := h.service.DeleteMail(uint(id)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// List retrieves all mails
func (h *MailHandler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	filter := core.SearchFilter{
		Query: query,
		Limit: limit,
		Offset: offset,
	}

	entities, err := h.service.ListMails(filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entities)
}
