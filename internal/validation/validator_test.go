package validation

import (
	"strings"
	"testing"

	"kalycs/db"
)

func TestValidateProject(t *testing.T) {
	tests := []struct {
		name    string
		project *db.Project
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil project",
			project: nil,
			wantErr: true,
			errMsg:  "project cannot be nil",
		},
		{
			name: "valid project",
			project: &db.Project{
				Name:        "Test Project",
				Description: "A test project",
				IsActive:    true,
			},
			wantErr: false,
		},
		{
			name: "empty name",
			project: &db.Project{
				Name:     "",
				IsActive: true,
			},
			wantErr: true,
			errMsg:  "project name is required",
		},
		{
			name: "name too long",
			project: &db.Project{
				Name:     "This project name is way too long and exceeds the maximum allowed length",
				IsActive: true,
			},
			wantErr: true,
			errMsg:  "project name must not exceed 25 characters",
		},
		{
			name: "description too long",
			project: &db.Project{
				Name:        "Valid Name",
				Description: "This description is extremely long and exceeds the maximum allowed length for project descriptions. It should cause a validation error because it contains way more than the allowed 200 characters limit for descriptions in the database schema.",
				IsActive:    true,
			},
			wantErr: true,
			errMsg:  "project description must not exceed 200 characters",
		},
		{
			name: "invalid UUID",
			project: &db.Project{
				ID:       "invalid-uuid",
				Name:     "Valid Name",
				IsActive: true,
			},
			wantErr: true,
			errMsg:  "ID must be a valid UUID format",
		},
		{
			name: "valid UUID",
			project: &db.Project{
				ID:       "550e8400-e29b-41d4-a716-446655440000",
				Name:     "Valid Name",
				IsActive: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProject(tt.project)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProject() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateProject() error = %q, want to contain %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestValidateRule(t *testing.T) {
	validProjectID := "550e8400-e29b-41d4-a716-446655440000"

	tests := []struct {
		name    string
		rule    *db.Rule
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil rule",
			rule:    nil,
			wantErr: true,
			errMsg:  "rule cannot be nil",
		},
		{
			name: "valid rule",
			rule: &db.Rule{
				Name:      "Test Rule",
				ProjectID: validProjectID,
				Rule:      "starts_with",
				Texts:     "test",
			},
			wantErr: false,
		},
		{
			name: "invalid rule type",
			rule: &db.Rule{
				Name:      "Test Rule",
				ProjectID: validProjectID,
				Rule:      "invalid_type",
				Texts:     "test",
			},
			wantErr: true,
			errMsg:  "rule type must be one of",
		},
		{
			name: "empty texts",
			rule: &db.Rule{
				Name:      "Test Rule",
				ProjectID: validProjectID,
				Rule:      "contains",
				Texts:     "",
			},
			wantErr: true,
			errMsg:  "rule texts cannot be empty",
		},
		{
			name: "invalid project ID",
			rule: &db.Rule{
				Name:      "Test Rule",
				ProjectID: "invalid-uuid",
				Rule:      "ends_with",
				Texts:     "test",
			},
			wantErr: true,
			errMsg:  "invalid project ID format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRule(tt.rule)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRule() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateRule() error = %q, want to contain %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	// Test error without value
	err1 := ValidationError{
		Field:   "name",
		Message: "is required",
		Value:   "",
	}
	expected1 := "validation failed for field 'name': is required"
	if err1.Error() != expected1 {
		t.Errorf("ValidationError.Error() = %v, want %v", err1.Error(), expected1)
	}

	// Test error with value
	err2 := ValidationError{
		Field:   "email",
		Message: "invalid format",
		Value:   "test@",
	}
	expected2 := "validation failed for field 'email': invalid format (value: test@)"
	if err2.Error() != expected2 {
		t.Errorf("ValidationError.Error() = %v, want %v", err2.Error(), expected2)
	}
}

func TestValidationErrors(t *testing.T) {
	var errors ValidationErrors

	// Test empty errors
	if errors.HasErrors() {
		t.Error("ValidationErrors.HasErrors() = true, want false for empty errors")
	}

	if errors.ToError() != nil {
		t.Error("ValidationErrors.ToError() = non-nil, want nil for empty errors")
	}

	// Add some errors
	errors.Add("name", "is required")
	errors.Add("email", "invalid format", "test@")

	if !errors.HasErrors() {
		t.Error("ValidationErrors.HasErrors() = false, want true after adding errors")
	}

	if errors.ToError() == nil {
		t.Error("ValidationErrors.ToError() = nil, want non-nil after adding errors")
	}

	if len(errors) != 2 {
		t.Errorf("len(ValidationErrors) = %d, want 2", len(errors))
	}
}

func TestValidateID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "valid uuid",
			id:      "550e8400-e29b-41d4-a716-446655440000",
			wantErr: false,
		},
		{
			name:    "invalid uuid",
			id:      "invalid-uuid",
			wantErr: true,
		},
		{
			name:    "empty string",
			id:      "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			id:      "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
