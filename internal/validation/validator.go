package validation

import (
	"strings"
	"unicode/utf8"

	"kalycs/db"
)

// ValidateProject validates a project struct and returns any validation errors
func ValidateProject(project *db.Project) error {
	if project == nil {
		return ValidationError{
			Field:   "project",
			Message: "project cannot be nil",
		}
	}

	var errors ValidationErrors

	// Validate name
	if err := validateProjectName(project.Name); err != nil {
		if ve, ok := err.(ValidationError); ok {
			errors = append(errors, ve)
		} else {
			errors.Add("name", err.Error(), project.Name)
		}
	}

	// Validate description
	if err := validateProjectDescription(project.Description); err != nil {
		if ve, ok := err.(ValidationError); ok {
			errors = append(errors, ve)
		} else {
			errors.Add("description", err.Error(), project.Description)
		}
	}

	// Validate ID format if provided
	if project.ID != "" {
		if err := validateUUID(project.ID); err != nil {
			if ve, ok := err.(ValidationError); ok {
				errors = append(errors, ve)
			} else {
				errors.Add("id", err.Error(), project.ID)
			}
		}
	}

	return errors.ToError()
}

// ValidateRule validates a rule struct and returns any validation errors
func ValidateRule(rule *db.Rule) error {
	if rule == nil {
		return ValidationError{
			Field:   "rule",
			Message: "rule cannot be nil",
		}
	}

	var errors ValidationErrors

	// Validate name
	if err := validateRuleName(rule.Name); err != nil {
		if ve, ok := err.(ValidationError); ok {
			errors = append(errors, ve)
		} else {
			errors.Add("name", err.Error(), rule.Name)
		}
	}

	// Validate project ID
	if err := validateUUID(rule.ProjectID); err != nil {
		errors.Add("project_id", "invalid project ID format", rule.ProjectID)
	}

	// Validate rule type
	if err := validateRuleType(rule.Rule); err != nil {
		if ve, ok := err.(ValidationError); ok {
			errors = append(errors, ve)
		} else {
			errors.Add("rule", err.Error(), rule.Rule)
		}
	}

	// Validate texts (should not be empty)
	if strings.TrimSpace(rule.Texts) == "" {
		errors.Add("texts", "rule texts cannot be empty", rule.Texts)
	}

	// Validate ID format if provided
	if rule.ID != "" {
		if err := validateUUID(rule.ID); err != nil {
			if ve, ok := err.(ValidationError); ok {
				errors = append(errors, ve)
			} else {
				errors.Add("id", err.Error(), rule.ID)
			}
		}
	}

	return errors.ToError()
}

// validateProjectName validates project name according to business rules
func validateProjectName(name string) error {
	trimmedName := strings.TrimSpace(name)

	if trimmedName == "" {
		return ValidationError{
			Field:   "name",
			Message: "project name is required",
		}
	}

	nameLength := utf8.RuneCountInString(trimmedName)
	if nameLength > MaxProjectNameLength {
		return ValidationError{
			Field:   "name",
			Message: "project name must not exceed 25 characters",
			Value:   name,
		}
	}

	if nameLength < MinProjectNameLength {
		return ValidationError{
			Field:   "name",
			Message: "project name must be at least 1 character",
			Value:   name,
		}
	}

	// Check for control characters
	if strings.ContainsAny(trimmedName, "\t\n\r\f\v") {
		return ValidationError{
			Field:   "name",
			Message: "project name cannot contain control characters",
			Value:   name,
		}
	}

	// Check for consecutive spaces
	if strings.Contains(trimmedName, "  ") {
		return ValidationError{
			Field:   "name",
			Message: "project name cannot contain consecutive spaces",
			Value:   name,
		}
	}

	return nil
}

// validateProjectDescription validates project description according to business rules
func validateProjectDescription(description string) error {
	if description == "" {
		return nil // Description is optional
	}

	trimmedDescription := strings.TrimSpace(description)
	if utf8.RuneCountInString(trimmedDescription) > MaxProjectDescriptionLength {
		return ValidationError{
			Field:   "description",
			Message: "project description must not exceed 200 characters",
			Value:   description,
		}
	}

	// Check for problematic control characters
	if strings.ContainsAny(trimmedDescription, "\f\v") {
		return ValidationError{
			Field:   "description",
			Message: "project description cannot contain form feed or vertical tab characters",
			Value:   description,
		}
	}

	return nil
}

// validateRuleName validates rule name according to business rules
func validateRuleName(name string) error {
	trimmedName := strings.TrimSpace(name)

	if trimmedName == "" {
		return ValidationError{
			Field:   "name",
			Message: "rule name is required",
		}
	}

	nameLength := utf8.RuneCountInString(trimmedName)
	if nameLength > MaxRuleNameLength {
		return ValidationError{
			Field:   "name",
			Message: "rule name must not exceed 25 characters",
			Value:   name,
		}
	}

	if nameLength < MinRuleNameLength {
		return ValidationError{
			Field:   "name",
			Message: "rule name must be at least 1 character",
			Value:   name,
		}
	}

	// Check for control characters
	if strings.ContainsAny(trimmedName, "\t\n\r\f\v") {
		return ValidationError{
			Field:   "name",
			Message: "rule name cannot contain control characters",
			Value:   name,
		}
	}

	return nil
}

// validateRuleType validates that the rule type is one of the allowed values
func validateRuleType(ruleType string) error {
	for _, validType := range ValidRuleTypes {
		if ruleType == validType {
			return nil
		}
	}

	return ValidationError{
		Field:   "rule",
		Message: "rule type must be one of: starts_with, contains, ends_with, extension, regex",
		Value:   ruleType,
	}
}

// validateUUID validates UUID format
func validateUUID(id string) error {
	if strings.TrimSpace(id) == "" {
		return ValidationError{
			Field:   "id",
			Message: "ID cannot be empty or whitespace",
		}
	}

	if len(id) != UUIDLength || strings.Count(id, "-") != UUIDHyphenCount {
		return ValidationError{
			Field:   "id",
			Message: "ID must be a valid UUID format",
			Value:   id,
		}
	}

	return nil
}
