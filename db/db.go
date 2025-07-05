package db

import (
	"database/sql"
	"fmt"
	"kalycs/internal/logging"
	"os"
	"path/filepath"
	"runtime"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Database connection instance
var db *sql.DB

// Project represents the project schema
type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	IsFavourite bool      `json:"is_favourite"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Rule represents the rules schema
type Rule struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	ProjectID     string    `json:"project_id"`
	Rule          string    `json:"rule"`  // starts_with, contains, ends_with, extension, regex
	Texts         string    `json:"texts"` // JSON array as string
	CaseSensitive bool      `json:"case_sensitive"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// File represents the file schema
type File struct {
	ID        string         `json:"id"`
	Path      string         `json:"path"`
	Name      string         `json:"name"`
	Ext       string         `json:"ext"`
	Size      int64          `json:"size"`
	Mtime     time.Time      `json:"mtime"`
	ProjectID sql.NullString `json:"project_id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// getAppDataDirectory returns the appropriate application data directory for the current OS
func getAppDataDirectory() (string, error) {
	var baseDir string

	switch runtime.GOOS {
	case "windows":
		baseDir = os.Getenv("APPDATA")
		if baseDir == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("failed to get user home directory: %w", err)
			}
			baseDir = filepath.Join(homeDir, "AppData", "Roaming")
		}
	case "darwin": // macOS
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		baseDir = filepath.Join(homeDir, "Library", "Application Support")
	default:
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		baseDir = filepath.Join(homeDir, ".kalycs")
	}

	appDir := filepath.Join(baseDir, "Kalycs")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(appDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create app directory: %w", err)
	}

	return appDir, nil
}

// InitializeDatabase sets up the SQLite database and creates tables
func InitializeDatabase() error {
	// Close existing connection if it exists to prevent connection leaks
	if db != nil {
		if err := db.Close(); err != nil {
			logging.L().Warnw("Failed to close existing database connection", "error", err)
		}
		db = nil // Clear the reference
	}

	appDir, err := getAppDataDirectory()
	if err != nil {
		return fmt.Errorf("failed to get app directory: %w", err)
	}

	dbPath := filepath.Join(appDir, "kalycs.db")

	// Open database connection
	db, err = sql.Open("sqlite3", dbPath+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Set secure file permissions
	if err := os.Chmod(dbPath, 0600); err != nil {
		logging.L().Warnw("Failed to set secure permissions on database file", "error", err, "path", dbPath)
	}

	// Create tables
	if err := createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	logging.L().Info("Database initialized successfully")
	return nil
}

// createTables creates the required database tables
func createTables() error {
	projectTable := `
	CREATE TABLE IF NOT EXISTS projects (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL UNIQUE CHECK(length(name) <= 25),
		description TEXT CHECK(length(description) <= 200),
		is_active BOOLEAN NOT NULL DEFAULT 1,
		is_favourite BOOLEAN NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`

	ruleTable := `
	CREATE TABLE IF NOT EXISTS rules (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL CHECK(length(name) <= 25),
		project_id TEXT NOT NULL,
		rule TEXT NOT NULL CHECK(rule IN ('starts_with', 'contains', 'ends_with', 'extension', 'regex')),
		texts TEXT NOT NULL,
		case_sensitive BOOLEAN NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
	);`

	fileTable := `
	CREATE TABLE IF NOT EXISTS files (
		id          TEXT PRIMARY KEY,
		path        TEXT UNIQUE,
		name        TEXT NOT NULL,
		ext         TEXT NOT NULL,
		size        INTEGER,
		mtime       DATETIME,
		project_id  TEXT,
		created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL
	);`

	// Create indexes
	projectNameIndex := `CREATE INDEX IF NOT EXISTS idx_projects_name ON projects(name);`
	ruleProjectIndex := `CREATE INDEX IF NOT EXISTS idx_rules_project_id ON rules(project_id);`

	// Create trigger for updated_at
	projectTrigger := `
	CREATE TRIGGER IF NOT EXISTS update_projects_updated_at 
	AFTER UPDATE ON projects
	BEGIN
		UPDATE projects SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
	END;`

	ruleTrigger := `
	CREATE TRIGGER IF NOT EXISTS update_rules_updated_at 
	AFTER UPDATE ON rules
	BEGIN
		UPDATE rules SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
	END;`

	fileTrigger := `
	CREATE TRIGGER IF NOT EXISTS trg_files_updated_at
	AFTER UPDATE ON files
	BEGIN
		UPDATE files SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
	END;`

	statements := []string{
		projectTable, ruleTable, fileTable, projectNameIndex, ruleProjectIndex, projectTrigger, ruleTrigger, fileTrigger,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute statement: %w", err)
		}
	}

	return nil
}

// CloseDatabase closes the database connection
func CloseDatabase() error {
	if db != nil {
		err := db.Close()
		db = nil // Clear the reference after closing
		return err
	}
	return nil
}

// GetDB returns the database instance
func GetDB() *sql.DB {
	return db
}
