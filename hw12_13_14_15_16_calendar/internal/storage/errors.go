package storage

import "errors"

var (
	ErrNotFound = errors.New("event not found")
	ErrDateBusy = errors.New("date busy")
)
