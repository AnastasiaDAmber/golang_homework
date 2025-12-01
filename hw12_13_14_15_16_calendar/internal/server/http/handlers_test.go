package internalhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/logger"
	"github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/storage"
	memorystorage "github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/storage/memory"
)

type mockApp struct {
	storage *memorystorage.Storage
}

func newMockApp() *mockApp {
	return &mockApp{
		storage: memorystorage.New(),
	}
}

func (m *mockApp) CreateEvent(ctx context.Context, e storage.Event) error {
	return m.storage.CreateEvent(ctx, e)
}

func (m *mockApp) UpdateEvent(ctx context.Context, e storage.Event) error {
	return m.storage.UpdateEvent(ctx, e)
}

func (m *mockApp) DeleteEvent(ctx context.Context, id string) error {
	return m.storage.DeleteEvent(ctx, id)
}

func (m *mockApp) GetEvent(ctx context.Context, id string) (storage.Event, error) {
	return m.storage.GetEvent(ctx, id)
}

func (m *mockApp) ListEvents(ctx context.Context) ([]storage.Event, error) {
	return m.storage.ListEvents(ctx)
}

func (m *mockApp) ListEventsDay(ctx context.Context, dayStart time.Time) ([]storage.Event, error) {
	return m.storage.ListEventsDay(ctx, dayStart)
}

func (m *mockApp) ListEventsWeek(ctx context.Context, weekStart time.Time) ([]storage.Event, error) {
	return m.storage.ListEventsWeek(ctx, weekStart)
}

func (m *mockApp) ListEventsMonth(ctx context.Context, monthStart time.Time) ([]storage.Event, error) {
	return m.storage.ListEventsMonth(ctx, monthStart)
}

func TestCreateEventHandler(t *testing.T) {
	logg := logger.New("debug")
	app := newMockApp()
	server := NewServer(logg, app, "127.0.0.1", 18080)

	eventData := map[string]interface{}{
		"id":            "test-1",
		"title":         "Test Event",
		"at":            time.Now().Format(time.RFC3339),
		"duration":      "1h",
		"description":   "Test description",
		"user_id":       "user1",
		"notify_before": "15m",
	}

	body, _ := json.Marshal(eventData)
	req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.createEventHandler(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}

	var resp eventResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.ID != "test-1" {
		t.Fatalf("expected id 'test-1', got '%s'", resp.ID)
	}
}

func TestCreateEventHandlerDuplicate(t *testing.T) {
	logg := logger.New("debug")
	app := newMockApp()
	server := NewServer(logg, app, "127.0.0.1", 18080)

	eventData := map[string]interface{}{
		"id":    "duplicate-1",
		"title": "Test Event",
		"at":    time.Now().Format(time.RFC3339),
	}

	body, _ := json.Marshal(eventData)
	req1 := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	server.createEventHandler(w1, req1)

	if w1.Code != http.StatusCreated {
		t.Fatalf("first create should succeed, got %d", w1.Code)
	}

	// Попытка создать дубликат
	body2, _ := json.Marshal(eventData)
	req2 := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	server.createEventHandler(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Fatalf("expected status 409 for duplicate, got %d", w2.Code)
	}
}

func TestGetEventHandler(t *testing.T) {
	logg := logger.New("debug")
	app := newMockApp()
	server := NewServer(logg, app, "127.0.0.1", 18080)

	// Сначала создаем событие
	event := storage.Event{
		ID:    "get-test-1",
		Title: "Get Test Event",
		At:    time.Now(),
	}
	_ = app.CreateEvent(context.Background(), event)

	req := httptest.NewRequest(http.MethodGet, "/api/events/get?id=get-test-1", nil)
	w := httptest.NewRecorder()

	server.getEventHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp eventResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.ID != "get-test-1" {
		t.Fatalf("expected id 'get-test-1', got '%s'", resp.ID)
	}

	if resp.Title != "Get Test Event" {
		t.Fatalf("expected title 'Get Test Event', got '%s'", resp.Title)
	}
}

