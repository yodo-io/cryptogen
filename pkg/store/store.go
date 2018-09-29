package store

import "errors"

// ErrNotFound is returned by storage providers if requested item wasn't found
var ErrNotFound = errors.New("ErrNotFound")

// Provider interface for storage handlers
type Provider interface {
	SetStatus(jobID, status string) error
	GetStatus(jobID string) (string, error)
}
