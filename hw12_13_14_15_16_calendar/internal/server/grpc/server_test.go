package grpcserver

import (
	"context"
	"testing"
	"time"

	"github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/api/event"
	"github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/logger"
	"github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/storage"
	memorystorage "github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/storage/memory"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
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

func TestGRPCCreateEvent(t *testing.T) {
	logg := logger.New("debug")
	app := newMockApp()
	server := NewServer(logg, app, "127.0.0.1", 18081)

	now := time.Now()
	req := &event.CreateEventRequest{
		Event: &event.Event{
			Id:           "grpc-test-1",
			Title:        "GRPC Test Event",
			At:           timestamppb.New(now),
			Duration:     durationpb.New(time.Hour),
			Description:  "Test description",
			UserId:       "user1",
			NotifyBefore: durationpb.New(15 * time.Minute),
		},
	}

	resp, err := server.CreateEvent(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateEvent failed: %v", err)
	}

	if resp.Id != "grpc-test-1" {
		t.Fatalf("expected id 'grpc-test-1', got '%s'", resp.Id)
	}
}

func TestGRPCCreateEventDuplicate(t *testing.T) {
	logg := logger.New("debug")
	app := newMockApp()
	server := NewServer(logg, app, "127.0.0.1", 18081)

	now := time.Now()
	req := &event.CreateEventRequest{
		Event: &event.Event{
			Id:    "grpc-duplicate-1",
			Title: "First Event",
			At:    timestamppb.New(now),
		},
	}

	// Первое создание должно пройти
	_, err := server.CreateEvent(context.Background(), req)
	if err != nil {
		t.Fatalf("first CreateEvent failed: %v", err)
	}

	// Попытка создать дубликат
	_, err = server.CreateEvent(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for duplicate event")
	}

	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.AlreadyExists {
		t.Fatalf("expected AlreadyExists error, got %v", err)
	}
}

func TestGRPCGetEvent(t *testing.T) {
	logg := logger.New("debug")
	app := newMockApp()
	server := NewServer(logg, app, "127.0.0.1", 18081)

	// Сначала создаем событие
	now := time.Now()
	createReq := &event.CreateEventRequest{
		Event: &event.Event{
			Id:    "grpc-get-1",
			Title: "Get Test Event",
			At:    timestamppb.New(now),
		},
	}
	_, _ = server.CreateEvent(context.Background(), createReq)

	// Получаем событие
	req := &event.GetEventRequest{Id: "grpc-get-1"}
	resp, err := server.GetEvent(context.Background(), req)
	if err != nil {
		t.Fatalf("GetEvent failed: %v", err)
	}

	if resp.Event.Id != "grpc-get-1" {
		t.Fatalf("expected id 'grpc-get-1', got '%s'", resp.Event.Id)
	}

	if resp.Event.Title != "Get Test Event" {
		t.Fatalf("expected title 'Get Test Event', got '%s'", resp.Event.Title)
	}
}

func TestGRPCGetEventNotFound(t *testing.T) {
	logg := logger.New("debug")
	app := newMockApp()
	server := NewServer(logg, app, "127.0.0.1", 18081)

	req := &event.GetEventRequest{Id: "non-existent"}
	_, err := server.GetEvent(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for non-existent event")
	}

	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.NotFound {
		t.Fatalf("expected NotFound error, got %v", err)
	}
}

func TestGRPCUpdateEvent(t *testing.T) {
	logg := logger.New("debug")
	app := newMockApp()
	server := NewServer(logg, app, "127.0.0.1", 18081)

	// Сначала создаем событие
	now := time.Now()
	createReq := &event.CreateEventRequest{
		Event: &event.Event{
			Id:    "grpc-update-1",
			Title: "Original Title",
			At:    timestamppb.New(now),
		},
	}
	_, _ = server.CreateEvent(context.Background(), createReq)

	// Обновляем событие
	updateReq := &event.UpdateEventRequest{
		Event: &event.Event{
			Id:    "grpc-update-1",
			Title: "Updated Title",
			At:    timestamppb.New(now),
		},
	}

	resp, err := server.UpdateEvent(context.Background(), updateReq)
	if err != nil {
		t.Fatalf("UpdateEvent failed: %v", err)
	}

	if !resp.Success {
		t.Fatal("expected success=true")
	}

	// Проверяем, что событие обновилось
	getReq := &event.GetEventRequest{Id: "grpc-update-1"}
	getResp, _ := server.GetEvent(context.Background(), getReq)
	if getResp.Event.Title != "Updated Title" {
		t.Fatalf("expected title 'Updated Title', got '%s'", getResp.Event.Title)
	}
}

func TestGRPCDeleteEvent(t *testing.T) {
	logg := logger.New("debug")
	app := newMockApp()
	server := NewServer(logg, app, "127.0.0.1", 18081)

	// Сначала создаем событие
	now := time.Now()
	createReq := &event.CreateEventRequest{
		Event: &event.Event{
			Id:    "grpc-delete-1",
			Title: "To Delete",
			At:    timestamppb.New(now),
		},
	}
	_, _ = server.CreateEvent(context.Background(), createReq)

	// Удаляем событие
	req := &event.DeleteEventRequest{Id: "grpc-delete-1"}
	resp, err := server.DeleteEvent(context.Background(), req)
	if err != nil {
		t.Fatalf("DeleteEvent failed: %v", err)
	}

	if !resp.Success {
		t.Fatal("expected success=true")
	}

	// Проверяем, что событие удалено
	getReq := &event.GetEventRequest{Id: "grpc-delete-1"}
	_, err = server.GetEvent(context.Background(), getReq)
	if err == nil {
		t.Fatal("expected error for deleted event")
	}

	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.NotFound {
		t.Fatalf("expected NotFound error, got %v", err)
	}
}

func TestGRPCListEvents(t *testing.T) {
	logg := logger.New("debug")
	app := newMockApp()
	server := NewServer(logg, app, "127.0.0.1", 18081)

	// Создаем несколько событий
	now := time.Now()
	events := []*event.CreateEventRequest{
		{
			Event: &event.Event{
				Id:    "grpc-list-1",
				Title: "Event 1",
				At:    timestamppb.New(now),
			},
		},
		{
			Event: &event.Event{
				Id:    "grpc-list-2",
				Title: "Event 2",
				At:    timestamppb.New(now.Add(time.Hour)),
			},
		},
	}

	for _, e := range events {
		_, _ = server.CreateEvent(context.Background(), e)
	}

	req := &event.ListEventsRequest{}
	resp, err := server.ListEvents(context.Background(), req)
	if err != nil {
		t.Fatalf("ListEvents failed: %v", err)
	}

	if len(resp.Events) < 2 {
		t.Fatalf("expected at least 2 events, got %d", len(resp.Events))
	}
}

func TestGRPCListEventsDay(t *testing.T) {
	logg := logger.New("debug")
	app := newMockApp()
	server := NewServer(logg, app, "127.0.0.1", 18081)

	// Используем UTC для избежания проблем с timezone
	now := time.Now().UTC()
	// Нормализуем к началу дня
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	// Создаем событие в середине дня (12:00)
	eventTime := dayStart.Add(12 * time.Hour)

	// Создаем событие на сегодня
	createReq := &event.CreateEventRequest{
		Event: &event.Event{
			Id:    "grpc-day-1",
			Title: "Today Event",
			At:    timestamppb.New(eventTime),
		},
	}
	_, err := server.CreateEvent(context.Background(), createReq)
	if err != nil {
		t.Fatalf("failed to create event: %v", err)
	}

	// Проверяем, что событие создано
	allEvents, _ := app.ListEvents(context.Background())
	if len(allEvents) == 0 {
		t.Fatal("event was not created")
	}

	// Используем тот же dayStart для запроса
	req := &event.ListEventsDayRequest{
		DayStart: timestamppb.New(dayStart),
	}

	resp, err := server.ListEventsDay(context.Background(), req)
	if err != nil {
		t.Fatalf("ListEventsDay failed: %v", err)
	}

	if len(resp.Events) == 0 {
		// Отладочная информация
		t.Logf("dayStart: %v, eventTime: %v", dayStart, eventTime)
		t.Logf("All events: %+v", allEvents)
		dayEvents, _ := app.ListEventsDay(context.Background(), dayStart)
		t.Logf("Direct ListEventsDay result: %+v", dayEvents)
		t.Fatalf("expected at least 1 event for today, got %d. All events: %d", len(resp.Events), len(allEvents))
	}
}

func TestGRPCListEventsWeek(t *testing.T) {
	logg := logger.New("debug")
	app := newMockApp()
	server := NewServer(logg, app, "127.0.0.1", 18081)

	now := time.Now()
	weekStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Создаем событие на этой неделе
	createReq := &event.CreateEventRequest{
		Event: &event.Event{
			Id:    "grpc-week-1",
			Title: "Week Event",
			At:    timestamppb.New(weekStart.Add(24 * time.Hour)),
		},
	}
	_, _ = server.CreateEvent(context.Background(), createReq)

	req := &event.ListEventsWeekRequest{
		WeekStart: timestamppb.New(weekStart),
	}

	resp, err := server.ListEventsWeek(context.Background(), req)
	if err != nil {
		t.Fatalf("ListEventsWeek failed: %v", err)
	}

	if len(resp.Events) == 0 {
		t.Fatal("expected at least 1 event for this week")
	}
}

func TestGRPCListEventsMonth(t *testing.T) {
	logg := logger.New("debug")
	app := newMockApp()
	server := NewServer(logg, app, "127.0.0.1", 18081)

	// Используем UTC для избежания проблем с timezone
	now := time.Now().UTC()
	// Нормализуем к началу месяца
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	// Создаем событие на 5-й день месяца в 12:00
	eventTime := monthStart.Add(5*24*time.Hour + 12*time.Hour)

	// Создаем событие в этом месяце
	createReq := &event.CreateEventRequest{
		Event: &event.Event{
			Id:    "grpc-month-1",
			Title: "Month Event",
			At:    timestamppb.New(eventTime),
		},
	}
	_, err := server.CreateEvent(context.Background(), createReq)
	if err != nil {
		t.Fatalf("failed to create event: %v", err)
	}

	// Проверяем, что событие создано
	allEvents, _ := app.ListEvents(context.Background())
	if len(allEvents) == 0 {
		t.Fatal("event was not created")
	}

	req := &event.ListEventsMonthRequest{
		MonthStart: timestamppb.New(monthStart),
	}

	resp, err := server.ListEventsMonth(context.Background(), req)
	if err != nil {
		t.Fatalf("ListEventsMonth failed: %v", err)
	}

	if len(resp.Events) == 0 {
		// Отладочная информация
		t.Logf("monthStart: %v, eventTime: %v", monthStart, eventTime)
		t.Logf("All events: %+v", allEvents)
		monthEvents, _ := app.ListEventsMonth(context.Background(), monthStart)
		t.Logf("Direct ListEventsMonth result: %+v", monthEvents)
		t.Fatalf("expected at least 1 event for this month, got %d. All events: %d", len(resp.Events), len(allEvents))
	}
}

func TestGRPCCreateEventInvalidRequest(t *testing.T) {
	logg := logger.New("debug")
	app := newMockApp()
	server := NewServer(logg, app, "127.0.0.1", 18081)

	req := &event.CreateEventRequest{
		Event: nil, // nil event
	}

	_, err := server.CreateEvent(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for nil event")
	}

	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument error, got %v", err)
	}
}

func TestGRPCUpdateEventNotFound(t *testing.T) {
	logg := logger.New("debug")
	app := newMockApp()
	server := NewServer(logg, app, "127.0.0.1", 18081)

	now := time.Now()
	req := &event.UpdateEventRequest{
		Event: &event.Event{
			Id:    "non-existent",
			Title: "Updated",
			At:    timestamppb.New(now),
		},
	}

	_, err := server.UpdateEvent(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for non-existent event")
	}

	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.NotFound {
		t.Fatalf("expected NotFound error, got %v", err)
	}
}
