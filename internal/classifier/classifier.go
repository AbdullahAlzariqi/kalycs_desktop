package classifier

import (
	"context"
	"database/sql"
	"encoding/json"
	"kalycs/db"
	"kalycs/internal/logging"
	"kalycs/internal/store"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

const IncomingProjectName = "Incoming"

type CompiledRule struct {
	RuleID        string
	ProjectID     string
	Kind          string
	Texts         []string
	CaseSensitive bool
	Regexp        *regexp.Regexp
	Priority      int
}

type Classifier struct {
	mu                sync.RWMutex
	set               []CompiledRule
	store             *store.Store
	incomingProjectID string
}

func NewClassifier(s *store.Store) *Classifier {
	return &Classifier{
		store: s,
	}
}

func (c *Classifier) LoadIncomingProject(ctx context.Context) error {
	incoming, err := c.store.Project.GetByName(ctx, IncomingProjectName)
	if err != nil {
		return err
	}

	if incoming == nil {
		logging.L().Infow("Incoming project not found, creating it", "project_name", IncomingProjectName)
		newProject := &db.Project{
			Name:        IncomingProjectName,
			Description: "Default project for unclassified files.",
			IsActive:    true,
		}
		if err := c.store.Project.Create(ctx, newProject); err != nil {
			return err
		}
		c.incomingProjectID = newProject.ID
	} else {
		c.incomingProjectID = incoming.ID
	}

	logging.L().Infow("Incoming project loaded", "project_name", IncomingProjectName, "project_id", c.incomingProjectID)
	return nil
}

func (c *Classifier) Reload(ctx context.Context) error {
	rules, err := c.store.Rule.ListActive(ctx)
	if err != nil {
		return err
	}

	compiled := make([]CompiledRule, 0, len(rules))
	for _, r := range rules {
		compiledRule, err := compileRule(r)
		if err != nil {
			logging.L().Warnw("Skipping invalid rule", "rule_name", r.Name, "rule_id", r.ID, "error", err)
			continue
		}
		compiled = append(compiled, compiledRule)
	}

	c.mu.Lock()
	c.set = compiled
	c.mu.Unlock()

	logging.L().Infow("Classifier reloaded", "rule_count", len(c.set))
	return nil
}

func compileRule(r db.Rule) (CompiledRule, error) {
	var texts []string
	if err := json.Unmarshal([]byte(r.Texts), &texts); err != nil {
		return CompiledRule{}, err
	}

	cr := CompiledRule{
		RuleID:        r.ID,
		ProjectID:     r.ProjectID,
		Kind:          r.Rule,
		CaseSensitive: r.CaseSensitive,
		Texts:         texts,
	}

	if !cr.CaseSensitive {
		for i, t := range cr.Texts {
			cr.Texts[i] = strings.ToLower(t)
		}
	}

	if cr.Kind == "regex" {
		re, err := regexp.Compile(cr.Texts[0])
		if err != nil {
			return CompiledRule{}, err
		}
		cr.Regexp = re
	}

	return cr, nil
}

func (c *Classifier) Classify(ctx context.Context, absPath string, meta os.FileInfo) error {
	name := meta.Name()
	ext := strings.ToLower(filepath.Ext(name))
	if len(ext) > 0 {
		ext = ext[1:] // remove dot
	}

	c.mu.RLock()
	rules := c.set
	c.mu.RUnlock()

	// TODO: Get default "Incoming" project ID
	projectID := ""
	matchedRule := ""

	for _, r := range rules {
		if matches(r, name, ext) {
			projectID = r.ProjectID
			matchedRule = r.RuleID
			break
		}
	}

	f := &db.File{
		Path:  absPath,
		Name:  name,
		Ext:   ext,
		Size:  meta.Size(),
		Mtime: meta.ModTime(),
	}

	if projectID != "" {
		f.ProjectID = sql.NullString{String: projectID, Valid: true}
		logging.L().Infow("File classified by rule", "file_path", absPath, "file_name", name, "rule_id", matchedRule, "project_id", projectID)
	} else {
		f.ProjectID = sql.NullString{String: c.incomingProjectID, Valid: true}
		logging.L().Infow("File classified to incoming project", "file_path", absPath, "file_name", name, "project_id", c.incomingProjectID)
	}

	err := c.store.File.Upsert(ctx, f)
	if err != nil {
		logging.L().Errorw("Failed to upsert classified file", "file_path", absPath, "file_name", name, "error", err)
	}
	return err
}

func matches(r CompiledRule, name, ext string) bool {
	testName := name
	if !r.CaseSensitive {
		testName = strings.ToLower(testName)
	}

	switch r.Kind {
	case "starts_with":
		return strings.HasPrefix(testName, r.Texts[0])
	case "contains":
		return strings.Contains(testName, r.Texts[0])
	case "ends_with":
		return strings.HasSuffix(testName, r.Texts[0])
	case "extension":
		// extension is already lowercased
		return ext == r.Texts[0]
	case "regex":
		return r.Regexp.MatchString(name)
	}
	return false
}
