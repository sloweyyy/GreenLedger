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
	if err := os.MkdirAll(rootDir, 0755); err != nil {
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
	fullPath, err := s.validatePath(path)
	if err != nil {
		return err
	}

	dir := filepath.Dir(fullPath)

	// Ensure directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(fullPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Delete deletes a file at the specified path
func (s *LocalFileStorage) Delete(ctx context.Context, path string) error {
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

	// Since s.rootDir is absolute (ensured in constructor), fullPath is absolute.
	// However, to be extra safe and consistent with checks, we can resolve it again or just check prefixes.
	// filepath.Join handles ".." elements by removing them if they are within the path,
	// but if cleanPath starts with "../", Join might still result in something outside if we aren't careful?
	// Actually filepath.Join("/root", "../foo") -> "/foo" which is outside "/root".
	// So we must check if the resulting path starts with s.rootDir.

	// We need to handle potential symlinks if we want to be 100% strict, but usually string prefix check is enough for basic protection
	// assuming we trust the rootDir creation.

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
