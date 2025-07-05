package db

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// prepareTestEnv sets up a temporary HOME/APPDATA directory so that database files
// are created in an isolated location during tests. It returns the temp directory
// so it can be inspected by callers if needed.
func prepareTestEnv(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	switch runtime.GOOS {
	case "windows":
		// On Windows the implementation checks APPDATA first and falls back to HOME.
		os.Setenv("APPDATA", tmpDir)
		os.Setenv("HOME", tmpDir)
	case "darwin":
		os.Setenv("HOME", tmpDir)
	default:
		// For Linux and other unix-like OSes the implementation relies on HOME.
		os.Setenv("HOME", tmpDir)
	}

	return tmpDir
}

// TestInitializeDatabase ensures that the database is created and the core tables exist.
func TestInitializeDatabase(t *testing.T) {
	tmpDir := prepareTestEnv(t)

	// Initialize the database and ensure no error is returned.
	if err := InitializeDatabase(); err != nil {
		t.Fatalf("InitializeDatabase() error = %v", err)
	}
	defer CloseDatabase()

	// Verify that the database file exists in the expected location.
	var expectedDBPath string
	switch runtime.GOOS {
	case "windows":
		expectedDBPath = filepath.Join(tmpDir, "Kalycs", "kalycs.db")
	case "darwin":
		expectedDBPath = filepath.Join(tmpDir, "Library", "Application Support", "Kalycs", "kalycs.db")
	default:
		expectedDBPath = filepath.Join(tmpDir, ".kalycs", "Kalycs", "kalycs.db")
	}

	if _, err := os.Stat(expectedDBPath); err != nil {
		t.Fatalf("expected database file %s to exist: %v", expectedDBPath, err)
	}

	// Verify that required tables exist.
	dbConn := GetDB()
	if dbConn == nil {
		t.Fatalf("GetDB() returned nil")
	}

	for _, table := range []string{"projects", "rules", "files"} {
		var name string
		err := dbConn.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Fatalf("table %s does not exist or query failed: %v", table, err)
		}
	}
}

// TestInitializeDatabaseIdempotent verifies that calling InitializeDatabase twice does not return an error.
func TestInitializeDatabaseIdempotent(t *testing.T) {
	prepareTestEnv(t)

	if err := InitializeDatabase(); err != nil {
		t.Fatalf("first InitializeDatabase() error = %v", err)
	}

	// Second call should succeed without error.
	if err := InitializeDatabase(); err != nil {
		t.Fatalf("second InitializeDatabase() error = %v", err)
	}

	CloseDatabase()
}

// TestCloseDatabase ensures that the database connection is properly closed.
func TestCloseDatabase(t *testing.T) {
	prepareTestEnv(t)

	if err := InitializeDatabase(); err != nil {
		t.Fatalf("InitializeDatabase() error = %v", err)
	}

	if err := CloseDatabase(); err != nil {
		t.Fatalf("CloseDatabase() error = %v", err)
	}

	// After closing, GetDB should return nil
	if GetDB() != nil {
		t.Fatalf("expected GetDB() to return nil after closing database, got non-nil")
	}
}

// TestGetAppDataDirectory checks that the app data directory is correctly determined.
func TestGetAppDataDirectory(t *testing.T) {
	// Save original env vars to restore them after the test.
	originalHome := os.Getenv("HOME")
	originalAppData := os.Getenv("APPDATA")
	originalUserProfile := os.Getenv("USERPROFILE")
	defer func() {
		os.Setenv("HOME", originalHome)
		os.Setenv("APPDATA", originalAppData)
		os.Setenv("USERPROFILE", originalUserProfile)
	}()

	t.Run("it returns the correct directory for linux", func(t *testing.T) {
		if runtime.GOOS != "linux" {
			t.Skip("test is for linux only")
		}
		tmpDir := t.TempDir()
		os.Setenv("HOME", tmpDir)

		appDir, err := getAppDataDirectory()
		if err != nil {
			t.Fatalf("getAppDataDirectory() error = %v", err)
		}

		expectedDir := filepath.Join(tmpDir, ".kalycs", "Kalycs")
		if appDir != expectedDir {
			t.Fatalf("expected app dir %s, got %s", expectedDir, appDir)
		}
	})

	t.Run("it returns the correct directory for macos", func(t *testing.T) {
		if runtime.GOOS != "darwin" {
			t.Skip("test is for darwin only")
		}
		tmpDir := t.TempDir()
		os.Setenv("HOME", tmpDir)

		appDir, err := getAppDataDirectory()
		if err != nil {
			t.Fatalf("getAppDataDirectory() error = %v", err)
		}

		expectedDir := filepath.Join(tmpDir, "Library", "Application Support", "Kalycs")
		if appDir != expectedDir {
			t.Fatalf("expected app dir %s, got %s", expectedDir, appDir)
		}
	})

	t.Run("it returns an error if home directory is not found", func(t *testing.T) {
		os.Unsetenv("HOME")
		os.Unsetenv("APPDATA")
		os.Unsetenv("USERPROFILE") // For Windows fallback

		_, err := getAppDataDirectory()
		if err == nil {
			t.Fatalf("expected error when HOME and APPDATA are not set, got nil")
		}
	})
}

// TestDatabaseState checks the state of the database connection before and after initialization.
func TestDatabaseState(t *testing.T) {
	// Ensure any existing database connection is closed first to avoid global state leakage
	CloseDatabase()

	// 1. Ensure GetDB is nil before initialization
	if GetDB() != nil {
		t.Fatalf("GetDB() should be nil before initialization")
	}

	// 2. Ensure closing a nil DB doesn't cause an error
	if err := CloseDatabase(); err != nil {
		t.Fatalf("CloseDatabase() on a nil db should not return an error, got %v", err)
	}

	// 3. Initialize DB and check if it's not nil
	prepareTestEnv(t)
	if err := InitializeDatabase(); err != nil {
		t.Fatalf("InitializeDatabase() error = %v", err)
	}
	defer CloseDatabase()

	if GetDB() == nil {
		t.Fatalf("GetDB() should not be nil after initialization")
	}
}
