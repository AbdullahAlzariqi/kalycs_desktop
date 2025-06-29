package store

import (
	"database/sql"
	"kalycs/db"
)

// ruleRepo implements RuleRepo
// (moved from repo.go)
type ruleRepo struct {
	db *sql.DB
}

// RuleRepo defines methods for rule data access
type RuleRepo interface {
	GetByID(id string) (*db.Rule, error)
	GetAllByProject(projectID string) ([]db.Rule, error)
	Create(rule *db.Rule) error
	Update(rule *db.Rule) error
	Delete(id string) error
}

func NewRuleRepo(db *sql.DB) RuleRepo {
	return &ruleRepo{db: db}
}

func (r *ruleRepo) GetByID(id string) (*db.Rule, error) {
	return nil, nil
}

func (r *ruleRepo) GetAllByProject(projectID string) ([]db.Rule, error) {
	return nil, nil
}

func (r *ruleRepo) Create(rule *db.Rule) error {
	return nil
}

func (r *ruleRepo) Update(rule *db.Rule) error {
	return nil
}

func (r *ruleRepo) Delete(id string) error {
	return nil
}
