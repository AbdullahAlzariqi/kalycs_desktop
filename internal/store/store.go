package store

import (
	"database/sql"
)

// Store holds all repository instances
type Store struct {
	Project ProjectRepo
	Rule    RuleRepo
	File    FileRepo
}

// NewStore initializes the repository store with the given *sql.DB
func NewStore(db *sql.DB) *Store {
	return &Store{
		Project: NewProjectRepo(db),
		Rule:    NewRuleRepo(db),
		File:    NewFileRepo(db),
	}
}
