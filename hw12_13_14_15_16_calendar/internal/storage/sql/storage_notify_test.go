package sqlstorage

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/storage"
)

func TestStorage_EventsToNotify(t *testing.T) {
	// Используем in-memory SQLite для тестов
	dsn := "file::memory:?cache=shared"
	s := New(dsn)
	ctx := context.Background()

	if err := s.Connect(ctx); err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer s.Close(ctx)

	// Создаем таблицу
	_, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS events (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			at TIMESTAMP NOT NULL,
			duration TEXT,
			description TEXT,
			user_id TEXT,
			notify_before TEXT
		)
	`)
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	// Событие, которое требует уведомления
	event1 := storage.Event{
		ID:           "event-1",
		Title:        "Event 1",
		At:           now.Add(2 * time.Hour),
		NotifyBefore: 1 * time.Hour,
		UserID:       "user-1",
	}

	// Событие, которое еще не требует уведомления
	event2 := storage.Event{
		ID:           "event-2",
		Title:        "Event 2",
		At:           now.Add(3 * time.Hour),
		NotifyBefore: 2 * time.Hour,
		UserID:       "user-2",
	}

	// Событие без уведомления
	event3 := storage.Event{
		ID:     "event-3",
		Title:  "Event 3",
		At:     now.Add(4 * time.Hour),
		UserID: "user-3",
	}

	// Создаем события
	if err := s.CreateEvent(ctx, event1); err != nil {
		t.Fatalf("failed to create event1: %v", err)
	}
	if err := s.CreateEvent(ctx, event2); err != nil {
		t.Fatalf("failed to create event2: %v", err)
	}
	if err := s.CreateEvent(ctx, event3); err != nil {
		t.Fatalf("failed to create event3: %v", err)
	}

	// Получаем события для уведомления
	events, err := s.EventsToNotify(ctx, now)
	if err != nil {
		t.Fatalf("failed to get events to notify: %v", err)
	}

	// Должно быть только event1
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if events[0].ID != event1.ID {
		t.Errorf("expected event ID %s, got %s", event1.ID, events[0].ID)
	}
}

func TestStorage_DeleteOldEvents(t *testing.T) {
	// Используем PostgreSQL из переменной окружения или пропускаем тест
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "postgres://calendar:calendar@localhost:5432/calendar_test?sslmode=disable"
	}

	s := New(dsn)
	ctx := context.Background()

	if err := s.Connect(ctx); err != nil {
		t.Skipf("skipping test: failed to connect to database: %v", err)
	}
	defer s.Close(ctx)

	// Создаем таблицу (если не существует)
	_, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS events (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			at TIMESTAMPTZ NOT NULL,
			duration INTERVAL,
			description TEXT,
			user_id TEXT,
			notify_before INTERVAL
		)
	`)
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	// Очищаем таблицу перед тестом
	_, _ = s.db.ExecContext(ctx, "DELETE FROM events")

	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	oneYearAgo := now.AddDate(-1, 0, 0)

	// Старое событие
	oldEvent := storage.Event{
		ID:     "old-event",
		Title:  "Old Event",
		At:     oneYearAgo.Add(-1 * time.Hour),
		UserID: "user-1",
	}

	// Новое событие
	newEvent := storage.Event{
		ID:     "new-event",
		Title:  "New Event",
		At:     now.AddDate(0, -6, 0), // 6 месяцев назад
		UserID: "user-2",
	}

	// Создаем события
	if err := s.CreateEvent(ctx, oldEvent); err != nil {
		t.Fatalf("failed to create old event: %v", err)
	}
	if err := s.CreateEvent(ctx, newEvent); err != nil {
		t.Fatalf("failed to create new event: %v", err)
	}

	// Удаляем старые события
	deletedCount, err := s.DeleteOldEvents(ctx, oneYearAgo)
	if err != nil {
		t.Fatalf("failed to delete old events: %v", err)
	}

	if deletedCount != 1 {
		t.Errorf("expected 1 deleted event, got %d", deletedCount)
	}

	// Проверяем, что старое событие удалено
	_, err = s.GetEvent(ctx, oldEvent.ID)
	if err == nil {
		t.Error("old event should be deleted")
	}

	// Проверяем, что новое событие осталось
	_, err = s.GetEvent(ctx, newEvent.ID)
	if err != nil {
		t.Errorf("new event should not be deleted: %v", err)
	}
}
