package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalFileStorage_SaveAndDelete(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	storage, err := NewLocalFileStorage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	ctx := context.Background()
	path := "test/file.txt"
	content := []byte("hello world")

	// Test Save
	if err := storage.Save(ctx, path, content); err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}

	// Verify file exists
	fullPath := filepath.Join(tmpDir, path)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Errorf("File was not created at %s", fullPath)
	}

	// Verify content
	readContent, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(readContent) != string(content) {
		t.Errorf("Expected content %s, got %s", content, readContent)
	}

	// Test Delete
	if err := storage.Delete(ctx, path); err != nil {
		t.Fatalf("Failed to delete file: %v", err)
	}

	// Verify file is gone
	if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
		t.Errorf("File still exists at %s", fullPath)
	}
}

func TestLocalFileStorage_PathTraversal(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	storage, err := NewLocalFileStorage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	ctx := context.Background()

	// Case 1: Attempt to write outside using ".."
	// Trying to write to a file that would be a sibling of the root directory
	err = storage.Save(ctx, "../outside.txt", []byte("malicious content"))
	if err == nil {
		t.Error("Expected error when saving with path traversal '..', got nil")
	} else if err.Error() == "" {
		// Verify error message contains expectation
		t.Error("Expected error message, got empty string")
	}

	// Case 2: Verify ".." staying inside is fine (if we supported it, but our clean logic removes it)
	// path/../path/file.txt -> path/file.txt which is inside.
	err = storage.Save(ctx, "subdir/../allowed.txt", []byte("ok"))
	if err != nil {
		t.Errorf("Expected success for 'subdir/../allowed.txt', got error: %v", err)
	}

	// Verify allowed.txt exists
	if _, err := os.Stat(filepath.Join(tmpDir, "allowed.txt")); os.IsNotExist(err) {
		t.Error("File allowed.txt was not created")
	}
}

func TestLocalFileStorage_ContextCancellation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	storage, err := NewLocalFileStorage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = storage.Save(ctx, "test.txt", []byte("content"))
	if err == nil {
		t.Error("Expected error due to canceled context, got nil")
	}
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}

	err = storage.Delete(ctx, "test.txt")
	if err == nil {
		t.Error("Expected error due to canceled context, got nil")
	}
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
}

func TestLocalFileStorage_DeleteNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	storage, err := NewLocalFileStorage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	ctx := context.Background()
	// Should not return error
	if err := storage.Delete(ctx, "non_existent.txt"); err != nil {
		t.Errorf("Expected no error when deleting non-existent file, got %v", err)
	}
}

func TestLocalFileStorage_EmptyPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	storage, err := NewLocalFileStorage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	ctx := context.Background()
	// Empty path resolves to ".", which is the root directory
	// Saving to directory path usually fails with EISDIR (Is a directory)
	err = storage.Save(ctx, "", []byte("content"))
	if err == nil {
		t.Error("Expected error when saving to empty path (root dir), got nil")
	}
}

func TestLocalFileStorage_RootPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	storage, err := NewLocalFileStorage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	ctx := context.Background()
	// "." resolves to root directory.
	// This tests the edge case in validation logic where fullPath == rootDir
	// Save should fail because it's a directory, but validation should pass (or fail if we consider root untoucheable, but implementation allows "inside").
	// Implementation: fullPath == rootDir. Prefix matches. Len equal. Returns fullPath.
	// Then os.WriteFile(fullPath) attempts to write to directory.

	err = storage.Save(ctx, ".", []byte("content"))
	if err == nil {
		t.Error("Expected error when saving to root directory, got nil")
	}

	// Ensure it didn't delete the root dir or something weird
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Error("Root directory disappeared")
	}
}
