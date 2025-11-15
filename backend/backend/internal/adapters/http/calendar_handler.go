// @kthulu:module:calendar
package adapterhttp

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/pmaojo/kthulu-go/backend/internal/domain"
	"github.com/pmaojo/kthulu-go/backend/internal/usecase"
)

// ensure domain types are recognized by documentation generators
var _ = domain.Calendar{}

// CalendarHandler handles calendar-related HTTP requests
type CalendarHandler struct {
	calendarUC *usecase.CalendarUseCase
	logger     *zap.Logger
}

// NewCalendarHandler creates a new calendar handler
func NewCalendarHandler(calendarUC *usecase.CalendarUseCase, logger *zap.Logger) *CalendarHandler {
	return &CalendarHandler{
		calendarUC: calendarUC,
		logger:     logger,
	}
}

// RegisterRoutes registers calendar routes
func (h *CalendarHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/calendar", func(r chi.Router) {
		// Calendar routes
		r.Route("/calendars", func(r chi.Router) {
			r.Post("/", h.CreateCalendar)
			r.Get("/", h.ListCalendars)
			r.Get("/{id}", h.GetCalendar)
		})

		// Event routes
		r.Route("/events", func(r chi.Router) {
			r.Post("/", h.CreateEvent)
			r.Get("/{id}", h.GetEvent)
			r.Get("/", h.ListEvents)
		})

		// Appointment booking routes
		r.Route("/appointments", func(r chi.Router) {
			r.Get("/available-slots", h.GetAvailableSlots)
			r.Post("/book", h.BookSlot)
		})

		// Utility routes
		r.Get("/business-day/{date}", h.IsBusinessDay)
	})

	h.logger.Info("Calendar routes registered")
}

