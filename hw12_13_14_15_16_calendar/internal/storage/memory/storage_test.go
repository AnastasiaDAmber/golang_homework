package memorystorage

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/storage"
)

func TestStorageBasicCRUD(t *testing.T) {
	s := New()
	ctx := context.Background()

	e := storage.Event{
		ID:    "1",
		Title: "test",
		At:    time.Now(),
	}

	if err := s.CreateEvent(ctx, e); err != nil {
		t.Fatalf("create failed: %v", err)
	}

	got, err := s.GetEvent(ctx, "1")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if got.ID != e.ID || got.Title != e.Title {
		t.Fatalf("mismatch got=%v want=%v", got, e)
	}

	e.Title = "updated"
	if err := s.UpdateEvent(ctx, e); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	got, _ = s.GetEvent(ctx, "1")
	if got.Title != "updated" {
		t.Fatalf("update didn't apply")
	}

	if err := s.DeleteEvent(ctx, "1"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	_, err = s.GetEvent(ctx, "1")
	if err == nil {
		t.Fatalf("expected not found after delete")
	}
}

func TestStorageConcurrency(t *testing.T) {
	s := New()
	ctx := context.Background()
	n := 100
	wg := sync.WaitGroup{}
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			e := storage.Event{
				ID:    string(rune(i + 65)),
				Title: "t",
				At:    time.Now(),
			}
			_ = s.CreateEvent(ctx, e)
		}()
	}
	wg.Wait()

	events, err := s.ListEvents(ctx)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(events) == 0 {
		t.Fatalf("expected some events")
	}
}

func TestStorageBusinessErrors(t *testing.T) {
	s := New()
	ctx := context.Background()

	// Тест ErrDateBusy - попытка создать событие с существующим ID
	e := storage.Event{
		ID:    "test-id",
		Title: "test",
		At:    time.Now(),
	}

	if err := s.CreateEvent(ctx, e); err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	// Попытка создать событие с тем же ID должна вернуть ErrDateBusy
	err := s.CreateEvent(ctx, e)
	if err == nil {
		t.Fatalf("expected ErrDateBusy when creating duplicate event")
	}
	if !errors.Is(err, storage.ErrDateBusy) {
		t.Fatalf("expected ErrDateBusy, got: %v", err)
	}

	// Тест ErrNotFound - попытка получить несуществующее событие
	_, err = s.GetEvent(ctx, "non-existent")
	if err == nil {
		t.Fatalf("expected ErrNotFound when getting non-existent event")
	}
	if !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}

	// Тест ErrNotFound - попытка обновить несуществующее событие
	e.ID = "non-existent"
	err = s.UpdateEvent(ctx, e)
	if err == nil {
		t.Fatalf("expected ErrNotFound when updating non-existent event")
	}
	if !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}

	// Тест ErrNotFound - попытка удалить несуществующее событие
	err = s.DeleteEvent(ctx, "non-existent")
	if err == nil {
		t.Fatalf("expected ErrNotFound when deleting non-existent event")
	}
	if !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}
