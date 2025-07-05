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

// ruleRepo implements RuleRepo
// (moved from repo.go)
type ruleRepo struct {
	db        *sql.DB
	validator *validation.RuleValidator
}

// RuleRepo defines methods for rule data access
type RuleRepo interface {
	GetByID(ctx context.Context, id string) (*db.Rule, error)
	GetAllByProject(ctx context.Context, projectID string) ([]db.Rule, error)
	ListActive(ctx context.Context) ([]db.Rule, error)
	Create(ctx context.Context, rule *db.Rule) error
	Update(ctx context.Context, rule *db.Rule) error
	Delete(ctx context.Context, id string) error
}

func NewRuleRepo(db *sql.DB) RuleRepo {
	return &ruleRepo{
		db:        db,
		validator: validation.NewRuleValidator(),
	}
}

func (r *ruleRepo) GetByID(ctx context.Context, id string) (*db.Rule, error) {
	q := `SELECT id, name, project_id, rule, texts, case_sensitive, created_at, updated_at FROM rules WHERE id = ?`
	row := r.db.QueryRowContext(ctx, q, id)
	rule := &db.Rule{}
	err := row.Scan(&rule.ID, &rule.Name, &rule.ProjectID, &rule.Rule, &rule.Texts, &rule.CaseSensitive, &rule.CreatedAt, &rule.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Consider not found as nil, not an error
		}
		return nil, err
	}
	return rule, nil
}

func (r *ruleRepo) GetAllByProject(ctx context.Context, projectID string) ([]db.Rule, error) {
	q := `SELECT id, name, project_id, rule, texts, case_sensitive, created_at, updated_at FROM rules WHERE project_id = ?`
	rows, err := r.db.QueryContext(ctx, q, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []db.Rule
	for rows.Next() {
		var rule db.Rule
		if err := rows.Scan(&rule.ID, &rule.Name, &rule.ProjectID, &rule.Rule, &rule.Texts, &rule.CaseSensitive, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

func (r *ruleRepo) ListActive(ctx context.Context) ([]db.Rule, error) {
	q := `
        SELECT r.id, r.name, r.project_id, r.rule, r.texts, r.case_sensitive, r.created_at, r.updated_at
        FROM rules r
        INNER JOIN projects p ON r.project_id = p.id
        WHERE p.is_active = 1`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []db.Rule
	for rows.Next() {
		var rule db.Rule
		if err := rows.Scan(&rule.ID, &rule.Name, &rule.ProjectID, &rule.Rule, &rule.Texts, &rule.CaseSensitive, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

func (r *ruleRepo) Create(ctx context.Context, rule *db.Rule) error {
	if err := r.validator.Validate(rule); err != nil {
		logging.L().Warnw("Rule validation failed", "rule_name", rule.Name, "error", err)
		return err
	}
	rule.ID = database.GenerateID()
	q := `INSERT INTO rules (id, name, project_id, rule, texts, case_sensitive) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, q, rule.ID, rule.Name, rule.ProjectID, rule.Rule, rule.Texts, rule.CaseSensitive)
	if err != nil {
		logging.L().Errorw("Failed to create rule", "rule_id", rule.ID, "rule_name", rule.Name, "project_id", rule.ProjectID, "error", err)
		return err
	}
	logging.L().Infow("Rule created successfully", "rule_id", rule.ID, "rule_name", rule.Name, "project_id", rule.ProjectID)
	return nil
}

func (r *ruleRepo) Update(ctx context.Context, rule *db.Rule) error {
	if err := r.validator.Validate(rule); err != nil {
		logging.L().Warnw("Rule validation failed during update", "rule_id", rule.ID, "rule_name", rule.Name, "error", err)
		return err
	}
	q := `UPDATE rules SET name = ?, project_id = ?, rule = ?, texts = ?, case_sensitive = ? WHERE id = ?`
	result, err := r.db.ExecContext(ctx, q, rule.Name, rule.ProjectID, rule.Rule, rule.Texts, rule.CaseSensitive, rule.ID)
	if err != nil {
		logging.L().Errorw("Failed to update rule", "rule_id", rule.ID, "rule_name", rule.Name, "error", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logging.L().Errorw("Failed to get rows affected for rule update", "rule_id", rule.ID, "error", err)
		return err
	}

	if rowsAffected == 0 {
		logging.L().Warnw("Rule update failed - rule not found", "rule_id", rule.ID)
		return fmt.Errorf("rule with ID '%s' not found", rule.ID)
	}

	logging.L().Infow("Rule updated successfully", "rule_id", rule.ID, "rule_name", rule.Name, "project_id", rule.ProjectID)
	return nil
}

func (r *ruleRepo) Delete(ctx context.Context, id string) error {
	q := `DELETE FROM rules WHERE id = ?`
	result, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		logging.L().Errorw("Failed to delete rule", "rule_id", id, "error", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logging.L().Errorw("Failed to get rows affected for rule deletion", "rule_id", id, "error", err)
		return err
	}

	if rowsAffected == 0 {
		logging.L().Warnw("Rule deletion failed - rule not found", "rule_id", id)
		return fmt.Errorf("rule with ID '%s' not found", id)
	}

	logging.L().Infow("Rule deleted successfully", "rule_id", id)
	return nil
}
