package app

import (
	"context"
	"time"

	"github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/storage"
)

type App struct {
	logger Logger
	store  Storage
}

type Logger interface {
	Info(msg string)
	Error(msg string)
	Debug(msg string)
}

type Storage interface {
	CreateEvent(ctx context.Context, e storage.Event) error
	UpdateEvent(ctx context.Context, e storage.Event) error
	DeleteEvent(ctx context.Context, id string) error
	GetEvent(ctx context.Context, id string) (storage.Event, error)
	ListEvents(ctx context.Context) ([]storage.Event, error)

	ListEventsDay(ctx context.Context, dayStart time.Time) ([]storage.Event, error)
	ListEventsWeek(ctx context.Context, weekStart time.Time) ([]storage.Event, error)
	ListEventsMonth(ctx context.Context, monthStart time.Time) ([]storage.Event, error)
}

func New(logger Logger, storage Storage) *App {
	return &App{
		logger: logger,
		store:  storage,
	}
}

func (a *App) CreateEvent(ctx context.Context, e storage.Event) error {
	a.logger.Debug("CreateEvent called")
	return a.store.CreateEvent(ctx, e)
}

func (a *App) UpdateEvent(ctx context.Context, e storage.Event) error {
	a.logger.Debug("UpdateEvent called")
	return a.store.UpdateEvent(ctx, e)
}

func (a *App) DeleteEvent(ctx context.Context, id string) error {
	a.logger.Debug("DeleteEvent called")
	return a.store.DeleteEvent(ctx, id)
}

func (a *App) GetEvent(ctx context.Context, id string) (storage.Event, error) {
	return a.store.GetEvent(ctx, id)
}

func (a *App) ListEvents(ctx context.Context) ([]storage.Event, error) {
	return a.store.ListEvents(ctx)
}
