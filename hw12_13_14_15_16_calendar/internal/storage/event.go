package storage

import "time"

type Event struct {
	ID           string
	Title        string
	At           time.Time
	Duration     time.Duration
	Description  string
	UserID       string
	NotifyBefore time.Duration
}
