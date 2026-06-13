// Package storage provides file-storage abstractions for generated reports.
package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileStorage defines the interface for file storage operations
type FileStorage interface {
	Save(ctx context.Context, path string, content []byte) error
	Delete(ctx context.Context, path string) error
}

// LocalFileStorage implements FileStorage for the local filesystem
type LocalFileStorage struct {
	rootDir string
}

// NewLocalFileStorage creates a new LocalFileStorage
func NewLocalFileStorage(rootDir string) (*LocalFileStorage, error) {
	// Ensure root directory exists
	if err := os.MkdirAll(rootDir, 0o750); err != nil {
		return nil, fmt.Errorf("failed to create root directory: %w", err)
	}

	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path of root directory: %w", err)
	}

	return &LocalFileStorage{rootDir: absRoot}, nil
}

// Save saves content to a file at the specified path
func (s *LocalFileStorage) Save(ctx context.Context, path string, content []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	fullPath, err := s.validatePath(path)
	if err != nil {
		return err
	}

	dir := filepath.Dir(fullPath)

	// Ensure directory exists
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Check context again before the potentially expensive write operation
	if err := ctx.Err(); err != nil {
		return err
	}

	// Write file
	if err := os.WriteFile(fullPath, content, 0o600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Delete deletes a file at the specified path
func (s *LocalFileStorage) Delete(ctx context.Context, path string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	fullPath, err := s.validatePath(path)
	if err != nil {
		return err
	}

	// Delete file
	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil // Treat as success if file doesn't exist
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// validatePath ensures the path is within the root directory and returns the full absolute path
func (s *LocalFileStorage) validatePath(path string) (string, error) {
	// Clean path to prevent directory traversal
	cleanPath := filepath.Clean(path)

	// Join with root directory
	fullPath := filepath.Join(s.rootDir, cleanPath)

	// Check if the resulting path is within the root directory
	if !strings.HasPrefix(fullPath, s.rootDir) {
		return "", fmt.Errorf("path traversal attempt: path %s is outside root %s", fullPath, s.rootDir)
	}

	// Also ensure that we don't accidentally match "/root_suffix" as a prefix of "/root"
	// by checking for separator or exact match.
	if len(fullPath) > len(s.rootDir) && fullPath[len(s.rootDir)] != filepath.Separator {
		return "", fmt.Errorf("path traversal attempt: path %s is outside root %s", fullPath, s.rootDir)
	}

	return fullPath, nil
}
