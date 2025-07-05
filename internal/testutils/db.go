package testutils

import (
	"database/sql"
	"kalycs/db"
	"os"
	"runtime"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// PrepareTestEnv sets up a temporary environment for testing
func PrepareTestEnv(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	// Store original values
	originalAppData := os.Getenv("APPDATA")
	originalHome := os.Getenv("HOME")

	switch runtime.GOOS {
	case "windows":
		os.Setenv("APPDATA", tmpDir)
	}
	// For both darwin and other linux/unix systems
	os.Setenv("HOME", tmpDir)

	t.Cleanup(func() {
		os.Setenv("APPDATA", originalAppData)
		os.Setenv("HOME", originalHome)
	})

	return tmpDir
}

// SetupTestDB initializes a test database
func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	if err := db.InitializeDatabase(); err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	testDB := db.GetDB()
	if testDB == nil {
		t.Fatalf("Failed to get database connection")
	}

	t.Cleanup(func() {
		if err := db.CloseDatabase(); err != nil {
			t.Errorf("Failed to close test database: %v", err)
		}
	})

	return testDB
}
