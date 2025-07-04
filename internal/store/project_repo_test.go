package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"kalycs/db"

	_ "github.com/mattn/go-sqlite3"
)

// prepareTestEnv sets up a temporary environment for testing
func prepareTestEnv(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	switch runtime.GOOS {
	case "windows":
		os.Setenv("APPDATA", tmpDir)
		os.Setenv("HOME", tmpDir)
	case "darwin":
		os.Setenv("HOME", tmpDir)
	default:
		os.Setenv("HOME", tmpDir)
	}

	return tmpDir
}

// setupTestDB initializes a test database
func setupTestDB(t *testing.T) *sql.DB {
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

// createTestProject creates a valid test project
func createTestProject(name string) *db.Project {
	return &db.Project{
		Name:        name,
		Description: "Test project description",
		IsActive:    true,
		IsFavourite: false,
	}
}

// createTestProjectWithID creates a test project with a specific ID
func createTestProjectWithID(id, name string) *db.Project {
	return &db.Project{
		ID:          id,
		Name:        name,
		Description: "Test project description",
		IsActive:    true,
		IsFavourite: false,
	}
}

func TestProjectRepo_Create(t *testing.T) {
	prepareTestEnv(t)
	testDB := setupTestDB(t)
	repo := NewProjectRepo(testDB)

	tests := []struct {
		name    string
		project *db.Project
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid project",
			project: createTestProject("Test Project"),
			wantErr: false,
		},
		{
			name:    "nil project",
			project: nil,
			wantErr: true,
			errMsg:  "project cannot be nil",
		},
		{
			name:    "empty name",
			project: createTestProject(""),
			wantErr: true,
			errMsg:  "validation failed",
		},
		{
			name:    "name too long",
			project: createTestProject("This project name is way too long and exceeds the maximum allowed length"),
			wantErr: true,
			errMsg:  "validation failed",
		},
		{
			name: "description too long",
			project: &db.Project{
				Name:        "Valid Name",
				Description: "This description is extremely long and exceeds the maximum allowed length for project descriptions. It should cause a validation error because it contains way more than the allowed 200 characters limit for descriptions in the database schema and validation rules.",
				IsActive:    true,
			},
			wantErr: true,
			errMsg:  "validation failed",
		},
		{
			name:    "invalid UUID",
			project: createTestProjectWithID("invalid-uuid", "Valid Project"),
			wantErr: true,
			errMsg:  "validation failed",
		},
		{
			name:    "valid UUID",
			project: createTestProjectWithID("550e8400-e29b-41d4-a716-446655440000", "Valid Project"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := repo.Create(ctx, tt.project)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Create() expected error, got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Create() error = %v, expected to contain %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Create() unexpected error = %v", err)
					return
				}

				if tt.project.ID == "" {
					t.Error("Create() should have generated an ID")
				}
				if tt.project.CreatedAt.IsZero() {
					t.Error("Create() should have set CreatedAt")
				}
				if tt.project.UpdatedAt.IsZero() {
					t.Error("Create() should have set UpdatedAt")
				}
			}
		})
	}
}

func TestProjectRepo_Create_DuplicateName(t *testing.T) {
	prepareTestEnv(t)
	testDB := setupTestDB(t)
	repo := NewProjectRepo(testDB)
	ctx := context.Background()

	// Create first project
	project1 := createTestProject("Duplicate Name")
	err := repo.Create(ctx, project1)
	if err != nil {
		t.Fatalf("Failed to create first project: %v", err)
	}

	// Try to create second project with same name
	project2 := createTestProject("Duplicate Name")
	err = repo.Create(ctx, project2)
	if err == nil {
		t.Error("Create() expected error for duplicate name, got nil")
	} else if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("Create() error = %v, expected to contain 'already exists'", err)
	}
}

