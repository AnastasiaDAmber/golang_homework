package memorystorage

import (
	"context"
	"testing"
	"time"

	"github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/storage"
)

func TestStorage_EventsToNotify(t *testing.T) {
	s := New()
	ctx := context.Background()

	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	// Событие, которое требует уведомления (время уведомления уже прошло, но событие еще не произошло)
	// Событие через 1 час, уведомление за 1 час - значит время уведомления = now, событие должно быть включено
	event1 := storage.Event{
		ID:           "event-1",
		Title:        "Event 1",
		At:           now.Add(1 * time.Hour), // событие через 1 час
		NotifyBefore: 1 * time.Hour,          // уведомление за 1 час
		UserID:       "user-1",
	}

	// Событие, которое еще не требует уведомления (время уведомления еще не наступило)
	event2 := storage.Event{
		ID:           "event-2",
		Title:        "Event 2",
		At:           now.Add(3 * time.Hour), // событие через 3 часа
		NotifyBefore: 2 * time.Hour,          // уведомление за 2 часа (еще не наступило)
		UserID:       "user-2",
	}

	// Событие без уведомления
	event3 := storage.Event{
		ID:     "event-3",
		Title:  "Event 3",
		At:     now.Add(4 * time.Hour),
		UserID: "user-3",
	}

	// Событие, которое уже произошло
	event4 := storage.Event{
		ID:           "event-4",
		Title:        "Event 4",
		At:           now.Add(-1 * time.Hour), // событие уже прошло
		NotifyBefore: 1 * time.Hour,
		UserID:       "user-4",
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
	if err := s.CreateEvent(ctx, event4); err != nil {
		t.Fatalf("failed to create event4: %v", err)
	}

	// Получаем события для уведомления
	events, err := s.EventsToNotify(ctx, now)
	if err != nil {
		t.Fatalf("failed to get events to notify: %v", err)
	}

	// Должно быть только event1 (event2 еще не требует уведомления, event3 без уведомления, event4 уже прошло)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if events[0].ID != event1.ID {
		t.Errorf("expected event ID %s, got %s", event1.ID, events[0].ID)
	}
}

func TestStorage_DeleteOldEvents(t *testing.T) {
	s := New()
	ctx := context.Background()

	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	oneYearAgo := now.AddDate(-1, 0, 0)

	// Старое событие (более года назад)
	oldEvent := storage.Event{
		ID:     "old-event",
		Title:  "Old Event",
		At:     oneYearAgo.Add(-1 * time.Hour),
		UserID: "user-1",
	}

	// Событие, которое еще не старое
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

func TestStorage_EventsToNotify_EdgeCases(t *testing.T) {
	s := New()
	ctx := context.Background()

	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	// Событие, где время уведомления точно сейчас
	event1 := storage.Event{
		ID:           "event-1",
		Title:        "Event 1",
		At:           now.Add(1 * time.Hour),
		NotifyBefore: 1 * time.Hour, // уведомление должно быть сейчас
		UserID:       "user-1",
	}

	if err := s.CreateEvent(ctx, event1); err != nil {
		t.Fatalf("failed to create event: %v", err)
	}

	events, err := s.EventsToNotify(ctx, now)
	if err != nil {
		t.Fatalf("failed to get events: %v", err)
	}

	// Событие должно быть включено (время уведомления наступило или прошло)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	// Событие, которое происходит прямо сейчас
	event2 := storage.Event{
		ID:           "event-2",
		Title:        "Event 2",
		At:           now,
		NotifyBefore: 1 * time.Hour,
		UserID:       "user-2",
	}

	if err := s.CreateEvent(ctx, event2); err != nil {
		t.Fatalf("failed to create event: %v", err)
	}

	events, err = s.EventsToNotify(ctx, now)
	if err != nil {
		t.Fatalf("failed to get events: %v", err)
	}

	// Событие, которое происходит сейчас, не должно быть включено (уже произошло)
	if len(events) != 1 {
		t.Fatalf("expected 1 event (event2 should not be included), got %d", len(events))
	}
}
