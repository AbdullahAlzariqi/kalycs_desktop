package validation

import (
	"encoding/json"
	"fmt"
	"kalycs/db"
	"regexp"
	"strings"
)

type RuleValidator struct{}

func NewRuleValidator() *RuleValidator {
	return &RuleValidator{}
}

func (v *RuleValidator) Validate(r *db.Rule) error {
	// 1. Trim whitespace
	r.Name = strings.TrimSpace(r.Name)

	var texts []string
	if err := json.Unmarshal([]byte(r.Texts), &texts); err != nil {
		return fmt.Errorf("invalid texts format: must be a JSON array of strings")
	}

	trimmedTexts := make([]string, 0, len(texts))
	for _, text := range texts {
		trimmed := strings.TrimSpace(text)
		if trimmed != "" {
			trimmedTexts = append(trimmedTexts, trimmed)
		}
	}

	// 2. Enforce max lengths
	if len(r.Name) == 0 {
		return fmt.Errorf("rule name cannot be empty")
	}
	if len(r.Name) > MaxRuleNameLength {
		return fmt.Errorf("rule name exceeds max length of %d", MaxRuleNameLength)
	}
	if len(trimmedTexts) == 0 {
		return fmt.Errorf("rule must have at least one text")
	}
	if len(trimmedTexts) > MaxRuleTextsItems {
		return fmt.Errorf("rule texts exceed max items of %d", MaxRuleTextsItems)
	}

	for _, text := range trimmedTexts {
		if len(text) > MaxRuleTextLength {
			return fmt.Errorf("rule text '%s' exceeds max length of %d", text, MaxRuleTextLength)
		}
	}

	// Update r.Texts with trimmed and validated texts
	textsJSON, err := json.Marshal(trimmedTexts)
	if err != nil {
		return fmt.Errorf("failed to marshal texts: %w", err)
	}
	r.Texts = string(textsJSON)

	// 3. For regex rules, compile the pattern
	if r.Rule == "regex" {
		if len(trimmedTexts) != 1 {
			return fmt.Errorf("regex rule must have exactly one pattern")
		}
		if _, err := regexp.Compile(trimmedTexts[0]); err != nil {
			return fmt.Errorf("invalid regex pattern: %w", err)
		}
	}

	return nil
}