func TestProjectRepo_GetByID(t *testing.T) {
	prepareTestEnv(t)
	testDB := setupTestDB(t)
	repo := NewProjectRepo(testDB)
	ctx := context.Background()

	// Create a test project
	project := createTestProject("Test Project for GetByID")
	err := repo.Create(ctx, project)
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	tests := []struct {
		name    string
		id      string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid ID",
			id:      project.ID,
			wantErr: false,
		},
		{
			name:    "empty ID",
			id:      "",
			wantErr: true,
			errMsg:  "project ID cannot be empty",
		},
		{
			name:    "invalid UUID format",
			id:      "invalid-uuid",
			wantErr: true,
			errMsg:  "invalid project ID format",
		},
		{
			name:    "non-existent ID",
			id:      "550e8400-e29b-41d4-a716-446655440000",
			wantErr: true,
			errMsg:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByID(tt.id)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetByID() expected error, got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("GetByID() error = %v, expected to contain %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("GetByID() unexpected error = %v", err)
					return
				}
				if result == nil {
					t.Error("GetByID() returned nil project")
					return
				}

				// Verify the returned project matches what we created
				if result.ID != project.ID {
					t.Errorf("GetByID() ID = %v, want %v", result.ID, project.ID)
				}
				if result.Name != project.Name {
					t.Errorf("GetByID() Name = %v, want %v", result.Name, project.Name)
				}
				if result.Description != project.Description {
					t.Errorf("GetByID() Description = %v, want %v", result.Description, project.Description)
				}
				if result.IsActive != project.IsActive {
					t.Errorf("GetByID() IsActive = %v, want %v", result.IsActive, project.IsActive)
				}
				if result.IsFavourite != project.IsFavourite {
					t.Errorf("GetByID() IsFavourite = %v, want %v", result.IsFavourite, project.IsFavourite)
				}
			}
		})
	}
}

