package validation

import "fmt"

// ValidationError represents a validation error with field-specific details
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// Error implements the error interface
func (e ValidationError) Error() string {
	if e.Value != "" {
		return fmt.Sprintf("validation failed for field '%s': %s (value: %s)", e.Field, e.Message, e.Value)
	}
	return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "no validation errors"
	}
	if len(e) == 1 {
		return e[0].Error()
	}
	return fmt.Sprintf("validation failed with %d errors: %s", len(e), e[0].Message)
}

// Add appends a new validation error
func (e *ValidationErrors) Add(field, message string, value ...string) {
	val := ""
	if len(value) > 0 {
		val = value[0]
	}
	*e = append(*e, ValidationError{
		Field:   field,
		Message: message,
		Value:   val,
	})
}

// HasErrors returns true if there are validation errors
func (e ValidationErrors) HasErrors() bool {
	return len(e) > 0
}

// ToError returns the ValidationErrors as an error if there are any errors, otherwise nil
func (e ValidationErrors) ToError() error {
	if e.HasErrors() {
		return e
	}
	return nil
}
