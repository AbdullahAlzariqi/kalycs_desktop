package store

import (
	"database/sql"
	"kalycs/db"
)

// projectRepo implements ProjectRepo
// (moved from repo.go)
type projectRepo struct {
	db *sql.DB
}

// ProjectRepo defines methods for project data access
type ProjectRepo interface {
	GetByID(id string) (*db.Project, error)
	GetAll() ([]db.Project, error)
	Create(project *db.Project) error
	Update(project *db.Project) error
	Delete(id string) error
}

func NewProjectRepo(db *sql.DB) ProjectRepo {
	return &projectRepo{db: db}
}

func (r *projectRepo) GetByID(id string) (*db.Project, error) {
	return nil, nil
}

func (r *projectRepo) GetAll() ([]db.Project, error) {
	return nil, nil
}

func (r *projectRepo) Create(project *db.Project) error {
	return nil
}

func (r *projectRepo) Update(project *db.Project) error {
	return nil
}

func (r *projectRepo) Delete(id string) error {
	return nil
}