func TestProjectRepo_GetAll(t *testing.T) {
	prepareTestEnv(t)
	testDB := setupTestDB(t)
	repo := NewProjectRepo(testDB)
	ctx := context.Background()

	// Test with empty database
	t.Run("empty database", func(t *testing.T) {
		projects, err := repo.GetAll()
		if err != nil {
			t.Errorf("GetAll() unexpected error = %v", err)
		}
		if len(projects) != 0 {
			t.Errorf("GetAll() returned %d projects, want 0", len(projects))
		}
	})

	// Create test projects
	project1 := createTestProject("First Project")
	project2 := createTestProject("Second Project")
	project3 := createTestProject("Third Project")

	err := repo.Create(ctx, project1)
	if err != nil {
		t.Fatalf("Failed to create first project: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	err = repo.Create(ctx, project2)
	if err != nil {
		t.Fatalf("Failed to create second project: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	err = repo.Create(ctx, project3)
	if err != nil {
		t.Fatalf("Failed to create third project: %v", err)
	}

	// Test with populated database
	t.Run("populated database", func(t *testing.T) {
		projects, err := repo.GetAll()
		if err != nil {
			t.Errorf("GetAll() unexpected error = %v", err)
			return
		}

		if len(projects) != 3 {
			t.Errorf("GetAll() returned %d projects, want 3", len(projects))
			return
		}

		// Verify projects are ordered by created_at DESC (newest first)
		if projects[0].Name != "Third Project" {
			t.Errorf("GetAll() first project name = %v, want 'Third Project'", projects[0].Name)
		}
		if projects[1].Name != "Second Project" {
			t.Errorf("GetAll() second project name = %v, want 'Second Project'", projects[1].Name)
		}
		if projects[2].Name != "First Project" {
			t.Errorf("GetAll() third project name = %v, want 'First Project'", projects[2].Name)
		}

		// Verify all fields are populated correctly
		for i, project := range projects {
			if project.ID == "" {
				t.Errorf("GetAll() project[%d] missing ID", i)
			}
			if project.Name == "" {
				t.Errorf("GetAll() project[%d] missing Name", i)
			}
			if project.CreatedAt.IsZero() {
				t.Errorf("GetAll() project[%d] missing CreatedAt", i)
			}
			if project.UpdatedAt.IsZero() {
				t.Errorf("GetAll() project[%d] missing UpdatedAt", i)
			}
		}
	})
}

func TestProjectRepo_Update(t *testing.T) {
	prepareTestEnv(t)
	testDB := setupTestDB(t)
	repo := NewProjectRepo(testDB)
	ctx := context.Background()

	// Create a test project
	project := createTestProject("Original Project")
	err := repo.Create(ctx, project)
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	originalUpdatedAt := project.UpdatedAt

	tests := []struct {
		name    string
		project *db.Project
		setup   func() *db.Project
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid update",
			setup: func() *db.Project {
				updated := *project
				updated.Name = "Updated Project"
				updated.Description = "Updated description"
				updated.IsActive = false
				updated.IsFavourite = true
				return &updated
			},
			wantErr: false,
		},
		{
			name:    "nil project",
			project: nil,
			wantErr: true,
			errMsg:  "project cannot be nil",
		},
		{
			name: "empty ID",
			setup: func() *db.Project {
				updated := *project
				updated.ID = ""
				return &updated
			},
			wantErr: true,
			errMsg:  "project ID cannot be empty for update",
		},
		{
			name: "invalid name",
			setup: func() *db.Project {
				updated := *project
				updated.Name = ""
				return &updated
			},
			wantErr: true,
			errMsg:  "validation failed",
		},
		{
			name: "non-existent project",
			setup: func() *db.Project {
				return &db.Project{
					ID:          "550e8400-e29b-41d4-a716-446655440000",
					Name:        "Non-existent Project",
					Description: "This project doesn't exist",
					IsActive:    true,
				}
			},
			wantErr: true,
			errMsg:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var updateProject *db.Project
			if tt.setup != nil {
				updateProject = tt.setup()
			} else {
				updateProject = tt.project
			}

			err := repo.Update(updateProject)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Update() expected error, got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Update() error = %v, expected to contain %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Update() unexpected error = %v", err)
					return
				}

				// Verify the project was updated
				if updateProject.UpdatedAt.Equal(originalUpdatedAt) || updateProject.UpdatedAt.Before(originalUpdatedAt) {
					t.Error("Update() should have updated the UpdatedAt timestamp")
				}

				// Verify the changes in the database
				updatedProject, err := repo.GetByID(updateProject.ID)
				if err != nil {
					t.Errorf("Failed to get updated project: %v", err)
					return
				}

				if updatedProject.Name != updateProject.Name {
					t.Errorf("Update() Name = %v, want %v", updatedProject.Name, updateProject.Name)
				}
				if updatedProject.Description != updateProject.Description {
					t.Errorf("Update() Description = %v, want %v", updatedProject.Description, updateProject.Description)
				}
				if updatedProject.IsActive != updateProject.IsActive {
					t.Errorf("Update() IsActive = %v, want %v", updatedProject.IsActive, updateProject.IsActive)
				}
				if updatedProject.IsFavourite != updateProject.IsFavourite {
					t.Errorf("Update() IsFavourite = %v, want %v", updatedProject.IsFavourite, updateProject.IsFavourite)
				}
			}
		})
	}
}

func TestProjectRepo_Update_DuplicateName(t *testing.T) {
	prepareTestEnv(t)
	testDB := setupTestDB(t)
	repo := NewProjectRepo(testDB)
	ctx := context.Background()

	// Create two projects
	project1 := createTestProject("Project One")
	project2 := createTestProject("Project Two")

	err := repo.Create(ctx, project1)
	if err != nil {
		t.Fatalf("Failed to create first project: %v", err)
	}

	err = repo.Create(ctx, project2)
	if err != nil {
		t.Fatalf("Failed to create second project: %v", err)
	}

	// Try to update project2 to have the same name as project1
	project2.Name = "Project One"
	err = repo.Update(project2)
	if err == nil {
		t.Error("Update() expected error for duplicate name, got nil")
	} else if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("Update() error = %v, expected to contain 'already exists'", err)
	}
}

