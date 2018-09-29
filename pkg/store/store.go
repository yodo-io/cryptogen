package store

import "errors"

// ErrNotFound is returned by storage providers if requested item wasn't found
var ErrNotFound = errors.New("ErrNotFound")

type JobStatus struct {
	Status      string
	SecretPaths []string
}

// Provider interface for storage handlers
type Provider interface {
	SetStatus(jobID string, status JobStatus) error
	GetStatus(jobID string) (JobStatus, error)
}
