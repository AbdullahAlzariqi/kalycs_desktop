package watcher_test

import (
	"context"
	"kalycs/internal/watcher"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewWatcher_Success(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "watcher-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	w, err := watcher.NewWatcher(context.Background(), tempDir)
	if err != nil {
		t.Fatalf("Expected no error from NewWatcher, got %v", err)
	}
	if w == nil {
		t.Fatal("Expected watcher to be non-nil")
	}

	w.Stop() // Clean up the watcher's goroutine
}

func TestNewWatcher_Error(t *testing.T) {
	nonExistentPath := filepath.Join(os.TempDir(), "non-existent-dir-for-kalycs-test")
	_, err := watcher.NewWatcher(context.Background(), nonExistentPath)
	if err == nil {
		t.Fatal("Expected an error from NewWatcher for non-existent path, got nil")
	}
}

func TestWatcher_StartStop(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "watcher-test-start-stop")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	w, err := watcher.NewWatcher(context.Background(), tempDir)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	w.Start()

	// Give the goroutine a moment to start up
	time.Sleep(10 * time.Millisecond)

	w.Stop()

	// Give the goroutine a moment to shut down
	time.Sleep(10 * time.Millisecond)
}