func TestProjectRepo_Delete(t *testing.T) {
	prepareTestEnv(t)
	testDB := setupTestDB(t)
	repo := NewProjectRepo(testDB)
	ctx := context.Background()

	tests := []struct {
		name    string
		id      string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty ID",
			id:      "",
			wantErr: true,
			errMsg:  "project ID cannot be empty",
		},
		{
			name:    "invalid UUID format",
			id:      "invalid-uuid",
			wantErr: true,
			errMsg:  "invalid project ID format",
		},
		{
			name:    "non-existent ID",
			id:      "550e8400-e29b-41d4-a716-446655440000",
			wantErr: true,
			errMsg:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Delete(tt.id)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Delete() expected error, got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Delete() error = %v, expected to contain %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Delete() unexpected error = %v", err)
				}
			}
		})
	}

	// Test valid deletion separately
	t.Run("valid deletion", func(t *testing.T) {
		project := createTestProject("Project to Delete")
		err := repo.Create(ctx, project)
		if err != nil {
			t.Fatalf("Failed to create test project: %v", err)
		}

		err = repo.Delete(project.ID)
		if err != nil {
			t.Errorf("Delete() unexpected error = %v", err)
			return
		}

		// Verify the project was deleted
		_, err = repo.GetByID(project.ID)
		if err == nil {
			t.Error("Delete() project still exists after deletion")
		} else if !strings.Contains(err.Error(), "not found") {
			t.Errorf("GetByID() after delete error = %v, expected 'not found'", err)
		}
	})
}

