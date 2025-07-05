package classifier

import (
	"encoding/json"
	"testing"

	"kalycs/db"
)

func TestCompileRuleRegexCaseInsensitive(t *testing.T) {
	texts, _ := json.Marshal([]string{"test"})
	rule := db.Rule{Rule: "regex", Texts: string(texts), CaseSensitive: false}
	cr, err := compileRule(rule)
	if err != nil {
		t.Fatalf("compileRule returned error: %v", err)
	}
	if cr.Regexp == nil {
		t.Fatalf("compiled regexp is nil")
	}

	if !cr.Regexp.MatchString("TEST") {
		t.Errorf("compiled regex should match case-insensitively")
	}
	if !cr.Regexp.MatchString("test") {
		t.Errorf("compiled regex should match lowercase string")
	}
}

func TestMatchesRegexCaseInsensitive(t *testing.T) {
	texts, _ := json.Marshal([]string{"^file"})
	rule := db.Rule{Rule: "regex", Texts: string(texts), CaseSensitive: false}
	cr, err := compileRule(rule)
	if err != nil {
		t.Fatalf("compileRule returned error: %v", err)
	}

	if !matches(cr, "FILE.TXT", "txt") {
		t.Errorf("regex rule should match filename regardless of case")
	}
	if !matches(cr, "file.txt", "txt") {
		t.Errorf("regex rule should match lowercase filename")
	}
}