// CreateCalendar godoc
// @Summary Create a new calendar
// @Description Creates a new calendar for a user
// @Tags Calendar
// @Accept json
// @Produce json
// @Param request body usecase.CreateCalendarRequest true "Calendar details"
// @Success 201 {object} domain.Calendar "Calendar created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/calendar/calendars [post]
func (h *CalendarHandler) CreateCalendar(w http.ResponseWriter, r *http.Request) {
	var req usecase.CreateCalendarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode create calendar request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	calendar, err := h.calendarUC.CreateCalendar(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create calendar", zap.Error(err))
		http.Error(w, "Failed to create calendar", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(calendar)
}

// GetCalendar godoc
// @Summary Get calendar by ID
// @Description Retrieves a calendar by its ID
// @Tags Calendar
// @Produce json
// @Param id path int true "Calendar ID"
// @Success 200 {object} domain.Calendar "Calendar details"
// @Failure 404 {object} map[string]string "Calendar not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/calendar/calendars/{id} [get]
func (h *CalendarHandler) GetCalendar(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid calendar ID", http.StatusBadRequest)
		return
	}

	calendar, err := h.calendarUC.GetCalendar(r.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to get calendar", zap.Uint64("id", id), zap.Error(err))
		http.Error(w, "Calendar not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(calendar)
}

// ListCalendars godoc
// @Summary List calendars
// @Description Retrieves a paginated list of calendars for a user
// @Tags Calendar
// @Produce json
// @Param ownerId query int true "Owner user ID"
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(10)
// @Success 200 {object} usecase.PaginatedResponse[domain.Calendar] "List of calendars"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/calendar/calendars [get]
func (h *CalendarHandler) ListCalendars(w http.ResponseWriter, r *http.Request) {
	ownerIDStr := r.URL.Query().Get("ownerId")
	if ownerIDStr == "" {
		http.Error(w, "ownerId parameter is required", http.StatusBadRequest)
		return
	}

	ownerID, err := strconv.ParseUint(ownerIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid ownerId", http.StatusBadRequest)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	req := usecase.ListCalendarsRequest{
		OwnerID:  uint(ownerID),
		Page:     page,
		PageSize: pageSize,
	}

	response, err := h.calendarUC.ListCalendars(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to list calendars", zap.Error(err))
		http.Error(w, "Failed to list calendars", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateEvent godoc
// @Summary Create a new event
// @Description Creates a new event in a calendar
// @Tags Calendar
// @Accept json
// @Produce json
// @Param request body usecase.CreateEventRequest true "Event details"
// @Success 201 {object} domain.Event "Event created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/calendar/events [post]
func (h *CalendarHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var req usecase.CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode create event request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	event, err := h.calendarUC.CreateEvent(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create event", zap.Error(err))
		http.Error(w, "Failed to create event", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(event)
}

// GetEvent godoc
// @Summary Get event by ID
// @Description Retrieves an event by its ID
// @Tags Calendar
// @Produce json
// @Param id path int true "Event ID"
// @Success 200 {object} domain.Event "Event details"
// @Failure 404 {object} map[string]string "Event not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/calendar/events/{id} [get]
func (h *CalendarHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	event, err := h.calendarUC.GetEvent(r.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to get event", zap.Uint64("id", id), zap.Error(err))
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(event)
}

// ListEvents godoc
// @Summary List events
// @Description Retrieves a paginated list of events for a calendar
// @Tags Calendar
// @Produce json
// @Param calendarId query int true "Calendar ID"
// @Param startTime query string false "Start time (RFC3339 format)"
// @Param endTime query string false "End time (RFC3339 format)"
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(10)
// @Success 200 {object} usecase.PaginatedResponse[domain.Event] "List of events"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/calendar/events [get]
func (h *CalendarHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	calendarIDStr := r.URL.Query().Get("calendarId")
	if calendarIDStr == "" {
		http.Error(w, "calendarId parameter is required", http.StatusBadRequest)
		return
	}

	calendarID, err := strconv.ParseUint(calendarIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid calendarId", http.StatusBadRequest)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	req := usecase.ListEventsRequest{
		CalendarID: uint(calendarID),
		Page:       page,
		PageSize:   pageSize,
	}

	// Parse optional time range
	if startTimeStr := r.URL.Query().Get("startTime"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			req.StartTime = &startTime
		}
	}

	if endTimeStr := r.URL.Query().Get("endTime"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			req.EndTime = &endTime
		}
	}

	response, err := h.calendarUC.ListEvents(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to list events", zap.Error(err))
		http.Error(w, "Failed to list events", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetAvailableSlots godoc
// @Summary Get available appointment slots
// @Description Retrieves available appointment slots for a calendar on a specific date
// @Tags Calendar
// @Produce json
// @Param calendarId query int true "Calendar ID"
// @Param date query string true "Date (YYYY-MM-DD format)"
// @Param duration query int false "Slot duration in minutes" default(30)
// @Success 200 {array} domain.AvailabilitySlot "Available slots"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/calendar/appointments/available-slots [get]
func (h *CalendarHandler) GetAvailableSlots(w http.ResponseWriter, r *http.Request) {
	calendarIDStr := r.URL.Query().Get("calendarId")
	if calendarIDStr == "" {
		http.Error(w, "calendarId parameter is required", http.StatusBadRequest)
		return
	}

	calendarID, err := strconv.ParseUint(calendarIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid calendarId", http.StatusBadRequest)
		return
	}

	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		http.Error(w, "date parameter is required", http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "Invalid date format, use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	duration := 30 // Default 30 minutes
	if durationStr := r.URL.Query().Get("duration"); durationStr != "" {
		if d, err := strconv.Atoi(durationStr); err == nil && d > 0 {
			duration = d
		}
	}

	req := usecase.GetAvailableSlotsRequest{
		CalendarID:   uint(calendarID),
		Date:         date,
		SlotDuration: time.Duration(duration) * time.Minute,
	}

	slots, err := h.calendarUC.GetAvailableSlots(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get available slots", zap.Error(err))
		http.Error(w, "Failed to get available slots", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(slots)
}

// BookSlot godoc
// @Summary Book an appointment slot
// @Description Books an available appointment slot
// @Tags Calendar
// @Accept json
// @Produce json
// @Param request body usecase.BookSlotRequest true "Booking details"
// @Success 201 {object} domain.Booking "Booking created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/calendar/appointments/book [post]
func (h *CalendarHandler) BookSlot(w http.ResponseWriter, r *http.Request) {
	var req usecase.BookSlotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode book slot request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	booking, err := h.calendarUC.BookSlot(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to book slot", zap.Error(err))
		http.Error(w, "Failed to book slot", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(booking)
}

// IsBusinessDay godoc
// @Summary Check if date is a business day
// @Description Checks if a given date is a business day
// @Tags Calendar
// @Produce json
// @Param date path string true "Date (YYYY-MM-DD format)"
// @Success 200 {object} map[string]bool "Business day status"
// @Failure 400 {object} map[string]string "Invalid date format"
// @Router /api/calendar/business-day/{date} [get]
func (h *CalendarHandler) IsBusinessDay(w http.ResponseWriter, r *http.Request) {
	dateStr := chi.URLParam(r, "date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "Invalid date format, use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	isBusinessDay := h.calendarUC.IsBusinessDay(r.Context(), date)

	response := map[string]interface{}{
		"date":          dateStr,
		"isBusinessDay": isBusinessDay,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