// Integration test that tests the full CRUD lifecycle
func TestProjectRepo_FullCRUDLifecycle(t *testing.T) {
	prepareTestEnv(t)
	testDB := setupTestDB(t)
	repo := NewProjectRepo(testDB)
	ctx := context.Background()

	// Create
	project := createTestProject("Lifecycle Test Project")
	err := repo.Create(ctx, project)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	originalID := project.ID
	if originalID == "" {
		t.Fatal("Create should have generated an ID")
	}

	// Read by ID
	retrievedProject, err := repo.GetByID(originalID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrievedProject.Name != project.Name {
		t.Errorf("Retrieved project name = %v, want %v", retrievedProject.Name, project.Name)
	}

	// Read all
	allProjects, err := repo.GetAll()
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}

	found := false
	for _, p := range allProjects {
		if p.ID == originalID {
			found = true
			break
		}
	}
	if !found {
		t.Error("Created project not found in GetAll results")
	}

	// Update
	retrievedProject.Name = "Updated Lifecycle Project"
	retrievedProject.IsActive = false
	retrievedProject.IsFavourite = true

	err = repo.Update(retrievedProject)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	updatedProject, err := repo.GetByID(originalID)
	if err != nil {
		t.Fatalf("GetByID after update failed: %v", err)
	}

	if updatedProject.Name != "Updated Lifecycle Project" {
		t.Errorf("Updated project name = %v, want 'Updated Lifecycle Project'", updatedProject.Name)
	}

	// Delete
	err = repo.Delete(originalID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	_, err = repo.GetByID(originalID)
	if err == nil {
		t.Error("GetByID after delete should have failed")
	} else if !strings.Contains(err.Error(), "not found") {
		t.Errorf("GetByID after delete error = %v, expected 'not found'", err)
	}
}

// Edge case tests for robustness
func TestProjectRepo_EdgeCases(t *testing.T) {
	prepareTestEnv(t)
	testDB := setupTestDB(t)
	repo := NewProjectRepo(testDB)
	ctx := context.Background()

	t.Run("create with whitespace name", func(t *testing.T) {
		project := &db.Project{
			Name:        "  Test Project  ",
			Description: "  Test description  ",
			IsActive:    true,
		}

		err := repo.Create(ctx, project)
		if err != nil {
			t.Errorf("Create() with whitespace failed: %v", err)
			return
		}

		// Verify the name was normalized (whitespace trimmed)
		retrieved, err := repo.GetByID(project.ID)
		if err != nil {
			t.Fatalf("Failed to retrieve project: %v", err)
		}

		if retrieved.Name != "Test Project" {
			t.Errorf("Expected normalized name 'Test Project', got '%s'", retrieved.Name)
		}
		if retrieved.Description != "Test description" {
			t.Errorf("Expected normalized description 'Test description', got '%s'", retrieved.Description)
		}
	})

	t.Run("create with control characters", func(t *testing.T) {
		project := &db.Project{
			Name:     "Test\tProject\n",
			IsActive: true,
		}

		err := repo.Create(ctx, project)
		if err == nil {
			t.Error("Create() should fail with control characters in name")
		} else if !strings.Contains(err.Error(), "validation failed") {
			t.Errorf("Create() error = %v, expected validation error", err)
		}
	})

	t.Run("create with exactly max length name", func(t *testing.T) {
		// Create a name with exactly 25 characters
		name := "1234567890123456789012345" // 25 chars
		project := &db.Project{
			Name:     name,
			IsActive: true,
		}

		err := repo.Create(ctx, project)
		if err != nil {
			t.Errorf("Create() with max length name failed: %v", err)
		}
	})

	t.Run("create with exactly max length description", func(t *testing.T) {
		// Create a description with exactly 200 characters
		description := strings.Repeat("a", 200)
		project := &db.Project{
			Name:        "Test Proj Max Desc",
			Description: description,
			IsActive:    true,
		}

		err := repo.Create(ctx, project)
		if err != nil {
			t.Errorf("Create() with max length description failed: %v", err)
		}
	})
}

// Concurrency tests for thread safety
func TestProjectRepo_ConcurrentOperations(t *testing.T) {
	prepareTestEnv(t)
	testDB := setupTestDB(t)
	repo := NewProjectRepo(testDB)
	ctx := context.Background()

	t.Run("concurrent creates", func(t *testing.T) {
		const numGoroutines = 10
		errors := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(index int) {
				project := createTestProject("Concurrent Project " + fmt.Sprintf("%d", index))
				errors <- repo.Create(ctx, project)
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			err := <-errors
			if err != nil {
				t.Errorf("Concurrent create %d failed: %v", i, err)
			}
		}

		// Verify all projects were created
		projects, err := repo.GetAll()
		if err != nil {
			t.Errorf("GetAll() failed: %v", err)
		}

		if len(projects) != numGoroutines {
			t.Errorf("Expected %d projects, got %d", numGoroutines, len(projects))
		}
	})

	t.Run("concurrent reads", func(t *testing.T) {
		// Create a project first
		project := createTestProject("Concurrent Read Project")
		err := repo.Create(ctx, project)
		if err != nil {
			t.Fatalf("Failed to create test project: %v", err)
		}

		const numGoroutines = 20
		errors := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				_, err := repo.GetByID(project.ID)
				errors <- err
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			err := <-errors
			if err != nil {
				t.Errorf("Concurrent read %d failed: %v", i, err)
			}
		}
	})
}

// Context cancellation tests
func TestProjectRepo_ContextCancellation(t *testing.T) {
	prepareTestEnv(t)
	testDB := setupTestDB(t)
	repo := NewProjectRepo(testDB)

	t.Run("create with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		project := createTestProject("Test Project")
		err := repo.Create(ctx, project)

		if err == nil {
			t.Error("Create() should fail with cancelled context")
		} else if !strings.Contains(err.Error(), "context canceled") {
			// Note: The error might be wrapped, so we check for context cancellation
			t.Logf("Create() with cancelled context error: %v", err)
		}
	})

	t.Run("create with timeout context", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		// Add a small delay to ensure context times out
		time.Sleep(10 * time.Millisecond)

		project := createTestProject("Test Project")
		err := repo.Create(ctx, project)

		if err == nil {
			t.Error("Create() should fail with timed out context")
		} else {
			t.Logf("Create() with timeout context error: %v", err)
		}
	})
}

