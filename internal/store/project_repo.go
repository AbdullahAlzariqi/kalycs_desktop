package store

import (
	"context"
	"database/sql"
	"fmt"

	"kalycs/db"
	"kalycs/internal/database"
	"kalycs/internal/logging"
	"kalycs/internal/validation"
)

// projectRepo implements ProjectRepo
// (moved from repo.go)
type projectRepo struct {
	db *sql.DB
}

// ProjectRepo defines methods for project data access
type ProjectRepo interface {
	GetByID(ctx context.Context, id string) (*db.Project, error)
	GetByName(ctx context.Context, name string) (*db.Project, error)
	GetAll(ctx context.Context) ([]db.Project, error)
	Create(ctx context.Context, project *db.Project) error
	Update(ctx context.Context, project *db.Project) error
	Delete(ctx context.Context, id string) error
}

// NewProjectRepo creates a new instance of ProjectRepo with the given database connection
func NewProjectRepo(db *sql.DB) ProjectRepo {
	return &projectRepo{db: db}
}

func (r *projectRepo) GetByID(ctx context.Context, id string) (*db.Project, error) {
	// Input validation
	if id == "" {
		return nil, fmt.Errorf("project ID cannot be empty")
	}

	if err := validation.ValidateID(id); err != nil {
		return nil, fmt.Errorf("invalid project ID format: %w", err)
	}

	query := `
		SELECT id, name, description, is_active, is_favourite, created_at, updated_at
		FROM projects
		WHERE id = ?
	`

	project := &db.Project{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.IsActive,
		&project.IsFavourite,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project with ID '%s' not found", id)
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return project, nil
}

func (r *projectRepo) GetByName(ctx context.Context, name string) (*db.Project, error) {
	query := `
		SELECT id, name, description, is_active, is_favourite, created_at, updated_at
		FROM projects
		WHERE name = ?
	`

	project := &db.Project{}
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.IsActive,
		&project.IsFavourite,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not an error, just not found
		}
		return nil, fmt.Errorf("failed to get project by name: %w", err)
	}

	return project, nil
}

func (r *projectRepo) GetAll(ctx context.Context) ([]db.Project, error) {
	query := `
		SELECT id, name, description, is_active, is_favourite, created_at, updated_at
		FROM projects
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}
	defer rows.Close()

	var projects []db.Project
	for rows.Next() {
		var project db.Project
		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Description,
			&project.IsActive,
			&project.IsFavourite,
			&project.CreatedAt,
			&project.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, project)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return projects, nil
}

// Create creates a new project with context support for cancellation and timeouts
func (r *projectRepo) Create(ctx context.Context, project *db.Project) error {
	// Input validation
	if project == nil {
		return fmt.Errorf("project cannot be nil")
	}

	// Validate using validation package
	if err := validation.ValidateProject(project); err != nil {
		logging.L().Warnw("Project validation failed", "error", err)
		return fmt.Errorf("validation failed: %w", err)
	}

	// Normalize and prepare data for creation
	database.NormalizeProjectData(project)
	database.PrepareProjectForCreation(project)

	// Direct insert - no transaction needed for simple insert
	query := `
		INSERT INTO projects (id, name, description, is_active, is_favourite, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		project.ID,
		project.Name,
		project.Description,
		project.IsActive,
		project.IsFavourite,
		project.CreatedAt,
		project.UpdatedAt,
	)

	if err != nil {
		// Handle specific database errors using database utilities
		if database.IsUniqueConstraintError(err) {
			logging.L().Warnw("Project creation failed - name already exists", "project_name", project.Name, "error", err)
			return fmt.Errorf("project with name '%s' already exists", project.Name)
		}
		logging.L().Errorw("Failed to create project", "project_name", project.Name, "error", err)
		return fmt.Errorf("failed to create project: %w", err)
	}

	logging.L().Infow("Project created successfully", "project_id", project.ID, "project_name", project.Name)
	return nil
}

func (r *projectRepo) Update(ctx context.Context, project *db.Project) error {
	// Input validation
	if project == nil {
		return fmt.Errorf("project cannot be nil")
	}

	if project.ID == "" {
		return fmt.Errorf("project ID cannot be empty for update")
	}

	// Validate using validation package
	if err := validation.ValidateProject(project); err != nil {
		logging.L().Warnw("Project validation failed during update", "project_id", project.ID, "error", err)
		return fmt.Errorf("validation failed: %w", err)
	}

	// Normalize and prepare data for update
	database.NormalizeProjectData(project)
	database.PrepareProjectForUpdate(project)

	query := `
		UPDATE projects 
		SET name = ?, description = ?, is_active = ?, is_favourite = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		project.Name,
		project.Description,
		project.IsActive,
		project.IsFavourite,
		project.UpdatedAt,
		project.ID,
	)

	if err != nil {
		// Handle specific database errors using database utilities
		if database.IsUniqueConstraintError(err) {
			logging.L().Warnw("Project update failed - name already exists", "project_id", project.ID, "project_name", project.Name, "error", err)
			return fmt.Errorf("project with name '%s' already exists", project.Name)
		}
		logging.L().Errorw("Failed to update project", "project_id", project.ID, "project_name", project.Name, "error", err)
		return fmt.Errorf("failed to update project: %w", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logging.L().Errorw("Failed to get rows affected for project update", "project_id", project.ID, "error", err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		logging.L().Warnw("Project update failed - project not found", "project_id", project.ID)
		return fmt.Errorf("project with ID '%s' not found", project.ID)
	}

	logging.L().Infow("Project updated successfully", "project_id", project.ID, "project_name", project.Name)
	return nil
}

func (r *projectRepo) Delete(ctx context.Context, id string) error {
	// Input validation
	if id == "" {
		return fmt.Errorf("project ID cannot be empty")
	}

	if err := validation.ValidateID(id); err != nil {
		logging.L().Warnw("Invalid project ID format for deletion", "project_id", id, "error", err)
		return fmt.Errorf("invalid project ID format: %w", err)
	}

	query := `DELETE FROM projects WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		// Handle specific database errors using database utilities
		if database.IsForeignKeyError(err) {
			logging.L().Warnw("Project deletion failed - has associated rules", "project_id", id, "error", err)
			return fmt.Errorf("cannot delete project '%s': it has associated rules", id)
		}
		logging.L().Errorw("Failed to delete project", "project_id", id, "error", err)
		return fmt.Errorf("failed to delete project: %w", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logging.L().Errorw("Failed to get rows affected for project deletion", "project_id", id, "error", err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		logging.L().Warnw("Project deletion failed - project not found", "project_id", id)
		return fmt.Errorf("project with ID '%s' not found", id)
	}

	logging.L().Infow("Project deleted successfully", "project_id", id)
	return nil
}
