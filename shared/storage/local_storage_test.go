package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalFileStorage(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "storage-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storage, err := NewLocalFileStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	ctx := context.Background()
	testPath := "users/user1/report.pdf"
	content := []byte("test content")

	// Test Save
	err = storage.Save(ctx, testPath, content)
	if err != nil {
		t.Errorf("Save failed: %v", err)
	}

	// Verify file exists and content matches
	fullPath := filepath.Join(tempDir, testPath)
	readContent, err := os.ReadFile(fullPath)
	if err != nil {
		t.Errorf("Failed to read saved file: %v", err)
	}

	if string(readContent) != string(content) {
		t.Errorf("Content mismatch. Expected %s, got %s", string(content), string(readContent))
	}

	// Test Delete
	err = storage.Delete(ctx, testPath)
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	// Verify file does not exist
	if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
		t.Errorf("File should have been deleted")
	}

	// Test Delete non-existent file (should not error)
	err = storage.Delete(ctx, "non-existent.pdf")
	if err != nil {
		t.Errorf("Delete non-existent file failed: %v", err)
	}
}