// Performance and stress tests
func TestProjectRepo_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	prepareTestEnv(t)
	testDB := setupTestDB(t)
	repo := NewProjectRepo(testDB)
	ctx := context.Background()

	t.Run("bulk create performance", func(t *testing.T) {
		const numProjects = 1000
		start := time.Now()

		for i := 0; i < numProjects; i++ {
			project := createTestProject("Bulk Project " + fmt.Sprintf("%d", i))
			err := repo.Create(ctx, project)
			if err != nil {
				t.Fatalf("Bulk create failed at iteration %d: %v", i, err)
			}
		}

		duration := time.Since(start)
		t.Logf("Created %d projects in %v (%.2f projects/sec)",
			numProjects, duration, float64(numProjects)/duration.Seconds())

		// Verify all projects were created
		projects, err := repo.GetAll()
		if err != nil {
			t.Fatalf("GetAll() failed: %v", err)
		}

		if len(projects) != numProjects {
			t.Errorf("Expected %d projects, got %d", numProjects, len(projects))
		}
	})
}

// Database constraint tests
func TestProjectRepo_DatabaseConstraints(t *testing.T) {
	prepareTestEnv(t)
	testDB := setupTestDB(t)
	repo := NewProjectRepo(testDB)
	ctx := context.Background()

	t.Run("unique name constraint", func(t *testing.T) {
		project1 := createTestProject("Unique Name Test")
		err := repo.Create(ctx, project1)
		if err != nil {
			t.Fatalf("Failed to create first project: %v", err)
		}

		project2 := createTestProject("Unique Name Test")
		err = repo.Create(ctx, project2)
		if err == nil {
			t.Error("Second project with same name should fail")
		} else if !strings.Contains(err.Error(), "already exists") {
			t.Errorf("Expected unique constraint error, got: %v", err)
		}
	})

	t.Run("name length constraint", func(t *testing.T) {
		// Test name that's too long (26 characters)
		longName := "12345678901234567890123456"
		project := &db.Project{
			Name:     longName,
			IsActive: true,
		}

		err := repo.Create(ctx, project)
		if err == nil {
			t.Error("Create() should fail with name longer than 25 characters")
		} else if !strings.Contains(err.Error(), "validation failed") {
			t.Errorf("Expected validation error, got: %v", err)
		}
	})

	t.Run("description length constraint", func(t *testing.T) {
		// Test description that's too long (201 characters)
		longDescription := strings.Repeat("a", 201)
		project := &db.Project{
			Name:        "Test Project",
			Description: longDescription,
			IsActive:    true,
		}

		err := repo.Create(ctx, project)
		if err == nil {
			t.Error("Create() should fail with description longer than 200 characters")
		} else if !strings.Contains(err.Error(), "validation failed") {
			t.Errorf("Expected validation error, got: %v", err)
		}
	})
}

