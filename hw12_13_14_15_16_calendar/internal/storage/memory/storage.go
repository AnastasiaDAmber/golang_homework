package memorystorage

import (
	"context"
	"sync"
	"time"

	"github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/storage"
)

// Используем общие ошибки из пакета storage

type Storage struct {
	mu     sync.RWMutex
	events map[string]storage.Event
}

func New() *Storage {
	return &Storage{
		events: make(map[string]storage.Event),
	}
}

func (s *Storage) CreateEvent(_ context.Context, e storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.events[e.ID]; ok {
		return storage.ErrDateBusy
	}
	s.events[e.ID] = e
	return nil
}

func (s *Storage) UpdateEvent(_ context.Context, e storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.events[e.ID]; !ok {
		return storage.ErrNotFound
	}
	s.events[e.ID] = e
	return nil
}

func (s *Storage) DeleteEvent(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.events[id]; !ok {
		return storage.ErrNotFound
	}
	delete(s.events, id)
	return nil
}

func (s *Storage) GetEvent(_ context.Context, id string) (storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.events[id]
	if !ok {
		return storage.Event{}, storage.ErrNotFound
	}
	return e, nil
}

func (s *Storage) ListEvents(_ context.Context) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]storage.Event, 0, len(s.events))
	for _, v := range s.events {
		out = append(out, v)
	}
	return out, nil
}

func (s *Storage) ListEventsDay(_ context.Context, dayStart time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	start := time.Date(dayStart.Year(), dayStart.Month(), dayStart.Day(), 0, 0, 0, 0, dayStart.Location())
	end := start.Add(24 * time.Hour)
	out := []storage.Event{}
	for _, ev := range s.events {
		if ev.At.Equal(start) || (ev.At.After(start) && ev.At.Before(end)) {
			out = append(out, ev)
		}
	}
	return out, nil
}

func (s *Storage) ListEventsWeek(_ context.Context, weekStart time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	start := time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())
	end := start.Add(7 * 24 * time.Hour)
	out := []storage.Event{}
	for _, ev := range s.events {
		if (ev.At.Equal(start) || ev.At.After(start)) && ev.At.Before(end) {
			out = append(out, ev)
		}
	}
	return out, nil
}

func (s *Storage) ListEventsMonth(_ context.Context, monthStart time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	start := time.Date(monthStart.Year(), monthStart.Month(), 1, 0, 0, 0, 0, monthStart.Location())
	end := start.AddDate(0, 1, 0)
	out := []storage.Event{}
	for _, ev := range s.events {
		if (ev.At.Equal(start) || ev.At.After(start)) && ev.At.Before(end) {
			out = append(out, ev)
		}
	}
	return out, nil
}
