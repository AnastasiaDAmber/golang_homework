package sqlstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/storage"
	"github.com/jmoiron/sqlx"
)

type Storage struct {
	db  *sqlx.DB
	dsn string
}

func New(dsn string) *Storage {
	return &Storage{dsn: dsn}
}

func (s *Storage) Connect(ctx context.Context) error {
	db, err := sqlx.ConnectContext(ctx, "postgres", s.dsn)
	if err != nil {
		return err
	}
	s.db = db
	return nil
}

func (s *Storage) Close(_ context.Context) error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Storage) CreateEvent(ctx context.Context, e storage.Event) error {
	query := `
		INSERT INTO events (id, title, at, duration, description, user_id, notify_before)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := s.db.ExecContext(ctx, query, e.ID, e.Title, e.At, pqInterval(e.Duration),
		e.Description, e.UserID, pqInterval(e.NotifyBefore))
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) UpdateEvent(ctx context.Context, e storage.Event) error {
	query := `
		UPDATE events
		SET title = $2, at = $3, duration = $4, description = $5, user_id = $6, notify_before = $7
		WHERE id = $1
	`
	res, err := s.db.ExecContext(ctx, query, e.ID, e.Title, e.At,
		pqInterval(e.Duration), e.Description, e.UserID, pqInterval(e.NotifyBefore))
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return storage.ErrNotFound
	}
	return nil
}

func (s *Storage) DeleteEvent(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM events WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return storage.ErrNotFound
	}
	return nil
}

func (s *Storage) GetEvent(ctx context.Context, id string) (storage.Event, error) {
	var e struct {
		ID           string         `db:"id"`
		Title        string         `db:"title"`
		At           time.Time      `db:"at"`
		Duration     sql.NullString `db:"duration"`
		Description  sql.NullString `db:"description"`
		UserID       sql.NullString `db:"user_id"`
		NotifyBefore sql.NullString `db:"notify_before"`
	}
	err := s.db.GetContext(ctx, &e, `
		SELECT id, title, at, duration::text as duration, description, user_id, notify_before::text as notify_before 
		FROM events 
		WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.Event{}, storage.ErrNotFound
		}
		return storage.Event{}, err
	}
	ev := storage.Event{
		ID:          e.ID,
		Title:       e.Title,
		At:          e.At,
		Description: nullStringToString(e.Description),
		UserID:      nullStringToString(e.UserID),
	}
	if e.Duration.Valid {
		if d, err := time.ParseDuration(sqlIntervalToDurationString(e.Duration.String)); err == nil {
			ev.Duration = d
		}
	}
	if e.NotifyBefore.Valid {
		if d, err := time.ParseDuration(sqlIntervalToDurationString(e.NotifyBefore.String)); err == nil {
			ev.NotifyBefore = d
		}
	}
	return ev, nil
}

func pqInterval(d time.Duration) interface{} {
	if d == 0 {
		return nil
	}
	return d.String()
}

func nullStringToString(ns sql.NullString) string {
	if !ns.Valid {
		return ""
	}
	return ns.String
}

func sqlIntervalToDurationString(s string) string {
	// if it is of form HH:MM:SS
	if len(s) >= 5 && s[2] == ':' {
		// parse HH:MM:SS -> convert to hours + minutes + seconds
		var hh, mm, ss int
		_, err := fmt.Sscanf(s, "%d:%d:%d", &hh, &mm, &ss)
		if err == nil {
			return fmt.Sprintf("%dh%dm%ds", hh, mm, ss)
		}
	}
	return s
}

func (s *Storage) rowsToEvents(rows *sqlx.Rows) ([]storage.Event, error) {
	out := make([]storage.Event, 0)

	for rows.Next() {
		var e struct {
			ID           string         `db:"id"`
			Title        string         `db:"title"`
			At           time.Time      `db:"at"`
			Duration     sql.NullString `db:"duration"`
			Description  sql.NullString `db:"description"`
			UserID       sql.NullString `db:"user_id"`
			NotifyBefore sql.NullString `db:"notify_before"`
		}

		if err := rows.StructScan(&e); err != nil {
			return nil, err
		}

		ev := storage.Event{
			ID:          e.ID,
			Title:       e.Title,
			At:          e.At,
			Description: nullStringToString(e.Description),
			UserID:      nullStringToString(e.UserID),
		}

		if e.Duration.Valid {
			if d, err := time.ParseDuration(sqlIntervalToDurationString(e.Duration.String)); err == nil {
				ev.Duration = d
			}
		}
		if e.NotifyBefore.Valid {
			if d, err := time.ParseDuration(sqlIntervalToDurationString(e.NotifyBefore.String)); err == nil {
				ev.NotifyBefore = d
			}
		}

		out = append(out, ev)
	}

	return out, nil
}

func (s *Storage) ListEvents(ctx context.Context) ([]storage.Event, error) {
	rows, err := s.db.QueryxContext(ctx, `
		SELECT id, title, at, duration::text as duration, description, user_id, notify_before::text as notify_before
		FROM events 
		ORDER BY at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return s.rowsToEvents(rows)
}

func (s *Storage) ListEventsDay(ctx context.Context, dayStart time.Time) ([]storage.Event, error) {
	rows, err := s.db.QueryxContext(ctx, `
		SELECT id, title, at, duration::text as duration, description, user_id, notify_before::text as notify_before
		FROM events 
		WHERE at >= $1 AND at < $2
		ORDER BY at`, dayStart, dayStart.Add(24*time.Hour))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return s.rowsToEvents(rows)
}

func (s *Storage) ListEventsWeek(ctx context.Context, weekStart time.Time) ([]storage.Event, error) {
	rows, err := s.db.QueryxContext(ctx, `
		SELECT id, title, at, duration::text as duration, description, user_id, notify_before::text as notify_before
		FROM events 
		WHERE at >= $1 AND at < $2
		ORDER BY at`, weekStart, weekStart.Add(7*24*time.Hour))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return s.rowsToEvents(rows)
}

func (s *Storage) ListEventsMonth(ctx context.Context, monthStart time.Time) ([]storage.Event, error) {
	end := time.Date(monthStart.Year(), monthStart.Month(), 1, 0, 0, 0, 0, monthStart.Location()).AddDate(0, 1, 0)
	rows, err := s.db.QueryxContext(ctx, `
		SELECT id, title, at, duration::text as duration, description, user_id, notify_before::text as notify_before
		FROM events
		WHERE at >= $1 AND at < $2
		ORDER BY at`, monthStart, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return s.rowsToEvents(rows)
}