// Test data integrity across operations
func TestProjectRepo_DataIntegrity(t *testing.T) {
	prepareTestEnv(t)
	testDB := setupTestDB(t)
	repo := NewProjectRepo(testDB)
	ctx := context.Background()

	t.Run("timestamps consistency", func(t *testing.T) {
		beforeCreate := time.Now().UTC()

		project := createTestProject("Timestamp Test")
		err := repo.Create(ctx, project)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		afterCreate := time.Now().UTC()

		// Verify creation timestamps
		if project.CreatedAt.Before(beforeCreate) || project.CreatedAt.After(afterCreate) {
			t.Errorf("CreatedAt timestamp %v is not between %v and %v",
				project.CreatedAt, beforeCreate, afterCreate)
		}

		if project.UpdatedAt.Before(beforeCreate) || project.UpdatedAt.After(afterCreate) {
			t.Errorf("UpdatedAt timestamp %v is not between %v and %v",
				project.UpdatedAt, beforeCreate, afterCreate)
		}

		originalUpdatedAt := project.UpdatedAt
		time.Sleep(10 * time.Millisecond) // Ensure time difference

		// Update the project
		beforeUpdate := time.Now().UTC()
		project.Name = "Updated Timestamp Test"
		err = repo.Update(project)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}
		afterUpdate := time.Now().UTC()

		// Verify update timestamp changed
		if !project.UpdatedAt.After(originalUpdatedAt) {
			t.Error("UpdatedAt should be newer after update")
		}

		if project.UpdatedAt.Before(beforeUpdate) || project.UpdatedAt.After(afterUpdate) {
			t.Errorf("Updated UpdatedAt timestamp %v is not between %v and %v",
				project.UpdatedAt, beforeUpdate, afterUpdate)
		}

		// Verify CreatedAt didn't change
		retrieved, err := repo.GetByID(project.ID)
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if !retrieved.CreatedAt.Equal(project.CreatedAt) {
			t.Error("CreatedAt should not change during update")
		}
	})

	t.Run("boolean fields persistence", func(t *testing.T) {
		testCases := []struct {
			name        string
			isActive    bool
			isFavourite bool
		}{
			{"both true", true, true},
			{"both false", false, false},
			{"active only", true, false},
			{"favourite only", false, true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				project := &db.Project{
					Name:        "Bool " + tc.name,
					IsActive:    tc.isActive,
					IsFavourite: tc.isFavourite,
				}

				err := repo.Create(ctx, project)
				if err != nil {
					t.Fatalf("Create failed: %v", err)
				}

				retrieved, err := repo.GetByID(project.ID)
				if err != nil {
					t.Fatalf("GetByID failed: %v", err)
				}

				if retrieved.IsActive != tc.isActive {
					t.Errorf("IsActive = %v, want %v", retrieved.IsActive, tc.isActive)
				}
				if retrieved.IsFavourite != tc.isFavourite {
					t.Errorf("IsFavourite = %v, want %v", retrieved.IsFavourite, tc.isFavourite)
				}
			})
		}
	})
}

// Benchmark tests for performance analysis
func BenchmarkProjectRepo_Create(b *testing.B) {
	prepareTestEnv(&testing.T{})

	if err := db.InitializeDatabase(); err != nil {
		b.Fatalf("Failed to initialize test database: %v", err)
	}
	defer db.CloseDatabase()

	testDB := db.GetDB()
	repo := NewProjectRepo(testDB)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		project := createTestProject("BenchCreate " + fmt.Sprintf("%d", i))
		err := repo.Create(ctx, project)
		if err != nil {
			b.Fatalf("Create failed: %v", err)
		}
	}
}

func BenchmarkProjectRepo_GetByID(b *testing.B) {
	prepareTestEnv(&testing.T{})

	if err := db.InitializeDatabase(); err != nil {
		b.Fatalf("Failed to initialize test database: %v", err)
	}
	defer db.CloseDatabase()

	testDB := db.GetDB()
	repo := NewProjectRepo(testDB)
	ctx := context.Background()

	// Create a test project
	project := createTestProject("BenchGetByID")
	err := repo.Create(ctx, project)
	if err != nil {
		b.Fatalf("Failed to create test project: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.GetByID(project.ID)
		if err != nil {
			b.Fatalf("GetByID failed: %v", err)
		}
	}
}

func BenchmarkProjectRepo_GetAll(b *testing.B) {
	prepareTestEnv(&testing.T{})

	if err := db.InitializeDatabase(); err != nil {
		b.Fatalf("Failed to initialize test database: %v", err)
	}
	defer db.CloseDatabase()

	testDB := db.GetDB()
	repo := NewProjectRepo(testDB)
	ctx := context.Background()

	// Create test projects
	for i := 0; i < 100; i++ {
		project := createTestProject("BenchGetAll " + fmt.Sprintf("%d", i))
		err := repo.Create(ctx, project)
		if err != nil {
			b.Fatalf("Failed to create test project: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.GetAll()
		if err != nil {
			b.Fatalf("GetAll failed: %v", err)
		}
	}
}
