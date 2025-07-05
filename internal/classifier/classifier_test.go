package classifier

import (
	"encoding/json"

	"sort"
	"testing"

	"kalycs/db"
)


func mustJSON(t *testing.T, items []string) string {
	t.Helper()
	b, err := json.Marshal(items)
	if err != nil {
		t.Fatalf("failed to marshal items: %v", err)
	}
	return string(b)
}

func TestCompileRule_RegexCaseSensitivity(t *testing.T) {
	rule := db.Rule{
		ID:            "1",
		ProjectID:     "p1",
		Rule:          "regex",
		Texts:         mustJSON(t, []string{"foo\\d+"}),
		CaseSensitive: true,
	}

	cr, err := compileRule(rule)
	if err != nil {
		t.Fatalf("compileRule returned error: %v", err)
	}

	if !cr.Regexp.MatchString("foo123") {
		t.Error("case sensitive regex failed to match lowercase")
	}
	if cr.Regexp.MatchString("FOO123") {
		t.Error("case sensitive regex matched uppercase")
	}

	rule.ID = "2"
	rule.CaseSensitive = false
	cr2, err := compileRule(rule)
	if err != nil {
		t.Fatalf("compileRule returned error: %v", err)
	}

	if !cr2.Regexp.MatchString("foo123") {
		t.Error("case insensitive regex failed to match lowercase")
	}
	if !cr2.Regexp.MatchString("FOO123") {
		t.Error("case insensitive regex failed to match uppercase")
	}
}

func TestMatches_MultiText(t *testing.T) {
	rule := db.Rule{
		ID:            "ext1",
		ProjectID:     "p1",
		Rule:          "extension",
		Texts:         mustJSON(t, []string{"jpg", "png"}),
		CaseSensitive: false,
	}

	cr, err := compileRule(rule)
	if err != nil {
		t.Fatalf("compileRule error: %v", err)
	}

	if !matches(cr, "photo.jpg", "jpg") {
		t.Error("expected jpg extension to match")
	}
	if !matches(cr, "graphic.PNG", "png") {
		t.Error("expected png extension to match regardless of case")
	}
	if matches(cr, "doc.txt", "txt") {
		t.Error("unexpected match for txt extension")
	}
}

func TestPriorityBehavior(t *testing.T) {
	r1 := db.Rule{
		ID:            "1",
		ProjectID:     "p1",
		Rule:          "starts_with",
		Texts:         mustJSON(t, []string{"report"}),
		CaseSensitive: false,
	}
	r2 := db.Rule{
		ID:            "2",
		ProjectID:     "p2",
		Rule:          "starts_with",
		Texts:         mustJSON(t, []string{"rep"}),
		CaseSensitive: false,
	}

	cr1, err := compileRule(r1)
	if err != nil {
		t.Fatalf("compileRule error: %v", err)
	}
	cr1.Priority = 1

	cr2, err := compileRule(r2)
	if err != nil {
		t.Fatalf("compileRule error: %v", err)
	}
	cr2.Priority = 0

	rules := []CompiledRule{cr1, cr2}
	sort.Slice(rules, func(i, j int) bool { return rules[i].Priority < rules[j].Priority })

	var matched *CompiledRule
	for i := range rules {
		if matches(rules[i], "report_final.txt", "txt") {
			matched = &rules[i]
			break
		}
	}
	if matched == nil {
		t.Fatal("no rule matched")
	}
	if matched.RuleID != cr2.RuleID {
		t.Errorf("expected rule %s to match first, got %s", cr2.RuleID, matched.RuleID)
	}
}
