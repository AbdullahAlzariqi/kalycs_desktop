package database

import (
	"strings"
)

// DatabaseError represents a database-specific error
type DatabaseError struct {
	Type    ErrorType `json:"type"`
	Message string    `json:"message"`
	Table   string    `json:"table,omitempty"`
	Field   string    `json:"field,omitempty"`
}

// ErrorType represents the type of database error
type ErrorType string

const (
	ErrorTypeUniqueConstraint  ErrorType = "unique_constraint"
	ErrorTypeForeignKey        ErrorType = "foreign_key"
	ErrorTypeNotNull           ErrorType = "not_null"
	ErrorTypeCheckConstraint   ErrorType = "check_constraint"
	ErrorTypeConnectionFailed  ErrorType = "connection_failed"
	ErrorTypeTransactionFailed ErrorType = "transaction_failed"
	ErrorTypeUnknown           ErrorType = "unknown"
)

// Error implements the error interface
func (e DatabaseError) Error() string {
	if e.Table != "" && e.Field != "" {
		return e.Message + " (table: " + e.Table + ", field: " + e.Field + ")"
	}
	return e.Message
}

// IsUniqueConstraintError checks if the error is due to a unique constraint violation
func IsUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "unique") ||
		strings.Contains(errStr, "constraint") ||
		strings.Contains(errStr, "duplicate")
}

// IsForeignKeyError checks if the error is due to a foreign key constraint violation
func IsForeignKeyError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "foreign key") ||
		strings.Contains(errStr, "references")
}

// IsNotNullError checks if the error is due to a not null constraint violation
func IsNotNullError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "not null") ||
		strings.Contains(errStr, "null constraint")
}

// ClassifyError attempts to classify a database error into a specific type
func ClassifyError(err error) DatabaseError {
	if err == nil {
		return DatabaseError{Type: ErrorTypeUnknown, Message: "no error"}
	}

	errStr := err.Error()

	switch {
	case IsUniqueConstraintError(err):
		return DatabaseError{
			Type:    ErrorTypeUniqueConstraint,
			Message: "unique constraint violation: " + errStr,
		}
	case IsForeignKeyError(err):
		return DatabaseError{
			Type:    ErrorTypeForeignKey,
			Message: "foreign key constraint violation: " + errStr,
		}
	case IsNotNullError(err):
		return DatabaseError{
			Type:    ErrorTypeNotNull,
			Message: "not null constraint violation: " + errStr,
		}
	default:
		return DatabaseError{
			Type:    ErrorTypeUnknown,
			Message: errStr,
		}
	}
}
