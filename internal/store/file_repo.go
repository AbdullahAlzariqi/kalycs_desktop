package store

import (
	"context"
	"database/sql"
	"fmt"
	"kalycs/db"
	"kalycs/internal/database"
	"kalycs/internal/logging"
)

type FileRepo interface {
	Upsert(ctx context.Context, f *db.File) error
	SetProject(ctx context.Context, fileID string, projectID string) error
	ByProject(ctx context.Context, projectID string) ([]db.File, error)
	GetByPath(ctx context.Context, path string) (*db.File, error)
}

type fileRepo struct {
	db *sql.DB
}

func NewFileRepo(db *sql.DB) FileRepo {
	return &fileRepo{db: db}
}

func (r *fileRepo) GetByPath(ctx context.Context, path string) (*db.File, error) {
	q := `SELECT id, path, name, ext, size, mtime, project_id, created_at, updated_at FROM files WHERE path = ?`
	row := r.db.QueryRowContext(ctx, q, path)
	f := &db.File{}
	err := row.Scan(&f.ID, &f.Path, &f.Name, &f.Ext, &f.Size, &f.Mtime, &f.ProjectID, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found is not an error, just means no file
		}
		return nil, err
	}
	return f, nil
}

func (r *fileRepo) Upsert(ctx context.Context, f *db.File) error {
	// Use ON CONFLICT to perform an upsert. This is more atomic and efficient.
	q := `
	INSERT INTO files (id, path, name, ext, size, mtime, project_id)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(path) DO UPDATE SET
		name = excluded.name,
		ext = excluded.ext,
		size = excluded.size,
		mtime = excluded.mtime,
		project_id = excluded.project_id,
		updated_at = CURRENT_TIMESTAMP`

	// If the file doesn't have an ID, it's new, so we generate one.
	if f.ID == "" {
		f.ID = database.GenerateID()
	}

	_, err := r.db.ExecContext(ctx, q, f.ID, f.Path, f.Name, f.Ext, f.Size, f.Mtime, f.ProjectID)
	if err != nil {
		logging.L().Errorw("Failed to upsert file", "file_path", f.Path, "file_name", f.Name, "error", err)
		return err
	}

	// Log successful upsert with project assignment
	projectID := "unassigned"
	if f.ProjectID.Valid {
		projectID = f.ProjectID.String
	}
	logging.L().Infow("File upserted successfully", "file_path", f.Path, "file_name", f.Name, "project_id", projectID, "size_bytes", f.Size)
	return nil
}

func (r *fileRepo) SetProject(ctx context.Context, fileID string, projectID string) error {
	var pid interface{}
	if projectID == "" {
		pid = nil
	} else {
		pid = projectID
	}

	q := `UPDATE files SET project_id = ? WHERE id = ?`
	result, err := r.db.ExecContext(ctx, q, pid, fileID)
	if err != nil {
		logging.L().Errorw("Failed to set project for file", "file_id", fileID, "project_id", projectID, "error", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logging.L().Errorw("Failed to get rows affected for file project update", "file_id", fileID, "error", err)
		return err
	}

	if rowsAffected == 0 {
		logging.L().Warnw("File project update failed - file not found", "file_id", fileID)
		return fmt.Errorf("file with ID '%s' not found", fileID)
	}

	logging.L().Infow("File project updated successfully", "file_id", fileID, "project_id", projectID)
	return nil
}

func (r *fileRepo) ByProject(ctx context.Context, projectID string) ([]db.File, error) {
	q := `SELECT id, path, name, ext, size, mtime, project_id, created_at, updated_at FROM files WHERE project_id = ?`
	rows, err := r.db.QueryContext(ctx, q, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []db.File
	for rows.Next() {
		var f db.File
		if err := rows.Scan(&f.ID, &f.Path, &f.Name, &f.Ext, &f.Size, &f.Mtime, &f.ProjectID, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		files = append(files, f)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return files, nil
}
