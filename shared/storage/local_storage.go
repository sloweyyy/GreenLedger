package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LocalFileStorage implements Storage interface for local filesystem
type LocalFileStorage struct {
	basePath string
}

// NewLocalFileStorage creates a new local file storage
func NewLocalFileStorage(basePath string) (*LocalFileStorage, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &LocalFileStorage{
		basePath: basePath,
	}, nil
}

// Save saves content to local filesystem
func (s *LocalFileStorage) Save(ctx context.Context, path string, content []byte) error {
	fullPath := filepath.Join(s.basePath, path)
	if !strings.HasPrefix(filepath.Clean(fullPath), filepath.Clean(s.basePath)) {
		return fmt.Errorf("invalid file path: %s", path)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(fullPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Delete deletes file from local filesystem
func (s *LocalFileStorage) Delete(ctx context.Context, path string) error {
	fullPath := filepath.Join(s.basePath, path)
	if !strings.HasPrefix(filepath.Clean(fullPath), filepath.Clean(s.basePath)) {
		return fmt.Errorf("invalid file path: %s", path)
	}

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		// If file doesn't exist, consider it deleted
		return nil
	}

	// Delete file
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}
