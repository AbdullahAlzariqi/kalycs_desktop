package database

import (
	"time"

	"kalycs/db"

	"github.com/google/uuid"
)

// GenerateID generates a new UUID string
func GenerateID() string {
	return uuid.New().String()
}

// PrepareProjectForCreation prepares a project for database insertion
// Sets ID if empty and sets creation/update timestamps
func PrepareProjectForCreation(project *db.Project) {
	now := time.Now().UTC()

	if project.ID == "" {
		project.ID = GenerateID()
	}

	project.CreatedAt = now
	project.UpdatedAt = now
}

// PrepareProjectForUpdate prepares a project for database update
// Sets the updated timestamp
func PrepareProjectForUpdate(project *db.Project) {
	project.UpdatedAt = time.Now().UTC()
}

// PrepareRuleForCreation prepares a rule for database insertion
// Sets ID if empty and sets creation/update timestamps
func PrepareRuleForCreation(rule *db.Rule) {
	now := time.Now().UTC()

	if rule.ID == "" {
		rule.ID = GenerateID()
	}

	rule.CreatedAt = now
	rule.UpdatedAt = now
}

// PrepareRuleForUpdate prepares a rule for database update
// Sets the updated timestamp
func PrepareRuleForUpdate(rule *db.Rule) {
	rule.UpdatedAt = time.Now().UTC()
}

// NormalizeProjectData normalizes project data by trimming whitespace
func NormalizeProjectData(project *db.Project) {
	project.Name = normalizeString(project.Name)
	project.Description = normalizeString(project.Description)
}

// NormalizeRuleData normalizes rule data by trimming whitespace
func NormalizeRuleData(rule *db.Rule) {
	rule.Name = normalizeString(rule.Name)
	rule.Texts = normalizeString(rule.Texts)
}

// normalizeString trims whitespace from a string
func normalizeString(s string) string {
	if s == "" {
		return s
	}
	// Only trim if not empty to preserve intentional empty strings
	return trimWhitespace(s)
}

// trimWhitespace is a helper function to trim whitespace
func trimWhitespace(s string) string {
	// Remove leading and trailing whitespace
	start := 0
	end := len(s)

	// Find first non-whitespace character
	for start < end && isWhitespace(s[start]) {
		start++
	}

	// Find last non-whitespace character
	for end > start && isWhitespace(s[end-1]) {
		end--
	}

	return s[start:end]
}

// isWhitespace checks if a byte is a whitespace character
func isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r' || b == '\f' || b == '\v'
}