func TestGetEventHandlerNotFound(t *testing.T) {
	logg := logger.New("debug")
	app := newMockApp()
	server := NewServer(logg, app, "127.0.0.1", 18080)

	req := httptest.NewRequest(http.MethodGet, "/api/events/get?id=non-existent", nil)
	w := httptest.NewRecorder()

	server.getEventHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestUpdateEventHandler(t *testing.T) {
	logg := logger.New("debug")
	app := newMockApp()
	server := NewServer(logg, app, "127.0.0.1", 18080)

	// Сначала создаем событие
	event := storage.Event{
		ID:    "update-test-1",
		Title: "Original Title",
		At:    time.Now(),
	}
	_ = app.CreateEvent(context.Background(), event)

	eventData := map[string]interface{}{
		"id":    "update-test-1",
		"title": "Updated Title",
		"at":    time.Now().Format(time.RFC3339),
	}

	body, _ := json.Marshal(eventData)
	req := httptest.NewRequest(http.MethodPut, "/api/events/update", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.updateEventHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	// Проверяем, что событие действительно обновилось
	updatedEvent, _ := app.GetEvent(context.Background(), "update-test-1")
	if updatedEvent.Title != "Updated Title" {
		t.Fatalf("expected title 'Updated Title', got '%s'", updatedEvent.Title)
	}
}

func TestDeleteEventHandler(t *testing.T) {
	logg := logger.New("debug")
	app := newMockApp()
	server := NewServer(logg, app, "127.0.0.1", 18080)

	// Сначала создаем событие
	event := storage.Event{
		ID:    "delete-test-1",
		Title: "To Delete",
		At:    time.Now(),
	}
	_ = app.CreateEvent(context.Background(), event)

	req := httptest.NewRequest(http.MethodDelete, "/api/events/delete?id=delete-test-1", nil)
	w := httptest.NewRecorder()

	server.deleteEventHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	// Проверяем, что событие удалено
	_, err := app.GetEvent(context.Background(), "delete-test-1")
	if err == nil {
		t.Fatal("expected event to be deleted")
	}
}

func TestListEventsHandler(t *testing.T) {
	logg := logger.New("debug")
	app := newMockApp()
	server := NewServer(logg, app, "127.0.0.1", 18080)

	// Создаем несколько событий
	events := []storage.Event{
		{ID: "list-1", Title: "Event 1", At: time.Now()},
		{ID: "list-2", Title: "Event 2", At: time.Now().Add(time.Hour)},
	}
	for _, e := range events {
		_ = app.CreateEvent(context.Background(), e)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/events", nil)
	w := httptest.NewRecorder()

	server.listEventsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp []eventResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(resp) < 2 {
		t.Fatalf("expected at least 2 events, got %d", len(resp))
	}
}

func TestListEventsDayHandler(t *testing.T) {
	logg := logger.New("debug")
	app := newMockApp()
	server := NewServer(logg, app, "127.0.0.1", 18080)

	now := time.Now()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Создаем событие на сегодня
	event := storage.Event{
		ID:    "day-1",
		Title: "Today Event",
		At:    dayStart.Add(10 * time.Hour),
	}
	_ = app.CreateEvent(context.Background(), event)

	req := httptest.NewRequest(http.MethodGet, "/api/events/day?day_start="+url.QueryEscape(dayStart.Format(time.RFC3339)), nil)
	w := httptest.NewRecorder()

	server.listEventsDayHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp []eventResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(resp) == 0 {
		t.Fatal("expected at least 1 event for today")
	}
}

func TestListEventsWeekHandler(t *testing.T) {
	logg := logger.New("debug")
	app := newMockApp()
	server := NewServer(logg, app, "127.0.0.1", 18080)

	now := time.Now()
	weekStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Создаем событие на этой неделе
	event := storage.Event{
		ID:    "week-1",
		Title: "Week Event",
		At:    weekStart.Add(24 * time.Hour),
	}
	_ = app.CreateEvent(context.Background(), event)

	req := httptest.NewRequest(http.MethodGet, "/api/events/week?week_start="+url.QueryEscape(weekStart.Format(time.RFC3339)), nil)
	w := httptest.NewRecorder()

	server.listEventsWeekHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp []eventResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(resp) == 0 {
		t.Fatal("expected at least 1 event for this week")
	}
}

func TestListEventsMonthHandler(t *testing.T) {
	logg := logger.New("debug")
	app := newMockApp()
	server := NewServer(logg, app, "127.0.0.1", 18080)

	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// Создаем событие в этом месяце
	event := storage.Event{
		ID:    "month-1",
		Title: "Month Event",
		At:    monthStart.Add(5 * 24 * time.Hour),
	}
	_ = app.CreateEvent(context.Background(), event)

	req := httptest.NewRequest(http.MethodGet, "/api/events/month?month_start="+url.QueryEscape(monthStart.Format(time.RFC3339)), nil)
	w := httptest.NewRecorder()

	server.listEventsMonthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp []eventResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(resp) == 0 {
		t.Fatal("expected at least 1 event for this month")
	}
}

func TestCreateEventHandlerInvalidJSON(t *testing.T) {
	logg := logger.New("debug")
	app := newMockApp()
	server := NewServer(logg, app, "127.0.0.1", 18080)

	req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.createEventHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for invalid JSON, got %d", w.Code)
	}
}

func TestCreateEventHandlerInvalidTime(t *testing.T) {
	logg := logger.New("debug")
	app := newMockApp()
	server := NewServer(logg, app, "127.0.0.1", 18080)

	eventData := map[string]interface{}{
		"id":    "test-1",
		"title": "Test",
		"at":    "invalid-time",
	}

	body, _ := json.Marshal(eventData)
	req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.createEventHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for invalid time, got %d", w.Code)
	}
}
