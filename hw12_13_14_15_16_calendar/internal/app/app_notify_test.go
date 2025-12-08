package app

import (
	"context"
	"testing"
	"time"

	"github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/logger"
	"github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/storage"
	memorystorage "github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/storage/memory"
)

func TestApp_EventsToNotify(t *testing.T) {
	logg := logger.New("debug")
	store := memorystorage.New()
	app := New(logg, store)
	ctx := context.Background()

	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	// Создаем событие, которое требует уведомления
	// Событие через 1 час, уведомление за 1 час - значит время уведомления = now, событие должно быть включено
	event := storage.Event{
		ID:           "event-1",
		Title:        "Test Event",
		At:           now.Add(1 * time.Hour),
		NotifyBefore: 1 * time.Hour,
		UserID:       "user-1",
	}

	if err := app.CreateEvent(ctx, event); err != nil {
		t.Fatalf("failed to create event: %v", err)
	}

	// Получаем события для уведомления
	events, err := app.EventsToNotify(ctx, now)
	if err != nil {
		t.Fatalf("failed to get events to notify: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if events[0].ID != event.ID {
		t.Errorf("expected event ID %s, got %s", event.ID, events[0].ID)
	}
}

func TestApp_DeleteOldEvents(t *testing.T) {
	logg := logger.New("debug")
	store := memorystorage.New()
	app := New(logg, store)
	ctx := context.Background()

	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	oneYearAgo := now.AddDate(-1, 0, 0)

	// Создаем старое событие
	oldEvent := storage.Event{
		ID:     "old-event",
		Title:  "Old Event",
		At:     oneYearAgo.Add(-1 * time.Hour),
		UserID: "user-1",
	}

	if err := app.CreateEvent(ctx, oldEvent); err != nil {
		t.Fatalf("failed to create old event: %v", err)
	}

	// Удаляем старые события
	deletedCount, err := app.DeleteOldEvents(ctx, oneYearAgo)
	if err != nil {
		t.Fatalf("failed to delete old events: %v", err)
	}

	if deletedCount != 1 {
		t.Errorf("expected 1 deleted event, got %d", deletedCount)
	}

	// Проверяем, что событие удалено
	_, err = app.GetEvent(ctx, oldEvent.ID)
	if err == nil {
		t.Error("old event should be deleted")
	}
}
