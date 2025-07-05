package watcher_test

import (
	"context"
	"encoding/json"
	"kalycs/db"
	"kalycs/internal/classifier"
	"kalycs/internal/store"
	"kalycs/internal/testutils"
	"kalycs/internal/watcher"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestClassifier(t *testing.T) (*classifier.Classifier, *store.Store) {
	t.Helper()
	testutils.PrepareTestEnv(t)
	db := testutils.SetupTestDB(t)
	s := store.NewStore(db)
	c := classifier.NewClassifier(s)
	ctx := context.Background()
	if err := c.LoadIncomingProject(ctx); err != nil {
		t.Fatalf("failed to load incoming project: %v", err)
	}
	return c, s
}

func TestNewWatcher_Success(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "watcher-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	c, _ := setupTestClassifier(t)
	w, err := watcher.NewWatcher(context.Background(), tempDir, c)
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
	c, _ := setupTestClassifier(t)
	_, err := watcher.NewWatcher(context.Background(), nonExistentPath, c)
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

	c, _ := setupTestClassifier(t)
	w, err := watcher.NewWatcher(context.Background(), tempDir, c)
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

func TestWatcher_FileClassification(t *testing.T) {
	// 1. Setup
	ctx := context.Background()
	c, s := setupTestClassifier(t)

	// Create a test project and rule
	project := &db.Project{
		Name:     "Test Project",
		IsActive: true, // Explicitly set to active
	}
	if err := s.Project.Create(ctx, project); err != nil {
		t.Fatalf("failed to create test project: %v", err)
	}
	t.Logf("Created project with ID: %s", project.ID)

	ruleTexts, _ := json.Marshal([]string{"pdf"})
	rule := &db.Rule{
		Name:      "PDFs",
		ProjectID: project.ID,
		Rule:      "extension",
		Texts:     string(ruleTexts),
	}
	if err := s.Rule.Create(ctx, rule); err != nil {
		t.Fatalf("failed to create test rule: %v", err)
	}
	t.Logf("Created rule with ID: %s, ProjectID: %s, Texts: %s", rule.ID, rule.ProjectID, rule.Texts)

	// Check if the rule was created successfully
	createdRule, err := s.Rule.GetByID(ctx, rule.ID)
	if err != nil {
		t.Fatalf("failed to get created rule: %v", err)
	}
	if createdRule == nil {
		t.Fatal("rule was not found after creation")
	}
	t.Logf("Retrieved rule: %+v", createdRule)

	// Check ListActive
	activeRules, err := s.Rule.ListActive(ctx)
	if err != nil {
		t.Fatalf("failed to list active rules: %v", err)
	}
	t.Logf("ListActive returned %d rules", len(activeRules))
	for i, r := range activeRules {
		t.Logf("Rule %d: %+v", i, r)
	}

	// Reload classifier with the new rule
	if err := c.Reload(ctx); err != nil {
		t.Fatalf("failed to reload classifier: %v", err)
	}

	// Create a temp directory to watch
	tempDir, err := os.MkdirTemp("", "watcher-classify-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 2. Start watcher
	w, err := watcher.NewWatcher(ctx, tempDir, c)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	w.Start()
	defer w.Stop()
	time.Sleep(20 * time.Millisecond) // give watcher time to start

	// 3. Create a file that matches the rule
	filePath := filepath.Join(tempDir, "test-document.pdf")
	if err := os.WriteFile(filePath, []byte("hello"), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// 4. Assert
	// Use a polling mechanism to wait for the file to be classified to avoid flaky tests.
	var file *db.File
	var getErr error
	timeout := time.After(2 * time.Second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatalf("timed out waiting for file to be classified. last error: %v", getErr)
		case <-ticker.C:
			file, getErr = s.File.GetByPath(ctx, filePath)
			if getErr == nil && file != nil {
				goto found
			}
		}
	}

found:
	if !file.ProjectID.Valid || file.ProjectID.String != project.ID {
		t.Errorf("file was not classified into the correct project. Got ProjectID: %v, want: %s", file.ProjectID, project.ID)
	}
}

func TestWatcher_FileNotClassified(t *testing.T) {
	// 1. Setup
	ctx := context.Background()
	c, s := setupTestClassifier(t)

	// Create a temp directory to watch
	tempDir, err := os.MkdirTemp("", "watcher-unclassify-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 2. Start watcher
	w, err := watcher.NewWatcher(ctx, tempDir, c)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	w.Start()
	defer w.Stop()
	time.Sleep(20 * time.Millisecond) // give watcher time to start

	// 3. Create a file that does not match any rule
	filePath := filepath.Join(tempDir, "test-document.txt")
	if err := os.WriteFile(filePath, []byte("hello"), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// 4. Assert
	// Use a polling mechanism to wait for the file to be processed.
	var file *db.File
	var getErr error
	timeout := time.After(2 * time.Second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatalf("timed out waiting for file to be processed. last error: %v", getErr)
		case <-ticker.C:
			file, getErr = s.File.GetByPath(ctx, filePath)
			if getErr == nil && file != nil {
				goto found
			}
		}
	}

found:
	// Files that don't match any rule should be classified into the "Incoming" project
	if !file.ProjectID.Valid {
		t.Errorf("file was not classified into any project, but should have been assigned to 'Incoming' project")
	}

	// Verify it was assigned to the "Incoming" project
	incomingProject, err := s.Project.GetByName(ctx, "Incoming")
	if err != nil {
		t.Fatalf("failed to get Incoming project: %v", err)
	}
	if file.ProjectID.String != incomingProject.ID {
		t.Errorf("file was not classified into the 'Incoming' project. Got ProjectID: %v, want: %s", file.ProjectID.String, incomingProject.ID)
	}
}

func TestWatcher_FileRename(t *testing.T) {
	// 1. Setup
	ctx := context.Background()
	c, s := setupTestClassifier(t)

	// Create a test project and rule
	project := &db.Project{Name: "Test Project", IsActive: true}
	if err := s.Project.Create(ctx, project); err != nil {
		t.Fatalf("failed to create test project: %v", err)
	}
	ruleTexts, _ := json.Marshal([]string{"log"})
	rule := &db.Rule{Name: "Logs", ProjectID: project.ID, Rule: "extension", Texts: string(ruleTexts)}
	if err := s.Rule.Create(ctx, rule); err != nil {
		t.Fatalf("failed to create test rule: %v", err)
	}
	if err := c.Reload(ctx); err != nil {
		t.Fatalf("failed to reload classifier: %v", err)
	}

	// Create a temp directory to watch and a temp file outside of it
	watchDir, err := os.MkdirTemp("", "watcher-rename-watch")
	if err != nil {
		t.Fatalf("Failed to create watch dir: %v", err)
	}
	defer os.RemoveAll(watchDir)

	otherDir, err := os.MkdirTemp("", "watcher-rename-other")
	if err != nil {
		t.Fatalf("Failed to create other dir: %v", err)
	}
	defer os.RemoveAll(otherDir)

	oldPath := filepath.Join(otherDir, "test.tmp")
	newPath := filepath.Join(watchDir, "renamed.log")

	if err := os.WriteFile(oldPath, []byte("log data"), 0600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	// 2. Start watcher
	w, err := watcher.NewWatcher(ctx, watchDir, c)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	w.Start()
	defer w.Stop()
	time.Sleep(20 * time.Millisecond) // give watcher time to start

	// 3. Rename the file into the watched directory
	if err := os.Rename(oldPath, newPath); err != nil {
		t.Fatalf("failed to rename file: %v", err)
	}

	// 4. Assert
	var file *db.File
	var getErr error
	timeout := time.After(2 * time.Second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatalf("timed out waiting for renamed file to be classified. last error: %v", getErr)
		case <-ticker.C:
			file, getErr = s.File.GetByPath(ctx, newPath)
			if getErr == nil && file != nil {
				goto found
			}
		}
	}

found:
	if !file.ProjectID.Valid || file.ProjectID.String != project.ID {
		t.Errorf("renamed file was not classified into the correct project. Got ProjectID: %v, want: %s", file.ProjectID, project.ID)
	}
}

func TestWatcher_DirectoryCreationIsIgnored(t *testing.T) {
	// 1. Setup
	ctx := context.Background()
	c, s := setupTestClassifier(t)

	// Create a temp directory to watch
	tempDir, err := os.MkdirTemp("", "watcher-dir-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 2. Start watcher
	w, err := watcher.NewWatcher(ctx, tempDir, c)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	w.Start()
	defer w.Stop()
	time.Sleep(20 * time.Millisecond) // give watcher time to start

	// 3. Create a directory
	dirPath := filepath.Join(tempDir, "new-directory")
	if err := os.Mkdir(dirPath, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// 4. Assert
	// Give the watcher a moment to process, then check that the dir was NOT added to the files table.
	// We poll for a short time to see if a file entry is created.
	timeout := time.After(200 * time.Millisecond)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			// The timeout was reached, which is the expected outcome as the directory should be ignored.
			goto success
		case <-ticker.C:
			file, err := s.File.GetByPath(ctx, dirPath)
			if err != nil {
				t.Fatalf("unexpected error when checking for directory in store: %v", err)
			}
			if file != nil {
				t.Errorf("directory was added to the files table, but it should have been ignored")
				return
			}
		}
	}

success:
	// If we reach here, it means the timeout occurred without the file ever appearing, which is correct.
}
