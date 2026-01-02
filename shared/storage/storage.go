package storage

import (
	"context"
)

// Storage interface defines methods for file storage operations
type Storage interface {
	// Save saves content to storage at the given path
	Save(ctx context.Context, path string, content []byte) error

	// Delete deletes the file at the given path
	Delete(ctx context.Context, path string) error
}
