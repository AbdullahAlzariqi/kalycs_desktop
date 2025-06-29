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

	for _, table := range []string{"projects", "rules"} {
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

	// After closing, operations on the db should fail.
	if err := GetDB().Ping(); err == nil {
		t.Fatalf("expected error when pinging closed database, got nil")
	}
}
