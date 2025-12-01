package internalhttp

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/storage"
)

type createEventRequest struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	At           string `json:"at"` // RFC3339 format
	Duration     string `json:"duration"` // Go duration format (e.g., "1h30m")
	Description  string `json:"description"`
	UserID       string `json:"user_id"`
	NotifyBefore string `json:"notify_before"` // Go duration format
}

type updateEventRequest struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	At           string `json:"at"` // RFC3339 format
	Duration     string `json:"duration"` // Go duration format
	Description  string `json:"description"`
	UserID       string `json:"user_id"`
	NotifyBefore string `json:"notify_before"` // Go duration format
}

type eventResponse struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	At           string `json:"at"` // RFC3339 format
	Duration     string `json:"duration"` // Go duration format
	Description  string `json:"description"`
	UserID       string `json:"user_id"`
	NotifyBefore string `json:"notify_before"` // Go duration format
}

type errorResponse struct {
	Error string `json:"error"`
}

func domainEventToResponse(e storage.Event) eventResponse {
	resp := eventResponse{
		ID:          e.ID,
		Title:       e.Title,
		Description: e.Description,
		UserID:      e.UserID,
	}

	if !e.At.IsZero() {
		resp.At = e.At.Format(time.RFC3339)
	}
	if e.Duration != 0 {
		resp.Duration = e.Duration.String()
	}
	if e.NotifyBefore != 0 {
		resp.NotifyBefore = e.NotifyBefore.String()
	}

	return resp
}

func (s *Server) createEventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req createEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	at, err := time.Parse(time.RFC3339, req.At)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid at format. Use RFC3339 format")
		return
	}

	var duration, notifyBefore time.Duration
	if req.Duration != "" {
		duration, err = time.ParseDuration(req.Duration)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid duration format. Use Go duration format (e.g., '1h30m')")
			return
		}
	}
	if req.NotifyBefore != "" {
		notifyBefore, err = time.ParseDuration(req.NotifyBefore)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid notify_before format. Use Go duration format")
			return
		}
	}

	event := storage.Event{
		ID:           req.ID,
		Title:        req.Title,
		At:           at,
		Duration:     duration,
		Description:  req.Description,
		UserID:       req.UserID,
		NotifyBefore: notifyBefore,
	}

	if err := s.app.CreateEvent(r.Context(), event); err != nil {
		if errors.Is(err, storage.ErrDateBusy) {
			respondError(w, http.StatusConflict, "Event with this ID already exists")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, eventResponse{ID: event.ID})
}

func (s *Server) updateEventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req updateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	at, err := time.Parse(time.RFC3339, req.At)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid at format. Use RFC3339 format")
		return
	}

	var duration, notifyBefore time.Duration
	if req.Duration != "" {
		duration, err = time.ParseDuration(req.Duration)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid duration format. Use Go duration format")
			return
		}
	}
	if req.NotifyBefore != "" {
		notifyBefore, err = time.ParseDuration(req.NotifyBefore)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid notify_before format. Use Go duration format")
			return
		}
	}

	event := storage.Event{
		ID:           req.ID,
		Title:        req.Title,
		At:           at,
		Duration:     duration,
		Description:  req.Description,
		UserID:       req.UserID,
		NotifyBefore: notifyBefore,
	}

	if err := s.app.UpdateEvent(r.Context(), event); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			respondError(w, http.StatusNotFound, "Event not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (s *Server) deleteEventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "id parameter is required")
		return
	}

	if err := s.app.DeleteEvent(r.Context(), id); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			respondError(w, http.StatusNotFound, "Event not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (s *Server) getEventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "id parameter is required")
		return
	}

	event, err := s.app.GetEvent(r.Context(), id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			respondError(w, http.StatusNotFound, "Event not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, domainEventToResponse(event))
}

func (s *Server) listEventsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	events, err := s.app.ListEvents(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := make([]eventResponse, 0, len(events))
	for _, e := range events {
		response = append(response, domainEventToResponse(e))
	}

	respondJSON(w, http.StatusOK, response)
}

func (s *Server) listEventsDayHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dayStartStr := r.URL.Query().Get("day_start")
	if dayStartStr == "" {
		respondError(w, http.StatusBadRequest, "day_start parameter is required (RFC3339 format)")
		return
	}

	dayStart, err := time.Parse(time.RFC3339, dayStartStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid day_start format. Use RFC3339 format")
		return
	}

	events, err := s.app.ListEventsDay(r.Context(), dayStart)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := make([]eventResponse, 0, len(events))
	for _, e := range events {
		response = append(response, domainEventToResponse(e))
	}

	respondJSON(w, http.StatusOK, response)
}

func (s *Server) listEventsWeekHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	weekStartStr := r.URL.Query().Get("week_start")
	if weekStartStr == "" {
		respondError(w, http.StatusBadRequest, "week_start parameter is required (RFC3339 format)")
		return
	}

	weekStart, err := time.Parse(time.RFC3339, weekStartStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid week_start format. Use RFC3339 format")
		return
	}

	events, err := s.app.ListEventsWeek(r.Context(), weekStart)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := make([]eventResponse, 0, len(events))
	for _, e := range events {
		response = append(response, domainEventToResponse(e))
	}

	respondJSON(w, http.StatusOK, response)
}

func (s *Server) listEventsMonthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	monthStartStr := r.URL.Query().Get("month_start")
	if monthStartStr == "" {
		respondError(w, http.StatusBadRequest, "month_start parameter is required (RFC3339 format)")
		return
	}

	monthStart, err := time.Parse(time.RFC3339, monthStartStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid month_start format. Use RFC3339 format")
		return
	}

	events, err := s.app.ListEventsMonth(r.Context(), monthStart)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := make([]eventResponse, 0, len(events))
	for _, e := range events {
		response = append(response, domainEventToResponse(e))
	}

	respondJSON(w, http.StatusOK, response)
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, errorResponse{Error: message})
}

